package render_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// validPNG is a 1x1 transparent PNG. The decode-config layer reads the
// IHDR width/height so a slot can place it with correct aspect — engine
// skips if it cannot parse the header. We reuse this for render-wire tests.
const validPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

func decodePNG(t *testing.T) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(validPNG)
	if err != nil {
		t.Fatalf("base64 decode PNG: %v", err)
	}
	return b
}

func assetImageDoc(id contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Image render",
		Slides: []contracts.Slide{
			{
				ID:     "hero",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Eyebrow: "P7A", Title: "Asset resolver", Subtitle: "image node"},
					&contracts.Image{AssetID: id, Alt: "Hero image"},
				},
			},
		},
	}
}

func missingAssetImageDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Image render — missing asset",
		Slides: []contracts.Slide{
			{
				ID:     "missing",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Title: "warn-don't-fail"},
					&contracts.Image{AssetID: "asset://missing-banner", Alt: "missing image"},
				},
			},
		},
	}
}

func TestRenderWithAssetsResolvesImage(t *testing.T) {
	t.Parallel()

	png := decodePNG(t)
	store := asset.NewMemoryStore()
	stored, err := store.Put("hero.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}

	doc := assetImageDoc(contracts.AssetID(stored.ID))
	resolver := raster.NewStoreResolver(store)

	buf, stats, err := render.RenderWithAssets(doc, soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes")
	}
	if stats.Slides != 1 {
		t.Errorf("stats.Slides = %d, want 1", stats.Slides)
	}
	if stats.Assets < 1 {
		t.Errorf("stats.Assets = %d, want >= 1", stats.Assets)
	}
	for _, w := range stats.Warnings {
		if strings.Contains(w, stored.ID) || strings.Contains(w, "missing") {
			t.Errorf("unexpected resolution warning for known asset: %q", w)
		}
	}
}

func TestRenderWithAssetsMissingStillRenders(t *testing.T) {
	t.Parallel()

	store := asset.NewMemoryStore()
	resolver := raster.NewStoreResolver(store)
	doc := missingAssetImageDoc()

	buf, stats, err := render.RenderWithAssets(doc, soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() with missing asset error = %v, want nil", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes with missing asset")
	}
	if stats.Slides != 1 {
		t.Errorf("stats.Slides = %d, want 1", stats.Slides)
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

func TestRenderNoResolverIsAssetFreeAndDeterministic(t *testing.T) {
	t.Parallel()

	store := asset.NewMemoryStore()
	png := decodePNG(t)
	stored, err := store.Put("hero.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}

	doc := assetImageDoc(contracts.AssetID(stored.ID))

	first, statsFirst, err := render.Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	second, statsSecond, err := render.Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() second call error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("Render() bytes differ across identical asset-free renders (determinism broken)")
	}
	if statsFirst.Assets != 0 || statsSecond.Assets != 0 {
		t.Errorf("Render() assets want 0 (asset-free path), got %d / %d", statsFirst.Assets, statsSecond.Assets)
	}

	resolver := raster.NewStoreResolver(store)
	third, statsThird, err := render.RenderWithAssets(doc, soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	if bytes.Equal(first, third) {
		t.Fatal("Render() and RenderWithAssets() should differ when an asset is resolved")
	}
	if statsThird.Assets < 1 {
		t.Errorf("RenderWithAssets() assets = %d, want >= 1", statsThird.Assets)
	}
}
