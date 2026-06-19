package handlers

import (
	"context"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestValidateSlideIROK(t *testing.T) {
	h := testHandlers()
	got, err := h.validateSlideIR(context.Background(), contracts.ValidateSlideIRInput{Slide: testSlide("Intro")})
	if err != nil {
		t.Fatalf("validateSlideIR: %v", err)
	}
	if !got.Structured.OK {
		t.Fatalf("validateSlideIR OK = false, issues = %v", got.Structured.Issues)
	}
}

func TestValidateSlideIRIssues(t *testing.T) {
	h := testHandlers()
	invalid := contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.List{Items: nil}}}
	got, err := h.validateSlideIR(context.Background(), contracts.ValidateSlideIRInput{Slide: invalid})
	if err != nil {
		t.Fatalf("validateSlideIR invalid: %v", err)
	}
	if got.Structured.OK {
		t.Fatal("validateSlideIR OK = true, want false")
	}
	if len(got.Structured.Issues) == 0 {
		t.Fatal("validateSlideIR issues empty, want at least one issue")
	}
}

func TestValidateDeckForExportAggregatesPerSlide(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Validation Deck"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	deckID := created.Structured.DeckID

	valid, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: testSlide("Valid")})
	if err != nil {
		t.Fatalf("addSlide valid: %v", err)
	}
	invalidSlide := contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.TwoColumn{Left: nil, Right: nil}}}
	invalid, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: invalidSlide})
	if err != nil {
		t.Fatalf("addSlide invalid: %v", err)
	}

	got, err := h.validateDeckForExport(ctx, contracts.ValidateDeckForExportInput{DeckID: deckID})
	if err != nil {
		t.Fatalf("validateDeckForExport: %v", err)
	}
	if got.Structured.OK {
		t.Fatal("validateDeckForExport OK = true, want false")
	}
	if len(got.Structured.PerSlide) != 2 {
		t.Fatalf("validateDeckForExport perSlide len = %d, want 2", len(got.Structured.PerSlide))
	}
	if got.Structured.PerSlide[0].SlideID != valid.Structured.SlideID || !got.Structured.PerSlide[0].OK {
		t.Fatalf("validateDeckForExport first result = %+v", got.Structured.PerSlide[0])
	}
	if got.Structured.PerSlide[1].SlideID != invalid.Structured.SlideID || got.Structured.PerSlide[1].OK {
		t.Fatalf("validateDeckForExport second result = %+v", got.Structured.PerSlide[1])
	}
	if len(got.Structured.Blockers) == 0 {
		t.Fatal("validateDeckForExport blockers empty, want at least one blocker")
	}
}

func TestValidateSlideUsesStoredSlide(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Stored Validation"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	added, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: created.Structured.DeckID, Slide: testSlide("Stored")})
	if err != nil {
		t.Fatalf("addSlide: %v", err)
	}

	got, err := h.validateSlide(ctx, contracts.ValidateSlideInput{DeckID: created.Structured.DeckID, SlideID: added.Structured.SlideID})
	if err != nil {
		t.Fatalf("validateSlide: %v", err)
	}
	if !got.Structured.OK || got.Structured.SlideID != added.Structured.SlideID {
		t.Fatalf("validateSlide got %+v", got.Structured)
	}
}
