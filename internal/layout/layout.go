// Package layout computes a slide's node geometry by MIRRORING pptx-go's
// deterministic box-stack layout (scene/render.go). It exists so the editor
// canvas can paint the exact geometry the export uses, without a second layout
// authority re-coded in the browser (the failure mode of the legacy editor).
//
// It calls pptx-go's own exported scene/layout helpers for container splits
// (Columns/Grid — no drift there) and mirrors only preferredHeight + the
// vertical stack. The preferredHeight constants are PINNED to the pptx-go
// version in go.mod; keep them in sync with scene/render.go preferredHeight.
package layout

import (
	"github.com/hurtener/pptx-go/pptx"
	slayout "github.com/hurtener/pptx-go/scene/layout"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// Body margin + container chrome estimates, mirrored from scene/render.go.
var (
	bodyMargin    = pptx.In(0.5)
	cardChromeEst = pptx.EMU(1097280) // ~1.2"
	estGap        = pptx.EMU(137160)  // ~0.15"

	// Content-aware increments and insets (Phase 22 / R1), mirrored verbatim
	// from scene/render.go constants block — PINNED to the pptx-go version in
	// go.mod. Keep in sync with that block.
	quoteLineEst     = pptx.EMU(411480) // ~0.45"; per extra wrapped line of a Quote
	calloutLineEst   = pptx.EMU(274320) // ~0.30"; per extra wrapped line of a Callout body
	calloutInsetEst  = pptx.EMU(182880) // ~0.20"; accent bar + text inset (renderCallout)
	cardBodyInsetEst = pptx.EMU(182880) // ~0.20"; per-side card body padding estimate
)

// Compute returns the canvas geometry for a slide. theme supplies the SpaceMD
// gap (resolved exactly as the renderer does); a nil theme falls back to the
// default theme so the function never panics.
func Compute(slide contracts.Slide, theme *pptx.Theme) contracts.SlideLayout {
	if theme == nil {
		theme = pptx.DefaultTheme()
	}
	gap := theme.ResolveSpace(pptx.SpaceMD)
	cw, ch := pptx.Slide16x9Width, pptx.Slide16x9Height
	body := pptx.Box{X: bodyMargin, Y: bodyMargin, W: cw - 2*bodyMargin, H: ch - 2*bodyMargin}
	full := pptx.Box{X: 0, Y: 0, W: cw, H: ch}

	c := &computer{gap: gap, theme: theme}
	c.stackTop(body, full, slide.Nodes, slide.Align)

	return contracts.SlideLayout{
		CanvasWidth:  int64(cw),
		CanvasHeight: int64(ch),
		Placements:   c.out,
		Overflow:     c.overflow,
	}
}

type computer struct {
	gap      pptx.EMU
	theme    *pptx.Theme
	out      []contracts.NodePlacement
	overflow bool
}

// stackTop mirrors scene.layout(): Decoration and SectionDivider are full-slide
// overlays that do NOT consume body-stack height; every other node stacks in the
// body region via alignedStackIn. Paths keep the ORIGINAL slide.Nodes index so
// the edit tools resolve correctly.
func (c *computer) stackTop(body, full pptx.Box, nodes []contracts.SlideNode, align contracts.Alignment) {
	// Separate full-slide overlays from body-stack nodes, retaining original indices.
	type indexed struct {
		idx  int
		node contracts.SlideNode
	}
	var bodyNodes []indexed
	for i, n := range nodes {
		path := []any{"nodes", i}
		switch n.(type) {
		case *contracts.Decoration, *contracts.SectionDivider:
			c.emit(full, n, path)
		default:
			bodyNodes = append(bodyNodes, indexed{i, n})
		}
	}
	if len(bodyNodes) == 0 {
		return
	}

	// Extract the node slice for alignedStackIn (no original-index info needed
	// there — we pair the returned boxes back up by position below).
	ns := make([]contracts.SlideNode, len(bodyNodes))
	for i, bn := range bodyNodes {
		ns[i] = bn.node
	}
	boxes := c.alignedStackIn(body, ns, align)
	for i, bn := range bodyNodes {
		c.emit(boxes[i], bn.node, []any{"nodes", bn.idx})
	}
}

// alignedStackIn mirrors scene/render.go alignedStackIn exactly (PINNED to the
// pptx-go version in go.mod). With a zero Alignment {VAlignTop, HAlignLeft}
// and no per-node Align overrides the placements are byte-identical to the
// pre-alignment stackIn (backward-compat guarantee).
//
// Returns one pptx.Box per node in ns (same order). Updates c.overflow when
// the total stack height exceeds box.H.
func (c *computer) alignedStackIn(box pptx.Box, ns []contracts.SlideNode, align contracts.Alignment) []pptx.Box {
	n := len(ns)
	if n == 0 {
		return nil
	}

	gap := c.gap

	// Per-node heights and total body height (sum of heights only).
	heights := make([]pptx.EMU, n)
	var bodyH pptx.EMU
	for i, nd := range ns {
		heights[i] = preferredHeight(nd, box.W, c.theme)
		bodyH += heights[i]
	}

	// totalH = bodyH + gap*(n-1); gap appears between nodes, not after the last.
	var gapCount pptx.EMU
	if n > 1 {
		gapCount = pptx.EMU(n - 1)
	}
	totalH := bodyH + gap*gapCount

	// Vertical: compute the Y coordinate of the first node.
	startY := box.Y
	switch align.Vertical {
	case contracts.VAlignCenter:
		slack := box.H - totalH
		if slack > 0 {
			startY = box.Y + slack/2
		}
	case contracts.VAlignBottom:
		candidate := box.Bottom() - totalH
		if candidate > box.Y {
			startY = candidate
		}
		// VAlignTop and VAlignJustify both start at box.Y; Justify adjusts the gap.
	}

	// Effective inter-node gap: VAlignJustify distributes slack into the gaps.
	effectiveGap := gap
	if align.Vertical == contracts.VAlignJustify && n > 1 {
		slack := box.H - bodyH
		if slack > gap*pptx.EMU(n-1) {
			effectiveGap = slack / pptx.EMU(n-1)
		}
	}

	// Overflow: fires when content is taller than the box, same semantics as
	// the engine's stackIn, regardless of how vertical alignment clamped startY.
	if totalH > box.H {
		c.overflow = true
	}

	// VAlignFill: grow the flexible nodes (containers + Image/Chart) to consume
	// the leftover body height, so the last flexible node's bottom reaches
	// box.Bottom(). Top-pinned (startY stays box.Y) with the standard gap; only
	// positive slack is distributed, so fill never overlaps and never fights the
	// overflow case. Mirrors scene/render.go alignedStackIn VAlignFill branch
	// and distributeFill (PINNED to the pptx-go version in go.mod).
	if align.Vertical == contracts.VAlignFill {
		if slack := box.H - totalH; slack > 0 {
			distributeFill(ns, heights, slack)
		}
	}

	out := make([]pptx.Box, n)
	y := startY
	for i, nd := range ns {
		h := heights[i]
		hAlign := nodeEffectiveHAlign(nd, align.Horizontal)

		plBox := pptx.Box{X: box.X, Y: y, W: box.W, H: h}

		if hAlign != contracts.HAlignLeft {
			nw := nodeNaturalWidth(nd, c.theme)
			if nw > box.W {
				nw = box.W
			}
			if nw > 0 && nw < box.W {
				var offsetX pptx.EMU
				switch hAlign {
				case contracts.HAlignCenter:
					offsetX = (box.W - nw) / 2
				case contracts.HAlignRight:
					offsetX = box.W - nw
				}
				plBox.X = box.X + offsetX
				plBox.W = nw
			}
		}

		out[i] = plBox
		y += h + effectiveGap
	}
	return out
}

// nodeEffectiveHAlign mirrors scene/render.go nodeEffectiveHAlign. The per-node
// Align field overrides the slide-level slideHAlign. Containers and visual nodes
// always return HAlignLeft (they keep their full box width). The return value is
// always a canonical non-empty HAlign (never the empty-string zero value).
func nodeEffectiveHAlign(n contracts.SlideNode, slideHAlign contracts.HAlign) contracts.HAlign {
	var nodeAlign contracts.HAlign
	switch v := n.(type) {
	case *contracts.Hero:
		nodeAlign = v.Align
	case *contracts.Heading:
		nodeAlign = v.Align
	case *contracts.Prose:
		nodeAlign = v.Align
	case *contracts.Quote:
		nodeAlign = v.Align
	case *contracts.Chip:
		nodeAlign = v.Align
	case *contracts.SectionDivider:
		nodeAlign = v.Align
	default:
		// Containers and visuals: always full-width; not subject to h-align.
		return contracts.HAlignLeft
	}
	// Non-empty per-node Align overrides the slide default.
	if nodeAlign != "" {
		return nodeAlign
	}
	// Empty slide horizontal (= unset) normalizes to left.
	if slideHAlign == "" {
		return contracts.HAlignLeft
	}
	return slideHAlign
}

// Pinned constants mirrored verbatim from scene/metrics.go (PINNED to the
// pptx-go version in go.mod — keep in sync).
const (
	avgCharWidthFactor = 0.5
	emuPerPointLayout  = 12700 // emuPerPointMetrics in scene/metrics.go
)

// naturalWidthRTAt estimates the rendered width of a contracts.RichText slice
// with a base TypeRole substituted for runs whose Style.TypeRole is empty
// (mirrors scene.naturalWidthAt). Use when the node's rendering base role is
// known (e.g. TypeH2 for a level-2 Heading). Each run contributes:
//
//	len(text) × floor(fontSize_pt × avgCharWidthFactor × emuPerPoint)
func naturalWidthRTAt(rt contracts.RichText, base pptx.TypeRole, theme *pptx.Theme) pptx.EMU {
	if len(rt) == 0 {
		return 0
	}
	var total pptx.EMU
	for _, run := range rt {
		if len(run.Text) == 0 {
			continue
		}
		role := layoutTypeRole(run.TypeRole)
		if role == pptx.TypeDisplay && run.TypeRole == "" {
			role = base
		}
		spec := theme.ResolveType(role)
		avgW := pptx.EMU(spec.Size * avgCharWidthFactor * emuPerPointLayout)
		total += avgW * pptx.EMU(len(run.Text))
	}
	return total
}

// naturalWidthPlain estimates the rendered width of a plain string at role
// (mirrors the Hero/Chip/SectionDivider calls in scene.nodeNaturalWidth which
// use TypeDisplay = zero value).
func naturalWidthPlain(text string, role pptx.TypeRole, theme *pptx.Theme) pptx.EMU {
	if len(text) == 0 {
		return 0
	}
	spec := theme.ResolveType(role)
	avgW := pptx.EMU(spec.Size * avgCharWidthFactor * emuPerPointLayout)
	return avgW * pptx.EMU(len(text))
}

// wrappedLinesLayout estimates how many lines rt occupies when laid out in a
// column of width avail (mirrors scene/metrics.go wrappedLines — PINNED to the
// pptx-go version in go.mod). Uses naturalWidthRTAt with base substituted for
// runs whose TypeRole is empty. Returns 1 when avail ≤ 0 or theme is nil, so
// a content-aware call that lacks a real width/theme reproduces the
// pre-R1 fixed-height output (fallback = single-line height).
func wrappedLinesLayout(rt contracts.RichText, base pptx.TypeRole, avail pptx.EMU, theme *pptx.Theme) int {
	if avail <= 0 || theme == nil {
		return 1
	}
	w := naturalWidthRTAt(rt, base, theme)
	if w <= 0 {
		return 1
	}
	// ceil(w / avail) with positive integers — identical to scene/metrics.go.
	lines := int((w + avail - 1) / avail)
	if lines < 1 {
		lines = 1
	}
	return lines
}

// nodeNaturalWidth mirrors scene/metrics.go nodeNaturalWidth (PINNED).
// Containers and visual nodes return 0 (they are always full-width).
func nodeNaturalWidth(n contracts.SlideNode, theme *pptx.Theme) pptx.EMU {
	switch v := n.(type) {
	case *contracts.Hero:
		// Title is the dominant visual; TypeDisplay is the zero value of
		// pptx.TypeRole — call naturalWidthPlain with TypeDisplay.
		return naturalWidthPlain(v.Title, pptx.TypeDisplay, theme)
	case *contracts.Heading:
		return naturalWidthRTAt(v.Text, headingRole(v.Level), theme)
	case *contracts.Prose:
		if len(v.Paragraphs) == 0 {
			return 0
		}
		return naturalWidthRTAt(v.Paragraphs[0], pptx.TypeBody, theme)
	case *contracts.Quote:
		return naturalWidthRTAt(v.Text, pptx.TypeH3, theme)
	case *contracts.Chip:
		return naturalWidthPlain(v.Label, pptx.TypeBodySmall, theme)
	case *contracts.SectionDivider:
		return naturalWidthPlain(v.Label, pptx.TypeDisplay, theme)
	}
	// Containers and visuals: always full-width (callers should not h-align these).
	return 0
}

// headingRole mirrors scene/render_leaves.go headingRole (PINNED).
func headingRole(level int) pptx.TypeRole {
	switch level {
	case 1:
		return pptx.TypeH1
	case 2:
		return pptx.TypeH2
	case 3:
		return pptx.TypeH3
	case 4:
		return pptx.TypeH4
	default:
		return pptx.TypeH5
	}
}

// layoutTypeRole converts the wire-level contracts.TypeRole string to the
// pptx.TypeRole integer used by the theme resolver.
func layoutTypeRole(r contracts.TypeRole) pptx.TypeRole {
	switch r {
	case contracts.TypeH1:
		return pptx.TypeH1
	case contracts.TypeH2:
		return pptx.TypeH2
	case contracts.TypeH3:
		return pptx.TypeH3
	case contracts.TypeH4:
		return pptx.TypeH4
	case contracts.TypeH5:
		return pptx.TypeH5
	case contracts.TypeBody:
		return pptx.TypeBody
	case contracts.TypeBodySmall:
		return pptx.TypeBodySmall
	case contracts.TypeCaption:
		return pptx.TypeCaption
	case contracts.TypeMono:
		return pptx.TypeMono
	case contracts.TypeCode:
		return pptx.TypeCode
	default:
		// TypeDisplay and unrecognized values → pptx.TypeDisplay (zero).
		return pptx.TypeDisplay
	}
}

// stack lays nodes top-to-bottom in box (full width, preferredHeight, gap),
// mirroring scene.stackIn, emitting a placement per node and recursing into
// containers. Records overflow into the shared computer.
func (c *computer) stack(box pptx.Box, nodes []contracts.SlideNode, prefix []any) {
	y := box.Y
	for i, n := range nodes {
		h := preferredHeight(n, box.W, c.theme)
		nb := pptx.Box{X: box.X, Y: y, W: box.W, H: h}
		c.emit(nb, n, appendPath(prefix, i))
		y += h + c.gap
	}
	if len(nodes) > 0 && y-c.gap > box.Bottom() {
		c.overflow = true
	}
}

// emit records a node's placement and recurses into two_column / grid children
// using pptx-go's own split helpers (exact, no drift). Other containers render
// as a single box in V1 (nested editing deferred).
func (c *computer) emit(box pptx.Box, n contracts.SlideNode, path []any) {
	c.out = append(c.out, contracts.NodePlacement{
		Path: path,
		Kind: string(contracts.KindOf(n)),
		Box:  emuBox(box),
	})
	switch v := n.(type) {
	case *contracts.TwoColumn:
		cols := slayout.Columns(box, ratioWeights(v.Ratio), c.gap)
		if len(cols) == 2 {
			c.stack(cols[0], v.Left, appendPath(path, "left"))
			c.stack(cols[1], v.Right, appendPath(path, "right"))
		}
	case *contracts.Grid:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		cells := slayout.Grid(box, cols, v.Ratio, c.gap, len(v.Cells))
		for j, cell := range v.Cells {
			if j < len(cells) {
				c.emit(cells[j], cell, appendPath(path, "cells", j))
			}
		}
	}
}

// preferredHeight mirrors scene/render.go preferredHeight EXACTLY (PINNED to the
// pptx-go version in go.mod — R1 / Phase 22 content-aware update). Keep in sync.
//
// Text-bearing nodes are content-aware: their height grows with the number of
// wrapped lines estimated by wrappedLinesLayout. avail ≤ 0 or a nil theme falls
// back to a single-line height, reproducing the pre-R1 fixed output byte-for-byte.
// Visual/atom nodes (Hero, Divider, Chip, Arrow, Image, Chart, CodeBlock, Flow)
// do not wrap and keep fixed slot heights.
func preferredHeight(n contracts.SlideNode, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	switch v := n.(type) {
	case *contracts.Hero:
		return pptx.In(2.2)
	case *contracts.Heading:
		lines := wrappedLinesLayout(v.Text, headingRole(v.Level), avail, theme)
		return pptx.In(0.6) * pptx.EMU(lines)
	case *contracts.Prose:
		if len(v.Paragraphs) == 0 {
			return pptx.In(0.4)
		}
		var h pptx.EMU
		for _, para := range v.Paragraphs {
			lines := wrappedLinesLayout(para, pptx.TypeBody, avail, theme)
			h += pptx.In(0.4) * pptx.EMU(lines)
		}
		return h
	case *contracts.List:
		if len(v.Items) == 0 {
			return pptx.In(0.32)
		}
		var h pptx.EMU
		for _, item := range v.Items {
			lines := wrappedLinesLayout(item.Text, pptx.TypeBody, avail, theme)
			h += pptx.In(0.32) * pptx.EMU(lines)
		}
		return h
	case *contracts.Divider:
		return pptx.In(0.2)
	case *contracts.Quote:
		// Fixed chrome (attribution + padding) = In(1.1); each extra wrapped
		// line of the quote text adds quoteLineEst (mirrors scene/render.go).
		lines := wrappedLinesLayout(v.Text, pptx.TypeH3, avail, theme)
		return pptx.In(1.1) + quoteLineEst*pptx.EMU(lines-1)
	case *contracts.Callout:
		// Body wraps within the box minus the accent bar + text inset
		// (mirrors renderCallout's In(0.2) inset via calloutInsetEst).
		lines := wrappedLinesLayout(v.Body, pptx.TypeBody, avail-calloutInsetEst, theme)
		return pptx.In(1.0) + calloutLineEst*pptx.EMU(lines-1)
	case *contracts.Chip:
		return pptx.In(0.4)
	case *contracts.Arrow:
		return pptx.In(0.6)
	case *contracts.CodeBlock:
		return pptx.In(2.6)
	case *contracts.Image:
		return pptx.In(3.0)
	case *contracts.Chart:
		return pptx.In(3.0)
	case *contracts.Table:
		return tableHeightLayout(v, avail, theme)
	case *contracts.TwoColumn:
		colW := (avail - estGap) / 2
		return maxEMU(nodesHeight(v.Left, colW, theme), nodesHeight(v.Right, colW, theme))
	case *contracts.Grid:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		rows := (len(v.Cells) + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
		cellW := (avail - estGap*pptx.EMU(cols-1)) / pptx.EMU(cols)
		var maxCell pptx.EMU
		for _, cell := range v.Cells {
			if hh := preferredHeight(cell, cellW, theme); hh > maxCell {
				maxCell = hh
			}
		}
		return pptx.EMU(rows)*maxCell + estGap*pptx.EMU(rows-1)
	case *contracts.Card:
		return cardChromeEst + nodesHeight(v.Body, avail-2*cardBodyInsetEst, theme) + estGap
	case *contracts.CardSection:
		return cardChromeEst + nodesHeight(v.Body, avail-2*cardBodyInsetEst, theme) + estGap
	case *contracts.Flow:
		if v.Orientation == contracts.FlowVertical {
			return pptx.In(0.9) * pptx.EMU(atLeast(len(v.Steps), 1))
		}
		return pptx.In(1.4)
	case *contracts.Bento:
		// Mirrors scene/render_bento.go bentoGeometry: each row is allotted an
		// equal fraction of the available height. We estimate the minimum row
		// height as a fixed 1.4" × row count plus gaps (PINNED to the pptx-go
		// version in go.mod). The Bento is flexible so VAlignFill will grow it.
		nRows := atLeast(len(v.Rows), 1)
		return pptx.In(1.4)*pptx.EMU(nRows) + estGap*pptx.EMU(nRows-1)
	default:
		return pptx.In(1.0)
	}
}

// nodesHeight estimates the stacked height of a node list laid out in a column
// of width avail (mirrors scene/render.go nodesHeight — PINNED).
func nodesHeight(nodes []contracts.SlideNode, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	var total pptx.EMU
	for i, n := range nodes {
		total += preferredHeight(n, avail, theme)
		if i > 0 {
			total += estGap
		}
	}
	return total
}

// tableHeightLayout mirrors scene/render.go tableHeight (PINNED). When avail,
// cols, or theme are unavailable it falls back to the count-based pre-R1 height.
func tableHeightLayout(v *contracts.Table, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	cols := tableColumnsLayout(v)
	if cols < 1 || avail <= 0 || theme == nil {
		rows := len(v.Rows)
		if len(v.Headers) > 0 {
			rows++
		}
		h := pptx.In(0.4) * pptx.EMU(rows)
		if v.Caption != "" {
			h += pptx.In(0.4)
		}
		return h
	}
	colW := avail / pptx.EMU(cols)
	var h pptx.EMU
	if len(v.Headers) > 0 {
		h += tableRowHeightLayout(v.Headers, colW, theme)
	}
	for _, row := range v.Rows {
		h += tableRowHeightLayout(row, colW, theme)
	}
	if v.Caption != "" {
		h += pptx.In(0.4)
	}
	return h
}

// tableColumnsLayout mirrors scene/render_table.go tableColumns (PINNED).
func tableColumnsLayout(v *contracts.Table) int {
	cols := len(v.Headers)
	for _, row := range v.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return cols
}

// tableRowHeightLayout is In(0.4) × the wrapped line count of the tallest cell
// in the row (mirrors scene/render.go tableRowHeight — PINNED).
func tableRowHeightLayout(cells []contracts.RichText, colW pptx.EMU, theme *pptx.Theme) pptx.EMU {
	maxLines := 1
	for _, cell := range cells {
		if l := wrappedLinesLayout(cell, pptx.TypeBody, colW, theme); l > maxLines {
			maxLines = l
		}
	}
	return pptx.In(0.4) * pptx.EMU(maxLines)
}

func ratioWeights(rt contracts.ColumnRatio) []int {
	switch rt {
	case contracts.Ratio12:
		return []int{1, 2}
	case contracts.Ratio21:
		return []int{2, 1}
	default:
		return []int{1, 1}
	}
}

func maxEMU(a, b pptx.EMU) pptx.EMU {
	if a > b {
		return a
	}
	return b
}

func atLeast(n, floor int) int {
	if n < floor {
		return floor
	}
	return n
}

func emuBox(b pptx.Box) contracts.EMUBox {
	return contracts.EMUBox{X: int64(b.X), Y: int64(b.Y), W: int64(b.W), H: int64(b.H)}
}

// isFlexible mirrors scene/render.go isFlexible (PINNED). The flexible node
// set is the containers (Grid, TwoColumn, Bento, Card, CardSection) and the
// two stretchable visuals (Table, Chart, Image). Text leaves, atoms, CodeBlock,
// and Flow are fixed — stretching text or monospaced code rasters is
// meaningless or distorts the output.
func isFlexible(n contracts.SlideNode) bool {
	switch n.(type) {
	case *contracts.Grid, *contracts.TwoColumn, *contracts.Bento,
		*contracts.Card, *contracts.CardSection,
		*contracts.Table, *contracts.Chart, *contracts.Image:
		return true
	}
	return false
}

// distributeFill mirrors scene/render.go distributeFill (PINNED). It grows
// the flexible nodes in place so their combined added heights equal exactly
// slack (slack > 0). The share is proportional to each flexible node's
// preferred height (larger nodes grow more, relative proportions preserved);
// the rounding remainder is assigned to the last flexible node so the total
// is exact. When all flexible heights are zero, slack is split equally. Pure
// integer EMU arithmetic — result is deterministic regardless of scheduling.
func distributeFill(nodes []contracts.SlideNode, heights []pptx.EMU, slack pptx.EMU) {
	var flex []int
	var flexH pptx.EMU
	for i, nd := range nodes {
		if isFlexible(nd) {
			flex = append(flex, i)
			flexH += heights[i]
		}
	}
	if len(flex) == 0 {
		return
	}
	var used pptx.EMU
	for k, idx := range flex {
		var add pptx.EMU
		switch {
		case k == len(flex)-1:
			add = slack - used // last flexible node absorbs the rounding remainder
		case flexH > 0:
			add = slack * heights[idx] / flexH
		default:
			add = slack / pptx.EMU(len(flex)) // all flexible heights zero → equal split
		}
		heights[idx] += add
		used += add
	}
}

// appendPath returns a fresh path slice (never aliases prefix's backing array).
func appendPath(prefix []any, legs ...any) []any {
	p := make([]any, 0, len(prefix)+len(legs))
	p = append(p, prefix...)
	p = append(p, legs...)
	return p
}
