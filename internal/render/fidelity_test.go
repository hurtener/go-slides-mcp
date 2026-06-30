package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

// brandFidelitySoul bootstraps a representative non-default brand soul: a
// complete light palette (distinctive canvas/surface/surfaceAlt + accent pair
// + a couple text roles), a navy dark palette overriding canvas/surface/
// surfaceAlt/primaryText/accent, and a named gradient — the R8.10 capstone
// fixture proving R8-A/B/D/E all reach render.
func brandFidelitySoul(t *testing.T) *soul.Soul {
	t.Helper()
	s, err := soul.Bootstrap(soul.BootstrapParams{
		Name: "Fidelity Brand",
		Palette: &soul.Palette{
			Surfaces: map[string]string{
				"canvas":     "F4F1EA",
				"surface":    "FFFFFF",
				"surfaceAlt": "E8E2D4",
				"accent":     "0D9488",
				"accentAlt":  "EA580C",
			},
			Text: map[string]string{
				"primary": "1A1A1A",
				"muted":   "5C5C5C",
			},
		},
		DarkPalette: &soul.DarkPalette{
			DarkSurfaces: map[string]string{
				"canvas":     "0A0E1A",
				"surface":    "14182B",
				"surfaceAlt": "1C2238",
				"accent":     "5EEAD4",
			},
			DarkText: map[string]string{
				"primary": "F4F6FF",
			},
		},
		Gradients: []soul.GradientSpec{
			{
				Name: "fidelityHero",
				Stops: []soul.GradientStop{
					{Pos: 0, ColorHex: "1E293B"},
					{Pos: 1, ColorHex: "0A0E1A"},
				},
				Radial: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}
	return s
}

// fidelityDoc returns a doc with one light and one dark slide, exercising
// both variants through SoulColorFidelity.
func fidelityDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Fidelity Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "light",
				Layout: contracts.LayoutFullBleed,
				Nodes:  []contracts.SlideNode{&contracts.Heading{Level: 1, Text: rt("Light slide")}},
			},
			{
				ID:      "dark",
				Layout:  contracts.LayoutFullBleed,
				Variant: contracts.VariantDark,
				Nodes:   []contracts.SlideNode{&contracts.Heading{Level: 1, Text: rt("Dark slide")}},
			},
		},
	}
}

// TestSoulColorFidelity_BrandSoulPasses is the R8.10 acceptance and the
// capstone proving R8-A/B/D/E's soul tokens all reach the rendered bytes: a
// non-default brand soul with a complete light palette, a navy dark palette,
// and a gradient renders with ZERO mismatches across both variants.
func TestSoulColorFidelity_BrandSoulPasses(t *testing.T) {
	t.Parallel()

	s := brandFidelitySoul(t)
	mismatches, err := SoulColorFidelity(fidelityDoc(), s)
	if err != nil {
		t.Fatalf("SoulColorFidelity() error = %v", err)
	}
	if len(mismatches) != 0 {
		t.Errorf("SoulColorFidelity() mismatches = %+v, want none", mismatches)
	}
}

// TestSoulColorFidelity_Deterministic asserts SoulColorFidelity is stable
// across repeated calls on the same doc+soul (render determinism, D-spec).
func TestSoulColorFidelity_Deterministic(t *testing.T) {
	t.Parallel()

	s := brandFidelitySoul(t)
	doc := fidelityDoc()
	first, err := SoulColorFidelity(doc, s)
	if err != nil {
		t.Fatalf("SoulColorFidelity() first call error = %v", err)
	}
	second, err := SoulColorFidelity(doc, s)
	if err != nil {
		t.Fatalf("SoulColorFidelity() second call error = %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("mismatch count differs across calls: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("mismatch[%d] differs across calls: %+v vs %+v", i, first[i], second[i])
		}
	}
}

// TestCompareSlideColors_CatchesDrift is the negative guard (R8.10
// acceptance): comparing a brand soul's RENDERED colors against a DIFFERENT
// soul's DECLARED tokens must surface the mismatch, naming the differing role
// with the correct Want/Got.
func TestCompareSlideColors_CatchesDrift(t *testing.T) {
	t.Parallel()

	brand := brandFidelitySoul(t)
	_, stats, err := Render(fidelityDoc(), brand)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.Colors) != 2 {
		t.Fatalf("stats.Colors len = %d, want 2", len(stats.Colors))
	}
	lightColors := stats.Colors[0]
	if lightColors.SlideID != "light" {
		t.Fatalf("stats.Colors[0].SlideID = %q, want %q", lightColors.SlideID, "light")
	}

	other, err := soul.Bootstrap(soul.BootstrapParams{Name: "Other Brand"}) // DeckardWhite-derived, no overrides
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}

	mismatches := compareSlideColors(lightColors, other, false)
	if len(mismatches) == 0 {
		t.Fatal("compareSlideColors() found no mismatches, want the drifted canvas/accent to be caught")
	}

	var canvasMismatch *ColorMismatch
	for i := range mismatches {
		if mismatches[i].Role == "canvas" {
			canvasMismatch = &mismatches[i]
			break
		}
	}
	if canvasMismatch == nil {
		t.Fatalf("compareSlideColors() mismatches = %+v, want a canvas mismatch", mismatches)
	}
	wantCanvas := string(other.Theme.ResolveColor(pptx.ColorCanvas))
	if canvasMismatch.Want != wantCanvas {
		t.Errorf("canvas mismatch Want = %q, want the other soul's declared canvas %q", canvasMismatch.Want, wantCanvas)
	}
	if canvasMismatch.Got != string(lightColors.Canvas) {
		t.Errorf("canvas mismatch Got = %q, want the rendered canvas %q", canvasMismatch.Got, lightColors.Canvas)
	}
	if canvasMismatch.Variant != "light" {
		t.Errorf("canvas mismatch Variant = %q, want %q", canvasMismatch.Variant, "light")
	}
}

// TestCompareSlideColors_DarkNoOverridesSkipsAll asserts a soul with no
// Theme.DarkColors contributes zero dark-variant comparisons (those roles are
// engine-pinned, not soul-owned) — compareSlideColors must not false-positive.
func TestCompareSlideColors_DarkNoOverridesSkipsAll(t *testing.T) {
	t.Parallel()

	s, err := soul.Bootstrap(soul.BootstrapParams{Name: "No Dark Overrides"})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}
	if s.Theme.DarkColors != nil {
		t.Fatalf("precondition: DarkColors = %+v, want nil", s.Theme.DarkColors)
	}

	mismatches, err := SoulColorFidelity(fidelityDoc(), s)
	if err != nil {
		t.Fatalf("SoulColorFidelity() error = %v", err)
	}
	for _, m := range mismatches {
		if m.Variant == "dark" {
			t.Errorf("dark mismatch reported for an undeclared role: %+v, want none", m)
		}
	}
}
