package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

func TestEditSlideFieldPersistsChange(t *testing.T) {
	h, deckID, slideID, revision := setupEditableSlide(t)

	got, err := h.editSlideField(context.Background(), contracts.EditSlideFieldInput{
		DeckID:               deckID,
		SlideID:              slideID,
		Path:                 contracts.IRPath{"nodes", 0},
		Field:                "level",
		Value:                json.RawMessage(`3`),
		ExpectedRevisionHash: revision,
	})
	if err != nil {
		t.Fatalf("editSlideField: %v", err)
	}
	if !got.Structured.Validation.OK {
		t.Fatalf("editSlideField validation = %+v, want OK", got.Structured.Validation)
	}
	assertHeadingLevel(t, got.Structured.Slide.Nodes[0], 3)

	persisted, err := h.deps.Store.GetSlide(deckID, slideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	assertHeadingLevel(t, persisted.Nodes[0], 3)
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
	h, deckID, slideID, revision := setupEditableSlide(t)

	_, err := h.editSlideField(context.Background(), contracts.EditSlideFieldInput{DeckID: deckID, SlideID: slideID, Path: contracts.IRPath{"nodes", 99}, Field: "level", Value: json.RawMessage(`2`), ExpectedRevisionHash: revision})
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

func assertHeadingLevel(t *testing.T, node contracts.SlideNode, want int) {
	t.Helper()
	heading, ok := node.(*contracts.Heading)
	if !ok {
		t.Fatalf("node type = %T, want *contracts.Heading", node)
	}
	if heading.Level != want {
		t.Fatalf("heading level = %d, want %d", heading.Level, want)
	}
}
