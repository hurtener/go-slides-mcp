package validate

import (
	"fmt"
	"strings"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// Fidelity walks a slide's IR and emits issues for every content-bearing leaf
// whose RichText is empty, recursing into container bodies. A wholesale-empty
// REPEATING node (all steps empty, all items empty, all cells empty) emits
// one extra Severity ERROR naming the correct JSON shape so the failure
// doubles as discoverability (docs/PHASE-12-FIDELITY.md C1-C4).
func Fidelity(slide contracts.Slide) []Issue {
	var issues []Issue
	for i, n := range slide.Nodes {
		walkNode(i, n, &issues)
	}
	return issues
}

// FidelityDoc walks every slide in a deck, prefixing node paths with the
// slide's container label so messages are unambiguous in a multi-slide deck.
func FidelityDoc(doc contracts.SlideDoc) []Issue {
	var issues []Issue
	for i, s := range doc.Slides {
		for j, n := range s.Nodes {
			walkNodeAt(fmt.Sprintf("slides[%d].nodes[%d]", i, j), n, &issues)
		}
	}
	return issues
}

// richTextEmpty is the emptiness primitive (docs/PHASE-12-FIDELITY.md C2).
func richTextEmpty(rt contracts.RichText) bool {
	return len(rt) == 0 || strings.TrimSpace(rt.PlainText()) == ""
}

// leafWarn emits an empty-leaf warning at the given node path.
func leafWarn(issues *[]Issue, path, label string) {
	*issues = append(*issues, Issue{
		Category: CategoryContent,
		Severity: SeverityWarning,
		Path:     path,
		Message:  fmt.Sprintf("%s: empty content (RichText missing or whitespace-only)", label),
	})
}

// wholesaleError emits the blocking error for an ALL-empty repeating
// container, naming the correct wire shape so the agent can self-correct.
func wholesaleError(issues *[]Issue, path, node, shape string) {
	*issues = append(*issues, Issue{
		Category: CategoryContent,
		Severity: SeverityError,
		Path:     path,
		Message: fmt.Sprintf(
			"%s at %s: no content in any entry — step is the leaf, you must nest the "+
				"shape %s inside each entry; verify the node's known fields.", node, path, shape),
	})
}

// walkNode walks a top-level slide child at the given prefix.
func walkNode(idx int, n contracts.SlideNode, issues *[]Issue) {
	walkNodeAt(fmt.Sprintf("nodes[%d]", idx), n, issues)
}

// walkNodeAt walks a node anchored at an arbitrary path prefix.
func walkNodeAt(path string, n contracts.SlideNode, issues *[]Issue) {
	switch v := n.(type) {
	case *contracts.Heading:
		if richTextEmpty(v.Text) {
			leafWarn(issues, path+".text", "heading")
		}
	case *contracts.Prose:
		for i, p := range v.Paragraphs {
			if richTextEmpty(p) {
				leafWarn(issues, fmt.Sprintf("%s.paragraphs[%d]", path, i), "prose paragraph")
			}
		}
	case *contracts.Quote:
		if richTextEmpty(v.Text) {
			leafWarn(issues, path+".text", "quote")
		}
	case *contracts.Callout:
		if richTextEmpty(v.Body) {
			leafWarn(issues, path+".body", "callout")
		}
	case *contracts.List:
		walkList(v, path, issues)
	case *contracts.Table:
		walkTable(v, path, issues)
	case *contracts.Flow:
		walkFlow(v, path, issues)
	case *contracts.Card:
		for i, child := range v.Body {
			walkNodeAt(fmt.Sprintf("%s.body[%d]", path, i), child, issues)
		}
	case *contracts.CardSection:
		for i, child := range v.Body {
			walkNodeAt(fmt.Sprintf("%s.body[%d]", path, i), child, issues)
		}
	case *contracts.Grid:
		walkGrid(v, path, issues)
	case *contracts.TwoColumn:
		for i, child := range v.Left {
			walkNodeAt(fmt.Sprintf("%s.left[%d]", path, i), child, issues)
		}
		for i, child := range v.Right {
			walkNodeAt(fmt.Sprintf("%s.right[%d]", path, i), child, issues)
		}
	}
}

// walkList walks a List, accumulating per-item labels. If every item carries
// no text, the list itself is a wholesale empty repeater (ERROR + shape hint).
func walkList(v *contracts.List, path string, issues *[]Issue) {
	allEmpty := len(v.Items) > 0
	for i, it := range v.Items {
		if richTextEmpty(it.Text) {
			leafWarn(issues, fmt.Sprintf("%s.items[%d].text", path, i), "list item")
		} else {
			allEmpty = false
		}
	}
	if allEmpty {
		wholesaleError(issues, path, "list",
			`flat {"listKind":"bullet|number|checklist","items":[{"text":<RichText>}]}`)
	}
}

// walkTable walks a Table; every empty header cell or body cell is a
// warning. Header-less tables are rejected by structural validation
// upstream so we only inspect what is present.
func walkTable(v *contracts.Table, path string, issues *[]Issue) {
	for i, h := range v.Headers {
		if richTextEmpty(h) {
			leafWarn(issues, fmt.Sprintf("%s.headers[%d]", path, i), "table header")
		}
	}
	for r, row := range v.Rows {
		for c, cell := range row {
			if richTextEmpty(cell) {
				leafWarn(issues, fmt.Sprintf("%s.rows[%d][%d]", path, r, c), "table cell")
			}
		}
	}
}

// walkFlow walks a Flow. Each empty step's label is a warning. A Flow in
// which EVERY step is empty emits one ERROR naming the correct flat step
// shape — the highest-leverage recoverability hint in this validator.
func walkFlow(v *contracts.Flow, path string, issues *[]Issue) {
	if len(v.Steps) == 0 {
		return // structural validator already flagged; no fidelity checks apply.
	}
	allEmpty := true
	for i, s := range v.Steps {
		if richTextEmpty(s.Label) {
			leafWarn(issues, fmt.Sprintf("%s.steps[%d].label", path, i), "flow step")
		} else {
			allEmpty = false
		}
	}
	if allEmpty {
		wholesaleError(issues, path, "flow",
			`flat {"label":<RichText>,"detail":<RichText>} (and "icon"? optional)`)
	}
}

// walkGrid walks a Grid's cells, recursing into each. A grid whose every
// cell has no renderable content (after recursion) is wholesale-empty. Each
// empty cell is also signalled per-cell so the author can locate every gap.
func walkGrid(v *contracts.Grid, path string, issues *[]Issue) {
	if len(v.Cells) == 0 {
		return
	}
	allEmpty := true
	for i, c := range v.Cells {
		sub := path + ".cells[" + fmt.Sprintf("%d", i) + "]"
		walkNodeAt(sub, c, issues)
		if cellIsEmpty(c) {
			leafWarn(issues, sub, "grid cell")
		} else {
			allEmpty = false
		}
	}
	if allEmpty {
		wholesaleError(issues, path, "grid",
			`{"columns":<2..4>,"cells":[<node>,<node>,...]}`)
	}
}

// cellIsEmpty reports whether a grid cell carries no renderable content.
// Mirrors the rules the inspector's "no body" cases flag; this is the
// recurring dispatch the fidelity pass uses.
func cellIsEmpty(n contracts.SlideNode) bool {
	switch v := n.(type) {
	case *contracts.Card:
		return v.Header == "" && len(v.Body) == 0
	case *contracts.CardSection:
		return v.Header == "" && len(v.Body) == 0
	case *contracts.Heading:
		return richTextEmpty(v.Text)
	case *contracts.Prose:
		if len(v.Paragraphs) == 0 {
			return true
		}
		for _, p := range v.Paragraphs {
			if !richTextEmpty(p) {
				return false
			}
		}
		return true
	case *contracts.Quote:
		return richTextEmpty(v.Text)
	case *contracts.Callout:
		return richTextEmpty(v.Body) && v.Title == ""
	case *contracts.List:
		return len(v.Items) == 0
	case *contracts.Flow:
		return len(v.Steps) == 0
	case *contracts.Table:
		return len(v.Headers) == 0 && len(v.Rows) == 0
	}
	return false
}
