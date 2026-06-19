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

	c := &computer{gap: gap}
	overflow := c.stack(body, slide.Nodes, []any{"nodes"})

	return contracts.SlideLayout{
		CanvasWidth:  int64(cw),
		CanvasHeight: int64(ch),
		Placements:   c.out,
		Overflow:     overflow,
	}
}

type computer struct {
	gap pptx.EMU
	out []contracts.NodePlacement
}

// stack lays nodes top-to-bottom in box (full width, preferredHeight, gap),
// mirroring scene.stackIn, emitting a placement per node and recursing into
// containers. Returns true if the content overflows box.
func (c *computer) stack(box pptx.Box, nodes []contracts.SlideNode, prefix []any) bool {
	y := box.Y
	for i, n := range nodes {
		h := preferredHeight(n)
		nb := pptx.Box{X: box.X, Y: y, W: box.W, H: h}
		c.emit(nb, n, appendPath(prefix, i))
		y += h + c.gap
	}
	return len(nodes) > 0 && y-c.gap > box.Bottom()
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
