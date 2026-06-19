package soul

import (
	"maps"
	"slices"

	"github.com/hurtener/pptx-go/pptx"
)

// Soul is a Deckard design soul: a complete pptx-go Theme (the deck's visual
// truth) plus the metadata and design voice that make it a product object. The
// soul is the source of truth — pptx-go code-authored themes do not round-trip
// through theme1.xml, so Deckard persists the soul and re-applies it per render.
type Soul struct {
	// ID is the stable soul identifier (e.g. "deckard-white").
	ID string
	// Name is the human-facing soul name.
	Name string
	// Description is an optional one-line summary.
	Description string
	// Status is the lifecycle state (e.g. "ready").
	Status string
	// Theme is the resolved pptx-go theme this soul renders through.
	Theme *pptx.Theme
	// StyleGuide is the design voice (north star + do/don't) surfaced to agents.
	StyleGuide StyleGuide
	// Extensions carries tokens that have NO native pptx.Theme field — e.g.
	// "border", "borderStrong", "tooltip", "accentSoft" — as hex strings.
	// Applied to deck shapes as literal strokes/washes by the renderer.
	Extensions map[string]string
}

// StyleGuide is a soul's design voice, shown to agents to steer authoring.
type StyleGuide struct {
	// NorthStar is the one-line design intent.
	NorthStar string
	// Do lists encouraged practices.
	Do []string
	// Dont lists discouraged practices.
	Dont []string
}

// Clone returns a deep, independent copy: the Theme is cloned (every map
// reallocated) and the metadata maps/slices are copied, so refining a clone
// never mutates the source soul.
func (s *Soul) Clone() *Soul {
	if s == nil {
		return nil
	}
	c := *s
	if s.Theme != nil {
		c.Theme = s.Theme.Clone()
	}
	c.Extensions = maps.Clone(s.Extensions)
	c.StyleGuide.Do = slices.Clone(s.StyleGuide.Do)
	c.StyleGuide.Dont = slices.Clone(s.StyleGuide.Dont)
	return &c
}
