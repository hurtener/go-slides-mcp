package contracts

// UploadAssetInput is the typed input for upload_asset.
type UploadAssetInput struct {
	// Filename is the original client-facing filename for the uploaded asset.
	Filename string `json:"filename"`
	// MIME is the asset media type, such as image/png.
	MIME string `json:"mime"`
	// DataBase64 is the raw asset content encoded as base64.
	DataBase64 string `json:"dataBase64"`
	// Origin is the optional caller surface identifier.
	Origin string `json:"origin,omitempty"`
}

// UploadAssetOutput is the structured result for upload_asset.
type UploadAssetOutput struct {
	// AssetID is the opaque stored asset URI.
	AssetID string `json:"assetId"`
	// Filename is the stored filename metadata.
	Filename string `json:"filename,omitempty"`
	// MIME is the stored asset media type.
	MIME string `json:"mime,omitempty"`
	// Bytes is the stored asset size in bytes.
	Bytes int `json:"bytes"`
}

// ListAssetsInput is the typed input for list_assets.
type ListAssetsInput struct{}

// AssetMetadata is the metadata summary for one stored asset.
type AssetMetadata struct {
	// AssetID is the opaque stored asset URI.
	AssetID string `json:"assetId"`
	// Filename is the original client-facing filename.
	Filename string `json:"filename,omitempty"`
	// MIME is the stored asset media type.
	MIME string `json:"mime,omitempty"`
	// Bytes is the stored asset size in bytes.
	Bytes int `json:"bytes"`
}

// ListAssetsOutput is the structured result for list_assets.
type ListAssetsOutput struct {
	// Assets is every stored asset metadata summary.
	Assets []AssetMetadata `json:"assets,omitempty"`
}

// GetAssetInput is the typed input for get_asset.
type GetAssetInput struct {
	// AssetID addresses one stored asset by opaque asset URI.
	AssetID string `json:"assetId"`
}

// GetAssetOutput is the structured result for get_asset.
type GetAssetOutput struct {
	// AssetID is the opaque stored asset URI.
	AssetID string `json:"assetId"`
	// Filename is the original client-facing filename.
	Filename string `json:"filename,omitempty"`
	// MIME is the stored asset media type.
	MIME string `json:"mime,omitempty"`
	// Bytes is the stored asset size in bytes.
	Bytes int `json:"bytes"`
}

// DeleteAssetInput is the typed input for delete_asset.
type DeleteAssetInput struct {
	// AssetID addresses one stored asset by opaque asset URI.
	AssetID string `json:"assetId"`
}

// DeleteAssetOutput is the structured result for delete_asset.
type DeleteAssetOutput struct {
	// AssetID is the deleted stored asset URI.
	AssetID string `json:"assetId"`
	// Deleted reports whether the asset was removed.
	Deleted bool `json:"deleted"`
}
