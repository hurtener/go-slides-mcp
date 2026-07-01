package autofit

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// enableShrinkToFit is ladder rung 1: set AutoFit=true on every Hero,
// Heading, and Stat node in the slide's node tree (recursing through
// two_column left/right, grid/bento cells, and card/card_section body).
// Idempotent — a node that already has AutoFit=true is left as the same
// value (no copy), so re-applying this rung to an already-fit slide is a
// no-op.
func enableShrinkToFit(s contracts.Slide) contracts.Slide {
	out := s
	out.Nodes = rewriteNodes(s.Nodes, setAutoFit)
	return out
}

// setAutoFit is the mutate callback for enableShrinkToFit: it flips AutoFit
// on Hero/Heading/Stat leaves, returning the same node unchanged when
// AutoFit is already true (purity: never writes to n's pointee).
func setAutoFit(n contracts.SlideNode) contracts.SlideNode {
	switch v := n.(type) {
	case *contracts.Hero:
		if v.AutoFit {
			return n
		}
		cp := *v
		cp.AutoFit = true
		return &cp
	case *contracts.Heading:
		if v.AutoFit {
			return n
		}
		cp := *v
		cp.AutoFit = true
		return &cp
	case *contracts.Stat:
		if v.AutoFit {
			return n
		}
		cp := *v
		cp.AutoFit = true
		return &cp
	default:
		return n
	}
}

// stepCardsDown is ladder rung 2: step every Card's Size down one level —
// lg -> md, md (or unset, the md default) -> sm — recursing through the same
// container shapes as rung 1. A Card already at sm is left unchanged (the sm
// floor; never steps below it).
func stepCardsDown(s contracts.Slide) contracts.Slide {
	out := s
	out.Nodes = rewriteNodes(s.Nodes, stepCardSize)
	return out
}

// stepCardSize is the mutate callback for stepCardsDown.
func stepCardSize(n contracts.SlideNode) contracts.SlideNode {
	v, ok := n.(*contracts.Card)
	if !ok {
		return n
	}
	next := nextCardSize(v.Size)
	if next == v.Size {
		return n
	}
	cp := *v
	cp.Size = next
	return &cp
}

// nextCardSize maps a CardSize to the next rung down: lg -> md, md (or the
// unset/"" default, which the engine treats as md) -> sm, sm -> sm (floor).
func nextCardSize(s contracts.CardSize) contracts.CardSize {
	switch s {
	case contracts.CardSizeLG:
		return contracts.CardSizeMD
	case contracts.CardSizeSM:
		return contracts.CardSizeSM
	default: // "" (unset, md default) and CardSizeMD both step to sm.
		return contracts.CardSizeSM
	}
}
