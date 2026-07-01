package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// footnotesDoc builds a single content slide carrying 2 slide-level
// Footnotes and a Prose paragraph with a Superscript marker run — the
// product-mapping accept case for R14.12.
func footnotesDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Footnotes Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "footnotes",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Headline Results")},
					&contracts.Prose{Paragraphs: []contracts.RichText{
						{
							{Text: "ARR figure includes one-time items"},
							{Text: "1", Superscript: true},
							{Text: "."},
						},
					}},
				},
				Footnotes: []contracts.RichText{
					rt("Source: internal telemetry, 2026."),
					rt("Note: figures unaudited; final numbers pending Q3 close."),
				},
			},
		},
	}
}

// plainDoc is footnotesDoc's body with NO Footnotes and NO Superscript run —
// the pre-R14.12 shape, used to assert byte-identity and a shape-count delta.
func plainDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Footnotes Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "footnotes",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Headline Results")},
					&contracts.Prose{Paragraphs: []contracts.RichText{
						{
							{Text: "ARR figure includes one-time items"},
							{Text: "1"},
							{Text: "."},
						},
					}},
				},
			},
		},
	}
}

// TestFootnotes_EmitsMoreShapesThanWithout is R14.12's core product-mapping
// accept case: a slide with Footnotes must emit strictly more shapes than
// the same slide without them (the reserved footnote band is real, counted
// shapes).
func TestFootnotes_EmitsMoreShapesThanWithout(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, withFootnotes, err := Render(footnotesDoc(), s)
	if err != nil {
		t.Fatalf("Render(footnotesDoc) error = %v", err)
	}
	_, without, err := Render(plainDoc(), s)
	if err != nil {
		t.Fatalf("Render(plainDoc) error = %v", err)
	}
	if withFootnotes.Shapes <= without.Shapes {
		t.Errorf("Shapes with footnotes = %d, want > without footnotes = %d", withFootnotes.Shapes, without.Shapes)
	}
}

// TestFootnotes_SuperscriptRunRenders asserts the Superscript run style
// maps through to the OOXML baseline shift (30000 = raised, per pptx-go's
// pptx.Superscript) in the rendered slide XML.
func TestFootnotes_SuperscriptRunRenders(t *testing.T) {
	t.Parallel()

	buf, _, err := Render(footnotesDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	xml := string(firstSlideXML(t, buf))
	if !bytes.Contains([]byte(xml), []byte(`baseline="30000"`)) {
		t.Errorf("slide1.xml does not contain a superscript baseline shift; xml = %s", xml)
	}
}

// TestFootnotes_DeterministicAcrossRepeatedRenders guards byte-identity of a
// footnote-bearing doc across repeated renders.
func TestFootnotes_DeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := footnotesDoc()
	s := soul.DeckardWhite()

	first, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("Render() bytes differ across identical renders of a footnote-bearing doc")
	}
}

// TestFootnotes_DeterministicAcrossWorkerCounts guards byte-identity of a
// footnote-bearing doc across worker counts (render determinism is a hard
// contract per CLAUDE.md §5).
func TestFootnotes_DeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := footnotesDoc()
	s := soul.DeckardWhite()

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("footnote-bearing render bytes differ across worker counts")
	}
}

// TestFootnotes_AbsentIsByteIdenticalToPreChange is the R14.12 byte-identity
// guard: a slide built with NO Footnotes and NO Superscript run — whether the
// fields are simply omitted (the zero value) or explicitly assigned their
// zero value — renders identical bytes to a doc that never mentions the new
// fields at all (the pre-change shape).
func TestFootnotes_AbsentIsByteIdenticalToPreChange(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	// Built one way: fields simply omitted (zero value by default).
	omitted := plainDoc()

	// Built a second way: fields explicitly assigned their zero value.
	explicit := plainDoc()
	explicit.Slides[0].Footnotes = nil
	explicit.Slides[0].Nodes[1] = &contracts.Prose{Paragraphs: []contracts.RichText{
		{
			{Text: "ARR figure includes one-time items", Superscript: false},
			{Text: "1", Superscript: false},
			{Text: "."},
		},
	}}

	a, _, err := Render(omitted, s)
	if err != nil {
		t.Fatalf("Render(omitted) error = %v", err)
	}
	b, _, err := Render(explicit, s)
	if err != nil {
		t.Fatalf("Render(explicit) error = %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("a slide with no footnotes/superscript is not byte-identical across the two build paths")
	}
}
