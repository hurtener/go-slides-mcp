package contracts

// ExampleNodeForKind returns a populated, schema-valid SlideNode for the given
// kind. Every example is built from a real struct so it always round-trips
// through UnmarshalSlideNode without dropping any field. Returns (nil, false)
// for an unknown kind.
//
// The examples deliberately exercise the fields that agents most often get
// wrong: flow steps use label+detail (NOT title/body); callout uses
// calloutKind (NOT kind); list uses listKind (NOT kind); a run is flat
// {text,bold} (NOT a nested {text,style:{bold:true}}).
func ExampleNodeForKind(kind Kind) (SlideNode, bool) {
	switch kind {
	case KindHero:
		return &Hero{
			Eyebrow:  "Q2 2026",
			Title:    "Results Overview",
			Subtitle: "What shipped and what is next",
		}, true

	case KindHeading:
		return &Heading{
			Text:  RichText{{Text: "Key Findings"}},
			Level: 2,
		}, true

	case KindProse:
		return &Prose{
			Paragraphs: []RichText{
				{
					{Text: "This quarter we shipped "},
					{Text: "three major features", Bold: true},
				},
			},
		}, true

	case KindList:
		return &List{
			Kind: ListBullet,
			Items: []ListItem{
				{Text: RichText{{Text: "Shipped the new dashboard"}}},
				{Text: RichText{
					{Text: "Reduced latency by "},
					{Text: "38%", Bold: true},
				}},
				{Text: RichText{{Text: "Onboarded 12 new customers"}}},
			},
		}, true

	case KindCallout:
		// calloutKind is the variant field; "kind" is always the node discriminator.
		return &Callout{
			Kind:  CalloutNote,
			Title: "Important",
			Body:  RichText{{Text: "Review these findings before the board meeting."}},
		}, true

	case KindTwoColumn:
		// Join and JoinLabel are additive (D-055): a "badge" join with label "VS"
		// draws a circular badge centered on the column seam. Omit both fields for
		// a plain two-column with no seam element (byte-identical to pre-R5 output).
		return &TwoColumn{
			Ratio:     Ratio11,
			Join:      JoinBadge,
			JoinLabel: "VS",
			Left:      []SlideNode{&Heading{Text: RichText{{Text: "Option A"}}, Level: 2}},
			Right:     []SlideNode{&Heading{Text: RichText{{Text: "Option B"}}, Level: 2}},
		}, true

	case KindGrid:
		return &Grid{
			Columns: 2,
			Gap:     SpaceMD,
			Cells: []SlideNode{
				&Card{Header: "Card A", Body: []SlideNode{
					&Prose{Paragraphs: []RichText{{{Text: "First cell"}}}},
				}},
				&Card{Header: "Card B", Body: []SlideNode{
					&Prose{Paragraphs: []RichText{{{Text: "Second cell"}}}},
				}},
			},
		}, true

	case KindCard:
		return &Card{
			Header: "Feature Highlight",
			Body: []SlideNode{
				&Prose{Paragraphs: []RichText{
					{{Text: "Cards hold child slide nodes in the body field."}},
				}},
			},
		}, true

	case KindCardSection:
		return &CardSection{
			Header: "Section Title",
			Body: []SlideNode{
				&Prose{Paragraphs: []RichText{{{Text: "Section body content."}}}},
			},
		}, true

	case KindDivider:
		return &Divider{Spacing: SpaceMD}, true

	case KindQuote:
		return &Quote{
			Text:        RichText{{Text: "The best way to predict the future is to invent it."}},
			Attribution: "Alan Kay",
		}, true

	case KindChip:
		return &Chip{Label: "New", Tone: ChipSolid, Color: ColorAccent}, true

	case KindArrow:
		return &Arrow{Direction: ArrowRight, Label: "next step"}, true

	case KindSectionDivider:
		return &SectionDivider{Eyebrow: "Part 2", Label: "Deep Dive"}, true

	case KindTable:
		return &Table{
			Headers: []RichText{
				{{Text: "Metric"}},
				{{Text: "Value", Bold: true}},
			},
			Rows: [][]RichText{
				{{{Text: "Revenue"}}, {{Text: "$2.4M"}}},
				{{{Text: "Users"}}, {{Text: "12,400"}}},
			},
			Caption: "Q2 summary",
		}, true

	case KindFlow:
		// Steps use label (RichText) + detail (RichText). Do NOT use title/body.
		return &Flow{
			Orientation: FlowHorizontal,
			Connector:   ConnectorArrow,
			Steps: []FlowStep{
				{
					Label:  RichText{{Text: "Discover"}},
					Detail: RichText{{Text: "Identify the problem"}},
				},
				{
					Label:  RichText{{Text: "Design"}},
					Detail: RichText{{Text: "Prototype solutions"}},
				},
				{
					Label:  RichText{{Text: "Deliver"}},
					Detail: RichText{{Text: "Ship and measure"}},
				},
			},
		}, true

	case KindImage:
		return &Image{
			AssetID: "brand-logo",
			Alt:     "Brand logo",
			Fit:     FitFill,
		}, true

	case KindCodeBlock:
		return &CodeBlock{
			AssetID:  "snippet-001",
			Language: "go",
			Caption:  "main.go",
		}, true

	case KindChart:
		return &Chart{
			AssetID: "revenue-chart",
			Caption: "Q2 Revenue",
		}, true

	case KindDecoration:
		return &Decoration{
			Kind:   DecorationPreset,
			Preset: "blob",
			Layer:  LayerBackground,
		}, true

	case KindStat:
		// A single stat: big metric, supporting label, directional delta. A
		// grid of stats forms a metric/pricing strip (D-057).
		return &Stat{
			Value:     "$2,200",
			Label:     "per month",
			Delta:     "+18%",
			DeltaTone: DeltaUp,
		}, true

	case KindBento:
		// Bento (D-056): a row-labeled grid with variable column spans. Columns
		// sets the shared column-unit count; each row's cell spans must sum to
		// <= Columns. An optional Label on a row reserves a left-gutter label
		// column for all rows. Cells hold any child SlideNode.
		return &Bento{
			Columns: 3,
			Rows: []BentoRow{
				{
					Label: "Core",
					Cells: []BentoCell{
						{Span: 2, Node: &Prose{Paragraphs: []RichText{{{Text: "Primary feature spans two columns."}}}}},
						{Span: 1, Node: &Chip{Label: "New", Tone: ChipSolid, Color: ColorAccent}},
					},
				},
				{
					Label: "Details",
					Cells: []BentoCell{
						{Span: 1, Node: &Prose{Paragraphs: []RichText{{{Text: "Detail A."}}}}},
						{Span: 1, Node: &Prose{Paragraphs: []RichText{{{Text: "Detail B."}}}}},
						{Span: 1, Node: &Prose{Paragraphs: []RichText{{{Text: "Detail C."}}}}},
					},
				},
			},
		}, true

	case KindTimeline:
		// Timeline (R14.4, D-119): Milestones drives markers when Lanes is
		// empty. Every Milestone field is set (position/label/detail/icon/
		// accentIndex) plus one Band, so the round-trip covers every field.
		return &Timeline{
			Milestones: []Milestone{
				{Position: 0, Label: "Kickoff", Detail: "Scope locked", Icon: "star", AccentIndex: 0},
				{Position: 0.5, Label: "Beta", Detail: "First external users", Icon: "diamond", AccentIndex: 1},
				{Position: 1, Label: "GA", Detail: "General availability", Icon: "check", AccentIndex: 2},
			},
			Bands: []TimelineBand{
				{From: 0, To: 0.5, Label: "Phase 1", Fill: ColorSurfaceAlt},
			},
		}, true

	case KindDataMark:
		// DataMark (R14.8, D-122): Kind "bars" so both Value and Values are
		// meaningful in the same example (Value is used by bar/donut/gauge;
		// Values by bars/sparkline), covering every field on round-trip.
		return &DataMark{
			Kind:        DataMarkBars,
			Value:       0.6,
			Values:      []float64{0.2, 0.6, 0.9},
			Orientation: FlowHorizontal,
			Color:       ColorAccent,
			Label:       "Q3",
		}, true

	case KindQuadrant:
		// Quadrant (R14.9, D-124): both axes labeled, all 4 quadrants titled
		// + tinted, and 2 items with every field, so the round-trip covers
		// the fixed [4] array and every sub-struct field.
		return &Quadrant{
			AxisX: QuadrantAxis{LowLabel: "Low Effort", HighLabel: "High Effort"},
			AxisY: QuadrantAxis{LowLabel: "Low Impact", HighLabel: "High Impact"},
			Quadrants: [4]QuadrantCell{
				{Title: "Quick Wins", Fill: ColorSurfaceAlt},
				{Title: "Big Bets", Fill: ColorAccentAlt},
				{Title: "Fill-Ins", Fill: ColorSurface},
				{Title: "Money Pits", Fill: ColorAccentWarm},
			},
			Items: []QuadrantItem{
				{X: 0.2, Y: 0.8, Label: "Onboarding revamp", AccentIndex: 0},
				{X: 0.8, Y: 0.9, Label: "Platform rebuild", AccentIndex: 1},
			},
		}, true

	case KindTree:
		// Tree (R14.10, D-127): a root with 2 children, one nested a level
		// deeper, covering Label/Detail/Icon/AccentIndex/Children at every
		// depth on round-trip.
		return &Tree{
			Root: TreeNode{
				Label:       "CEO",
				Detail:      "Executive lead",
				Icon:        "star",
				AccentIndex: 0,
				Children: []TreeNode{
					{
						Label:       "VP Engineering",
						Detail:      "Platform + product",
						Icon:        "diamond",
						AccentIndex: 1,
						Children: []TreeNode{
							{Label: "Eng Manager", Detail: "Core platform team", Icon: "check", AccentIndex: 2},
						},
					},
					{Label: "VP Sales", Detail: "Revenue + partnerships", Icon: "square", AccentIndex: 2},
				},
			},
			Orientation: FlowVertical,
		}, true

	case KindFunnel:
		// Funnel (R14.11, D-128): 3 stages, each with a Label/Value/
		// AccentIndex, covering every field on round-trip.
		return &Funnel{
			Stages: []FunnelStage{
				{Label: "Visitors", Value: "10,000", AccentIndex: 0},
				{Label: "Signups", Value: "2,400", AccentIndex: 1},
				{Label: "Customers", Value: "380", AccentIndex: 2},
			},
		}, true

	case KindCycle:
		// Cycle (R14.11, D-128): 4 stages, each with a Label/Icon/
		// AccentIndex, covering every field on round-trip.
		return &Cycle{
			Stages: []CycleStage{
				{Label: "Plan", Icon: "star", AccentIndex: 0},
				{Label: "Build", Icon: "diamond", AccentIndex: 1},
				{Label: "Ship", Icon: "check", AccentIndex: 2},
				{Label: "Learn", Icon: "circle", AccentIndex: 0},
			},
		}, true

	default:
		return nil, false
	}
}
