package contracts

// Funnel is a staged conversion/drop-off diagram (R14.11, D-128): N stages
// rendered as progressively narrowing bands, top to bottom, each with a
// label and an optional value caption. Native shapes; pure integer-EMU
// layout → byte-identical across renders/worker counts. Mirror of pptx-go's
// scene.Funnel. A deck with no Funnel is byte-identical (a new node, absent
// until used).
type Funnel struct {
	// Stages are the funnel's bands, top (widest) to bottom (narrowest).
	Stages []FunnelStage `json:"stages,omitempty"`
}

func (Funnel) slideNodeKind() Kind { return KindFunnel }

// MarshalJSON injects the "funnel" kind discriminator via marshalNode.
func (f *Funnel) MarshalJSON() ([]byte, error) { return marshalNode(KindFunnel, *f) }

func init() { registerNodeKind(KindFunnel, func() SlideNode { return &Funnel{} }) }

// FunnelStage is one band of a Funnel (D-128): a label + optional value
// caption + a soul-driven series accent color.
type FunnelStage struct {
	// Label is the stage's headline text.
	Label string `json:"label,omitempty"`
	// Value is an optional caption (e.g. a count or percentage) drawn
	// alongside the label.
	Value string `json:"value,omitempty"`
	// AccentIndex selects a soul-driven series accent color for the stage
	// band (0 = the first accent). A plain int passthrough — 0 is a real
	// value (the first accent), not "unset".
	AccentIndex int `json:"accentIndex,omitempty"`
}
