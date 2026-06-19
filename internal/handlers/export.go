package handlers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func (h *handlers) exportDeck(_ context.Context, in contracts.ExportDeckInput) (tool.Result[contracts.ExportDeckOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.ExportDeckOutput]{}, mapDeckError(in.DeckID, err)
	}

	deckSoul := soul.DeckardWhite()
	if stored.SoulID != "" {
		if resolved, ok := h.deps.Souls.Get(stored.SoulID); ok {
			deckSoul = resolved
		}
	}

	doc := contracts.SlideDoc{Title: stored.Title, Slides: append([]contracts.Slide(nil), stored.Slides...)}
	resolver := raster.NewStoreResolver(h.deps.Assets)
	path, stats, err := exportstore.ExportWithResolver(h.deps.Workspace, stored.ID, doc, deckSoul, resolver)
	if err != nil {
		return tool.Result[contracts.ExportDeckOutput]{}, err
	}
	out := contracts.ExportDeckOutput{
		Path:        path,
		ResourceURI: exportstore.DeckResourceURI(stored.ID),
		Stats:       contracts.ExportStats{Slides: stats.Slides, Shapes: stats.Shapes, Warnings: append([]string(nil), stats.Warnings...)},
	}
	return tool.Result[contracts.ExportDeckOutput]{Text: fmt.Sprintf("Exported deck %q to %s.", deckLabel(stored), out.Path), Structured: out}, nil
}

func (h *handlers) listResources(_ context.Context, _ contracts.ListResourcesInput) (tool.Result[contracts.ListResourcesOutput], error) {
	entries, err := os.ReadDir(filepath.Join(h.deps.Workspace, "exports"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return tool.Result[contracts.ListResourcesOutput]{Text: "Found 0 resources.", Structured: contracts.ListResourcesOutput{Resources: []contracts.ResourceSummary{}}}, nil
		}
		return tool.Result[contracts.ListResourcesOutput]{}, err
	}

	resources := make([]contracts.ResourceSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pptx" {
			continue
		}
		deckID := strings.TrimSuffix(entry.Name(), ".pptx")
		resources = append(resources, contracts.ResourceSummary{URI: exportstore.DeckResourceURI(deckID), MIME: exportstore.PPTXMIMEType, Title: entry.Name()})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].URI < resources[j].URI })
	out := contracts.ListResourcesOutput{Resources: resources}
	return tool.Result[contracts.ListResourcesOutput]{Text: fmt.Sprintf("Found %d resources.", len(resources)), Structured: out}, nil
}

func (h *handlers) getResource(_ context.Context, in contracts.GetResourceInput) (tool.Result[contracts.GetResourceOutput], error) {
	deckID, err := exportstore.ParseDeckID(in.URI)
	if err != nil {
		return tool.Result[contracts.GetResourceOutput]{}, err
	}
	path, err := filepath.Abs(exportstore.ExportPath(h.deps.Workspace, deckID))
	if err != nil {
		return tool.Result[contracts.GetResourceOutput]{}, err
	}
	_, err = os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			out := contracts.GetResourceOutput{URI: in.URI, MIME: exportstore.PPTXMIMEType, Path: path, Found: false}
			return tool.Result[contracts.GetResourceOutput]{Text: fmt.Sprintf("Resource %s not found.", in.URI), Structured: out}, nil
		}
		return tool.Result[contracts.GetResourceOutput]{}, err
	}
	out := contracts.GetResourceOutput{URI: in.URI, MIME: exportstore.PPTXMIMEType, Path: path, Found: true}
	return tool.Result[contracts.GetResourceOutput]{Text: fmt.Sprintf("Resolved resource %s.", in.URI), Structured: out}, nil
}
