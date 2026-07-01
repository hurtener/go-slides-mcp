package autofit

// This file is the rest of R10.11 — the overflow-remediation ladder. Fill
// (autofit.go) only redistributes existing slack; it can never introduce or
// fix overflow. This file adds the second half: render, detect overflow from
// the engine's LayoutWarnings, and apply a fixed, deterministic ladder of
// pure IR rewrites to ONLY the offending slides, re-rendering after each
// rung, until no slide overflows or the ladder is exhausted.
//
// SCOPE — BEST-EFFORT, not a zero-overflow guarantee. The rungs are the only
// overflow levers the product IR exposes today, and they are weak against a
// VERTICAL stack overflow: AutoFit is a width-only shrink-to-one-line (it does
// not reduce a leaf's vertical slot height), and stepping a Card down only
// trims its size-driven padding. The dominant leaf heights (Hero/Heading/Prose
// slots) are fixed and unreachable from the IR, and the engine already applies
// its own gap-/body-shrink before it warns — so a warning means it could not
// fit even after that. R10.11's strict acceptance ("zero unresolved overflow
// warnings") therefore needs the ENGINE to expose a real FitToRegion knob that
// shrinks content vertically AND clears the warning (filed engine gap). Until
// then this ladder measurably improves overflowing slides (AutoFit prevents
// display-title clipping; card-shrink reclaims height) but cannot promise a
// fully overflow-free deck — hence "ladder exhausted" is a valid terminal
// state and the export never fails on residual overflow.

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// ladderRungs is the fixed remediation ladder, in application order. Each
// rung is a pure IR rewrite applied only to slides currently reported as
// overflowing; render-independence (no import of internal/render) is kept by
// having the caller inject overflow detection via OverflowFunc.
var ladderRungs = []func(contracts.Slide) contracts.Slide{
	enableShrinkToFit,
	stepCardsDown,
}

// OverflowFunc renders doc and returns the set of slide IDs whose render
// produced an overflow LayoutWarning. Injected by the caller (the export
// handler renders via internal/render and classifies) so this package stays
// render-independent and deterministic.
type OverflowFunc func(contracts.SlideDoc) (map[string]bool, error)

// Remediate renders doc via overflowing, and while any slide overflows
// applies the next ladder rung to exactly those slides, re-rendering after
// each rung, until no slide overflows or the ladder is exhausted. Returns the
// remediated doc and the number of rungs applied. Deterministic: rung order
// is fixed and each rung is a pure IR rewrite of the offending slides only. A
// doc that already fits (overflowing returns empty on the first call) is
// returned unchanged (0 rungs). Caller applies Fill BEFORE Remediate.
func Remediate(doc contracts.SlideDoc, overflowing OverflowFunc) (contracts.SlideDoc, int, error) {
	current := doc
	rungsApplied := 0

	for _, rung := range ladderRungs {
		bad, err := overflowing(current)
		if err != nil {
			return contracts.SlideDoc{}, rungsApplied, err
		}
		if len(bad) == 0 {
			return current, rungsApplied, nil
		}
		current = applyRung(current, bad, rung)
		rungsApplied++
	}

	return current, rungsApplied, nil
}

// applyRung returns a copy of doc with rung applied to every slide whose ID
// is in targets. Slides not in targets are left as-is (same value, sharing
// the same Nodes backing array — Slide.Nodes is never mutated in place by a
// rung, only rebuilt for the slides it touches).
func applyRung(doc contracts.SlideDoc, targets map[string]bool, rung func(contracts.Slide) contracts.Slide) contracts.SlideDoc {
	out := doc
	out.Slides = append([]contracts.Slide(nil), doc.Slides...)
	for i, s := range out.Slides {
		if !targets[s.ID] {
			continue
		}
		out.Slides[i] = rung(s)
	}
	return out
}
