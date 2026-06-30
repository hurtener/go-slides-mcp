package soul

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestFromTemplateUsesExtractedAccent(t *testing.T) {
	theme := pptx.DefaultTheme().Clone()
	theme.Colors.Surfaces[pptx.ColorAccent] = pptx.RGB("DB2777")

	s, err := FromTemplate("Acme Brand", "From brand kit", theme)
	if err != nil {
		t.Fatalf("FromTemplate: %v", err)
	}
	if got := s.Theme.ResolveColor(pptx.ColorAccent); got != "DB2777" {
		t.Fatalf("accent = %q, want DB2777", got)
	}
	if s.ID != "acme-brand" {
		t.Fatalf("ID = %q, want acme-brand", s.ID)
	}
	if s.Status != "ready" {
		t.Fatalf("status = %q, want ready", s.Status)
	}
	if s.Description != "From brand kit" {
		t.Fatalf("description = %q, want %q", s.Description, "From brand kit")
	}
}

func TestFromTemplateRejectsNilTheme(t *testing.T) {
	if _, err := FromTemplate("Acme", "", nil); err == nil {
		t.Fatal("expected nil-theme error")
	}
}

func TestFromTemplateRejectsEmptyName(t *testing.T) {
	if _, err := FromTemplate("  ", "", pptx.DefaultTheme()); err == nil {
		t.Fatal("expected empty-name error")
	}
}
