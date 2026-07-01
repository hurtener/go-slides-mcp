package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// dataMarkDoc builds a one-slide asset-free doc exercising the DataMark node
// (R14.8, D-122): a bar and a donut side by side under a heading.
func dataMarkDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "DataMark Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "marks",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Marks")},
					&contracts.DataMark{Kind: contracts.DataMarkBar, Value: 0.6, Label: "60%"},
					&contracts.DataMark{Kind: contracts.DataMarkDonut, Value: 0.92, Label: "92%"},
				},
			},
		},
	}
}

// TestDataMarkRendersWithoutError asserts a DataMark-bearing doc renders to
// valid, non-empty PPTX bytes.
func TestDataMarkRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(dataMarkDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
	assertValidPPTX(t, buf)
	if stats.Slides != 1 {
		t.Fatalf("stats.Slides = %d, want 1", stats.Slides)
	}
}

// TestDataMarkEmitsMoreShapesThanEmptySlide proves the DataMark node has a
// render effect, not dead infra: its slide must emit strictly more shapes
// than a blank slide with no nodes.
func TestDataMarkEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, markStats, err := Render(dataMarkDoc(), s)
	if err != nil {
		t.Fatalf("Render(dataMark) error = %v", err)
	}
	if markStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("DataMark shapes = %d, want > empty-slide shapes %d", markStats.Shapes, emptyStats.Shapes)
	}
	if len(markStats.Warnings) != 0 {
		t.Fatalf("DataMark stats.Warnings = %v, want empty (zero-overflow)", markStats.Warnings)
	}
}

// TestDataMarkDeterministicAcrossRepeatedRenders asserts byte-identical
// output across two renders of the same doc+soul (the render-determinism
// hard contract, CLAUDE §5).
func TestDataMarkDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := dataMarkDoc()
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
		t.Fatal("Render() bytes differ across identical renders")
	}
}

// TestDataMarkDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestDataMarkDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := dataMarkDoc()
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
		t.Fatal("render bytes differ across worker counts")
	}
}
