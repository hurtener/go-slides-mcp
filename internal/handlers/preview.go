package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
)

// getDeckPreview builds the glanceable deck-preview payload for the
// deck-preview surface: a brand config, a deck summary, and per-slide thumbnail
// descriptors rendered natively from the IR.
func (h *handlers) getDeckPreview(_ context.Context, in contracts.DeckPreviewInput) (tool.Result[contracts.DeckPreviewOutput], error) {
	deckID := in.DeckID
	if deckID == "" {
		active, _, _ := h.deps.Session.Snapshot()
		deckID = active
	}

	brand := h.resolveBrand()
	if deckID == "" {
		return previewResult(contracts.DeckPreviewOutput{
			State:   "empty",
			Message: "No active deck. Create a deck to preview it here.",
			Brand:   brand,
		}), nil
	}

	stored, err := h.deps.Store.GetDeck(deckID)
	if err != nil {
		return previewResult(contracts.DeckPreviewOutput{
			State:   "error",
			Message: fmt.Sprintf("Deck %q not found.", deckID),
			Brand:   brand,
		}), nil
	}

	out := contracts.DeckPreviewOutput{
		Brand: brand,
		Deck: contracts.DeckSummary{
			DeckID:     stored.ID,
			Slug:       stored.Slug,
			Title:      stored.Title,
			SlideCount: len(stored.Slides),
			SoulID:     stored.SoulID,
		},
		ResourceURI: exportstore.DeckResourceURI(stored.ID),
	}
	if len(stored.Slides) == 0 {
		out.State = "empty"
		out.Message = "This deck has no slides yet."
		return previewResult(out), nil
	}

	out.State = "ready"
	out.Slides = make([]contracts.SlidePreview, 0, len(stored.Slides))
	for i, s := range stored.Slides {
		sp := contracts.SlidePreview{
			ID:     s.ID,
			Index:  i,
			Layout: string(s.Layout),
			Title:  slideTitle(s),
		}
		for _, n := range s.Nodes {
			sp.Nodes = append(sp.Nodes, nodeToThumb(n))
		}
		out.Slides = append(out.Slides, sp)
	}
	return previewResult(out), nil
}

func previewResult(out contracts.DeckPreviewOutput) tool.Result[contracts.DeckPreviewOutput] {
	text := fmt.Sprintf("Deck preview: %s (%d slides).", out.Deck.Title, out.Deck.SlideCount)
	if out.State != "ready" {
		text = "Deck preview: " + out.Message
	}
	return tool.Result[contracts.DeckPreviewOutput]{Text: text, Structured: out}
}

// resolveBrand returns the configured brand, filling Deckard defaults for any
// unset field so the surface always has a usable config.
func (h *handlers) resolveBrand() contracts.AppBrand {
	b := h.deps.Brand
	if b.Title == "" {
		b.Title = "Deckard Slides"
	}
	if b.DefaultTheme == "" {
		b.DefaultTheme = "deckard-white"
	}
	return b
}

// slideTitle is a best-effort label: the first hero title or heading text.
func slideTitle(s contracts.Slide) string {
	for _, n := range s.Nodes {
		switch v := n.(type) {
		case *contracts.Hero:
			if v.Title != "" {
				return v.Title
			}
		case *contracts.Heading:
			if t := v.Text.PlainText(); t != "" {
				return t
			}
		}
	}
	return ""
}

// nodeToThumb reduces one IR node to its glanceable thumbnail descriptor.
func nodeToThumb(n contracts.SlideNode) contracts.ThumbNode {
	t := contracts.ThumbNode{Kind: string(contracts.KindOf(n))}
	switch v := n.(type) {
	case *contracts.Hero:
		t.Text, t.Detail = v.Title, v.Eyebrow
	case *contracts.Heading:
		t.Text = v.Text.PlainText()
	case *contracts.Prose:
		if len(v.Paragraphs) > 0 {
			t.Text = v.Paragraphs[0].PlainText()
		}
		t.Count = len(v.Paragraphs)
	case *contracts.List:
		t.Count = len(v.Items)
	case *contracts.Callout:
		t.Text, t.Accent = v.Title, true
	case *contracts.Quote:
		t.Text, t.Detail = v.Text.PlainText(), v.Attribution
	case *contracts.Chip:
		t.Text, t.Accent = v.Label, v.Tone == contracts.ChipSolid
	case *contracts.Table:
		t.Count = len(v.Rows)
	case *contracts.Flow:
		t.Count = len(v.Steps)
	case *contracts.Grid:
		t.Count = len(v.Cells)
	case *contracts.Card:
		t.Text = v.Header
	case *contracts.CardSection:
		t.Text = v.Header
	case *contracts.SectionDivider:
		t.Text = v.Label
	case *contracts.Arrow:
		t.Text = v.Label
	}
	return t
}
