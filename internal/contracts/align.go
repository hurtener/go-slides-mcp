package contracts

// VAlign selects vertical alignment of the slide's body stack within the body
// region. Wire value is a JSON string; the zero value (empty string) is treated
// as "top" by the renderer — backward-compatible with pre-alignment slides.
type VAlign string

// Vertical alignment wire values. Empty string == "top" (the default).
const (
	// VAlignTop (default) starts the body stack at the top of the body region.
	// This is the backward-compatible default when the field is absent.
	VAlignTop VAlign = "top"
	// VAlignCenter distributes the remaining vertical space equally above and
	// below the body stack; the stack never starts above the top edge.
	VAlignCenter VAlign = "center"
	// VAlignBottom places the body stack flush with the body region's bottom
	// edge; the stack never starts above the top edge.
	VAlignBottom VAlign = "bottom"
	// VAlignJustify distributes vertical slack evenly into the inter-node gaps.
	// Equivalent to VAlignTop for a single node or when there is no slack.
	VAlignJustify VAlign = "justify"
	// VAlignFill pins fixed leaf nodes (heading, prose, list, callout, divider,
	// etc.) at the top (like "top") and grows the flexible container nodes
	// (grid, two_column, card, card_section, table, chart, image) to consume
	// the remaining body height, so a heading-plus-content slide fills its
	// frame instead of reading thin at the bottom. When no flexible node is
	// present, or when content already overflows the body height, it behaves
	// exactly like "top". Beats "center"/"justify" for heading-plus-content
	// slides where the content block should dominate the frame.
	VAlignFill VAlign = "fill"
)

// HAlign selects horizontal alignment of leaf nodes within the body region.
// Wire value is a JSON string; the zero value (empty string) is treated as
// "left" by the renderer — backward-compatible with pre-alignment slides.
// Containers (grid, two_column, table, card, card_section, flow, callout,
// image, chart, code_block, divider, arrow) always use left alignment
// regardless of this field and are always full-width.
type HAlign string

// Horizontal alignment wire values. Empty string == "left" (the default).
const (
	// HAlignLeft (default) spans each leaf node across the full body width,
	// left-flush. This is the backward-compatible default when the field is absent.
	HAlignLeft HAlign = "left"
	// HAlignCenter narrows each affected leaf node to its estimated natural text
	// width and centers it within the body region.
	HAlignCenter HAlign = "center"
	// HAlignRight narrows each affected leaf node to its estimated natural text
	// width and places it flush with the body right edge.
	HAlignRight HAlign = "right"
)

// Alignment sets how the slide's body content sits within the body frame.
// Both axes are optional; omitting a field (or using an empty string) defaults
// to top/left, reproducing the pre-alignment layout unchanged.
type Alignment struct {
	// Vertical sets the body stack's vertical position within the body region:
	// "top" (default), "center", "bottom", "justify", or "fill". Empty = top.
	// Use "fill" to grow container nodes (grid, card, image, chart, etc.) to
	// consume the remaining body height — ideal for heading-plus-content slides.
	Vertical VAlign `json:"vertical,omitempty"`
	// Horizontal sets the default horizontal alignment for leaf nodes in the
	// body stack: "left" (default), "center", or "right". Empty = left.
	// Per-node align fields (on hero, heading, prose, quote, chip,
	// section_divider) override this for individual blocks.
	Horizontal HAlign `json:"horizontal,omitempty"`
}
