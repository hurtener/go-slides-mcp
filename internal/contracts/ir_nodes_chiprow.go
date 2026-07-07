package contracts

// ChipSpec is one chip in a ChipRow (R12.5, D-096): a label, a tone, the
// tone's color role, and an optional leading icon (a closed-name curated or
// extension icon, Stage-1 validated). It mirrors the single Chip node's
// vocabulary (ChipTone / ColorRole). For ChipTint the Color is ignored (the
// chip uses ColorSurfaceAlt); ChipSolid / ChipOutline use Color.
type ChipSpec struct {
	// Label is the chip text.
	Label string `json:"label,omitempty"`
	// Tone selects the chip fill treatment; empty = ChipTint (tinted pill).
	Tone ChipTone `json:"tone,omitempty"`
	// Color is the chip color role; ignored for ChipTint (uses ColorSurfaceAlt).
	Color ColorRole `json:"color,omitempty"`
	// Icon is an optional leading glyph (curated/extension icon name); "" = none.
	Icon string `json:"icon,omitempty"`
}

// ChipRow is a horizontal, wrap-to-next-line row of content-fit chip pills
// with an optional leading label (R12.5, D-096): a tag / category / capability
// strip. Each chip sizes to its label (plus an optional leading icon); chips
// lay left-to-right and, when Wrap is set, reflow onto new lines when the row
// width is exceeded. Mirror of pptx-go's scene.ChipRow.
//
// Wrap is the engine mechanism: the zero value lays all chips on a single
// line (the minimal behavior); a product that wants a reflowing strip sets
// Wrap true (D-026). A non-empty Label renders as a leading TypeCaption label
// before the first chip. Align offsets each line's chips (zero = inherit the
// slide's Content.Horizontal). Additive: a deck with no ChipRow is
// byte-identical.
type ChipRow struct {
	// Label is an optional leading TypeCaption label before the first chip; "" = none.
	Label string `json:"label,omitempty"`
	// Chips is the sequence of chip pills; at least one is required.
	Chips []ChipSpec `json:"chips,omitempty"`
	// Wrap reflows chips onto new lines when the row width is exceeded; false = single line.
	Wrap bool `json:"wrap,omitempty"`
	// Align is a per-node horizontal alignment override; 0 = inherit slide.
	Align HAlign `json:"align,omitempty"`
}

func (ChipRow) slideNodeKind() Kind { return KindChipRow }

// MarshalJSON injects the "chip_row" kind discriminator via marshalNode.
// ChipSpec is a plain concrete sub-struct (no nested SlideNode), so the
// default strictUnmarshal path decodes Chips directly — no custom
// UnmarshalJSON (CONVENTIONS §3, the datamark leaf pattern).
func (c *ChipRow) MarshalJSON() ([]byte, error) { return marshalNode(KindChipRow, *c) }

func init() { registerNodeKind(KindChipRow, func() SlideNode { return &ChipRow{} }) }
