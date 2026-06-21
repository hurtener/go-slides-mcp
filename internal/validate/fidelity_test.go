package validate

import (
	"math"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// richText is a small helper that builds a non-empty RichText.
func richText(s string) contracts.RichText {
	return contracts.RichText{{Text: s}}
}

func findIssuesWithMessage(issues []Issue, needle string) []Issue {
	var out []Issue
	for _, is := range issues {
		if strings.Contains(is.Message, needle) {
			out = append(out, is)
		}
	}
	return out
}

// TestFidelityFlowAllEmptyScoresBelowOne locks C1+C3: a Flow with five empty
// FlowSteps yields one SeverityError (mentioning label+detail) and five
// per-step warnings (each at the steps[N].label path), and the aggregated
// StyleScore drops below 1.0 with Passed=false.
func TestFidelityFlowAllEmptyScoresBelowOne(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Flow{Steps: []contracts.FlowStep{
				{}, {}, {}, {}, {},
			}},
		},
	}
	report := Score(Fidelity(slide))
	if report.Passed {
		t.Fatal("all-empty Flow must flip Passed=false")
	}
	if report.Score >= 1.0 {
		t.Fatalf("all-empty Flow score = %.4f, want < 1.0", report.Score)
	}
	issues := Fidelity(slide)
	var ws, es int
	var errMsg string
	for _, is := range issues {
		if is.Category != CategoryContent {
			t.Fatalf("issue category = %q, want content", is.Category)
		}
		switch is.Severity {
		case SeverityWarning:
			ws++
		case SeverityError:
			es++
			errMsg = is.Message
		}
	}
	if ws != 5 {
		t.Fatalf("warnings = %d, want 5", ws)
	}
	if es != 1 {
		t.Fatalf("errors = %d, want 1 (wholesale-empty Flow)", es)
	}
	if !strings.Contains(errMsg, "label") || !strings.Contains(errMsg, "detail") {
		t.Fatalf("wholesale-error message = %q, want it to mention label and detail", errMsg)
	}
}

// TestFidelityFlowPartialEmptyStaysPassing: one empty step among four is a
// single warning; Ok must stay true (per-empty is a warning, not an error).
func TestFidelityFlowPartialEmptyStaysPassing(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Flow{Steps: []contracts.FlowStep{
				{Label: richText("a")}, {}, {Label: richText("b")}, {Label: richText("c")},
			}},
		},
	}
	issues := Fidelity(slide)
	if len(issues) != 1 || issues[0].Severity != SeverityWarning {
		t.Fatalf("partial-empty Flow issues = %+v, want 1 warning", issues)
	}
	if !strings.Contains(issues[0].Path, "steps[1].label") {
		t.Fatalf("warning path = %q, want steps[1].label", issues[0].Path)
	}
	if Score(issues).Passed != true {
		t.Fatal("partial-empty Flow must still pass (warnings only)")
	}
}

// TestFidelityEmptyPerKind checks C2: each kind-level leaf emptiness surfaces
// as a finding whose Path points at the right node field.
func TestFidelityEmptyPerKind(t *testing.T) {
	cases := []struct {
		name  string
		slide contracts.Slide
		want  string
	}{
		{
			name: "empty heading",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 1},
			}},
			want: "nodes[0].text",
		},
		{
			name: "empty list item",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{
					{Text: richText("real")}, {},
				}},
			}},
			want: "nodes[0].items[1].text",
		},
		{
			name: "empty callout body",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.Callout{Kind: contracts.CalloutNote, Title: "Hello"},
			}},
			want: "nodes[0].body",
		},
		{
			name: "empty table cell",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.Table{
					Headers: []contracts.RichText{richText("col")},
					Rows:    [][]contracts.RichText{{richText("ok"), {}}},
				},
			}},
			want: "nodes[0].rows[0][1]",
		},
		{
			name: "empty prose paragraph",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.Prose{Paragraphs: []contracts.RichText{richText("ok"), {}}},
			}},
			want: "nodes[0].paragraphs[1]",
		},
		{
			name: "empty flow step",
			slide: contracts.Slide{Nodes: []contracts.SlideNode{
				&contracts.Flow{Steps: []contracts.FlowStep{
					{Label: richText("ok")}, {},
				}},
			}},
			want: "nodes[0].steps[1].label",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issues := Fidelity(tc.slide)
			if len(issues) == 0 {
				t.Fatalf("Fidelity(%s): no findings", tc.name)
			}
			found := findIssuesWithMessage(issues, "empty content")
			if len(found) == 0 {
				t.Fatalf("no empty-content findings: %+v", issues)
			}
			foundPath := false
			for _, is := range found {
				if is.Path == tc.want {
					foundPath = true
					break
				}
			}
			if !foundPath {
				t.Fatalf("no finding with Path=%q; got %+v", tc.want, issues)
			}
		})
	}
}

// TestFidelityPopulatedSlideHasNoFindings: a fully-populated slide yields ZERO
// fidelity findings (no false positives).
func TestFidelityPopulatedSlideHasNoFindings(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 1, Text: richText("Intro")},
			&contracts.Prose{Paragraphs: []contracts.RichText{
				richText("body line 1"), richText("body line 2"),
			}},
			&contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{
				{Text: richText("first")}, {Text: richText("second")},
			}},
			&contracts.Callout{Kind: contracts.CalloutNote, Title: "Note", Body: richText("be careful")},
			&contracts.Flow{Steps: []contracts.FlowStep{
				{Label: richText("Step 1"), Detail: richText("do thing")},
				{Label: richText("Step 2"), Detail: richText("do other")},
			}},
			&contracts.Table{
				Headers: []contracts.RichText{richText("A"), richText("B")},
				Rows: [][]contracts.RichText{
					{richText("a1"), richText("b1")},
				},
			},
		},
	}
	issues := Fidelity(slide)
	if len(issues) != 0 {
		t.Fatalf("populated slide must yield zero fidelity findings, got %+v", issues)
	}
}

// TestFidelityCardGridEmptyCells covers C4: a Grid with all-empty Cards
// surfaces a wholesale-empty error (which flips Passed).
func TestFidelityCardGridEmptyCells(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Card{}, &contracts.Card{},
			}},
		},
	}
	report := Score(Fidelity(slide))
	if report.Passed {
		t.Fatal("all-empty Grid must flip Passed=false")
	}
	issues := Fidelity(slide)
	errs := findIssuesWithMessage(issues, "columns")
	if len(errs) == 0 {
		t.Fatalf("wholesale grid error message missing: %+v", issues)
	}
}

// TestFidelitySingleEmptyCardGridCell: one empty cell among two is a warning
// (per-cell), no wholesale error.
func TestFidelitySingleEmptyCardGridCell(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Card{Header: "ok"}, &contracts.Card{},
			}},
		},
	}
	issues := Fidelity(slide)
	if len(issues) != 1 {
		t.Fatalf("single empty cell issues = %+v, want exactly 1 warning", issues)
	}
	if issues[0].Severity != SeverityWarning {
		t.Fatalf("cell issue severity = %q, want warning", issues[0].Severity)
	}
	if issues[0].Path != "nodes[0].cells[1]" {
		t.Fatalf("cell issue path = %q", issues[0].Path)
	}
	if Score(issues).Passed != true {
		t.Fatal("single empty cell must still pass (warnings only)")
	}
}

// TestFidelityDocPrefixesPaths locks FidelityDoc's path prefix so multi-slide
// messages are unambiguous.
func TestFidelityDocPrefixesPaths(t *testing.T) {
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 1},
		}},
	}}
	issues := FidelityDoc(doc)
	if len(issues) == 0 {
		t.Fatal("FidelityDoc: no findings")
	}
	if !strings.Contains(issues[0].Path, "slides[0]") {
		t.Fatalf("FidelityDoc path %q missing slides[0] prefix", issues[0].Path)
	}
}

// TestFidelityConnectedToScore: re-derive the wiring so a future weight bump
// can't quietly drop the fidelity category (C6 plumbing).
func TestFidelityConnectedToScore(t *testing.T) {
	slide := contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Flow{Steps: []contracts.FlowStep{{}, {}, {}, {}, {}}},
		},
	}
	// Build a full Slide() report (Slide uses no theme + no render warnings).
	report := Slide(slide, nil, nil, nil)
	if report.Score.Passed {
		t.Fatal("Slide() with all-empty flow must flip Passed=false")
	}
	// 1 wholesale error + 5 per-step warnings => Score should differ from
	// 1.0 by notice. With content weight 0.14: 0.14*0.20 + 5*0.14*0.05 = 0.063.
	// Score = 1 - 0.063 = 0.937. Confirm it's < 0.94 (loose bound to survive
	// future weight rebalancing toward the C-binding behavior of <1.0).
	if math.Abs(report.Score.Score-0.937) > 1e-9 {
		t.Fatalf("Score = %.4f, want 0.937", report.Score.Score)
	}
	if report.Score.ByCategory[CategoryContent] >= 1.0 {
		t.Fatalf("content subscore = %.4f, want < 1.0", report.Score.ByCategory[CategoryContent])
	}
}
