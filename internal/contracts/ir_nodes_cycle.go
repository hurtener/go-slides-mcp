package contracts

// Cycle is a closed-loop process diagram (R14.11, D-128): N stages placed
// evenly on a ring with directional connectors showing the loop. Native
// shapes; pure integer-EMU layout → byte-identical across renders/worker
// counts. Mirror of pptx-go's scene.Cycle. A deck with no Cycle is
// byte-identical (a new node, absent until used).
type Cycle struct {
	// Stages are the ring's nodes, placed evenly around the loop in order.
	Stages []CycleStage `json:"stages,omitempty"`
}

func (Cycle) slideNodeKind() Kind { return KindCycle }

// MarshalJSON injects the "cycle" kind discriminator via marshalNode.
func (c *Cycle) MarshalJSON() ([]byte, error) { return marshalNode(KindCycle, *c) }

func init() { registerNodeKind(KindCycle, func() SlideNode { return &Cycle{} }) }

// CycleStage is one node on a Cycle ring (D-128): a label + optional
// curated icon + a soul-driven series accent color.
type CycleStage struct {
	// Label is the stage's headline text.
	Label string `json:"label,omitempty"`
	// Icon is an optional curated icon name drawn inside the stage node.
	Icon string `json:"icon,omitempty"`
	// AccentIndex selects a soul-driven series accent color for the stage
	// node (0 = the first accent). A plain int passthrough — 0 is a real
	// value (the first accent), not "unset".
	AccentIndex int `json:"accentIndex,omitempty"`
}
