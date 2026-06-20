package handlers

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// TestNodeToThumbRecursesGridOfCards proves Phase 12 D1: a Grid of Cards no
// longer flattens to a count — each card's header (Text) and body text survive
// as nested Children.
func TestNodeToThumbRecursesGridOfCards(t *testing.T) {
	grid := &contracts.Grid{
		Columns: 2,
		Cells: []contracts.SlideNode{
			&contracts.Card{Header: "Alpha", Body: []contracts.SlideNode{
				&contracts.Prose{Paragraphs: []contracts.RichText{rt("alpha body")}},
			}},
			&contracts.Card{Header: "Beta", Body: []contracts.SlideNode{
				&contracts.Prose{Paragraphs: []contracts.RichText{rt("beta body")}},
			}},
		},
	}

	thumb := nodeToThumb(grid)
	if thumb.Kind != "grid" {
		t.Fatalf("kind = %q, want grid", thumb.Kind)
	}
	if len(thumb.Children) != 2 {
		t.Fatalf("grid children = %d, want 2 (not flattened to a count)", len(thumb.Children))
	}

	card0 := asThumb(t, thumb.Children[0])
	if card0.Text != "Alpha" {
		t.Errorf("card[0].Text = %q, want Alpha (header dropped)", card0.Text)
	}
	if len(card0.Children) != 1 || asThumb(t, card0.Children[0]).Text != "alpha body" {
		t.Errorf("card[0] body = %+v, want one prose child with %q", card0.Children, "alpha body")
	}
	if asThumb(t, thumb.Children[1]).Text != "Beta" {
		t.Errorf("card[1].Text = %q, want Beta", asThumb(t, thumb.Children[1]).Text)
	}
}

// asThumb recovers the concrete ThumbNode from the []any Children slot (typed
// `any` only to satisfy the V1 schema generator — D-052).
func asThumb(t *testing.T, v any) contracts.ThumbNode {
	t.Helper()
	tn, ok := v.(contracts.ThumbNode)
	if !ok {
		t.Fatalf("child = %T, want contracts.ThumbNode", v)
	}
	return tn
}

// TestNodeToThumbTwoColumn proves the new TwoColumn case merges left+right
// children recursively.
func TestNodeToThumbTwoColumn(t *testing.T) {
	tc := &contracts.TwoColumn{
		Left:  []contracts.SlideNode{&contracts.Heading{Text: rt("L")}},
		Right: []contracts.SlideNode{&contracts.Heading{Text: rt("R")}},
	}
	thumb := nodeToThumb(tc)
	if len(thumb.Children) != 2 {
		t.Fatalf("two_column children = %d, want 2", len(thumb.Children))
	}
	if asThumb(t, thumb.Children[0]).Text != "L" || asThumb(t, thumb.Children[1]).Text != "R" {
		t.Errorf("children = [%q,%q], want [L,R]", asThumb(t, thumb.Children[0]).Text, asThumb(t, thumb.Children[1]).Text)
	}
}

// TestNodeToThumbListItems proves Phase 12 D3: list item text is carried (and
// capped) as Items.
func TestNodeToThumbListItems(t *testing.T) {
	list := &contracts.List{Items: []contracts.ListItem{
		{Text: rt("one")}, {Text: rt("two")}, {Text: rt("three")},
		{Text: rt("four")}, {Text: rt("five")},
	}}
	thumb := nodeToThumb(list)
	if thumb.Count != 5 {
		t.Errorf("count = %d, want 5", thumb.Count)
	}
	if len(thumb.Items) != thumbItemCap {
		t.Fatalf("items = %d, want capped at %d", len(thumb.Items), thumbItemCap)
	}
	if thumb.Items[0] != "one" || thumb.Items[3] != "four" {
		t.Errorf("items = %v, want first four item texts", thumb.Items)
	}
}

// TestNodeToThumbFlowSteps proves Phase 12 D5: flow step labels are carried as
// Items.
func TestNodeToThumbFlowSteps(t *testing.T) {
	flow := &contracts.Flow{Steps: []contracts.FlowStep{
		{Label: rt("Draft")}, {Label: rt("Review")}, {Label: rt("Ship")},
	}}
	thumb := nodeToThumb(flow)
	if thumb.Count != 3 {
		t.Errorf("count = %d, want 3", thumb.Count)
	}
	want := []string{"Draft", "Review", "Ship"}
	if len(thumb.Items) != len(want) {
		t.Fatalf("items = %v, want %v", thumb.Items, want)
	}
	for i, w := range want {
		if thumb.Items[i] != w {
			t.Errorf("items[%d] = %q, want %q", i, thumb.Items[i], w)
		}
	}
}

// TestNodeToThumbCalloutBody proves Phase 12 D4: the callout body lands in
// Detail.
func TestNodeToThumbCalloutBody(t *testing.T) {
	callout := &contracts.Callout{Title: "Heads up", Body: rt("the important detail")}
	thumb := nodeToThumb(callout)
	if thumb.Text != "Heads up" {
		t.Errorf("text = %q, want title", thumb.Text)
	}
	if thumb.Detail != "the important detail" {
		t.Errorf("detail = %q, want callout body", thumb.Detail)
	}
	if !thumb.Accent {
		t.Errorf("callout should be accent")
	}
}
