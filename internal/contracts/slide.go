package contracts

// SlideValidation is the structural validation result for one slide.
type SlideValidation struct {
	// OK reports whether ValidateSlide returned no issues.
	OK bool `json:"ok"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
}

// AddSlideInput is the typed input for add_slide.
type AddSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// Slide is the slide snapshot to insert. Its ID is ignored on add.
	Slide Slide `json:"slide"`
	// Position is the optional zero-based insertion index.
	Position *int `json:"position,omitempty"`
}

// AddSlideOutput is the structured result for add_slide.
type AddSlideOutput struct {
	// SlideID is the inserted slide identifier assigned by the store.
	SlideID string `json:"slideId"`
	// Slide is the stored slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// UpdateSlideInput is the typed input for update_slide.
type UpdateSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to replace.
	SlideID string `json:"slideId"`
	// Slide is the replacement slide snapshot.
	Slide Slide `json:"slide"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// UpdateSlideOutput is the structured result for update_slide.
type UpdateSlideOutput struct {
	// SlideID is the updated slide identifier.
	SlideID string `json:"slideId"`
	// Slide is the stored replacement slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// GetSlideInput is the typed input for get_slide.
type GetSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to load.
	SlideID string `json:"slideId"`
}

// GetSlideOutput is the structured result for get_slide.
type GetSlideOutput struct {
	// SlideID is the loaded slide identifier.
	SlideID string `json:"slideId"`
	// Slide is the loaded slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the loaded slide.
	Validation SlideValidation `json:"validation"`
}

// RemoveSlideInput is the typed input for remove_slide.
type RemoveSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to delete.
	SlideID string `json:"slideId"`
}

// RemoveSlideOutput is the structured result for remove_slide.
type RemoveSlideOutput struct {
	// DeckID is the updated deck identifier.
	DeckID string `json:"deckId"`
	// Removed reports whether the slide was removed.
	Removed bool `json:"removed"`
}

// ReorderSlidesInput is the typed input for reorder_slides.
type ReorderSlidesInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// Order is the complete ordered list of slide IDs.
	Order []string `json:"order,omitempty"`
}

// ReorderSlidesOutput is the structured result for reorder_slides.
type ReorderSlidesOutput struct {
	// Kind identifies this payload as a deck result.
	Kind DeckKind `json:"kind"`
	// DeckID is the reordered deck identifier.
	DeckID string `json:"deckId"`
	// Slides is the ordered preview summary after reordering.
	Slides []SlideSummary `json:"slides,omitempty"`
}

// DuplicateSlideInput is the typed input for duplicate_slide.
type DuplicateSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to copy.
	SlideID string `json:"slideId"`
	// Position is the optional zero-based insertion index for the copy.
	Position *int `json:"position,omitempty"`
}

// DuplicateSlideOutput is the structured result for duplicate_slide.
type DuplicateSlideOutput struct {
	// SlideID is the inserted duplicate slide identifier.
	SlideID string `json:"slideId"`
	// Slide is the stored duplicate slide snapshot.
	Slide Slide `json:"slide"`
}
