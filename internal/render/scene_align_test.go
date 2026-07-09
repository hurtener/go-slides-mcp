package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapSlideAlignment asserts that a Slide with Align:{Vertical:"center"}
// maps to a SceneSlide with Content.Vertical == scene.VAlignCenter.
func TestMapSlideAlignment(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{
		ID:    "cover",
		Align: contracts.Alignment{Vertical: contracts.VAlignCenter, Horizontal: contracts.HAlignCenter},
		Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Hello"}},
	}
	got := mapSlide(slide, 0)

	if got.Content.Vertical != scene.VAlignCenter {
		t.Errorf("Content.Vertical = %v, want %v", got.Content.Vertical, scene.VAlignCenter)
	}
	if got.Content.Horizontal != scene.HAlignCenter {
		t.Errorf("Content.Horizontal = %v, want %v", got.Content.Horizontal, scene.HAlignCenter)
	}
}

// TestMapSlideZeroAlignment asserts that a Slide with zero Align maps to a
// SceneSlide with the zero Content (VAlignTop, HAlignLeft) — backward compat.
func TestMapSlideZeroAlignment(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{ID: "s", Nodes: []contracts.SlideNode{&contracts.Heading{Level: 1}}}
	got := mapSlide(slide, 0)

	if got.Content != (scene.Alignment{}) {
		t.Errorf("zero Align should produce zero scene.Alignment, got %+v", got.Content)
	}
}

// TestMapNodeAlign asserts that per-node Align fields map to the scene node's
// Align field for each supported node type.
func TestMapNodeAlign(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		node     contracts.SlideNode
		wantKind string
		check    func(t *testing.T, sn scene.SlideNode)
	}{
		{
			name:     "hero center",
			node:     &contracts.Hero{Title: "T", Align: contracts.HAlignCenter},
			wantKind: "hero",
			check: func(t *testing.T, sn scene.SlideNode) {
				h, ok := sn.(scene.Hero)
				if !ok {
					t.Fatalf("want scene.Hero, got %T", sn)
				}
				if h.Align != scene.HAlignCenter {
					t.Errorf("Hero.Align = %v, want HAlignCenter", h.Align)
				}
			},
		},
		{
			name:     "heading right",
			node:     &contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "H"}}, Align: contracts.HAlignRight},
			wantKind: "heading",
			check: func(t *testing.T, sn scene.SlideNode) {
				h, ok := sn.(scene.Heading)
				if !ok {
					t.Fatalf("want scene.Heading, got %T", sn)
				}
				if h.Align != scene.HAlignRight {
					t.Errorf("Heading.Align = %v, want HAlignRight", h.Align)
				}
			},
		},
		{
			name:     "prose center",
			node:     &contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "P"}}}, Align: contracts.HAlignCenter},
			wantKind: "prose",
			check: func(t *testing.T, sn scene.SlideNode) {
				p, ok := sn.(scene.Prose)
				if !ok {
					t.Fatalf("want scene.Prose, got %T", sn)
				}
				if p.Align != scene.HAlignCenter {
					t.Errorf("Prose.Align = %v, want HAlignCenter", p.Align)
				}
			},
		},
		{
			name:     "quote right",
			node:     &contracts.Quote{Text: contracts.RichText{{Text: "Q"}}, Align: contracts.HAlignRight},
			wantKind: "quote",
			check: func(t *testing.T, sn scene.SlideNode) {
				q, ok := sn.(scene.Quote)
				if !ok {
					t.Fatalf("want scene.Quote, got %T", sn)
				}
				if q.Align != scene.HAlignRight {
					t.Errorf("Quote.Align = %v, want HAlignRight", q.Align)
				}
			},
		},
		{
			name:     "chip center",
			node:     &contracts.Chip{Label: "C", Align: contracts.HAlignCenter},
			wantKind: "chip",
			check: func(t *testing.T, sn scene.SlideNode) {
				ch, ok := sn.(scene.Chip)
				if !ok {
					t.Fatalf("want scene.Chip, got %T", sn)
				}
				if ch.Align != scene.HAlignCenter {
					t.Errorf("Chip.Align = %v, want HAlignCenter", ch.Align)
				}
			},
		},
		{
			name:     "section_divider right",
			node:     &contracts.SectionDivider{Label: "S", Align: contracts.HAlignRight},
			wantKind: "section_divider",
			check: func(t *testing.T, sn scene.SlideNode) {
				sd, ok := sn.(scene.SectionDivider)
				if !ok {
					t.Fatalf("want scene.SectionDivider, got %T", sn)
				}
				if sd.Align != scene.HAlignRight {
					t.Errorf("SectionDivider.Align = %v, want HAlignRight", sd.Align)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sn := mapNode(tc.node)
			if sn == nil {
				t.Fatal("mapNode returned nil")
			}
			tc.check(t, sn)
		})
	}
}

// TestMapAlignEnums asserts every VAlign and HAlign wire value maps to the
// correct scene enum constant.
func TestMapAlignEnums(t *testing.T) {
	t.Parallel()

	vCases := []struct {
		in   contracts.VAlign
		want scene.VAlign
	}{
		{contracts.VAlignTop, scene.VAlignTop},
		{"", scene.VAlignTop}, // empty = default = top
		{contracts.VAlignCenter, scene.VAlignCenter},
		{contracts.VAlignBottom, scene.VAlignBottom},
		{contracts.VAlignJustify, scene.VAlignJustify},
		{contracts.VAlignFill, scene.VAlignFill},
		{contracts.VAlignBalanced, scene.VAlignBalanced},
	}
	for _, tc := range vCases {
		if got := mapVAlign(tc.in); got != tc.want {
			t.Errorf("mapVAlign(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}

	hCases := []struct {
		in   contracts.HAlign
		want scene.HAlign
	}{
		{contracts.HAlignLeft, scene.HAlignLeft},
		{"", scene.HAlignLeft}, // empty = default = left
		{contracts.HAlignCenter, scene.HAlignCenter},
		{contracts.HAlignRight, scene.HAlignRight},
	}
	for _, tc := range hCases {
		if got := mapHAlign(tc.in); got != tc.want {
			t.Errorf("mapHAlign(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
