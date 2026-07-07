package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func TestMapNodeCardRibbonAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Card{
		Header: "Business",
		Ribbon: &contracts.Ribbon{Text: "MOST POPULAR", Position: contracts.RibbonCornerTR, Color: contracts.ColorAccent, TextColor: contracts.TextInverse},
		Body:   []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Our recommended plan"}}}}},
	}
	sn := mapNode(node)
	c, ok := sn.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", sn)
	}
	if c.Ribbon == nil {
		t.Fatal("Ribbon: got nil, want non-nil")
	}
	if c.Ribbon.Position != scene.RibbonCornerTR {
		t.Errorf("Ribbon.Position: got %v, want scene.RibbonCornerTR", c.Ribbon.Position)
	}
	if c.Ribbon.Color == nil || *c.Ribbon.Color != scene.ColorAccent {
		t.Errorf("Ribbon.Color: got %v, want pointer-to-ColorAccent", c.Ribbon.Color)
	}
	if c.Ribbon.TextColor != scene.TextInverse {
		t.Errorf("Ribbon.TextColor: got %v, want scene.TextInverse", c.Ribbon.TextColor)
	}
}

func TestMapNodeGridConnectorsBiArrow(t *testing.T) {
	t.Parallel()

	node := &contracts.Grid{
		Columns:    3,
		Gap:        contracts.SpaceMD,
		Connectors: []contracts.GridConnector{{Between: [2]int{0, 1}, Kind: contracts.ConnectorBiArrow, Label: "sync"}},
		Cells: []contracts.SlideNode{
			&contracts.Card{Header: "A", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "a"}}}}}},
			&contracts.Card{Header: "B", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "b"}}}}}},
			&contracts.Card{Header: "C", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "c"}}}}}},
		},
	}
	sn := mapNode(node)
	g, ok := sn.(scene.Grid)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Grid", sn)
	}
	if len(g.Connectors) != 1 {
		t.Fatalf("len(Connectors): got %d, want 1", len(g.Connectors))
	}
	if g.Connectors[0].Kind != scene.ConnectorBiArrow {
		t.Errorf("Connectors[0].Kind: got %v, want scene.ConnectorBiArrow", g.Connectors[0].Kind)
	}
	if g.Connectors[0].Between != [2]int{0, 1} {
		t.Errorf("Connectors[0].Between: got %v, want [0 1]", g.Connectors[0].Between)
	}
}

func TestMapNodeTwoColumnJoinPosition(t *testing.T) {
	t.Parallel()

	node := &contracts.TwoColumn{
		Ratio:        contracts.Ratio11,
		Join:         contracts.JoinBadge,
		JoinLabel:    "One agent",
		JoinPosition: contracts.JoinTopBridge,
		Left:         []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Build internally"}}}}},
		Right:        []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Sell externally"}}}}},
	}
	sn := mapNode(node)
	tc, ok := sn.(scene.TwoColumn)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.TwoColumn", sn)
	}
	if tc.JoinPosition != scene.JoinTopBridge {
		t.Errorf("JoinPosition: got %v, want scene.JoinTopBridge", tc.JoinPosition)
	}
}

func TestRenderRibbonGridConnJoinPosEffect(t *testing.T) {
	t.Parallel()

	empty := contracts.SlideDoc{Title: "baseline", Slides: []contracts.Slide{{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}}}}}}
	withNodes := contracts.SlideDoc{Title: "effect", Slides: []contracts.Slide{{ID: "s", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
		&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Header"}}},
		&contracts.Grid{
			Columns:    3,
			Gap:        contracts.SpaceMD,
			Connectors: []contracts.GridConnector{{Between: [2]int{0, 1}, Kind: contracts.ConnectorBiArrow, Label: "sync"}, {Between: [2]int{1, 2}, Kind: contracts.ConnectorArrow, Label: "ship"}},
			Cells: []contracts.SlideNode{
				&contracts.Card{Header: "Starter", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "For small teams"}}}}}},
				&contracts.Card{Header: "Business", Ribbon: &contracts.Ribbon{Text: "MOST POPULAR", Position: contracts.RibbonTopBar, Color: contracts.ColorAccent, TextColor: contracts.TextInverse}, Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Recommended"}}}}}},
				&contracts.Card{Header: "Enterprise", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Custom scale"}}}}}},
			},
		},
		&contracts.TwoColumn{Ratio: contracts.Ratio11, Join: contracts.JoinBadge, JoinLabel: "One agent", JoinPosition: contracts.JoinTopBridge, Left: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Build internally"}}}}}, Right: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Sell externally"}}}}}},
	}}}}
	s := soul.DeckardWhite()

	base, baseStats, err := Render(empty, s)
	if err != nil {
		t.Fatalf("baseline Render() error = %v", err)
	}
	assertValidPPTX(t, base)

	got, gotStats, err := Render(withNodes, s)
	if err != nil {
		t.Fatalf("withNodes Render() error = %v", err)
	}
	assertValidPPTX(t, got)
	if gotStats.Shapes <= baseStats.Shapes {
		t.Errorf("Ribbon+GridConn+JoinPos Shapes = %d, want > baseline %d (fields render no shapes?)", gotStats.Shapes, baseStats.Shapes)
	}
	if _, err := pptx.NewFromBytes(got); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}
	again, _, err := Render(withNodes, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(got, again) {
		t.Fatal("Render() bytes differ across identical renders")
	}
	defW, _, err := renderWithWorkers(withNodes, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneW, _, err := renderWithWorkers(withNodes, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defW, oneW) {
		t.Fatal("render bytes differ across worker counts")
	}
}
