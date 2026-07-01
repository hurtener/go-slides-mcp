package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// validBrandIconSVG mirrors pptx-go's known-good icon fixture
// (scene/icons_validate_test.go validIconSVG): a single-path triangle that
// satisfies the icon translator subset (D-040), so scene.ValidateIcon /
// scene.WithIconExtension accept it.
const validBrandIconSVG = `<svg viewBox="0 0 24 24"><path d="M12 2 L22 22 L2 22 Z"/></svg>`

// iconCardDoc returns a minimal single-slide doc containing one Card whose
// Icon references glyphName.
func iconCardDoc(glyphName string) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Icon Set Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "s1",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Card{
						Icon:   glyphName,
						Header: "Featured",
						Body:   []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("body")}}},
					},
				},
			},
		},
	}
}

// TestRenderIconSetBindingResolvesBrandGlyph proves the R14.16 binding has
// effect: a Card referencing a glyph name that is NOT in the curated set
// fails Stage-1 (the engine's scene.Render errors, naming the icon) when the
// soul carries no IconSet, but succeeds once the soul's IconSet registers
// that glyph via scene.WithIconExtension.
func TestRenderIconSetBindingResolvesBrandGlyph(t *testing.T) {
	t.Parallel()

	doc := iconCardDoc("brandmark")

	// Baseline: no IconSet on the soul -> "brandmark" resolves to neither the
	// curated set nor any extension, so the engine's Stage-1 icon-ref
	// validation errors, naming the icon.
	plain := soul.DeckardWhite()
	if _, _, err := Render(doc, plain); err == nil {
		t.Fatal("Render() with no IconSet and an unregistered glyph: error = nil, want a Stage-1 unknown-icon error")
	} else if !strings.Contains(err.Error(), "brandmark") {
		t.Errorf("error %q should name the unresolved icon %q", err, "brandmark")
	}

	// With the soul's IconSet carrying "brandmark", the same Card resolves
	// and renders without error.
	branded := plain.Clone()
	branded.IconSet = map[string]string{"brandmark": validBrandIconSVG}
	buf, stats, err := Render(doc, branded)
	if err != nil {
		t.Fatalf("Render() with a bound brand glyph: error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("Render() with a bound brand glyph returned empty bytes")
	}
	if stats.Slides == 0 {
		t.Fatalf("Render() stats slides = %d, want > 0", stats.Slides)
	}
}

// TestRenderIconSetDeterministic asserts the bound-icon render path stays
// deterministic across repeated renders and worker counts (the same hard
// contract every other render path carries).
func TestRenderIconSetDeterministic(t *testing.T) {
	t.Parallel()

	doc := iconCardDoc("brandmark")
	s := soul.DeckardWhite().Clone()
	s.IconSet = map[string]string{"brandmark": validBrandIconSVG}

	first, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("Render() bytes differ across identical renders with a bound IconSet")
	}

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("render bytes differ across worker counts with a bound IconSet")
	}
}

// TestRenderEmptyIconSetByteIdentical asserts that a soul with an empty,
// non-nil IconSet renders byte-identical to the same soul without the field
// set at all — no scene.WithIconExtension option is appended when IconSet is
// empty, so the render path is untouched.
func TestRenderEmptyIconSetByteIdentical(t *testing.T) {
	t.Parallel()

	doc := testDoc()

	withoutField := soul.DeckardWhite()
	withEmptySet := soul.DeckardWhite().Clone()
	withEmptySet.IconSet = map[string]string{}

	base, _, err := Render(doc, withoutField)
	if err != nil {
		t.Fatalf("Render(no IconSet) error = %v", err)
	}
	empty, _, err := Render(doc, withEmptySet)
	if err != nil {
		t.Fatalf("Render(empty IconSet) error = %v", err)
	}
	if !bytes.Equal(base, empty) {
		t.Fatal("Render() bytes differ between a soul with no IconSet and one with an empty IconSet")
	}
}
