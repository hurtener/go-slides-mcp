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
	// Gradients is an optional set of named brand gradients (R8.5), each
	// registered on the soul's theme under its Name and requested at render
	// time by a slide Background's GradientName. Unset/nil leaves
	// s.Theme.Gradients nil, byte-identical to today.
	Gradients []GradientSpec
}

// GradientSpec is a named brand gradient (R8.5) for Bootstrap. Stops are an
// ordered 2..8 list; each stop's color is either a pinned hex (ColorHex,
// variant-independent) or a surface role (ColorRole, follows the variant) —
// exactly one. Radial picks a radial wash; otherwise Angle (deg, clockwise
// from the positive x-axis) drives a linear gradient.
type GradientSpec struct {
	// Name is the gradient's stable identifier, requested by a slide
	// Background's GradientName. Must be non-empty and unique within one
	// Bootstrap call.
	Name string
	// Stops is the ordered 2..8 color-stop list, each Pos strictly ascending
	// in [0,1].
	Stops []GradientStop
	// Angle is the linear gradient angle in degrees clockwise from the
	// positive x-axis. Ignored when Radial is true.
	Angle int
	// Radial selects a radial wash from the centre outward instead of a
	// linear gradient.
	Radial bool
}

// GradientStop is one color stop in a GradientSpec. Exactly one of ColorHex
// (a pinned 6-digit hex, variant-independent) or ColorRole (a surface-role
// token that follows the active theme variant) must be set.
type GradientStop struct {
	// Pos is the stop position along the gradient axis, in [0,1].
	Pos float64
	// ColorHex pins the stop to an exact 6-digit hex color (no '#'),
	// unaffected by a light/dark variant swap. Mutually exclusive with
	// ColorRole.
	ColorHex string
	// ColorRole names a surface-role token (the same names surfaceRole
	// validates) whose resolved color follows the active theme variant.
	// Mutually exclusive with ColorHex.
	ColorRole string
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
	}
	if p.AccentAlt != "" {
		accentAlt, err := parseHexColor(p.AccentAlt)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accentAlt: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccentAlt] = accentAlt
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
	if err := applyGradients(s, p.Gradients); err != nil {
		return nil, err
	}

	deriveAccentText(s, p)
	applyTypographyDefaults(s.Theme)

	// R13.1/R13.12: seed the paper tint, register the default cover/dark
	// named gradients, and build the soul's default decor policy — all
	// derived from the now-fully-built theme, so a bootstrapped soul is
	// tastefully decorated out of the box without the caller hand-placing
	// ornaments.
	s.Theme.Colors.Surfaces[pptx.ColorPaper] = pptx.RGB(paperTint(
		string(s.Theme.ResolveColor(pptx.ColorCanvas)),
		string(s.Theme.ResolveColor(pptx.ColorSurfaceAlt)),
	))
	registerDecorGradients(s)
	s.Decor = DefaultDecorPolicy(s.Theme)

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

// applyGradients validates and registers each named brand gradient (R8.5)
// onto s.Theme.Gradients. An empty/nil specs leaves s.Theme.Gradients nil,
// byte-identical to today. Gradients are bootstrap-only (D-105/R8.5); there
// is intentionally no refine_soul path for them — a structured stop list
// does not fit the flat category/token/value refine shape.
func applyGradients(s *Soul, specs []GradientSpec) error {
	if len(specs) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(specs))
	for _, spec := range specs {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			return fmt.Errorf("soul: bootstrap gradient %q: name must be non-empty", spec.Name)
		}
		if seen[name] {
			return fmt.Errorf("soul: bootstrap gradient %q: duplicate name", name)
		}
		seen[name] = true

		if len(spec.Stops) < 2 || len(spec.Stops) > 8 {
			return fmt.Errorf("soul: bootstrap gradient %q: must have 2..8 stops, got %d", name, len(spec.Stops))
		}
		stops := make([]pptx.GradientStop, len(spec.Stops))
		prevPos := -1.0
		for i, st := range spec.Stops {
			if st.Pos < 0 || st.Pos > 1 {
				return fmt.Errorf("soul: bootstrap gradient %q: stop %d pos %v out of [0,1]", name, i, st.Pos)
			}
			if st.Pos <= prevPos {
				return fmt.Errorf("soul: bootstrap gradient %q: stop %d pos %v not strictly ascending", name, i, st.Pos)
			}
			prevPos = st.Pos

			hasHex := st.ColorHex != ""
			hasRole := st.ColorRole != ""
			if hasHex == hasRole {
				return fmt.Errorf("soul: bootstrap gradient %q: stop %d must set exactly one of ColorHex/ColorRole", name, i)
			}
			var color pptx.Color
			if hasHex {
				rgb, err := parseHexColor(st.ColorHex)
				if err != nil {
					return fmt.Errorf("soul: bootstrap gradient %q: stop %d: %w", name, i, err)
				}
				color = rgb
			} else {
				role, ok := surfaceRole(st.ColorRole)
				if !ok {
					return fmt.Errorf("soul: bootstrap gradient %q: stop %d: unknown surface role %q", name, i, st.ColorRole)
				}
				color = pptx.TokenColor(role)
			}
			stops[i] = pptx.GradientStop{Pos: st.Pos, Color: color}
		}

		if s.Theme.Gradients == nil {
			s.Theme.Gradients = map[string]pptx.GradientSpec{}
		}
		s.Theme.Gradients[name] = pptx.GradientSpec{Stops: stops, Angle: spec.Angle, Radial: spec.Radial}
	}
	return nil
}

// deriveAccentText is the final bootstrap pass (R8.6): it WCAG-contrast-checks
// the accent/accentAlt text colors against the resolved canvas, for both the
// light and (when present) dark variant, and replaces them with a legible
// derivation ONLY when the caller actually overrode that accent surface and
// did not also supply an explicit text override for it. A name-only bootstrap
// (no accent set) never reaches the derivation branch, so DeckardWhite's
// hand-tuned text accent stays byte-identical; a soul with no dark palette
// leaves s.Theme.DarkColors nil, so no dark text is invented.
func deriveAccentText(s *Soul, p BootstrapParams) {
	type accentSpec struct {
		key      string
		textRole pptx.TextColorRole
		surfRole pptx.ColorRole
	}
	specs := []accentSpec{
		{"accent", pptx.TextAccent, pptx.ColorAccent},
		{"accentAlt", pptx.TextAccentAlt, pptx.ColorAccentAlt},
	}

	canvas := s.Theme.Colors.Surfaces[pptx.ColorCanvas]
	for _, spec := range specs {
		if accentSurfaceSet(p, spec.key) && !paletteTextSet(p, spec.key) {
			s.Theme.Colors.Text[spec.textRole] = legibleAccentText(s.Theme.Colors.Surfaces[spec.surfRole], canvas)
		}
	}

	if s.Theme.DarkColors == nil {
		return
	}
	darkCanvas := s.Theme.DarkColors.Surfaces[pptx.ColorCanvas]
	if darkCanvas == "" {
		darkCanvas = pptx.RGB("111827") // engine-pinned dark canvas default
	}
	for _, spec := range specs {
		if darkTextSet(p, spec.key) {
			continue
		}
		darkAccent := s.Theme.DarkColors.Surfaces[spec.surfRole]
		if darkAccent == "" {
			darkAccent = s.Theme.Colors.Surfaces[spec.surfRole] // fall back to the light accent
		}
		s.Theme.DarkColors.Text[spec.textRole] = legibleAccentText(darkAccent, darkCanvas)
	}
}

// accentSurfaceSet reports whether the caller overrode the accent surface for
// key ("accent" or "accentAlt"), via the dedicated field or via Palette.Surfaces.
func accentSurfaceSet(p BootstrapParams, key string) bool {
	switch key {
	case "accent":
		if p.Accent != "" {
			return true
		}
	case "accentAlt":
		if p.AccentAlt != "" {
			return true
		}
	}
	if p.Palette == nil {
		return false
	}
	_, ok := p.Palette.Surfaces[key]
	return ok
}

// paletteTextSet reports whether the caller supplied an explicit light text
// override for key via Palette.Text.
func paletteTextSet(p BootstrapParams, key string) bool {
	if p.Palette == nil {
		return false
	}
	_, ok := p.Palette.Text[key]
	return ok
}

// darkTextSet reports whether the caller supplied an explicit dark text
// override for key via DarkPalette.DarkText.
func darkTextSet(p BootstrapParams, key string) bool {
	if p.DarkPalette == nil {
		return false
	}
	_, ok := p.DarkPalette.DarkText[key]
	return ok
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
