package contracts

// SlideArchetype names a slide's role in the deck for the purpose of
// soul-driven decoration (R13.12): cover/section/content/dark/closing. It
// selects the per-archetype background+decoration recipe a bootstrapped soul
// carries (soul.DecorPolicy) when the slide itself sets no explicit
// Background/decorations.
type SlideArchetype string

// Slide archetypes (wire values per R13.12).
const (
	ArchetypeCover   SlideArchetype = "cover"
	ArchetypeSection SlideArchetype = "section"
	ArchetypeContent SlideArchetype = "content"
	ArchetypeDark    SlideArchetype = "dark"
	ArchetypeClosing SlideArchetype = "closing"
)

// IsValid reports whether v is one of the closed SlideArchetype wire values
// (mirrors the ColorRole pattern).
func (v SlideArchetype) IsValid() bool { return IsValidEnum(v, AllowedSlideArchetype()) }

// AllowedSlideArchetype returns the closed set of SlideArchetype wire values.
func AllowedSlideArchetype() []SlideArchetype {
	return []SlideArchetype{
		ArchetypeCover, ArchetypeSection, ArchetypeContent,
		ArchetypeDark, ArchetypeClosing,
	}
}

// allowedStrings returns the closed set of SlideArchetype wire values as
// plain strings — for inclusion in an error message.
func (SlideArchetype) allowedStrings() []string { return stringsFrom(AllowedSlideArchetype()) }

// ArchetypeDecor is the background+decoration recipe for one SlideArchetype
// (R13.12): a bootstrapped soul's DecorPolicy carries one of these per
// archetype it decorates. Applied only when the slide itself sets no explicit
// Background and no top-level Decoration node — an explicit per-slide setting
// always wins.
type ArchetypeDecor struct {
	// Background is the archetype's default full-bleed fill.
	Background *Background `json:"background,omitempty"`
	// Decorations are the ornament nodes prepended to the slide's node tree.
	Decorations []Decoration `json:"decorations,omitempty"`
}

// DecorPolicy is a soul's per-archetype background+decoration policy
// (R13.12/R13.13): a bootstrapped soul carries one so every slide is
// tastefully decorated by default without the caller hand-placing ornaments.
// Nil (the default) is a no-op — the built-in Deckard White soul carries a
// nil policy so its render stays byte-identical to a policy-free deck.
type DecorPolicy struct {
	// ByArchetype maps a SlideArchetype to its default recipe.
	ByArchetype map[SlideArchetype]ArchetypeDecor `json:"byArchetype,omitempty"`
}

// cloneBackground returns a deep copy of b (a new Gradient slice), or nil for
// a nil b.
func cloneBackground(b *Background) *Background {
	if b == nil {
		return nil
	}
	cp := *b
	if b.Gradient != nil {
		cp.Gradient = append([]ColorRole(nil), b.Gradient...)
	}
	return &cp
}

// Clone returns a deep, independent copy of p: a new map, and for each entry
// a deep copy of Background (its Gradient slice) and the Decorations slice.
// Decoration itself has no slice/pointer fields, so copying each element by
// value is already a deep copy. Returns nil for a nil receiver.
func (p *DecorPolicy) Clone() *DecorPolicy {
	if p == nil {
		return nil
	}
	cp := &DecorPolicy{}
	if p.ByArchetype != nil {
		cp.ByArchetype = make(map[SlideArchetype]ArchetypeDecor, len(p.ByArchetype))
		for k, v := range p.ByArchetype {
			entry := ArchetypeDecor{Background: cloneBackground(v.Background)}
			if v.Decorations != nil {
				entry.Decorations = append([]Decoration(nil), v.Decorations...)
			}
			cp.ByArchetype[k] = entry
		}
	}
	return cp
}
