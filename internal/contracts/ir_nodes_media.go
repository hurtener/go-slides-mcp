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
}

func (Image) slideNodeKind() Kind { return KindImage }

// MarshalJSON injects the "image" kind discriminator via marshalNode.
func (im *Image) MarshalJSON() ([]byte, error) { return marshalNode(KindImage, *im) }

func init() { registerNodeKind(KindImage, func() SlideNode { return &Image{} }) }

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
