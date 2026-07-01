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

	"github.com/hurtener/go-slides-mcp/internal/autofit"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/exportstore"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// overflowMarkers is the closed set of substrings (R10.11) the engine's
// layout warning messages carry when a slide's content overflows its
// region — the remediation ladder targets exactly these slides. Matched
// case-sensitively against the engine's (lowercase) Message text.
var overflowMarkers = []string{
	"overflow",
	"exceeds the region",
	"exceeds the slide safe area",
	"breadth exceeds",
}

// isOverflow reports whether msg is an overflow-shaped layout warning (R10.11).
func isOverflow(msg string) bool {
	for _, marker := range overflowMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func (h *handlers) exportDeck(_ context.Context, in contracts.ExportDeckInput) (tool.Result[contracts.ExportDeckOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.ExportDeckOutput]{}, mapDeckError(in.DeckID, err)
	}

	deckSoul := soul.DeckardWhite()
	effectiveSoulID := soul.DeckardWhiteID
	resolvedOK := false
	if stored.SoulID != "" {
		if resolved, ok := h.deps.Souls.Get(stored.SoulID); ok {
			deckSoul = resolved
			effectiveSoulID = stored.SoulID
			resolvedOK = true
		}
	}
	established := brandSoulEstablished(stored.SoulID) && resolvedOK

	doc := contracts.SlideDoc{Title: stored.Title, Chrome: mapChrome(stored.Chrome), Slides: append([]contracts.Slide(nil), stored.Slides...)}
	resolver := raster.NewStoreResolver(h.deps.Assets)
	if in.Autofit {
		doc = autofit.Fill(doc)
		doc, _, err = autofit.Remediate(doc, func(d contracts.SlideDoc) (map[string]bool, error) {
			_, st, err := render.RenderWithAssets(d, deckSoul, resolver)
			if err != nil {
				return nil, err
			}
			overflowing := map[string]bool{}
			for _, w := range st.LayoutWarnings {
				if isOverflow(w.Message) {
					overflowing[w.SlideID] = true
				}
			}
			return overflowing, nil
		})
		if err != nil {
			return tool.Result[contracts.ExportDeckOutput]{}, err
		}
	}
	path, stats, err := exportstore.ExportWithResolver(h.deps.Workspace, stored.ID, doc, deckSoul, resolver)
	if err != nil {
		return tool.Result[contracts.ExportDeckOutput]{}, err
	}
	out := contracts.ExportDeckOutput{
		Path:                 path,
		ResourceURI:          exportstore.DeckResourceURI(stored.ID),
		SoulID:               effectiveSoulID,
		BrandSoulEstablished: established,
		Stats:                contracts.ExportStats{Slides: stats.Slides, Shapes: stats.Shapes, Warnings: append([]string(nil), stats.Warnings...)},
	}
	text := fmt.Sprintf("Exported deck %q to %s.", deckLabel(stored), out.Path)
	if !established {
		text = fmt.Sprintf("Exported deck %q to %s. %s", deckLabel(stored), out.Path, noBrandSoulNotice)
	}
	return tool.Result[contracts.ExportDeckOutput]{Text: text, Structured: out}, nil
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
