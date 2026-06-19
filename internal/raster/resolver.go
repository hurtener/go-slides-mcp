package raster

import (
	"context"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/pptx-go/scene"
)

// StoreResolver resolves asset://<id> references against an in-memory asset
// store and satisfies scene.AssetResolver. It is the seam that links the
// upload-style asset store to the scene renderer's pic composes
// (see pptx-go's register-an-asset + scene.URIAssetResolver conventions).
type StoreResolver struct {
	Assets *asset.MemoryStore
}

// NewStoreResolver wraps an asset store in a scene AssetResolver. A nil store
// is treated as empty so callers can pass one unconditionally.
func NewStoreResolver(s *asset.MemoryStore) StoreResolver {
	return StoreResolver{Assets: s}
}

// Resolve returns the stored bytes and the recorded MIME for id, or
// scene.ErrAssetNotFound when no asset matches. Ids may be bare keys or
// "asset://<uuid>" URIs (the store canonicalizes on Put).
func (r StoreResolver) Resolve(_ context.Context, id scene.AssetID) ([]byte, string, error) {
	if r.Assets == nil {
		return nil, "", scene.ErrAssetNotFound
	}
	a, ok := r.Assets.Get(string(id))
	if !ok {
		return nil, "", scene.ErrAssetNotFound
	}
	return a.Bytes, a.MIME, nil
}
