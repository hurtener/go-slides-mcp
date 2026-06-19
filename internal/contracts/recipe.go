package contracts

// SaveAsTemplateInput is the typed input for save_as_template.
type SaveAsTemplateInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to save as a recipe.
	SlideID string `json:"slideId"`
	// Name is the caller-facing recipe name.
	Name string `json:"name"`
	// Description is the optional recipe summary.
	Description string `json:"description,omitempty"`
	// Tags are the optional recipe labels used for filtering.
	Tags []string `json:"tags,omitempty"`
}

// SaveAsTemplateOutput is the structured result for save_as_template.
type SaveAsTemplateOutput struct {
	// RecipeID is the stored recipe identifier.
	RecipeID string `json:"recipeId"`
	// Name is the stored recipe name.
	Name string `json:"name"`
}

// ListRecipesInput is the typed input for list_recipes.
type ListRecipesInput struct {
	// Tag filters the returned recipes to one tag when set.
	Tag string `json:"tag,omitempty"`
	// SoulID is reserved for future soul-aware recipe selection.
	SoulID string `json:"soulId,omitempty"`
}

// RecipeSummary is one list_recipes result item.
type RecipeSummary struct {
	// RecipeID is the stable recipe identifier.
	RecipeID string `json:"recipeId"`
	// Name is the caller-facing recipe name.
	Name string `json:"name"`
	// Tags are the recipe labels used for filtering.
	Tags []string `json:"tags,omitempty"`
	// Source reports whether the recipe is builtin or user-saved.
	Source string `json:"source,omitempty"`
}

// ListRecipesOutput is the structured result for list_recipes.
type ListRecipesOutput struct {
	// Recipes is the ordered recipe list, built-ins first then user recipes.
	Recipes []RecipeSummary `json:"recipes,omitempty"`
}

// ApplyRecipeInput is the typed input for apply_recipe.
type ApplyRecipeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// RecipeID is the stable recipe identifier to instantiate.
	RecipeID string `json:"recipeId"`
	// Position is the optional zero-based insertion index.
	Position *int `json:"position,omitempty"`
}

// ApplyRecipeOutput is the structured result for apply_recipe.
type ApplyRecipeOutput struct {
	// SlideID is the inserted slide identifier assigned by the deck store.
	SlideID string `json:"slideId"`
	// Slide is the stored slide snapshot created from the recipe.
	Slide Slide `json:"slide"`
}
