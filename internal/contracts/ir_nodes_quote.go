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
	// Testimonial enrichments (R14.5, D-120). Each zero value omits its
	// element, so a Quote with only Text+Attribution renders byte-for-byte
	// as before. When any of these is set the enriched testimonial layout
	// runs: an optional oversized quotation Mark behind the text, an
	// optional rounded Avatar, a structured attribution (Name/Role/Company),
	// and an optional customer Logo.
	//
	// Mark draws a large, low-emphasis quotation glyph behind the quote text.
	Mark bool `json:"mark,omitempty"`
	// AvatarAssetID is the author's avatar (resolved via the AssetResolver,
	// drawn as a rounded picture); "" = no avatar.
	AvatarAssetID AssetID `json:"avatarAssetId,omitempty"`
	// AttributionName is the structured attribution's name; when set it
	// supersedes the flat Attribution string in the enriched layout.
	AttributionName string `json:"attributionName,omitempty"`
	// AttributionRole is the structured attribution's role/title.
	AttributionRole string `json:"attributionRole,omitempty"`
	// AttributionCompany is the structured attribution's company/org.
	AttributionCompany string `json:"attributionCompany,omitempty"`
	// LogoAssetID is the customer/brand logo (resolved via the
	// AssetResolver); "" = no logo.
	LogoAssetID AssetID `json:"logoAssetId,omitempty"`
}

func (Quote) slideNodeKind() Kind { return KindQuote }

// MarshalJSON injects the "quote" kind discriminator via marshalNode.
func (q *Quote) MarshalJSON() ([]byte, error) { return marshalNode(KindQuote, *q) }

func init() { registerNodeKind(KindQuote, func() SlideNode { return &Quote{} }) }
