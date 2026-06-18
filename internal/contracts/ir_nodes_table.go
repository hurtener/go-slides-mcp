package contracts

// Table is a grid of cells with a header row and a caption. Renders as native
// PPTX shapes. Mirror of pptx-go's scene.Table. Each header cell and each
// body cell is a RichText; every row width MUST equal the header width
// (validation, later unit).
type Table struct {
	// Headers is the header row, one RichText per column.
	Headers []RichText `json:"headers,omitempty"`
	// Rows is the body rows; each row is one RichText per column.
	Rows [][]RichText `json:"rows,omitempty"`
	// Caption is the optional table caption.
	Caption string `json:"caption,omitempty"`
}

func (Table) slideNodeKind() Kind { return KindTable }

// MarshalJSON injects the "table" kind discriminator via marshalNode.
// Headers/Rows marshal through each TextRun's own MarshalJSON.
func (t *Table) MarshalJSON() ([]byte, error) { return marshalNode(KindTable, *t) }

func init() { registerNodeKind(KindTable, func() SlideNode { return &Table{} }) }
