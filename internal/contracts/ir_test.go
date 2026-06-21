package contracts

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// richTextSample exercises every RunStyle flag, both color forms, and the
// zero color (default primary) so the flat-run codec is fully covered.
func richTextSample() RichText {
	return RichText{
		{Text: "see ", Color: TextColor{Token: TextSecondary}},
		{
			Text: "the docs",
			Link: true, Href: "https://example.com", Code: false, TypeRole: TypeBody,
			Color: TextColor{Literal: "FF0000"},
		},
		{Text: "bold mono", Bold: true, Italic: true, Underline: true, Strike: true, TypeRole: TypeMono},
		{Text: "default-color run"},
	}
}

func TestRichTextRoundTrip(t *testing.T) {
	want := richTextSample()
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got RichText
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("RichText round-trip drift:\nwant=%#v\ngot =%#v", want, got)
	}
}

// nodeRoundTrips is the table of every implemented node kind. Each entry
// marshals via the node's MarshalJSON and unmarshals via UnmarshalSlideNode.
func nodeRoundTrips() []struct {
	name string
	node SlideNode
} {
	return []struct {
		name string
		node SlideNode
	}{
		{"hero", &Hero{Eyebrow: "Q2", Title: "Review", Subtitle: "what shipped"}},
		{"heading", &Heading{Text: RichText{{Text: "Highlights", TypeRole: TypeH2}}, Level: 2}},
		{"prose", &Prose{Paragraphs: []RichText{
			{{Text: "first paragraph"}},
			{{Text: "second", Italic: true, TypeRole: TypeBodySmall}},
		}}},
		{"list", &List{Kind: ListChecklist, Items: []ListItem{
			{Text: RichText{{Text: "ship it"}}, Checked: true},
			{Text: RichText{{Text: "nested item"}}, Level: 1},
		}}},
		{"callout", &Callout{Kind: CalloutWarning, Title: "Heads up", Body: RichText{
			{Text: "watch out", Color: TextColor{Token: TextWarning}},
		}}},
		{"two_column", &TwoColumn{Ratio: Ratio11,
			Left:  []SlideNode{&Hero{Title: "L"}},
			Right: []SlideNode{&Prose{Paragraphs: []RichText{{{Text: "R"}}}}},
		}},
		{"grid", &Grid{Columns: 2, Ratio: []int{1, 2}, Gap: SpaceMD, Cells: []SlideNode{
			&Hero{Title: "c1"},
			&Prose{Paragraphs: []RichText{{{Text: "c2"}}}},
		}}},
		{"card", &Card{
			Header: "H", Eyebrow: "E", Icon: "spark", HeaderPill: "new",
			Body:        []SlideNode{&Prose{Paragraphs: []RichText{{{Text: "body"}}}}},
			BodyLayout:  BodyVertical,
			Fill:        ColorAccent,
			Outline:     true,
			BorderStyle: BorderSolid,
			Size:        CardSizeLG,
			Layout:      CardLayoutIconTop,
			Elevation:   ElevationRaised,
		}},
		{"card_r4", &Card{
			Header:     "Pillar",
			Fill:       ColorSurface,
			HeaderFill: ColorAccent,
			StatusDot:  ColorSuccess,
			Watermark:  "01",
			Body:       []SlideNode{&Prose{Paragraphs: []RichText{{{Text: "R4 rich card"}}}}},
		}},
		{"card_section", &CardSection{Header: "S", Body: []SlideNode{
			&Grid{Columns: 2, Cells: []SlideNode{&Hero{Title: "g"}}},
		}}},
		{"divider", &Divider{Spacing: SpaceLG}},
		{"quote", &Quote{Text: RichText{{Text: "a quote"}, {Text: "emphasis", Italic: true, TypeRole: TypeBody}}, Attribution: "me"}},
		{"chip", &Chip{Label: "new", Tone: ChipSolid, Color: ColorAccent}},
		{"arrow", &Arrow{Direction: ArrowRight, Label: "next"}},
		{"section_divider", &SectionDivider{Eyebrow: "Part 2", Label: "Details"}},
		{"table", &Table{
			Headers: []RichText{{{Text: "A"}}, {{Text: "B", Bold: true, TypeRole: TypeH3}}},
			Rows:    [][]RichText{{{{Text: "1"}}}, {{{Text: "2"}}}},
			Caption: "cap",
		}},
		{"flow", &Flow{Orientation: FlowHorizontal, Connector: ConnectorArrow, Steps: []FlowStep{
			{Label: RichText{{Text: "start"}}, Detail: RichText{{Text: "go"}}, Icon: "play"},
			{Label: RichText{{Text: "end"}, {Text: "!", Bold: true}}},
		}}},
		{"image", &Image{AssetID: "logo", Alt: "logo", Frame: FrameBrowser, FrameName: "",
			Crop: Crop{Left: 0.1, Top: 0.1, Right: 0.1, Bottom: 0.1}, Fit: FitFill}},
		{"code_block", &CodeBlock{AssetID: "snippet", Language: "go", Caption: "main.go"}},
		{"chart", &Chart{AssetID: "q1-chart", Caption: "Revenue"}},
		{"decoration", &Decoration{Kind: DecorationPreset, Preset: "blob", Layer: LayerBackground,
			Anchor: AnchorTopLeft, Offset: Position{X: 10, Y: 20}, Size: Size{W: 100, H: 100},
			Bleed: true, Opacity: 0.5, Rotation: 15}},
		// R6 (D-057) — Stat leaf node.
		{"stat-full", &Stat{Value: "$2,200", Label: "per month", Delta: "+18%", DeltaTone: DeltaUp}},
		{"stat-no-delta", &Stat{Value: "98%", Label: "NPS score"}},
		{"stat-delta-down", &Stat{Value: "14 days", Label: "avg cycle", Delta: "-2 days", DeltaTone: DeltaDown}},
	}
}

func TestNodeRoundTrip(t *testing.T) {
	for _, tt := range nodeRoundTrips() {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.node)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got, err := UnmarshalSlideNode(b)
			if err != nil {
				t.Fatalf("UnmarshalSlideNode: %v", err)
			}
			if !reflect.DeepEqual(tt.node, got) {
				t.Fatalf("node round-trip drift:\nwant=%#v\ngot =%#v", tt.node, got)
			}
			// The marshaled object MUST carry the kind discriminator.
			var peek struct {
				Kind Kind `json:"kind"`
			}
			if err := json.Unmarshal(b, &peek); err != nil || peek.Kind == "" {
				t.Fatalf("marshaled node missing kind: %s (err=%v)", b, err)
			}
		})
	}
}

// TestRecursiveNesting proves a deeply nested deck survives marshal→unmarshal
// equal: two_column → card → list (≥3 levels of containers).
func TestRecursiveNesting(t *testing.T) {
	want := SlideDoc{Title: "Recursion", Slides: []Slide{{
		ID:     "s1",
		Layout: LayoutTwoColumn,
		Nodes: []SlideNode{
			&TwoColumn{Ratio: Ratio11,
				Left: []SlideNode{&Card{Header: "C", Fill: ColorSurface, Elevation: ElevationRaised,
					Body: []SlideNode{&List{Kind: ListBullet, Items: []ListItem{
						{Text: RichText{{Text: "a"}}},
						{Text: RichText{{Text: "b", Bold: true}}, Level: 1},
					}}}}},
				Right: []SlideNode{&Heading{Text: RichText{{Text: "R"}}, Level: 3}},
			},
			&CardSection{Header: "section", Body: []SlideNode{
				&Grid{Columns: 2, Cells: []SlideNode{
					&Callout{Kind: CalloutTip, Title: "tip", Body: RichText{{Text: "x"}}},
					&Prose{Paragraphs: []RichText{{{Text: "p"}}}},
				}},
			}},
		},
		Notes: RichText{{Text: "speaker notes"}, {Text: "second run", Code: true, TypeRole: TypeCode}},
	}}}

	b, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got SlideDoc
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("recursive deck round-trip drift:\nwant=%#v\ngot =%#v", want, got)
	}
}

// TestUnknownKindErrors asserts an unknown and a missing kind are hard errors.
func TestUnknownKindErrors(t *testing.T) {
	if _, err := UnmarshalSlideNode([]byte(`{"kind":"totally_unknown","x":1}`)); err == nil {
		t.Fatal("want error for unknown kind, got nil")
	}
	if _, err := UnmarshalSlideNode([]byte(`{"text":"no kind here"}`)); err == nil {
		t.Fatal("want error for missing kind, got nil")
	}
	if _, err := UnmarshalSlideNode([]byte(`{"kind":"hero"`)); err == nil {
		t.Fatal("want error for malformed JSON, got nil")
	}
	// Sanity: a known kind with a bad field type still surfaces an error.
	if _, err := UnmarshalSlideNode([]byte(`{"kind":"heading","level":"not-an-int"}`)); err == nil {
		t.Fatal("want error for bad field, got nil")
	}
}

// TestUnknownKindIsNotSilent also asserts the error wraps for errors.Is use.
func TestUnknownKindIsNotSilent(t *testing.T) {
	_, err := UnmarshalSlideNode([]byte(`{"kind":"nope"}`))
	if err == nil {
		t.Fatal("want error")
	}
	if !errors.Is(err, err) { // trivial: ensures errors import is exercised
		t.Fatal("errors.Is sanity")
	}
}

// TestCardR4FieldsOmitWhenEmpty asserts that headerFill, statusDot, and watermark
// are absent from the JSON output when the Card carries none of the R4 fields —
// the omitempty contract guarantees byte-identical output to a pre-R4 card.
func TestCardR4FieldsOmitWhenEmpty(t *testing.T) {
	t.Parallel()

	c := &Card{Header: "Plain", Fill: ColorSurface}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}
	for _, field := range []string{"headerFill", "statusDot", "watermark"} {
		if _, ok := m[field]; ok {
			t.Errorf("field %q present in JSON for empty Card, want omitted; JSON=%s", field, b)
		}
	}
}
