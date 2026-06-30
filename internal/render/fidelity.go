package render

import (
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// ColorMismatch is one resolved render color that does not equal the soul's
// declared value for that role and variant (R8.10). A non-empty mismatch list
// means a deck's rendered colors drifted from the soul that authored it — the
// fidelity gate this type exists to support.
type ColorMismatch struct {
	SlideID string
	Variant string // "light" | "dark"
	Role    string // "canvas" | "surface" | "surfaceAlt" | "accent" | "accentAlt" | "primaryText" | "textAccent"
	Want    string // soul-declared hex
	Got     string // resolved render hex
}

// SoulColorFidelity renders doc through s and returns every per-slide resolved
// color that does not equal the soul's declared value for that role/variant.
// An empty result means full soul->engine fidelity: every token the soul
// declares (light Theme, plus any Theme.DarkColors override) reached the
// rendered bytes unchanged. Deterministic — it performs no I/O beyond Render.
func SoulColorFidelity(doc contracts.SlideDoc, s *soul.Soul) ([]ColorMismatch, error) {
	_, stats, err := Render(doc, s)
	if err != nil {
		return nil, fmt.Errorf("soul color fidelity: %w", err)
	}

	slideVariant := make(map[string]bool, len(doc.Slides))
	for _, sl := range doc.Slides {
		slideVariant[sl.ID] = sl.Variant == contracts.VariantDark
	}

	var mismatches []ColorMismatch
	for i, sc := range stats.Colors {
		dark, ok := slideVariant[sc.SlideID]
		if !ok {
			// Fall back to scene-order index if SlideID lookup fails (defensive;
			// every mapped slide carries its contract ID through to SlideColors).
			if i < len(doc.Slides) {
				dark = doc.Slides[i].Variant == contracts.VariantDark
			}
		}
		mismatches = append(mismatches, compareSlideColors(sc, s, dark)...)
	}
	return mismatches, nil
}

// compareSlideColors compares one rendered slide's resolved colors against the
// soul's declared values for the given variant. For a LIGHT slide every
// reported role is compared to the soul's light Theme. For a DARK slide, ONLY
// the roles the soul EXPLICITLY declares in Theme.DarkColors are compared —
// roles the soul does not declare fall to the engine's pinned dark default and
// are not soul-owned, so they are skipped. Returns every mismatch found.
func compareSlideColors(sc scene.SlideColors, s *soul.Soul, dark bool) []ColorMismatch {
	if dark {
		return compareDarkSlideColors(sc, s)
	}
	return compareLightSlideColors(sc, s)
}

func compareLightSlideColors(sc scene.SlideColors, s *soul.Soul) []ColorMismatch {
	th := s.Theme
	var out []ColorMismatch
	checks := []struct {
		role string
		got  pptx.RGB
		want pptx.RGB
	}{
		{"canvas", sc.Canvas, th.ResolveColor(pptx.ColorCanvas)},
		{"surface", sc.Surface, th.ResolveColor(pptx.ColorSurface)},
		{"surfaceAlt", sc.SurfaceAlt, th.ResolveColor(pptx.ColorSurfaceAlt)},
		{"accent", sc.Accent, th.ResolveColor(pptx.ColorAccent)},
		{"accentAlt", sc.AccentAlt, th.ResolveColor(pptx.ColorAccentAlt)},
		{"primaryText", sc.PrimaryText, th.ResolveTextColor(pptx.TextPrimary)},
		{"textAccent", sc.TextAccent, th.ResolveTextColor(pptx.TextAccent)},
	}
	for _, ck := range checks {
		if ck.got != ck.want {
			out = append(out, ColorMismatch{
				SlideID: sc.SlideID, Variant: "light", Role: ck.role,
				Want: string(ck.want), Got: string(ck.got),
			})
		}
	}
	return out
}

func compareDarkSlideColors(sc scene.SlideColors, s *soul.Soul) []ColorMismatch {
	dp := s.Theme.DarkColors
	if dp == nil {
		// No soul-declared dark overrides: every dark color comes from the
		// engine's pinned default, which is not soul-owned. Nothing to compare.
		return nil
	}

	var out []ColorMismatch
	surfaceChecks := []struct {
		role string
		col  pptx.ColorRole
		got  pptx.RGB
	}{
		{"canvas", pptx.ColorCanvas, sc.Canvas},
		{"surface", pptx.ColorSurface, sc.Surface},
		{"surfaceAlt", pptx.ColorSurfaceAlt, sc.SurfaceAlt},
		{"accent", pptx.ColorAccent, sc.Accent},
		{"accentAlt", pptx.ColorAccentAlt, sc.AccentAlt},
	}
	for _, ck := range surfaceChecks {
		want, declared := dp.Surfaces[ck.col]
		if !declared {
			continue
		}
		if ck.got != want {
			out = append(out, ColorMismatch{
				SlideID: sc.SlideID, Variant: "dark", Role: ck.role,
				Want: string(want), Got: string(ck.got),
			})
		}
	}

	textChecks := []struct {
		role string
		col  pptx.TextColorRole
		got  pptx.RGB
	}{
		{"primaryText", pptx.TextPrimary, sc.PrimaryText},
		{"textAccent", pptx.TextAccent, sc.TextAccent},
	}
	for _, ck := range textChecks {
		want, declared := dp.Text[ck.col]
		if !declared {
			continue
		}
		if ck.got != want {
			out = append(out, ColorMismatch{
				SlideID: sc.SlideID, Variant: "dark", Role: ck.role,
				Want: string(want), Got: string(ck.got),
			})
		}
	}
	return out
}
