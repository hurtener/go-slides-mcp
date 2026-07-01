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

// gridSlide returns a top-heavy slide (a heading plus a 2-cell grid) with no
// explicit alignment — the exact shape R10-A's autofit.Fill is meant to act
// on, used to exercise the Autofit export flag end to end.
func gridSlide(title string) contracts.Slide {
	return contracts.Slide{
		Layout: contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: title}}},
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Stat{Value: "1", Label: "One"},
				&contracts.Stat{Value: "2", Label: "Two"},
			}},
		},
	}
}

// TestExportDeckAutofitTrueRendersGridSlideWithoutError proves R10-A: an
// export with Autofit=true on a deck containing a top-level grid slide
// succeeds and renders a valid .pptx — autofit.Fill only redistributes
// existing slack onto the grid via VAlignFill, so it cannot introduce a
// render error or overflow.
func TestExportDeckAutofitTrueRendersGridSlideWithoutError(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Autofit Deck"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	if _, _, err := h.deps.Store.AddSlide(created.Structured.DeckID, gridSlide("Highlights"), nil); err != nil {
		t.Fatalf("store AddSlide: %v", err)
	}

	exported, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: created.Structured.DeckID, Autofit: true})
	if err != nil {
		t.Fatalf("exportDeck (autofit): %v", err)
	}
	buf, err := os.ReadFile(exported.Structured.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(export path): %v", err)
	}
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}
}

// TestExportDeckAutofitFalseLeavesStoredSlideUnchanged proves R10-A: Autofit
// defaults to false, and an export with Autofit false (the zero value) does
// not mutate the stored slide's alignment — confirming the autofit.Fill pass
// is never applied unless explicitly opted in, so default export behavior is
// byte-identical to before R10-A.
func TestExportDeckAutofitFalseLeavesStoredSlideUnchanged(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "No Autofit Deck"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	stored, _, err := h.deps.Store.AddSlide(created.Structured.DeckID, gridSlide("Highlights"), nil)
	if err != nil {
		t.Fatalf("store AddSlide: %v", err)
	}
	if stored.Slides[0].Align.Vertical != "" {
		t.Fatalf("precondition: stored slide Align.Vertical = %q, want empty", stored.Slides[0].Align.Vertical)
	}

	if _, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: created.Structured.DeckID}); err != nil {
		t.Fatalf("exportDeck (no autofit): %v", err)
	}

	after, err := h.deps.Store.GetDeck(created.Structured.DeckID)
	if err != nil {
		t.Fatalf("GetDeck after export: %v", err)
	}
	if after.Slides[0].Align.Vertical != "" {
		t.Fatalf("exportDeck with Autofit=false mutated stored slide: Align.Vertical = %q, want empty", after.Slides[0].Align.Vertical)
	}
}

// overflowingSlide stacks far more Hero nodes than the body region can hold
// (mirrors render's overflowingDoc fixture), guaranteeing at least one
// overflow LayoutWarning so the export path's Autofit=true wiring — Fill,
// then autofit.Remediate driven by real render.RenderWithAssets output —
// exercises a real (not faked) overflow signal end to end.
func overflowingSlide() contracts.Slide {
	nodes := make([]contracts.SlideNode, 0, 8)
	for i := 0; i < 8; i++ {
		nodes = append(nodes, &contracts.Hero{Title: "Overflow driver"})
	}
	return contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: nodes}
}

// TestExportDeckAutofitTrueRemediatesOverflowingDeck proves R10-C wiring: an
// export with Autofit=true on a deck whose slide overflows still succeeds and
// renders a valid .pptx (Fill + the remediation ladder run against the real
// render pipeline, capped at 2 rungs, never erroring even if the ladder can't
// fully clear the overflow).
func TestExportDeckAutofitTrueRemediatesOverflowingDeck(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Overflow Deck"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	if _, _, err := h.deps.Store.AddSlide(created.Structured.DeckID, overflowingSlide(), nil); err != nil {
		t.Fatalf("store AddSlide: %v", err)
	}

	exported, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: created.Structured.DeckID, Autofit: true})
	if err != nil {
		t.Fatalf("exportDeck (autofit, overflowing): %v", err)
	}
	buf, err := os.ReadFile(exported.Structured.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(export path): %v", err)
	}
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
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
