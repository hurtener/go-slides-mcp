package soul

import (
	"github.com/hurtener/go-slides-mcp/internal/soul/fonts"
	"github.com/hurtener/pptx-go/pptx"
)

// DeckardWhiteID is the stable id of the built-in default soul.
const DeckardWhiteID = "deckard-white"

// DeckardWhite returns the built-in default soul: a warm, editorial, off-white
// theme — teal accent, warm near-black text, serif display headings, two type
// weights (400/500), airy spacing. Values are the docs/research/04-design-tokens.md
// Deckard White → *pptx.Theme mapping. Each call returns a fresh, complete theme.
func DeckardWhite() *Soul {
	t := pptx.DefaultTheme().Clone()
	t.Name = "Deckard White"
	t.HeadingFont = "Playfair Display"
	t.BodyFont = "Inter"

	// Surface colors (warm off-white palette + teal accent split).
	t.Colors.Surfaces[pptx.ColorCanvas] = "FAF7F2"
	t.Colors.Surfaces[pptx.ColorSurface] = "FFFFFF"
	t.Colors.Surfaces[pptx.ColorSurfaceAlt] = "F4EFE6"
	t.Colors.Surfaces[pptx.ColorAccent] = "3B9C94"
	t.Colors.Surfaces[pptx.ColorAccentAlt] = "2B7A73"
	t.Colors.Surfaces[pptx.ColorAccentWarm] = "D97B1A"
	t.Colors.Surfaces[pptx.ColorSuccess] = "3F8E6B"
	t.Colors.Surfaces[pptx.ColorWarning] = "D97B1A"
	t.Colors.Surfaces[pptx.ColorError] = "B64A4A"
	t.Colors.Surfaces[pptx.ColorInfo] = "2B7A73"

	// Text colors (warm near-black; AA-safe teal for accent text).
	t.Colors.Text[pptx.TextPrimary] = "2B2723"
	t.Colors.Text[pptx.TextSecondary] = "6A625B"
	t.Colors.Text[pptx.TextTertiary] = "6A625B"
	t.Colors.Text[pptx.TextInverse] = "FAF7F2"
	t.Colors.Text[pptx.TextMuted] = "B8B0A4"
	t.Colors.Text[pptx.TextAccent] = "2B7A73"
	t.Colors.Text[pptx.TextAccentAlt] = "2B7A73"
	t.Colors.Text[pptx.TextSuccess] = "3F8E6B"
	t.Colors.Text[pptx.TextWarning] = "D97B1A"
	t.Colors.Text[pptx.TextError] = "B64A4A"

	// Typography: serif display/titles (Playfair/Lora), sans subheads/body
	// (Inter), mono unchanged. Weights clamped to the 400/500 rule.
	setType(t, pptx.TypeDisplay, "Playfair Display", 40, 400)
	setType(t, pptx.TypeH1, "Lora", 32, 400)
	setType(t, pptx.TypeH2, "Lora", 28, 400)
	setType(t, pptx.TypeH3, "Lora", 24, 400)
	setType(t, pptx.TypeH4, "Inter", 20, 500)
	setType(t, pptx.TypeH5, "Inter", 16, 500)
	setType(t, pptx.TypeBody, "Inter", 14, 400)
	setType(t, pptx.TypeBodySmall, "Inter", 12, 400)
	setType(t, pptx.TypeCaption, "Inter", 10, 500)
	setType(t, pptx.TypeMono, "Consolas", 13, 400)
	setType(t, pptx.TypeCode, "Consolas", 12, 400)

	// Spacing: tighten the small end to the warm 4/8/12 rhythm; LG/XL/2XL match
	// the default already.
	t.Spacing[pptx.SpaceXS] = pptx.Pt(4)
	t.Spacing[pptx.SpaceSM] = pptx.Pt(8)
	t.Spacing[pptx.SpaceMD] = pptx.Pt(12)

	// Radii: softer corners (sm 8 / md 12 / lg 16).
	t.Radii[pptx.RadiusSM] = pptx.Pt(8)
	t.Radii[pptx.RadiusMD] = pptx.Pt(12)
	t.Radii[pptx.RadiusLG] = pptx.Pt(16)

	return &Soul{
		ID:          DeckardWhiteID,
		Name:        "Deckard White",
		Description: "Warm editorial off-white — teal accent, serif display, airy and rebrandable.",
		Status:      "ready",
		Theme:       t,
		// The default soul names bundled OFL serif/sans faces (Playfair Display,
		// Lora, Inter); register the provider so they embed on export (R9.1) and
		// the editorial serif renders on any machine.
		FontProvider: fonts.Provider(),
		StyleGuide: StyleGuide{
			NorthStar: "Calm, editorial, premium: warm off-white, generous whitespace, one teal accent, serif titles.",
			Do: []string{
				"Left-align everything; use sentence case.",
				"Use exactly one italic serif accent word per title set, in teal.",
				"Lean on whitespace; keep two type weights (400/500).",
			},
			Dont: []string{
				"Use heavy weights (600/700/900) or all-caps outside small eyebrows.",
				"Introduce a second accent hue; the teal is the only brand color.",
				"Center body content or crowd containers below the 16pt padding floor.",
			},
		},
		// Tokens with no native pptx.Theme field; applied as literal strokes/washes.
		Extensions: map[string]string{
			"border":       "E0D5CA",
			"borderStrong": "D8D0C4",
			"accentSoft":   "3B9C94", // rendered as a 12% wash via TokenColorAlpha(ColorAccent)
			// tooltip (#FF9645) is app-only chrome, intentionally NOT a deck token.
		},
	}
}

// setType overrides one type role's family/size/weight, preserving Italic.
func setType(t *pptx.Theme, role pptx.TypeRole, family string, size float64, weight int) {
	fs := t.Typography[role]
	fs.Family = family
	fs.Size = size
	fs.Weight = weight
	t.Typography[role] = fs
}
