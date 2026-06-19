package soul

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestBootstrapInheritsDeckardWhiteWithNameOnly(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "x"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.ID != "x" {
		t.Fatalf("ID = %q, want x", s.ID)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "3B9C94" {
		t.Fatalf("accent = %q, want 3B9C94", got)
	}
	if got := s.Theme.ResolveTextColor(pptx.TextAccent); got != "2B7A73" {
		t.Fatalf("text accent = %q, want 2B7A73", got)
	}
	if s.StyleGuide.NorthStar != "" {
		t.Fatalf("style guide should be cleared for renamed soul, got %q", s.StyleGuide.NorthStar)
	}
	if s.Status != "ready" {
		t.Fatalf("status = %q, want ready", s.Status)
	}
}

func TestBootstrapAccentOverride(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "Acme", Accent: "DB2777"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "DB2777" {
		t.Fatalf("accent = %q, want DB2777", got)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccentAlt); got != "2B7A73" {
		t.Fatalf("accentAlt = %q, want inherited 2B7A73", got)
	}
}

func TestBootstrapRejectsEmptyName(t *testing.T) {
	if _, err := Bootstrap(BootstrapParams{}); err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestBootstrapSlugifiesID(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{Name: "Acme Labs 2.0!"})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.ID != "acme-labs-2-0" {
		t.Fatalf("ID = %q, want acme-labs-2-0", s.ID)
	}
}

func fatalBootstrap(t *testing.T, err error) {
	t.Helper()
	t.Fatal(err)
}
