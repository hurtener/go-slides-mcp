package render

import (
	"reflect"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestApplySectionThemes_DarkSectionSetsUnsetVariant(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1"},
		{ID: "s2"},
	}
	sections := []contracts.DeckSection{
		{Name: "Deep Dive", SlideIDs: []string{"s1"}, Variant: contracts.VariantDark},
	}

	out := ApplySectionThemes(slides, sections)

	if out[0].Variant != contracts.VariantDark {
		t.Fatalf("s1 Variant = %q, want dark", out[0].Variant)
	}
	if out[1].Variant != "" {
		t.Fatalf("s2 Variant = %q, want empty (not a section member)", out[1].Variant)
	}
	// Original input must not be mutated.
	if slides[0].Variant != "" {
		t.Fatalf("input slide mutated: Variant = %q, want empty", slides[0].Variant)
	}
}

func TestApplySectionThemes_ExplicitVariantWins(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1", Variant: contracts.VariantLight},
	}
	sections := []contracts.DeckSection{
		{Name: "Deep Dive", SlideIDs: []string{"s1"}, Variant: contracts.VariantDark},
	}

	out := ApplySectionThemes(slides, sections)

	if out[0].Variant != contracts.VariantLight {
		t.Fatalf("explicit slide Variant overridden: got %q, want light", out[0].Variant)
	}
}

func TestApplySectionThemes_ArchetypeOverride(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1"},
	}
	sections := []contracts.DeckSection{
		{Name: "Intro", SlideIDs: []string{"s1"}, Archetype: contracts.ArchetypeSection},
	}

	out := ApplySectionThemes(slides, sections)

	if out[0].Archetype != contracts.ArchetypeSection {
		t.Fatalf("s1 Archetype = %q, want section", out[0].Archetype)
	}
}

func TestApplySectionThemes_ExplicitArchetypeWins(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1", Archetype: contracts.ArchetypeCover},
	}
	sections := []contracts.DeckSection{
		{Name: "Intro", SlideIDs: []string{"s1"}, Archetype: contracts.ArchetypeSection},
	}

	out := ApplySectionThemes(slides, sections)

	if out[0].Archetype != contracts.ArchetypeCover {
		t.Fatalf("explicit slide Archetype overridden: got %q, want cover", out[0].Archetype)
	}
}

func TestApplySectionThemes_NoOverrideIsByteIdenticalOptOut(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1"},
		{ID: "s2"},
	}

	cases := [][]contracts.DeckSection{
		nil,
		{},
		{{Name: "Untouched", SlideIDs: []string{"s1", "s2"}}},
	}
	for _, sections := range cases {
		out := ApplySectionThemes(slides, sections)
		if !reflect.DeepEqual(out, slides) {
			t.Fatalf("expected byte-identical opt-out for sections=%v, got %+v vs %+v", sections, out, slides)
		}
	}
}

func TestApplySectionThemes_SlideNotInAnyOverridingSectionUnchanged(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1"},
		{ID: "s2"},
	}
	sections := []contracts.DeckSection{
		{Name: "Deep Dive", SlideIDs: []string{"s1"}, Variant: contracts.VariantDark},
	}

	out := ApplySectionThemes(slides, sections)

	if !reflect.DeepEqual(out[1], slides[1]) {
		t.Fatalf("s2 unexpectedly changed: got %+v, want %+v", out[1], slides[1])
	}
}

func TestApplySectionThemes_BothFieldsSetFromSection(t *testing.T) {
	slides := []contracts.Slide{
		{ID: "s1"},
	}
	sections := []contracts.DeckSection{
		{Name: "Deep Dive", SlideIDs: []string{"s1"}, Variant: contracts.VariantDark, Archetype: contracts.ArchetypeDark},
	}

	out := ApplySectionThemes(slides, sections)

	if out[0].Variant != contracts.VariantDark || out[0].Archetype != contracts.ArchetypeDark {
		t.Fatalf("got %+v, want both overrides applied", out[0])
	}
}
