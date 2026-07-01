package contracts

// Quadrant is a 2x2 positioning map (R14.9, D-124): labeled X/Y axes with
// low/high end captions, optional per-quadrant tint + title, and items
// plotted at caller (x,y) coordinates in [0,1] (origin bottom-left). Axes,
// dividers, item dots, and labels draw as native shapes — no asset. Mirror
// of pptx-go's scene.Quadrant. Pure integer-EMU layout → byte-identical
// across renders/worker counts; a deck with no Quadrant is byte-identical
// (a new node, absent until used).
type Quadrant struct {
	// AxisX carries the horizontal axis's low/high end captions.
	AxisX QuadrantAxis `json:"axisX,omitempty"`
	// AxisY carries the vertical axis's low/high end captions.
	AxisY QuadrantAxis `json:"axisY,omitempty"`
	// Quadrants are optional per-cell tint + title, indexed 0=top-left,
	// 1=top-right, 2=bottom-left, 3=bottom-right. An empty Fill draws no
	// tint for that cell.
	Quadrants [4]QuadrantCell `json:"quadrants,omitempty"`
	// Items are plotted points; X/Y in [0,1] with the origin at the
	// bottom-left.
	Items []QuadrantItem `json:"items,omitempty"`
}

func (Quadrant) slideNodeKind() Kind { return KindQuadrant }

// MarshalJSON injects the "quadrant" kind discriminator via marshalNode.
func (q *Quadrant) MarshalJSON() ([]byte, error) { return marshalNode(KindQuadrant, *q) }

func init() { registerNodeKind(KindQuadrant, func() SlideNode { return &Quadrant{} }) }

// QuadrantAxis is one axis's end captions (D-124).
type QuadrantAxis struct {
	// LowLabel captions the axis's low (origin) end.
	LowLabel string `json:"lowLabel,omitempty"`
	// HighLabel captions the axis's high (far) end.
	HighLabel string `json:"highLabel,omitempty"`
}

// QuadrantCell is an optional per-quadrant tint + title (D-124). Mirror of
// pptx-go's scene.QuadrantCell, whose Fill is a *ColorRole (nil = no tint);
// the product uses a plain ColorRole string where "" = no tint, mapped via
// mapColorRolePtr at render time.
type QuadrantCell struct {
	// Title is the quadrant's caption, drawn inside the cell.
	Title string `json:"title,omitempty"`
	// Fill is the quadrant's tint color role; "" draws no tint.
	Fill ColorRole `json:"fill,omitempty"`
}

// QuadrantItem is a plotted point (D-124): X/Y in [0,1] (origin
// bottom-left), a Label, and an AccentIndex selecting the dot color from a
// pinned token cycle.
type QuadrantItem struct {
	// X is the item's horizontal position, 0..1 (0 = left, 1 = right).
	X float64 `json:"x"`
	// Y is the item's vertical position, 0..1 (0 = bottom, 1 = top).
	Y float64 `json:"y"`
	// Label is the item's caption drawn beside its plotted dot.
	Label string `json:"label,omitempty"`
	// AccentIndex selects a soul-driven series accent color for the dot
	// (0 = the first accent). A plain int passthrough — 0 is a real value
	// (the first accent), not "unset".
	AccentIndex int `json:"accentIndex,omitempty"`
}
