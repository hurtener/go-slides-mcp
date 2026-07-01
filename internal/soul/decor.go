package soul

import (
	"strconv"
	"strings"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
)

// Subtle-alpha bands a bootstrapped soul's DecorPolicy draws from (R13.13).
// The engine's Decoration.Opacity is "0 = opaque" — a decoration with no
// explicit opacity is FULLY OPAQUE, not invisible — so the policy must set an
// explicit subtle Opacity on every decoration it injects; it never relies on
// an engine default. An explicit per-slide opacity on a caller-authored
// decoration still overrides (the policy only fills slides that set none).
const (
	// textureAlpha is the opacity for full-bleed pattern textures (dot grids,
	// starfields): band 4-10%.
	textureAlpha = 0.06
	// glowAlpha is the opacity for atmospheric glow ornaments: band 8-15%.
	glowAlpha = 0.12
	// watermarkAlpha is the opacity for oversized ghost-text watermarks: band
	// 6-12%. Unused this PR (Decoration.Text/FontSize are deferred) — defined
	// here for the record so the band is documented alongside its siblings.
	watermarkAlpha = 0.09
)

// paperTint derives a low-chroma off-white ~2-4% removed from pure white
// (R13.1), used as the bootstrapped soul's ColorPaper. If canvasHex is
// already off-white (not "FFFFFF"), it is returned unchanged — a caller who
// already picked a tinted canvas gets no further tinting. If canvasHex IS
// pure white, 10% of surfaceAltHex is blended into white per RGB channel:
//
//	channel = 255 - round((255 - surfaceAltChannel) * 0.10)
//
// Parsing is defensive: a malformed hex returns canvasHex unchanged.
func paperTint(canvasHex, surfaceAltHex string) string {
	if !strings.EqualFold(canvasHex, "FFFFFF") {
		return strings.ToUpper(canvasHex)
	}
	sr, sg, sb, ok := parseRGBHex(surfaceAltHex)
	if !ok {
		return canvasHex
	}
	r := 255 - roundInt((255-float64(sr))*0.10)
	g := 255 - roundInt((255-float64(sg))*0.10)
	b := 255 - roundInt((255-float64(sb))*0.10)
	return strings.ToUpper(hex2(r) + hex2(g) + hex2(b))
}

// parseRGBHex parses a 6-digit hex string (optionally "#"-prefixed) into its
// three channel bytes. Returns ok=false on any malformed input.
func parseRGBHex(s string) (r, g, b int, ok bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	rv, err1 := strconv.ParseInt(s[0:2], 16, 32)
	gv, err2 := strconv.ParseInt(s[2:4], 16, 32)
	bv, err3 := strconv.ParseInt(s[4:6], 16, 32)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return int(rv), int(gv), int(bv), true
}

// roundInt is deterministic integer rounding (round-half-away-from-zero) for
// the non-negative channel deltas paperTint computes.
func roundInt(f float64) int {
	return int(f + 0.5)
}

// hex2 formats a 0..255 channel as a zero-padded 2-digit uppercase hex byte.
func hex2(v int) string {
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	s := strconv.FormatInt(int64(v), 16)
	if len(s) < 2 {
		s = "0" + s
	}
	return s
}

// Ornament preset names, verified against the engine's curated registry
// (github.com/hurtener/pptx-go/scene/ornaments.Curated(), NameGridDots /
// NameStarfield — both exist; R13-D uses no other presets). grid_dots is a
// full-bleed dot lattice; starfield is the organic size/alpha-varied scatter
// (R13.6, D-110).
const (
	presetGridDots  = "grid_dots"
	presetStarfield = "starfield"
)

// Named brand gradients the decor policy references (R8.5 style), registered
// on the theme by Bootstrap alongside DefaultDecorPolicy (decor_gradients.go).
const (
	gradientCoverWash = "coverWash"
	gradientHeroDark  = "heroDark"
)

// DefaultDecorPolicy builds a bootstrapped soul's per-archetype decor recipe
// (R13.12) from the given theme. It uses only vocabulary that renders today:
// the content archetype gets a tinted-paper canvas plus a neutral full-bleed
// dot texture; cover and dark get a named brand gradient (registered by
// Bootstrap alongside this policy, see decor_gradients.go) plus, for dark, an
// accent-tinted starfield; section gets a tinted color break with no
// decoration; closing mirrors dark.
//
// Dark starfield color: the engine swaps in a derived DARK theme for a
// VariantDark slide (scene/render.go darkThemeFrom), so a decoration's surface
// color role resolves against the dark palette — canvas/surface/surfaceAlt all
// darken and would render near-invisible dark-on-dark specks. Accent (and the
// semantic roles) are the roles the dark derivation PRESERVES, so the starfield
// uses ColorAccent to stay a visible, subtly brand-tinted scatter on the dark
// wash. A truly pale/white starfield needs a literal-or-inverse decoration
// color the current role vocabulary can't express under VariantDark — an
// engine gap (see the R13 gaps note). Every injected Decoration carries an explicit subtle
// Opacity — the engine's Opacity zero value is "opaque", so relying on the
// default would render a solid slab, not a faint wash. Deterministic: a
// literal map, no clock/rand, no map-order-dependent construction. The recipe
// is expressed entirely in color ROLES (resolved against the active theme at
// render time), not resolved hexes, so the theme parameter is currently
// unused by this vocabulary-only recipe; it stays part of the signature so a
// richer, theme-derived recipe can grow here without an API break.
func DefaultDecorPolicy(_ *pptx.Theme) *contracts.DecorPolicy {
	surfaceAlt := contracts.ColorSurfaceAlt

	return &contracts.DecorPolicy{
		ByArchetype: map[contracts.SlideArchetype]contracts.ArchetypeDecor{
			contracts.ArchetypeContent: {
				Background: &contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorPaper},
				Decorations: []contracts.Decoration{
					{
						Kind:    contracts.DecorationPreset,
						Preset:  presetGridDots,
						Layer:   contracts.LayerBackground,
						Anchor:  contracts.AnchorCenter,
						Bleed:   true,
						Opacity: textureAlpha,
						Color:   surfaceAlt,
						Pitch:   28,
					},
				},
			},
			contracts.ArchetypeCover: {
				Background: &contracts.Background{Kind: contracts.BackgroundGradient, GradientName: gradientCoverWash},
			},
			contracts.ArchetypeDark: {
				Background: &contracts.Background{Kind: contracts.BackgroundGradient, GradientName: gradientHeroDark},
				Decorations: []contracts.Decoration{
					{
						Kind:    contracts.DecorationPreset,
						Preset:  presetStarfield,
						Layer:   contracts.LayerBackground,
						Anchor:  contracts.AnchorCenter,
						Bleed:   true,
						Opacity: textureAlpha,
						Color:   contracts.ColorAccent,
						Pitch:   36,
					},
				},
			},
			contracts.ArchetypeSection: {
				Background: &contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorSurfaceAlt},
			},
			// Closing mirrors dark (R13.12): a dark closing slide gets the
			// same radial wash + pale starfield treatment as an in-deck dark
			// slide. Documented choice — cover was the other candidate, but a
			// closing slide reads better as a dark "curtain close" than a
			// repeat of the opening light wash.
			contracts.ArchetypeClosing: {
				Background: &contracts.Background{Kind: contracts.BackgroundGradient, GradientName: gradientHeroDark},
				Decorations: []contracts.Decoration{
					{
						Kind:    contracts.DecorationPreset,
						Preset:  presetStarfield,
						Layer:   contracts.LayerBackground,
						Anchor:  contracts.AnchorCenter,
						Bleed:   true,
						Opacity: textureAlpha,
						Color:   contracts.ColorAccent,
						Pitch:   36,
					},
				},
			},
		},
	}
}
