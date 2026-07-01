package contracts

// DeltaTone selects the color direction of a Stat's delta (mirrors
// pptx-go's scene.DeltaTone; D-057). The zero value "neutral" is muted,
// so a delta with no tone set reads as neutral.
type DeltaTone string

// Delta tone wire values (D-057).
const (
	DeltaNeutral DeltaTone = "neutral" // muted — the zero / omitted value
	DeltaUp      DeltaTone = "up"      // positive (success color)
	DeltaDown    DeltaTone = "down"    // negative (error color)
)

// IsValid reports whether v is one of the closed DeltaTone wire values.
// The validator additionally accepts "" (empty / omitted) as neutral.
func (v DeltaTone) IsValid() bool { return IsValidEnum(v, AllowedDeltaTone()) }

// Stat is a hero big-number metric: a display-scale Value with a Label
// and an optional directional Delta (e.g. "$2,200" / "ARR" / "+18%"). A
// row of Stats inside a Grid forms a metric or pricing strip. Mirror of
// pptx-go's scene.Stat. By default the engine renders Value/Delta
// verbatim — it formats no numbers (D-026) — unless the typed Number
// path below is set (R14.13, D-121).
type Stat struct {
	// Value is the display-scale metric (e.g. "$2,200", "98%"). The engine
	// renders it verbatim at display type scale — format it before passing.
	Value string `json:"value,omitempty"`
	// Label is the supporting caption below the value (e.g. "ARR", "per month").
	Label string `json:"label,omitempty"`
	// Delta is the optional change indicator (e.g. "+18%", "-3 pp"). Omit
	// or leave empty to render no delta line.
	Delta string `json:"delta,omitempty"`
	// DeltaTone colors the delta: "neutral" (muted, default) | "up" (success
	// color) | "down" (error color). Ignored when Delta is empty.
	DeltaTone DeltaTone `json:"deltaTone,omitempty"`
	// AutoFit shrinks the Value (the big-number display run) to fit its box
	// instead of clipping/overflowing when a long number/price would
	// otherwise overflow (shrink-to-fit). Default false = the Value renders
	// at its full type size.
	AutoFit bool `json:"autoFit,omitempty"`
	// Number is an optional numeric value (R14.13, D-121). When set, it is
	// formatted via Format (or a plain integer-ish default when Format is
	// nil/omitted) and SUPERSEDES the raw Value string, then shrink-to-fits
	// via AutoFit so a long formatted price/number stays on one line. Nil
	// (omitted) renders the raw Value verbatim — byte-identical to a Stat
	// that predates this field. A pointer because 0 is a real value (the
	// D-054 nil-means-unset pattern).
	//
	// Note: this is a per-Stat mechanism only. A soul-level default number
	// format (e.g. "this brand's numbers are € by default") is a deferred
	// follow-up, not built here.
	Number *float64 `json:"number,omitempty"`
	// Format is an optional number/currency/percent/locale format applied
	// to Number (R14.13, D-121). Nil with a non-nil Number uses a plain
	// default (no grouping/decimals/affixes, i.e. the zero NumberFormat).
	// Ignored when Number is nil.
	Format *NumberFormat `json:"format,omitempty"`
}

func (Stat) slideNodeKind() Kind { return KindStat }

// MarshalJSON injects the "stat" kind discriminator via marshalNode.
func (s *Stat) MarshalJSON() ([]byte, error) { return marshalNode(KindStat, *s) }

func init() { registerNodeKind(KindStat, func() SlideNode { return &Stat{} }) }
