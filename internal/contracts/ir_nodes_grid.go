package contracts

import "encoding/json"

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
}

func (Grid) slideNodeKind() Kind { return KindGrid }

// MarshalJSON injects the "grid" kind; child slices marshal through each
// child's own MarshalJSON (kind injected per child).
func (g *Grid) MarshalJSON() ([]byte, error) { return marshalNode(KindGrid, *g) }

// UnmarshalJSON dispatches Cells through UnmarshalSlideNode so the container
// nests recursively (CONVENTIONS §3).
func (g *Grid) UnmarshalJSON(data []byte) error {
	type raw struct {
		Columns int               `json:"columns,omitempty"`
		Ratio   []int             `json:"ratio,omitempty"`
		Gap     SpaceRole         `json:"gap,omitempty"`
		Cells   []json.RawMessage `json:"cells,omitempty"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	g.Columns = r.Columns
	g.Ratio = r.Ratio
	g.Gap = r.Gap
	cells, err := unmarshalNodes(r.Cells)
	if err != nil {
		return err
	}
	g.Cells = cells
	return nil
}

func init() { registerNodeKind(KindGrid, func() SlideNode { return &Grid{} }) }
