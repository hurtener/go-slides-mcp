package soul

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// TestSubtleAlphaBandsAreInTheDocumentedRange guards the R13.13 alpha bands
// this package's DefaultDecorPolicy draws from — including watermarkAlpha,
// which no recipe in this PR consumes yet (Decoration.Text/FontSize are
// deferred) but which is defined and documented for the record.
func TestSubtleAlphaBandsAreInTheDocumentedRange(t *testing.T) {
	t.Parallel()

	if textureAlpha <= 0 || textureAlpha > 0.10 {
		t.Errorf("textureAlpha = %v, want in (0, 0.10]", textureAlpha)
	}
	if glowAlpha <= 0 || glowAlpha > 0.15 {
		t.Errorf("glowAlpha = %v, want in (0, 0.15]", glowAlpha)
	}
	if watermarkAlpha <= 0 || watermarkAlpha > 0.12 {
		t.Errorf("watermarkAlpha = %v, want in (0, 0.12]", watermarkAlpha)
	}
}

func TestPaperTint_PureWhiteYieldsLowChromaOffWhite(t *testing.T) {
	t.Parallel()

	got := paperTint("FFFFFF", "F4EFE6")
	if got == "FFFFFF" {
		t.Fatalf("paperTint(FFFFFF, ...) = %q, want != FFFFFF", got)
	}
	r, g, b, ok := parseRGBHex(got)
	if !ok {
		t.Fatalf("paperTint returned unparsable hex %q", got)
	}
	for _, c := range []int{r, g, b} {
		if c < 244 {
			t.Errorf("paperTint(%q) channel %d < 244", got, c)
		}
	}
	maxC, minC := r, r
	for _, c := range []int{g, b} {
		if c > maxC {
			maxC = c
		}
		if c < minC {
			minC = c
		}
	}
	if maxC-minC > 12 {
		t.Errorf("paperTint(%q) chroma spread = %d, want <= 12", got, maxC-minC)
	}
}

func TestPaperTint_AlreadyOffWhiteIsUnchanged(t *testing.T) {
	t.Parallel()

	got := paperTint("FAF7F2", "F4EFE6")
	if got != "FAF7F2" {
		t.Fatalf("paperTint(FAF7F2, ...) = %q, want unchanged FAF7F2", got)
	}
}

func TestPaperTint_MalformedInputReturnsCanvasUnchanged(t *testing.T) {
	t.Parallel()

	got := paperTint("FFFFFF", "not-a-hex")
	if got != "FFFFFF" {
		t.Fatalf("paperTint with malformed surfaceAlt = %q, want canvas unchanged FFFFFF", got)
	}
}

func TestDefaultDecorPolicy_HasContentCoverDarkEntriesAtSubtleAlpha(t *testing.T) {
	t.Parallel()

	theme := DeckardWhite().Theme
	p := DefaultDecorPolicy(theme)
	if p == nil {
		t.Fatal("DefaultDecorPolicy() = nil, want non-nil")
	}

	for _, arch := range []contracts.SlideArchetype{
		contracts.ArchetypeContent, contracts.ArchetypeCover, contracts.ArchetypeDark,
		contracts.ArchetypeSection, contracts.ArchetypeClosing,
	} {
		if _, ok := p.ByArchetype[arch]; !ok {
			t.Errorf("ByArchetype[%q] missing", arch)
		}
	}

	for arch, entry := range p.ByArchetype {
		for i, d := range entry.Decorations {
			if d.Opacity <= 0 || d.Opacity > 0.15 {
				t.Errorf("%s decoration[%d].Opacity = %v, want in (0, 0.15]", arch, i, d.Opacity)
			}
			if d.Preset == presetGridDots || d.Preset == presetStarfield {
				if d.Opacity < 0.04 || d.Opacity > 0.10 {
					t.Errorf("%s texture decoration[%d].Opacity = %v, want in [0.04, 0.10]", arch, i, d.Opacity)
				}
			}
		}
	}
}

func TestDefaultDecorPolicy_ContentUsesPaperAndNeutralGrain(t *testing.T) {
	t.Parallel()

	theme := DeckardWhite().Theme
	p := DefaultDecorPolicy(theme)
	content := p.ByArchetype[contracts.ArchetypeContent]
	if content.Background == nil || content.Background.Color != contracts.ColorPaper {
		t.Fatalf("content.Background = %+v, want Color=paper", content.Background)
	}
	if len(content.Decorations) != 1 {
		t.Fatalf("content.Decorations = %d, want 1", len(content.Decorations))
	}
	d := content.Decorations[0]
	if d.Color == contracts.ColorAccent {
		t.Errorf("content texture Color = accent, want a neutral role (not accent)")
	}
}

func TestDefaultDecorPolicy_DeterministicAcrossCalls(t *testing.T) {
	t.Parallel()

	theme := DeckardWhite().Theme
	a := DefaultDecorPolicy(theme)
	b := DefaultDecorPolicy(theme)
	if len(a.ByArchetype) != len(b.ByArchetype) {
		t.Fatalf("ByArchetype length differs across calls: %d vs %d", len(a.ByArchetype), len(b.ByArchetype))
	}
	for k, av := range a.ByArchetype {
		bv, ok := b.ByArchetype[k]
		if !ok {
			t.Fatalf("archetype %q missing on second call", k)
		}
		if len(av.Decorations) != len(bv.Decorations) {
			t.Fatalf("archetype %q decoration count differs: %d vs %d", k, len(av.Decorations), len(bv.Decorations))
		}
	}
}

func TestDecorPolicyClone_MutationIndependence(t *testing.T) {
	t.Parallel()

	theme := DeckardWhite().Theme
	src := DefaultDecorPolicy(theme)
	clone := src.Clone()

	entry := clone.ByArchetype[contracts.ArchetypeContent]
	entry.Background.Color = contracts.ColorAccent
	if len(entry.Decorations) > 0 {
		entry.Decorations[0].Opacity = 0.99
	}
	clone.ByArchetype[contracts.ArchetypeContent] = entry
	clone.ByArchetype[contracts.SlideArchetype("new")] = contracts.ArchetypeDecor{}

	if src.ByArchetype[contracts.ArchetypeContent].Background.Color != contracts.ColorPaper {
		t.Error("mutating the clone's background mutated the source")
	}
	if len(src.ByArchetype[contracts.ArchetypeContent].Decorations) > 0 &&
		src.ByArchetype[contracts.ArchetypeContent].Decorations[0].Opacity == 0.99 {
		t.Error("mutating the clone's decoration mutated the source")
	}
	if _, ok := src.ByArchetype[contracts.SlideArchetype("new")]; ok {
		t.Error("adding a key to the clone's map mutated the source's map")
	}
}
