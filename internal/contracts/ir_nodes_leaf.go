package contracts

// Hero is a cover/title slide node: an eyebrow, a title, and a subtitle.
// Renders as native PPTX shapes. Mirror of pptx-go's scene.Hero.
type Hero struct {
	// Eyebrow is the small label above the title.
	Eyebrow string `json:"eyebrow,omitempty"`
	// Title is the headline.
	Title string `json:"title,omitempty"`
	// Subtitle is the supporting line under the title.
	Subtitle string `json:"subtitle,omitempty"`
	// Align overrides the slide's horizontal alignment for this block:
	// "left" | "center" | "right". Empty = inherit the slide's align.horizontal.
	Align HAlign `json:"align,omitempty"`
}

func (Hero) slideNodeKind() Kind { return KindHero }

// MarshalJSON injects the "hero" kind discriminator via marshalNode.
func (h *Hero) MarshalJSON() ([]byte, error) { return marshalNode(KindHero, *h) }

func init() { registerNodeKind(KindHero, func() SlideNode { return &Hero{} }) }

// Heading is a typed heading line at a 1..6 depth. Mirror of scene.Heading.
type Heading struct {
	// Text is the heading content.
	Text RichText `json:"text,omitempty"`
	// Level is the heading depth, 1..6.
	Level int `json:"level,omitempty"`
	// Align overrides the slide's horizontal alignment for this block:
	// "left" | "center" | "right". Empty = inherit the slide's align.horizontal.
	Align HAlign `json:"align,omitempty"`
}

func (Heading) slideNodeKind() Kind { return KindHeading }

// MarshalJSON injects the "heading" kind discriminator via marshalNode.
func (h *Heading) MarshalJSON() ([]byte, error) { return marshalNode(KindHeading, *h) }

func init() { registerNodeKind(KindHeading, func() SlideNode { return &Heading{} }) }

// Prose is body text: an ordered list of paragraphs, each a RichText. Mirror
// of scene.Prose.
type Prose struct {
	// Paragraphs is the ordered body text, one RichText per paragraph.
	Paragraphs []RichText `json:"paragraphs,omitempty"`
	// Align overrides the slide's horizontal alignment for this block:
	// "left" | "center" | "right". Empty = inherit the slide's align.horizontal.
	Align HAlign `json:"align,omitempty"`
}

func (Prose) slideNodeKind() Kind { return KindProse }

// MarshalJSON injects the "prose" kind discriminator via marshalNode.
func (p *Prose) MarshalJSON() ([]byte, error) { return marshalNode(KindProse, *p) }

func init() { registerNodeKind(KindProse, func() SlideNode { return &Prose{} }) }
