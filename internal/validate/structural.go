package validate

import (
	"strings"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
)

// Structural runs the IR structural validator over a slide and maps each error
// to a blocking structural Issue.
func Structural(slide contracts.Slide) []Issue {
	return fromError(ir.ValidateSlide(slide))
}

// StructuralDoc runs the IR structural validator over a whole document.
func StructuralDoc(doc contracts.SlideDoc) []Issue {
	return fromError(ir.ValidateDoc(doc))
}

func fromError(err error) []Issue {
	msgs := flatten(err)
	if len(msgs) == 0 {
		return nil
	}
	issues := make([]Issue, 0, len(msgs))
	for _, m := range msgs {
		issues = append(issues, Issue{
			Category: CategoryStructural,
			Severity: SeverityError,
			Message:  m,
		})
	}
	return issues
}

// flatten unwraps a joined error (errors.Join) into individual messages.
func flatten(err error) []string {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var out []string
		for _, item := range joined.Unwrap() {
			out = append(out, flatten(item)...)
		}
		return out
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return nil
	}
	return []string{msg}
}
