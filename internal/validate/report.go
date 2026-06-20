package validate

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
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

// Slide scores a single slide: structural IR checks, soul-theme contrast,
// content-fidelity (empty RichText leaves), and layout-overflow warnings
// (renderWarnings comes from a render of that slide; pass nil to skip the
// render-truth pass). theme may be nil (contrast skipped).
func Slide(slide contracts.Slide, theme *pptx.Theme, renderWarnings []string) Report {
	var issues []Issue
	issues = append(issues, Structural(slide)...)
	issues = append(issues, Fidelity(slide)...)
	issues = append(issues, AuditTheme(theme)...)
	issues = append(issues, OverflowIssues(renderWarnings)...)
	return Report{Score: Score(issues), Issues: issues}
}

// Deck scores a whole deck for export. Contrast is a soul-level audit run once;
// structural, fidelity, and overflow are per-slide. perSlideWarnings is keyed
// by slide index in doc order (pass nil to skip the render-truth pass).
func Deck(doc contracts.SlideDoc, theme *pptx.Theme, perSlideWarnings [][]string) (Report, []Report) {
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
		slideIssues = append(slideIssues, OverflowIssues(warnings)...)
		perSlide[i] = Report{Score: Score(slideIssues), Issues: slideIssues}
		deckIssues = append(deckIssues, slideIssues...)
	}

	return Report{Score: Score(deckIssues), Issues: deckIssues}, perSlide
}
