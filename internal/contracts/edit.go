package contracts

// IRPath is a structural path into a slide's node tree.
type IRPath = []any

// EditSlideNodeInput is the typed input for edit_slide_node.
type EditSlideNodeInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node to replace.
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
	// Node is the replacement slide node as a JSON object with a "kind" discriminator.
	Node map[string]any `json:"node"`
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

// EditSlideFieldInput is the typed input for edit_slide_field.
type EditSlideFieldInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node whose field will be replaced.
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
	// Field is the JSON field name to replace on the addressed node.
	Field string `json:"field"`
	// Value is the replacement value for a string-valued field (e.g. a title,
	// label, or eyebrow). For rich text use patch_slide_text; for structured
	// fields (objects/arrays/numbers) replace the whole node via edit_slide_node.
	Value string `json:"value"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// EditSlideFieldOutput is the structured result for edit_slide_field.
type EditSlideFieldOutput struct {
	// Slide is the stored edited slide snapshot.
	Slide Slide `json:"slide"`
	// Validation is the structural validation result for the stored slide.
	Validation SlideValidation `json:"validation"`
}

// PatchSlideTextInput is the typed input for patch_slide_text.
type PatchSlideTextInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to edit.
	SlideID string `json:"slideId"`
	// Path addresses the existing node whose text field will be replaced.
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
	// Field is the RichText JSON field name to replace on the addressed node.
	Field string `json:"field"`
	// Text is the plain text to encode as a single RichText run.
	Text string `json:"text"`
	// ExpectedRevisionHash enforces optimistic concurrency when set.
	ExpectedRevisionHash string `json:"expectedRevisionHash,omitempty"`
}

// PatchSlideTextOutput is the structured result for patch_slide_text.
type PatchSlideTextOutput struct {
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
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
	// Node is the inserted slide node as a JSON object with a "kind" discriminator.
	Node map[string]any `json:"node"`
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
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
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
	Path IRPath `json:"path,omitempty" jsonschema:"path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
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
	From IRPath `json:"from,omitempty" jsonschema:"source path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
	// To addresses the destination insertion point.
	To IRPath `json:"to,omitempty" jsonschema:"destination path to the node as legs from the slide root: the first leg is always nodes, then an integer node index; nested container legs are left/right (two_column), cells (grid), or body (card). NOTE: list items and prose paragraphs are node fields, NOT legs — to change them, replace the whole list/prose node with edit_slide_node. examples: [nodes, 2] or [nodes, 0, left, 1]"`
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
