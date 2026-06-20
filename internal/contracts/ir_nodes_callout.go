package contracts

// CalloutKind selects a callout variant (mirrors pptx-go's scene.CalloutKind).
type CalloutKind string

// Callout variants.
const (
	CalloutNote      CalloutKind = "note"
	CalloutWarning   CalloutKind = "warning"
	CalloutTip       CalloutKind = "tip"
	CalloutImportant CalloutKind = "important"
)

// IsValid reports whether v is one of the closed CalloutKind wire values
// (Phase 12 A4). The validator additionally accepts "" (empty) so an
// unset/omitempty field defaults at render time.
func (v CalloutKind) IsValid() bool { return IsValidEnum(v, AllowedCalloutKind()) }

// Callout is a highlighted note with a title and rich body. Mirror of
// scene.Callout. The JSON field for the variant is "calloutKind" (not
// "kind", which is the node discriminator).
type Callout struct {
	// Kind is the callout variant.
	Kind CalloutKind `json:"calloutKind,omitempty"`
	// Title is the callout heading.
	Title string `json:"title,omitempty"`
	// Body is the callout body content.
	Body RichText `json:"body,omitempty"`
}

func (Callout) slideNodeKind() Kind { return KindCallout }

// MarshalJSON injects the "callout" kind discriminator via marshalNode.
func (c *Callout) MarshalJSON() ([]byte, error) { return marshalNode(KindCallout, *c) }

func init() { registerNodeKind(KindCallout, func() SlideNode { return &Callout{} }) }
