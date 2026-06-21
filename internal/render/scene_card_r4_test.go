package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeCardR4AllSet asserts that a Card with HeaderFill, StatusDot, and
// Watermark set maps to a scene.Card with all three R4 fields populated (D-054).
func TestMapNodeCardR4AllSet(t *testing.T) {
	t.Parallel()

	node := &contracts.Card{
		Header:     "Pillar",
		Fill:       contracts.ColorSurface,
		HeaderFill: contracts.ColorAccent,
		StatusDot:  contracts.ColorSuccess,
		Watermark:  "01",
		Body:       []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "body"}}}}},
	}
	sn := mapNode(node)
	c, ok := sn.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", sn)
	}

	wantHF := mapColorRole(contracts.ColorAccent)
	if c.HeaderFill == nil {
		t.Error("HeaderFill: got nil, want non-nil")
	} else if *c.HeaderFill != wantHF {
		t.Errorf("HeaderFill: got %v, want %v", *c.HeaderFill, wantHF)
	}

	wantSD := mapColorRole(contracts.ColorSuccess)
	if c.StatusDot == nil {
		t.Error("StatusDot: got nil, want non-nil")
	} else if *c.StatusDot != wantSD {
		t.Errorf("StatusDot: got %v, want %v", *c.StatusDot, wantSD)
	}

	if c.Watermark != "01" {
		t.Errorf("Watermark: got %q, want %q", c.Watermark, "01")
	}
}

// TestMapNodeCardR4EmptyOptionals asserts that a Card with empty (unset)
// HeaderFill and StatusDot maps to nil — the engine's "no band / no dot"
// sentinel — so a plain card renders byte-identical to before D-054.
func TestMapNodeCardR4EmptyOptionals(t *testing.T) {
	t.Parallel()

	node := &contracts.Card{Header: "Plain", Fill: contracts.ColorSurfaceAlt}
	sn := mapNode(node)
	c, ok := sn.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", sn)
	}

	if c.HeaderFill != nil {
		t.Errorf("empty HeaderFill: got %v, want nil (no band)", c.HeaderFill)
	}
	if c.StatusDot != nil {
		t.Errorf("empty StatusDot: got %v, want nil (no dot)", c.StatusDot)
	}
	if c.Watermark != "" {
		t.Errorf("empty Watermark: got %q, want empty string", c.Watermark)
	}
}

// TestMapNodeCardR4RegressionNoSpuriousBand asserts that a Card with no R4 fields
// produces a scene.Card whose R4 fields are all at their zero/nil values — a
// regression guard that no spurious band, dot, or watermark is injected.
func TestMapNodeCardR4RegressionNoSpuriousBand(t *testing.T) {
	t.Parallel()

	plain := &contracts.Card{
		Header:    "H",
		Eyebrow:   "E",
		Fill:      contracts.ColorSurface,
		Elevation: contracts.ElevationRaised,
		Body:      []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "body"}}}}},
	}
	sn := mapNode(plain)
	c, ok := sn.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", sn)
	}

	if c.HeaderFill != nil {
		t.Errorf("regression: plain card got HeaderFill=%v, want nil", c.HeaderFill)
	}
	if c.StatusDot != nil {
		t.Errorf("regression: plain card got StatusDot=%v, want nil", c.StatusDot)
	}
	if c.Watermark != "" {
		t.Errorf("regression: plain card got Watermark=%q, want empty", c.Watermark)
	}
}
