package contracts

// ButtonTone selects a Button's fill treatment (mirrors pptx-go's
// scene.ButtonTone, an int enum — the product mirror is a string enum; see
// R12.1, D-094). Each tone maps to theme color tokens (P2), so a theme swap
// re-skins every button. The zero value ButtonPrimary is a solid accent pill
// — the default "do this next" affordance. The empty string is accepted
// (acceptEmpty=true) and maps to ButtonPrimary at render time.
type ButtonTone string

// Button tones (wire values per compose-a-scene).
const (
	// ButtonPrimary is a solid ColorAccent fill, inverse label (zero value).
	ButtonPrimary ButtonTone = "primary"
	// ButtonAccentAlt is a solid ColorAccentAlt fill, inverse label.
	ButtonAccentAlt ButtonTone = "accent_alt"
	// ButtonGhost is no fill + an accent hairline outline, accent label.
	ButtonGhost ButtonTone = "ghost"
	// ButtonNeutral is a solid ColorSurfaceAlt fill, default label.
	ButtonNeutral ButtonTone = "neutral"
)

// ButtonSize scales a Button's height, interior padding, and icon size
// (mirrors pptx-go's scene.ButtonSize; see R12.1, D-094). The zero value
// ButtonSizeMD is the default; SM/LG step it down/up. A pinned layout metric,
// not a theme token (it sizes geometry, not a visual property). The empty
// string is accepted (acceptEmpty=true) and maps to ButtonSizeMD at render.
type ButtonSize string

// Button sizes (wire values per compose-a-scene).
const (
	// ButtonSizeMD is the default size.
	ButtonSizeMD ButtonSize = "md"
	// ButtonSizeSM steps the geometry down.
	ButtonSizeSM ButtonSize = "sm"
	// ButtonSizeLG steps the geometry up.
	ButtonSizeLG ButtonSize = "lg"
)

// Button is a presentational CTA / action affordance (R12.1, D-094): a
// content-fit RadiusFull pill with a label and optional leading/trailing
// icons, droppable standalone (a closing slide), inside a card body (a
// pricing card), or inside a banner. It is a shape only — no hyperlink/action
// wiring (the deck is static). Mirror of pptx-go's scene.Button.
//
// Width is content-fit (label + icons + padding) clamped to its box; Align
// offsets the pill within the box (zero = inherit the slide's
// Content.Horizontal). Tone selects the token fill (ghost = outline); Size
// scales the geometry. LeadingIcon / TrailingIcon are closed-name curated or
// extension icons (Stage-1 validated); their zero value ("") renders no
// glyph. Additive: absent ⇒ byte-identical.
type Button struct {
	// Label is the button text.
	Label string `json:"label,omitempty"`
	// Tone selects the token fill treatment; empty = ButtonPrimary (solid accent).
	Tone ButtonTone `json:"tone,omitempty"`
	// Size scales the geometry; empty = ButtonSizeMD (default).
	Size ButtonSize `json:"size,omitempty"`
	// LeadingIcon is a closed-name curated/extension icon; "" = none.
	LeadingIcon string `json:"leadingIcon,omitempty"`
	// TrailingIcon is a closed-name curated/extension icon; "" = none.
	TrailingIcon string `json:"trailingIcon,omitempty"`
	// Align is a per-node horizontal alignment override; 0 = inherit slide.
	Align HAlign `json:"align,omitempty"`
}

func (Button) slideNodeKind() Kind { return KindButton }

// MarshalJSON injects the "button" kind discriminator via marshalNode.
func (b *Button) MarshalJSON() ([]byte, error) { return marshalNode(KindButton, *b) }

func init() { registerNodeKind(KindButton, func() SlideNode { return &Button{} }) }
