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
		heights[i] = preferredHeight(nd)
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
		role := layoutTypeRole(run.Style.TypeRole)
		if role == pptx.TypeDisplay && run.Style.TypeRole == "" {
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
		h := preferredHeight(n)
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
// pptx-go version in go.mod). Keep in sync.
func preferredHeight(n contracts.SlideNode) pptx.EMU {
	switch v := n.(type) {
	case *contracts.Hero:
		return pptx.In(2.2)
	case *contracts.Heading:
		return pptx.In(0.6)
	case *contracts.Prose:
		return pptx.In(0.4) * pptx.EMU(atLeast(len(v.Paragraphs), 1))
	case *contracts.List:
		return pptx.In(0.32) * pptx.EMU(atLeast(len(v.Items), 1))
	case *contracts.Divider:
		return pptx.In(0.2)
	case *contracts.Quote:
		return pptx.In(1.1)
	case *contracts.Callout:
		return pptx.In(1.0)
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
		rows := len(v.Rows)
		if len(v.Headers) > 0 {
			rows++
		}
		h := pptx.In(0.4) * pptx.EMU(rows)
		if v.Caption != "" {
			h += pptx.In(0.4)
		}
		return h
	case *contracts.TwoColumn:
		return maxEMU(nodesHeight(v.Left), nodesHeight(v.Right))
	case *contracts.Grid:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		rows := (len(v.Cells) + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
		var maxCell pptx.EMU
		for _, cell := range v.Cells {
			if hh := preferredHeight(cell); hh > maxCell {
				maxCell = hh
			}
		}
		return pptx.EMU(rows)*maxCell + estGap*pptx.EMU(rows-1)
	case *contracts.Card:
		return cardChromeEst + nodesHeight(v.Body) + estGap
	case *contracts.CardSection:
		return cardChromeEst + nodesHeight(v.Body) + estGap
	case *contracts.Flow:
		if v.Orientation == contracts.FlowVertical {
			return pptx.In(0.9) * pptx.EMU(atLeast(len(v.Steps), 1))
		}
		return pptx.In(1.4)
	default:
		return pptx.In(1.0)
	}
}

func nodesHeight(nodes []contracts.SlideNode) pptx.EMU {
	var total pptx.EMU
	for i, n := range nodes {
		total += preferredHeight(n)
		if i > 0 {
			total += estGap
		}
	}
	return total
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

// appendPath returns a fresh path slice (never aliases prefix's backing array).
func appendPath(prefix []any, legs ...any) []any {
	p := make([]any, 0, len(prefix)+len(legs))
	p = append(p, prefix...)
	p = append(p, legs...)
	return p
}
