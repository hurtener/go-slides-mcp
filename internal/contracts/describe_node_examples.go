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
		return &TwoColumn{
			Ratio: Ratio11,
			Left:  []SlideNode{&Heading{Text: RichText{{Text: "Left Column"}}, Level: 2}},
			Right: []SlideNode{&Prose{Paragraphs: []RichText{{{Text: "Right column content."}}}}},
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

	default:
		return nil, false
	}
}
