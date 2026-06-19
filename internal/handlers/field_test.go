package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

// addCalloutSlide adds a slide whose node[0] is a Callout (it has the string
// field "title" that edit_slide_field targets).
func addCalloutSlide(t *testing.T, h *handlers, deckID string) (slideID, revision string) {
	t.Helper()
	slide := contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
		&contracts.Callout{Kind: contracts.CalloutTip, Title: "Old title", Body: contracts.RichText{{Text: "Body"}}},
	}}
	added, err := h.addSlide(context.Background(), contracts.AddSlideInput{DeckID: deckID, Slide: slide})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}
	return added.Structured.SlideID, deckRevision(t, h, deckID)
}

func TestEditSlideFieldPersistsChange(t *testing.T) {
	h := testHandlers()
	created, err := h.createDeck(context.Background(), contracts.CreateDeckInput{Title: "Editable"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	deckID := created.Structured.DeckID
	slideID, revision := addCalloutSlide(t, h, deckID)

	got, err := h.editSlideField(context.Background(), contracts.EditSlideFieldInput{
		DeckID:               deckID,
		SlideID:              slideID,
		Path:                 contracts.IRPath{"nodes", 0},
		Field:                "title",
		Value:                "New title",
		ExpectedRevisionHash: revision,
	})
	if err != nil {
		t.Fatalf("editSlideField: %v", err)
	}
	assertCalloutTitle(t, got.Structured.Slide.Nodes[0], "New title")

	persisted, err := h.deps.Store.GetSlide(deckID, slideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	assertCalloutTitle(t, persisted.Nodes[0], "New title")
}

func TestPatchSlideTextPersistsChange(t *testing.T) {
	h, deckID, slideID, revision := setupEditableSlide(t)

	got, err := h.patchSlideText(context.Background(), contracts.PatchSlideTextInput{
		DeckID:               deckID,
		SlideID:              slideID,
		Path:                 contracts.IRPath{"nodes", 0},
		Field:                "text",
		Text:                 "Patched title",
		ExpectedRevisionHash: revision,
	})
	if err != nil {
		t.Fatalf("patchSlideText: %v", err)
	}
	if !got.Structured.Validation.OK {
		t.Fatalf("patchSlideText validation = %+v, want OK", got.Structured.Validation)
	}
	assertHeadingText(t, got.Structured.Slide.Nodes[0], "Patched title")

	persisted, err := h.deps.Store.GetSlide(deckID, slideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	assertHeadingText(t, persisted.Nodes[0], "Patched title")
}

func TestEditSlideFieldBadPathReturnsError(t *testing.T) {
	h := testHandlers()
	created, _ := h.createDeck(context.Background(), contracts.CreateDeckInput{Title: "Editable"})
	deckID := created.Structured.DeckID
	slideID, revision := addCalloutSlide(t, h, deckID)

	_, err := h.editSlideField(context.Background(), contracts.EditSlideFieldInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 99}, Field: "title", Value: "x", ExpectedRevisionHash: revision})
	if err == nil {
		t.Fatal("editSlideField error = nil, want error")
	}
}

func TestPatchSlideTextRevisionConflict(t *testing.T) {
	h, deckID, slideID, _ := setupEditableSlide(t)

	_, err := h.patchSlideText(context.Background(), contracts.PatchSlideTextInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 0}, Field: "text", Text: "Conflict", ExpectedRevisionHash: "stale-revision"})
	if !errors.Is(err, deck.ErrRevisionConflict) {
		t.Fatalf("patchSlideText conflict err = %v, want ErrRevisionConflict", err)
	}
}

func assertCalloutTitle(t *testing.T, node contracts.SlideNode, want string) {
	t.Helper()
	callout, ok := node.(*contracts.Callout)
	if !ok {
		t.Fatalf("node type = %T, want *contracts.Callout", node)
	}
	if callout.Title != want {
		t.Fatalf("callout title = %q, want %q", callout.Title, want)
	}
}
