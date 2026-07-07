package contracts

import "encoding/json"

// Banner is a full-width filled "big takeaway / promo / CTA" strip
// (R12.6, D-097): a leading icon + a bold lead phrase + a supporting body
// on the left, with optional right-aligned Trailing children (a Stat
// and/or a Button). Distinct from the side-bar Callout — the banner is a
// wide, full-fill band. Mirror of pptx-go's scene.Banner.
//
// Fill is the strip color; an empty string leaves the engine's zero
// (ColorCanvas) in effect, which the engine renderer promotes to
// ColorAccent (a banner is always a filled strip — a canvas-colored one
// would be invisible). TextColor colors the lead/body; an empty string
// leaves the engine's zero (TextPrimary) in effect, which the renderer
// auto-contrasts against Fill (light on a dark fill), and any explicit
// non-default value is honored. Trailing children render in a right
// region per their own policy. Additive: a deck with no Banner is
// byte-identical.
type Banner struct {
	// Lead is the bold headline phrase shown left-of-center.
	Lead RichText `json:"lead,omitempty"`
	// Body is the supporting copy shown below Lead (also left-aligned).
	Body RichText `json:"body,omitempty"`
	// Icon is a leading curated/extension icon name; "" = none.
	Icon string `json:"icon,omitempty"`
	// Fill is the strip fill color role; "" = engine zero (Canvas → promoted to Accent).
	Fill ColorRole `json:"fill,omitempty"`
	// TextColor colors the lead/body; "" = engine zero (Primary, auto-contrasted on Fill).
	TextColor TextColorRole `json:"textColor,omitempty"`
	// Trailing children render right-aligned in their own region (Stat/Button/Lockup); nil = none.
	Trailing []SlideNode `json:"trailing,omitempty"`
}

func (Banner) slideNodeKind() Kind { return KindBanner }

// MarshalJSON injects the "banner" kind; Trailing marshals through each
// child's own MarshalJSON (the kind injected per child, CONVENTIONS §3).
func (b *Banner) MarshalJSON() ([]byte, error) { return marshalNode(KindBanner, *b) }

// UnmarshalJSON dispatches Trailing through UnmarshalSlideNode so the
// banner's right-aligned children nest recursively (CONVENTIONS §3 — the
// same pattern Card uses for Body, TwoColumn uses for Left/Right).
func (b *Banner) UnmarshalJSON(data []byte) error {
	type raw struct {
		Lead      RichText          `json:"lead,omitempty"`
		Body      RichText          `json:"body,omitempty"`
		Icon      string            `json:"icon,omitempty"`
		Fill      ColorRole         `json:"fill,omitempty"`
		TextColor TextColorRole     `json:"textColor,omitempty"`
		Trailing  []json.RawMessage `json:"trailing,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	b.Lead = r.Lead
	b.Body = r.Body
	b.Icon = r.Icon
	b.Fill = r.Fill
	b.TextColor = r.TextColor
	trailing, err := unmarshalNodes(r.Trailing)
	if err != nil {
		return err
	}
	b.Trailing = trailing
	return nil
}

func init() { registerNodeKind(KindBanner, func() SlideNode { return &Banner{} }) }
