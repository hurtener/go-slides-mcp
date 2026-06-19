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
}

// GetDeckStateInput is the typed input for get_deck_state.
type GetDeckStateInput struct {
	// DeckID addresses the deck by stable ID or slug.
	DeckID string `json:"deckId"`
	// SelectedSlideID is the optional currently-focused slide identifier
	// surfaced in the State payload.
	SelectedSlideID string `json:"selectedSlideId,omitempty"`
}

// DeckStateSelection is the focused slide descriptor surfaced in the
// get_deck_state payload.
type DeckStateSelection struct {
	// SlideID is the currently-focused slide identifier.
	SlideID string `json:"slideId"`
	// Layout is the focused slide's layout kind, when known.
	Layout LayoutKind `json:"layout,omitempty"`
	// Title is a best-effort headline extracted from the focused slide.
	Title string `json:"title,omitempty"`
}

// GetDeckStateOutput is the structured result for get_deck_state.
type GetDeckStateOutput struct {
	// Kind identifies this payload as the aggregate state surface result.
	Kind SurfaceKind `json:"kind"`
	// DeckID is the resolved deck identifier.
	DeckID string `json:"deckId"`
	// Slides is the ordered preview summary of the deck's slides.
	Slides []SlideSummary `json:"slides,omitempty"`
	// Selected is the focused slide descriptor when one is selected.
	Selected *DeckStateSelection `json:"selected,omitempty"`
	// Souls is every stored soul summary available to the deck.
	Souls []SoulSummary `json:"souls,omitempty"`
	// Validation is the aggregated validation result for the deck.
	Validation ValidateDeckForExportOutput `json:"validation"`
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
