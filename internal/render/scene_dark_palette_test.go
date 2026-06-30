package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
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
