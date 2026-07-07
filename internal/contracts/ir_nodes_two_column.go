package contracts

import "encoding/json"

// ColumnRatio names a left:right width split (mirrors pptx-go's
// scene.ColumnRatio; CONVENTIONS §2 uses "1:1").
type ColumnRatio string

// Column ratios.
const (
	Ratio11 ColumnRatio = "1:1"
	Ratio12 ColumnRatio = "1:2"
	Ratio21 ColumnRatio = "2:1"
)

// IsValid reports whether v is one of the closed ColumnRatio wire values
// (Phase 12 A4).
func (v ColumnRatio) IsValid() bool { return IsValidEnum(v, AllowedColumnRatio()) }

// ColumnJoin selects the optional element drawn centered on the column seam
// (mirrors pptx-go's scene.ColumnJoin; D-055). JoinNone (zero value / empty)
// draws nothing; an existing TwoColumn with no join field renders byte-for-byte
// unchanged.
type ColumnJoin string

// Column join styles (wire values per compose-a-scene).
const (
	// JoinNone (default) draws nothing on the column seam.
	JoinNone ColumnJoin = ""
	// JoinBadge draws a circular text badge (JoinLabel) centered on the seam.
	JoinBadge ColumnJoin = "badge"
	// JoinArrow draws a right-arrow connector between the columns.
	JoinArrow ColumnJoin = "arrow"
)

// IsValid reports whether v is one of the closed ColumnJoin wire values.
// The empty string (JoinNone) is also valid.
func (v ColumnJoin) IsValid() bool { return IsValidEnum(v, AllowedColumnJoin()) }

// JoinPosition selects where a TwoColumn's Join element sits (mirrors
// pptx-go's scene.JoinPosition, an int enum). The zero value JoinSeam
// centers it on the vertical seam between the columns; JoinTopBridge /
// JoinBottomBridge draw a horizontal accent bracket spanning both columns'
// combined width at the top / bottom edge, with the JoinLabel as a
// centered pill on it. (R12.8, D-101.)
type JoinPosition string

// Join positions (wire values per compose-a-scene).
const (
	// JoinSeam centers the Join element on the vertical seam (zero value).
	JoinSeam JoinPosition = "seam"
	// JoinTopBridge spans a bracket across the top of the two columns.
	JoinTopBridge JoinPosition = "top_bridge"
	// JoinBottomBridge spans a bracket across the bottom of the two columns.
	JoinBottomBridge JoinPosition = "bottom_bridge"
)

// TwoColumn splits a slide into left and right child columns. Both sides
// must be non-empty (validation, later unit). Mirror of scene.TwoColumn.
// Children are []SlideNode and nest recursively.
//
// Join and JoinLabel are additive (D-055): their zero values draw no element
// on the seam, so an existing TwoColumn renders byte-for-byte unchanged.
type TwoColumn struct {
	// Ratio is the left:right width split.
	Ratio ColumnRatio `json:"ratio,omitempty"`
	// Left is the left-column children.
	Left []SlideNode `json:"left,omitempty"`
	// Right is the right-column children.
	Right []SlideNode `json:"right,omitempty"`
	// Join is the optional element drawn centered on the column seam.
	// JoinNone (empty / omitted) draws nothing (default).
	Join ColumnJoin `json:"join,omitempty"`
	// JoinLabel is the badge text when Join == "badge" (e.g. "VS").
	// Ignored for other Join values.
	JoinLabel string `json:"joinLabel,omitempty"`
	// JoinPosition selects where the Join element sits; empty = JoinSeam.
	JoinPosition JoinPosition `json:"joinPosition,omitempty"`
}

func (TwoColumn) slideNodeKind() Kind { return KindTwoColumn }

// MarshalJSON injects the "two_column" kind; child slices marshal through
// each child's own MarshalJSON (kind injected per child).
func (t *TwoColumn) MarshalJSON() ([]byte, error) { return marshalNode(KindTwoColumn, *t) }

// UnmarshalJSON dispatches Left and Right children through UnmarshalSlideNode
// so the container nests recursively (CONVENTIONS §3).
func (t *TwoColumn) UnmarshalJSON(data []byte) error {
	type raw struct {
		Ratio        ColumnRatio       `json:"ratio,omitempty"`
		Left         []json.RawMessage `json:"left,omitempty"`
		Right        []json.RawMessage `json:"right,omitempty"`
		Join         ColumnJoin        `json:"join,omitempty"`
		JoinLabel    string            `json:"joinLabel,omitempty"`
		JoinPosition JoinPosition      `json:"joinPosition,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	t.Ratio = r.Ratio
	t.Join = r.Join
	t.JoinLabel = r.JoinLabel
	t.JoinPosition = r.JoinPosition
	left, err := unmarshalNodes(r.Left)
	if err != nil {
		return err
	}
	t.Left = left
	right, err := unmarshalNodes(r.Right)
	if err != nil {
		return err
	}
	t.Right = right
	return nil
}

func init() { registerNodeKind(KindTwoColumn, func() SlideNode { return &TwoColumn{} }) }
