package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/ir"
)

func (h *handlers) addSlide(_ context.Context, in contracts.AddSlideInput) (tool.Result[contracts.AddSlideOutput], error) {
	_, slide, err := h.deps.Store.AddSlide(in.DeckID, in.Slide, in.Position)
	if err != nil {
		return tool.Result[contracts.AddSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	validation := validateSlide(*slide)
	out := contracts.AddSlideOutput{SlideID: slide.ID, Slide: *slide, Validation: validation}
	return tool.Result[contracts.AddSlideOutput]{Text: fmt.Sprintf("Added slide %q to deck %q.", slide.ID, in.DeckID), Structured: out}, nil
}

func (h *handlers) updateSlide(_ context.Context, in contracts.UpdateSlideInput) (tool.Result[contracts.UpdateSlideOutput], error) {
	_, slide, err := h.deps.Store.UpdateSlide(in.DeckID, in.SlideID, in.Slide, in.ExpectedRevisionHash)
	if err != nil {
		return tool.Result[contracts.UpdateSlideOutput]{}, mapSlideMutationError(in.DeckID, in.SlideID, err)
	}
	validation := validateSlide(*slide)
	out := contracts.UpdateSlideOutput{SlideID: slide.ID, Slide: *slide, Validation: validation}
	return tool.Result[contracts.UpdateSlideOutput]{Text: fmt.Sprintf("Updated slide %q in deck %q.", slide.ID, in.DeckID), Structured: out}, nil
}

func (h *handlers) getSlide(_ context.Context, in contracts.GetSlideInput) (tool.Result[contracts.GetSlideOutput], error) {
	slide, err := h.deps.Store.GetSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.GetSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	validation := validateSlide(*slide)
	out := contracts.GetSlideOutput{SlideID: slide.ID, Slide: *slide, Validation: validation}
	return tool.Result[contracts.GetSlideOutput]{Text: fmt.Sprintf("Loaded slide %q from deck %q.", slide.ID, in.DeckID), Structured: out}, nil
}

func (h *handlers) removeSlide(_ context.Context, in contracts.RemoveSlideInput) (tool.Result[contracts.RemoveSlideOutput], error) {
	stored, err := h.deps.Store.RemoveSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.RemoveSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.RemoveSlideOutput{DeckID: stored.ID, Removed: true}
	return tool.Result[contracts.RemoveSlideOutput]{Text: fmt.Sprintf("Removed slide %q from deck %q.", in.SlideID, stored.ID), Structured: out}, nil
}

func (h *handlers) reorderSlides(_ context.Context, in contracts.ReorderSlidesInput) (tool.Result[contracts.ReorderSlidesOutput], error) {
	stored, err := h.deps.Store.ReorderSlides(in.DeckID, in.Order)
	if err != nil {
		return tool.Result[contracts.ReorderSlidesOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.ReorderSlidesOutput{Kind: contracts.DeckKindDeck, DeckID: stored.ID, Slides: slideSummaries(stored)}
	return tool.Result[contracts.ReorderSlidesOutput]{Text: fmt.Sprintf("Reordered %d slide(s) in deck %q.", len(out.Slides), stored.ID), Structured: out}, nil
}

func (h *handlers) duplicateSlide(_ context.Context, in contracts.DuplicateSlideInput) (tool.Result[contracts.DuplicateSlideOutput], error) {
	_, slide, err := h.deps.Store.DuplicateSlide(in.DeckID, in.SlideID, in.Position)
	if err != nil {
		return tool.Result[contracts.DuplicateSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.DuplicateSlideOutput{SlideID: slide.ID, Slide: *slide}
	return tool.Result[contracts.DuplicateSlideOutput]{Text: fmt.Sprintf("Duplicated slide %q as %q in deck %q.", in.SlideID, slide.ID, in.DeckID), Structured: out}, nil
}

func validateSlide(slide contracts.Slide) contracts.SlideValidation {
	err := ir.ValidateSlide(slide)
	if err == nil {
		return contracts.SlideValidation{OK: true}
	}
	return contracts.SlideValidation{OK: false, Issues: collectIssues(err)}
}

func collectIssues(err error) []string {
	type joiner interface{ Unwrap() []error }
	if err == nil {
		return nil
	}
	if joined, ok := err.(joiner); ok {
		var issues []string
		for _, item := range joined.Unwrap() {
			issues = append(issues, collectIssues(item)...)
		}
		return issues
	}
	return []string{err.Error()}
}

func mapSlideMutationError(deckID, slideID string, err error) error {
	if errors.Is(err, deck.ErrRevisionConflict) {
		return fmt.Errorf("deck %q slide %q revision conflict: %w", deckID, slideID, err)
	}
	return mapDeckError(deckID, err)
}
