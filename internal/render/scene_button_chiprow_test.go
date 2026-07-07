package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeButtonAllFields asserts a fully-populated Button maps to a
// scene.Button with every field set (R12.1, D-094).
func TestMapNodeButtonAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Button{
		Label:        "Talk to the team",
		Tone:         contracts.ButtonAccentAlt,
		Size:         contracts.ButtonSizeLG,
		LeadingIcon:  "check",
		TrailingIcon: "arrow-right",
		Align:        contracts.HAlignCenter,
	}
	sn := mapNode(node)
	b, ok := sn.(scene.Button)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Button", sn)
	}
	if b.Label != "Talk to the team" {
		t.Errorf("Label: got %q, want %q", b.Label, "Talk to the team")
	}
	if b.Tone != scene.ButtonAccentAlt {
		t.Errorf("Tone: got %v, want scene.ButtonAccentAlt", b.Tone)
	}
	if b.Size != scene.ButtonLG {
		t.Errorf("Size: got %v, want scene.ButtonLG", b.Size)
	}
	if b.LeadingIcon != "check" {
		t.Errorf("LeadingIcon: got %q, want %q", b.LeadingIcon, "check")
	}
	if b.TrailingIcon != "arrow-right" {
		t.Errorf("TrailingIcon: got %q, want %q", b.TrailingIcon, "arrow-right")
	}
	if b.Align != scene.HAlignCenter {
		t.Errorf("Align: got %v, want scene.HAlignCenter", b.Align)
	}
}

// TestMapNodeButtonZeroDefaults asserts that empty Tone/Size/Align map to the
// engine's zero values (ButtonPrimary / ButtonMD / HAlignLeft) — the
// additive path, byte-identical to the engine's default.
func TestMapNodeButtonZeroDefaults(t *testing.T) {
	t.Parallel()

	sn := mapNode(&contracts.Button{Label: "Go"})
	b, ok := sn.(scene.Button)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Button", sn)
	}
	if b.Tone != scene.ButtonPrimary {
		t.Errorf("Tone: got %v, want scene.ButtonPrimary (zero)", b.Tone)
	}
	if b.Size != scene.ButtonMD {
		t.Errorf("Size: got %v, want scene.ButtonMD (zero)", b.Size)
	}
	if b.Align != scene.HAlignLeft {
		t.Errorf("Align: got %v, want scene.HAlignLeft (zero)", b.Align)
	}
}

// TestMapNodeChipRowAllFields asserts a fully-populated ChipRow maps to a
// scene.ChipRow with every field set (R12.5, D-096), incl. per-chip Tone +
// Color mapping through their existing mappers.
func TestMapNodeChipRowAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.ChipRow{
		Label: "CAPABILITIES",
		Chips: []contracts.ChipSpec{
			{Label: "Operate", Tone: contracts.ChipTint, Color: contracts.ColorAccent},
			{Label: "Execute", Tone: contracts.ChipSolid, Color: contracts.ColorAccent, Icon: "check"},
			{Label: "Build", Tone: contracts.ChipOutline, Color: contracts.ColorAccentAlt, Icon: "star"},
		},
		Wrap:  true,
		Align: contracts.HAlignCenter,
	}
	sn := mapNode(node)
	c, ok := sn.(scene.ChipRow)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.ChipRow", sn)
	}
	if c.Label != "CAPABILITIES" {
		t.Errorf("Label: got %q, want %q", c.Label, "CAPABILITIES")
	}
	if !c.Wrap {
		t.Error("Wrap: got false, want true")
	}
	if c.Align != scene.HAlignCenter {
		t.Errorf("Align: got %v, want scene.HAlignCenter", c.Align)
	}
	if len(c.Chips) != 3 {
		t.Fatalf("len(Chips): got %d, want 3", len(c.Chips))
	}
	if c.Chips[0].Tone != scene.ChipTint {
		t.Errorf("Chips[0].Tone: got %v, want scene.ChipTint", c.Chips[0].Tone)
	}
	if c.Chips[1].Tone != scene.ChipSolid {
		t.Errorf("Chips[1].Tone: got %v, want scene.ChipSolid", c.Chips[1].Tone)
	}
	if c.Chips[2].Tone != scene.ChipOutline {
		t.Errorf("Chips[2].Tone: got %v, want scene.ChipOutline", c.Chips[2].Tone)
	}
	if c.Chips[1].Icon != "check" {
		t.Errorf("Chips[1].Icon: got %q, want %q", c.Chips[1].Icon, "check")
	}
	if c.Chips[2].Color != scene.ColorAccentAlt {
		t.Errorf("Chips[2].Color: got %v, want scene.ColorAccentAlt", c.Chips[2].Color)
	}
}

// TestRenderButtonChipRowEffect verifies the two new nodes render REAL
// shapes (strictly more than an empty slide), produce a valid PPTX, and are
// byte-identical across repeated renders and across worker counts (the
// determinism contract, CLAUDE §5).
func TestRenderButtonChipRowEffect(t *testing.T) {
	t.Parallel()

	empty := contracts.SlideDoc{
		Title: "Button+ChipRow baseline",
		Slides: []contracts.Slide{
			{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}},
			}},
		},
	}
	withNodes := contracts.SlideDoc{
		Title: "Button+ChipRow effect",
		Slides: []contracts.Slide{
			{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}},
				&contracts.Button{
					Label:        "Talk to the team",
					Tone:         contracts.ButtonPrimary,
					Size:         contracts.ButtonSizeLG,
					TrailingIcon: "arrow-right",
				},
				&contracts.ChipRow{
					Label: "TAGS",
					Chips: []contracts.ChipSpec{
						{Label: "Finance", Tone: contracts.ChipTint, Color: contracts.ColorAccent},
						{Label: "HR", Tone: contracts.ChipSolid, Color: contracts.ColorAccent, Icon: "check"},
						{Label: "Sales", Tone: contracts.ChipOutline, Color: contracts.ColorAccentAlt},
					},
					Wrap: true,
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
	// The nodes must render REAL shapes — strictly more than the baseline slide.
	// Guards against the stub trap (a node that maps to a no-op).
	if gotStats.Shapes <= baseStats.Shapes {
		t.Errorf("Button+ChipRow Shapes = %d, want > baseline %d (nodes render no shapes?)",
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

	// Byte-identical across worker counts (race-free determinism).
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
