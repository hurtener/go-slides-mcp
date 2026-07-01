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
	// Style, when non-nil, turns the table into a comparison matrix — a
	// header band, zebra body striping, a highlighted column, an emphasized
	// row-label column, and grouped header spans — all resolved from soul
	// tokens (R14.3, D-118). nil (the default) keeps today's plain banded
	// table, byte-identical. Cell-value glyphs (check/cross/dot/mini-bar) are
	// deliberately NOT a Table feature — a native OOXML table cell holds only
	// text, so compose those with a Bento of Checklist/IconRows cells instead.
	Style *TableStyle `json:"style,omitempty"`
}

// TableStyle is the additive visual styling for a comparison-matrix Table
// (R14.3, D-118). Every field's zero value reproduces an unstyled column, so
// a caller turns features on one at a time. Colors resolve from theme
// tokens: the header band and the highlighted column use the accent color
// role; zebra striping and the row-label column use the surface-alt color
// role. Mirror of pptx-go's scene.TableStyle.
type TableStyle struct {
	// HeaderFill fills the header row with the accent band (contrast text).
	HeaderFill bool `json:"headerFill,omitempty"`
	// Zebra alternates a subtle surface-alt fill on odd body rows.
	Zebra bool `json:"zebra,omitempty"`
	// HighlightCol is the 1-based column to emphasize (accent tint + heavier
	// accent border) — e.g. a "recommended" plan column. 0 (the default) means
	// no column is highlighted.
	HighlightCol int `json:"highlightCol,omitempty"`
	// RowLabelCol emphasizes the first column as row labels (surface-alt fill
	// + bold).
	RowLabelCol bool `json:"rowLabelCol,omitempty"`
	// HeaderGroups, when non-empty, adds a grouped header row above the
	// headers: each group's Label spans Span columns (merged), laid
	// left-to-right from column 0. The spans should sum to the column count.
	HeaderGroups []HeaderGroup `json:"headerGroups,omitempty"`
}

// HeaderGroup is one merged span in a Table's grouped header row (R14.3,
// D-118). Mirror of pptx-go's scene.HeaderGroup.
type HeaderGroup struct {
	// Label is the group heading (e.g. "Enterprise").
	Label string `json:"label,omitempty"`
	// Span is the number of columns the group covers (>= 1).
	Span int `json:"span,omitempty"`
}

func (Table) slideNodeKind() Kind { return KindTable }

// MarshalJSON injects the "table" kind discriminator via marshalNode.
// Headers/Rows marshal through each TextRun's own MarshalJSON.
func (t *Table) MarshalJSON() ([]byte, error) { return marshalNode(KindTable, *t) }

func init() { registerNodeKind(KindTable, func() SlideNode { return &Table{} }) }
