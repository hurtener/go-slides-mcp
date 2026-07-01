package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// funnelDoc builds a one-slide asset-free doc exercising the Funnel node
// (R14.11, D-128): a 4-stage marketing conversion funnel, covering every
// FunnelStage field.
func funnelDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Funnel Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "conversion-funnel",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Conversion Funnel")},
					&contracts.Funnel{
						Stages: []contracts.FunnelStage{
							{Label: "Visitors", Value: "10,000", AccentIndex: 0},
							{Label: "Signups", Value: "2,400", AccentIndex: 1},
							{Label: "Trials", Value: "820", AccentIndex: 2},
							{Label: "Customers", Value: "380", AccentIndex: 0},
						},
					},
				},
			},
		},
	}
}

// TestFunnelRendersWithoutError asserts a Funnel-bearing doc renders to
// valid, non-empty PPTX bytes.
func TestFunnelRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(funnelDoc(), soul.DeckardWhite())
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

// TestFunnelEmitsMoreShapesThanEmptySlide proves the Funnel node has a
// render effect, not dead infra: its slide must emit strictly more shapes
// than a blank slide with no nodes.
func TestFunnelEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, funnelStats, err := Render(funnelDoc(), s)
	if err != nil {
		t.Fatalf("Render(funnel) error = %v", err)
	}
	if funnelStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("Funnel shapes = %d, want > empty-slide shapes %d", funnelStats.Shapes, emptyStats.Shapes)
	}
	if len(funnelStats.Warnings) != 0 {
		t.Fatalf("Funnel stats.Warnings = %v, want empty (zero-overflow)", funnelStats.Warnings)
	}
}

// TestFunnelDeterministicAcrossRepeatedRenders asserts byte-identical output
// across two renders of the same doc+soul (the render-determinism hard
// contract, CLAUDE §5).
func TestFunnelDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := funnelDoc()
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

// TestFunnelDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestFunnelDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := funnelDoc()
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
