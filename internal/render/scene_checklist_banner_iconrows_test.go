package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeChecklistAllFields asserts a fully-populated Checklist maps to
// a scene.Checklist with every field set (R12.2, D-095), incl. per-item
// State mapping and GlyphTone pointer wrapping.
func TestMapNodeChecklistAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Checklist{
		Items: []contracts.ChecklistItem{
			{Text: contracts.RichText{{Text: "Done by default"}}, State: contracts.CheckDone},
			{Text: contracts.RichText{{Text: "Not here"}}, State: contracts.CheckNo},
			{Text: contracts.RichText{{Text: "Later"}}, State: contracts.CheckNeutral},
		},
		Columns:   2,
		GlyphTone: contracts.ColorAccent,
		Fill:      true,
	}
	sn := mapNode(node)
	c, ok := sn.(scene.Checklist)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Checklist", sn)
	}
	if c.Columns != 2 {
		t.Errorf("Columns: got %d, want 2", c.Columns)
	}
	if !c.Fill {
		t.Error("Fill: got false, want true")
	}
	if c.GlyphTone == nil || *c.GlyphTone != scene.ColorAccent {
		t.Errorf("GlyphTone: got %v, want pointer-to-ColorAccent", c.GlyphTone)
	}
	if len(c.Items) != 3 {
		t.Fatalf("len(Items): got %d, want 3", len(c.Items))
	}
	if c.Items[0].State != scene.CheckDone {
		t.Errorf("Items[0].State: got %v, want scene.CheckDone", c.Items[0].State)
	}
	if c.Items[1].State != scene.CheckNo {
		t.Errorf("Items[1].State: got %v, want scene.CheckNo", c.Items[1].State)
	}
	if c.Items[2].State != scene.CheckNeutral {
		t.Errorf("Items[2].State: got %v, want scene.CheckNeutral", c.Items[2].State)
	}
}

// TestMapNodeChecklistGlyphToneNil asserts an empty GlyphTone maps to a
// NIL engine *ColorRole (D-054 pointer-sentinel pattern) — preserves the
// engine's per-state glyph color default for the additive byte-identical path.
func TestMapNodeChecklistGlyphToneNil(t *testing.T) {
	t.Parallel()

	sn := mapNode(&contracts.Checklist{Items: []contracts.ChecklistItem{
		{Text: contracts.RichText{{Text: "x"}}},
	}})
	c, ok := sn.(scene.Checklist)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Checklist", sn)
	}
	if c.GlyphTone != nil {
		t.Errorf("GlyphTone: got %v, want nil (empty product string = engine nil ptr)", c.GlyphTone)
	}
}

// TestMapNodeBannerAllFields asserts a fully-populated Banner maps to a
// scene.Banner with every field set (R12.6, D-097), incl. Trailing
// child nesting + explicit Fill (no auto-promotion).
func TestMapNodeBannerAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Banner{
		Lead:      contracts.RichText{{Text: "Run it internally"}},
		Body:      contracts.RichText{{Text: "Or sell it externally."}},
		Icon:      "star",
		Fill:      contracts.ColorAccent,
		TextColor: contracts.TextInverse,
		Trailing: []contracts.SlideNode{
			&contracts.Button{
				Label:        "Start free",
				Tone:         contracts.ButtonGhost,
				TrailingIcon: "arrow-right",
			},
		},
	}
	sn := mapNode(node)
	b, ok := sn.(scene.Banner)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Banner", sn)
	}
	if len(b.Lead) == 0 {
		t.Error("Lead: got empty, want non-empty")
	}
	if b.Icon != "star" {
		t.Errorf("Icon: got %q, want %q", b.Icon, "star")
	}
	if b.Fill != scene.ColorAccent {
		t.Errorf("Fill: got %v, want scene.ColorAccent", b.Fill)
	}
	if b.TextColor != scene.TextInverse {
		t.Errorf("TextColor: got %v, want scene.TextInverse", b.TextColor)
	}
	if len(b.Trailing) != 1 {
		t.Fatalf("len(Trailing): got %d, want 1", len(b.Trailing))
	}
	if _, ok := b.Trailing[0].(scene.Button); !ok {
		t.Errorf("Trailing[0]: got %T, want scene.Button", b.Trailing[0])
	}
}

// TestMapNodeBannerFillEmptyPromotesToCanvas asserts the Banner.Fill
// special-case: an empty product ColorRole string maps to the engine's
// ColorCanvas zero (NOT the generic ColorSurface default), so the
// renderer promotes it to ColorAccent (D-097). The generic
// mapColorRole("") = ColorSurface would break this contract.
func TestMapNodeBannerFillEmptyPromotesToCanvas(t *testing.T) {
	t.Parallel()

	sn := mapNode(&contracts.Banner{
		Lead: contracts.RichText{{Text: "Hello"}},
	})
	b, ok := sn.(scene.Banner)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Banner", sn)
	}
	if b.Fill != scene.ColorCanvas {
		t.Errorf("Fill: got %v, want scene.ColorCanvas (engine zero, so renderer promotes to Accent)", b.Fill)
	}
}

// TestMapNodeIconRowsAllFields asserts a fully-populated IconRows maps to
// scene.IconRows with every field set (R12.7, D-100), incl. per-row Tone
// + Meta rich text.
func TestMapNodeIconRowsAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.IconRows{
		Rows: []contracts.IconRow{
			{Icon: "check", Label: contracts.RichText{{Text: "Chat"}}, Tone: contracts.RowPlain},
			{Icon: "star", Label: contracts.RichText{{Text: "Agents"}}, Meta: contracts.RichText{{Text: "12"}}, Tone: contracts.RowPill},
		},
		Fill:       true,
		GlyphColor: contracts.ColorAccent,
	}
	sn := mapNode(node)
	ir, ok := sn.(scene.IconRows)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.IconRows", sn)
	}
	if !ir.Fill {
		t.Error("Fill: got false, want true")
	}
	if ir.GlyphColor != scene.ColorAccent {
		t.Errorf("GlyphColor: got %v, want scene.ColorAccent", ir.GlyphColor)
	}
	if len(ir.Rows) != 2 {
		t.Fatalf("len(Rows): got %d, want 2", len(ir.Rows))
	}
	if ir.Rows[1].Tone != scene.RowPill {
		t.Errorf("Rows[1].Tone: got %v, want scene.RowPill", ir.Rows[1].Tone)
	}
	if len(ir.Rows[1].Meta) == 0 {
		t.Error("Rows[1].Meta: got empty, want non-empty")
	}
}

// TestMapNodeIconRowsGlyphColorEmptyPromotesToCanvas asserts the
// IconRows.GlyphColor special-case mirrors mapBannerFill: an empty
// product ColorRole string maps to scene.ColorCanvas so the renderer
// promotes it to ColorAccent (D-100).
func TestMapNodeIconRowsGlyphColorEmptyPromotesToCanvas(t *testing.T) {
	t.Parallel()

	sn := mapNode(&contracts.IconRows{
		Rows: []contracts.IconRow{{Icon: "star", Label: contracts.RichText{{Text: "x"}}}},
	})
	ir, ok := sn.(scene.IconRows)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.IconRows", sn)
	}
	if ir.GlyphColor != scene.ColorCanvas {
		t.Errorf("GlyphColor: got %v, want scene.ColorCanvas (engine zero, so renderer promotes to Accent)", ir.GlyphColor)
	}
}

// TestRenderChecklistBannerIconRowsEffect verifies the three new nodes
// render REAL shapes (strictly more than the baseline), produce a valid
// PPTX, and are byte-identical across repeated renders and worker counts.
func TestRenderChecklistBannerIconRowsEffect(t *testing.T) {
	t.Parallel()

	empty := contracts.SlideDoc{
		Title: "C+B+IR baseline",
		Slides: []contracts.Slide{
			{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}},
			}},
		},
	}
	withNodes := contracts.SlideDoc{
		Title: "C+B+IR effect",
		Slides: []contracts.Slide{
			{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}},
				&contracts.Checklist{
					Items: []contracts.ChecklistItem{
						{Text: contracts.RichText{{Text: "Byte-identical exports"}}, State: contracts.CheckDone, Icon: "check"},
						{Text: contracts.RichText{{Text: "No headless browser"}}, State: contracts.CheckDone},
						{Text: contracts.RichText{{Text: "Soul-driven theming"}}, State: contracts.CheckNeutral, Icon: "dot"},
					},
					Columns:   2,
					GlyphTone: contracts.ColorAccent,
					Fill:      true,
				},
				&contracts.Banner{
					Lead:      contracts.RichText{{Text: "Run it internally"}},
					Body:      contracts.RichText{{Text: "Or sell it externally."}},
					Icon:      "star",
					Fill:      contracts.ColorAccent,
					TextColor: contracts.TextInverse,
					Trailing: []contracts.SlideNode{
						&contracts.Button{
							Label:        "Start free",
							Tone:         contracts.ButtonGhost,
							TrailingIcon: "arrow-right",
						},
					},
				},
				&contracts.IconRows{
					Rows: []contracts.IconRow{
						{Icon: "check", Label: contracts.RichText{{Text: "Chat & Q&A"}}, Tone: contracts.RowPlain},
						{Icon: "star", Label: contracts.RichText{{Text: "Salesforce · Slack"}}, Tone: contracts.RowPill},
					},
					Fill:       true,
					GlyphColor: contracts.ColorAccent,
				},
			}},
		},
	}
	s := soul.DeckardWhite()

	base, baseStats, err := Render(empty, s)
	if err != nil {
		t.Fatalf("baseline Render() error = %v", err)
	}
	assertValidPPTX(t, base)

	got, gotStats, err := Render(withNodes, s)
	if err != nil {
		t.Fatalf("withNodes Render() error = %v", err)
	}
	assertValidPPTX(t, got)
	// Stub-trap guard: the new nodes must emit REAL shapes, strictly more than
	// the baseline slide that uses none of them.
	if gotStats.Shapes <= baseStats.Shapes {
		t.Errorf("Checklist+Banner+IconRows Shapes = %d, want > baseline %d (nodes render no shapes?)",
			gotStats.Shapes, baseStats.Shapes)
	}
	if _, err := pptx.NewFromBytes(got); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}

	// Byte-identical across two Render calls (canonical determinism).
	again, _, err := Render(withNodes, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(got, again) {
		t.Fatal("Render() bytes differ across identical renders")
	}

	// Byte-identical across worker counts.
	defW, _, err := renderWithWorkers(withNodes, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneW, _, err := renderWithWorkers(withNodes, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defW, oneW) {
		t.Fatal("render bytes differ across worker counts")
	}
}
