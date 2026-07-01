package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

func (h *handlers) createDeck(_ context.Context, in contracts.CreateDeckInput) (tool.Result[contracts.CreateDeckOutput], error) {
	stored, err := h.deps.Store.CreateDeck(deck.CreateDeckInput{Title: in.Title, Author: in.Author, SoulID: in.SoulID})
	if err != nil {
		return tool.Result[contracts.CreateDeckOutput]{}, err
	}
	established := brandSoulEstablished(stored.SoulID)
	out := contracts.CreateDeckOutput{
		Kind:                 contracts.DeckKindDeck,
		DeckID:               stored.ID,
		Slug:                 stored.Slug,
		Title:                stored.Title,
		SoulID:               stored.SoulID,
		BrandSoulEstablished: established,
		Slides:               slideSummaries(stored),
	}
	text := fmt.Sprintf("Created deck %q (%s).", deckLabel(stored), stored.ID)
	if !established {
		text = fmt.Sprintf("Created deck %q (%s). %s", deckLabel(stored), stored.ID, noBrandSoulNotice)
	}
	return tool.Result[contracts.CreateDeckOutput]{Text: text, Structured: out}, nil
}

func (h *handlers) listDecks(_ context.Context, _ contracts.ListDecksInput) (tool.Result[contracts.ListDecksOutput], error) {
	decks := h.deps.Store.ListDecks()
	out := contracts.ListDecksOutput{Decks: make([]contracts.DeckSummary, 0, len(decks))}
	for _, stored := range decks {
		out.Decks = append(out.Decks, contracts.DeckSummary{
			DeckID:     stored.ID,
			Slug:       stored.Slug,
			Title:      stored.Title,
			SoulID:     stored.SoulID,
			SlideCount: len(stored.Slides),
			UpdatedAt:  stored.UpdatedAt,
		})
	}
	return tool.Result[contracts.ListDecksOutput]{Text: fmt.Sprintf("Found %d deck(s).", len(out.Decks)), Structured: out}, nil
}

func (h *handlers) getDeck(_ context.Context, in contracts.GetDeckInput) (tool.Result[contracts.GetDeckOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.GetDeckOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.GetDeckOutput{
		Kind:     contracts.DeckKindDeck,
		DeckID:   stored.ID,
		Slug:     stored.Slug,
		Title:    stored.Title,
		SoulID:   stored.SoulID,
		Chrome:   mapChrome(stored.Chrome),
		Sections: mapSections(stored.Sections),
		Slides:   slideSummaries(stored),
	}
	return tool.Result[contracts.GetDeckOutput]{Text: fmt.Sprintf("Loaded deck %q with %d slide(s).", deckLabel(stored), len(stored.Slides)), Structured: out}, nil
}

func (h *handlers) deleteDeck(_ context.Context, in contracts.DeleteDeckInput) (tool.Result[contracts.DeleteDeckOutput], error) {
	if err := h.deps.Store.DeleteDeck(in.DeckID); err != nil {
		return tool.Result[contracts.DeleteDeckOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.DeleteDeckOutput{DeckID: in.DeckID, Deleted: true}
	return tool.Result[contracts.DeleteDeckOutput]{Text: fmt.Sprintf("Deleted deck %q.", in.DeckID), Structured: out}, nil
}

func (h *handlers) setDeckChrome(_ context.Context, in contracts.SetDeckChromeInput) (tool.Result[contracts.SetDeckChromeOutput], error) {
	stored, err := h.deps.Store.SetChrome(in.DeckID, deck.Chrome{Enabled: in.Chrome.Enabled, BrandAssetID: in.Chrome.BrandAssetID, BrandText: in.Chrome.BrandText})
	if err != nil {
		return tool.Result[contracts.SetDeckChromeOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.SetDeckChromeOutput{DeckID: stored.ID, Chrome: mapChrome(stored.Chrome)}
	return tool.Result[contracts.SetDeckChromeOutput]{Text: fmt.Sprintf("Updated chrome for deck %q.", deckLabel(stored)), Structured: out}, nil
}

func (h *handlers) setDeckSections(_ context.Context, in contracts.SetDeckSectionsInput) (tool.Result[contracts.SetDeckSectionsOutput], error) {
	stored, err := h.deps.Store.SetSections(in.DeckID, unmapSections(in.Sections))
	if err != nil {
		return tool.Result[contracts.SetDeckSectionsOutput]{}, mapDeckError(in.DeckID, err)
	}
	out := contracts.SetDeckSectionsOutput{DeckID: stored.ID, Sections: mapSections(stored.Sections)}
	return tool.Result[contracts.SetDeckSectionsOutput]{Text: fmt.Sprintf("Updated %d section(s) for deck %q.", len(out.Sections), deckLabel(stored)), Structured: out}, nil
}

func mapDeckError(id string, err error) error {
	if errors.Is(err, deck.ErrNotFound) {
		return fmt.Errorf("deck %q not found: %w", id, err)
	}
	return err
}

func mapChrome(chrome deck.Chrome) contracts.DeckChrome {
	return contracts.DeckChrome{Enabled: chrome.Enabled, BrandAssetID: chrome.BrandAssetID, BrandText: chrome.BrandText}
}

func mapSections(sections []deck.Section) []contracts.DeckSection {
	out := make([]contracts.DeckSection, 0, len(sections))
	for _, section := range sections {
		out = append(out, contracts.DeckSection{
			Name:      section.Name,
			SlideIDs:  append([]string(nil), section.SlideIDs...),
			Variant:   contracts.Variant(section.Variant),
			Archetype: contracts.SlideArchetype(section.Archetype),
		})
	}
	return out
}

func unmapSections(sections []contracts.DeckSection) []deck.Section {
	out := make([]deck.Section, 0, len(sections))
	for _, section := range sections {
		out = append(out, deck.Section{
			Name:      section.Name,
			SlideIDs:  append([]string(nil), section.SlideIDs...),
			Variant:   string(section.Variant),
			Archetype: string(section.Archetype),
		})
	}
	return out
}

func slideSummaries(stored *deck.Deck) []contracts.SlideSummary {
	out := make([]contracts.SlideSummary, 0, len(stored.Slides))
	for _, slide := range stored.Slides {
		title, preview := summarizeSlide(slide)
		out = append(out, contracts.SlideSummary{SlideID: slide.ID, Layout: slide.Layout, Title: title, PreviewText: preview, Revision: stored.Revision})
	}
	return out
}

func summarizeSlide(slide contracts.Slide) (string, string) {
	parts := make([]string, 0, 2)
	for _, node := range slide.Nodes {
		switch n := node.(type) {
		case *contracts.Hero:
			appendIfSet(&parts, n.Title)
			appendIfSet(&parts, n.Subtitle)
		case *contracts.Heading:
			appendIfSet(&parts, richTextString(n.Text))
		case *contracts.Prose:
			for _, paragraph := range n.Paragraphs {
				appendIfSet(&parts, richTextString(paragraph))
				if len(parts) >= 2 {
					break
				}
			}
		case *contracts.Callout:
			appendIfSet(&parts, n.Title)
			appendIfSet(&parts, richTextString(n.Body))
		}
		if len(parts) >= 2 {
			break
		}
	}
	if len(parts) == 0 {
		return "Untitled slide", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func appendIfSet(parts *[]string, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	*parts = append(*parts, trimmed)
}

func richTextString(text contracts.RichText) string {
	var b strings.Builder
	for _, run := range text {
		b.WriteString(run.Text)
	}
	return strings.TrimSpace(b.String())
}

func deckLabel(stored *deck.Deck) string {
	if strings.TrimSpace(stored.Title) != "" {
		return stored.Title
	}
	return stored.Slug
}
