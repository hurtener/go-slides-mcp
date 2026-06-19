package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

func TestEditSlideNodePersistsChange(t *testing.T) {
	h, deckID, slideID, revision := setupEditableSlide(t)

	raw := mustNodeMap(t, &contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Updated title"}}})
	got, err := h.editSlideNode(context.Background(), contracts.EditSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 0}, Node: raw, ExpectedRevisionHash: revision})
	if err != nil {
		t.Fatalf("editSlideNode: %v", err)
	}
	if !got.Structured.Validation.OK {
		t.Fatalf("editSlideNode validation = %+v, want OK", got.Structured.Validation)
	}
	assertHeadingText(t, got.Structured.Slide.Nodes[0], "Updated title")

	persisted, err := h.deps.Store.GetSlide(deckID, slideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	assertHeadingText(t, persisted.Nodes[0], "Updated title")
}

func TestNodeEditHandlersMutateTree(t *testing.T) {
	h, deckID, slideID, revision := setupEditableSlide(t)
	ctx := context.Background()

	inserted, err := h.insertSlideNode(ctx, contracts.InsertSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 1}, Node: mustNodeMap(t, &contracts.Callout{Kind: contracts.CalloutTip, Title: "Tip", Body: contracts.RichText{{Text: "Inserted"}}}), ExpectedRevisionHash: revision})
	if err != nil {
		t.Fatalf("insertSlideNode: %v", err)
	}
	if len(inserted.Structured.Slide.Nodes) != 3 {
		t.Fatalf("insertSlideNode node count = %d, want 3", len(inserted.Structured.Slide.Nodes))
	}
	revision = deckRevision(t, h, deckID)

	removed, err := h.removeSlideNode(ctx, contracts.RemoveSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 1}, ExpectedRevisionHash: revision})
	if err != nil {
		t.Fatalf("removeSlideNode: %v", err)
	}
	if len(removed.Structured.Slide.Nodes) != 2 {
		t.Fatalf("removeSlideNode node count = %d, want 2", len(removed.Structured.Slide.Nodes))
	}
	revision = deckRevision(t, h, deckID)

	duplicated, err := h.duplicateSlideNode(ctx, contracts.DuplicateSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 0}, ExpectedRevisionHash: revision})
	if err != nil {
		t.Fatalf("duplicateSlideNode: %v", err)
	}
	if len(duplicated.Structured.Slide.Nodes) != 3 {
		t.Fatalf("duplicateSlideNode node count = %d, want 3", len(duplicated.Structured.Slide.Nodes))
	}
	assertHeadingText(t, duplicated.Structured.Slide.Nodes[1], "Intro")
	revision = deckRevision(t, h, deckID)

	moved, err := h.moveSlideNode(ctx, contracts.MoveSlideNodeInput{DeckID: deckID, SlideID: slideID, From: contracts.IRPath{"nodes", 2}, To: contracts.IRPath{"nodes", 0}, ExpectedRevisionHash: revision})
	if err != nil {
		t.Fatalf("moveSlideNode: %v", err)
	}
	assertProseBody(t, moved.Structured.Slide.Nodes[0], "Body")
	assertHeadingText(t, moved.Structured.Slide.Nodes[1], "Intro")
}

func TestEditSlideNodeBadPathReturnsError(t *testing.T) {
	h, deckID, slideID, revision := setupEditableSlide(t)

	_, err := h.editSlideNode(context.Background(), contracts.EditSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 99}, Node: mustNodeMap(t, &contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Nope"}}}), ExpectedRevisionHash: revision})
	if err == nil {
		t.Fatal("editSlideNode error = nil, want error")
	}
}

func TestEditSlideNodeRevisionConflict(t *testing.T) {
	h, deckID, slideID, _ := setupEditableSlide(t)

	_, err := h.editSlideNode(context.Background(), contracts.EditSlideNodeInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 0}, Node: mustNodeMap(t, &contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Conflict"}}}), ExpectedRevisionHash: "stale-revision"})
	if !errors.Is(err, deck.ErrRevisionConflict) {
		t.Fatalf("editSlideNode conflict err = %v, want ErrRevisionConflict", err)
	}
}

func setupEditableSlide(t *testing.T) (*handlers, string, string, string) {
	t.Helper()
	h := testHandlers()
	createdDeck, err := h.createDeck(context.Background(), contracts.CreateDeckInput{Title: "Editable"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	added, err := h.addSlide(context.Background(), contracts.AddSlideInput{DeckID: createdDeck.Structured.DeckID, Slide: testSlide("Intro")})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}
	return h, createdDeck.Structured.DeckID, added.Structured.SlideID, deckRevision(t, h, createdDeck.Structured.DeckID)
}

func deckRevision(t *testing.T, h *handlers, deckID string) string {
	t.Helper()
	stored, err := h.deps.Store.GetDeck(deckID)
	if err != nil {
		t.Fatalf("store GetDeck: %v", err)
	}
	return stored.Revision
}

func mustNodeMap(t *testing.T, node contracts.SlideNode) map[string]any {
	t.Helper()
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("json.Marshal node: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal node to map: %v", err)
	}
	return m
}

func assertHeadingText(t *testing.T, node contracts.SlideNode, want string) {
	t.Helper()
	heading, ok := node.(*contracts.Heading)
	if !ok {
		t.Fatalf("node type = %T, want *contracts.Heading", node)
	}
	if got := richTextString(heading.Text); got != want {
		t.Fatalf("heading text = %q, want %q", got, want)
	}
}

func assertProseBody(t *testing.T, node contracts.SlideNode, want string) {
	t.Helper()
	prose, ok := node.(*contracts.Prose)
	if !ok {
		t.Fatalf("node type = %T, want *contracts.Prose", node)
	}
	if len(prose.Paragraphs) == 0 {
		t.Fatal("prose paragraphs empty")
	}
	if got := richTextString(prose.Paragraphs[0]); got != want {
		t.Fatalf("prose text = %q, want %q", got, want)
	}
}
