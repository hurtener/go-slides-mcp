package contracts

// Kind is the JSON "kind" discriminator that selects a concrete SlideNode.
// Values are snake_case and match the pptx-go scene node (CONVENTIONS §2).
type Kind string

// Node kinds implemented in this unit (Phase 1A). Later units register the
// remaining leaf kinds (divider, quote, chip, arrow, section_divider, table,
// flow, image, code_block, chart, decoration).
const (
	KindHero        Kind = "hero"
	KindHeading     Kind = "heading"
	KindProse       Kind = "prose"
	KindList        Kind = "list"
	KindCallout     Kind = "callout"
	KindTwoColumn   Kind = "two_column"
	KindGrid        Kind = "grid"
	KindCard        Kind = "card"
	KindCardSection Kind = "card_section"
)

// LayoutKind is a slide's structural intent, mapping to a master layout
// (mirrors pptx-go's scene.LayoutKind; CONVENTIONS §2).
type LayoutKind string

// Slide layouts (wire values per CONVENTIONS §2).
const (
	LayoutCover        LayoutKind = "cover"
	LayoutTitleContent LayoutKind = "title_content"
	LayoutTwoColumn    LayoutKind = "two_column"
	LayoutCardGrid     LayoutKind = "card_grid"
	LayoutFullBleed    LayoutKind = "full_bleed"
	LayoutBlank        LayoutKind = "blank"
)
