package soul

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestRefineReturnsCloneAndRecolorsAccent(t *testing.T) {
	s := DeckardWhite()
	refined, err := Refine(s, []TokenOverride{{Category: "surface", Token: "accent", Value: "DB2777"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := refined.Theme.ResolveColor(pptx.ColorAccent); got != "DB2777" {
		t.Fatalf("refined accent = %q, want DB2777", got)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "3B9C94" {
		t.Fatalf("source accent = %q, want 3B9C94", got)
	}
}

func TestRefineUnknownTokenErrors(t *testing.T) {
	_, err := Refine(DeckardWhite(), []TokenOverride{{Category: "surface", Token: "missing", Value: "DB2777"}})
	if err == nil || !strings.Contains(err.Error(), `unknown surface token "missing"`) {
		t.Fatalf("error = %v, want unknown surface token", err)
	}
}

func TestRefineMalformedSpaceValueErrors(t *testing.T) {
	_, err := Refine(DeckardWhite(), []TokenOverride{{Category: "space", Token: "md", Value: "abc"}})
	if err == nil || !strings.Contains(err.Error(), `invalid point value "abc"`) {
		t.Fatalf("error = %v, want invalid point value", err)
	}
}

func TestRefineExtensionOverrideWritesThrough(t *testing.T) {
	refined, err := Refine(DeckardWhite(), []TokenOverride{{Category: "extension", Token: "outline", Value: "ABCDEF"}})
	if err != nil {
		t.Fatal(err)
	}
	if refined.Extensions["outline"] != "ABCDEF" {
		t.Fatalf("extension = %q, want ABCDEF", refined.Extensions["outline"])
	}
}

func TestRefineDarkSurfaceAndDarkTextOverridesPopulateDarkColors(t *testing.T) {
	s := DeckardWhite()
	if s.Theme.DarkColors != nil {
		t.Fatalf("precondition: DeckardWhite().Theme.DarkColors = %+v, want nil", s.Theme.DarkColors)
	}
	refined, err := Refine(s, []TokenOverride{
		{Category: "darkSurface", Token: "canvas", Value: "0A1622"},
		{Category: "darkText", Token: "primary", Value: "F5F1E8"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if refined.Theme.DarkColors == nil {
		t.Fatal("refined.Theme.DarkColors = nil, want non-nil (allocated without panic)")
	}
	if got := refined.Theme.DarkColors.Surfaces[pptx.ColorCanvas]; got != "0A1622" {
		t.Fatalf("DarkColors.Surfaces[ColorCanvas] = %q, want 0A1622", got)
	}
	if got := refined.Theme.DarkColors.Text[pptx.TextPrimary]; got != "F5F1E8" {
		t.Fatalf("DarkColors.Text[TextPrimary] = %q, want F5F1E8", got)
	}
	// The source soul (nil DarkColors) is untouched by the clone-and-mutate.
	if s.Theme.DarkColors != nil {
		t.Fatalf("source soul DarkColors mutated: %+v, want still nil", s.Theme.DarkColors)
	}
}

func TestRefineDarkSurfaceAndDarkTextUnknownTokenErrors(t *testing.T) {
	if _, err := Refine(DeckardWhite(), []TokenOverride{{Category: "darkSurface", Token: "missing", Value: "DB2777"}}); err == nil || !strings.Contains(err.Error(), `unknown dark surface token "missing"`) {
		t.Fatalf("error = %v, want unknown dark surface token", err)
	}
	if _, err := Refine(DeckardWhite(), []TokenOverride{{Category: "darkText", Token: "missing", Value: "DB2777"}}); err == nil || !strings.Contains(err.Error(), `unknown dark text token "missing"`) {
		t.Fatalf("error = %v, want unknown dark text token", err)
	}
}
