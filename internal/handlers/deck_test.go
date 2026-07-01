package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func TestDeckHandlersRoundTripThroughStore(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Quarterly Review", Author: "Deckard", SoulID: "deckard-white"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	if created.Structured.DeckID == "" {
		t.Fatal("createDeck returned empty deck id")
	}

	listed, err := h.listDecks(ctx, contracts.ListDecksInput{})
	if err != nil {
		t.Fatalf("listDecks: %v", err)
	}
	if len(listed.Structured.Decks) != 1 {
		t.Fatalf("listDecks len = %d, want 1", len(listed.Structured.Decks))
	}

	stored, err := h.deps.Store.GetDeck(created.Structured.DeckID)
	if err != nil {
		t.Fatalf("store GetDeck: %v", err)
	}
	stored, _, err = h.deps.Store.AddSlide(stored.ID, testSlide("Agenda"), nil)
	if err != nil {
		t.Fatalf("store AddSlide: %v", err)
	}

	got, err := h.getDeck(ctx, contracts.GetDeckInput{DeckID: stored.Slug})
	if err != nil {
		t.Fatalf("getDeck: %v", err)
	}
	if got.Structured.DeckID != stored.ID {
		t.Fatalf("getDeck id = %q, want %q", got.Structured.DeckID, stored.ID)
	}
	if len(got.Structured.Slides) != 1 {
		t.Fatalf("getDeck slides len = %d, want 1", len(got.Structured.Slides))
	}
	if got.Structured.Slides[0].Title != "Agenda" {
		t.Fatalf("getDeck first slide title = %q, want Agenda", got.Structured.Slides[0].Title)
	}

	chrome, err := h.setDeckChrome(ctx, contracts.SetDeckChromeInput{DeckID: stored.ID, Chrome: contracts.DeckChrome{Enabled: true, BrandText: "Deckard"}})
	if err != nil {
		t.Fatalf("setDeckChrome: %v", err)
	}
	if !chrome.Structured.Chrome.Enabled || chrome.Structured.Chrome.BrandText != "Deckard" {
		t.Fatalf("setDeckChrome got %+v", chrome.Structured.Chrome)
	}

	sections, err := h.setDeckSections(ctx, contracts.SetDeckSectionsInput{DeckID: stored.ID, Sections: []contracts.DeckSection{{Name: "Opening", SlideIDs: []string{got.Structured.Slides[0].SlideID}}}})
	if err != nil {
		t.Fatalf("setDeckSections: %v", err)
	}
	if len(sections.Structured.Sections) != 1 || sections.Structured.Sections[0].Name != "Opening" {
		t.Fatalf("setDeckSections got %+v", sections.Structured.Sections)
	}

	deleted, err := h.deleteDeck(ctx, contracts.DeleteDeckInput{DeckID: stored.Slug})
	if err != nil {
		t.Fatalf("deleteDeck: %v", err)
	}
	if !deleted.Structured.Deleted {
		t.Fatal("deleteDeck Deleted = false, want true")
	}
	if _, err := h.deps.Store.GetDeck(stored.ID); !errors.Is(err, deck.ErrNotFound) {
		t.Fatalf("store GetDeck after delete err = %v, want ErrNotFound", err)
	}
}

// TestCreateDeckBrandSoulSignal proves R8.8: creating a deck on the built-in
// default soul (empty or "deckard-white" SoulID) reports
// BrandSoulEstablished == false and the Text carries the non-fatal notice; a
// deck created on a real brand soul reports true with no notice.
func TestCreateDeckBrandSoulSignal(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	defaultDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Default Soul Deck"})
	if err != nil {
		t.Fatalf("createDeck (default): %v", err)
	}
	if defaultDeck.Structured.BrandSoulEstablished {
		t.Fatal("createDeck (empty SoulID) BrandSoulEstablished = true, want false")
	}
	if !strings.Contains(defaultDeck.Text, noBrandSoulNotice) {
		t.Fatalf("createDeck (empty SoulID) Text = %q, want it to contain the no-brand-soul notice", defaultDeck.Text)
	}

	explicitDefault, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Explicit Default Soul Deck", SoulID: soul.DeckardWhiteID})
	if err != nil {
		t.Fatalf("createDeck (explicit default): %v", err)
	}
	if explicitDefault.Structured.BrandSoulEstablished {
		t.Fatal("createDeck (deckard-white SoulID) BrandSoulEstablished = true, want false")
	}
	if !strings.Contains(explicitDefault.Text, noBrandSoulNotice) {
		t.Fatalf("createDeck (deckard-white SoulID) Text = %q, want it to contain the no-brand-soul notice", explicitDefault.Text)
	}

	brandDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Brand Soul Deck", SoulID: "soul_acme"})
	if err != nil {
		t.Fatalf("createDeck (brand): %v", err)
	}
	if !brandDeck.Structured.BrandSoulEstablished {
		t.Fatal("createDeck (brand SoulID) BrandSoulEstablished = false, want true")
	}
	if strings.Contains(brandDeck.Text, noBrandSoulNotice) {
		t.Fatalf("createDeck (brand SoulID) Text = %q, want no no-brand-soul notice", brandDeck.Text)
	}
}

func TestGetDeckMissingReturnsNotFound(t *testing.T) {
	h := testHandlers()
	_, err := h.getDeck(context.Background(), contracts.GetDeckInput{DeckID: "missing"})
	if !errors.Is(err, deck.ErrNotFound) {
		t.Fatalf("getDeck missing err = %v, want ErrNotFound", err)
	}
}

func testHandlers() *handlers {
	return &handlers{deps: ToolDeps{
		Store:     deck.NewMemoryStore(),
		Souls:     soul.NewMemoryRegistry(),
		Assets:    asset.NewMemoryStore(),
		Comments:  comment.NewMemoryStore(),
		Recipes:   recipe.NewMemoryStore(),
		Session:   &SessionState{},
		BuildInfo: contracts.BuildInfo{Name: "go-slides-mcp", Version: "test"},
		Workspace: "/tmp/deckard-test",
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}}
}

func testSlide(title string) contracts.Slide {
	return contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: title}}},
			&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Body"}}}},
		},
	}
}
