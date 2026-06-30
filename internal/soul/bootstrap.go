package soul

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/hurtener/pptx-go/pptx"
)

var nonSlugRunes = regexp.MustCompile(`[^a-z0-9]+`)

// BootstrapParams seeds a complete soul from a small set of structured tokens.
type BootstrapParams struct {
	// Name is the required soul name and the source for the slugified ID.
	Name string
	// Description is an optional one-line summary.
	Description string
	// Accent overrides the accent surface color as a 6-digit hex string without '#'.
	Accent string
	// AccentAlt overrides the alternate accent surface and text colors.
	AccentAlt string
	// AccentWarm overrides the warm accent and warning surface colors.
	AccentWarm string
	// HeadingFont overrides the display and H1-H3 families.
	HeadingFont string
	// BodyFont overrides the H4-H5 and body families.
	BodyFont string
	// MonoFont overrides the mono and code families.
	MonoFont string
	// Palette is an optional complete color palette covering every surface,
	// text, and extension token in one call.
	Palette *Palette
	// DarkPalette is an optional soul-driven VariantDark color override set (R8.3).
	DarkPalette *DarkPalette
}

// DarkPalette is an optional soul-driven VariantDark color override set (R8.3).
// Each map is keyed by the SAME token names Refine validates; the values become
// pptx.Theme.DarkColors, which the engine overlays over its pinned neutral-gray
// dark default so a brand renders its own deep dark side (e.g. navy). Unset =>
// the engine keeps its pinned gray, byte-identical.
type DarkPalette struct {
	DarkSurfaces map[string]string // surface-role token -> 6-digit hex
	DarkText     map[string]string // text-role token -> 6-digit hex
}

// Palette is a complete optional color palette for Bootstrap. Each map is keyed
// by the SAME token names Refine validates (surfaceRole/textRole); an Extensions
// map carries the non-native tokens (border/borderStrong/accentSoft). Unset keys
// inherit DeckardWhite byte-for-byte; an unknown key is a typed error.
type Palette struct {
	Surfaces   map[string]string // surface-role token -> 6-digit hex
	Text       map[string]string // text-role token -> 6-digit hex
	Extensions map[string]string // extension token -> 6-digit hex
}

// Bootstrap seeds a complete soul from p, inheriting every unset token from
// DeckardWhite().
func Bootstrap(p BootstrapParams) (*Soul, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, fmt.Errorf("soul: bootstrap requires a non-empty name")
	}

	s := DeckardWhite().Clone()
	id := slugify(name)
	if id == "" {
		return nil, fmt.Errorf("soul: bootstrap produced an empty id from name %q", p.Name)
	}

	if p.Description != "" {
		s.Description = p.Description
	}
	if p.Accent != "" {
		accent, err := parseHexColor(p.Accent)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accent: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccent] = accent
		s.Theme.Colors.Text[pptx.TextAccent] = derivedTextAccent(accent)
	}
	if p.AccentAlt != "" {
		accentAlt, err := parseHexColor(p.AccentAlt)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accentAlt: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccentAlt] = accentAlt
		s.Theme.Colors.Text[pptx.TextAccentAlt] = derivedTextAccent(accentAlt)
	}
	if p.AccentWarm != "" {
		accentWarm, err := parseHexColor(p.AccentWarm)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accentWarm: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccentWarm] = accentWarm
		s.Theme.Colors.Surfaces[pptx.ColorWarning] = accentWarm
	}
	if p.HeadingFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.HeadingFont, fs.Size, fs.Weight)
		}
		s.Theme.HeadingFont = p.HeadingFont
	}
	if p.BodyFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeH4, pptx.TypeH5, pptx.TypeBody, pptx.TypeBodySmall, pptx.TypeCaption} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.BodyFont, fs.Size, fs.Weight)
		}
		s.Theme.BodyFont = p.BodyFont
	}
	if p.MonoFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeMono, pptx.TypeCode} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.MonoFont, fs.Size, fs.Weight)
		}
	}

	if p.Palette != nil {
		if err := applyPalette(s, p.Palette); err != nil {
			return nil, err
		}
	}
	if p.DarkPalette != nil {
		if err := applyDarkPalette(s, p.DarkPalette); err != nil {
			return nil, err
		}
	}

	s.ID = id
	s.Name = name
	s.Status = "ready"
	if name != "Deckard White" {
		s.StyleGuide.NorthStar = ""
		s.StyleGuide.Do = nil
		s.StyleGuide.Dont = nil
	}

	return s, nil
}

// applyPalette writes every supplied palette token onto s.Theme/s.Extensions,
// in sorted key order per map, for stable error reporting.
func applyPalette(s *Soul, p *Palette) error {
	for _, token := range sortedPaletteKeys(p.Surfaces) {
		role, ok := surfaceRole(token)
		if !ok {
			return fmt.Errorf("soul: bootstrap palette: unknown surface token %q", token)
		}
		rgb, err := parseHexColor(p.Surfaces[token])
		if err != nil {
			return fmt.Errorf("soul: bootstrap palette surface %q: %w", token, err)
		}
		s.Theme.Colors.Surfaces[role] = rgb
	}
	for _, token := range sortedPaletteKeys(p.Text) {
		role, ok := textRole(token)
		if !ok {
			return fmt.Errorf("soul: bootstrap palette: unknown text token %q", token)
		}
		rgb, err := parseHexColor(p.Text[token])
		if err != nil {
			return fmt.Errorf("soul: bootstrap palette text %q: %w", token, err)
		}
		s.Theme.Colors.Text[role] = rgb
	}
	for _, token := range sortedPaletteKeys(p.Extensions) {
		if token == "" {
			return fmt.Errorf("soul: bootstrap palette: unknown extension token %q", token)
		}
		rgb, err := parseHexColor(p.Extensions[token])
		if err != nil {
			return fmt.Errorf("soul: bootstrap palette extension %q: %w", token, err)
		}
		if s.Extensions == nil {
			s.Extensions = make(map[string]string)
		}
		s.Extensions[token] = string(rgb)
	}
	return nil
}

// applyDarkPalette writes every supplied dark-palette token into a new
// pptx.DarkPalette and assigns it to s.Theme.DarkColors. An all-empty p
// leaves s.Theme.DarkColors nil, byte-identical to no dark palette at all.
func applyDarkPalette(s *Soul, p *DarkPalette) error {
	dp := &pptx.DarkPalette{
		Surfaces: make(map[pptx.ColorRole]pptx.RGB, len(p.DarkSurfaces)),
		Text:     make(map[pptx.TextColorRole]pptx.RGB, len(p.DarkText)),
	}
	for _, token := range sortedPaletteKeys(p.DarkSurfaces) {
		role, ok := surfaceRole(token)
		if !ok {
			return fmt.Errorf("soul: bootstrap dark palette: unknown surface token %q", token)
		}
		rgb, err := parseHexColor(p.DarkSurfaces[token])
		if err != nil {
			return fmt.Errorf("soul: bootstrap dark palette surface %q: %w", token, err)
		}
		dp.Surfaces[role] = rgb
	}
	for _, token := range sortedPaletteKeys(p.DarkText) {
		role, ok := textRole(token)
		if !ok {
			return fmt.Errorf("soul: bootstrap dark palette: unknown text token %q", token)
		}
		rgb, err := parseHexColor(p.DarkText[token])
		if err != nil {
			return fmt.Errorf("soul: bootstrap dark palette text %q: %w", token, err)
		}
		dp.Text[role] = rgb
	}
	if len(dp.Surfaces) > 0 || len(dp.Text) > 0 {
		s.Theme.DarkColors = dp
	}
	return nil
}

func sortedPaletteKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func slugify(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteByte('-')
	}
	return strings.Trim(nonSlugRunes.ReplaceAllString(b.String(), "-"), "-")
}
