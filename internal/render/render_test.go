package render

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

func TestRenderProducesValidPPTX(t *testing.T) {
	t.Parallel()

	buf, stats, err := Render(testDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
	if stats.Slides == 0 {
		t.Fatalf("Render() stats slides = %d, want > 0", stats.Slides)
	}

	assertValidPPTX(t, buf)
	if _, err := pptx.NewFromBytes(buf); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}
}

func TestRenderDeterministicAcrossRepeatedRenders(t *testing.T) {
	t.Parallel()

	doc := testDoc()
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

func TestRenderDeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	doc := testDoc()
	s := soul.DeckardWhite()

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("render bytes differ across worker counts")
	}
}

func assertValidPPTX(t *testing.T, buf []byte) {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	for _, f := range r.File {
		if f.Name == "[Content_Types].xml" {
			return
		}
	}
	t.Fatal("rendered zip missing [Content_Types].xml")
}

func testDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Render Adapter Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "cover",
				Layout: contracts.LayoutCover,
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Eyebrow: "Phase 3A", Title: "Render adapter", Subtitle: "IR to scene to PPTX"},
					&contracts.Quote{Text: rt("Determinism is a hard contract."), Attribution: "TASK.md"},
				},
				Notes: rt("Speaker notes remain wired through the scene."),
			},
			{
				ID:     "content",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Native node coverage")},
					&contracts.Prose{Paragraphs: []contracts.RichText{rt("This deck exercises the core native render nodes."), rtStyled("Token and literal rich text both map through the adapter.")}},
					&contracts.List{Kind: contracts.ListChecklist, Items: []contracts.ListItem{{Text: rt("Hero"), Checked: true}, {Text: rt("Heading"), Checked: true}, {Text: rt("Callout"), Checked: true}}},
					&contracts.Callout{Kind: contracts.CalloutImportant, Title: "Open switch", Body: rt("Unsupported nodes can be added in Phase 3B without rewiring the driver.")},
					&contracts.Table{Headers: []contracts.RichText{rt("Kind"), rt("Status")}, Rows: [][]contracts.RichText{{rt("table"), rt("native")}, {rt("cards"), rt("recursive")}}, Caption: "Native table render"},
					&contracts.TwoColumn{
						Ratio: contracts.Ratio11,
						Left: []contracts.SlideNode{
							&contracts.Card{Header: "Left card", Eyebrow: "Recursion", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Cards hold child nodes.")}}}, Fill: contracts.ColorSurfaceAlt, Elevation: contracts.ElevationRaised},
						},
						Right: []contracts.SlideNode{
							&contracts.CardSection{
								Header: "Right section",
								Body: []contracts.SlideNode{
									&contracts.Grid{
										Columns: 2,
										Gap:     contracts.SpaceMD,
										Cells: []contracts.SlideNode{
											&contracts.Card{Header: "A", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Grid cell A")}}}},
											&contracts.Card{Header: "B", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Grid cell B")}}}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func rt(text string) contracts.RichText {
	return contracts.RichText{{Text: text}}
}

func rtStyled(text string) contracts.RichText {
	return contracts.RichText{{
		Text: text,
		Style: contracts.RunStyle{
			TypeRole:  contracts.TypeBody,
			Bold:      true,
			Underline: true,
			Link:      true,
			Href:      "https://example.test/render",
		},
		Color: contracts.TextColor{Literal: strings.ToUpper("3b9c94")},
	}}
}
