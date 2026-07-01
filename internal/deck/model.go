package deck

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// Deck is one stored slide deck and its current revision snapshot.
type Deck struct {
	// ID is the stable deck identifier.
	ID string `json:"id"`
	// Slug is the human-readable unique lookup key.
	Slug string `json:"slug"`
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// Author is the deck author.
	Author string `json:"author,omitempty"`
	// SoulID is the design soul applied to the deck.
	SoulID string `json:"soulId,omitempty"`
	// Chrome is the deck-level header/footer chrome configuration.
	Chrome Chrome `json:"chrome,omitempty"`
	// Sections groups slide IDs into named sections.
	Sections []Section `json:"sections,omitempty"`
	// Slides is the deck's slides, in order.
	Slides []contracts.Slide `json:"slides,omitempty"`
	// Revision is the current lowercase-hex content hash.
	Revision string `json:"revision"`
	// CreatedAt is the deck creation timestamp in ISO-8601 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the last mutation timestamp in ISO-8601 format.
	UpdatedAt string `json:"updatedAt"`
}

// Chrome is the deck-level slide chrome configuration (R3).
type Chrome struct {
	// Enabled is the master chrome switch.
	Enabled bool `json:"enabled,omitempty"`
	// BrandAssetID is the footer-left brand image asset id.
	BrandAssetID string `json:"brandAssetId,omitempty"`
	// BrandText is the footer-left brand label text used when BrandAssetID
	// is not set.
	BrandText string `json:"brandText,omitempty"`
}

// Section is one named grouping of slide IDs.
type Section struct {
	// Name is the section title.
	Name string `json:"name,omitempty"`
	// SlideIDs is the ordered set of slide IDs in the section.
	SlideIDs []string `json:"slideIds,omitempty"`
	// Variant is a section-scoped theme variant default (R14.14), applied to
	// every member slide that sets no explicit Variant of its own. Empty =
	// no override.
	Variant string `json:"variant,omitempty"`
	// Archetype is a section-scoped decoration archetype default (R14.14),
	// applied to every member slide that sets no explicit Archetype of its
	// own. Empty = no override.
	Archetype string `json:"archetype,omitempty"`
}
