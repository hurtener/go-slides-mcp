// Package autofit applies opt-in density transforms to a contracts.SlideDoc
// before export (R10.11). The first rung of the ladder is "fill": growing
// container-bearing slides to consume leftover body slack so an
// agent-authored deck doesn't read top-heavy. Transforms here are pure IR
// rewrites — no engine/layout math — so they are safe to run unconditionally
// behind the opt-in flag and never touch render code.
package autofit

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// fillableKinds is the closed set of top-level node kinds Fill treats as
// "flexible containers" worth growing. VAlignFill (R2, engine-done) only
// grows multi-child container nodes — grid, two_column, bento, table, and
// card_section — never bare leaves (hero/heading/prose/list/...) or a lone
// card/stat, so a sparse single-block slide is never ballooned.
var fillableKinds = map[contracts.Kind]bool{
	contracts.KindGrid:        true,
	contracts.KindTwoColumn:   true,
	contracts.KindBento:       true,
	contracts.KindTable:       true,
	contracts.KindCardSection: true,
}

// Fill returns a copy of doc with VAlignFill applied to every slide that (a)
// has no explicit vertical alignment set (Slide.Align.Vertical == "") and (b)
// contains at least one fillable multi-child container node at top level
// (grid, two_column, bento, table, card_section). Such slides grow their
// containers to consume the body region instead of clustering at the top.
// Slides with an explicit alignment, or with no fillable container (e.g. a
// hero/heading/prose-only cover, or a lone card/stat), are left untouched —
// so a lone sparse node is never ballooned.
//
// VAlignFill only redistributes EXISTING leftover slack to flexible nodes —
// it never adds content — so it cannot cause overflow; a slide with no slack
// is a no-op. That makes applying it unconditionally (whenever the shape
// matches) safe.
//
// Pure + deterministic: the input doc is not mutated. The Slides slice is
// copied by value (Align is a value field on Slide, so writing
// copy[i].Align.Vertical never touches the stored slide); Nodes are pointers
// but Fill only reads them to detect top-level kinds, never mutates them.
func Fill(doc contracts.SlideDoc) contracts.SlideDoc {
	out := doc
	out.Slides = append([]contracts.Slide(nil), doc.Slides...)
	for i, s := range out.Slides {
		if s.Align.Vertical != "" {
			continue
		}
		if !hasFillableContainer(s.Nodes) {
			continue
		}
		s.Align.Vertical = contracts.VAlignFill
		out.Slides[i] = s
	}
	return out
}

// hasFillableContainer reports whether any top-level node in nodes is one of
// the fillable container kinds (grid, two_column, bento, table,
// card_section). It only inspects the top level — a fillable container
// nested inside a card/stat does not count, matching what VAlignFill itself
// grows (the slide's top-level body stack).
func hasFillableContainer(nodes []contracts.SlideNode) bool {
	for _, n := range nodes {
		if fillableKinds[contracts.KindOf(n)] {
			return true
		}
	}
	return false
}
