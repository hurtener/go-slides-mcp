package contracts

// RowTone selects an IconRow's framing (mirrors pptx-go's scene.RowTone, an
// int enum — the product mirror is a string enum). The zero value RowPlain
// draws no frame; RowPill wraps the row in a SurfaceAlt rounded-rect
// (R12.7, D-100).
type RowTone string

// Row tones (wire values per compose-a-scene).
const (
	// RowPlain is the default — no frame.
	RowPlain RowTone = "plain"
	// RowPill wraps the row in a SurfaceAlt rounded-rect.
	RowPill RowTone = "pill"
)

// IconRow is one row of an IconRows (R12.7, D-100): a leading icon, a
// rich label, and an optional right-aligned meta. Icon is a closed-name
// curated/extension icon name; an empty string renders no glyph (the
// label starts at the left).
type IconRow struct {
	// Icon is the leading curated/extension glyph name; "" = no glyph at the left.
	Icon string `json:"icon,omitempty"`
	// Label is the row's rich text.
	Label RichText `json:"label,omitempty"`
	// Meta is the optional right-aligned rich text; nil = none.
	Meta RichText `json:"meta,omitempty"`
	// Tone selects the row's framing; empty = RowPlain.
	Tone RowTone `json:"tone,omitempty"`
}

// IconRows is a vertical stack of [icon | label | optional meta] rows
// (R12.7, D-100): the "integrations / capabilities / sources" list that
// reads as designed rows rather than bullets. Fill distributes inter-row
// spacing so the rows span the box height (like VAlignFill); GlyphColor
// tints every row's icon. Mirror of pptx-go's scene.IconRows.
//
// GlyphColor's empty string leaves the engine's zero (ColorCanvas) in
// effect, which the renderer promotes to ColorAccent (a canvas-colored
// glyph would be invisible). Additive: a deck with no IconRows is
// byte-identical.
type IconRows struct {
	// Rows is the sequence of icon-rows; at least one is required.
	Rows []IconRow `json:"rows,omitempty"`
	// Fill distributes rows to fill the box height; false = top-anchored.
	Fill bool `json:"fill,omitempty"`
	// GlyphColor tints every row's icon; "" = engine zero (Canvas → promoted to Accent).
	GlyphColor ColorRole `json:"glyphColor,omitempty"`
}

func (IconRows) slideNodeKind() Kind { return KindIconRows }

// MarshalJSON injects the "icon_rows" kind discriminator via marshalNode.
// IconRow is a plain concrete sub-struct (no nested SlideNode), so the
// default strictUnmarshal path decodes Rows directly — no custom
// UnmarshalJSON (the datamark leaf pattern).
func (ir *IconRows) MarshalJSON() ([]byte, error) { return marshalNode(KindIconRows, *ir) }

func init() { registerNodeKind(KindIconRows, func() SlideNode { return &IconRows{} }) }
