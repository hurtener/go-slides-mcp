package validate

// Severity ranks a finding. An error blocks export; a warning is advisory.
type Severity string

const (
	// SeverityError is a blocking problem (passed = errorCount == 0).
	SeverityError Severity = "error"
	// SeverityWarning is advisory and does not block export.
	SeverityWarning Severity = "warning"
)

// Category groups a finding for the weighted StyleScore.
type Category string

// Validation categories, each carrying a StyleScore weight (see categoryWeights).
const (
	CategoryStructural Category = "structural"
	CategoryContrast   Category = "contrast"
	CategoryTypography Category = "typography"
	CategorySpacing    Category = "spacing"
	CategoryToken      Category = "token"
)

// Issue is one validation finding against a slide or its theme.
type Issue struct {
	Category Category
	Severity Severity
	Message  string
	Path     string // optional IR node path (e.g. "nodes[2].body[0]")
}

// categoryWeights are the StyleScore weights (sum to 1.0). Re-derived for the
// IR-native model from research/01 §4.5: token 30%, contrast 25%, typography
// 15%, spacing 15%, structural 15%.
var categoryWeights = map[Category]float64{
	CategoryToken:      0.30,
	CategoryContrast:   0.25,
	CategoryTypography: 0.15,
	CategorySpacing:    0.15,
	CategoryStructural: 0.15,
}

const (
	errorPenalty   = 0.20 // subtracted from a category subscore per error
	warningPenalty = 0.05 // subtracted from a category subscore per warning
)

// StyleScore is the weighted 0..1 quality score plus its breakdown.
type StyleScore struct {
	// Score is the weighted aggregate in [0,1].
	Score float64
	// Passed is true when there are zero errors (warnings do not block).
	Passed bool
	// ErrorCount / WarnCount are totals across all categories.
	ErrorCount int
	WarnCount  int
	// ByCategory is the clamped subscore for each category in [0,1].
	ByCategory map[Category]float64
}

// Score aggregates issues into a StyleScore. Each category starts at 1.0, loses
// errorPenalty per error and warningPenalty per warning, is clamped to [0,1],
// then weighted into the total. passed = (errorCount == 0).
func Score(issues []Issue) StyleScore {
	sub := make(map[Category]float64, len(categoryWeights))
	for c := range categoryWeights {
		sub[c] = 1.0
	}

	errs, warns := 0, 0
	for _, is := range issues {
		switch is.Severity {
		case SeverityError:
			sub[is.Category] -= errorPenalty
			errs++
		case SeverityWarning:
			sub[is.Category] -= warningPenalty
			warns++
		}
	}

	total := 0.0
	for c, weight := range categoryWeights {
		s := clamp01(sub[c])
		sub[c] = s
		total += weight * s
	}

	return StyleScore{
		Score:      clamp01(total),
		Passed:     errs == 0,
		ErrorCount: errs,
		WarnCount:  warns,
		ByCategory: sub,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
