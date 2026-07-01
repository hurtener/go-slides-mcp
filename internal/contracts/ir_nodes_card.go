package contracts

import "encoding/json"

// BodyLayout selects how a card body stacks its children (mirrors pptx-go's
// scene.BodyLayout; BodyVertical is the default).
type BodyLayout string

// Body layouts.
const (
	BodyVertical   BodyLayout = "vertical"
	BodyHorizontal BodyLayout = "horizontal"
)

// BorderStyle selects a card border style (mirrors scene.BorderStyle;
// BorderDefault defers to the Outline flag).
type BorderStyle string

// Border styles.
const (
	BorderDefault BorderStyle = "default"
	BorderNone    BorderStyle = "none"
	BorderSolid   BorderStyle = "solid"
	BorderAccent  BorderStyle = "accent"
)

// CardSize selects a card size variant (mirrors scene.CardSize).
type CardSize string

// Card sizes.
const (
	CardSizeMD CardSize = "md"
	CardSizeSM CardSize = "sm"
	CardSizeLG CardSize = "lg"
)

// CardLayout selects a card layout variant (mirrors scene.CardLayout).
type CardLayout string

// Card layouts.
const (
	CardLayoutDefault CardLayout = "default"
	CardLayoutIconTop CardLayout = "iconTop"
)

// Card is an accent card holding a body of child nodes. All fields beyond
// Header/Body are additive (zero values reproduce the default render).
// Mirror of scene.Card. Children nest recursively.
type Card struct {
	// Header is the card title.
	Header string `json:"header,omitempty"`
	// Eyebrow is the small label above the header.
	Eyebrow string `json:"eyebrow,omitempty"`
	// Icon is a closed-name curated/extension icon.
	Icon string `json:"icon,omitempty"`
	// HeaderPill is a pill-shaped badge in the header.
	HeaderPill string `json:"headerPill,omitempty"`
	// Body is the card body children.
	Body []SlideNode `json:"body,omitempty"`
	// BodyLayout stacks the body vertically (default) or horizontally.
	BodyLayout BodyLayout `json:"bodyLayout,omitempty"`
	// Fill is the card surface color role.
	Fill ColorRole `json:"fill,omitempty"`
	// Outline enables a card border.
	Outline bool `json:"outline,omitempty"`
	// BorderStyle selects the border style.
	BorderStyle BorderStyle `json:"borderStyle,omitempty"`
	// Size is the card size variant.
	Size CardSize `json:"size,omitempty"`
	// Layout is the card layout variant.
	Layout CardLayout `json:"layout,omitempty"`
	// Elevation is the card shadow role.
	Elevation ElevationRole `json:"elevation,omitempty"`
	// HeaderFill is the color role of a banded header region drawn above the
	// body (D-054). The body keeps `fill`. Omit (=empty string) to skip the
	// band — an unset value is byte-identical to a pre-Phase-14 card.
	HeaderFill ColorRole `json:"headerFill,omitempty"`
	// StatusDot is the color role of a small status dot placed in the top-right
	// corner of the card (D-054). Useful for "live / won / at-risk" cues.
	// Omit (=empty string) to draw no dot.
	StatusDot ColorRole `json:"statusDot,omitempty"`
	// Watermark is a large, low-opacity label drawn behind the body
	// (D-054) — e.g. "01", "Q4", or a section number. Omit (=empty string)
	// to draw no watermark.
	Watermark string `json:"watermark,omitempty"`
	// Backdrop is a decoration drawn behind the card's box before its fill
	// (R13.10) — a focal glow/halo that tracks the card; typically a
	// center-anchored, bleeding `radial_glow`. nil = none, byte-identical.
	Backdrop *Decoration `json:"backdrop,omitempty"`
	// ImageFill fills the card surface with a cover-fit photo (resolved via
	// the registered asset store) instead of the solid Fill — the
	// image-as-surface treatment for photographic cards (R14.1). "" = solid,
	// byte-identical to a pre-R14.1 card.
	ImageFill AssetID `json:"imageFill,omitempty"`
}

func (Card) slideNodeKind() Kind { return KindCard }

// MarshalJSON injects the "card" kind; Body marshals through each child's
// own MarshalJSON (kind injected per child).
func (c *Card) MarshalJSON() ([]byte, error) { return marshalNode(KindCard, *c) }

// UnmarshalJSON dispatches Body through UnmarshalSlideNode so the card nests
// recursively (CONVENTIONS §3).
func (c *Card) UnmarshalJSON(data []byte) error {
	type raw struct {
		Header      string            `json:"header,omitempty"`
		Eyebrow     string            `json:"eyebrow,omitempty"`
		Icon        string            `json:"icon,omitempty"`
		HeaderPill  string            `json:"headerPill,omitempty"`
		Body        []json.RawMessage `json:"body,omitempty"`
		BodyLayout  BodyLayout        `json:"bodyLayout,omitempty"`
		Fill        ColorRole         `json:"fill,omitempty"`
		Outline     bool              `json:"outline,omitempty"`
		BorderStyle BorderStyle       `json:"borderStyle,omitempty"`
		Size        CardSize          `json:"size,omitempty"`
		Layout      CardLayout        `json:"layout,omitempty"`
		Elevation   ElevationRole     `json:"elevation,omitempty"`
		HeaderFill  ColorRole         `json:"headerFill,omitempty"`
		StatusDot   ColorRole         `json:"statusDot,omitempty"`
		Watermark   string            `json:"watermark,omitempty"`
		Backdrop    *Decoration       `json:"backdrop,omitempty"`
		ImageFill   AssetID           `json:"imageFill,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	c.Header = r.Header
	c.Eyebrow = r.Eyebrow
	c.Icon = r.Icon
	c.HeaderPill = r.HeaderPill
	body, err := unmarshalNodes(r.Body)
	if err != nil {
		return err
	}
	c.Body = body
	c.BodyLayout = r.BodyLayout
	c.Fill = r.Fill
	c.Outline = r.Outline
	c.BorderStyle = r.BorderStyle
	c.Size = r.Size
	c.Layout = r.Layout
	c.Elevation = r.Elevation
	c.HeaderFill = r.HeaderFill
	c.StatusDot = r.StatusDot
	c.Watermark = r.Watermark
	c.Backdrop = r.Backdrop
	c.ImageFill = r.ImageFill
	return nil
}

func init() { registerNodeKind(KindCard, func() SlideNode { return &Card{} }) }
