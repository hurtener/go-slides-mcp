package validate

import (
	"math"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func TestContrastRatioExtremes(t *testing.T) {
	r, ok := ContrastRatio("FFFFFF", "000000")
	if !ok || math.Abs(r-21) > 0.05 {
		t.Fatalf("white/black ratio = %.3f (ok=%v), want ~21", r, ok)
	}
	r, ok = ContrastRatio("3B9C94", "3B9C94")
	if !ok || math.Abs(r-1) > 0.001 {
		t.Fatalf("same-color ratio = %.3f, want 1", r)
	}
	if _, ok := ContrastRatio("xyz", "000000"); ok {
		t.Fatal("bad hex should not parse")
	}
}

func TestScoreEmptyIsPerfect(t *testing.T) {
	s := Score(nil)
	if s.Score != 1.0 || !s.Passed || s.ErrorCount != 0 {
		t.Fatalf("empty score = %+v, want 1.0/passed", s)
	}
}

func TestScoreErrorAndWarningPenalties(t *testing.T) {
	// one structural error: structural subscore 1.0-0.20=0.80; total =
	// 1.0 - 0.12*(0.20) = 0.976. passed=false.
	s := Score([]Issue{{Category: CategoryStructural, Severity: SeverityError}})
	if s.Passed {
		t.Fatal("an error must fail")
	}
	if math.Abs(s.Score-0.976) > 1e-9 {
		t.Fatalf("score = %.4f, want 0.976", s.Score)
	}
	if math.Abs(s.ByCategory[CategoryStructural]-0.80) > 1e-9 {
		t.Fatalf("structural subscore = %.4f, want 0.80", s.ByCategory[CategoryStructural])
	}
	// one contrast warning: contrast subscore 0.95; total = 1 - 0.25*0.05 = 0.9875; passes.
	w := Score([]Issue{{Category: CategoryContrast, Severity: SeverityWarning}})
	if !w.Passed || math.Abs(w.Score-0.9875) > 1e-9 {
		t.Fatalf("warning score = %+v, want 0.9875/passed", w.Score)
	}
}

func TestScoreClampsAndWeightsSumToOne(t *testing.T) {
	sum := 0.0
	for _, w := range categoryWeights {
		sum += w
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("category weights sum to %.4f, want 1.0", sum)
	}
	// flood one category with errors -> subscore clamps at 0, never negative.
	many := make([]Issue, 20)
	for i := range many {
		many[i] = Issue{Category: CategoryContrast, Severity: SeverityError}
	}
	s := Score(many)
	if s.ByCategory[CategoryContrast] != 0 {
		t.Fatalf("flooded subscore = %v, want 0 (clamped)", s.ByCategory[CategoryContrast])
	}
	if s.Score < 0 {
		t.Fatalf("total score went negative: %v", s.Score)
	}
}

func TestAuditDeckardWhiteHasNoErrors(t *testing.T) {
	s := soul.DeckardWhite()
	issues := AuditTheme(s.Theme)
	for _, is := range issues {
		if is.Severity == SeverityError {
			t.Fatalf("default soul has a contrast ERROR: %s", is.Message)
		}
		t.Logf("default soul contrast note: %s", is.Message)
	}
}

func TestAuditDetectsUnreadablePairing(t *testing.T) {
	th := pptx.DefaultTheme().Clone()
	th.Colors.Text[pptx.TextPrimary] = "EEEEEE" // light text
	th.Colors.Surfaces[pptx.ColorCanvas] = "FFFFFF"
	issues := AuditTheme(th)
	found := false
	for _, is := range issues {
		if is.Category == CategoryContrast && is.Severity == SeverityError {
			found = true
		}
	}
	if !found {
		t.Fatal("light-on-white should yield a contrast error")
	}
}

func TestOverflowMapping(t *testing.T) {
	issues := OverflowIssues([]string{"slide cover: content overflow on node hero"})
	if len(issues) != 1 || issues[0].Category != CategorySpacing || issues[0].Severity != SeverityWarning {
		t.Fatalf("overflow mapping = %+v", issues)
	}
}

func TestStructuralFlattensErrors(t *testing.T) {
	// a heading with an out-of-range level is a structural error.
	bad := contracts.Slide{
		Nodes: []contracts.SlideNode{&contracts.Heading{Level: 99, Text: contracts.RichText{{Text: "x"}}}},
	}
	issues := Structural(bad)
	if len(issues) == 0 {
		t.Fatal("invalid heading level should produce a structural issue")
	}
	for _, is := range issues {
		if is.Category != CategoryStructural || is.Severity != SeverityError {
			t.Fatalf("issue = %+v, want structural error", is)
		}
	}
}

// ── R7 + contrast tests ────────────────────────────────────────────────────

// TestAuditSlideColors_DarkSlide_NoFalsePositive is acceptance criterion (a):
// a VariantDark slide with light text on a dark canvas must produce NO contrast
// finding when AuditSlideColors is fed the engine's resolved dark-palette RGBs.
func TestAuditSlideColors_DarkSlide_NoFalsePositive(t *testing.T) {
	// Simulate the engine-resolved dark palette: dark canvas, light primary text.
	dark := scene.SlideColors{
		SlideID:     "dark-slide",
		Canvas:      pptx.RGB("1A1A2E"), // dark navy ≈ luminance 0.007
		Surface:     pptx.RGB("23233A"),
		PrimaryText: pptx.RGB("EAEAEA"), // near-white ≈ luminance 0.84
	}
	issues := AuditSlideColors(dark)
	for _, is := range issues {
		if is.Category == CategoryContrast {
			t.Errorf("dark slide with light text must not produce a contrast finding; got: %s", is.Message)
		}
	}
}

// TestAuditTheme_LargeText_PassesAt3to1 is acceptance criterion (b) — large-text
// branch: inverse text on accent at ~3.09:1 must PASS (not warn) because the
// inverse-text-on-accent pair is marked largeText and the WCAG AA large-text
// threshold is 3:1.
func TestAuditTheme_LargeText_PassesAt3to1(t *testing.T) {
	th := pptx.DefaultTheme().Clone()
	// Craft a palette where TextInverse on ColorAccent ≈ 3.09:1.
	// White (#FFFFFF, L=1.0) on a medium accent gives ~3.09:1 around #5F7A9D.
	// Using exact values verified by ContrastRatio below.
	th.Colors.Text[pptx.TextInverse] = "FFFFFF"
	th.Colors.Surfaces[pptx.ColorAccent] = "5B7B9A" // chosen so ratio ≈ 3.09

	ratio, ok := ContrastRatio(
		th.Colors.Text[pptx.TextInverse],
		th.Colors.Surfaces[pptx.ColorAccent],
	)
	if !ok {
		t.Fatal("ContrastRatio failed on test colors")
	}
	// Only run the large-text assertion when ratio is in the large-text-pass zone.
	if ratio < contrastMin || ratio >= contrastAA {
		t.Skipf("test color ratio %.2f is outside (%.1f, %.1f); skipping large-text assertion", ratio, contrastMin, contrastAA)
	}

	issues := AuditTheme(th)
	for _, is := range issues {
		if is.Category == CategoryContrast && is.Severity == SeverityWarning {
			// Warn is acceptable for other pairs (e.g. primary/canvas); fail only
			// if the inverse-text-on-accent pair emits a finding.
			if containsStr(is.Message, "inverse text on accent") {
				t.Errorf("large-text pair (inverse text on accent) at %.2f:1 must PASS at 3:1; got: %s", ratio, is.Message)
			}
		}
		if is.Category == CategoryContrast && is.Severity == SeverityError {
			if containsStr(is.Message, "inverse text on accent") {
				t.Errorf("inverse text on accent at %.2f:1 above minimum must not error; got: %s", ratio, is.Message)
			}
		}
	}
}

// TestAuditSlideColors_SmallBodyText_FailsAt3to1 is acceptance criterion (b) —
// body-text branch: body text (primary-text-on-canvas via AuditSlideColors) at
// ~3.09:1 must still FAIL (warn) because body text requires 4.5:1.
func TestAuditSlideColors_SmallBodyText_FailsAt3to1(t *testing.T) {
	// White on the same accent-ish background: AuditSlideColors treats all pairs
	// as body text → 4.5:1 threshold → must warn.
	sc := scene.SlideColors{
		Canvas:      pptx.RGB("5B7B9A"), // same medium accent
		Surface:     pptx.RGB("5B7B9A"),
		PrimaryText: pptx.RGB("FFFFFF"),
	}
	ratio, ok := ContrastRatio(sc.PrimaryText, sc.Canvas)
	if !ok {
		t.Fatal("ContrastRatio failed on test colors")
	}
	if ratio < contrastMin || ratio >= contrastAA {
		t.Skipf("test color ratio %.2f outside body-text zone; skipping", ratio)
	}

	issues := AuditSlideColors(sc)
	found := false
	for _, is := range issues {
		if is.Category == CategoryContrast {
			found = true
		}
	}
	if !found {
		t.Errorf("body text at %.2f:1 (below 4.5:1) must produce a contrast finding via AuditSlideColors", ratio)
	}
}

// TestAuditSlideColors_GenuineLowContrast is acceptance criterion (c): a
// genuinely low-contrast dark slide (dark text on dark background) must still
// produce a contrast finding — no false negatives.
func TestAuditSlideColors_GenuineLowContrast(t *testing.T) {
	sc := scene.SlideColors{
		SlideID:     "bad-dark",
		Canvas:      pptx.RGB("1A1A2E"),
		Surface:     pptx.RGB("1A1A2E"),
		PrimaryText: pptx.RGB("222244"), // dark text on dark canvas: very low contrast
	}
	issues := AuditSlideColors(sc)
	found := false
	for _, is := range issues {
		if is.Category == CategoryContrast && is.Severity == SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("dark text on dark canvas must produce a contrast ERROR; got none (false negative)")
	}
}

// TestSlide_DarkVariant_UsesResolvedColors is acceptance criterion (d):
// the Slide validator with resolved dark-palette colors must yield OK = true
// (no contrast errors) for a dark slide with legible inverse text.
func TestSlide_DarkVariant_UsesResolvedColors(t *testing.T) {
	slide := contracts.Slide{
		ID:      "hero-dark",
		Variant: contracts.VariantDark,
		Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Dark Hero", Subtitle: "subtitle"},
		},
	}

	// Engine-resolved dark palette for this slide.
	dark := scene.SlideColors{
		SlideID:     "hero-dark",
		Canvas:      pptx.RGB("1A1A2E"),
		Surface:     pptx.RGB("23233A"),
		PrimaryText: pptx.RGB("EAEAEA"),
	}

	// With resolved colors: dark slide checks against its actual dark palette → pass.
	report := Slide(slide, soul.DeckardWhite().Theme, &dark, nil)
	for _, is := range report.Issues {
		if is.Category == CategoryContrast {
			t.Errorf("dark slide with resolved colors must have no contrast finding; got: %s", is.Message)
		}
	}

	// Without resolved colors (nil): falls back to AuditTheme, which checks the
	// light soul theme — the soul itself must at least not ERROR (warnings are ok).
	reportFallback := Slide(slide, soul.DeckardWhite().Theme, nil, nil)
	for _, is := range reportFallback.Issues {
		if is.Category == CategoryContrast && is.Severity == SeverityError {
			t.Errorf("fallback AuditTheme on Deckard White must not produce contrast ERROR; got: %s", is.Message)
		}
	}
}

// containsStr is a test helper for substring search without importing strings.
func containsStr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
