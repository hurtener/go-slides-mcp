package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// watermarkDoc builds a one-slide content doc whose Nodes carry a
// DecorationText watermark ("03") in the background layer (R13.9).
func watermarkDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Text Watermark",
		Slides: []contracts.Slide{{
			ID:     "content",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Decoration{
					Kind:  contracts.DecorationText,
					Text:  "03",
					Layer: contracts.LayerBackground,
				},
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Section Three"}}},
			},
		}},
	}
}

// noWatermarkDoc mirrors watermarkDoc's shape without the Decoration node —
// the shape-count baseline.
func noWatermarkDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "No Watermark",
		Slides: []contracts.Slide{{
			ID:     "content",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Section Three"}}},
			},
		}},
	}
}

// TestRenderTextWatermarkEmitsMoreShapesThanNone is the R13.9 product-level
// accept case: a DecorationText watermark renders without error and emits
// more shapes than the same slide with no decoration (proves effect, not dead
// infra), and repeated renders are byte-identical.
func TestRenderTextWatermarkEmitsMoreShapesThanNone(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	baseBuf, baseStats, err := Render(noWatermarkDoc(), s)
	if err != nil {
		t.Fatalf("Render(none) error = %v", err)
	}
	if len(baseBuf) == 0 {
		t.Fatal("Render(none) returned empty bytes")
	}

	doc := watermarkDoc()
	first, firstStats, err := Render(doc, s)
	if err != nil {
		t.Fatalf("Render(watermark) error = %v", err)
	}
	if len(first) == 0 {
		t.Fatal("Render(watermark) returned empty bytes")
	}
	if firstStats.Shapes <= baseStats.Shapes {
		t.Errorf("watermark Shapes = %d, want > none Shapes %d", firstStats.Shapes, baseStats.Shapes)
	}

	second, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("text watermark Render() bytes differ across identical renders")
	}
}

// TestMapNodeDecorationText asserts a DecorationText node maps its Text and
// FontSize fields through to the scene.Decoration (R13.9).
func TestMapNodeDecorationText(t *testing.T) {
	t.Parallel()

	node := &contracts.Decoration{
		Kind:     contracts.DecorationText,
		Text:     "03",
		FontSize: 240,
		Layer:    contracts.LayerBackground,
	}
	sn := mapNode(node)
	d, ok := sn.(scene.Decoration)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Decoration", sn)
	}
	if d.Kind != scene.DecorationText {
		t.Errorf("Kind = %v, want DecorationText", d.Kind)
	}
	if d.Text != "03" {
		t.Errorf("Text = %q, want %q", d.Text, "03")
	}
	if d.FontSize != 240 {
		t.Errorf("FontSize = %v, want 240", d.FontSize)
	}
}

// cardBackdropDoc builds a one-slide content doc with a single card carrying
// a radial_glow Backdrop (R13.10).
func cardBackdropDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Card Backdrop",
		Slides: []contracts.Slide{{
			ID:     "content",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Card{
					Header: "Focal",
					Backdrop: &contracts.Decoration{
						Kind:   contracts.DecorationPreset,
						Preset: "radial_glow",
						Anchor: contracts.AnchorCenter,
						Bleed:  true,
					},
					Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Center card"}}}},
					},
				},
			},
		}},
	}
}

// noCardBackdropDoc mirrors cardBackdropDoc's shape with Backdrop left nil —
// the shape-count baseline.
func noCardBackdropDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "No Card Backdrop",
		Slides: []contracts.Slide{{
			ID:     "content",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Card{
					Header: "Focal",
					Body: []contracts.SlideNode{
						&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Center card"}}}},
					},
				},
			},
		}},
	}
}

// TestRenderCardBackdropEmitsMoreShapesThanNone is the R13.10 product-level
// accept case: a card with a radial_glow Backdrop renders without error and
// emits more shapes than the same card with no Backdrop.
func TestRenderCardBackdropEmitsMoreShapesThanNone(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	baseBuf, baseStats, err := Render(noCardBackdropDoc(), s)
	if err != nil {
		t.Fatalf("Render(none) error = %v", err)
	}
	if len(baseBuf) == 0 {
		t.Fatal("Render(none) returned empty bytes")
	}

	buf, stats, err := Render(cardBackdropDoc(), s)
	if err != nil {
		t.Fatalf("Render(backdrop) error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("Render(backdrop) returned empty bytes")
	}
	if stats.Shapes <= baseStats.Shapes {
		t.Errorf("backdrop Shapes = %d, want > none Shapes %d", stats.Shapes, baseStats.Shapes)
	}
}

// TestMapNodeCardBackdropNilByteIdentical asserts that a Card with no
// Backdrop maps to a scene.Card with a nil Backdrop — byte-identical to a
// card literal built without the field (R13.10).
func TestMapNodeCardBackdropNilByteIdentical(t *testing.T) {
	t.Parallel()

	withoutField := mapNode(&contracts.Card{Header: "Plain"})
	explicitNil := mapNode(&contracts.Card{Header: "Plain", Backdrop: nil})

	cWithout, ok := withoutField.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", withoutField)
	}
	cExplicit, ok := explicitNil.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", explicitNil)
	}
	if cWithout.Backdrop != nil {
		t.Errorf("Backdrop (field unset) = %v, want nil", cWithout.Backdrop)
	}
	if cExplicit.Backdrop != nil {
		t.Errorf("Backdrop (explicit nil) = %v, want nil", cExplicit.Backdrop)
	}
}

// TestMapNodeCardBackdropMapsFields asserts a set Backdrop maps its preset
// and geometry fields through mapDecorationPtr/mapDecoration (R13.10).
func TestMapNodeCardBackdropMapsFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Card{
		Header: "Focal",
		Backdrop: &contracts.Decoration{
			Kind:   contracts.DecorationPreset,
			Preset: "radial_glow",
			Anchor: contracts.AnchorCenter,
			Bleed:  true,
		},
	}
	sn := mapNode(node)
	c, ok := sn.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", sn)
	}
	if c.Backdrop == nil {
		t.Fatal("Backdrop = nil, want non-nil")
	}
	if c.Backdrop.Preset != "radial_glow" {
		t.Errorf("Backdrop.Preset = %q, want %q", c.Backdrop.Preset, "radial_glow")
	}
	if !c.Backdrop.Bleed {
		t.Error("Backdrop.Bleed = false, want true")
	}
	if c.Backdrop.Anchor != mapAnchor(contracts.AnchorCenter) {
		t.Errorf("Backdrop.Anchor = %v, want %v", c.Backdrop.Anchor, mapAnchor(contracts.AnchorCenter))
	}
}
