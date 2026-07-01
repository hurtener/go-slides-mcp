package contracts

// Timeline is a roadmap / horizontal-axis node: a single-lane or swimlaned
// series of Milestones, optionally overlaid with phase/horizon Bands drawn
// behind the axis. Renders as native PPTX shapes (axis, marker dots,
// custGeom icons) — no asset. Mirror of pptx-go's scene.Timeline (D-119).
type Timeline struct {
	// Milestones is the single-lane milestone list, used when Lanes is
	// empty. Ignored (superseded) when Lanes is non-empty.
	Milestones []Milestone `json:"milestones,omitempty"`
	// Lanes are swimlanes (rows), each with its own milestones. When
	// non-empty, Lanes supersedes Milestones.
	Lanes []TimelineLane `json:"lanes,omitempty"`
	// Bands are optional phase/horizon regions drawn behind the axis, each
	// spanning [From,To] of the timeline width.
	Bands []TimelineBand `json:"bands,omitempty"`
}

func (Timeline) slideNodeKind() Kind { return KindTimeline }

// MarshalJSON injects the "timeline" kind discriminator via marshalNode.
func (t *Timeline) MarshalJSON() ([]byte, error) { return marshalNode(KindTimeline, *t) }

func init() { registerNodeKind(KindTimeline, func() SlideNode { return &Timeline{} }) }

// Milestone is one point on a Timeline axis. Mirror of pptx-go's
// scene.Milestone.
type Milestone struct {
	// Position is the milestone's location along the axis, 0..1 (0 = start,
	// 1 = end).
	Position float64 `json:"position"`
	// Label is the milestone's headline text.
	Label string `json:"label,omitempty"`
	// Detail is optional supporting text under the label.
	Detail string `json:"detail,omitempty"`
	// Icon is an optional curated icon name drawn at the marker.
	Icon string `json:"icon,omitempty"`
	// AccentIndex selects a soul-driven series accent color for the marker
	// (0 = the first accent). A plain int passthrough — 0 is a real value
	// (the first accent), not "unset".
	AccentIndex int `json:"accentIndex,omitempty"`
}

// TimelineLane is one swimlane (row) of a Timeline: a left-gutter Label and
// its own Milestones. Mirror of pptx-go's scene.TimelineLane.
type TimelineLane struct {
	// Label is the swimlane's left-gutter caption.
	Label string `json:"label,omitempty"`
	// Milestones are this lane's milestones.
	Milestones []Milestone `json:"milestones,omitempty"`
}

// TimelineBand is a phase/horizon region drawn behind a Timeline axis: it
// spans [From,To] (each 0..1) of the timeline width, filled with Fill (a
// low-alpha surface role) and labeled at the top. Mirror of pptx-go's
// scene.TimelineBand.
type TimelineBand struct {
	// From is the band's start position along the axis, 0..1.
	From float64 `json:"from"`
	// To is the band's end position along the axis, 0..1. Must be >= From.
	To float64 `json:"to"`
	// Label is the band's caption drawn at its top.
	Label string `json:"label,omitempty"`
	// Fill is the band's surface color role (low-alpha).
	Fill ColorRole `json:"fill,omitempty"`
}
