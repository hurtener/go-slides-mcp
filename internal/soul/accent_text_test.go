package soul

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/validate"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestLegibleAccentTextMatchesEngineHelper is a sanity check that the soul's
// wrapper is a thin, deterministic pass-through to the engine's contrast
// mechanism (D-026) at the 4.5:1 body-text ratio (minRatioX10 = 45).
func TestLegibleAccentTextMatchesEngineHelper(t *testing.T) {
	const accent, bg pptx.RGB = "F5DEB3", "FAF7F2"
	got := legibleAccentText(accent, bg)
	want := scene.LegibleTextOn(accent, bg, 45)
	if got != want {
		t.Fatalf("legibleAccentText(%q, %q) = %q, want %q", accent, bg, got, want)
	}
}

// TestBootstrapDerivesLegibleAccentTextOnLightCanvas is the R8.6 light-side
// acceptance: an accent surface override that is illegible against the
// resolved canvas is replaced with a derivation that clears 4.5:1.
func TestBootstrapDerivesLegibleAccentTextOnLightCanvas(t *testing.T) {
	const paleAccent = "F5DEB3" // wheat, near-white — illegible on cream
	const cream = "FAF7F2"      // DeckardWhite's default canvas

	if ratio, ok := validate.ContrastRatio(paleAccent, cream); !ok || ratio >= 4.5 {
		t.Fatalf("precondition: contrast(%q, %q) = %v (ok=%v), want < 4.5", paleAccent, cream, ratio, ok)
	}

	s, err := Bootstrap(BootstrapParams{
		Name:    "Pale Brand",
		Palette: &Palette{Surfaces: map[string]string{"accent": paleAccent}},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}

	textAccent := s.Theme.Colors.Text[pptx.TextAccent]
	ratio, ok := validate.ContrastRatio(textAccent, s.Theme.Colors.Surfaces[pptx.ColorCanvas])
	if !ok {
		t.Fatalf("derived text accent %q did not parse", textAccent)
	}
	if ratio < 4.5 {
		t.Fatalf("derived text accent %q contrast = %v, want >= 4.5", textAccent, ratio)
	}
}

// TestBootstrapNameOnlyLeavesTextAccentByteIdentical guards against
// derivation running when the caller never overrode the accent surface: a
// name-only bootstrap must keep DeckardWhite's hand-tuned text accent.
func TestBootstrapNameOnlyLeavesTextAccentByteIdentical(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "x"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if got := s.Theme.Colors.Text[pptx.TextAccent]; got != "2B7A73" {
		t.Fatalf("Text[TextAccent] = %q, want byte-identical 2B7A73", got)
	}
}

// TestBootstrapExplicitPaletteTextWinsOverDerivation guards the override
// precedence: an explicit Palette.Text["accent"] must survive the final
// derivation pass untouched, even though the accent surface was also
// overridden (which would otherwise trigger derivation).
func TestBootstrapExplicitPaletteTextWinsOverDerivation(t *testing.T) {
	const explicit = "112233"
	s, err := Bootstrap(BootstrapParams{
		Name:   "Acme",
		Accent: "F5DEB3",
		Palette: &Palette{
			Text: map[string]string{"accent": explicit},
		},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if got := s.Theme.Colors.Text[pptx.TextAccent]; got != explicit {
		t.Fatalf("Text[TextAccent] = %q, want explicit override %q untouched", got, explicit)
	}
}

// TestBootstrapDerivesLegibleDarkAccentText is the R8.6 dark-side acceptance:
// a soul with a dark palette and no explicit dark text-accent override gets a
// derived dark accent text that clears 4.5:1 against the dark canvas and
// differs from the light accent text (each is legible only on its own bg).
func TestBootstrapDerivesLegibleDarkAccentText(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{
		Name:    "Navy Brand",
		Palette: &Palette{Surfaces: map[string]string{"accent": "F5DEB3"}},
		DarkPalette: &DarkPalette{
			DarkSurfaces: map[string]string{"canvas": "0A1622", "accent": "1B3A5C"},
		},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.Theme.DarkColors == nil {
		t.Fatal("DarkColors = nil, want non-nil")
	}

	lightTextAccent := s.Theme.Colors.Text[pptx.TextAccent]
	darkTextAccent := s.Theme.DarkColors.Text[pptx.TextAccent]

	ratio, ok := validate.ContrastRatio(darkTextAccent, s.Theme.DarkColors.Surfaces[pptx.ColorCanvas])
	if !ok {
		t.Fatalf("dark text accent %q did not parse", darkTextAccent)
	}
	if ratio < 4.5 {
		t.Fatalf("dark text accent %q contrast = %v, want >= 4.5", darkTextAccent, ratio)
	}
	if darkTextAccent == lightTextAccent {
		t.Fatalf("dark text accent %q must differ from light text accent %q", darkTextAccent, lightTextAccent)
	}
}

// TestBootstrapNoDarkPaletteInventsNoDarkText is the byte-identity guard: a
// soul with NO dark palette must leave Theme.DarkColors nil — no dark text is
// invented when there is no dark side to derive it for.
func TestBootstrapNoDarkPaletteInventsNoDarkText(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{
		Name:    "No Dark Brand",
		Palette: &Palette{Surfaces: map[string]string{"accent": "F5DEB3"}},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.Theme.DarkColors != nil {
		t.Fatalf("DarkColors = %+v, want nil", s.Theme.DarkColors)
	}
}
