package contracts

// AssetID names caller-supplied bytes resolved by the render-time
// AssetResolver (mirrors pptx-go's scene.AssetID; see register-an-asset).
// It is free-form: bare keys, content hashes, UUIDs, or "asset://<uuid>"
// URIs — the engine imposes no scheme.
type AssetID string

// Crop is a per-edge image crop expressed as fractions in [0,1] (mirrors
// pptx-go's builder Crop; see compose-a-scene → Validation: each edge in
// [0,1], Left+Right < 1, Top+Bottom < 1). The zero value is no crop.
type Crop struct {
	// Left is the fraction cropped from the left edge, in [0,1].
	Left float64 `json:"left,omitempty"`
	// Top is the fraction cropped from the top edge, in [0,1].
	Top float64 `json:"top,omitempty"`
	// Right is the fraction cropped from the right edge, in [0,1].
	Right float64 `json:"right,omitempty"`
	// Bottom is the fraction cropped from the bottom edge, in [0,1].
	Bottom float64 `json:"bottom,omitempty"`
}

// Fit selects how an image fills its slot (mirrors pptx-go's builder Fit;
// FitFill is the default).
type Fit string

// Fit modes (wire values mirror the builder enum).
const (
	FitFill Fit = "fill"
	FitNone Fit = "none"
)

// Position is a 2D point in slide coordinates (mirrors pptx-go's builder
// position types). Used by Decoration.Offset.
type Position struct {
	// X is the horizontal coordinate (points).
	X float64 `json:"x,omitempty"`
	// Y is the vertical coordinate (points).
	Y float64 `json:"y,omitempty"`
}

// Size is a 2D extent in slide coordinates (mirrors pptx-go's builder size
// types). Used by Decoration.Size.
type Size struct {
	// W is the width (points).
	W float64 `json:"w,omitempty"`
	// H is the height (points).
	H float64 `json:"h,omitempty"`
}

// FrameKind selects a device frame around an image (mirrors pptx-go's
// scene.FrameKind). FrameName, when set, overrides Frame and selects a
// WithFrameExtension frame.
type FrameKind string

// Frame kinds (wire values per compose-a-scene).
const (
	FrameNone    FrameKind = "none"
	FrameBrowser FrameKind = "browser"
	FramePhone   FrameKind = "phone"
	FrameDesktop FrameKind = "desktop"
	FrameLaptop  FrameKind = "laptop"
)

// Anchor names a slide anchor point for a Decoration (mirrors pptx-go's
// scene.Anchor). The Offset is interpreted relative to this anchor.
type Anchor string

// Slide anchor positions (9-point compass; wire values are snake_case).
const (
	AnchorTopLeft     Anchor = "top_left"
	AnchorTop         Anchor = "top"
	AnchorTopRight    Anchor = "top_right"
	AnchorLeft        Anchor = "left"
	AnchorCenter      Anchor = "center"
	AnchorRight       Anchor = "right"
	AnchorBottomLeft  Anchor = "bottom_left"
	AnchorBottom      Anchor = "bottom"
	AnchorBottomRight Anchor = "bottom_right"
)
