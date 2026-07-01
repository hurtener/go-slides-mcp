package soul

import (
	"strings"

	"github.com/hurtener/pptx-go/pptx"
)

// engineDarkCanvasDefault mirrors the engine-pinned dark canvas default used
// elsewhere in this package (deriveAccentText's darkCanvas fallback) — used
// here when no DarkPalette override supplies
// s.Theme.DarkColors.Surfaces[ColorCanvas].
const engineDarkCanvasDefault = "111827"

// registerDecorGradients registers the named brand gradients the default
// decor policy (DefaultDecorPolicy) references, reusing the existing R8.5
// named-gradient storage (s.Theme.Gradients — the same map applyGradients
// writes):
//   - coverWash: a soft 3-stop light diagonal wash for the cover archetype
//     (paper -> surfaceAlt -> a faint accent-tinted canvas).
//   - heroDark: a radial dark wash for the dark/closing archetypes (a
//     slightly accent-lit center fading to the pinned dark canvas at the
//     edge).
//
// Every stop is derived from the already-resolved theme palette —
// deterministic, no clock/rand. A caller-supplied gradient of the same name
// (via BootstrapParams.Gradients) is never overwritten: the caller's own
// brand gradient wins.
func registerDecorGradients(s *Soul) {
	canvas := s.Theme.ResolveColor(pptx.ColorCanvas)
	paper := s.Theme.ResolveColor(pptx.ColorPaper)
	accent := s.Theme.ResolveColor(pptx.ColorAccent)
	surfaceAlt := s.Theme.ResolveColor(pptx.ColorSurfaceAlt)

	darkCanvas := pptx.RGB(engineDarkCanvasDefault)
	if s.Theme.DarkColors != nil {
		if v, ok := s.Theme.DarkColors.Surfaces[pptx.ColorCanvas]; ok && v != "" {
			darkCanvas = v
		}
	}

	if s.Theme.Gradients == nil {
		s.Theme.Gradients = map[string]pptx.GradientSpec{}
	}
	if _, exists := s.Theme.Gradients[gradientCoverWash]; !exists {
		s.Theme.Gradients[gradientCoverWash] = pptx.GradientSpec{
			Angle: 135,
			Stops: []pptx.GradientStop{
				{Pos: 0, Color: paper},
				{Pos: 0.5, Color: surfaceAlt},
				{Pos: 1, Color: pptx.RGB(blendHex(string(canvas), string(accent), 0.90))},
			},
		}
	}
	if _, exists := s.Theme.Gradients[gradientHeroDark]; !exists {
		s.Theme.Gradients[gradientHeroDark] = pptx.GradientSpec{
			Radial: true,
			Stops: []pptx.GradientStop{
				{Pos: 0, Color: pptx.RGB(blendHex(string(darkCanvas), string(accent), 0.88))},
				{Pos: 1, Color: darkCanvas},
			},
		}
	}
}

// blendHex blends baseHex toward mixHex: each channel is
// round(base*baseWeight + mix*(1-baseWeight)), baseWeight in [0,1].
// Deterministic integer rounding; malformed input returns baseHex unchanged.
func blendHex(baseHex, mixHex string, baseWeight float64) string {
	br, bg, bb, ok1 := parseRGBHex(baseHex)
	mr, mg, mb, ok2 := parseRGBHex(mixHex)
	if !ok1 || !ok2 {
		return strings.ToUpper(baseHex)
	}
	r := roundInt(float64(br)*baseWeight + float64(mr)*(1-baseWeight))
	g := roundInt(float64(bg)*baseWeight + float64(mg)*(1-baseWeight))
	b := roundInt(float64(bb)*baseWeight + float64(mb)*(1-baseWeight))
	return strings.ToUpper(hex2(r) + hex2(g) + hex2(b))
}
