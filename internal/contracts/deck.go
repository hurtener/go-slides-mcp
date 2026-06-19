package contracts

// CreateDeckInput is the typed input for create_deck.
type CreateDeckInput struct {
	// Title is the initial deck title.
	Title string `json:"title"`
	// Author is the initial deck author.
	Author string `json:"author,omitempty"`
	// SoulID selects the design soul to apply to the deck.
	SoulID string `json:"soulId,omitempty"`
}

// DeckKind identifies a deck-shaped output payload.
type DeckKind string

const (
	// DeckKindDeck marks a deck payload for future UI dispatch.
	DeckKindDeck DeckKind = "deck"
)

// SlideSummary is the small preview payload for one slide.
type SlideSummary struct {
	// SlideID is the stable slide identifier.
	SlideID string `json:"slideId"`
	// Layout is the slide layout kind.
	Layout LayoutKind `json:"layout,omitempty"`
	// Title is a best-effort headline extracted from the slide content.
	Title string `json:"title,omitempty"`
	// PreviewText is a short secondary text extracted from the slide content.
	PreviewText string `json:"previewText,omitempty"`
	// Revision is the deck revision snapshot this slide summary belongs to.
	Revision string `json:"revision,omitempty"`
}

// DeckChrome is the deck-level header/footer chrome contract.
type DeckChrome struct {
	// Header is the header text shown on slides.
	Header string `json:"header,omitempty"`
	// Footer is the footer text shown on slides.
	Footer string `json:"footer,omitempty"`
	// ShowOnCover controls whether chrome appears on the cover slide.
	ShowOnCover bool `json:"showOnCover,omitempty"`
}

// DeckSection is one named grouping of slide IDs.
type DeckSection struct {
	// Name is the section title.
	Name string `json:"name,omitempty"`
	// SlideIDs is the ordered set of slide IDs in the section.
	SlideIDs []string `json:"slideIds,omitempty"`
}

// CreateDeckOutput is the structured result for create_deck.
type CreateDeckOutput struct {
	// Kind identifies this payload as a deck result.
	Kind DeckKind `json:"kind"`
	// DeckID is the stable deck identifier.
	DeckID string `json:"deckId"`
	// Slug is the human-readable unique lookup key.
	Slug string `json:"slug"`
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// SoulID is the design soul applied to the deck.
	SoulID string `json:"soulId,omitempty"`
	// Slides is the ordered preview summary of the deck's slides.
	Slides []SlideSummary `json:"slides,omitempty"`
}

// DeckSummary is the small list payload for one deck.
type DeckSummary struct {
	// DeckID is the stable deck identifier.
	DeckID string `json:"deckId"`
	// Slug is the human-readable unique lookup key.
	Slug string `json:"slug"`
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// SoulID is the design soul applied to the deck.
	SoulID string `json:"soulId,omitempty"`
	// SlideCount is the number of slides in the deck.
	SlideCount int `json:"slideCount"`
	// UpdatedAt is the last mutation timestamp in ISO-8601 format.
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// ListDecksInput is the typed input for list_decks.
type ListDecksInput struct{}

// ListDecksOutput is the structured result for list_decks.
type ListDecksOutput struct {
	// Decks is every stored deck summary.
	Decks []DeckSummary `json:"decks,omitempty"`
}

// GetDeckInput is the typed input for get_deck.
type GetDeckInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
}

// GetDeckOutput is the structured result for get_deck.
type GetDeckOutput struct {
	// Kind identifies this payload as a deck result.
	Kind DeckKind `json:"kind"`
	// DeckID is the stable deck identifier.
	DeckID string `json:"deckId"`
	// Slug is the human-readable unique lookup key.
	Slug string `json:"slug"`
	// Title is the deck title.
	Title string `json:"title,omitempty"`
	// SoulID is the design soul applied to the deck.
	SoulID string `json:"soulId,omitempty"`
	// Chrome is the deck-level header/footer chrome configuration.
	Chrome DeckChrome `json:"chrome,omitempty"`
	// Sections is the deck's named grouping of slide IDs.
	Sections []DeckSection `json:"sections,omitempty"`
	// Slides is the ordered preview summary of the deck's slides.
	Slides []SlideSummary `json:"slides,omitempty"`
}

// DeleteDeckInput is the typed input for delete_deck.
type DeleteDeckInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
}

// DeleteDeckOutput is the structured result for delete_deck.
type DeleteDeckOutput struct {
	// DeckID is the deleted deck identifier or slug input.
	DeckID string `json:"deckId"`
	// Deleted reports whether the deck was removed.
	Deleted bool `json:"deleted"`
}

// SetDeckChromeInput is the typed input for set_deck_chrome.
type SetDeckChromeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// Chrome is the replacement chrome configuration.
	Chrome DeckChrome `json:"chrome"`
}

// SetDeckChromeOutput is the structured result for set_deck_chrome.
type SetDeckChromeOutput struct {
	// DeckID is the updated deck identifier.
	DeckID string `json:"deckId"`
	// Chrome is the applied deck-level chrome configuration.
	Chrome DeckChrome `json:"chrome,omitempty"`
}

// SetDeckSectionsInput is the typed input for set_deck_sections.
type SetDeckSectionsInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// Sections is the replacement section grouping.
	Sections []DeckSection `json:"sections,omitempty"`
}

// SetDeckSectionsOutput is the structured result for set_deck_sections.
type SetDeckSectionsOutput struct {
	// DeckID is the updated deck identifier.
	DeckID string `json:"deckId"`
	// Sections is the applied section grouping.
	Sections []DeckSection `json:"sections,omitempty"`
}
