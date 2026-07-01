package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// plainQuoteDoc builds a one-slide asset-free doc with a plain Quote (only
// Text+Attribution set) — no testimonial enrichment.
func plainQuoteDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Quote Coverage — Plain",
		Slides: []contracts.Slide{
			{
				ID:     "plain-quote",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Quote{
						Text:        rt("The best way to predict the future is to invent it."),
						Attribution: "Alan Kay",
					},
				},
			},
		},
	}
}

// enrichedQuoteDoc builds a one-slide asset-free doc with a testimonial
// Quote (R14.5, D-120): Mark + structured attribution, covering the
// enriched layout without requiring an AssetResolver.
func enrichedQuoteDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Quote Coverage — Enriched",
		Slides: []contracts.Slide{
			{
				ID:     "enriched-quote",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Quote{
						Text:               rt("Deckard cut our deck-build time from days to minutes."),
						Mark:               true,
						AttributionName:    "Priya Natarajan",
						AttributionRole:    "VP Marketing",
						AttributionCompany: "Northwind Labs",
					},
				},
			},
		},
	}
}

// TestQuoteEnrichedEmitsMoreShapesThanPlain proves the testimonial
// enrichment has a render effect, not dead infra: the enriched layout must
// emit strictly more shapes than a plain Quote.
func TestQuoteEnrichedEmitsMoreShapesThanPlain(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	_, plainStats, err := Render(plainQuoteDoc(), s)
	if err != nil {
		t.Fatalf("Render(plain) error = %v", err)
	}
	_, enrichedStats, err := Render(enrichedQuoteDoc(), s)
	if err != nil {
		t.Fatalf("Render(enriched) error = %v", err)
	}
	if enrichedStats.Shapes <= plainStats.Shapes {
		t.Fatalf("enriched quote shapes = %d, want > plain quote shapes %d", enrichedStats.Shapes, plainStats.Shapes)
	}
}

// TestQuotePlainByteIdenticalToPreChange asserts a plain Quote (only
// Text+Attribution set, all R14.5 fields zero) renders without error and
// deterministically — the byte-identity contract for the additive fields.
func TestQuotePlainByteIdenticalToPreChange(t *testing.T) {
	t.Parallel()

	doc := plainQuoteDoc()
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
		t.Fatal("Render() bytes differ across identical plain-quote renders")
	}
}
