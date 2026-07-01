package render

import (
	"fmt"
	"sort"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// LayoutWarning is a structured, per-warning mirror of one entry in Warnings
// (R10.11): the same slide/node/message triple the engine records, kept
// structured so a caller (the export remediation ladder) can group and act on
// warnings by SlideID without parsing the formatted string in Warnings.
type LayoutWarning struct {
	SlideID string
	Node    string
	Message string
}

// Stats is the render summary returned alongside the rendered PPTX bytes.
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []string
	// LayoutWarnings is the structured form of Warnings (R10.11), one entry
	// per warning, carrying SlideID/Node/Message separately. Additive: it
	// mirrors Warnings 1:1 in the same order and never replaces it — existing
	// callers that consume Warnings are unaffected.
	LayoutWarnings []LayoutWarning
	// Colors are the per-slide resolved canvas/surface/primary-text RGBs the
	// engine rendered each slide with (R7, scene.Stats.Colors). In scene order.
	// VariantDark slides carry their derived dark palette here, not the soul's
	// light theme — callers use these for variant-aware contrast auditing.
	Colors []scene.SlideColors
}

// Render maps doc+soul to a scene and renders deterministic .pptx bytes.
// Asset-free decks stay byte-identical across calls; pass a resolver via
// RenderWithAssets to resolve asset-backed nodes (Image, Chart, CodeBlock,
// asset Decoration). Render is the public, resolver-less entry point — see
// RenderWithAssets for the asset-aware variant.
func Render(doc contracts.SlideDoc, s *soul.Soul) ([]byte, Stats, error) {
	return RenderWithAssets(doc, s, nil)
}

// RenderWithAssets renders doc into PPTX bytes and threads resolver into the
// scene renderer so asset-backed nodes (Image/Chart/CodeBlock/asset
// Decoration) can resolve their bytes via scene.WithAssetResolver. A nil
// resolver is allowed and yields the asset-free rendering path (no asset
// bytes, no resolution, no asset-loss warnings).
//
//nolint:revive // RenderWithAssets reads clearly and pairs with Render(doc, soul); the suggested "WithAssets" is less clear.
func RenderWithAssets(doc contracts.SlideDoc, s *soul.Soul, resolver scene.AssetResolver) ([]byte, Stats, error) {
	return renderWithWorkers(doc, s, 0, resolver)
}

func renderWithWorkers(doc contracts.SlideDoc, s *soul.Soul, workers int, resolver scene.AssetResolver) ([]byte, Stats, error) {
	if s == nil || s.Theme == nil {
		return nil, Stats{}, fmt.Errorf("render: nil soul theme")
	}

	// R13.12: fill in each slide's Background/decorations from the soul's
	// per-archetype DecorPolicy when the slide sets none. A nil/empty policy
	// (the built-in Deckard White soul) returns doc unchanged — byte-identical
	// to the pre-R13-D render path. Runs on every render (preview, export,
	// autofit probes): composition is deterministic and idempotent for a
	// given (doc, soul), so re-running it across probe renders stays
	// consistent.
	doc = applyDecorPolicy(doc, s)

	popts := []pptx.Option{pptx.WithTheme(s.Theme)}
	if s.FontProvider != nil {
		// Embed the brand faces the deck actually uses (R9.1): the engine's
		// save-time pass walks every run, collects the distinct (family, weight,
		// italic) set in a stable sorted order, and resolves each through the
		// provider — so the serif display/heading faces ship inside the .pptx and
		// render on any machine. A nil provider skips both options and keeps the
		// render byte-identical to the pre-embedding output.
		popts = append(popts, pptx.WithFontEmbedding(), pptx.WithFontSource(s.FontProvider))
	}
	pres := pptx.New(popts...)
	sc := scene.Scene{
		Theme:  s.Theme,
		Slides: mapSlides(doc.Slides),
		Meta: scene.Metadata{
			Title: doc.Title,
		},
		Chrome: mapDocChrome(doc.Chrome),
	}

	opts := []scene.RenderOption{scene.WithWorkers(workers)}
	if resolver != nil {
		opts = append(opts, scene.WithAssetResolver(resolver))
	}
	// R14.16: register each brand glyph in the soul's IconSet so every
	// Card/Flow/Milestone/etc. icon reference resolves from the brand set
	// before the curated set. Sorted name order keeps registration (and thus
	// the rendered bytes) deterministic. An empty/nil IconSet appends nothing
	// — byte-identical to a soul without the field.
	if len(s.IconSet) > 0 {
		names := make([]string, 0, len(s.IconSet))
		for name := range s.IconSet {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			opts = append(opts, scene.WithIconExtension(name, []byte(s.IconSet[name])))
		}
	}
	sceneStats, err := scene.Render(pres, sc, opts...)
	if err != nil {
		return nil, Stats{}, fmt.Errorf("render scene: %w", err)
	}

	buf, err := pres.WriteToBytes()
	if err != nil {
		return nil, Stats{}, fmt.Errorf("write pptx bytes: %w", err)
	}

	return buf, statsFromScene(sceneStats), nil
}

// mapDocChrome converts the deck-level DeckChrome contract to the engine's
// scene.Chrome. The zero value (Enabled == false) returns an empty
// scene.Chrome so a chrome-free deck renders byte-identical to before R3.
// When enabled, Total is left at 0: the engine uses len(Scene.Slides) as the
// page-number denominator by convention.
func mapDocChrome(c contracts.DeckChrome) scene.Chrome {
	if !c.Enabled {
		return scene.Chrome{}
	}
	return scene.Chrome{
		Enabled:    true,
		Brand:      c.BrandText,
		BrandAsset: scene.AssetID(c.BrandAssetID),
	}
}

func statsFromScene(s scene.Stats) Stats {
	warnings := make([]string, 0, len(s.Warnings))
	layoutWarnings := make([]LayoutWarning, 0, len(s.Warnings))
	for _, warning := range s.Warnings {
		warnings = append(warnings, fmt.Sprintf("slide=%s node=%s: %s", warning.SlideID, warning.Node, warning.Message))
		layoutWarnings = append(layoutWarnings, LayoutWarning{SlideID: warning.SlideID, Node: warning.Node, Message: warning.Message})
	}
	return Stats{
		Slides:         s.Slides,
		Shapes:         s.Shapes,
		Assets:         s.Assets,
		Warnings:       warnings,
		LayoutWarnings: layoutWarnings,
		Colors:         s.Colors,
	}
}
