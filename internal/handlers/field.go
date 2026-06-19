package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
)

func (h *handlers) editSlideField(_ context.Context, in contracts.EditSlideFieldInput) (tool.Result[contracts.EditSlideFieldOutput], error) {
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		node, err := ir.Resolve(slide, in.Path)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(in.Value) // string → a JSON string value
		if err != nil {
			return err
		}
		updated, err := setNodeField(node, in.Field, raw)
		if err != nil {
			return err
		}
		return ir.Set(slide, in.Path, updated)
	})
	if err != nil {
		return tool.Result[contracts.EditSlideFieldOutput]{}, err
	}
	out := contracts.EditSlideFieldOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.EditSlideFieldOutput]{Text: fmt.Sprintf("Edited field %q in slide %q in deck %q.", in.Field, in.SlideID, in.DeckID), Structured: out}, nil
}

func (h *handlers) patchSlideText(_ context.Context, in contracts.PatchSlideTextInput) (tool.Result[contracts.PatchSlideTextOutput], error) {
	raw, err := json.Marshal([]map[string]string{{"text": in.Text}})
	if err != nil {
		return tool.Result[contracts.PatchSlideTextOutput]{}, fmt.Errorf("marshal text patch: %w", err)
	}
	slide, validation, err := h.mutateSlide(in.DeckID, in.SlideID, in.ExpectedRevisionHash, func(slide *contracts.Slide) error {
		node, err := ir.Resolve(slide, in.Path)
		if err != nil {
			return err
		}
		updated, err := setNodeField(node, in.Field, raw)
		if err != nil {
			return err
		}
		return ir.Set(slide, in.Path, updated)
	})
	if err != nil {
		return tool.Result[contracts.PatchSlideTextOutput]{}, err
	}
	out := contracts.PatchSlideTextOutput{Slide: slide, Validation: validation}
	return tool.Result[contracts.PatchSlideTextOutput]{Text: fmt.Sprintf("Patched text field %q in slide %q in deck %q.", in.Field, in.SlideID, in.DeckID), Structured: out}, nil
}

func setNodeField(node contracts.SlideNode, field string, raw json.RawMessage) (contracts.SlideNode, error) {
	encoded, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("marshal slide node: %w", err)
	}
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, fmt.Errorf("decode slide node object: %w", err)
	}
	fields[field] = raw
	merged, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("re-encode slide node object: %w", err)
	}
	updated, err := contracts.UnmarshalSlideNode(merged)
	if err != nil {
		return nil, fmt.Errorf("decode edited slide node: %w", err)
	}
	return updated, nil
}
