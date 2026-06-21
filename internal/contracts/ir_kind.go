package contracts

// Kind is the JSON "kind" discriminator that selects a concrete SlideNode.
// Values are snake_case and match the pptx-go scene node (CONVENTIONS §2).
type Kind string

// Node kinds implemented in Phase 1A.
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

// Node kinds implemented in Phase 1B — the remaining leaf kinds (mirror the
// compose-a-scene catalog). Wire values are snake_case per CONVENTIONS §2.
const (
	KindDivider        Kind = "divider"
	KindQuote          Kind = "quote"
	KindChip           Kind = "chip"
	KindArrow          Kind = "arrow"
	KindSectionDivider Kind = "section_divider"
	KindTable          Kind = "table"
	KindFlow           Kind = "flow"
	KindImage          Kind = "image"
	KindCodeBlock      Kind = "code_block"
	KindChart          Kind = "chart"
	KindDecoration     Kind = "decoration"
)

// Node kind added in R6 — the Stat leaf node (D-057).
const (
	KindStat Kind = "stat"
)

// Node kinds added in R5 — the Bento grid (D-056).
const (
	KindBento Kind = "bento"
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

// IsValid reports whether v is one of the closed LayoutKind wire values
// (Phase 12 A4).
func (v LayoutKind) IsValid() bool { return IsValidEnum(v, AllowedLayoutKind()) }
