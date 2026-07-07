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
// corpus.go's package doc — org-chart/quadrant/funnel/agenda/logo-wall/RTL
// nodes are not built yet and are NOT stubbed in here; timeline landed in
// R14.4). Every entry uses only native, asset-free nodes so a fixture render
// never depends on an AssetResolver.
var Archetypes = []Fixture{
	{"cover", coverDoc},
	{"section", sectionDoc},
	{"content", contentDoc},
	{"card-grid", cardGridDoc},
	{"stat-strip", statStripDoc},
	{"two-column", twoColumnDoc},
	{"quote", quoteDoc},
	{"quote-testimonial", quoteTestimonialDoc},
	{"flow", flowDoc},
	{"comparison-table", comparisonTableDoc},
	{"bento", bentoDoc},
	{"dark-feature", darkFeatureDoc},
	{"closing", closingDoc},
	{"dashboard", dashboardDoc},
	{"cover-mesh", coverMeshDoc},
	{"watermark-content", watermarkContentDoc},
	{"focal-card", focalCardDoc},
	{"scrim-cover", scrimCoverDoc},
	{"timeline", timelineDoc},
	{"datamark-kpi", dataMarkKPIDoc},
	{"quadrant", quadrantDoc},
	{"org-tree", treeDoc},
	{"funnel", funnelDoc},
	{"cycle", cycleDoc},
	{"footnotes-sources", footnotesSourcesDoc},
	{"agenda", agendaDoc},
	{"button-cta", buttonCTADoc},
	{"chip-row", chipRowDoc},
	{"checklist", checklistDoc},
	{"banner", bannerDoc},
	{"icon-rows", iconRowsDoc},
	{"lockup", lockupDoc},
	{"connector-grid", connectorGridDoc},
	{"bridge-two-column", bridgeTwoColumnDoc},
	{"ribbon-card", ribbonCardDoc},
	{"pricing-offer-card", pricingOfferCardDoc},
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

// quoteTestimonialDoc exercises the Quote testimonial enrichment (R14.5,
// D-120): Mark + structured attribution (Name/Role/Company), asset-free (no
// avatar/logo AssetID — the corpus stays asset-free per its scope note).
func quoteTestimonialDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Quote Testimonial", contracts.Slide{
		ID:        "quote-testimonial",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Quote{
				Text:               rt("Deckard cut our deck-build time from days to minutes."),
				Mark:               true,
				AttributionName:    "Priya Natarajan",
				AttributionRole:    "VP Marketing",
				AttributionCompany: "Northwind Labs",
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

// scrimCoverDoc is a cover slide whose Background is a solid color with a
// gradient Scrim overlay (R14.1) — the corpus's product-level accept case
// for Background.Scrim over a color background (no asset needed: a scrim
// applies over any drawn background kind). Asset-free like every other
// archetype fixture in this file.
func scrimCoverDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Scrim Cover", contracts.Slide{
		ID:        "scrim-cover",
		Archetype: contracts.ArchetypeCover,
		Layout:    contracts.LayoutCover,
		Background: &contracts.Background{
			Kind:  contracts.BackgroundColor,
			Color: contracts.ColorAccent,
			Scrim: &contracts.Scrim{
				Color:    contracts.ColorCanvas,
				Opacity:  0.5,
				Gradient: true,
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

// timelineDoc exercises the Timeline node (R14.4, D-119): a single-lane
// roadmap of 4 milestones overlaid with 2 phase bands. Kept modest (4
// milestones, staggered labels only) so it stays clear of the safe-area
// (INV-2 zero-overflow, strict).
func timelineDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Timeline", contracts.Slide{
		ID:        "timeline",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
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
	})
}

// dataMarkKPIDoc exercises the DataMark node (R14.8, D-122): a 3-card KPI
// row, one DataMark per card body — a donut, a bar, and a bar group. Kept to
// one native-geometry mark per card (mirrors cardGridDoc's one-node-per-cell
// shape) so it stays clear of the safe area (INV-2 zero-overflow, strict).
func dataMarkKPIDoc() contracts.SlideDoc {
	return oneSlide("Conformance — DataMark KPI", contracts.Slide{
		ID:        "datamark-kpi",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("This Quarter's Marks")},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Card{Header: "Uptime", Body: []contracts.SlideNode{
						&contracts.DataMark{Kind: contracts.DataMarkDonut, Value: 0.92, Label: "92%"},
					}},
					&contracts.Card{Header: "Capacity", Body: []contracts.SlideNode{
						&contracts.DataMark{Kind: contracts.DataMarkBar, Value: 0.6, Label: "60%"},
					}},
					&contracts.Card{Header: "Trend", Body: []contracts.SlideNode{
						&contracts.DataMark{Kind: contracts.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5}},
					}},
				},
			},
		},
	})
}

// quadrantDoc exercises the Quadrant node (R14.9, D-124): a labeled 2x2
// prioritization map with all 4 quadrants titled + tinted and 5 plotted
// items — a modest count that stays clear of the safe area (INV-2
// zero-overflow, strict).
func quadrantDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Quadrant", contracts.Slide{
		ID:        "quadrant",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
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
	})
}

// treeDoc exercises the Tree node (R14.10, D-127): a shallow org chart —
// root -> 3 children, one with 2 leaves — kept to modest depth/breadth so
// the layout stays clear of the safe area (INV-2 zero-overflow, strict).
func treeDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Org Tree", contracts.Slide{
		ID:        "org-tree",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
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
	})
}

// funnelDoc exercises the Funnel node (R14.11, D-128): a 4-stage marketing
// conversion funnel with a value caption on every stage — a modest count
// that stays clear of the safe area (INV-2 zero-overflow, strict).
func funnelDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Funnel", contracts.Slide{
		ID:        "funnel",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
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
	})
}

// cycleDoc exercises the Cycle node (R14.11, D-128): a 5-stage lifecycle
// loop with a curated icon on every stage — a modest count that stays clear
// of the safe area (INV-2 zero-overflow, strict).
func cycleDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Cycle", contracts.Slide{
		ID:        "cycle",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Product Lifecycle")},
			&contracts.Cycle{
				Stages: []contracts.CycleStage{
					{Label: "Discover", Icon: "star", AccentIndex: 0},
					{Label: "Plan", Icon: "diamond", AccentIndex: 1},
					{Label: "Build", Icon: "square", AccentIndex: 2},
					{Label: "Ship", Icon: "check", AccentIndex: 0},
					{Label: "Learn", Icon: "circle", AccentIndex: 1},
				},
			},
		},
	})
}

// footnotesSourcesDoc exercises slide-level Footnotes + the Superscript run
// style (R14.12): two source/disclaimer lines pinned to the reserved bottom
// band, plus a superscript marker run referencing them from a stat callout.
// The footnote band shrinks the body region, so the body here stays modest
// (a heading + a 2-up stat row + one short caption) to hold clear of the
// safe area (INV-2 zero-overflow, strict). Asset-free like every other
// archetype fixture in this file.
func footnotesSourcesDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Footnotes & Sources", contracts.Slide{
		ID:        "footnotes-sources",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Headline Results")},
			&contracts.Grid{
				Columns: 2,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Stat{Value: "$2.2M", Label: "ARR", Delta: "+18%", DeltaTone: contracts.DeltaUp},
					&contracts.Stat{Value: "99.98%", Label: "Uptime", Delta: "+0.1pp", DeltaTone: contracts.DeltaUp},
				},
			},
			&contracts.Prose{Paragraphs: []contracts.RichText{
				{
					{Text: "ARR figure includes one-time items"},
					{Text: "1", Superscript: true},
					{Text: "."},
				},
			}},
		},
		Footnotes: []contracts.RichText{
			rt("Source: internal telemetry, 2026."),
			rt("Note: figures unaudited; final numbers pending Q3 close."),
		},
	})
}

// agendaDoc mirrors the rcp_agenda builtin recipe (internal/recipe/store.go,
// R14.6): a card_grid of 4 numbered Cards forming a section index. Kept to 4
// cells across a 4-column Grid — modest and asset-free, like every other
// archetype fixture in this file — so it stays clear of the safe area
// (INV-2 zero-overflow, strict).
func agendaDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Agenda", contracts.Slide{
		ID:        "agenda",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Agenda")},
			&contracts.Grid{
				Columns: 4,
				Gap:     contracts.SpaceMD,
				Cells: []contracts.SlideNode{
					&contracts.Card{Eyebrow: "01", Header: "Context", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("Where we are today")}},
					}},
					&contracts.Card{Eyebrow: "02", Header: "Strategy", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("Where we're headed")}},
					}},
					&contracts.Card{Eyebrow: "03", Header: "Roadmap", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("How we get there")}},
					}},
					&contracts.Card{Eyebrow: "04", Header: "Ask", Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{rt("What we need from you")}},
					}},
				},
			},
		},
	})
}

// buttonCTADoc exercises the Button CTA primitive (R12.1, D-094) as a
// standalone affordance on a content slide: a large primary Button with a
// trailing arrow-right glyph below a short heading. Kept to two modest nodes
// so it stays clear of the safe area (INV-2 zero-overflow, strict);
// asset-free like every other archetype fixture in this file.
func buttonCTADoc() contracts.SlideDoc {
	return oneSlide("Conformance — Button CTA", contracts.Slide{
		ID:        "button-cta",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Ready to ship?")},
			&contracts.Button{
				Label:        "Talk to the team",
				Tone:         contracts.ButtonPrimary,
				Size:         contracts.ButtonSizeLG,
				TrailingIcon: "arrow-right",
				Align:        contracts.HAlignLeft,
			},
		},
	})
}

// chipRowDoc exercises the ChipRow wrap-to-next-line chip group (R12.5,
// D-096): a labeled capability strip of 4 chips across all three ChipTone
// variants, two with curated leading icons, Wrap on. Asset-free + modest so
// it stays clear of the safe area (INV-2 zero-overflow, strict).
func chipRowDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Chip Row", contracts.Slide{
		ID:        "chip-row",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Common Builds")},
			&contracts.ChipRow{
				Label: "CATEGORIES",
				Chips: []contracts.ChipSpec{
					{Label: "Finance", Tone: contracts.ChipTint, Color: contracts.ColorAccent},
					{Label: "HR", Tone: contracts.ChipSolid, Color: contracts.ColorAccent, Icon: "check"},
					{Label: "Sales", Tone: contracts.ChipOutline, Color: contracts.ColorAccentAlt, Icon: "star"},
					{Label: "Operations", Tone: contracts.ChipTint, Color: contracts.ColorSurfaceAlt},
				},
				Wrap:  true,
				Align: contracts.HAlignLeft,
			},
		},
	})
}

// checklistDoc exercises the Checklist dense feature list (R12.2, D-095):
// 4 rows (each a different CheckState — done/no/neutral/done-default),
// 2-column row-major reflow, an accent GlyphTone, Fill on. Asset-free so
// no resolver path runs (the placeholder glyph is native curGeom).
// Kept modest so it stays clear of the safe area (INV-2 strict).
func checklistDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Checklist", contracts.Slide{
		ID:        "checklist",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("What You Get")},
			&contracts.Checklist{
				Items: []contracts.ChecklistItem{
					{Text: contracts.RichText{{Text: "Byte-identical exports"}}, State: contracts.CheckDone, Icon: "check"},
					{Text: contracts.RichText{{Text: "No headless browser"}}, State: contracts.CheckDone},
					{Text: contracts.RichText{{Text: "Soul-driven theming"}}, State: contracts.CheckDone, Icon: "star"},
					{Text: contracts.RichText{{Text: "Single binary"}}, State: contracts.CheckNeutral, Icon: "dot"},
				},
				Columns:   2,
				GlyphTone: contracts.ColorAccent,
				Fill:      true,
			},
		},
	})
}

// bannerDoc exercises the Banner full-width filled strip (R12.6, D-097):
// a Lead + Body, a leading curated Icon, an explicit accent Fill, a
// custom TextColor, and one Trailing Button child. Asset-free + modest so
// it stays clear of the safe area (INV-2 strict).
func bannerDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Banner", contracts.Slide{
		ID:        "banner",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("The Pitch")},
			&contracts.Banner{
				Lead:      contracts.RichText{{Text: "Run it internally"}},
				Body:      contracts.RichText{{Text: "Or sell it externally — without building an agentic platform."}},
				Icon:      "star",
				Fill:      contracts.ColorAccent,
				TextColor: contracts.TextInverse,
				Trailing: []contracts.SlideNode{
					&contracts.Button{
						Label:        "Start free",
						Tone:         contracts.ButtonGhost,
						TrailingIcon: "arrow-right",
					},
				},
			},
		},
	})
}

// iconRowsDoc exercises the IconRows vertical icon-label row list (R12.7,
// D-100): 4 rows across RowPlain and RowPill tones, every row has a
// leading curated Icon + label, one carries right-aligned Meta; Fill on.
// Asset-free + modest so it stays clear of the safe area (INV-2 strict).
func iconRowsDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Icon Rows", contracts.Slide{
		ID:        "icon-rows",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Integrations")},
			&contracts.IconRows{
				Rows: []contracts.IconRow{
					{Icon: "check", Label: contracts.RichText{{Text: "Chat & Q&A"}}, Tone: contracts.RowPlain},
					{Icon: "star", Label: contracts.RichText{{Text: "Salesforce · Slack"}}, Tone: contracts.RowPill},
					{Icon: "diamond", Label: contracts.RichText{{Text: "Microsoft 365"}}, Meta: contracts.RichText{{Text: "12 sources"}}, Tone: contracts.RowPlain},
					{Icon: "circle", Label: contracts.RichText{{Text: "Workspace"}}, Tone: contracts.RowPill},
				},
				Fill:       true,
				GlyphColor: contracts.ColorAccent,
			},
		},
	})
}

// lockupDoc exercises the Lockup attribution mark (R12.9, D-102) on the
// ICON path (media-free): a centered "POWERED BY" caption paired with a
// curated star glyph, logo trailing the caption. Asset-free so the corpus
// keeps its no-resolver invariant; a separate render test exercises the
// AssetID path with a stub resolver. Modest so it stays clear of the safe
// area (INV-2 strict).
func lockupDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Lockup", contracts.Slide{
		ID:        "lockup",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Partnered With")},
			&contracts.Lockup{
				Caption:   "POWERED BY",
				Icon:      "star",
				AssetSide: contracts.LeadCaption,
				MaxHeight: 18,
				Align:     contracts.HAlignCenter,
			},
		},
	})
}

// connectorGridDoc exercises Grid.Connectors (R12.4, D-099): a 3-column
// architecture row with two gutter connectors, one bi_arrow and one arrow.
// Modest node count so it stays clear of the safe area (INV-2 strict).
func connectorGridDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Connector Grid", contracts.Slide{
		ID:        "connector-grid",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Architecture")},
			&contracts.Grid{
				Columns: 3,
				Gap:     contracts.SpaceMD,
				Connectors: []contracts.GridConnector{
					{Between: [2]int{0, 1}, Kind: contracts.ConnectorBiArrow, Label: "sync"},
					{Between: [2]int{1, 2}, Kind: contracts.ConnectorArrow, Label: "ship"},
				},
				Cells: []contracts.SlideNode{
					&contracts.Card{Header: "People", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Users + operators")}}}},
					&contracts.Card{Header: "Agents", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Planning + execution")}}}},
					&contracts.Card{Header: "Data", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Knowledge + memory")}}}},
				},
			},
		},
	})
}

// bridgeTwoColumnDoc exercises TwoColumn.JoinPosition (R12.8, D-101): a
// badge join labeled "One agent" spanning the top of both columns.
func bridgeTwoColumnDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Bridge Two Column", contracts.Slide{
		ID:        "bridge-two-column",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutTwoColumn,
		Nodes: []contracts.SlideNode{
			&contracts.TwoColumn{
				Ratio:        contracts.Ratio11,
				Join:         contracts.JoinBadge,
				JoinLabel:    "One agent",
				JoinPosition: contracts.JoinTopBridge,
				Left: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{rt("Build internally")}},
				},
				Right: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{rt("Sell externally")}},
				},
			},
		},
	})
}

// ribbonCardDoc exercises Card.Ribbon (R12.3, D-098): the center card in a
// 3-up row carries a top-bar ribbon highlighting it as the recommended plan.
func ribbonCardDoc() contracts.SlideDoc {
	return oneSlide("Conformance — Ribbon Card", contracts.Slide{
		ID:        "ribbon-card",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Plans")},
			&contracts.Grid{Columns: 3, Gap: contracts.SpaceMD, Cells: []contracts.SlideNode{
				&contracts.Card{Header: "Starter", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("For small teams")}}}},
				&contracts.Card{Header: "Business", Ribbon: &contracts.Ribbon{Text: "MOST POPULAR", Position: contracts.RibbonTopBar, Color: contracts.ColorAccent, TextColor: contracts.TextInverse}, Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Our recommended plan")}}}},
				&contracts.Card{Header: "Enterprise", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{rt("Custom scale")}}}},
			}},
		},
	})
}

// pricingOfferCardDoc exercises the enriched pricing/offer-card recipe
// shape (R12.10): one highlighted tier + one plain tier, each card =
// Stat(price) + cap line + filled Checklist + footer CTA Button. The
// full built-in recipe carries 4 tiers (apply_recipe/large.json points to
// it); the corpus fixture stays to 2 cells so INV-2 strict zero-overflow
// holds while still proving the composite on-bar.
func pricingOfferCardDoc() contracts.SlideDoc {
	numPtr := func(v float64) *float64 { return &v }
	featureItems := func(items ...string) []contracts.ChecklistItem {
		out := make([]contracts.ChecklistItem, 0, len(items))
		for _, item := range items {
			out = append(out, contracts.ChecklistItem{Text: rt(item), State: contracts.CheckDone, Icon: "check"})
		}
		return out
	}
	tier := func(plan string, price float64, capLine string, highlight bool, cta string, features ...string) *contracts.Card {
		card := &contracts.Card{
			Eyebrow: "PLAN",
			Header:  plan,
			Body: []contracts.SlideNode{
				&contracts.Stat{Label: "per seat / month", Number: numPtr(price), Format: &contracts.NumberFormat{CurrencySymbol: "$"}},
				&contracts.Prose{Paragraphs: []contracts.RichText{rt(capLine)}},
				&contracts.Checklist{Items: featureItems(features...), Fill: true},
				&contracts.Button{Label: cta, Tone: contracts.ButtonPrimary, TrailingIcon: "arrow-right"},
			},
		}
		if highlight {
			card.HeaderFill = contracts.ColorAccent
			card.Ribbon = &contracts.Ribbon{Text: "MOST POPULAR", Position: contracts.RibbonTopBar, Color: contracts.ColorAccent, TextColor: contracts.TextInverse}
		}
		return card
	}
	return oneSlide("Conformance — Pricing Offer Card", contracts.Slide{
		ID:        "pricing-offer-card",
		Archetype: contracts.ArchetypeContent,
		Layout:    contracts.LayoutCardGrid,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: rt("Pricing")},
			&contracts.Grid{Columns: 2, Gap: contracts.SpaceMD, Cells: []contracts.SlideNode{
				tier("Growth", 79, "up to 25 agents", true, "Start free", "Priority", "Analytics"),
				tier("Scale", 199, "up to 100 agents", false, "Contact sales", "Dedicated", "SLA"),
			}},
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
