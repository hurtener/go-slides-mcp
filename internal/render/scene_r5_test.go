package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeTwoColumnJoinBadge asserts that a TwoColumn with Join=badge and
// JoinLabel maps to a scene.TwoColumn with the matching engine fields (R5 / D-055).
func TestMapNodeTwoColumnJoinBadge(t *testing.T) {
	t.Parallel()

	node := &contracts.TwoColumn{
		Ratio:     contracts.Ratio11,
		Join:      contracts.JoinBadge,
		JoinLabel: "VS",
		Left:      []contracts.SlideNode{&contracts.Hero{Title: "Option A"}},
		Right:     []contracts.SlideNode{&contracts.Hero{Title: "Option B"}},
	}
	sn := mapNode(node)
	tc, ok := sn.(scene.TwoColumn)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.TwoColumn", sn)
	}
	if tc.Join != scene.JoinBadge {
		t.Errorf("Join: got %v, want scene.JoinBadge", tc.Join)
	}
	if tc.JoinLabel != "VS" {
		t.Errorf("JoinLabel: got %q, want %q", tc.JoinLabel, "VS")
	}
}

// TestMapNodeTwoColumnJoinArrow asserts that Join=arrow maps to scene.JoinArrow.
func TestMapNodeTwoColumnJoinArrow(t *testing.T) {
	t.Parallel()

	node := &contracts.TwoColumn{
		Ratio: contracts.Ratio11,
		Join:  contracts.JoinArrow,
		Left:  []contracts.SlideNode{&contracts.Hero{Title: "Before"}},
		Right: []contracts.SlideNode{&contracts.Hero{Title: "After"}},
	}
	sn := mapNode(node)
	tc, ok := sn.(scene.TwoColumn)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.TwoColumn", sn)
	}
	if tc.Join != scene.JoinArrow {
		t.Errorf("Join: got %v, want scene.JoinArrow", tc.Join)
	}
}

// TestMapNodeTwoColumnNoJoinByteIdentical asserts that a TwoColumn with no
// join fields maps to a scene.TwoColumn with JoinNone and empty JoinLabel —
// byte-identical to the pre-R5 render path (additive D-055 contract).
func TestMapNodeTwoColumnNoJoinByteIdentical(t *testing.T) {
	t.Parallel()

	node := &contracts.TwoColumn{
		Ratio: contracts.Ratio11,
		Left:  []contracts.SlideNode{&contracts.Hero{Title: "L"}},
		Right: []contracts.SlideNode{&contracts.Hero{Title: "R"}},
	}
	sn := mapNode(node)
	tc, ok := sn.(scene.TwoColumn)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.TwoColumn", sn)
	}
	if tc.Join != scene.JoinNone {
		t.Errorf("Join: got %v, want scene.JoinNone (zero)", tc.Join)
	}
	if tc.JoinLabel != "" {
		t.Errorf("JoinLabel: got %q, want empty", tc.JoinLabel)
	}
}

// TestMapNodeBentoLabeledRows asserts that a Bento with labeled rows maps to a
// scene.Bento with the matching structure (R5 / D-056).
func TestMapNodeBentoLabeledRows(t *testing.T) {
	t.Parallel()

	node := &contracts.Bento{
		Columns: 3,
		Rows: []contracts.BentoRow{
			{
				Label: "Core",
				Cells: []contracts.BentoCell{
					{Span: 2, Node: &contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Primary"}}}}},
					{Span: 1, Node: &contracts.Chip{Label: "New", Tone: contracts.ChipSolid, Color: contracts.ColorAccent}},
				},
			},
			{
				Label: "Details",
				Cells: []contracts.BentoCell{
					{Span: 1, Node: &contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "A"}}}}},
					{Span: 1, Node: &contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "B"}}}}},
					{Span: 1, Node: &contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "C"}}}}},
				},
			},
		},
	}
	sn := mapNode(node)
	b, ok := sn.(scene.Bento)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Bento", sn)
	}
	if b.Columns != 3 {
		t.Errorf("Columns: got %d, want 3", b.Columns)
	}
	if len(b.Rows) != 2 {
		t.Fatalf("Rows len: got %d, want 2", len(b.Rows))
	}
	if b.Rows[0].Label != "Core" {
		t.Errorf("Rows[0].Label: got %q, want %q", b.Rows[0].Label, "Core")
	}
	if len(b.Rows[0].Cells) != 2 {
		t.Fatalf("Rows[0].Cells len: got %d, want 2", len(b.Rows[0].Cells))
	}
	if b.Rows[0].Cells[0].Span != 2 {
		t.Errorf("Rows[0].Cells[0].Span: got %d, want 2", b.Rows[0].Cells[0].Span)
	}
	if b.Rows[1].Label != "Details" {
		t.Errorf("Rows[1].Label: got %q, want %q", b.Rows[1].Label, "Details")
	}
	if len(b.Rows[1].Cells) != 3 {
		t.Fatalf("Rows[1].Cells len: got %d, want 3", len(b.Rows[1].Cells))
	}
}

// TestMapNodeBentoNoLabels asserts that a Bento with no row labels maps correctly.
func TestMapNodeBentoNoLabels(t *testing.T) {
	t.Parallel()

	node := &contracts.Bento{
		Columns: 2,
		Rows: []contracts.BentoRow{
			{
				Cells: []contracts.BentoCell{
					{Span: 1, Node: &contracts.Hero{Title: "Left"}},
					{Span: 1, Node: &contracts.Hero{Title: "Right"}},
				},
			},
		},
	}
	sn := mapNode(node)
	b, ok := sn.(scene.Bento)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Bento", sn)
	}
	if b.Columns != 2 {
		t.Errorf("Columns: got %d, want 2", b.Columns)
	}
	if len(b.Rows) != 1 {
		t.Fatalf("Rows len: got %d, want 1", len(b.Rows))
	}
	if b.Rows[0].Label != "" {
		t.Errorf("Rows[0].Label: got %q, want empty (no label)", b.Rows[0].Label)
	}
}
