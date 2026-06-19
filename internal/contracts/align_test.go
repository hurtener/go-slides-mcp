package contracts

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestAlignmentRoundTrip asserts Alignment marshals/unmarshals cleanly,
// including the zero value (omitempty fields absent from JSON).
func TestAlignmentRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		align Alignment
		// wantEmpty: if true, the JSON should be "{}" (zero-value = omitempty).
		wantEmpty bool
	}{
		{
			name:      "zero value omits both fields",
			align:     Alignment{},
			wantEmpty: true,
		},
		{
			name:  "center-center",
			align: Alignment{Vertical: VAlignCenter, Horizontal: HAlignCenter},
		},
		{
			name:  "bottom-right",
			align: Alignment{Vertical: VAlignBottom, Horizontal: HAlignRight},
		},
		{
			name:  "justify only",
			align: Alignment{Vertical: VAlignJustify},
		},
		{
			name:  "horizontal only",
			align: Alignment{Horizontal: HAlignCenter},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.align)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got Alignment
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tc.align, got) {
				t.Fatalf("Alignment round-trip drift:\nwant=%#v\ngot =%#v", tc.align, got)
			}
			if tc.wantEmpty && string(b) != "{}" {
				t.Fatalf("zero Alignment should marshal to {}, got %s", b)
			}
		})
	}
}

// TestSlideAlignRoundTrip asserts a Slide with Align survives the full
// marshal→unmarshal cycle (including the custom Slide codec that handles Nodes
// and Notes).
func TestSlideAlignRoundTrip(t *testing.T) {
	t.Parallel()

	want := Slide{
		ID:    "cover",
		Align: Alignment{Vertical: VAlignCenter, Horizontal: HAlignCenter},
		Nodes: []SlideNode{
			&Hero{Title: "Hello", Eyebrow: "World"},
		},
	}
	b, err := json.Marshal(&want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Slide
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Align != want.Align {
		t.Fatalf("Slide.Align round-trip drift: want %#v got %#v", want.Align, got.Align)
	}
	if len(got.Nodes) != 1 {
		t.Fatalf("nodes len: want 1 got %d", len(got.Nodes))
	}
}

// TestNodeAlignRoundTrip asserts per-node Align fields survive marshal→unmarshal
// for each node type that carries the field.
func TestNodeAlignRoundTrip(t *testing.T) {
	t.Parallel()

	nodes := []SlideNode{
		&Hero{Title: "T", Align: HAlignCenter},
		&Heading{Level: 2, Text: RichText{{Text: "H"}}, Align: HAlignRight},
		&Prose{Paragraphs: []RichText{{{Text: "P"}}}, Align: HAlignCenter},
		&Quote{Text: RichText{{Text: "Q"}}, Attribution: "A", Align: HAlignRight},
		&Chip{Label: "C", Tone: ChipTint, Align: HAlignCenter},
		&SectionDivider{Label: "S", Align: HAlignRight},
	}

	for _, n := range nodes {
		b, err := json.Marshal(n)
		if err != nil {
			t.Fatalf("marshal %T: %v", n, err)
		}
		got, err := UnmarshalSlideNode(b)
		if err != nil {
			t.Fatalf("UnmarshalSlideNode %T: %v", n, err)
		}
		if !reflect.DeepEqual(n, got) {
			t.Fatalf("node align round-trip drift for %T:\nwant=%#v\ngot =%#v", n, n, got)
		}
	}
}
