package contracts

// SurfaceKind identifies which UI surface a surface tool's structured output
// is shaped for. The discriminator lets a single dispatcher branch on the
// payload shape without sniffing its fields.
type SurfaceKind string

const (
	// SurfaceKindOverview marks the deck-overview hydration payload.
	SurfaceKindOverview SurfaceKind = "overview"
	// SurfaceKindEditor marks the single-slide editor payload.
	SurfaceKindEditor SurfaceKind = "editor"
	// SurfaceKindState marks the aggregate deck-state hydration payload.
	SurfaceKindState SurfaceKind = "state"
	// SurfaceKindActiveWorkspace marks the active-workspace session payload.
	SurfaceKindActiveWorkspace SurfaceKind = "active_workspace"
)

// GetDeckOverviewInput is the typed input for get_deck_overview.
type GetDeckOverviewInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
}

// DeckOverviewSection is one named section of the deck overview.
type DeckOverviewSection struct {
	// Name is the section title.
	Name string `json:"name,omitempty"`
	// SlideIDs is the ordered set of slide IDs in the section.
	SlideIDs []string `json:"slideIds,omitempty"`
}

// GetDeckOverviewOutput is the structured result for get_deck_overview.
type GetDeckOverviewOutput struct {
	// Kind identifies this payload as the overview surface result.
	Kind SurfaceKind `json:"kind"`
	// State is the four-state page state (ready | empty | error | permission | loading).
	State string `json:"state,omitempty"`
	// Message is the human-readable note for empty/error/permission states.
	Message string `json:"message,omitempty"`
	// DeckID is the resolved deck identifier.
	DeckID string `json:"deckId"`
	// Title is the deck title at fetch time.
	Title string `json:"title,omitempty"`
	// Sections is the deck's named section grouping.
	Sections []DeckOverviewSection `json:"sections,omitempty"`
	// Slides is the ordered preview summary of the deck's slides.
	Slides []SlideSummary `json:"slides,omitempty"`
	// Brand is the white-label brand config for the surface chrome/theme.
	Brand AppBrand `json:"brand"`
}

// OpenSlideEditorInput is the typed input for open_slide_editor.
type OpenSlideEditorInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SlideID is the stable slide identifier to open in the editor.
	SlideID string `json:"slideId"`
}

// OpenSlideEditorOutput is the structured result for open_slide_editor.
type OpenSlideEditorOutput struct {
	// Kind identifies this payload as the editor surface result.
	Kind SurfaceKind `json:"kind"`
	// State is the four-state page state (ready | empty | error | permission | loading).
	State string `json:"state,omitempty"`
	// Message is the human-readable note for empty/error/permission states.
	Message string `json:"message,omitempty"`
	// DeckID is the deck the slide belongs to (needed for the node-edit tools).
	DeckID string `json:"deckId"`
	// SlideID is the slide identifier opened in the editor.
	SlideID string `json:"slideId"`
	// IR is the slide's full IR snapshot for the editor.
	IR Slide `json:"ir"`
	// SoulID is the design soul the deck is bound to.
	SoulID string `json:"soulId,omitempty"`
	// Validation is the structural validation result for the slide IR.
	Validation SlideValidation `json:"validation"`
	// Brand is the white-label brand config for the surface chrome/theme.
	Brand AppBrand `json:"brand"`
	// Layout is the server-computed node geometry for the canvas (EMU boxes,
	// mirroring the renderer) — the editor paints this, never re-computes layout.
	Layout SlideLayout `json:"layout"`
	// Palette is the deck soul's resolved colors + fonts so the canvas paints in
	// the deck's visual language (matching the export).
	Palette SoulPalette `json:"palette"`
}

// SetActiveWorkspaceInput is the typed input for set_active_workspace.
type SetActiveWorkspaceInput struct {
	// DeckID is the optional deck ID to mark as the active workspace deck.
	DeckID string `json:"deckId,omitempty"`
	// SoulID is the optional soul ID to mark as the active workspace soul.
	SoulID string `json:"soulId,omitempty"`
}

// SetActiveWorkspaceOutput is the structured result for set_active_workspace.
type SetActiveWorkspaceOutput struct {
	// Kind identifies this payload as the active-workspace write result.
	Kind SurfaceKind `json:"kind"`
	// ActiveDeckID is the new active deck identifier.
	ActiveDeckID string `json:"activeDeckId,omitempty"`
	// ActiveSoulID is the new active soul identifier.
	ActiveSoulID string `json:"activeSoulId,omitempty"`
}
