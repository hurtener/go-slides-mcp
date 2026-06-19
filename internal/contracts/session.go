package contracts

// GetSessionInput is the typed input for get_session.
type GetSessionInput struct{}

// BuildInfo describes the running server build.
type BuildInfo struct {
	// Name is the server binary name.
	Name string `json:"name"`
	// Version is the server build version.
	Version string `json:"version"`
}

// GetSessionOutput is the structured result for get_session.
type GetSessionOutput struct {
	// ActiveDeckID is the currently active deck, if one is selected.
	ActiveDeckID string `json:"activeDeckId,omitempty"`
	// ActiveSoulID is the currently active soul, if one is selected.
	ActiveSoulID string `json:"activeSoulId,omitempty"`
	// OpenPanels is the ordered list of currently open panel identifiers.
	OpenPanels []string `json:"openPanels"`
	// BuildInfo identifies the running Deckard build.
	BuildInfo BuildInfo `json:"buildInfo"`
}
