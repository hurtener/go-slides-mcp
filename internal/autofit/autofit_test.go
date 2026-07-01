package autofit

import (
	"reflect"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// slideWithNodes builds a minimal slide carrying the given top-level nodes,
// with no explicit Align — the shape Fill is meant to act on.
func slideWithNodes(id string, nodes ...contracts.SlideNode) contracts.Slide {
	return contracts.Slide{ID: id, Nodes: nodes}
}

// TestFillAppliesToEachFillableContainerKind asserts every fillable
// container kind (grid, two_column, bento, table, card_section) at the top
// level triggers VAlignFill on a slide with no explicit alignment.
func TestFillAppliesToEachFillableContainerKind(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		node contracts.SlideNode
	}{
		{"grid", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Stat{Value: "1"}, &contracts.Stat{Value: "2"}}}},
		{"two_column", &contracts.TwoColumn{Left: []contracts.SlideNode{&contracts.Prose{}}, Right: []contracts.SlideNode{&contracts.Prose{}}}},
		{"bento", &contracts.Bento{Columns: 2}},
		{"table", &contracts.Table{Headers: []contracts.RichText{{{Text: "A"}}}}},
		{"card_section", &contracts.CardSection{Header: "Section", Body: []contracts.SlideNode{&contracts.Prose{}}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := contracts.SlideDoc{Slides: []contracts.Slide{
				slideWithNodes("s1", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "Title"}}}, tc.node),
			}}
			got := Fill(doc)
			if got.Slides[0].Align.Vertical != contracts.VAlignFill {
				t.Fatalf("Fill() Align.Vertical = %q, want %q", got.Slides[0].Align.Vertical, contracts.VAlignFill)
			}
		})
	}
}

// TestFillLeavesUnfillableSlidesUntouched asserts slides with no fillable
// container at the top level — a hero-only cover, a prose+heading slide, and
// a lone single card — are left with Align.Vertical == "" (so a sparse lone
// node is never ballooned).
func TestFillLeavesUnfillableSlidesUntouched(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		slide contracts.Slide
	}{
		{
			name:  "hero only",
			slide: slideWithNodes("hero", &contracts.Hero{Title: "Welcome"}),
		},
		{
			name: "prose plus heading",
			slide: slideWithNodes("prose",
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Heading"}}},
				&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Body"}}}},
			),
		},
		{
			name:  "lone single card",
			slide: slideWithNodes("card", &contracts.Card{Header: "Solo"}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := contracts.SlideDoc{Slides: []contracts.Slide{tc.slide}}
			got := Fill(doc)
			if got.Slides[0].Align.Vertical != "" {
				t.Fatalf("Fill() Align.Vertical = %q, want empty (untouched)", got.Slides[0].Align.Vertical)
			}
		})
	}
}

// TestFillRespectsExplicitAlignment asserts a slide with an explicit
// Align.Vertical already set is left untouched — explicit wins over the
// auto-fill heuristic even when a fillable container is present.
func TestFillRespectsExplicitAlignment(t *testing.T) {
	t.Parallel()

	slide := slideWithNodes("explicit", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Stat{Value: "1"}}})
	slide.Align.Vertical = contracts.VAlignCenter
	doc := contracts.SlideDoc{Slides: []contracts.Slide{slide}}

	got := Fill(doc)
	if got.Slides[0].Align.Vertical != contracts.VAlignCenter {
		t.Fatalf("Fill() Align.Vertical = %q, want unchanged %q", got.Slides[0].Align.Vertical, contracts.VAlignCenter)
	}
}

// TestFillDoesNotMutateInput asserts the original doc's slide is unaffected
// after Fill — the transform copies the Slides slice (and each Slide is a
// value type for Align) rather than mutating in place.
func TestFillDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		slideWithNodes("s1", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Stat{Value: "1"}}}),
	}}

	_ = Fill(doc)

	if doc.Slides[0].Align.Vertical != "" {
		t.Fatalf("Fill() mutated input doc: Align.Vertical = %q, want empty", doc.Slides[0].Align.Vertical)
	}
}

// TestFillIsDeterministic asserts two Fill calls over the same input produce
// equal output docs.
func TestFillIsDeterministic(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		slideWithNodes("s1", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Stat{Value: "1"}}}),
		slideWithNodes("s2", &contracts.Hero{Title: "Cover"}),
	}}

	got1 := Fill(doc)
	got2 := Fill(doc)
	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("Fill() is not deterministic:\nfirst =%#v\nsecond=%#v", got1, got2)
	}
}
