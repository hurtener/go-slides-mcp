package render

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// logoWallPNG is a 1x1 transparent PNG, reused from the image/asset test
// pattern (render_assets_test.go's validPNG) for a resolvable logo asset.
const logoWallPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

func decodeLogoWallPNG(t *testing.T) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(logoWallPNG)
	if err != nil {
		t.Fatalf("base64 decode PNG: %v", err)
	}
	return b
}

// logoWallDoc builds a one-slide doc with a 3-logo LogoWall (R14.7, D-125),
// covering Columns/Tone/Caption + every LogoEntry field.
func logoWallDoc(ids [3]contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "LogoWall Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "logo-wall",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Trusted By")},
					&contracts.LogoWall{
						Logos: []contracts.LogoEntry{
							{AssetID: ids[0], Alt: "Acme Corp"},
							{AssetID: ids[1], Alt: "Globex"},
							{AssetID: ids[2], Alt: "Initech"},
						},
						Columns: 3,
						Tone:    contracts.LogoToneMono,
						Caption: "Trusted by",
					},
				},
			},
		},
	}
}

// missingLogoWallDoc builds a one-slide doc whose single logo AssetID does
// not resolve through the AssetResolver — the render must warn, not panic.
func missingLogoWallDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "LogoWall Coverage — missing asset",
		Slides: []contracts.Slide{
			{
				ID:     "logo-wall-missing",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.LogoWall{
						Logos: []contracts.LogoEntry{
							{AssetID: "asset://missing-logo", Alt: "missing logo"},
						},
						Columns: 1,
					},
				},
			},
		},
	}
}

// storeThreeLogos puts 3 identical 1x1 PNGs into a fresh MemoryStore and
// returns their resolved AssetIDs plus a StoreResolver over the store.
func storeThreeLogos(t *testing.T) ([3]contracts.AssetID, *asset.MemoryStore) {
	t.Helper()
	png := decodeLogoWallPNG(t)
	store := asset.NewMemoryStore()
	var ids [3]contracts.AssetID
	names := [3]string{"acme.png", "globex.png", "initech.png"}
	for i, name := range names {
		stored, err := store.Put(name, "image/png", png)
		if err != nil {
			t.Fatalf("store.Put(%q) error = %v", name, err)
		}
		ids[i] = contracts.AssetID(stored.ID)
	}
	return ids, store
}

// TestLogoWallRendersWithoutError asserts a resolved 3-logo wall renders to
// valid, non-empty PPTX bytes.
func TestLogoWallRendersWithoutError(t *testing.T) {
	t.Parallel()

	ids, store := storeThreeLogos(t)
	resolver := raster.NewStoreResolver(store)

	buf, stats, err := RenderWithAssets(logoWallDoc(ids), soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes")
	}
	assertValidPPTX(t, buf)
	if stats.Slides != 1 {
		t.Fatalf("stats.Slides = %d, want 1", stats.Slides)
	}
	if stats.Assets < 3 {
		t.Errorf("stats.Assets = %d, want >= 3", stats.Assets)
	}
}

// TestLogoWallEmitsMoreShapesThanEmptySlide proves the LogoWall node has a
// render effect, not dead infra: its slide must emit strictly more shapes
// than a blank slide with no nodes.
func TestLogoWallEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()
	ids, store := storeThreeLogos(t)
	resolver := raster.NewStoreResolver(store)

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, wallStats, err := RenderWithAssets(logoWallDoc(ids), s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets(logoWall) error = %v", err)
	}
	if wallStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("LogoWall shapes = %d, want > empty-slide shapes %d", wallStats.Shapes, emptyStats.Shapes)
	}
}

// TestLogoWallMissingAssetWarnsNotPanics asserts an unresolved logo asset
// warns and still renders — the warn-don't-fail contract (RFC §10.2).
func TestLogoWallMissingAssetWarnsNotPanics(t *testing.T) {
	t.Parallel()

	store := asset.NewMemoryStore()
	resolver := raster.NewStoreResolver(store)

	buf, stats, err := RenderWithAssets(missingLogoWallDoc(), soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() with missing asset error = %v, want nil", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes with missing asset")
	}
	if stats.Assets != 0 {
		t.Errorf("stats.Assets = %d, want 0 (no resolved asset)", stats.Assets)
	}
	found := false
	for _, w := range stats.Warnings {
		if strings.Contains(w, "asset") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected an unresolved-asset warning; got %v", stats.Warnings)
	}
}

// TestLogoWallDeterministicAcrossRepeatedRenders asserts byte-identical
// output across two renders of the same doc+soul+resolver.
func TestLogoWallDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	ids, store := storeThreeLogos(t)
	resolver := raster.NewStoreResolver(store)
	doc := logoWallDoc(ids)
	s := soul.DeckardWhite()

	first, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("first RenderWithAssets() error = %v", err)
	}
	second, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("second RenderWithAssets() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("RenderWithAssets() bytes differ across identical renders")
	}
}

// TestLogoWallDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestLogoWallDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	ids, store := storeThreeLogos(t)
	resolver := raster.NewStoreResolver(store)
	doc := logoWallDoc(ids)
	s := soul.DeckardWhite()

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("render bytes differ across worker counts")
	}
}
