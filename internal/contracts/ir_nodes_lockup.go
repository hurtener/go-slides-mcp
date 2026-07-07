package contracts

// AssetSide selects where a Lockup's logo sits relative to its caption
// (mirrors pptx-go's scene.AssetSide, an int enum — the product mirror is
// a string enum). The zero value LeadCaption places the caption first
// (caption leads, logo trails); TrailCaption places the logo first. (R12.9,
// D-102.)
type AssetSide string

// Asset sides (wire values per compose-a-scene).
const (
	// LeadCaption places the caption then the logo (zero value).
	LeadCaption AssetSide = "lead_caption"
	// TrailCaption places the logo then the caption.
	TrailCaption AssetSide = "trail_caption"
)

// Lockup is a compact "powered by / in partnership with" attribution mark
// (R12.9, D-102): a caption paired with a small partner logo composed as
// one inline, centerable unit. The mark is either an AssetID (a partner
// logo, resolved via the AssetResolver — renders as a pic) or an Icon (a
// curated/extension glyph — media-free); exactly one is set per render.
// Mirror of pptx-go's scene.Lockup.
//
// MaxHeight is a height bound on the logo — the field's zero value (0)
// lets the engine apply a pinned default. Align positions the whole group
// (zero = inherit the slide's Content.Horizontal). Additive: a deck with
// no Lockup is byte-identical.
type Lockup struct {
	// Caption is a TypeCaption muted caption (e.g. "POWERED BY").
	Caption string `json:"caption,omitempty"`
	// AssetID is the partner-logo AssetID resolved via AssetResolver;
	// empty → uses Icon instead.
	AssetID AssetID `json:"assetId,omitempty"`
	// Icon is a curated/extension glyph used instead of an asset;
	// empty → uses AssetID.
	Icon string `json:"icon,omitempty"`
	// AssetSide places the logo before or after the caption; "" = LeadCaption.
	AssetSide AssetSide `json:"assetSide,omitempty"`
	// MaxHeight bounds the logo height, in points (mapped to pptx.EMU at
	// render). Zero lets the engine apply a pinned default.
	MaxHeight float64 `json:"maxHeight,omitempty"`
	// Align is a per-node horizontal alignment override; 0 = inherit slide.
	Align HAlign `json:"align,omitempty"`
}

func (Lockup) slideNodeKind() Kind { return KindLockup }

// MarshalJSON injects the "lockup" kind discriminator via marshalNode.
// Lockup is a leaf (no nested SlideNode), so the default strictUnmarshal
// path decodes its fields directly — no custom UnmarshalJSON (the
// datamark leaf pattern).
func (l *Lockup) MarshalJSON() ([]byte, error) { return marshalNode(KindLockup, *l) }

func init() { registerNodeKind(KindLockup, func() SlideNode { return &Lockup{} }) }
