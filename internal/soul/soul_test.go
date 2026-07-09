package soul

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestDeckardWhiteGolden(t *testing.T) {
	s := DeckardWhite()
	if s.Theme == nil {
		t.Fatal("DeckardWhite: nil theme")
	}
	th := s.Theme

	for _, c := range []struct {
		role pptx.ColorRole
		want pptx.RGB
	}{
		{pptx.ColorCanvas, "FAF7F2"},
		{pptx.ColorSurface, "FFFFFF"},
		{pptx.ColorSurfaceAlt, "F4EFE6"},
		{pptx.ColorAccent, "3B9C94"},
		{pptx.ColorAccentAlt, "2B7A73"},
		{pptx.ColorAccentWarm, "D97B1A"},
		{pptx.ColorSuccess, "3F8E6B"},
		{pptx.ColorError, "B64A4A"},
		{pptx.ColorInfo, "2B7A73"},
	} {
		if got := th.ResolveColor(c.role); got != c.want {
			t.Errorf("surface role %d = %q, want %q", c.role, got, c.want)
		}
	}

	for _, c := range []struct {
		role pptx.TextColorRole
		want pptx.RGB
	}{
		{pptx.TextPrimary, "2B2723"},
		{pptx.TextSecondary, "6A625B"},
		{pptx.TextInverse, "FAF7F2"},
		{pptx.TextAccent, "2B7A73"},
		{pptx.TextError, "B64A4A"},
	} {
		if got := th.ResolveTextColor(c.role); got != c.want {
			t.Errorf("text role %d = %q, want %q", c.role, got, c.want)
		}
	}

	for _, c := range []struct {
		role   pptx.TypeRole
		fam    string
		size   float64
		weight int
	}{
		{pptx.TypeDisplay, "Playfair Display", 40, 400},
		{pptx.TypeH1, "Lora", 32, 400},
		{pptx.TypeH4, "Inter", 20, 500},
		{pptx.TypeBody, "Inter", 14, 400},
	} {
		fs := th.ResolveType(c.role)
		if fs.Family != c.fam || fs.Size != c.size || fs.Weight != c.weight {
			t.Errorf("type role %d = %+v, want %s/%v/%d", c.role, fs, c.fam, c.size, c.weight)
		}
		if fs.AvgCharWidth <= 0 {
			t.Errorf("type role %d AvgCharWidth = %.4f, want > 0", c.role, fs.AvgCharWidth)
		}
		if len(fs.Fallback) == 0 {
			t.Errorf("type role %d Fallback = nil, want a controlled fallback chain", c.role)
		}
	}
	if display := th.ResolveType(pptx.TypeDisplay); len(display.Fallback) == 0 || display.Fallback[0] != "Lora" {
		t.Errorf("display fallback = %v, want first fallback Lora", display.Fallback)
	}
	if body := th.ResolveType(pptx.TypeBody); len(body.Fallback) == 0 || body.Fallback[0] != "Calibri" {
		t.Errorf("body fallback = %v, want first fallback Calibri", body.Fallback)
	}

	// The 400/500-weights-only rule.
	for _, role := range []pptx.TypeRole{
		pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3,
		pptx.TypeH4, pptx.TypeH5, pptx.TypeBody, pptx.TypeBodySmall,
	} {
		if w := th.ResolveType(role).Weight; w > 500 {
			t.Errorf("type role %d weight %d violates the 400/500 rule", role, w)
		}
	}

	if th.ResolveSpace(pptx.SpaceMD) != pptx.Pt(12) {
		t.Errorf("SpaceMD = %v, want Pt(12)", th.ResolveSpace(pptx.SpaceMD))
	}
	if th.ResolveRadius(pptx.RadiusMD) != pptx.Pt(12) {
		t.Errorf("RadiusMD = %v, want Pt(12)", th.ResolveRadius(pptx.RadiusMD))
	}

	if s.ID != DeckardWhiteID || s.Name != "Deckard White" || s.Status != "ready" {
		t.Errorf("metadata = %q/%q/%q", s.ID, s.Name, s.Status)
	}
	if s.Extensions["border"] != "E0D5CA" {
		t.Errorf("border extension = %q, want E0D5CA", s.Extensions["border"])
	}
	if s.StyleGuide.NorthStar == "" || len(s.StyleGuide.Do) == 0 {
		t.Error("style guide should be populated")
	}
}

func TestSoulCloneIndependent(t *testing.T) {
	s := DeckardWhite()
	s.IconSet = map[string]string{"brandmark": "<svg/>"}
	c := s.Clone()
	c.Theme.Colors.Surfaces[pptx.ColorAccent] = "000000"
	c.Extensions["border"] = "ZZZZZZ"
	c.IconSet["brandmark"] = "<svg mutated/>"
	c.Name = "Mutated"
	c.StyleGuide.Do[0] = "changed"

	if s.Theme.ResolveColor(pptx.ColorAccent) != "3B9C94" {
		t.Error("clone mutated the source theme")
	}
	if s.Extensions["border"] != "E0D5CA" {
		t.Error("clone mutated the source extensions")
	}
	if s.IconSet["brandmark"] != "<svg/>" {
		t.Error("clone mutated the source icon set")
	}
	if s.Name != "Deckard White" {
		t.Error("clone mutated the source name")
	}
	if s.StyleGuide.Do[0] == "changed" {
		t.Error("clone mutated the source style guide")
	}
}

func TestRegistry(t *testing.T) {
	r := NewMemoryRegistry()
	got, ok := r.Get(DeckardWhiteID)
	if !ok || got.Name != "Deckard White" {
		t.Fatal("registry not seeded with Deckard White")
	}
	// Get returns a clone — mutating it must not touch the store.
	got.Name = "tampered"
	again, _ := r.Get(DeckardWhiteID)
	if again.Name != "Deckard White" {
		t.Error("Get did not return a clone")
	}

	if err := r.Put(&Soul{ID: "mono", Name: "Mono"}); err != nil {
		t.Fatal(err)
	}
	if list := r.List(); len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
	if _, ok := r.Get("missing"); ok {
		t.Error("Get of missing id should be false")
	}
	if err := r.Put(&Soul{ID: ""}); err == nil {
		t.Error("Put with empty ID should error")
	}
}
