package contracts

// Quote is a block quotation with an optional attribution. Renders as native
// PPTX shapes. Mirror of pptx-go's scene.Quote.
type Quote struct {
	// Text is the quotation content.
	Text RichText `json:"text,omitempty"`
	// Attribution is the optional source/author line.
	Attribution string `json:"attribution,omitempty"`
	// Align overrides the slide's horizontal alignment for this block:
	// "left" | "center" | "right". Empty = inherit the slide's align.horizontal.
	Align HAlign `json:"align,omitempty"`
}

func (Quote) slideNodeKind() Kind { return KindQuote }

// MarshalJSON injects the "quote" kind discriminator via marshalNode.
func (q *Quote) MarshalJSON() ([]byte, error) { return marshalNode(KindQuote, *q) }

func init() { registerNodeKind(KindQuote, func() SlideNode { return &Quote{} }) }
