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
	return nil
}

func init() { registerNodeKind(KindCard, func() SlideNode { return &Card{} }) }
