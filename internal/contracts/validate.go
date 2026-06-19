package contracts

// StyleFinding is one structured validation finding behind a StyleScore.
type StyleFinding struct {
	// Category groups the finding: structural, contrast, typography, spacing, token.
	Category string `json:"category"`
	// Severity is "error" (blocks export) or "warning" (advisory).
	Severity string `json:"severity"`
	// Message is the human-readable description.
	Message string `json:"message"`
	// Path is the optional IR node path the finding refers to.
	Path string `json:"path,omitempty"`
}

// ValidateSlideIRInput is the typed input for validate_slide_ir.
type ValidateSlideIRInput struct {
	// Slide is the slide snapshot to validate without storage.
	Slide Slide `json:"slide"`
	// SoulID is the optional soul context; when set, contrast and overflow run
	// against that soul's theme (otherwise only structural checks run).
	SoulID string `json:"soulId,omitempty"`
}

// ValidateSlideIROutput is the structured result for validate_slide_ir.
type ValidateSlideIROutput struct {
	// OK reports whether the slide passed (no errors).
	OK bool `json:"ok"`
	// Score is the weighted StyleScore in [0,1].
	Score float64 `json:"score"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
	// Findings is the structured list of issues with category and severity.
	Findings []StyleFinding `json:"findings,omitempty"`
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
	// OK reports whether the slide passed (no errors).
	OK bool `json:"ok"`
	// Score is the weighted StyleScore in [0,1].
	Score float64 `json:"score"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
	// Findings is the structured list of issues with category and severity.
	Findings []StyleFinding `json:"findings,omitempty"`
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
	// OK reports whether the slide passed (no errors).
	OK bool `json:"ok"`
	// Score is the per-slide weighted StyleScore in [0,1].
	Score float64 `json:"score"`
	// Issues is the flattened list of validation issue messages.
	Issues []string `json:"issues,omitempty"`
}

// ValidateDeckForExportOutput is the structured result for validate_deck_for_export.
type ValidateDeckForExportOutput struct {
	// OK reports whether the deck passed (no errors anywhere).
	OK bool `json:"ok"`
	// Score is the deck-wide weighted StyleScore in [0,1].
	Score float64 `json:"score"`
	// PerSlide is the validation result for each slide in deck order.
	PerSlide []DeckSlideValidation `json:"perSlide,omitempty"`
	// Blockers is the flattened list of slide-scoped export blockers.
	Blockers []string `json:"blockers,omitempty"`
	// Findings is the structured list of deck-wide issues (incl. contrast/overflow).
	Findings []StyleFinding `json:"findings,omitempty"`
}
