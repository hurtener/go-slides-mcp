package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// TestEmitGoldenDeckArtifact builds a rich representative deck through the real
// handlers and writes the exported .pptx to DECKARD_EMIT_PATH for visual review.
// Skipped unless DECKARD_EMIT_PATH is set, so normal test runs never emit files.
//
//	DECKARD_EMIT_PATH=/tmp/deckard-golden.pptx go test -run TestEmitGoldenDeckArtifact ./internal/handlers/
func TestEmitGoldenDeckArtifact(t *testing.T) {
	dest := os.Getenv("DECKARD_EMIT_PATH")
	if dest == "" {
		t.Skip("set DECKARD_EMIT_PATH to emit the artifact")
	}

	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Deckard — Platform Review"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	deckID := created.Structured.DeckID

	chart, err := h.compileChart(ctx, contracts.CompileChartInput{
		Spec: contracts.ChartSpec{
			Type:   "bar",
			Title:  "p99 latency by quarter (ms)",
			Labels: []string{"Q1", "Q2", "Q3", "Q4"},
			Series: []contracts.ChartSeries{{Values: []float64{120, 98, 61, 47}}},
		},
	})
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	code, err := h.compileCode(ctx, contracts.CompileCodeInput{
		Code:     "stats, err := scene.Render(pres, sc,\n\tscene.WithWorkers(8),\n\tscene.WithAssetResolver(r),\n)\nif err != nil {\n\treturn fmt.Errorf(\"render: %w\", err)\n}",
		Language: "go",
		Caption:  "render.go",
	})
	if err != nil {
		t.Fatalf("compileCode: %v", err)
	}
	md, err := h.compileMarkdown(ctx, contracts.CompileMarkdownInput{
		Markdown: "## What shipped this half\n\n- Native pure-Go render path (no Chromium)\n- Chart + code rasterizers\n- StyleScore validation with WCAG contrast",
	})
	if err != nil {
		t.Fatalf("compileMarkdown: %v", err)
	}

	slides := []contracts.Slide{
		{Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{
			&contracts.Hero{Eyebrow: "2026 · H1", Title: "Platform Review", Subtitle: "What shipped, and what's next for Deckard"},
		}},
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Agenda")},
			&contracts.List{Kind: contracts.ListChecklist, Items: []contracts.ListItem{
				{Text: rt("Results this half"), Checked: true},
				{Text: rt("Architecture: native render")},
				{Text: rt("Roadmap to GA")},
			}},
			&contracts.Divider{Spacing: contracts.SpaceLG},
			&contracts.Chip{Label: "internal", Tone: contracts.ChipOutline, Color: contracts.ColorAccent},
		}},
		{Layout: contracts.LayoutTwoColumn, Nodes: []contracts.SlideNode{
			&contracts.TwoColumn{Ratio: contracts.Ratio11,
				Left: []contracts.SlideNode{&contracts.Card{Header: "Before", Eyebrow: "v1", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{rt("HTML → Chromium → screenshot. Slow, fragile, heavyweight runtime.")}},
				}}},
				Right: []contracts.SlideNode{&contracts.Card{Header: "After", Eyebrow: "v2", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{rt("Typed IR → pptx-go → native shapes. One static binary, deterministic.")}},
				}}},
			},
		}},
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Callout{Kind: contracts.CalloutTip, Title: "Headline", Body: rt("38% lower p99 latency, half the deploy size.")},
			&contracts.Quote{Text: rt("Everything enters through the eyes."), Attribution: "Design principle"},
		}},
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Latency trend")},
			&chart.Structured.Node,
			&contracts.Table{
				Headers: []contracts.RichText{rt("Quarter"), rt("p99 (ms)"), rt("Δ")},
				Rows: [][]contracts.RichText{
					{rt("Q1"), rt("120"), rt("—")},
					{rt("Q4"), rt("47"), rt("-61%")},
				},
				Caption: "Quarterly p99",
			},
		}},
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("The render call")},
			&code.Structured.Node,
		}},
		{Layout: contracts.LayoutTitleContent, Nodes: append(append([]contracts.SlideNode{}, md.Structured.Nodes...),
			&contracts.Flow{Orientation: contracts.FlowHorizontal, Connector: contracts.ConnectorArrow, Steps: []contracts.FlowStep{
				{Label: rt("Author"), Detail: rt("MCP tools")},
				{Label: rt("Validate"), Detail: rt("StyleScore")},
				{Label: rt("Export"), Detail: rt("deck://")},
			}},
		)},
		{Layout: contracts.LayoutCardGrid, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("What's in the box")},
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Card{Header: "49 tools", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Agent-facing authoring surface")}}}},
				&contracts.Card{Header: "Souls", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Deckard White + variants, bootstrap + refine")}}}},
				&contracts.Card{Header: "Native render", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Pure Go, no Chromium")}}}},
				&contracts.Card{Header: "StyleScore", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("WCAG contrast + overflow")}}}},
			}},
		}},
	}
	for i, s := range slides {
		if _, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: s}); err != nil {
			t.Fatalf("addSlide #%d: %v", i+1, err)
		}
	}

	exp, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: deckID})
	if err != nil {
		t.Fatalf("exportDeck: %v", err)
	}
	buf, err := os.ReadFile(exp.Structured.Path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if err := os.WriteFile(dest, buf, 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	t.Logf("wrote %d-slide deck (%d shapes, %d bytes) to %s",
		exp.Structured.Stats.Slides, exp.Structured.Stats.Shapes, len(buf), dest)
}
