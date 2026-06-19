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
