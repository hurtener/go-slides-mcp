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

// TestR1ContentAwareHeightMultiLineProse asserts that a Prose node with a very
// long paragraph is allotted ≥ 2 line-heights (R1 mirror fidelity), and that a
// second stacked node's Y is strictly below the Prose's bottom edge (no overlap).
func TestR1ContentAwareHeightMultiLineProse(t *testing.T) {
	theme := pptx.DefaultTheme()
	// Build a long string guaranteed to wrap in the body column.
	// Body width ≈ Slide16x9Width − 2×bodyMargin ≈ 11.3M EMU.
	// TypeBody avgW ≈ 14pt × 0.5 × 12700 ≈ 88,900 EMU/char.
	// ~200 chars → ~17.7M EMU → at least 2 lines.
	longText := ""
	for i := 0; i < 200; i++ {
		longText += "w"
	}

	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.Prose{Paragraphs: []contracts.RichText{rt(longText)}},
		&contracts.Heading{Level: 2, Text: rt("After")},
	}}
	lay := Compute(slide, theme)

	if len(lay.Placements) != 2 {
		t.Fatalf("got %d placements, want 2", len(lay.Placements))
	}
	prose := lay.Placements[0]
	heading := lay.Placements[1]

	// The prose height must be ≥ 2 × In(0.4) (at least 2 lines).
	twoLineH := int64(pptx.In(0.4)) * 2
	if prose.Box.H < twoLineH {
		t.Fatalf("prose height %d < 2×In(0.4)=%d; R1 content-aware height not applied",
			prose.Box.H, twoLineH)
	}

	// The heading must start below the prose bottom edge (no overlap).
	proseBottom := prose.Box.Y + prose.Box.H
	if heading.Box.Y <= proseBottom {
		t.Fatalf("heading Y=%d ≤ prose bottom=%d: nodes overlap", heading.Box.Y, proseBottom)
	}
}

// TestR1ContentAwareHeightMultiLineList asserts that a List with long items is
// taller than a single-line-per-item height (R1 list wrapping).
func TestR1ContentAwareHeightMultiLineList(t *testing.T) {
	theme := pptx.DefaultTheme()
	// ~150 chars per item wraps to >1 line at body width.
	longItem := ""
	for i := 0; i < 150; i++ {
		longItem += "x"
	}

	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.List{Items: []contracts.ListItem{
			{Text: rt(longItem)},
			{Text: rt(longItem)},
		}},
	}}
	lay := Compute(slide, theme)
	if len(lay.Placements) != 1 {
		t.Fatalf("got %d placements, want 1", len(lay.Placements))
	}
	list := lay.Placements[0]

	// Single-line height for 2 items = 2 × In(0.32). With wrapping, it must be larger.
	twoItemH := int64(pptx.In(0.32)) * 2
	if list.Box.H <= twoItemH {
		t.Fatalf("list height %d ≤ 2×In(0.32)=%d; R1 content-aware height not applied",
			list.Box.H, twoItemH)
	}
}

// TestR1WrappedContentOverflow asserts that layout.Compute sets Overflow=true
// when a Prose with many long paragraphs overflows the body region (C7 mirror).
// Before R1 the fixed height under-counted; now the snapshot matches the engine.
func TestR1WrappedContentOverflow(t *testing.T) {
	theme := pptx.DefaultTheme()
	// Build a slide whose wrapped content exceeds the body height (~7.5" at 16:9).
	// Each Prose paragraph wraps to ~2+ lines; stacking many pushes past body.H.
	longPara := ""
	for i := 0; i < 200; i++ {
		longPara += "m"
	}
	// 15 paragraphs of ~2 lines each = ~30 × In(0.4) ≈ 12" >> 6.5" body.
	paras := make([]contracts.RichText, 15)
	for i := range paras {
		paras[i] = rt(longPara)
	}
	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.Prose{Paragraphs: paras},
	}}
	lay := Compute(slide, theme)
	if !lay.Overflow {
		t.Fatalf("slide with many wrapped paragraphs should set Overflow=true (C7)")
	}
}

// TestR1ShortContentUnchanged asserts that single-line / short content produces
// the same geometry as the pre-R1 fixed-height values (backward-compat).
func TestR1ShortContentUnchanged(t *testing.T) {
	theme := pptx.DefaultTheme()
	slide := contracts.Slide{Nodes: []contracts.SlideNode{
		&contracts.Heading{Level: 2, Text: rt("Hi")},
		&contracts.Prose{Paragraphs: []contracts.RichText{rt("Short.")}},
		&contracts.List{Items: []contracts.ListItem{{Text: rt("one")}, {Text: rt("two")}}},
	}}
	lay := Compute(slide, theme)

	if len(lay.Placements) != 3 {
		t.Fatalf("got %d placements, want 3", len(lay.Placements))
	}
	h0 := lay.Placements[0]
	h1 := lay.Placements[1]
	h2 := lay.Placements[2]

	// Heading: "Hi" is 2 chars, 1 line → In(0.6).
	if h0.Box.H != int64(pptx.In(0.6)) {
		t.Errorf("short heading H=%d, want %d (1 line)", h0.Box.H, int64(pptx.In(0.6)))
	}
	// Prose: "Short." is short, 1 line → In(0.4).
	if h1.Box.H != int64(pptx.In(0.4)) {
		t.Errorf("short prose H=%d, want %d (1 line)", h1.Box.H, int64(pptx.In(0.4)))
	}
	// List: "one" and "two" are short → 2 × In(0.32).
	if h2.Box.H != int64(pptx.In(0.32))*2 {
		t.Errorf("short list H=%d, want %d (2 single-line items)", h2.Box.H, int64(pptx.In(0.32))*2)
	}
}
