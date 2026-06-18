package contracts

// ChipTone selects a chip's visual treatment (mirrors pptx-go's
// scene.ChipTone).
type ChipTone string

// Chip tones (wire values per compose-a-scene).
const (
	ChipTint    ChipTone = "tint"
	ChipSolid   ChipTone = "solid"
	ChipOutline ChipTone = "outline"
)

// Chip is a compact label/badge. Renders as native PPTX shapes. Mirror of
// pptx-go's scene.Chip. The JSON field for the tone is "tone" (the node
// discriminator "kind" is reserved — CONVENTIONS §2).
type Chip struct {
	// Label is the chip text.
	Label string `json:"label,omitempty"`
	// Tone is the chip visual treatment.
	Tone ChipTone `json:"tone,omitempty"`
	// Color is the chip color role.
	Color ColorRole `json:"color,omitempty"`
}

func (Chip) slideNodeKind() Kind { return KindChip }

// MarshalJSON injects the "chip" kind discriminator via marshalNode.
func (c *Chip) MarshalJSON() ([]byte, error) { return marshalNode(KindChip, *c) }

func init() { registerNodeKind(KindChip, func() SlideNode { return &Chip{} }) }

// ArrowDirection selects an arrow's direction (mirrors pptx-go's
// scene.ArrowDirection).
type ArrowDirection string

// Arrow directions (wire values per compose-a-scene).
const (
	ArrowRight ArrowDirection = "right"
	ArrowLeft  ArrowDirection = "left"
	ArrowUp    ArrowDirection = "up"
	ArrowDown  ArrowDirection = "down"
)

// Arrow is a directional connector with an optional label. Renders as native
// PPTX shapes. Mirror of pptx-go's scene.Arrow.
type Arrow struct {
	// Direction is the arrow direction.
	Direction ArrowDirection `json:"direction,omitempty"`
	// Label is the optional arrow label.
	Label string `json:"label,omitempty"`
}

func (Arrow) slideNodeKind() Kind { return KindArrow }

// MarshalJSON injects the "arrow" kind discriminator via marshalNode.
func (a *Arrow) MarshalJSON() ([]byte, error) { return marshalNode(KindArrow, *a) }

func init() { registerNodeKind(KindArrow, func() SlideNode { return &Arrow{} }) }

// SectionDivider is a full-bleed section break with an eyebrow and a label.
// Renders as native PPTX shapes. Mirror of pptx-go's scene.SectionDivider.
type SectionDivider struct {
	// Eyebrow is the small label above the section label.
	Eyebrow string `json:"eyebrow,omitempty"`
	// Label is the section heading.
	Label string `json:"label,omitempty"`
}

func (SectionDivider) slideNodeKind() Kind { return KindSectionDivider }

// MarshalJSON injects the "section_divider" kind discriminator via marshalNode.
func (s *SectionDivider) MarshalJSON() ([]byte, error) {
	return marshalNode(KindSectionDivider, *s)
}

func init() {
	registerNodeKind(KindSectionDivider, func() SlideNode { return &SectionDivider{} })
}
