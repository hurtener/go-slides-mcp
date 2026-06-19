package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"
	"github.com/hurtener/pptx-go/pptx"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/layout"
)

func (h *handlers) getDeckOverview(_ context.Context, in contracts.GetDeckOverviewInput) (tool.Result[contracts.GetDeckOverviewOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.GetDeckOverviewOutput]{}, mapDeckError(in.DeckID, err)
	}
	sections := make([]contracts.DeckOverviewSection, 0, len(stored.Sections))
	for _, section := range stored.Sections {
		ids := append([]string(nil), section.SlideIDs...)
		sections = append(sections, contracts.DeckOverviewSection{Name: section.Name, SlideIDs: ids})
	}
	out := contracts.GetDeckOverviewOutput{
		Kind:     contracts.SurfaceKindOverview,
		State:    "ready",
		DeckID:   stored.ID,
		Title:    stored.Title,
		Sections: sections,
		Slides:   slideSummaries(stored),
		Brand:    h.resolveBrand(),
	}
	return tool.Result[contracts.GetDeckOverviewOutput]{Text: fmt.Sprintf("Loaded overview for deck %q with %d slide(s).", deckLabel(stored), len(out.Slides)), Structured: out}, nil
}

func (h *handlers) openSlideEditor(_ context.Context, in contracts.OpenSlideEditorInput) (tool.Result[contracts.OpenSlideEditorOutput], error) {
	slide, err := h.deps.Store.GetSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.OpenSlideEditorOutput]{}, mapDeckError(in.DeckID, err)
	}
	validation := validateSlide(*slide)
	deckID := in.DeckID
	soulID := ""
	if stored, err := h.deps.Store.GetDeck(in.DeckID); err == nil {
		deckID = stored.ID
		soulID = stored.SoulID
	}
	soulCtx := h.resolveSoul(soulID)
	out := contracts.OpenSlideEditorOutput{
		Kind:       contracts.SurfaceKindEditor,
		State:      "ready",
		DeckID:     deckID,
		SlideID:    slide.ID,
		IR:         *slide,
		SoulID:     soulID,
		Validation: validation,
		Brand:      h.resolveBrand(),
		Layout:     layout.Compute(*slide, soulCtx.Theme),
		Palette:    soulPalette(soulCtx.Theme),
	}
	return tool.Result[contracts.OpenSlideEditorOutput]{Text: fmt.Sprintf("Opened editor for slide %q in deck %q.", slide.ID, deckID), Structured: out}, nil
}

func (h *handlers) setActiveWorkspace(_ context.Context, in contracts.SetActiveWorkspaceInput) (tool.Result[contracts.SetActiveWorkspaceOutput], error) {
	if deckID := strings.TrimSpace(in.DeckID); deckID != "" {
		if _, err := h.deps.Store.GetDeck(deckID); err != nil {
			return tool.Result[contracts.SetActiveWorkspaceOutput]{}, mapDeckError(deckID, err)
		}
	}
	if soulID := strings.TrimSpace(in.SoulID); soulID != "" {
		if _, ok := h.deps.Souls.Get(soulID); !ok {
			return tool.Result[contracts.SetActiveWorkspaceOutput]{}, fmt.Errorf("soul %q not found", soulID)
		}
	}
	h.deps.Session.SetActive(strings.TrimSpace(in.DeckID), strings.TrimSpace(in.SoulID))
	activeDeckID, activeSoulID, _ := h.deps.Session.Snapshot()
	out := contracts.SetActiveWorkspaceOutput{
		Kind:         contracts.SurfaceKindActiveWorkspace,
		ActiveDeckID: activeDeckID,
		ActiveSoulID: activeSoulID,
	}
	return tool.Result[contracts.SetActiveWorkspaceOutput]{Text: fmt.Sprintf("Set active workspace: deck=%q soul=%q.", activeDeckID, activeSoulID), Structured: out}, nil
}

// soulPalette resolves a soul theme's colors + fonts into CSS-ready hex/family
// strings so the editor canvas paints in the deck's visual language.
func soulPalette(t *pptx.Theme) contracts.SoulPalette {
	if t == nil {
		t = pptx.DefaultTheme()
	}
	hex := func(s pptx.RGB) string { return "#" + string(s) }
	return contracts.SoulPalette{
		Canvas:        hex(t.ResolveColor(pptx.ColorCanvas)),
		Surface:       hex(t.ResolveColor(pptx.ColorSurface)),
		SurfaceAlt:    hex(t.ResolveColor(pptx.ColorSurfaceAlt)),
		Accent:        hex(t.ResolveColor(pptx.ColorAccent)),
		AccentText:    hex(t.ResolveTextColor(pptx.TextAccent)),
		TextPrimary:   hex(t.ResolveTextColor(pptx.TextPrimary)),
		TextSecondary: hex(t.ResolveTextColor(pptx.TextSecondary)),
		TextInverse:   hex(t.ResolveTextColor(pptx.TextInverse)),
		Border:        hex(t.ResolveColor(pptx.ColorSurfaceAlt)),
		HeadingFont:   t.ResolveType(pptx.TypeH1).Family,
		BodyFont:      t.ResolveType(pptx.TypeBody).Family,
		MonoFont:      t.ResolveType(pptx.TypeMono).Family,
	}
}
