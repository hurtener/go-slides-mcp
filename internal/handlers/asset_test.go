package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestAssetHandlersRoundTripMetadataOnly(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()
	encoded := base64.StdEncoding.EncodeToString([]byte("hello asset"))

	uploaded, err := h.uploadAsset(ctx, contracts.UploadAssetInput{Filename: "hello.txt", MIME: "text/plain", DataBase64: encoded})
	if err != nil {
		t.Fatalf("uploadAsset: %v", err)
	}
	if uploaded.Structured.AssetID == "" {
		t.Fatal("uploadAsset returned empty asset id")
	}
	if uploaded.Structured.Bytes != len("hello asset") {
		t.Fatalf("uploadAsset bytes = %d, want %d", uploaded.Structured.Bytes, len("hello asset"))
	}

	listed, err := h.listAssets(ctx, contracts.ListAssetsInput{})
	if err != nil {
		t.Fatalf("listAssets: %v", err)
	}
	if len(listed.Structured.Assets) != 1 {
		t.Fatalf("listAssets len = %d, want 1", len(listed.Structured.Assets))
	}
	if listed.Structured.Assets[0].AssetID != uploaded.Structured.AssetID {
		t.Fatalf("listAssets asset id = %q, want %q", listed.Structured.Assets[0].AssetID, uploaded.Structured.AssetID)
	}

	got, err := h.getAsset(ctx, contracts.GetAssetInput{AssetID: uploaded.Structured.AssetID})
	if err != nil {
		t.Fatalf("getAsset: %v", err)
	}
	if got.Structured.Filename != "hello.txt" || got.Structured.MIME != "text/plain" {
		t.Fatalf("getAsset got %+v", got.Structured)
	}
	if stored, ok := h.deps.Assets.Get(uploaded.Structured.AssetID); !ok || string(stored.Bytes) != "hello asset" {
		t.Fatalf("store bytes = %q, want %q", string(stored.Bytes), "hello asset")
	}

	deleted, err := h.deleteAsset(ctx, contracts.DeleteAssetInput{AssetID: uploaded.Structured.AssetID})
	if err != nil {
		t.Fatalf("deleteAsset: %v", err)
	}
	if !deleted.Structured.Deleted {
		t.Fatal("deleteAsset Deleted = false, want true")
	}
	if _, ok := h.deps.Assets.Get(uploaded.Structured.AssetID); ok {
		t.Fatal("asset still present after delete")
	}
	if err := h.deps.Assets.Delete(uploaded.Structured.AssetID); !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("store Delete after delete err = %v, want ErrNotFound", err)
	}
}

func TestAssetHandlersMissingReturnsNotFound(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	if _, err := h.getAsset(ctx, contracts.GetAssetInput{AssetID: "asset://missing"}); !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("getAsset missing err = %v, want ErrNotFound", err)
	}
	if _, err := h.deleteAsset(ctx, contracts.DeleteAssetInput{AssetID: "asset://missing"}); !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("deleteAsset missing err = %v, want ErrNotFound", err)
	}
}
