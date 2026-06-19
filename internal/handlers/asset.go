package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func (h *handlers) uploadAsset(_ context.Context, in contracts.UploadAssetInput) (tool.Result[contracts.UploadAssetOutput], error) {
	data, err := base64.StdEncoding.DecodeString(in.DataBase64)
	if err != nil {
		return tool.Result[contracts.UploadAssetOutput]{}, fmt.Errorf("decode asset base64: %w", err)
	}
	stored, err := h.deps.Assets.Put(in.Filename, in.MIME, data)
	if err != nil {
		return tool.Result[contracts.UploadAssetOutput]{}, err
	}
	out := contracts.UploadAssetOutput{AssetID: stored.ID, Filename: stored.Filename, MIME: stored.MIME, Bytes: len(stored.Bytes)}
	return tool.Result[contracts.UploadAssetOutput]{Text: fmt.Sprintf("Uploaded asset %q (%d bytes).", stored.ID, len(stored.Bytes)), Structured: out}, nil
}

func (h *handlers) listAssets(_ context.Context, _ contracts.ListAssetsInput) (tool.Result[contracts.ListAssetsOutput], error) {
	stored := h.deps.Assets.List()
	out := contracts.ListAssetsOutput{Assets: make([]contracts.AssetMetadata, 0, len(stored))}
	for _, item := range stored {
		out.Assets = append(out.Assets, assetMetadata(item))
	}
	return tool.Result[contracts.ListAssetsOutput]{Text: fmt.Sprintf("Found %d asset(s).", len(out.Assets)), Structured: out}, nil
}

func (h *handlers) getAsset(_ context.Context, in contracts.GetAssetInput) (tool.Result[contracts.GetAssetOutput], error) {
	stored, ok := h.deps.Assets.Get(in.AssetID)
	if !ok {
		return tool.Result[contracts.GetAssetOutput]{}, mapAssetNotFound(in.AssetID)
	}
	out := contracts.GetAssetOutput{AssetID: stored.ID, Filename: stored.Filename, MIME: stored.MIME, Bytes: len(stored.Bytes)}
	return tool.Result[contracts.GetAssetOutput]{Text: fmt.Sprintf("Loaded asset %q.", stored.ID), Structured: out}, nil
}

func (h *handlers) deleteAsset(_ context.Context, in contracts.DeleteAssetInput) (tool.Result[contracts.DeleteAssetOutput], error) {
	if err := h.deps.Assets.Delete(in.AssetID); err != nil {
		if errors.Is(err, asset.ErrNotFound) {
			return tool.Result[contracts.DeleteAssetOutput]{}, mapAssetNotFound(in.AssetID)
		}
		return tool.Result[contracts.DeleteAssetOutput]{}, err
	}
	out := contracts.DeleteAssetOutput{AssetID: in.AssetID, Deleted: true}
	return tool.Result[contracts.DeleteAssetOutput]{Text: fmt.Sprintf("Deleted asset %q.", in.AssetID), Structured: out}, nil
}

func assetMetadata(item *asset.Asset) contracts.AssetMetadata {
	return contracts.AssetMetadata{AssetID: item.ID, Filename: item.Filename, MIME: item.MIME, Bytes: len(item.Bytes)}
}

func mapAssetNotFound(id string) error {
	return fmt.Errorf("asset %q not found: %w", id, asset.ErrNotFound)
}
