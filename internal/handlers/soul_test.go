package handlers

import (
	"context"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func TestListSoulsIncludesDeckardWhite(t *testing.T) {
	h := testHandlers()

	got, err := h.listSouls(context.Background(), contracts.ListSoulsInput{})
	if err != nil {
		t.Fatalf("listSouls: %v", err)
	}
	if len(got.Structured.Souls) == 0 {
		t.Fatal("listSouls returned no souls")
	}
	if got.Structured.Souls[0].SoulID != soul.DeckardWhiteID {
		t.Fatalf("first soul id = %q, want %q", got.Structured.Souls[0].SoulID, soul.DeckardWhiteID)
	}
}

func TestBootstrapSoulAndGetDesignTokensRoundTrip(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Teal Variant", Accent: "112233"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}
	if bootstrapped.Structured.TokenCount == 0 {
		t.Fatal("bootstrapSoul token count = 0, want > 0")
	}

	tokens, err := h.getDesignTokens(ctx, contracts.GetDesignTokensInput{SoulID: bootstrapped.Structured.SoulID})
	if err != nil {
		t.Fatalf("getDesignTokens: %v", err)
	}
	if got := tokenValue(tokens.Structured.Tokens, contracts.TokenLayerSurface, "accent"); got != "112233" {
		t.Fatalf("accent token = %q, want 112233", got)
	}
	if len(tokens.Structured.Tokens) != bootstrapped.Structured.TokenCount {
		t.Fatalf("getDesignTokens len = %d, want %d", len(tokens.Structured.Tokens), bootstrapped.Structured.TokenCount)
	}
}

func TestRefineSoulChangesToken(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Refine Me"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}

	refined, err := h.refineSoul(ctx, contracts.RefineSoulInput{SoulID: bootstrapped.Structured.SoulID, Overrides: []contracts.SoulOverride{{Category: "surface", Token: "accent", Value: "ABCDEF"}}})
	if err != nil {
		t.Fatalf("refineSoul: %v", err)
	}
	if len(refined.Structured.Changed) != 1 || refined.Structured.Changed[0] != "surface.accent" {
		t.Fatalf("refineSoul changed = %+v, want [surface.accent]", refined.Structured.Changed)
	}

	tokens, err := h.getDesignTokens(ctx, contracts.GetDesignTokensInput{SoulID: bootstrapped.Structured.SoulID})
	if err != nil {
		t.Fatalf("getDesignTokens: %v", err)
	}
	if got := tokenValue(tokens.Structured.Tokens, contracts.TokenLayerSurface, "accent"); got != "ABCDEF" {
		t.Fatalf("accent token = %q, want ABCDEF", got)
	}
}

func TestGetSoulIncludesDeckardWhiteStyleGuide(t *testing.T) {
	h := testHandlers()

	got, err := h.getSoul(context.Background(), contracts.GetSoulInput{SoulID: soul.DeckardWhiteID, IncludeStyleGuide: true})
	if err != nil {
		t.Fatalf("getSoul: %v", err)
	}
	if got.Structured.StyleGuide == nil {
		t.Fatal("getSoul style guide = nil, want value")
	}
	if got.Structured.StyleGuide.NorthStar == "" {
		t.Fatal("getSoul north star empty, want Deckard White voice")
	}
	if len(got.Structured.StyleGuide.Do) == 0 || len(got.Structured.StyleGuide.Dont) == 0 {
		t.Fatalf("getSoul style guide = %+v, want do/dont guidance", got.Structured.StyleGuide)
	}
}

func tokenValue(tokens []contracts.TokenEntry, layer contracts.TokenLayer, name string) string {
	for _, token := range tokens {
		if token.Layer == layer && token.Name == name {
			return token.Value
		}
	}
	return ""
}
