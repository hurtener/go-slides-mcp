package render

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// applyDecorPolicy fills in each slide's Background/decorations from the
// soul's per-archetype DecorPolicy (R13.12), when the slide itself sets
// neither. It is copy-on-write throughout: doc/slide/nodes are never mutated
// in place, so a caller-held SlideDoc is untouched.
//
// The nil-policy path returns doc UNCHANGED (the same value, no copy) — this
// is the byte-identity fast path the built-in Deckard White soul (Decor ==
// nil) takes on every render.
func applyDecorPolicy(doc contracts.SlideDoc, s *soul.Soul) contracts.SlideDoc {
	if s == nil || s.Decor == nil || len(s.Decor.ByArchetype) == 0 {
		return doc
	}

	n := len(doc.Slides)
	slides := make([]contracts.Slide, n)
	for i, slide := range doc.Slides {
		arch := slide.Archetype
		if arch == "" {
			arch = inferArchetype(slide, i, n)
		}
		entry, ok := s.Decor.ByArchetype[arch]
		if !ok {
			slides[i] = slide
			continue
		}
		slides[i] = applyArchetypeDecor(slide, entry)
	}
	doc.Slides = slides
	return doc
}

// applyArchetypeDecor returns a copy of slide with entry's Background and
// decorations filled in wherever the slide itself sets none. Explicit
// per-slide settings always win.
func applyArchetypeDecor(slide contracts.Slide, entry contracts.ArchetypeDecor) contracts.Slide {
	out := slide
	if out.Background == nil && entry.Background != nil {
		out.Background = cloneBackground(entry.Background)
	}
	if len(entry.Decorations) > 0 && !hasDecoration(out.Nodes) {
		nodes := make([]contracts.SlideNode, 0, len(entry.Decorations)+len(out.Nodes))
		for _, d := range entry.Decorations {
			cp := d
			nodes = append(nodes, &cp)
		}
		nodes = append(nodes, out.Nodes...)
		out.Nodes = nodes
	}
	return out
}

// cloneBackground is a package-local passthrough to contracts' deep-copy
// helper — kept named identically so a reader sees the same contract at both
// call sites (composer + DecorPolicy.Clone).
func cloneBackground(b *contracts.Background) *contracts.Background {
	cp := *b
	if b.Gradient != nil {
		cp.Gradient = append([]contracts.ColorRole(nil), b.Gradient...)
	}
	return &cp
}

// hasDecoration reports whether any top-level node in nodes is a Decoration —
// the "does not set its own decorations" test (R13.12).
func hasDecoration(nodes []contracts.SlideNode) bool {
	for _, n := range nodes {
		if _, ok := n.(*contracts.Decoration); ok {
			return true
		}
	}
	return false
}

// inferArchetype infers a SlideArchetype for a slide that set none, from its
// Variant/Layout/index (R13.12). Closing is only reachable via an explicit
// Slide.Archetype — there is no positional signal for "last slide is a
// closing slide" (a deck's last slide is very often ordinary content). The
// slide count parameter is currently unused by this inference table; it
// stays part of the signature (matching applyDecorPolicy's call site) so a
// future rule (e.g. "last slide") can be added without an API break.
func inferArchetype(s contracts.Slide, idx int, _ int) contracts.SlideArchetype {
	switch {
	case s.Variant == contracts.VariantDark:
		return contracts.ArchetypeDark
	case s.Layout == contracts.LayoutCover:
		return contracts.ArchetypeCover
	case idx == 0:
		return contracts.ArchetypeCover
	case s.Layout == contracts.LayoutFullBleed:
		return contracts.ArchetypeSection
	default:
		return contracts.ArchetypeContent
	}
}
