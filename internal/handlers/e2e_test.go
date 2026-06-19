package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
)

func rt(s string) contracts.RichText { return contracts.RichText{{Text: s}} }

// TestEndToEndGoldenDeck drives a realistic authoring session through the real
// agent-facing handlers: create a deck, compile a chart / code / markdown into
// IR nodes, assemble a multi-slide deck exercising most of the node catalog,
// validate it for export (StyleScore), then export and confirm a valid,
// multi-slide .pptx comes out. This proves the whole backend surface composes,
// not just each unit in isolation.
func TestEndToEndGoldenDeck(t *testing.T) {
	h := testHandlers()
	h.deps.Workspace = t.TempDir()
	ctx := context.Background()

	created, err := h.createDeck(ctx, contracts.CreateDeckInput{Title: "Deckard — Platform Review"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	deckID := created.Structured.DeckID

	// compile a chart, code, and markdown through the authoring helpers.
	chart, err := h.compileChart(ctx, contracts.CompileChartInput{
		Spec: contracts.ChartSpec{
			Type:   "bar",
			Title:  "Latency by quarter",
			Labels: []string{"Q1", "Q2", "Q3"},
			Series: []contracts.ChartSeries{{Values: []float64{120, 98, 61}}},
		},
	})
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	code, err := h.compileCode(ctx, contracts.CompileCodeInput{
		Code:     "func main() {\n\tprintln(\"deckard\")\n}",
		Language: "go",
		Caption:  "main.go",
	})
	if err != nil {
		t.Fatalf("compileCode: %v", err)
	}
	md, err := h.compileMarkdown(ctx, contracts.CompileMarkdownInput{
		Markdown: "## What shipped\n\n- Native render\n- Pure-Go rasterizers\n- StyleScore",
	})
	if err != nil {
		t.Fatalf("compileMarkdown: %v", err)
	}

	slides := []contracts.Slide{
		// 1: cover
		{Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{
			&contracts.Hero{Eyebrow: "2026 H1", Title: "Platform Review", Subtitle: "What shipped, what's next"},
		}},
		// 2: agenda — heading + checklist + divider
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Agenda")},
			&contracts.List{Kind: contracts.ListChecklist, Items: []contracts.ListItem{
				{Text: rt("Results"), Checked: true},
				{Text: rt("Architecture")},
				{Text: rt("Roadmap")},
			}},
			&contracts.Divider{Spacing: contracts.SpaceLG},
		}},
		// 3: two-column of cards
		{Layout: contracts.LayoutTwoColumn, Nodes: []contracts.SlideNode{
			&contracts.TwoColumn{Ratio: contracts.Ratio11,
				Left:  []contracts.SlideNode{&contracts.Card{Header: "Before", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Chromium render path")}}}}},
				Right: []contracts.SlideNode{&contracts.Card{Header: "After", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Native pptx-go render")}}}}},
			},
		}},
		// 4: callout + quote + chips
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Callout{Kind: contracts.CalloutTip, Title: "Highlight", Body: rt("38% lower p99 latency")},
			&contracts.Quote{Text: rt("Everything enters through the eyes."), Attribution: "Design principle"},
			&contracts.Chip{Label: "shipped", Tone: contracts.ChipSolid, Color: contracts.ColorSuccess},
		}},
		// 5: chart + table
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Latency")},
			&chart.Structured.Node,
			&contracts.Table{
				Headers: []contracts.RichText{rt("Quarter"), rt("p99 (ms)")},
				Rows:    [][]contracts.RichText{{rt("Q1"), rt("120")}, {rt("Q3"), rt("61")}},
				Caption: "Quarterly p99",
			},
		}},
		// 6: code block
		{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Entry point")},
			&code.Structured.Node,
		}},
		// 7: markdown-compiled nodes + a flow
		{Layout: contracts.LayoutTitleContent, Nodes: append(append([]contracts.SlideNode{}, md.Structured.Nodes...),
			&contracts.Flow{Orientation: contracts.FlowHorizontal, Connector: contracts.ConnectorArrow, Steps: []contracts.FlowStep{
				{Label: rt("Author")}, {Label: rt("Validate")}, {Label: rt("Export")},
			}},
		)},
		// 8: grid of cards
		{Layout: contracts.LayoutCardGrid, Nodes: []contracts.SlideNode{
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Card{Header: "Tools", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("49 agent-facing")}}}},
				&contracts.Card{Header: "Souls", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Deckard White + variants")}}}},
				&contracts.Card{Header: "Render", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Native, no Chromium")}}}},
				&contracts.Card{Header: "Validate", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("StyleScore + WCAG")}}}},
			}},
		}},
	}

	for i, s := range slides {
		if _, err := h.addSlide(ctx, contracts.AddSlideInput{DeckID: deckID, Slide: s}); err != nil {
			t.Fatalf("addSlide #%d: %v", i+1, err)
		}
	}

	// validate the whole deck for export.
	val, err := h.validateDeckForExport(ctx, contracts.ValidateDeckForExportInput{DeckID: deckID})
	if err != nil {
		t.Fatalf("validateDeckForExport: %v", err)
	}
	if len(val.Structured.PerSlide) != len(slides) {
		t.Fatalf("perSlide = %d, want %d", len(val.Structured.PerSlide), len(slides))
	}
	if val.Structured.Score <= 0 || val.Structured.Score > 1 {
		t.Fatalf("deck StyleScore = %v, want in (0,1]", val.Structured.Score)
	}
	if !val.Structured.OK {
		t.Fatalf("golden deck failed validation (errors present): %v", val.Structured.Blockers)
	}

	// export and confirm a valid, multi-slide PPTX.
	exp, err := h.exportDeck(ctx, contracts.ExportDeckInput{DeckID: deckID})
	if err != nil {
		t.Fatalf("exportDeck: %v", err)
	}
	if exp.Structured.Stats.Slides != len(slides) {
		t.Fatalf("exported slides = %d, want %d", exp.Structured.Stats.Slides, len(slides))
	}
	if exp.Structured.Stats.Shapes == 0 {
		t.Fatal("exported deck has zero shapes")
	}
	buf, err := os.ReadFile(exp.Structured.Path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("exported bytes are not a valid pptx: %v", err)
	}
	t.Logf("golden deck: %d slides, %d shapes, StyleScore %.2f, %d bytes",
		exp.Structured.Stats.Slides, exp.Structured.Stats.Shapes, val.Structured.Score, len(buf))
}
