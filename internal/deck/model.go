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

// Chrome is the deck-level slide chrome.
type Chrome struct {
	// Header is the header text.
	Header string `json:"header,omitempty"`
	// Footer is the footer text.
	Footer string `json:"footer,omitempty"`
	// ShowOnCover controls whether chrome appears on the cover slide.
	ShowOnCover bool `json:"showOnCover,omitempty"`
}

// Section is one named grouping of slide IDs.
type Section struct {
	// Name is the section title.
	Name string `json:"name,omitempty"`
	// SlideIDs is the ordered set of slide IDs in the section.
	SlideIDs []string `json:"slideIds,omitempty"`
}
