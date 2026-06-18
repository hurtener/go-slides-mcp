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

// TwoColumn splits a slide into left and right child columns. Both sides
// must be non-empty (validation, later unit). Mirror of scene.TwoColumn.
// Children are []SlideNode and nest recursively.
type TwoColumn struct {
	// Ratio is the left:right width split.
	Ratio ColumnRatio `json:"ratio,omitempty"`
	// Left is the left-column children.
	Left []SlideNode `json:"left,omitempty"`
	// Right is the right-column children.
	Right []SlideNode `json:"right,omitempty"`
}

func (TwoColumn) slideNodeKind() Kind { return KindTwoColumn }

// MarshalJSON injects the "two_column" kind; child slices marshal through
// each child's own MarshalJSON (kind injected per child).
func (t *TwoColumn) MarshalJSON() ([]byte, error) { return marshalNode(KindTwoColumn, *t) }

// UnmarshalJSON dispatches Left and Right children through UnmarshalSlideNode
// so the container nests recursively (CONVENTIONS §3).
func (t *TwoColumn) UnmarshalJSON(data []byte) error {
	type raw struct {
		Ratio ColumnRatio       `json:"ratio,omitempty"`
		Left  []json.RawMessage `json:"left,omitempty"`
		Right []json.RawMessage `json:"right,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	t.Ratio = r.Ratio
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
