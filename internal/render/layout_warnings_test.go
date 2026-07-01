package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// overflowingDoc returns a single-slide doc stacking far more Hero nodes
// (each a fixed 2.2in slot) than the body region can hold, so the engine's
// stackIn overflow check fires a "content overflows its region" LayoutWarning
// for the slide (R10.11 — the remediation ladder needs a real overflow signal
// to target).
func overflowingDoc() contracts.SlideDoc {
	nodes := make([]contracts.SlideNode, 0, 8)
	for i := 0; i < 8; i++ {
		nodes = append(nodes, &contracts.Hero{Title: "Overflow driver"})
	}
	return contracts.SlideDoc{
		Slides: []contracts.Slide{
			{ID: "overflow-slide", Layout: contracts.LayoutTitleContent, Nodes: nodes},
		},
	}
}

// TestRenderPopulatesLayoutWarningsForOverflow proves R10.11's additive
// Stats.LayoutWarnings field mirrors Warnings 1:1 and carries a non-empty
// SlideID/Message for a slide that overflows its region.
func TestRenderPopulatesLayoutWarningsForOverflow(t *testing.T) {
	t.Parallel()

	_, stats, err := Render(overflowingDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.Warnings) == 0 {
		t.Fatal("Render() produced no Warnings for a deliberately overflowing slide")
	}
	if len(stats.LayoutWarnings) != len(stats.Warnings) {
		t.Fatalf("Render() LayoutWarnings len = %d, want %d (mirrors Warnings 1:1)", len(stats.LayoutWarnings), len(stats.Warnings))
	}

	foundOverflow := false
	for _, w := range stats.LayoutWarnings {
		if w.SlideID == "" {
			t.Fatalf("Render() LayoutWarning has empty SlideID: %#v", w)
		}
		if w.Message == "" {
			t.Fatalf("Render() LayoutWarning has empty Message: %#v", w)
		}
		if w.SlideID == "overflow-slide" {
			foundOverflow = true
		}
	}
	if !foundOverflow {
		t.Fatalf("Render() LayoutWarnings did not carry the overflowing slide's ID: %#v", stats.LayoutWarnings)
	}
}

// TestRenderCleanDocHasNoLayoutWarnings asserts a simple, sparse,
// non-overflowing doc (a lone Hero slide) renders with an empty
// LayoutWarnings slice, so the field is not spuriously populated.
func TestRenderCleanDocHasNoLayoutWarnings(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{
			&contracts.Hero{Eyebrow: "Clean", Title: "No overflow here", Subtitle: "Just one node"},
		}},
	}}

	_, stats, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(stats.LayoutWarnings) != 0 {
		t.Fatalf("Render() LayoutWarnings = %#v, want empty for a non-overflowing doc", stats.LayoutWarnings)
	}
}
