package handlers

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

func TestExportHandlersProducePathAndResourceMetadata(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Quarterly Review"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	stored, _, err := h.deps.Store.AddSlide(created.Structured.DeckID, testSlide("Agenda"), nil)
	if err != nil {
		t.Fatalf("store AddSlide: %v", err)
	}

	exported, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: stored.ID})
	if err != nil {
		t.Fatalf("exportDeck: %v", err)
	}
	if exported.Structured.Path == "" {
		t.Fatal("exportDeck returned empty path")
	}
	if exported.Structured.ResourceURI != exportstore.DeckResourceURI(stored.ID) {
		t.Fatalf("exportDeck resource uri = %q, want %q", exported.Structured.ResourceURI, exportstore.DeckResourceURI(stored.ID))
	}
	buf, err := os.ReadFile(exported.Structured.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(export path): %v", err)
	}
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}

	listed, err := h.listResources(ctx, contracts.ListResourcesInput{})
	if err != nil {
		t.Fatalf("listResources: %v", err)
	}
	if len(listed.Structured.Resources) != 1 {
		t.Fatalf("listResources len = %d, want 1", len(listed.Structured.Resources))
	}
	if listed.Structured.Resources[0].URI != exported.Structured.ResourceURI {
		t.Fatalf("listResources uri = %q, want %q", listed.Structured.Resources[0].URI, exported.Structured.ResourceURI)
	}

	resolved, err := h.getResource(ctx, contracts.GetResourceInput{URI: exported.Structured.ResourceURI})
	if err != nil {
		t.Fatalf("getResource: %v", err)
	}
	if !resolved.Structured.Found {
		t.Fatal("getResource Found = false, want true")
	}
	if resolved.Structured.Path != exported.Structured.Path {
		t.Fatalf("getResource path = %q, want %q", resolved.Structured.Path, exported.Structured.Path)
	}
}

// TestExportDeckBrandSoulSignal proves R8.8: exporting a deck on the
// built-in default soul reports BrandSoulEstablished == false, SoulID ==
// the default id, and the Text carries the notice; exporting a deck whose
// SoulID resolves to a stored brand soul reports true with the brand id and
// no notice; a SoulID that does NOT resolve in the store falls back to the
// default (established == false).
func TestExportDeckBrandSoulSignal(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	// Default soul (empty SoulID).
	defaultDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Default Soul Deck"})
	if err != nil {
		t.Fatalf("createDeck (default): %v", err)
	}
	if _, _, err := h.deps.Store.AddSlide(defaultDeck.Structured.DeckID, testSlide("Agenda"), nil); err != nil {
		t.Fatalf("store AddSlide (default): %v", err)
	}
	exportedDefault, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: defaultDeck.Structured.DeckID})
	if err != nil {
		t.Fatalf("exportDeck (default): %v", err)
	}
	if exportedDefault.Structured.BrandSoulEstablished {
		t.Fatal("exportDeck (default soul) BrandSoulEstablished = true, want false")
	}
	if exportedDefault.Structured.SoulID != soul.DeckardWhiteID {
		t.Fatalf("exportDeck (default soul) SoulID = %q, want %q", exportedDefault.Structured.SoulID, soul.DeckardWhiteID)
	}
	if !strings.Contains(exportedDefault.Text, noBrandSoulNotice) {
		t.Fatalf("exportDeck (default soul) Text = %q, want it to contain the no-brand-soul notice", exportedDefault.Text)
	}

	// Brand soul that resolves in the store.
	brandSoul := soul.DeckardWhite()
	brandSoul.ID = "soul_acme"
	brandSoul.Name = "Acme"
	if err := h.deps.Souls.Put(brandSoul); err != nil {
		t.Fatalf("Souls.Put: %v", err)
	}
	brandDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Brand Soul Deck", SoulID: "soul_acme"})
	if err != nil {
		t.Fatalf("createDeck (brand): %v", err)
	}
	if _, _, err := h.deps.Store.AddSlide(brandDeck.Structured.DeckID, testSlide("Agenda"), nil); err != nil {
		t.Fatalf("store AddSlide (brand): %v", err)
	}
	exportedBrand, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: brandDeck.Structured.DeckID})
	if err != nil {
		t.Fatalf("exportDeck (brand): %v", err)
	}
	if !exportedBrand.Structured.BrandSoulEstablished {
		t.Fatal("exportDeck (brand soul) BrandSoulEstablished = false, want true")
	}
	if exportedBrand.Structured.SoulID != "soul_acme" {
		t.Fatalf("exportDeck (brand soul) SoulID = %q, want soul_acme", exportedBrand.Structured.SoulID)
	}
	if strings.Contains(exportedBrand.Text, noBrandSoulNotice) {
		t.Fatalf("exportDeck (brand soul) Text = %q, want no no-brand-soul notice", exportedBrand.Text)
	}

	// SoulID set but unresolved in the store: falls back to the default soul.
	unresolvedDeck, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Unresolved Soul Deck", SoulID: "soul_ghost"})
	if err != nil {
		t.Fatalf("createDeck (unresolved): %v", err)
	}
	if _, _, err := h.deps.Store.AddSlide(unresolvedDeck.Structured.DeckID, testSlide("Agenda"), nil); err != nil {
		t.Fatalf("store AddSlide (unresolved): %v", err)
	}
	exportedUnresolved, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: unresolvedDeck.Structured.DeckID})
	if err != nil {
		t.Fatalf("exportDeck (unresolved): %v", err)
	}
	if exportedUnresolved.Structured.BrandSoulEstablished {
		t.Fatal("exportDeck (unresolved soul) BrandSoulEstablished = true, want false (falls back to default)")
	}
	if exportedUnresolved.Structured.SoulID != soul.DeckardWhiteID {
		t.Fatalf("exportDeck (unresolved soul) SoulID = %q, want %q (falls back to default)", exportedUnresolved.Structured.SoulID, soul.DeckardWhiteID)
	}
	if !strings.Contains(exportedUnresolved.Text, noBrandSoulNotice) {
		t.Fatalf("exportDeck (unresolved soul) Text = %q, want it to contain the no-brand-soul notice", exportedUnresolved.Text)
	}
}

func TestGetResourceMissingReturnsFoundFalse(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()

	got, err := h.getResource(context.Background(), contracts.GetResourceInput{URI: "deck://export/missing.pptx"})
	if err != nil {
		t.Fatalf("getResource missing: %v", err)
	}
	if got.Structured.Found {
		t.Fatal("getResource Found = true, want false")
	}
}
