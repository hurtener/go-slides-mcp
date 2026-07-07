package contracts

// CheckState selects a Checklist item's status glyph (mirrors pptx-go's
// scene.CheckState, an int enum — the product mirror is a string enum).
// The zero value CheckDone is a filled affirmative check (the common "you
// get this" row), so a checklist with no per-item State field defaults to
// the engine's accent-tinted check (R12.2, D-095).
type CheckState string

// Check states (wire values per compose-a-scene).
const (
	// CheckDone is a filled check glyph (default), accent-tinted.
	CheckDone CheckState = "done"
	// CheckNo is a filled cross glyph, muted.
	CheckNo CheckState = "no"
	// CheckNeutral is a filled dot glyph, muted.
	CheckNeutral CheckState = "neutral"
)

// ChecklistItem is one row of a Checklist (R12.2, D-095): rich text, a
// status, and an optional icon name that overrides the state's default
// glyph (a closed-name curated/extension icon). The zero value of State
// (CheckDone) renders a filled check glyph before the text.
type ChecklistItem struct {
	// Text is the row's rich text.
	Text RichText `json:"text,omitempty"`
	// State selects the row's status glyph; empty = CheckDone.
	State CheckState `json:"state,omitempty"`
	// Icon is an optional glyph override (curated/extension icon name); "" = state's default glyph.
	Icon string `json:"icon,omitempty"`
}

// Checklist is a dense feature / "what you get" list (R12.2, D-095): rows
// of a filled status glyph (check / cross / dot) before rich text, reflowed
// row-major into 1-3 balanced columns, with the text hanging-indented from
// the glyph width. The glyph is a true filled custGeom (the curated
// check / x / dot icon), never an empty font checkbox. Mirror of
// pptx-go's scene.Checklist.
//
// GlyphTone overrides the per-state glyph color for every item; an empty
// string leaves the per-state default in effect (CheckDone → accent,
// others → muted), mirroring the D-054 pointer/sentinel pattern at the
// product layer (the engine uses *ColorRole; nil there = default; we map
// "" → nil via the existing mapColorRolePtr helper). Fill distributes
// inter-row slack so a short list spans the box height. Additive: a deck
// with no Checklist is byte-identical.
type Checklist struct {
	// Items is the sequence of checklist rows; at least one is required.
	Items []ChecklistItem `json:"items,omitempty"`
	// Columns selects 1..3 column reflow (row-major); 0 = 1 column (default).
	Columns int `json:"columns,omitempty"`
	// GlyphTone overrides the per-state glyph color role; "" = per-state default.
	GlyphTone ColorRole `json:"glyphTone,omitempty"`
	// Fill distributes rows to fill the box height (like VAlignFill); false = top-aligned.
	Fill bool `json:"fill,omitempty"`
}

func (Checklist) slideNodeKind() Kind { return KindChecklist }

// MarshalJSON injects the "checklist" kind discriminator via marshalNode.
// ChecklistItem is a plain concrete sub-struct (no nested SlideNode), so
// the default strictUnmarshal path decodes Items directly — no custom
// UnmarshalJSON (the datamark leaf pattern).
func (c *Checklist) MarshalJSON() ([]byte, error) { return marshalNode(KindChecklist, *c) }

func init() { registerNodeKind(KindChecklist, func() SlideNode { return &Checklist{} }) }
