package soul

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestBootstrapInheritsDeckardWhiteWithNameOnly(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "x"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.ID != "x" {
		t.Fatalf("ID = %q, want x", s.ID)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "3B9C94" {
		t.Fatalf("accent = %q, want 3B9C94", got)
	}
	if got := s.Theme.ResolveTextColor(pptx.TextAccent); got != "2B7A73" {
		t.Fatalf("text accent = %q, want 2B7A73", got)
	}
	if s.StyleGuide.NorthStar != "" {
		t.Fatalf("style guide should be cleared for renamed soul, got %q", s.StyleGuide.NorthStar)
	}
	if s.Status != "ready" {
		t.Fatalf("status = %q, want ready", s.Status)
	}
}

func TestBootstrapAccentOverride(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "Acme", Accent: "DB2777"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "DB2777" {
		t.Fatalf("accent = %q, want DB2777", got)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccentAlt); got != "2B7A73" {
		t.Fatalf("accentAlt = %q, want inherited 2B7A73", got)
	}
}

func TestBootstrapRejectsEmptyName(t *testing.T) {
	if _, err := Bootstrap(BootstrapParams{}); err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestBootstrapSlugifiesID(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "Acme Labs 2.0!"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.ID != "acme-labs-2-0" {
		t.Fatalf("ID = %q, want acme-labs-2-0", s.ID)
	}
}

func TestBootstrapNilPaletteMatchesDeckardWhite(t *testing.T) {
	want := DeckardWhite()
	for _, p := range []*Palette{nil, {}} {
		s, err := Bootstrap(BootstrapParams{Name: "x", Palette: p})
		if err != nil {
			fatalBootstrap(t, err)
		}
		for token := range surfaceTokens {
			role, _ := surfaceRole(token)
			if got, exp := s.Theme.Colors.Surfaces[role], want.Theme.Colors.Surfaces[role]; got != exp {
				t.Fatalf("surface %q = %q, want %q (DeckardWhite)", token, got, exp)
			}
		}
		for token := range textTokens {
			role, _ := textRole(token)
			if got, exp := s.Theme.Colors.Text[role], want.Theme.Colors.Text[role]; got != exp {
				t.Fatalf("text %q = %q, want %q (DeckardWhite)", token, got, exp)
			}
		}
		if len(s.Extensions) != len(want.Extensions) {
			t.Fatalf("extensions = %v, want %v", s.Extensions, want.Extensions)
		}
		for k, v := range want.Extensions {
			if s.Extensions[k] != v {
				t.Fatalf("extension %q = %q, want %q", k, s.Extensions[k], v)
			}
		}
	}
}

func TestBootstrapFullPaletteSetsEverySuppliedToken(t *testing.T) {
	palette := &Palette{
		Surfaces: map[string]string{
			"canvas": "0B1B2B", "surface": "122838", "surfaceAlt": "1A3548",
			"accent": "2FA37E", "accentAlt": "7A4FE0", "accentWarm": "E08A2B",
			"success": "3FA66B", "warning": "E0A82B", "error": "D14F4F", "info": "3E8FD0",
		},
		Text: map[string]string{
			"primary": "F5F1E8", "secondary": "D8D2C2", "tertiary": "A9A28E",
			"inverse": "0B1B2B", "muted": "7C7868", "accent": "E8F7F0",
			"accentAlt": "F2ECFE", "success": "E6F6EC", "warning": "FBF1DD", "error": "FBE7E7",
		},
		Extensions: map[string]string{
			"border": "2A4458", "borderStrong": "3E5C73", "accentSoft": "1E4538",
		},
	}
	s, err := Bootstrap(BootstrapParams{Name: "Full Brand", Palette: palette})
	if err != nil {
		fatalBootstrap(t, err)
	}
	for token, hex := range palette.Surfaces {
		role, _ := surfaceRole(token)
		if got := string(s.Theme.Colors.Surfaces[role]); got != hex {
			t.Fatalf("surface %q = %q, want %q", token, got, hex)
		}
	}
	for token, hex := range palette.Text {
		role, _ := textRole(token)
		if got := string(s.Theme.Colors.Text[role]); got != hex {
			t.Fatalf("text %q = %q, want %q", token, got, hex)
		}
	}
	for token, hex := range palette.Extensions {
		if got := s.Extensions[token]; got != hex {
			t.Fatalf("extension %q = %q, want %q", token, got, hex)
		}
	}
}

func TestBootstrapPaletteRejectsUnknownTokens(t *testing.T) {
	cases := []struct {
		name    string
		palette *Palette
	}{
		{"unknown surface", &Palette{Surfaces: map[string]string{"notAToken": "112233"}}},
		{"unknown text", &Palette{Text: map[string]string{"notAToken": "112233"}}},
		{"empty extension key", &Palette{Extensions: map[string]string{"": "112233"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Bootstrap(BootstrapParams{Name: "x", Palette: tc.palette}); err == nil {
				t.Fatal("expected a typed error, got nil")
			}
		})
	}
}

var surfaceTokens = map[string]struct{}{
	"canvas": {}, "surface": {}, "surfaceAlt": {}, "accent": {}, "accentAlt": {},
	"accentWarm": {}, "success": {}, "warning": {}, "error": {}, "info": {},
}

var textTokens = map[string]struct{}{
	"primary": {}, "secondary": {}, "tertiary": {}, "inverse": {}, "muted": {},
	"accent": {}, "accentAlt": {}, "success": {}, "warning": {}, "error": {},
}

func TestBootstrapDarkPaletteSetsDarkColors(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{
		Name: "Acme",
		DarkPalette: &DarkPalette{
			DarkSurfaces: map[string]string{"canvas": "0A1622"},
			DarkText:     map[string]string{"primary": "F5F1E8"},
		},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.Theme.DarkColors == nil {
		t.Fatal("DarkColors = nil, want non-nil")
	}
	if got := s.Theme.DarkColors.Surfaces[pptx.ColorCanvas]; got != "0A1622" {
		t.Fatalf("DarkColors.Surfaces[ColorCanvas] = %q, want 0A1622", got)
	}
	if got := s.Theme.DarkColors.Text[pptx.TextPrimary]; got != "F5F1E8" {
		t.Fatalf("DarkColors.Text[TextPrimary] = %q, want F5F1E8", got)
	}
}

func TestBootstrapNilOrEmptyDarkPaletteLeavesDarkColorsNil(t *testing.T) {
	for _, p := range []*DarkPalette{nil, {}} {
		s, err := Bootstrap(BootstrapParams{Name: "x", DarkPalette: p})
		if err != nil {
			fatalBootstrap(t, err)
		}
		if s.Theme.DarkColors != nil {
			t.Fatalf("DarkColors = %+v, want nil for %+v", s.Theme.DarkColors, p)
		}
	}
}

func fatalBootstrap(t *testing.T, err error) {
	t.Helper()
	t.Fatal(err)
}
