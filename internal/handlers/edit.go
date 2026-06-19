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

func (h *handlers) editSlideNode(_ context.Context, in contracts.EditSlideNodeInput) (tool.Result[contracts.EditSlideNodeOutput], error) {
	node, err := decodeSlideNode(in.Node)
	if err != nil {
		return tool.Result[contracts.EditSlideNodeOutput]{}, err
	}
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		return ir.Set(slide, in.Path, node)
	})
	if err != nil {
		return tool.Result[contracts.EditSlideNodeOutput]{}, err
	}
	out := contracts.EditSlideNodeOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.EditSlideNodeOutput]{Text: fmt.Sprintf("Edited one node in slide %q in deck %q.", in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) insertSlideNode(_ context.Context, in contracts.InsertSlideNodeInput) (tool.Result[contracts.InsertSlideNodeOutput], error) {
	node, err := decodeSlideNode(in.Node)
	if err != nil {
		return tool.Result[contracts.InsertSlideNodeOutput]{}, err
	}
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		return ir.Insert(slide, in.Path, node)
	})
	if err != nil {
		return tool.Result[contracts.InsertSlideNodeOutput]{}, err
	}
	out := contracts.InsertSlideNodeOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.InsertSlideNodeOutput]{Text: fmt.Sprintf("Inserted one node in slide %q in deck %q.", in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) removeSlideNode(_ context.Context, in contracts.RemoveSlideNodeInput) (tool.Result[contracts.RemoveSlideNodeOutput], error) {
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		_, err := ir.Remove(slide, in.Path)
		return err
	})
	if err != nil {
		return tool.Result[contracts.RemoveSlideNodeOutput]{}, err
	}
	out := contracts.RemoveSlideNodeOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.RemoveSlideNodeOutput]{Text: fmt.Sprintf("Removed one node from slide %q in deck %q.", in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) duplicateSlideNode(_ context.Context, in contracts.DuplicateSlideNodeInput) (tool.Result[contracts.DuplicateSlideNodeOutput], error) {
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		_, err := ir.Duplicate(slide, in.Path)
		return err
	})
	if err != nil {
		return tool.Result[contracts.DuplicateSlideNodeOutput]{}, err
	}
	out := contracts.DuplicateSlideNodeOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.DuplicateSlideNodeOutput]{Text: fmt.Sprintf("Duplicated one node in slide %q in deck %q.", in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) moveSlideNode(_ context.Context, in contracts.MoveSlideNodeInput) (tool.Result[contracts.MoveSlideNodeOutput], error) {
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		return ir.Move(slide, in.From, in.To)
	})
	if err != nil {
		return tool.Result[contracts.MoveSlideNodeOutput]{}, err
	}
	out := contracts.MoveSlideNodeOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.MoveSlideNodeOutput]{Text: fmt.Sprintf("Moved one node in slide %q in deck %q.", in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) mutateSlide(deckID, slideID, expectedRevision string, apply func(*contracts.Slide) error) (contracts.Slide, contracts.SlideValidation, error) {
	slide, err := h.deps.Store.GetSlide(deckID, slideID)
	if err != nil {
		return contracts.Slide{}, contracts.SlideValidation{}, mapDeckError(deckID, err)
	}
	if err := apply(slide); err != nil {
		return contracts.Slide{}, contracts.SlideValidation{}, mapSlideEditError(deckID, slideID, err)
	}
	validation := validateSlide(*slide)
	_, stored, err := h.deps.Store.UpdateSlide(deckID, slideID, *slide, expectedRevision)
	if err != nil {
		return contracts.Slide{}, contracts.SlideValidation{}, mapSlideMutationError(deckID, slideID, err)
	}
	return *stored, validation, nil
}

func decodeSlideNode(raw []byte) (contracts.SlideNode, error) {
	node, err := contracts.UnmarshalSlideNode(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid slide node: %w", err)
	}
	return node, nil
}

func mapSlideEditError(deckID, slideID string, err error) error {
	if errors.Is(err, deck.ErrNotFound) {
		return mapDeckError(deckID, err)
	}
	return fmt.Errorf("deck %q slide %q invalid edit path: %w", deckID, slideID, err)
}
