package validate

import (
	"fmt"
	"math"
	"strconv"

	"github.com/hurtener/pptx-go/pptx"
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
type contrastPair struct {
	name    string
	text    pptx.TextColorRole
	surface pptx.ColorRole
}

// defaultPairs are the core text-on-surface combinations Deckard relies on.
// Because the IR addresses colors by token, a soul that fails one of these
// fails it on every slide that uses that pairing — so the audit is soul-level.
var defaultPairs = []contrastPair{
	{"primary text on canvas", pptx.TextPrimary, pptx.ColorCanvas},
	{"primary text on surface", pptx.TextPrimary, pptx.ColorSurface},
	{"secondary text on surface", pptx.TextSecondary, pptx.ColorSurface},
	{"inverse text on accent", pptx.TextInverse, pptx.ColorAccent},
	{"inverse text on accent-alt", pptx.TextInverse, pptx.ColorAccentAlt},
}

// AuditTheme checks the soul theme's core text-on-surface pairings against WCAG.
// Below 3:1 is an error; below the 4.5:1 AA target is a warning.
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
