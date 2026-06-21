package render

import (
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Stats is the render summary returned alongside the rendered PPTX bytes.
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []string
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

	pres := pptx.New(pptx.WithTheme(s.Theme))
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
	for _, warning := range s.Warnings {
		warnings = append(warnings, fmt.Sprintf("slide=%s node=%s: %s", warning.SlideID, warning.Node, warning.Message))
	}
	return Stats{
		Slides:   s.Slides,
		Shapes:   s.Shapes,
		Assets:   s.Assets,
		Warnings: warnings,
	}
}
