package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// treeDoc builds a one-slide asset-free doc exercising the Tree node
// (R14.10, D-127): a shallow org chart — root -> 3 children, one with 2
// leaves — covering recursion and every TreeNode field.
func treeDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Tree Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "org-chart",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Org Structure")},
					&contracts.Tree{
						Root: contracts.TreeNode{
							Label:       "CEO",
							Detail:      "Executive lead",
							Icon:        "star",
							AccentIndex: 0,
							Children: []contracts.TreeNode{
								{
									Label:       "VP Engineering",
									Detail:      "Platform + product",
									Icon:        "diamond",
									AccentIndex: 1,
									Children: []contracts.TreeNode{
										{Label: "Eng Manager", Icon: "check", AccentIndex: 2},
										{Label: "Staff Engineer", Icon: "circle", AccentIndex: 2},
									},
								},
								{Label: "VP Sales", Detail: "Revenue + partnerships", Icon: "square", AccentIndex: 2},
								{Label: "VP People", Detail: "Talent + culture", Icon: "triangle", AccentIndex: 1},
							},
						},
						Orientation: contracts.FlowVertical,
					},
				},
			},
		},
	}
}

// TestTreeRendersWithoutError asserts a Tree-bearing doc renders to valid,
// non-empty PPTX bytes.
func TestTreeRendersWithoutError(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(treeDoc(), soul.DeckardWhite())
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

// TestTreeEmitsMoreShapesThanEmptySlide proves the Tree node has a render
// effect, not dead infra: its slide must emit strictly more shapes than a
// blank slide with no nodes.
func TestTreeEmitsMoreShapesThanEmptySlide(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, emptyStats, err := Render(emptySlideDoc(), s)
	if err != nil {
		t.Fatalf("Render(empty) error = %v", err)
	}
	_, treeStats, err := Render(treeDoc(), s)
	if err != nil {
		t.Fatalf("Render(tree) error = %v", err)
	}
	if treeStats.Shapes <= emptyStats.Shapes {
		t.Fatalf("Tree shapes = %d, want > empty-slide shapes %d", treeStats.Shapes, emptyStats.Shapes)
	}
	if len(treeStats.Warnings) != 0 {
		t.Fatalf("Tree stats.Warnings = %v, want empty (zero-overflow)", treeStats.Warnings)
	}
}

// TestTreeDeterministicAcrossRepeatedRenders asserts byte-identical output
// across two renders of the same doc+soul (the render-determinism hard
// contract, CLAUDE §5).
func TestTreeDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := treeDoc()
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

// TestTreeDeterministicAcrossWorkerCounts asserts byte-identical output
// regardless of worker count (the render-determinism hard contract).
func TestTreeDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := treeDoc()
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
