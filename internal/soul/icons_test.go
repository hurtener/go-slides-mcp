package soul

import (
	"strings"
	"testing"
)

// validIconSVG and arcIconSVG mirror pptx-go's scene package fixtures
// (scene/icons_validate_test.go): a single-path triangle satisfies the icon
// translator subset; a path with an elliptical arc command violates it
// (D-040), so ValidateIcon/ApplyIcons must reject it.
const validIconSVG = `<svg viewBox="0 0 24 24"><path d="M12 2 L22 22 L2 22 Z"/></svg>`
const arcIconSVG = `<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`

// TestApplyIconsAcceptsValidGlyph asserts that ApplyIcons binds a valid
// single-path SVG to the soul's IconSet under its glyph name, without
// mutating the source soul (Refine/Clone convention).
func TestApplyIconsAcceptsValidGlyph(t *testing.T) {
	s := DeckardWhite()
	refined, err := ApplyIcons(s, map[string]string{"brandmark": validIconSVG})
	if err != nil {
		t.Fatalf("ApplyIcons() error = %v", err)
	}
	if got := refined.IconSet["brandmark"]; got != validIconSVG {
		t.Fatalf("refined.IconSet[brandmark] = %q, want %q", got, validIconSVG)
	}
	if len(s.IconSet) != 0 {
		t.Fatalf("source soul IconSet mutated: %v", s.IconSet)
	}
}

// TestApplyIconsRejectsInvalidGlyph asserts that ApplyIcons rejects an SVG
// outside the translator subset (an elliptical arc) via scene.ValidateIcon,
// naming the offending glyph, and applies no change.
func TestApplyIconsRejectsInvalidGlyph(t *testing.T) {
	s := DeckardWhite()
	_, err := ApplyIcons(s, map[string]string{"bad-glyph": arcIconSVG})
	if err == nil {
		t.Fatal("ApplyIcons() error = nil, want a validation error")
	}
	if !strings.Contains(err.Error(), "bad-glyph") {
		t.Errorf("error %q should name the invalid glyph", err)
	}
}

// TestApplyIconsMergesAndNewWins asserts that a second ApplyIcons call
// preserves prior glyphs and lets a same-named new glyph win, building a
// fresh map each time (copy-on-write).
func TestApplyIconsMergesAndNewWins(t *testing.T) {
	s := DeckardWhite()
	first, err := ApplyIcons(s, map[string]string{"brandmark": validIconSVG})
	if err != nil {
		t.Fatalf("first ApplyIcons() error = %v", err)
	}
	other := `<svg viewBox="0 0 24 24"><path d="M2 2 L20 2 L20 20 Z"/></svg>`
	second, err := ApplyIcons(first, map[string]string{"other": other})
	if err != nil {
		t.Fatalf("second ApplyIcons() error = %v", err)
	}
	if second.IconSet["brandmark"] != validIconSVG {
		t.Errorf("second.IconSet[brandmark] = %q, want it preserved", second.IconSet["brandmark"])
	}
	if second.IconSet["other"] != other {
		t.Errorf("second.IconSet[other] = %q, want %q", second.IconSet["other"], other)
	}
	if len(first.IconSet) != 1 {
		t.Errorf("first.IconSet mutated by second ApplyIcons call: %v", first.IconSet)
	}
}

// TestApplyIconsEmptyIsNoop asserts that an empty/nil icons map leaves the
// soul's IconSet unchanged (still a clone, per the Refine convention).
func TestApplyIconsEmptyIsNoop(t *testing.T) {
	s := DeckardWhite()
	refined, err := ApplyIcons(s, nil)
	if err != nil {
		t.Fatalf("ApplyIcons(nil) error = %v", err)
	}
	if len(refined.IconSet) != 0 {
		t.Errorf("refined.IconSet = %v, want empty", refined.IconSet)
	}
}
