package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
)

func (h *handlers) validateSlideIR(_ context.Context, in contracts.ValidateSlideIRInput) (tool.Result[contracts.ValidateSlideIROutput], error) {
	validation := validateSlide(in.Slide)
	out := contracts.ValidateSlideIROutput(validation)
	return tool.Result[contracts.ValidateSlideIROutput]{Text: validationText("slide IR", out.OK, out.Issues), Structured: out}, nil
}

func (h *handlers) validateSlide(_ context.Context, in contracts.ValidateSlideInput) (tool.Result[contracts.ValidateSlideOutput], error) {
	slide, err := h.deps.Store.GetSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.ValidateSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	validation := validateSlide(*slide)
	out := contracts.ValidateSlideOutput{SlideID: in.SlideID, OK: validation.OK, Issues: validation.Issues}
	return tool.Result[contracts.ValidateSlideOutput]{Text: validationText(fmt.Sprintf("slide %q", in.SlideID), out.OK, out.Issues), Structured: out}, nil
}

func (h *handlers) validateDeckForExport(_ context.Context, in contracts.ValidateDeckForExportInput) (tool.Result[contracts.ValidateDeckForExportOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.ValidateDeckForExportOutput]{}, mapDeckError(in.DeckID, err)
	}

	doc := contracts.SlideDoc{Title: stored.Title, Slides: append([]contracts.Slide(nil), stored.Slides...)}
	out := validateDoc(doc)
	out.PerSlide = make([]contracts.DeckSlideValidation, 0, len(stored.Slides))
	for _, slide := range stored.Slides {
		validation := validateSlide(slide)
		perSlide := contracts.DeckSlideValidation{SlideID: slide.ID, OK: validation.OK, Issues: validation.Issues}
		out.PerSlide = append(out.PerSlide, perSlide)
	}

	return tool.Result[contracts.ValidateDeckForExportOutput]{Text: validationText(fmt.Sprintf("deck %q", stored.ID), out.OK, out.Blockers), Structured: out}, nil
}

func validationText(target string, ok bool, issues []string) string {
	if ok {
		return fmt.Sprintf("Validated %s: OK.", target)
	}
	return fmt.Sprintf("Validated %s: %d issue(s).", target, len(issues))
}

func validateDoc(doc contracts.SlideDoc) contracts.ValidateDeckForExportOutput {
	err := ir.ValidateDoc(doc)
	if err == nil {
		return contracts.ValidateDeckForExportOutput{OK: true}
	}
	issues := collectIssues(err)
	blockers := make([]string, 0, len(issues))
	blockers = append(blockers, issues...)
	return contracts.ValidateDeckForExportOutput{OK: false, Blockers: blockers}
}
