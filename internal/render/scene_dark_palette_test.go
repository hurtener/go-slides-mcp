package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

// darkSlideDoc returns a minimal doc with one VariantDark slide, used to
// exercise the soul-driven dark palette (R8.3).
func darkSlideDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Dark Palette Coverage",
		Slides: []contracts.Slide{
			{
				ID:      "dark",
				Layout:  contracts.LayoutFullBleed,
				Variant: contracts.VariantDark,
				Nodes:   []contracts.SlideNode{&contracts.Heading{Level: 1, Text: rt("Dark slide")}},
			},
		},
	}
}

// TestRenderSoulDarkPaletteOverridesCanvas is the R8.3 acceptance: a soul
// bootstrapped with a navy dark palette renders its VariantDark slide's
// resolved canvas to the supplied brand hex, not the engine's pinned gray.
func TestRenderSoulDarkPaletteOverridesCanvas(t *testing.T) {
	t.Parallel()

	const navyCanvas = "0A1622"
	s, err := soul.Bootstrap(soul.BootstrapParams{
		Name: "Navy Brand",
		DarkPalette: &soul.DarkPalette{
			DarkSurfaces: map[string]string{"canvas": navyCanvas},
		},
	})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}

	_, stats, err := Render(darkSlideDoc(), s)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.Colors) != 1 {
		t.Fatalf("stats.Colors len = %d, want 1", len(stats.Colors))
	}
	if got := string(stats.Colors[0].Canvas); got != navyCanvas {
		t.Errorf("resolved dark canvas = %q, want %q", got, navyCanvas)
	}
}

// lightSlideDoc returns a minimal doc with one default-variant (light) slide.
func lightSlideDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Light Accent Text Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "light",
				Layout: contracts.LayoutFullBleed,
				Nodes:  []contracts.SlideNode{&contracts.Heading{Level: 1, Text: rt("Light slide")}},
			},
		},
	}
}

// TestRenderSoulDerivedAccentTextReachesRender is the R8.6 acceptance: a soul
// bootstrapped with an accent surface override resolves its render-time
// TextAccent to the soul's WCAG-derived legible value, not a raw scale.
func TestRenderSoulDerivedAccentTextReachesRender(t *testing.T) {
	t.Parallel()

	const paleAccent = "F5DEB3"
	s, err := soul.Bootstrap(soul.BootstrapParams{
		Name:    "Pale Brand",
		Palette: &soul.Palette{Surfaces: map[string]string{"accent": paleAccent}},
	})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}
	wantTextAccent := s.Theme.Colors.Text[pptx.TextAccent]

	_, stats, err := Render(lightSlideDoc(), s)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.Colors) != 1 {
		t.Fatalf("stats.Colors len = %d, want 1", len(stats.Colors))
	}
	if got := stats.Colors[0].TextAccent; got != wantTextAccent {
		t.Errorf("rendered TextAccent = %q, want soul-derived %q", got, wantTextAccent)
	}
}

// TestRenderNoDarkPaletteKeepsPinnedGray is the byte-identity guard: a soul
// with NO dark palette renders its VariantDark slide's canvas to the engine's
// pinned neutral-gray default (111827), unchanged from before R8.3.
func TestRenderNoDarkPaletteKeepsPinnedGray(t *testing.T) {
	t.Parallel()

	const pinnedGrayCanvas = "111827"
	s, err := soul.Bootstrap(soul.BootstrapParams{Name: "No Dark Palette"})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}
	if s.Theme.DarkColors != nil {
		t.Fatalf("precondition: DarkColors = %+v, want nil", s.Theme.DarkColors)
	}

	_, stats, err := Render(darkSlideDoc(), s)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.Colors) != 1 {
		t.Fatalf("stats.Colors len = %d, want 1", len(stats.Colors))
	}
	if got := string(stats.Colors[0].Canvas); got != pinnedGrayCanvas {
		t.Errorf("resolved dark canvas = %q, want pinned gray %q", got, pinnedGrayCanvas)
	}
}
