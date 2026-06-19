package contracts

// AppBrand is the white-label brand configuration delivered to the UI surfaces.
// It is resolved at server startup from the brand-token JSON (or defaults) and
// lets a deployment re-skin the surfaces without a rebuild.
type AppBrand struct {
	// Title is the product/brand name shown in surface chrome.
	Title string `json:"title,omitempty"`
	// DefaultTheme is the built-in theme id selected by default
	// (deckard-white | deckard-dark | midnight | slate | editorial-sepia).
	DefaultTheme string `json:"defaultTheme,omitempty"`
	// Tokens are per-token "--app-*" overrides (key without the leading "--"),
	// applied over the selected preset.
	Tokens map[string]string `json:"tokens,omitempty"`
	// AllowThemeSwitch shows or hides the theme selector (false = locked brand).
	AllowThemeSwitch bool `json:"allowThemeSwitch"`
}

// ThumbNode is a glanceable descriptor of one IR node, rendered as a block in a
// slide thumbnail. It carries just enough of the IR to compose a recognizable
// preview — never the full node payload.
type ThumbNode struct {
	// Kind is the IR node kind (heading, prose, list, callout, chart, ...).
	Kind string `json:"kind"`
	// Text is the primary text snippet (title, first line, label).
	Text string `json:"text,omitempty"`
	// Detail is an optional secondary snippet (subtitle, attribution).
	Detail string `json:"detail,omitempty"`
	// Count is an item count for repeating nodes (list items, grid cells, steps).
	Count int `json:"count,omitempty"`
	// Accent marks a node that renders in the accent color (callout, chip, accent card).
	Accent bool `json:"accent,omitempty"`
}

// SlidePreview is one slide reduced to a thumbnail-renderable form.
type SlidePreview struct {
	// ID is the stable slide identifier.
	ID string `json:"id"`
	// Index is the zero-based position in the deck.
	Index int `json:"index"`
	// Layout is the slide layout kind (cover, title_content, two_column, ...).
	Layout string `json:"layout"`
	// Title is a best-effort slide title for labelling the thumbnail.
	Title string `json:"title,omitempty"`
	// Nodes are the top-level node descriptors composing the thumbnail.
	Nodes []ThumbNode `json:"nodes,omitempty"`
}

// DeckPreviewInput drives the deck-preview surface.
type DeckPreviewInput struct {
	// DeckID addresses the deck by stable ID or slug (empty = the active deck).
	DeckID string `json:"deckId,omitempty"`
}

// DeckPreviewOutput is the structured payload the deck-preview surface renders.
type DeckPreviewOutput struct {
	// State is the four-state page state (ready | empty | error | permission | loading).
	State string `json:"state"`
	// Message is the human-readable note for empty/error/permission states.
	Message string `json:"message,omitempty"`
	// Brand is the white-label brand configuration for the surface.
	Brand AppBrand `json:"brand"`
	// Deck is the deck-level summary.
	Deck DeckSummary `json:"deck"`
	// Slides are the per-slide thumbnail descriptors in deck order.
	Slides []SlidePreview `json:"slides,omitempty"`
	// ResourceURI is the deck:// export resource for the [download] action.
	ResourceURI string `json:"resourceUri,omitempty"`
}
