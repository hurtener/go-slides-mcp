package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func TestListRecipesIncludesBuiltins(t *testing.T) {
	h := testHandlers()
	listed, err := h.listRecipes(context.Background(), contracts.ListRecipesInput{})
	if err != nil {
		t.Fatalf("listRecipes: %v", err)
	}
	// 4 original builtins (title_cover, bulleted_content, two_column,
	// section_break) + 4 R14.18/R14.6/R14.20 additions (agenda, pricing
	// tiers, feature card, comparison matrix).
	if len(listed.Structured.Recipes) < 8 {
		t.Fatalf("listRecipes len = %d, want at least 8 builtins", len(listed.Structured.Recipes))
	}
	if listed.Structured.Recipes[0].Source != "builtin" {
		t.Fatalf("first recipe source = %q, want builtin", listed.Structured.Recipes[0].Source)
	}
	if listed.Structured.Recipes[0].RecipeID == "" {
		t.Fatal("first recipe id empty")
	}
	// "comparison" now tags 4 builtins: the original rcp_two_column plus the
	// three R14.20 offer-card-family/matrix additions below.
	filtered, err := h.listRecipes(context.Background(), contracts.ListRecipesInput{Tag: "comparison"})
	if err != nil {
		t.Fatalf("listRecipes filtered: %v", err)
	}
	if len(filtered.Structured.Recipes) != 4 {
		t.Fatalf("listRecipes filtered len = %d, want 4: %+v", len(filtered.Structured.Recipes), filtered.Structured.Recipes)
	}
	wantComparisonIDs := map[string]bool{
		"rcp_two_column":        true,
		"rcp_pricing_tiers":     true,
		"rcp_feature_card":      true,
		"rcp_comparison_matrix": true,
	}
	for _, r := range filtered.Structured.Recipes {
		if !wantComparisonIDs[r.RecipeID] {
			t.Fatalf("unexpected recipe %q in comparison-tag filter", r.RecipeID)
		}
	}

	// "agenda" tags exactly the new R14.6 builtin.
	agendaFiltered, err := h.listRecipes(context.Background(), contracts.ListRecipesInput{Tag: "agenda"})
	if err != nil {
		t.Fatalf("listRecipes agenda: %v", err)
	}
	if len(agendaFiltered.Structured.Recipes) != 1 || agendaFiltered.Structured.Recipes[0].RecipeID != "rcp_agenda" {
		t.Fatalf("agenda-filtered recipes = %+v", agendaFiltered.Structured.Recipes)
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

// TestApplyRecipeAgendaAddsSlideToDeck mirrors
// TestApplyRecipeAddsSlideToDeck for the new R14.6 agenda builtin: apply it
// to a fresh deck and confirm the composed card_grid slide lands.
func TestApplyRecipeAgendaAddsSlideToDeck(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	createdDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Agenda"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	applied, err := h.applyRecipe(ctx, contracts.ApplyRecipeInput{DeckID: createdDeck.Structured.DeckID, RecipeID: "rcp_agenda"})
	if err != nil {
		t.Fatalf("applyRecipe: %v", err)
	}
	if applied.Structured.SlideID == "" {
		t.Fatal("applyRecipe slide id empty")
	}
	got, err := h.deps.Store.GetSlide(createdDeck.Structured.DeckID, applied.Structured.SlideID)
	if err != nil {
		t.Fatalf("store GetSlide: %v", err)
	}
	if got.Layout != contracts.LayoutCardGrid {
		t.Fatalf("applied slide layout = %q, want %q", got.Layout, contracts.LayoutCardGrid)
	}
}

// TestBuiltinRecipesValidateAndRender is the R14.18 determinism acceptance:
// every builtin recipe's Slide must pass ir.ValidateSlide AND render through
// render.Render into non-empty PPTX bytes — proving each recipe is a real
// renderable slide, not dead data, without a new NodeKind or engine change.
func TestBuiltinRecipesValidateAndRender(t *testing.T) {
	h := testHandlers()
	deckardWhite := soul.DeckardWhite()

	stored := h.deps.Recipes.List("")
	builtinCount := 0
	for _, r := range stored {
		if r.Source != "builtin" {
			continue
		}
		builtinCount++
		t.Run(r.ID, func(t *testing.T) {
			if err := ir.ValidateSlide(r.Slide); err != nil {
				t.Fatalf("recipe %q: ValidateSlide: %v", r.ID, err)
			}
			doc := contracts.SlideDoc{Slides: []contracts.Slide{r.Slide}}
			pptxBytes, _, err := render.Render(doc, deckardWhite)
			if err != nil {
				t.Fatalf("recipe %q: render.Render: %v", r.ID, err)
			}
			if len(pptxBytes) == 0 {
				t.Fatalf("recipe %q: render.Render returned empty bytes", r.ID)
			}
		})
	}
	if builtinCount < 8 {
		t.Fatalf("builtin recipe count = %d, want at least 8", builtinCount)
	}
}

// TestOfferCardFamilyBothRenderFromCardCompositeShape is the R12.10/R14.20
// family assertion: rcp_pricing_tiers and rcp_feature_card are both Grid-of-
// Cards compositions and both render; the pricing variant must now carry the
// richer R12.10 composite (Checklist + Button + optional Ribbon highlight)
// while the non-price variant stays the simpler Card+List offer card.
func TestOfferCardFamilyBothRenderFromCardCompositeShape(t *testing.T) {
	h := testHandlers()
	deckardWhite := soul.DeckardWhite()

	for _, id := range []string{"rcp_pricing_tiers", "rcp_feature_card"} {
		r, err := h.deps.Recipes.Get(id)
		if err != nil {
			t.Fatalf("recipe %q: Get: %v", id, err)
		}
		if len(r.Slide.Nodes) < 2 {
			t.Fatalf("recipe %q: want a heading + grid, got %d nodes", id, len(r.Slide.Nodes))
		}
		grid, ok := r.Slide.Nodes[len(r.Slide.Nodes)-1].(*contracts.Grid)
		if !ok {
			t.Fatalf("recipe %q: last node is %T, want *contracts.Grid", id, r.Slide.Nodes[len(r.Slide.Nodes)-1])
		}
		for i, cell := range grid.Cells {
			card, ok := cell.(*contracts.Card)
			if !ok {
				t.Fatalf("recipe %q: cell[%d] is %T, want *contracts.Card", id, i, cell)
			}
			hasList := false
			hasChecklist := false
			hasButton := false
			for _, b := range card.Body {
				switch b.(type) {
				case *contracts.List:
					hasList = true
				case *contracts.Checklist:
					hasChecklist = true
				case *contracts.Button:
					hasButton = true
				}
			}
			switch id {
			case "rcp_pricing_tiers":
				if !hasChecklist || !hasButton {
					t.Fatalf("recipe %q: card[%d] body needs Checklist + Button (got list=%v checklist=%v button=%v)", id, i, hasList, hasChecklist, hasButton)
				}
			case "rcp_feature_card":
				if !hasList && !hasChecklist {
					t.Fatalf("recipe %q: card[%d] body has neither *contracts.List nor *contracts.Checklist", id, i)
				}
			}
		}
		if id == "rcp_pricing_tiers" {
			ribboned := 0
			for _, cell := range grid.Cells {
				card := cell.(*contracts.Card)
				if card.Ribbon != nil {
					ribboned++
				}
			}
			if ribboned != 1 {
				t.Fatalf("recipe %q: ribboned card count = %d, want 1", id, ribboned)
			}
		}
		doc := contracts.SlideDoc{Slides: []contracts.Slide{r.Slide}}
		if _, _, err := render.Render(doc, deckardWhite); err != nil {
			t.Fatalf("recipe %q: render.Render: %v", id, err)
		}
	}
}
