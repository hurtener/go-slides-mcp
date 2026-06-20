package validate

import (
	"math"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
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
