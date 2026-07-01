package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// timelineDoc builds a one-slide asset-free doc exercising the Timeline node
// (R14.4, D-119): a single-lane roadmap of milestones plus two phase bands.
func timelineDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Timeline Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "roadmap",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Roadmap")},
					&contracts.Timeline{
						Milestones: []contracts.Milestone{
							{Position: 0, Label: "Kickoff", Detail: "Scope locked", AccentIndex: 0},
							{Position: 0.33, Label: "Alpha", Detail: "Internal dogfood", AccentIndex: 1},
							{Position: 0.66, Label: "Beta", Detail: "External pilot", AccentIndex: 2},
							{Position: 1, Label: "GA", Detail: "General availability", AccentIndex: 0},
						},
						Bands: []contracts.TimelineBand{
							{From: 0, To: 0.5, Label: "Build", Fill: contracts.ColorSurfaceAlt},
							{From: 0.5, To: 1, Label: "Launch", Fill: contracts.ColorAccentAlt},
						},
					},
				},
			},
		},
	}
}

// emptySlideDoc is the baseline used to prove the Timeline node has a
// render effect (more shapes than an otherwise-empty slide), not dead infra.
func emptySlideDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Empty Baseline",
		Slides: []contracts.Slide{
			{ID: "empty", Layout: contracts.LayoutBlank},
		},
	}
}

// TestTimelineRendersWithoutError asserts a Timeline-bearing doc renders to
// valid, non-empty PPTX bytes.
func TestTimelineRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(timelineDoc(), soul.DeckardWhite())
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

// TestTimelineEmitsMoreShapesThanEmptySlide proves the Timeline node has a
// render effect, not dead infra: its slide must emit strictly more shapes
// than a blank slide with no nodes.
func TestTimelineEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, timelineStats, err := Render(timelineDoc(), s)
	if err != nil {
		t.Fatalf("Render(timeline) error = %v", err)
	}
	if timelineStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("Timeline shapes = %d, want > empty-slide shapes %d", timelineStats.Shapes, emptyStats.Shapes)
	}
	if len(timelineStats.Warnings) != 0 {
		t.Fatalf("Timeline stats.Warnings = %v, want empty (zero-overflow)", timelineStats.Warnings)
	}
}

// TestTimelineDeterministicAcrossRepeatedRenders asserts byte-identical
// output across two renders of the same doc+soul (the render-determinism
// hard contract, CLAUDE §5).
func TestTimelineDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := timelineDoc()
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

// TestTimelineDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestTimelineDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := timelineDoc()
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
