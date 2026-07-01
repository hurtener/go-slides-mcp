// Package conformance is the R14.19 archetype conformance corpus: a
// deterministic set of one-slide fixtures, one per professional deck
// archetype, plus the soul variants they render through. corpus.go holds the
// fixture/soul builders (a plain Go package, not test-only, so a future
// non-test caller — e.g. a CLI audit — can reuse them); conformance_test.go
// holds the invariant runner.
//
// Adding a new archetype is a ONE-FIXTURE addition: write a `func() SlideDoc`
// builder and append one Fixture{Name, Build} entry to the Archetypes slice
// below (see TestConformanceCorpus_IsExtensible in conformance_test.go).
package conformance

import (
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// Fixture is one archetype's fixture: a stable Name (used in test output to
// pinpoint a regressing class) and a Build func returning a fresh one-slide
// SlideDoc each call (fixtures are rebuilt per soul so no render mutates a
// shared doc).
type Fixture struct {
	Name  string
	Build func() contracts.SlideDoc
}

// rt builds a single-run, unstyled RichText — the shared helper every
// fixture builder below uses for plain text fields.
func rt(text string) contracts.RichText {
	return contracts.RichText{{Text: text}}
}

// Archetypes is the corpus registry (R14.19): one entry per professional deck
// archetype expressible with today's node vocabulary (see the scope note in
// corpus.go's package doc — timeline/org-chart/quadrant/funnel/agenda/
// logo-wall/RTL nodes are not built yet and are NOT stubbed in here). Every
// entry uses only native, asset-free nodes so a fixture render never depends
// on an AssetResolver.
var Archetypes = []Fixture{
	{"cover", coverDoc},
	{"section", sectionDoc},
	{"content", contentDoc},
	{"card-grid", cardGridDoc},
	{"stat-strip", statStripDoc},
	{"two-column", twoColumnDoc},
	{"quote", quoteDoc},
	{"flow", flowDoc},
	{"comparison-table", comparisonTableDoc},
	{"bento", bentoDoc},
	{"dark-feature", darkFeatureDoc},
	{"closing", closingDoc},
	{"dashboard", dashboardDoc},
	{"cover-mesh", coverMeshDoc},
	{"watermark-content", watermarkContentDoc},
	{"focal-card", focalCardDoc},
}

// oneSlide wraps a single slide into a titled SlideDoc — the shared shape
// every fixture builder below returns.
func oneSlide(title string, s contracts.Slide) contracts.SlideDoc {
	return contracts.SlideDoc{Title: title, Slides: []contracts.Slide{s}}
}

func coverDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Cover", contracts.Slide{
		ID:        "cover",
		Archetype: contracts.ArchetypeCover,
		Layout:    contracts.LayoutCover,
		Nodes: []contracts.SlideNode{
			&contracts.Hero{
				Eyebrow:  "FY26 Board Review",
				Title:    "The State of the Platform",
				Subtitle: "A quarterly look at growth, reliability, and what's next",
			},
		},
	})
}

// coverMeshDoc is a cover slide whose Background is a two-glow mesh wash
// (R13.4) — the corpus's product-level accept case for BackgroundMesh,
// asset-free like every other archetype fixture in this file.
func coverMeshDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Cover Mesh", contracts.Slide{
		ID:        "cover-mesh",
		Archetype: contracts.ArchetypeCover,
		Layout:    contracts.LayoutCover,
		Background: &contracts.Background{
			Kind: contracts.BackgroundMesh,
			Mesh: []contracts.MeshGlow{
				{Anchor: contracts.AnchorTopLeft, Color: contracts.ColorAccent, Radius: 240, Alpha: 0.12},
				{Anchor: contracts.AnchorBottomRight, Color: contracts.ColorAccentAlt, Radius: 200, Alpha: 0.08},
			},
		},
		Nodes: []contracts.SlideNode{
			&contracts.Hero{
				Eyebrow:  "FY26 Board Review",
				Title:    "The State of the Platform",
				Subtitle: "A quarterly look at growth, reliability, and what's next",
			},
		},
	})
}

func sectionDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Section", contracts.Slide{
		ID:        "section",
		Archetype: contracts.ArchetypeSection,
		Layout:    contracts.LayoutFullBleed,
		Nodes: []contracts.SlideNode{
			&contracts.SectionDivider{Eyebrow: "Part 2", Label: "Product Deep Dive"},
		},
	})
}

func contentDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Content", contracts.Slide{
		ID:        "content",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Key Findings")},
			&contracts.Prose{Paragraphs: []contracts.RichText{
				rt("This quarter we shipped three major features and cut latency across the board."),
			}},
			&contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{
				{Text: rt("Shipped the new dashboard")},
				{Text: rt("Reduced p95 latency by 38%")},
				{Text: rt("Onboarded 12 new customers")},
			}},
		},
	})
}

func cardGridDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Card Grid", contracts.Slide{
		ID:        "card-grid",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Three Pillars")},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Card{Header: "Reliability", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("99.98% uptime this quarter.")}},
					}},
					&contracts.Card{Header: "Speed", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("Median render time under 400ms.")}},
					}},
					&contracts.Card{Header: "Scale", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("3x the decks rendered vs. last quarter.")}},
					}},
				},
			},
		},
	})
}

func statStripDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Stat Strip", contracts.Slide{
		ID:        "stat-strip",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("This Quarter By The Numbers")},
			&contracts.Grid{
				Columns: 4,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Stat{Value: "$2.2M", Label: "ARR", Delta: "+18%", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "12,400", Label: "Active users", Delta: "+9%", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "99.98%", Label: "Uptime", Delta: "+0.1pp", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "1.8%", Label: "Churn", Delta: "-0.4pp", DeltaTone: contracts.DeltaDown},
				},
			},
		},
	})
}

func twoColumnDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Two Column", contracts.Slide{
		ID:        "two-column",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTwoColumn,
		Nodes: []contracts.SlideNode{
			// Left/Right each carry exactly one node: crossing two different
			// node kinds in one column stacks them through the top-level
			// gap estimate (nodesHeight's estGap) vs. the theme's actual
			// SpaceMD, and the two can differ by soul — a single node per
			// column sidesteps that estimator/renderer gap entirely.
			&contracts.TwoColumn{
				Ratio: contracts.Ratio11,
				Left: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{
						rt("What changed: a single deterministic renderer replaced the old pipeline."),
					}},
				},
				Right: []contracts.SlideNode{
					&contracts.List{Kind: contracts.ListChecklist, Items: []contracts.ListItem{
						{Text: rt("Byte-identical re-renders"), Checked: true},
						{Text: rt("Zero Chromium dependency"), Checked: true},
					}},
				},
			},
		},
	})
}

func quoteDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Quote", contracts.Slide{
		ID:        "quote",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Quote{
				Text:        rt("The best way to predict the future is to invent it."),
				Attribution: "Alan Kay",
			},
		},
	})
}

func flowDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Flow", contracts.Slide{
		ID:        "flow",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("How A Deck Ships")},
			&contracts.Flow{
				Orientation: contracts.FlowHorizontal,
				Connector:   contracts.ConnectorArrow,
				Steps: []contracts.FlowStep{
					{Label: rt("Draft"), Detail: rt("Agent composes the IR")},
					{Label: rt("Validate"), Detail: rt("Contrast + overflow checks")},
					{Label: rt("Render"), Detail: rt("Deterministic PPTX bytes")},
					{Label: rt("Export"), Detail: rt("Deck resource is ready")},
				},
			},
		},
	})
}

func comparisonTableDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Comparison Table", contracts.Slide{
		ID:        "comparison-table",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Plan Comparison")},
			&contracts.Table{
				Headers: []contracts.RichText{rt("Plan"), rt("Seats"), rt("Support"), rt("Price")},
				Rows: [][]contracts.RichText{
					{rt("Starter"), rt("5"), rt("Community"), rt("$0")},
					{rt("Team"), rt("25"), rt("Email"), rt("$49/mo")},
					{rt("Business"), rt("100"), rt("Priority"), rt("$199/mo")},
					{rt("Enterprise"), rt("Unlimited"), rt("Dedicated"), rt("Custom")},
				},
				Caption: "Effective FY26 Q3",
			},
		},
	})
}

func bentoDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Bento", contracts.Slide{
		ID:        "bento",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Highlights")},
			&contracts.Bento{
				Columns: 3,
				Rows: []contracts.BentoRow{
					{
						Label: "Core",
						Cells: []contracts.BentoCell{
							{Span: 2, Node: &contracts.Card{Header: "Deterministic export", Body: []contracts.SlideNode{
								&contracts.Prose{Paragraphs: []contracts.RichText{rt("Same doc + soul always renders identical bytes.")}},
							}}},
							{Span: 1, Node: &contracts.Chip{Label: "New", Tone: contracts.ChipSolid, Color: contracts.ColorAccent}},
						},
					},
					{
						Label: "Details",
						Cells: []contracts.BentoCell{
							{Span: 1, Node: &contracts.Card{Header: "Fast", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Sub-second.")}}}}},
							{Span: 1, Node: &contracts.Card{Header: "Safe", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("AA-checked.")}}}}},
							{Span: 1, Node: &contracts.Card{Header: "Portable", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("One binary.")}}}}},
						},
					},
				},
			},
		},
	})
}

func darkFeatureDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Dark Feature", contracts.Slide{
		ID:        "dark-feature",
		Archetype: contracts.ArchetypeDark,
		Variant:   contracts.VariantDark,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Built For Scale")},
			&contracts.Prose{Paragraphs: []contracts.RichText{
				rt("A single Go binary renders every deck — no headless browser, no measure-the-DOM stage."),
			}},
		},
	})
}

func closingDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Closing", contracts.Slide{
		ID:        "closing",
		Archetype: contracts.ArchetypeClosing,
		Variant:   contracts.VariantDark,
		Layout:    contracts.LayoutCover,
		Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Thank you"},
		},
	})
}

func dashboardDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Dashboard", contracts.Slide{
		ID:        "dashboard",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Operating Metrics")},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceSM,
				Cells: []contracts.SlideNode{
					&contracts.Stat{Value: "$2.2M", Label: "ARR", Delta: "+18%", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "99.98%", Label: "Uptime", Delta: "+0.1pp", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "1.8%", Label: "Churn", Delta: "-0.4pp", DeltaTone: contracts.DeltaDown},
				},
			},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceSM,
				Cells: []contracts.SlideNode{
					&contracts.Card{Header: "Reliability", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("99.98% uptime.")}}}},
					&contracts.Card{Header: "Speed", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("400ms median.")}}}},
					&contracts.Card{Header: "Scale", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("3x throughput.")}}}},
				},
			},
		},
	})
}

// watermarkContentDoc is a content slide carrying a DecorationText "03"
// background watermark plus a starfield scatter preset with Pitch and a
// neutral Color (R13.9 corpus accept case). Asset-free.
func watermarkContentDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Watermark Content", contracts.Slide{
		ID:        "watermark-content",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Decoration{
				Kind:  contracts.DecorationText,
				Text:  "03",
				Layer: contracts.LayerBackground,
			},
			&contracts.Decoration{
				Kind:   contracts.DecorationPreset,
				Preset: "starfield",
				Layer:  contracts.LayerBackground,
				Color:  contracts.ColorSurfaceAlt,
				Pitch:  24,
				Bleed:  true,
			},
			&contracts.Heading{Level: 2, Text: rt("Section Three")},
			&contracts.Prose{Paragraphs: []contracts.RichText{
				rt("A watermark and a starfield scatter both render behind body content."),
			}},
		},
	})
}

// focalCardDoc is a 3-card row where the center card carries a radial_glow
// Backdrop (R13.10 corpus accept case). Asset-free.
func focalCardDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Focal Card", contracts.Slide{
		ID:        "focal-card",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Featured Plan")},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Card{Header: "Starter", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("For small teams.")}},
					}},
					&contracts.Card{
						Header: "Business",
						Backdrop: &contracts.Decoration{
							Kind:   contracts.DecorationPreset,
							Preset: "radial_glow",
							Anchor: contracts.AnchorCenter,
							Bleed:  true,
						},
						Body: []contracts.SlideNode{
							&contracts.Prose{Paragraphs: []contracts.RichText{rt("Our most popular plan.")}},
						},
					},
					&contracts.Card{Header: "Enterprise", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("Custom scale.")}},
					}},
				},
			},
		},
	})
}

// SoulVariant is one named soul under test — "light/dark/cream" per R14.19:
// cream is the built-in Deckard White (flat, undecorated); brand and
// brandCream are both R13-D-bootstrapped (decor policy + paper tint), one on
// a white canvas and one on Deckard White's inherited warm/cream canvas, so
// the corpus exercises the decorated-render path against two distinct light
// bases. Slide-level "dark" is covered by the dark-feature/closing fixtures
// above, rendered through every soul variant here.
type SoulVariant struct {
	Name string
	Soul *soul.Soul
}

// Souls builds the three soul variants the corpus renders every archetype
// through. Deterministic — no I/O, no clock/rand.
func Souls() ([]SoulVariant, error) {
	brand, err := soul.Bootstrap(soul.BootstrapParams{
		Name:   "Conformance Brand",
		Accent: "3B5BDB",
		Palette: &soul.Palette{
			Surfaces: map[string]string{"canvas": "FFFFFF", "surfaceAlt": "E8ECFB"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("conformance: bootstrap brand soul: %w", err)
	}

	brandCream, err := soul.Bootstrap(soul.BootstrapParams{
		Name:   "Conformance Brand Cream",
		Accent: "3B5BDB",
	})
	if err != nil {
		return nil, fmt.Errorf("conformance: bootstrap brandCream soul: %w", err)
	}

	return []SoulVariant{
		{Name: "cream", Soul: soul.DeckardWhite()},
		{Name: "brand", Soul: brand},
		{Name: "brandCream", Soul: brandCream},
	}, nil
}
