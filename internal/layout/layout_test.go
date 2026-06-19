package layout

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func rt(s string) contracts.RichText { return contracts.RichText{{Text: s}} }

func TestComputeStacksTopLevel(t *testing.T) {
	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.Heading{Level: 2, Text: rt("Title")},
		&contracts.List{Items: []contracts.ListItem{{Text: rt("a")}, {Text: rt("b")}}},
	}}
	lay := Compute(slide, pptx.DefaultTheme())

	if lay.CanvasWidth != int64(pptx.Slide16x9Width) || lay.CanvasHeight != int64(pptx.Slide16x9Height) {
		t.Fatalf("canvas = %dx%d, want 16:9", lay.CanvasWidth, lay.CanvasHeight)
	}
	if len(lay.Placements) != 2 {
		t.Fatalf("got %d placements, want 2", len(lay.Placements))
	}
	margin := int64(pptx.In(0.5))
	bodyW := int64(pptx.Slide16x9Width) - 2*margin
	h0 := lay.Placements[0]
	if h0.Kind != "heading" {
		t.Fatalf("placement[0] kind = %q, want heading", h0.Kind)
	}
	if h0.Box.X != margin || h0.Box.Y != margin || h0.Box.W != bodyW || h0.Box.H != int64(pptx.In(0.6)) {
		t.Fatalf("heading box = %+v (want x=%d y=%d w=%d h=%d)", h0.Box, margin, margin, bodyW, int64(pptx.In(0.6)))
	}
	// list sits below heading by heading-height + SpaceMD gap.
	gap := int64(pptx.DefaultTheme().ResolveSpace(pptx.SpaceMD))
	wantListY := margin + int64(pptx.In(0.6)) + gap
	l := lay.Placements[1]
	if l.Kind != "list" || l.Box.Y != wantListY || l.Box.H != int64(pptx.In(0.32))*2 {
		t.Fatalf("list box = %+v (want y=%d h=%d)", l.Box, wantListY, int64(pptx.In(0.32))*2)
	}
	// paths are "nodes"-prefixed (what the edit tools resolve).
	if len(h0.Path) != 2 || h0.Path[0] != "nodes" || h0.Path[1] != 0 {
		t.Fatalf("heading path = %v, want [nodes 0]", h0.Path)
	}
}

func TestComputeRecursesTwoColumn(t *testing.T) {
	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.TwoColumn{
			Ratio: contracts.Ratio11,
			Left:  []contracts.SlideNode{&contracts.Heading{Level: 3, Text: rt("L")}},
			Right: []contracts.SlideNode{&contracts.Heading{Level: 3, Text: rt("R")}},
		},
	}}
	lay := Compute(slide, pptx.DefaultTheme())
	if len(lay.Placements) != 3 {
		t.Fatalf("got %d placements, want 3 (column + 2 children)", len(lay.Placements))
	}
	var left, right *contracts.NodePlacement
	for i := range lay.Placements {
		p := &lay.Placements[i]
		if len(p.Path) >= 3 && p.Path[2] == "left" {
			left = p
		}
		if len(p.Path) >= 3 && p.Path[2] == "right" {
			right = p
		}
	}
	if left == nil || right == nil {
		t.Fatal("missing left/right child placements with nodes/i/left|right paths")
	}
	if left.Box.X >= right.Box.X {
		t.Fatalf("left.X (%d) should be < right.X (%d)", left.Box.X, right.Box.X)
	}
}

// TestComputeCenteredSlideY asserts that a slide with Align.Vertical = center
// places the first node's Y strictly below the body top edge (slack > 0 because
// a single Hero is much shorter than the body height).
func TestComputeCenteredSlideY(t *testing.T) {
	slide := contracts.Slide{
		Align: contracts.Alignment{Vertical: contracts.VAlignCenter},
		Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Cover"},
		},
	}
	lay := Compute(slide, pptx.DefaultTheme())

	if len(lay.Placements) != 1 {
		t.Fatalf("got %d placements, want 1", len(lay.Placements))
	}
	bodyTop := int64(pptx.In(0.5))
	heroY := lay.Placements[0].Box.Y
	if heroY <= bodyTop {
		t.Fatalf("centered hero Y = %d, want strictly > body top %d", heroY, bodyTop)
	}
}

// TestComputeZeroAlignIdenticalToDefault asserts that a slide with the zero
// Alignment{} produces placements byte-identical to the same slide without an
// explicit Align field. This is the backward-compat regression guard.
func TestComputeZeroAlignIdenticalToDefault(t *testing.T) {
	nodes := []contracts.SlideNode{
		&contracts.Heading{Level: 2, Text: rt("Title")},
		&contracts.List{Items: []contracts.ListItem{{Text: rt("a")}, {Text: rt("b")}}},
	}
	theme := pptx.DefaultTheme()

	withZero := Compute(contracts.Slide{Nodes: nodes, Align: contracts.Alignment{}}, theme)
	withNone := Compute(contracts.Slide{Nodes: nodes}, theme)

	if len(withZero.Placements) != len(withNone.Placements) {
		t.Fatalf("placement count differs: zero=%d none=%d", len(withZero.Placements), len(withNone.Placements))
	}
	for i, a := range withZero.Placements {
		b := withNone.Placements[i]
		if a.Box != b.Box {
			t.Errorf("placement[%d] box differs: zero=%+v none=%+v", i, a.Box, b.Box)
		}
	}
}

// TestComputeBottomAlign asserts that a slide with Align.Vertical = bottom
// places the last node flush with (or near) the body bottom edge.
func TestComputeBottomAlign(t *testing.T) {
	slide := contracts.Slide{
		Align: contracts.Alignment{Vertical: contracts.VAlignBottom},
		Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Cover"},
		},
	}
	lay := Compute(slide, pptx.DefaultTheme())

	if len(lay.Placements) != 1 {
		t.Fatalf("got %d placements, want 1", len(lay.Placements))
	}
	// The hero's bottom should equal the body bottom
	// (body bottom = slide height - margin = Slide16x9Height - In(0.5)).
	margin := int64(pptx.In(0.5))
	bodyBottom := int64(pptx.Slide16x9Height) - margin
	p := lay.Placements[0]
	heroBottom := p.Box.Y + p.Box.H
	if heroBottom != bodyBottom {
		t.Errorf("bottom-aligned hero bottom = %d, want %d (body bottom)", heroBottom, bodyBottom)
	}
}

func TestComputeDetectsOverflow(t *testing.T) {
	// many tall nodes overflow the 6.5" body region.
	var nodes []contracts.SlideNode
	for i := 0; i < 6; i++ {
		nodes = append(nodes, &contracts.Image{AssetID: "asset://x"}) // 3" each
	}
	lay := Compute(contracts.Slide{Nodes: nodes}, pptx.DefaultTheme())
	if !lay.Overflow {
		t.Fatal("6×3in images should overflow the body region")
	}
}
