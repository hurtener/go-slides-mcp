package render

import "github.com/hurtener/go-slides-mcp/internal/contracts"

// ApplySectionThemes returns slides with each section's Variant/Archetype
// override filled into the section's member slides that set none (R14.14).
// Copy-on-write: input slides are never mutated. Explicit per-slide
// Variant/Archetype always wins over a section default. A nil/empty sections
// list, or sections that carry no override, returns the input slice
// unchanged (byte-identical opt-out).
func ApplySectionThemes(slides []contracts.Slide, sections []contracts.DeckSection) []contracts.Slide {
	overrides := sectionOverridesBySlideID(sections)
	if len(overrides) == 0 {
		return slides
	}

	out := slides
	copied := false
	for i, slide := range slides {
		ov, ok := overrides[slide.ID]
		if !ok {
			continue
		}
		setsVariant := slide.Variant == "" && ov.Variant != ""
		setsArchetype := slide.Archetype == "" && ov.Archetype != ""
		if !setsVariant && !setsArchetype {
			continue
		}
		if !copied {
			out = append([]contracts.Slide(nil), slides...)
			copied = true
		}
		s := out[i]
		if setsVariant {
			s.Variant = ov.Variant
		}
		if setsArchetype {
			s.Archetype = ov.Archetype
		}
		out[i] = s
	}
	return out
}

// sectionOverridesBySlideID maps each slide ID belonging to a section that
// carries a Variant and/or Archetype override to that section's override
// values. A section with neither field set contributes no entries.
func sectionOverridesBySlideID(sections []contracts.DeckSection) map[string]contracts.DeckSection {
	var overrides map[string]contracts.DeckSection
	for _, section := range sections {
		if section.Variant == "" && section.Archetype == "" {
			continue
		}
		for _, id := range section.SlideIDs {
			if overrides == nil {
				overrides = make(map[string]contracts.DeckSection)
			}
			overrides[id] = section
		}
	}
	return overrides
}
