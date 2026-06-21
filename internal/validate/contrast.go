package validate

import (
	"fmt"
	"math"
	"strconv"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// WCAG contrast thresholds.
const (
	contrastAA  = 4.5 // AA for normal text
	contrastMin = 3.0 // AA for large text / UI; below this is an error
)

// ContrastRatio returns the WCAG 2.1 contrast ratio between two colors (1..21)
// and whether both parsed. The ratio is symmetric and order-independent.
func ContrastRatio(a, b pptx.RGB) (float64, bool) {
	la, ok1 := relLuminance(a)
	lb, ok2 := relLuminance(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	hi, lo := la+0.05, lb+0.05
	if hi < lo {
		hi, lo = lo, hi
	}
	return hi / lo, true
}

// relLuminance is the WCAG relative luminance of a 6-hex color in [0,1].
func relLuminance(c pptx.RGB) (float64, bool) {
	r, g, b, ok := parseHex(string(c))
	if !ok {
		return 0, false
	}
	lin := func(ch float64) float64 {
		ch /= 255
		if ch <= 0.03928 {
			return ch / 12.92
		}
		return math.Pow((ch+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b), true
}

// parseHex parses a 6-digit RRGGBB hex string (no leading '#').
func parseHex(s string) (r, g, b float64, ok bool) {
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64((v >> 16) & 0xFF), float64((v >> 8) & 0xFF), float64(v & 0xFF), true
}

// contrastPair is a text-on-surface pairing the soul promises to render legibly.
// largeText marks pairs used for display/heading/hero text (WCAG AA large-text
// threshold 3:1 applies instead of the normal-text 4.5:1 threshold).
type contrastPair struct {
	name      string
	text      pptx.TextColorRole
	surface   pptx.ColorRole
	largeText bool // true = WCAG AA large-text (3:1); false = normal text (4.5:1)
}

// defaultPairs are the core text-on-surface combinations Deckard relies on.
// Because the IR addresses colors by token, a soul that fails one of these
// fails it on every slide that uses that pairing — so the audit is soul-level.
// TextInverse-on-accent pairings are large-text (hero/display roles) so the
// WCAG AA large-text threshold (3:1) applies; the normal-text pairs keep 4.5:1.
var defaultPairs = []contrastPair{
	{"primary text on canvas", pptx.TextPrimary, pptx.ColorCanvas, false},
	{"primary text on surface", pptx.TextPrimary, pptx.ColorSurface, false},
	{"secondary text on surface", pptx.TextSecondary, pptx.ColorSurface, false},
	{"inverse text on accent", pptx.TextInverse, pptx.ColorAccent, true},        // hero/display = large text
	{"inverse text on accent-alt", pptx.TextInverse, pptx.ColorAccentAlt, true}, // hero/display = large text
}

// AuditTheme checks the soul theme's core text-on-surface pairings against WCAG.
// Below 3:1 is an error for all text. Between 3:1 and 4.5:1 is a warning for
// normal-text pairs; large-text pairs (hero/display roles) pass at ≥ 3:1.
func AuditTheme(t *pptx.Theme) []Issue {
	if t == nil {
		return nil
	}
	var issues []Issue
	for _, p := range defaultPairs {
		ratio, ok := ContrastRatio(t.ResolveTextColor(p.text), t.ResolveColor(p.surface))
		if !ok {
			continue
		}
		// WCAG AA: normal text requires 4.5:1; large text (≥ 18pt or ≥ 14pt bold)
		// requires 3:1. Below 3:1 is always an error regardless of text size.
		aaTarget := contrastAA // 4.5 for normal text
		if p.largeText {
			aaTarget = contrastMin // large text: the AA target equals the minimum (3:1)
		}
		switch {
		case ratio < contrastMin:
			issues = append(issues, Issue{
				Category: CategoryContrast,
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: contrast %.2f:1 is below the %.0f:1 minimum", p.name, ratio, contrastMin),
			})
		case ratio < aaTarget:
			issues = append(issues, Issue{
				Category: CategoryContrast,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("%s: contrast %.2f:1 is below the AA %.1f:1 target", p.name, ratio, aaTarget),
			})
		}
	}
	return issues
}

// AuditSlideColors checks the primary-text/canvas and primary-text/surface pairs
// against the engine's R7 per-slide resolved colors (scene.SlideColors). It is
// variant-aware: for a VariantDark slide the engine resolves dark-palette RGBs
// into Colors, so this check evaluates the slide against what it actually renders
// with rather than the soul's light theme. Both checked pairs are body-text
// weight → normal-text 4.5:1 AA threshold.
func AuditSlideColors(sc scene.SlideColors) []Issue {
	pairs := []struct {
		name    string
		text    pptx.RGB
		surface pptx.RGB
	}{
		{"primary text on canvas", sc.PrimaryText, sc.Canvas},
		{"primary text on surface", sc.PrimaryText, sc.Surface},
	}
	var issues []Issue
	for _, p := range pairs {
		ratio, ok := ContrastRatio(p.text, p.surface)
		if !ok {
			continue
		}
		switch {
		case ratio < contrastMin:
			issues = append(issues, Issue{
				Category: CategoryContrast,
				Severity: SeverityError,
				Message:  fmt.Sprintf("%s: contrast %.2f:1 is below the %.0f:1 minimum", p.name, ratio, contrastMin),
			})
		case ratio < contrastAA:
			issues = append(issues, Issue{
				Category: CategoryContrast,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("%s: contrast %.2f:1 is below the AA %.1f:1 target", p.name, ratio, contrastAA),
			})
		}
	}
	return issues
}
