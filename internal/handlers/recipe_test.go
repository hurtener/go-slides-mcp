package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
)

func TestListRecipesIncludesBuiltins(t *testing.T) {
	h := testHandlers()
	listed, err := h.listRecipes(context.Background(), contracts.ListRecipesInput{})
	if err != nil {
		t.Fatalf("listRecipes: %v", err)
	}
	if len(listed.Structured.Recipes) < 4 {
		t.Fatalf("listRecipes len = %d, want at least 4 builtins", len(listed.Structured.Recipes))
	}
	if listed.Structured.Recipes[0].Source != "builtin" {
		t.Fatalf("first recipe source = %q, want builtin", listed.Structured.Recipes[0].Source)
	}
	if listed.Structured.Recipes[0].RecipeID == "" {
		t.Fatal("first recipe id empty")
	}
	filtered, err := h.listRecipes(context.Background(), contracts.ListRecipesInput{Tag: "comparison"})
	if err != nil {
		t.Fatalf("listRecipes filtered: %v", err)
	}
	if len(filtered.Structured.Recipes) != 1 || filtered.Structured.Recipes[0].RecipeID != "rcp_two_column" {
		t.Fatalf("filtered recipes = %+v", filtered.Structured.Recipes)
	}
}

func TestSaveAsTemplateThenListShowsUserRecipe(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Templates"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	added, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: createdDeck.Structured.DeckID, Slide: testSlide("Reusable")})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}

	saved, err := h.saveAsTemplate(ctx, contracts.SaveAsTemplateInput{DeckID: createdDeck.Structured.DeckID, SlideID: added.Structured.SlideID, Name: "Reusable Slide", Description: "Saved from a deck", Tags: []string{"custom", "agenda"}})
	if err != nil {
		t.Fatalf("saveAsTemplate: %v", err)
	}
	if saved.Structured.RecipeID == "" {
		t.Fatal("saveAsTemplate recipe id empty")
	}

	listed, err := h.listRecipes(ctx, contracts.ListRecipesInput{Tag: "custom"})
	if err != nil {
		t.Fatalf("listRecipes custom: %v", err)
	}
	if len(listed.Structured.Recipes) != 1 {
		t.Fatalf("listRecipes custom len = %d, want 1", len(listed.Structured.Recipes))
	}
	if listed.Structured.Recipes[0].RecipeID != saved.Structured.RecipeID || listed.Structured.Recipes[0].Source != "user" {
		t.Fatalf("listed user recipe = %+v", listed.Structured.Recipes[0])
	}
	stored, err := h.deps.Recipes.Get(saved.Structured.RecipeID)
	if err != nil {
		t.Fatalf("recipe store Get: %v", err)
	}
	if stored.Slide.ID != "" {
		t.Fatalf("stored recipe slide id = %q, want empty", stored.Slide.ID)
	}
}

func TestApplyRecipeAddsSlideToDeck(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Apply"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	applied, err := h.applyRecipe(ctx, contracts.ApplyRecipeInput{DeckID: createdDeck.Structured.DeckID, RecipeID: "rcp_title_cover"})
	if err != nil {
		t.Fatalf("applyRecipe: %v", err)
	}
	if applied.Structured.SlideID == "" {
		t.Fatal("applyRecipe slide id empty")
	}
	if applied.Structured.Slide.ID != applied.Structured.SlideID {
		t.Fatalf("applied slide id = %q, want %q", applied.Structured.Slide.ID, applied.Structured.SlideID)
	}
	got, err := h.deps.Store.GetSlide(createdDeck.Structured.DeckID, applied.Structured.SlideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	if got.Layout != contracts.LayoutCover {
		t.Fatalf("applied slide layout = %q, want %q", got.Layout, contracts.LayoutCover)
	}
}

func TestApplyRecipeMissingReturnsNotFound(t *testing.T) {
	h := testHandlers()
	_, err := h.applyRecipe(context.Background(), contracts.ApplyRecipeInput{DeckID: "deck_missing", RecipeID: "rcp_missing"})
	if !errors.Is(err, recipe.ErrNotFound) {
		t.Fatalf("applyRecipe missing err = %v, want ErrNotFound", err)
	}
}
