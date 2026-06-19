package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
)

func (h *handlers) saveAsTemplate(_ context.Context, in contracts.SaveAsTemplateInput) (tool.Result[contracts.SaveAsTemplateOutput], error) {
	slide, err := h.deps.Store.GetSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.SaveAsTemplateOutput]{}, mapDeckError(in.DeckID, err)
	}
	stored, err := h.deps.Recipes.Save(recipe.Recipe{Name: in.Name, Description: in.Description, Tags: in.Tags, Slide: *slide})
	if err != nil {
		return tool.Result[contracts.SaveAsTemplateOutput]{}, err
	}
	out := contracts.SaveAsTemplateOutput{RecipeID: stored.ID, Name: stored.Name}
	return tool.Result[contracts.SaveAsTemplateOutput]{Text: fmt.Sprintf("Saved slide %q from deck %q as recipe %q.", in.SlideID, in.DeckID, stored.ID), Structured: out}, nil
}

func (h *handlers) listRecipes(_ context.Context, in contracts.ListRecipesInput) (tool.Result[contracts.ListRecipesOutput], error) {
	stored := h.deps.Recipes.List(in.Tag)
	out := contracts.ListRecipesOutput{Recipes: make([]contracts.RecipeSummary, 0, len(stored))}
	for _, item := range stored {
		out.Recipes = append(out.Recipes, contracts.RecipeSummary{RecipeID: item.ID, Name: item.Name, Tags: append([]string(nil), item.Tags...), Source: item.Source})
	}
	return tool.Result[contracts.ListRecipesOutput]{Text: fmt.Sprintf("Found %d recipe(s).", len(out.Recipes)), Structured: out}, nil
}

func (h *handlers) applyRecipe(_ context.Context, in contracts.ApplyRecipeInput) (tool.Result[contracts.ApplyRecipeOutput], error) {
	storedRecipe, err := h.deps.Recipes.Get(in.RecipeID)
	if err != nil {
		return tool.Result[contracts.ApplyRecipeOutput]{}, mapRecipeError(in.RecipeID, err)
	}
	if err := ir.ValidateSlide(storedRecipe.Slide); err != nil {
		return tool.Result[contracts.ApplyRecipeOutput]{}, fmt.Errorf("recipe %q slide invalid: %w", in.RecipeID, err)
	}
	_, slide, err := h.deps.Store.AddSlide(in.DeckID, storedRecipe.Slide, in.Position)
	if err != nil {
		return tool.Result[contracts.ApplyRecipeOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.ApplyRecipeOutput{SlideID: slide.ID, Slide: *slide}
	return tool.Result[contracts.ApplyRecipeOutput]{Text: fmt.Sprintf("Applied recipe %q to deck %q as slide %q.", in.RecipeID, in.DeckID, slide.ID), Structured: out}, nil
}

func mapRecipeError(id string, err error) error {
	if errors.Is(err, recipe.ErrNotFound) {
		return fmt.Errorf("recipe %q not found: %w", id, err)
	}
	return err
}
