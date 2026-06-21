package validate

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Report is the full validation result: a StyleScore plus the issues behind it.
type Report struct {
	Score  StyleScore
	Issues []Issue
}

// Messages returns every issue message in order (for the flat Issues string list).
func (r Report) Messages() []string {
	out := make([]string, 0, len(r.Issues))
	for _, is := range r.Issues {
		out = append(out, is.Message)
	}
	return out
}

// Slide scores a single slide: structural IR checks, contrast, content-fidelity
// (empty RichText leaves), and layout-overflow warnings (renderWarnings comes
// from a render of that slide; pass nil to skip the render-truth pass).
//
// Contrast is variant-aware (R7): when slideColors is non-nil the check uses
// the engine's per-slide resolved RGBs (including the dark palette for
// VariantDark slides) via AuditSlideColors. When slideColors is nil it falls
// back to the soul-level AuditTheme(theme). theme may be nil (contrast skipped
// in the fallback path).
func Slide(slide contracts.Slide, theme *pptx.Theme, slideColors *scene.SlideColors, renderWarnings []string) Report {
	var issues []Issue
	issues = append(issues, Structural(slide)...)
	issues = append(issues, Fidelity(slide)...)
	if slideColors != nil {
		issues = append(issues, AuditSlideColors(*slideColors)...)
	} else {
		issues = append(issues, AuditTheme(theme)...)
	}
	issues = append(issues, OverflowIssues(renderWarnings)...)
	return Report{Score: Score(issues), Issues: issues}
}

// Deck scores a whole deck for export. Soul-level contrast (AuditTheme) runs
// once at the deck level. Per-slide: structural, fidelity, overflow, and
// variant-aware contrast when perSlideColors is provided (R7).
//
// perSlideWarnings and perSlideColors are keyed by slide index in doc order;
// pass nil for either to skip the corresponding per-slide check.
func Deck(doc contracts.SlideDoc, theme *pptx.Theme, perSlideWarnings [][]string, perSlideColors []scene.SlideColors) (Report, []Report) {
	var deckIssues []Issue
	deckIssues = append(deckIssues, StructuralDoc(doc)...)
	deckIssues = append(deckIssues, AuditTheme(theme)...)

	perSlide := make([]Report, len(doc.Slides))
	for i, slide := range doc.Slides {
		var warnings []string
		if i < len(perSlideWarnings) {
			warnings = perSlideWarnings[i]
		}
		var slideIssues []Issue
		slideIssues = append(slideIssues, Structural(slide)...)
		slideIssues = append(slideIssues, Fidelity(slide)...)
		// Variant-aware per-slide contrast (R7): use resolved engine colors when
		// available so dark slides are evaluated against their dark palette.
		if i < len(perSlideColors) {
			slideIssues = append(slideIssues, AuditSlideColors(perSlideColors[i])...)
		}
		slideIssues = append(slideIssues, OverflowIssues(warnings)...)
		perSlide[i] = Report{Score: Score(slideIssues), Issues: slideIssues}
		deckIssues = append(deckIssues, slideIssues...)
	}

	return Report{Score: Score(deckIssues), Issues: deckIssues}, perSlide
}
