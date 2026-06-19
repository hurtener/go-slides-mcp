package raster_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/pptx-go/scene"
)

func TestStoreResolverReturnsStoredBytes(t *testing.T) {
	t.Parallel()

	store := asset.NewMemoryStore()
	stored, err := store.Put("hero.png", "image/png", []byte("\x89PNG\r\n\x1a\nframed"))
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}

	resolver := raster.NewStoreResolver(store)
	got, mime, err := resolver.Resolve(context.Background(), scene.AssetID(stored.ID))
	if err != nil {
		t.Fatalf("StoreResolver.Resolve() error = %v", err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q, want %q", mime, "image/png")
	}
	if string(got) != "\x89PNG\r\n\x1a\nframed" {
		t.Errorf("bytes = %q, want stored PNG prefix", got)
	}
}

func TestStoreResolverReportsErrAssetNotFound(t *testing.T) {
	t.Parallel()

	resolver := raster.NewStoreResolver(asset.NewMemoryStore())
	_, _, err := resolver.Resolve(context.Background(), scene.AssetID("asset://missing"))
	if !errors.Is(err, scene.ErrAssetNotFound) {
		t.Errorf("Resolve() err = %v, want scene.ErrAssetNotFound", err)
	}
}

func TestStoreResolverNilStoreReportsErrAssetNotFound(t *testing.T) {
	t.Parallel()

	resolver := raster.NewStoreResolver(nil)
	_, _, err := resolver.Resolve(context.Background(), scene.AssetID("asset://anything"))
	if !errors.Is(err, scene.ErrAssetNotFound) {
		t.Errorf("Resolve() err = %v, want scene.ErrAssetNotFound", err)
	}
}
