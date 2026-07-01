package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// cycleDoc builds a one-slide asset-free doc exercising the Cycle node
// (R14.11, D-128): a 5-stage lifecycle loop, covering every CycleStage
// field.
func cycleDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Cycle Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "product-lifecycle",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Product Lifecycle")},
					&contracts.Cycle{
						Stages: []contracts.CycleStage{
							{Label: "Discover", Icon: "star", AccentIndex: 0},
							{Label: "Plan", Icon: "diamond", AccentIndex: 1},
							{Label: "Build", Icon: "square", AccentIndex: 2},
							{Label: "Ship", Icon: "check", AccentIndex: 0},
							{Label: "Learn", Icon: "circle", AccentIndex: 1},
						},
					},
				},
			},
		},
	}
}

// TestCycleRendersWithoutError asserts a Cycle-bearing doc renders to
// valid, non-empty PPTX bytes.
func TestCycleRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(cycleDoc(), soul.DeckardWhite())
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

// TestCycleEmitsMoreShapesThanEmptySlide proves the Cycle node has a render
// effect, not dead infra: its slide must emit strictly more shapes than a
// blank slide with no nodes.
func TestCycleEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, cycleStats, err := Render(cycleDoc(), s)
	if err != nil {
		t.Fatalf("Render(cycle) error = %v", err)
	}
	if cycleStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("Cycle shapes = %d, want > empty-slide shapes %d", cycleStats.Shapes, emptyStats.Shapes)
	}
	if len(cycleStats.Warnings) != 0 {
		t.Fatalf("Cycle stats.Warnings = %v, want empty (zero-overflow)", cycleStats.Warnings)
	}
}

// TestCycleDeterministicAcrossRepeatedRenders asserts byte-identical output
// across two renders of the same doc+soul (the render-determinism hard
// contract, CLAUDE §5).
func TestCycleDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := cycleDoc()
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

// TestCycleDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestCycleDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := cycleDoc()
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
