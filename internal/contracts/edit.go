package contracts

import "encoding/json"

// IRPath is a structural path into a slide's node tree.
type IRPath = []any

// EditSlideNodeInput is the typed input for edit_slide_node.
type EditSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node to replace.
	Path IRPath `json:"path,omitempty"`
	// Node is the replacement slide node encoded as raw JSON with a kind discriminator.
	Node json.RawMessage `json:"node"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// EditSlideNodeOutput is the structured result for edit_slide_node.
type EditSlideNodeOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// InsertSlideNodeInput is the typed input for insert_slide_node.
type InsertSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the insertion point in a node slice.
	Path IRPath `json:"path,omitempty"`
	// Node is the inserted slide node encoded as raw JSON with a kind discriminator.
	Node json.RawMessage `json:"node"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// InsertSlideNodeOutput is the structured result for insert_slide_node.
type InsertSlideNodeOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// RemoveSlideNodeInput is the typed input for remove_slide_node.
type RemoveSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node to remove.
	Path IRPath `json:"path,omitempty"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// RemoveSlideNodeOutput is the structured result for remove_slide_node.
type RemoveSlideNodeOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// DuplicateSlideNodeInput is the typed input for duplicate_slide_node.
type DuplicateSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node to duplicate.
	Path IRPath `json:"path,omitempty"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// DuplicateSlideNodeOutput is the structured result for duplicate_slide_node.
type DuplicateSlideNodeOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// MoveSlideNodeInput is the typed input for move_slide_node.
type MoveSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// From addresses the existing node to move.
	From IRPath `json:"from,omitempty"`
	// To addresses the destination insertion point.
	To IRPath `json:"to,omitempty"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// MoveSlideNodeOutput is the structured result for move_slide_node.
type MoveSlideNodeOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}
