package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
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
