package soul

import (
	"maps"
	"slices"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
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
	// FontProvider, when non-nil, resolves the soul's named font families to
	// embeddable bytes (R9.1). The render path registers it as the engine
	// FontSource and enables save-time embedding, so the faces the deck actually
	// uses ship inside the .pptx and render on any machine without a host
	// install. A nil provider means "embed nothing" — the render path is then
	// byte-identical to the pre-embedding output. It is a runtime capability, not
	// a serialized token: Clone carries the (immutable, shared) provider by
	// reference, and it is never marshaled into a contract.
	FontProvider pptx.FontSource
	// Decor is the soul's per-archetype background+decoration policy
	// (R13.12). Nil (the default) is a no-op — the composer leaves a doc
	// byte-identical when Decor is nil. A non-nil policy decorates each slide
	// by its archetype at render time; it is a runtime capability applied by
	// the composer, not a flattened design token.
	Decor *contracts.DecorPolicy
	// IconSet is the soul's brand icon set (R14.16): glyph-name -> single-path
	// SVG string. It is a runtime capability like FontProvider/Decor, not a
	// flattened design token — the render path registers each entry via
	// scene.WithIconExtension so every Card/Flow/Milestone/etc. icon
	// reference resolves from the brand set before the curated set. A
	// nil/empty IconSet means curated-set-only and renders byte-identical to
	// a soul without the field. Set copy-on-write by ApplyIcons and treated
	// as immutable-after-set.
	IconSet map[string]string
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
	c.IconSet = maps.Clone(s.IconSet)
	c.StyleGuide.Do = slices.Clone(s.StyleGuide.Do)
	c.StyleGuide.Dont = slices.Clone(s.StyleGuide.Dont)
	c.Decor = s.Decor.Clone()
	return &c
}
