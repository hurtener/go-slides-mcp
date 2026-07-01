package contracts

// Image is a picture placed from caller-supplied bytes resolved by
// AssetID. Renders as a PPTX picture (D-011). Mirror of pptx-go's
// scene.Image. Stage-1 validation: non-empty AssetID and crop bounds (each
// edge in [0,1], Left+Right < 1, Top+Bottom < 1) — later unit.
type Image struct {
	// AssetID is the resolver key for the image bytes.
	AssetID AssetID `json:"assetId,omitempty"`
	// Alt is the accessibility/alt-text description.
	Alt string `json:"alt,omitempty"`
	// Frame selects a device frame around the image.
	Frame FrameKind `json:"frame,omitempty"`
	// FrameName overrides Frame and selects a WithFrameExtension frame.
	FrameName string `json:"frameName,omitempty"`
	// Crop is the per-edge crop fractions; the zero value is no crop.
	Crop Crop `json:"crop,omitempty"`
	// Fit selects how the image fills its slot; FitFill is the default.
	Fit Fit `json:"fit,omitempty"`
	// CornerRadius rounds the picture's corners from a theme radius token
	// (R13.11). RadiusNone (the zero value) leaves the picture rectangular —
	// byte-identical to a pre-R13.11 Image.
	CornerRadius RadiusRole `json:"cornerRadius,omitempty"`
	// Elevation casts a soft drop shadow on the picture from a theme
	// elevation token (R13.11). ElevationFlat (the zero value) emits no
	// shadow — byte-identical to a pre-R13.11 Image.
	Elevation ElevationRole `json:"elevation,omitempty"`
	// Annotations overlays numbered pins and/or highlight rectangles on the
	// picture (R14.17). A nil Annotations (the zero value) emits nothing —
	// byte-identical to a pre-R14.17 Image.
	Annotations *ImageAnnotations `json:"annotations,omitempty"`
}

func (Image) slideNodeKind() Kind { return KindImage }

// MarshalJSON injects the "image" kind discriminator via marshalNode.
func (im *Image) MarshalJSON() ([]byte, error) { return marshalNode(KindImage, *im) }

func init() { registerNodeKind(KindImage, func() SlideNode { return &Image{} }) }

// ImageAnnotations is an optional overlay on an Image (R14.17): numbered
// pins at fractional coordinates and/or highlight rectangles around regions,
// each drawn as a native soul-styled shape over the picture. Mirror of
// pptx-go's scene.ImageAnnotations.
type ImageAnnotations struct {
	// Pins are numbered callout markers placed at fractional coordinates of
	// the image box, each with an optional off-pin caption.
	Pins []ImagePin `json:"pins,omitempty"`
	// Highlights are rectangles (fractions of the image box) outlined to
	// draw attention to a region.
	Highlights []ImageHighlight `json:"highlights,omitempty"`
}

// ImagePin is a numbered callout marker at (X,Y) in [0,1] of the image box,
// with an optional caption drawn beside it and a leader line from the pin to
// it. Mirror of pptx-go's scene.ImagePin.
type ImagePin struct {
	// X is the pin's horizontal position, a fraction [0,1] of the image box.
	X float64 `json:"x"`
	// Y is the pin's vertical position, a fraction [0,1] of the image box.
	Y float64 `json:"y"`
	// Label is the pin's number/letter (e.g. "1").
	Label string `json:"label,omitempty"`
	// Caption is an optional off-pin caption; empty means no caption/leader.
	Caption string `json:"caption,omitempty"`
	// AccentIndex selects the soul accent color for this pin.
	AccentIndex int `json:"accentIndex,omitempty"`
}

// ImageHighlight is a rectangle (fractions [0,1] of the image box) outlined
// to draw attention to a region. Mirror of pptx-go's scene.ImageHighlight.
type ImageHighlight struct {
	// X is the highlight's left edge, a fraction [0,1] of the image box.
	X float64 `json:"x"`
	// Y is the highlight's top edge, a fraction [0,1] of the image box.
	Y float64 `json:"y"`
	// W is the highlight's width, a fraction [0,1] of the image box.
	W float64 `json:"w"`
	// H is the highlight's height, a fraction [0,1] of the image box.
	H float64 `json:"h"`
	// AccentIndex selects the soul accent color for this highlight.
	AccentIndex int `json:"accentIndex,omitempty"`
}

// CodeBlock is a source-code listing placed from a pre-rasterized image
// resolved by AssetID (D-014). Renders as a PPTX picture. Mirror of
// pptx-go's scene.CodeBlock. Stage-1 validation: non-empty AssetID — later
// unit. Language and Caption are free-form and may be empty.
type CodeBlock struct {
	// AssetID is the resolver key for the rendered code image.
	AssetID AssetID `json:"assetId,omitempty"`
	// Language is the source language label (free-form).
	Language string `json:"language,omitempty"`
	// Caption is the optional listing caption/filename.
	Caption string `json:"caption,omitempty"`
}

func (CodeBlock) slideNodeKind() Kind { return KindCodeBlock }

// MarshalJSON injects the "code_block" kind discriminator via marshalNode.
func (c *CodeBlock) MarshalJSON() ([]byte, error) { return marshalNode(KindCodeBlock, *c) }

func init() { registerNodeKind(KindCodeBlock, func() SlideNode { return &CodeBlock{} }) }

// Chart is a chart placed from a pre-rasterized image resolved by AssetID
// (V1: image-shape; native c:chart is V2 — D-004). Renders as a PPTX
// picture, or a labeled ChartPlaceholder when the asset is unresolved.
// Mirror of pptx-go's scene.Chart. Stage-1 validation: non-empty AssetID —
// later unit.
type Chart struct {
	// AssetID is the resolver key for the chart image bytes.
	AssetID AssetID `json:"assetId,omitempty"`
	// Caption is the optional chart caption.
	Caption string `json:"caption,omitempty"`
}

func (Chart) slideNodeKind() Kind { return KindChart }

// MarshalJSON injects the "chart" kind discriminator via marshalNode.
func (c *Chart) MarshalJSON() ([]byte, error) { return marshalNode(KindChart, *c) }

func init() { registerNodeKind(KindChart, func() SlideNode { return &Chart{} }) }
