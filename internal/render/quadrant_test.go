package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// quadrantDoc builds a one-slide asset-free doc exercising the Quadrant node
// (R14.9, D-124): a labeled 2x2 map with all 4 quadrants titled + tinted and
// 5 plotted items.
func quadrantDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Quadrant Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "matrix",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Prioritization Matrix")},
					&contracts.Quadrant{
						AxisX: contracts.QuadrantAxis{LowLabel: "Low Effort", HighLabel: "High Effort"},
						AxisY: contracts.QuadrantAxis{LowLabel: "Low Impact", HighLabel: "High Impact"},
						Quadrants: [4]contracts.QuadrantCell{
							{Title: "Quick Wins", Fill: contracts.ColorSurfaceAlt},
							{Title: "Big Bets", Fill: contracts.ColorAccentAlt},
							{Title: "Fill-Ins", Fill: contracts.ColorSurface},
							{Title: "Money Pits", Fill: contracts.ColorAccentWarm},
						},
						Items: []contracts.QuadrantItem{
							{X: 0.15, Y: 0.85, Label: "Onboarding revamp", AccentIndex: 0},
							{X: 0.8, Y: 0.9, Label: "Platform rebuild", AccentIndex: 1},
							{X: 0.2, Y: 0.2, Label: "Docs polish", AccentIndex: 2},
							{X: 0.75, Y: 0.15, Label: "Legacy migration", AccentIndex: 0},
							{X: 0.5, Y: 0.5, Label: "API v2", AccentIndex: 1},
						},
					},
				},
			},
		},
	}
}

// TestQuadrantRendersWithoutError asserts a Quadrant-bearing doc renders to
// valid, non-empty PPTX bytes.
func TestQuadrantRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(quadrantDoc(), soul.DeckardWhite())
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

// TestQuadrantEmitsMoreShapesThanEmptySlide proves the Quadrant node has a
// render effect, not dead infra: its slide must emit strictly more shapes
// than a blank slide with no nodes.
func TestQuadrantEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, quadrantStats, err := Render(quadrantDoc(), s)
	if err != nil {
		t.Fatalf("Render(quadrant) error = %v", err)
	}
	if quadrantStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("Quadrant shapes = %d, want > empty-slide shapes %d", quadrantStats.Shapes, emptyStats.Shapes)
	}
	if len(quadrantStats.Warnings) != 0 {
		t.Fatalf("Quadrant stats.Warnings = %v, want empty (zero-overflow)", quadrantStats.Warnings)
	}
}

// TestQuadrantDeterministicAcrossRepeatedRenders asserts byte-identical
// output across two renders of the same doc+soul (the render-determinism
// hard contract, CLAUDE §5).
func TestQuadrantDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := quadrantDoc()
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

// TestQuadrantDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestQuadrantDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := quadrantDoc()
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
