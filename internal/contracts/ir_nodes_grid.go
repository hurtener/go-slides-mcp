package contracts

import "encoding/json"

// GridConnector draws a connector glyph in the gutter between two adjacent
// columns of a Grid (R12.4, D-099), so an architecture / pipeline grid reads
// as data flow, not just adjacency. Between holds the two adjacent column
// indices ({c, c+1}); Kind reuses the Flow connector set (plus ConnectorBiArrow);
// Label is an optional caption in the gutter.
type GridConnector struct {
	// Between holds the two adjacent column indices, e.g. {0,1}.
	Between [2]int `json:"between,omitempty"`
	// Kind selects the gutter glyph (reuse Flow's connector vocabulary incl. bi_arrow).
	Kind ConnectorKind `json:"kind,omitempty"`
	// Label is an optional gutter caption; "" = none.
	Label string `json:"label,omitempty"`
}

// Grid lays out children in a column grid. Columns is 2..4; Cells length is a
// multiple of Columns; Ratio is empty or len == Columns (validation, later
// unit). Mirror of scene.Grid. Children nest recursively.
type Grid struct {
	// Columns is the cell count per row, 2..4.
	Columns int `json:"columns,omitempty"`
	// Ratio is an optional per-column width ratio; empty or len == Columns.
	Ratio []int `json:"ratio,omitempty"`
	// Gap is the spacing between cells.
	Gap SpaceRole `json:"gap,omitempty"`
	// Cells is the grid children, a multiple of Columns.
	Cells []SlideNode `json:"cells,omitempty"`
	// Connectors are optional inter-column gutter glyphs; empty = none,
	// byte-identical to a pre-R12.4 Grid.
	Connectors []GridConnector `json:"connectors,omitempty"`
}

func (Grid) slideNodeKind() Kind { return KindGrid }

// MarshalJSON injects the "grid" kind; child slices marshal through each
// child's own MarshalJSON (kind injected per child).
func (g *Grid) MarshalJSON() ([]byte, error) { return marshalNode(KindGrid, *g) }

// UnmarshalJSON dispatches Cells through UnmarshalSlideNode so the container
// nests recursively (CONVENTIONS §3).
func (g *Grid) UnmarshalJSON(data []byte) error {
	type raw struct {
		Columns    int               `json:"columns,omitempty"`
		Ratio      []int             `json:"ratio,omitempty"`
		Gap        SpaceRole         `json:"gap,omitempty"`
		Cells      []json.RawMessage `json:"cells,omitempty"`
		Connectors []GridConnector   `json:"connectors,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	g.Columns = r.Columns
	g.Ratio = r.Ratio
	g.Gap = r.Gap
	g.Connectors = r.Connectors
	cells, err := unmarshalNodes(r.Cells)
	if err != nil {
		return err
	}
	g.Cells = cells
	return nil
}

func init() { registerNodeKind(KindGrid, func() SlideNode { return &Grid{} }) }
