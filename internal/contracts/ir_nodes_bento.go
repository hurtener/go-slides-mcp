package contracts

import "encoding/json"

// BentoCell is one cell of a BentoRow: the child node and how many of the
// bento's shared column units this cell spans (>= 1; defaults to 1 when zero).
// Mirror of pptx-go's scene.BentoCell (D-056).
type BentoCell struct {
	// Span is the number of shared column units this cell occupies (>= 1).
	Span int `json:"span,omitempty"`
	// Node is the child slide node rendered in this cell.
	Node SlideNode `json:"node,omitempty"`
}

// bentoRawCell is the wire form used by BentoCell's Unmarshal, where Node is
// an opaque JSON object dispatched through UnmarshalSlideNode.
type bentoRawCell struct {
	Span int             `json:"span,omitempty"`
	Node json.RawMessage `json:"node,omitempty"`
}

// MarshalJSON encodes a BentoCell, injecting the child node's own kind via its
// MarshalJSON method. wire is the same shape as BentoCell but is a distinct
// type so json.Marshal encodes it field-by-field without re-entering this method.
func (c BentoCell) MarshalJSON() ([]byte, error) {
	type wire BentoCell
	return json.Marshal(wire(c))
}

// UnmarshalJSON decodes a BentoCell, dispatching the child node through
// UnmarshalSlideNode.
func (c *BentoCell) UnmarshalJSON(data []byte) error {
	var r bentoRawCell
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	c.Span = r.Span
	if r.Node != nil {
		n, err := UnmarshalSlideNode(r.Node)
		if err != nil {
			return err
		}
		c.Node = n
	}
	return nil
}

// BentoRow is one row of a Bento: an optional left-gutter label and a
// left-to-right sequence of span-weighted cells. Mirror of pptx-go's
// scene.BentoRow (D-056).
type BentoRow struct {
	// Label is the optional left-gutter row label. Empty means no label for
	// this row; if any row has a label a gutter column is reserved for all rows.
	Label string `json:"label,omitempty"`
	// Cells is the left-to-right sequence of cells in this row.
	Cells []BentoCell `json:"cells,omitempty"`
}

// Bento is a row-labeled grid (D-056): rows that each carry an optional left
// label and cells of variable column span, measured against Columns shared
// column units (a span-S cell occupies S units, so columns align across rows).
// A row's cell spans must sum to <= Columns. Bento is a container node — its
// cells render per their own policy — and is distinct from Grid (uniform
// columns, one child per cell). Mirror of pptx-go's scene.Bento.
type Bento struct {
	// Columns is the shared column-unit count all rows are measured against (>= 1).
	Columns int `json:"columns,omitempty"`
	// Rows is the ordered sequence of bento rows.
	Rows []BentoRow `json:"rows,omitempty"`
}

func (Bento) slideNodeKind() Kind { return KindBento }

// MarshalJSON injects the "bento" kind discriminator via marshalNode. BentoCell
// child nodes are encoded through each node's own MarshalJSON.
func (b *Bento) MarshalJSON() ([]byte, error) { return marshalNode(KindBento, *b) }

// UnmarshalJSON strict-decodes a Bento and dispatches each BentoCell's child
// node through UnmarshalSlideNode, so nesting is recursive (CONVENTIONS §3).
func (b *Bento) UnmarshalJSON(data []byte) error {
	type rowWire struct {
		Label string            `json:"label,omitempty"`
		Cells []json.RawMessage `json:"cells,omitempty"`
	}
	type wire struct {
		Columns int               `json:"columns,omitempty"`
		Rows    []json.RawMessage `json:"rows,omitempty"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	b.Columns = w.Columns
	if w.Rows == nil {
		return nil
	}
	b.Rows = make([]BentoRow, len(w.Rows))
	for ri, rawRow := range w.Rows {
		var rw rowWire
		if err := json.Unmarshal(rawRow, &rw); err != nil {
			return err
		}
		b.Rows[ri].Label = rw.Label
		if rw.Cells != nil {
			b.Rows[ri].Cells = make([]BentoCell, len(rw.Cells))
			for ci, rawCell := range rw.Cells {
				if err := json.Unmarshal(rawCell, &b.Rows[ri].Cells[ci]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func init() { registerNodeKind(KindBento, func() SlideNode { return &Bento{} }) }
