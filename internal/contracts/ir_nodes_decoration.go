package contracts

// DecorationKind selects a decoration's render path (mirrors pptx-go's
// scene.DecorationKind). A preset decoration renders as native shapes; an
// asset decoration renders as a PPTX picture (D-018).
type DecorationKind string

// Decoration kinds (wire values per compose-a-scene).
const (
	DecorationPreset DecorationKind = "preset"
	DecorationAsset  DecorationKind = "asset"
	// DecorationText renders an oversized, low-opacity ghost number/word
	// behind body content (R13.9) — e.g. a large "03" watermark.
	DecorationText DecorationKind = "text"
)

// Layer selects a decoration's z-order layer (mirrors pptx-go's
// scene.Layer).
type Layer string

// Decoration layers (wire values per compose-a-scene).
const (
	LayerBackground Layer = "background"
	LayerForeground Layer = "foreground"
)

// Decoration is an ornamental element layered over a slide. A preset
// decoration renders as native shapes; an asset decoration renders as a
// PPTX picture from AssetID-resolved bytes. Mirror of pptx-go's
// scene.Decoration. The JSON field for the variant is "decorationKind"
// (not "kind", which is the node discriminator — CONVENTIONS §2).
// Stage-1 validation: preset needs a Preset, asset needs an AssetID,
// Opacity in [0,1] (Opacity 0 = opaque) — later unit.
type Decoration struct {
	// Kind is the decoration variant (preset or asset).
	Kind DecorationKind `json:"decorationKind,omitempty"`
	// Preset is the curated ornament name (preset kind only).
	Preset string `json:"preset,omitempty"`
	// AssetID is the resolver key for the ornament bytes (asset kind only).
	AssetID AssetID `json:"assetId,omitempty"`
	// Layer is the z-order layer.
	Layer Layer `json:"layer,omitempty"`
	// Anchor is the slide anchor the Offset is relative to.
	Anchor Anchor `json:"anchor,omitempty"`
	// Offset is the position offset from the anchor.
	Offset Position `json:"offset,omitempty"`
	// Size is the decoration extent.
	Size Size `json:"size,omitempty"`
	// Bleed extends the decoration to the slide edges.
	Bleed bool `json:"bleed,omitempty"`
	// Opacity is the decoration opacity in [0,1]; 0 means opaque.
	Opacity float64 `json:"opacity,omitempty"`
	// Rotation is the clockwise rotation in degrees.
	Rotation float64 `json:"rotation,omitempty"`
	// Color overrides the ornament's color role (R13.5). Empty = the engine
	// default (accent). Set to `canvas`/`paper` for a neutral paper grain or
	// `surface`/an inverse role for a pale starfield on dark slides. Applies
	// to preset decorations; asset decorations ignore it.
	Color ColorRole `json:"color,omitempty"`
	// Pitch is the lattice spacing in POINTS for pattern presets
	// (grid_dots/noise_overlay/starfield); the dot count derives from the box
	// at this pitch so a full-bleed texture keeps a consistent density
	// (R13.7). 0 keeps the preset's legacy fixed count — byte-identical.
	Pitch float64 `json:"pitch,omitempty"`
	// Text is the watermark string for `decorationKind:"text"` (an oversized
	// ghost number/word, e.g. "03") — R13.9. Required for text kind; ignored
	// otherwise.
	Text string `json:"text,omitempty"`
	// FontSize is the watermark text size in POINTS for text kind (R13.9);
	// 0 uses a box-height "fill the box" default. Ignored by other kinds.
	FontSize float64 `json:"fontSize,omitempty"`
}

func (Decoration) slideNodeKind() Kind { return KindDecoration }

// MarshalJSON injects the "decoration" kind discriminator via marshalNode.
func (d *Decoration) MarshalJSON() ([]byte, error) { return marshalNode(KindDecoration, *d) }

func init() { registerNodeKind(KindDecoration, func() SlideNode { return &Decoration{} }) }
