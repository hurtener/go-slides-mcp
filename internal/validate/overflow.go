package validate

import "strings"

// OverflowIssues maps render-time layout warnings to validation issues. pptx-go
// warns (it never fails) on content overflow and unmapped layouts; Deckard
// surfaces those as spacing/structural findings so the agent can react.
//
// Overflow is the render-truth check that replaces the dropped Chromium Stage 2:
// instead of measuring a DOM, we read the warnings the native layout already
// emits.
func OverflowIssues(renderWarnings []string) []Issue {
	var issues []Issue
	for _, w := range renderWarnings {
		lw := strings.ToLower(w)
		switch {
		case strings.Contains(lw, "overflow"):
			issues = append(issues, Issue{
				Category: CategorySpacing,
				Severity: SeverityWarning,
				Message:  "layout overflow: " + w,
			})
		case strings.Contains(lw, "fell back to") || strings.Contains(lw, "unmapped") || strings.Contains(lw, "blank layout"):
			issues = append(issues, Issue{
				Category: CategoryStructural,
				Severity: SeverityWarning,
				Message:  "layout fallback: " + w,
			})
		default:
			issues = append(issues, Issue{
				Category: CategoryStructural,
				Severity: SeverityWarning,
				Message:  "render warning: " + w,
			})
		}
	}
	return issues
}
