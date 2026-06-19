package contracts

// ValidateSlideIRInput is the typed input for validate_slide_ir.
type ValidateSlideIRInput struct {
	// Slide is the slide snapshot to validate without storage.
	Slide Slide `json:"slide"`
	// SoulID is the optional soul context for future validation expansion.
	SoulID string `json:"soulId,omitempty"`
}

// ValidateSlideIROutput is the structured result for validate_slide_ir.
type ValidateSlideIROutput struct {
	// OK reports whether the slide passed structural validation.
	OK bool `json:"ok"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
}

// ValidateSlideInput is the typed input for validate_slide.
type ValidateSlideInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to validate.
	SlideID string `json:"slideId"`
}

// ValidateSlideOutput is the structured result for validate_slide.
type ValidateSlideOutput struct {
	// SlideID is the validated slide identifier.
	SlideID string `json:"slideId"`
	// OK reports whether the slide passed structural validation.
	OK bool `json:"ok"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
}

// ValidateDeckForExportInput is the typed input for validate_deck_for_export.
type ValidateDeckForExportInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
}

// DeckSlideValidation is the per-slide validation result for deck export checks.
type DeckSlideValidation struct {
	// SlideID is the validated slide identifier.
	SlideID string `json:"slideId"`
	// OK reports whether the slide passed structural validation.
	OK bool `json:"ok"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
}

// ValidateDeckForExportOutput is the structured result for validate_deck_for_export.
type ValidateDeckForExportOutput struct {
	// OK reports whether every slide in the deck passed validation.
	OK bool `json:"ok"`
	// PerSlide is the validation result for each slide in deck order.
	PerSlide []DeckSlideValidation `json:"perSlide,omitempty"`
	// Blockers is the flattened list of slide-scoped export blockers.
	Blockers []string `json:"blockers,omitempty"`
}
