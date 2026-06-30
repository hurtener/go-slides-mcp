package contracts

// ExportDeckInput is the typed input for export_deck.
type ExportDeckInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
}

// ExportStats is the structured render summary for one exported deck.
type ExportStats struct {
	// Slides is the number of rendered slides.
	Slides int `json:"slides"`
	// Shapes is the number of rendered shapes.
	Shapes int `json:"shapes"`
	// Warnings is the ordered list of render warnings.
	Warnings []string `json:"warnings,omitempty"`
}

// ExportDeckOutput is the structured result for export_deck.
type ExportDeckOutput struct {
	// Path is the absolute workspace path of the exported .pptx file.
	Path string `json:"path"`
	// ResourceURI is the readable deck:// resource URI for the exported .pptx file.
	ResourceURI string `json:"resourceUri"`
	// SoulID is the design soul actually used to render the export (R8.8):
	// the deck's stored SoulID when it resolves to a brand soul, otherwise
	// the built-in Deckard White default id.
	SoulID string `json:"soulId,omitempty"`
	// BrandSoulEstablished reports whether the export rendered on a real
	// brand soul rather than the built-in Deckard White default (R8.8).
	// False means the deck rendered on the default soul — run bootstrap_soul
	// to establish a brand soul before exporting.
	BrandSoulEstablished bool `json:"brandSoulEstablished"`
	// Stats is the render summary for the export.
	Stats ExportStats `json:"stats"`
}

// ListResourcesInput is the typed input for list_resources.
type ListResourcesInput struct{}

// ResourceSummary is one exported deck resource summary.
type ResourceSummary struct {
	// URI is the deck:// resource URI.
	URI string `json:"uri"`
	// MIME is the resource MIME type.
	MIME string `json:"mime"`
	// Title is the human-readable exported filename.
	Title string `json:"title"`
}

// ListResourcesOutput is the structured result for list_resources.
type ListResourcesOutput struct {
	// Resources is one entry per exported deck file present on disk.
	Resources []ResourceSummary `json:"resources,omitempty"`
}

// GetResourceInput is the typed input for get_resource.
type GetResourceInput struct {
	// URI is the absolute deck:// resource URI to resolve.
	URI string `json:"uri"`
}

// GetResourceOutput is the structured result for get_resource.
type GetResourceOutput struct {
	// URI is the requested deck:// resource URI.
	URI string `json:"uri"`
	// MIME is the resource MIME type.
	MIME string `json:"mime,omitempty"`
	// Path is the absolute workspace path backing the resource.
	Path string `json:"path,omitempty"`
	// Found reports whether the resource exists on disk.
	Found bool `json:"found"`
}
