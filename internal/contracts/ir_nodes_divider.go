package contracts

// Divider is a horizontal rule with surrounding spacing. Renders as native
// PPTX shapes. Mirror of pptx-go's scene.Divider.
type Divider struct {
	// Spacing is the spacing token role above and below the rule.
	Spacing SpaceRole `json:"spacing,omitempty"`
}

func (Divider) slideNodeKind() Kind { return KindDivider }

// MarshalJSON injects the "divider" kind discriminator via marshalNode.
func (d *Divider) MarshalJSON() ([]byte, error) { return marshalNode(KindDivider, *d) }

func init() { registerNodeKind(KindDivider, func() SlideNode { return &Divider{} }) }
