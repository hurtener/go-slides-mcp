package contracts

// DataMarkKind selects a DataMark's render shape (mirrors pptx-go's
// scene.DataMarkKind, an int enum — the product mirror is a string enum per
// the D-054-adjacent "string enum, not int" convention).
type DataMarkKind string

// DataMark kinds (wire values per compose-a-scene). The empty string is
// accepted (acceptEmpty) and maps to DataMarkBar at render time, mirroring
// the engine's zero value.
const (
	// DataMarkBar is a single progress/capacity bar: a track + a fill to Value.
	DataMarkBar DataMarkKind = "bar"
	// DataMarkBars is a small bar group, one bar per Values entry.
	DataMarkBars DataMarkKind = "bars"
	// DataMarkSparkline is a polyline through Values (a trend line).
	DataMarkSparkline DataMarkKind = "sparkline"
	// DataMarkDonut is a single-value ring (Value 0..1) with a centered label.
	DataMarkDonut DataMarkKind = "donut"
	// DataMarkGauge is a single-value speedometer arc (Value 0..1) with a label.
	DataMarkGauge DataMarkKind = "gauge"
)

// DataMark is a native (no-raster) micro-chart (R14.8, D-122): a crisp,
// brand-colored vector data mark drawn entirely from preset shapes — a
// progress bar, a small bar group, a sparkline, a donut, or a gauge. It is
// driven by numeric values in [0,1] and theme colors, sizes to its box, and
// embeds in a Card/Bento cell. Mirror of pptx-go's scene.DataMark. Pure
// integer-EMU geometry → byte-identical across renders/worker counts; no
// AssetResolver. The JSON field for the variant is "markKind" (not "kind",
// which is the node discriminator — CONVENTIONS §2, same pattern as
// Decoration's "decorationKind").
type DataMark struct {
	// Kind selects the mark shape (bar/bars/sparkline/donut/gauge). Empty
	// defaults to "bar" at render time.
	Kind DataMarkKind `json:"markKind,omitempty"`
	// Value is the single fraction in [0,1] for bar/donut/gauge.
	Value float64 `json:"value,omitempty"`
	// Values are the per-element fractions in [0,1] for bars/sparkline.
	Values []float64 `json:"values,omitempty"`
	// Orientation selects a horizontal (default) or vertical bar. Ignored by
	// bars/sparkline/donut/gauge.
	Orientation FlowOrientation `json:"orientation,omitempty"`
	// Color overrides the mark's color role; empty defaults to the accent
	// (the track is always the surface-alt role).
	Color ColorRole `json:"color,omitempty"`
	// Label is an optional inline caption (drawn to the right of a
	// horizontal bar, or centered for donut/gauge).
	Label string `json:"label,omitempty"`
}

func (DataMark) slideNodeKind() Kind { return KindDataMark }

// MarshalJSON injects the "data_mark" kind discriminator via marshalNode.
func (d *DataMark) MarshalJSON() ([]byte, error) { return marshalNode(KindDataMark, *d) }

func init() { registerNodeKind(KindDataMark, func() SlideNode { return &DataMark{} }) }
