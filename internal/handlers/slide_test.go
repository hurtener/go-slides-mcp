package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

func TestSlideHandlersRoundTripThroughStore(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Roadmap"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	deckID := createdDeck.Structured.DeckID

	added, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: testSlide("Intro")})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}
	if !added.Structured.Validation.OK {
		t.Fatalf("addSlide validation = %+v, want OK", added.Structured.Validation)
	}

	got, err := h.getSlide(ctx, contracts.GetSlideInput{DeckID: deckID, SlideID: added.Structured.SlideID})
	if err != nil {
		t.Fatalf("getSlide: %v", err)
	}
	if got.Structured.SlideID != added.Structured.SlideID {
		t.Fatalf("getSlide id = %q, want %q", got.Structured.SlideID, added.Structured.SlideID)
	}

	updatedSlide := got.Structured.Slide
	updatedSlide.Nodes = append(updatedSlide.Nodes, &contracts.Callout{Kind: contracts.CalloutTip, Title: "Tip", Body: contracts.RichText{{Text: "Ship it"}}})
	storedDeck, err := h.deps.Store.GetDeck(deckID)
	if err != nil {
		t.Fatalf("store GetDeck: %v", err)
	}
	updated, err := h.updateSlide(ctx, contracts.UpdateSlideInput{DeckID: deckID, SlideID: added.Structured.SlideID, Slide: updatedSlide, ExpectedRevisionHash: storedDeck.Revision})
	if err != nil {
		t.Fatalf("updateSlide: %v", err)
	}
	if !updated.Structured.Validation.OK {
		t.Fatalf("updateSlide validation = %+v, want OK", updated.Structured.Validation)
	}

	second, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: testSlide("Second")})
	if err != nil {
		t.Fatalf("addSlide second: %v", err)
	}

	reordered, err := h.reorderSlides(ctx, contracts.ReorderSlidesInput{DeckID: deckID, Order: []string{second.Structured.SlideID, updated.Structured.SlideID}})
	if err != nil {
		t.Fatalf("reorderSlides: %v", err)
	}
	if len(reordered.Structured.Slides) != 2 || reordered.Structured.Slides[0].SlideID != second.Structured.SlideID {
		t.Fatalf("reorderSlides got %+v", reordered.Structured.Slides)
	}

	duplicated, err := h.duplicateSlide(ctx, contracts.DuplicateSlideInput{DeckID: deckID, SlideID: second.Structured.SlideID})
	if err != nil {
		t.Fatalf("duplicateSlide: %v", err)
	}
	if duplicated.Structured.SlideID == second.Structured.SlideID {
		t.Fatal("duplicateSlide returned original slide id")
	}

	removed, err := h.removeSlide(ctx, contracts.RemoveSlideInput{DeckID: deckID, SlideID: updated.Structured.SlideID})
	if err != nil {
		t.Fatalf("removeSlide: %v", err)
	}
	if !removed.Structured.Removed {
		t.Fatal("removeSlide Removed = false, want true")
	}
	if _, err := h.deps.Store.GetSlide(deckID, updated.Structured.SlideID); !errors.Is(err, deck.ErrNotFound) {
		t.Fatalf("store GetSlide after remove err = %v, want ErrNotFound", err)
	}
}

func TestSlideValidationIssuesSurface(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Validation"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}

	invalid := contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.List{Items: nil}}}
	added, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: createdDeck.Structured.DeckID, Slide: invalid})
	if err != nil {
		t.Fatalf("addSlide invalid: %v", err)
	}
	if added.Structured.Validation.OK {
		t.Fatal("addSlide validation OK = true, want false")
	}
	if len(added.Structured.Validation.Issues) == 0 {
		t.Fatal("addSlide validation issues empty, want at least one issue")
	}
}

func TestUpdateSlideRevisionConflict(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Conflicts"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	added, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: createdDeck.Structured.DeckID, Slide: testSlide("Current")})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}

	_, err = h.updateSlide(ctx, contracts.UpdateSlideInput{DeckID: createdDeck.Structured.DeckID, SlideID: added.Structured.SlideID, Slide: added.Structured.Slide, ExpectedRevisionHash: "stale-revision"})
	if !errors.Is(err, deck.ErrRevisionConflict) {
		t.Fatalf("updateSlide conflict err = %v, want ErrRevisionConflict", err)
	}
}
