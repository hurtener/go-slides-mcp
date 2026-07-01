package render_test

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// plainAssetBgDoc is a cover slide with a plain full-bleed asset background —
// the shape-count baseline for the scrim/duotone accept case below.
func plainAssetBgDoc(id contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Asset Background — plain",
		Slides: []contracts.Slide{
			{
				ID:     "cover",
				Layout: contracts.LayoutCover,
				Background: &contracts.Background{
					Kind:    contracts.BackgroundAsset,
					AssetID: string(id),
				},
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Title: "Photo background"},
				},
			},
		},
	}
}

// scrimDuotoneAssetBgDoc mirrors plainAssetBgDoc but adds a gradient Scrim
// and a Duotone recolor over the same photo background (R14.1).
func scrimDuotoneAssetBgDoc(id contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Asset Background — scrim + duotone",
		Slides: []contracts.Slide{
			{
				ID:     "cover",
				Layout: contracts.LayoutCover,
				Background: &contracts.Background{
					Kind:    contracts.BackgroundAsset,
					AssetID: string(id),
					Scrim: &contracts.Scrim{
						Color:    contracts.ColorCanvas,
						Opacity:  0.6,
						Gradient: true,
					},
					Duotone: &contracts.Duotone{
						Shadow:    contracts.ColorAccent,
						Highlight: contracts.ColorSurface,
					},
				},
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Title: "Photo background"},
				},
			},
		},
	}
}

// cardImageFillDoc is a single card whose surface is filled by a cover-fit
// photo instead of a solid Fill (R14.1).
func cardImageFillDoc(id contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Card ImageFill",
		Slides: []contracts.Slide{
			{
				ID:     "content",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Card{
						Header:    "Featured",
						ImageFill: id,
						Body: []contracts.SlideNode{
							&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Photographic card surface."}}}},
						},
					},
				},
			},
		},
	}
}

// roundedImageDoc is a single Image node with CornerRadius + Elevation set
// (R13.11).
func roundedImageDoc(id contracts.AssetID) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Rounded Image",
		Slides: []contracts.Slide{
			{
				ID:     "content",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Image{
						AssetID:      id,
						Alt:          "Rounded hero",
						CornerRadius: contracts.RadiusLG,
						Elevation:    contracts.ElevationRaised,
					},
				},
			},
		},
	}
}

// TestRenderBackgroundScrimDuotoneEmitsMoreShapesThanPlain is the R14.1
// product-level accept case: a photo background with a Scrim + Duotone
// renders without error and emits more shapes than the same photo
// background with neither (the scrim/duotone overlay draws extra shapes).
func TestRenderBackgroundScrimDuotoneEmitsMoreShapesThanPlain(t *testing.T) {
	t.Parallel()

	png := decodePNG(t)
	store := asset.NewMemoryStore()
	stored, err := store.Put("bg.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}
	resolver := raster.NewStoreResolver(store)
	s := soul.DeckardWhite()

	baseBuf, baseStats, err := render.RenderWithAssets(plainAssetBgDoc(contracts.AssetID(stored.ID)), s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets(plain) error = %v", err)
	}
	if len(baseBuf) == 0 {
		t.Fatal("RenderWithAssets(plain) returned empty bytes")
	}

	buf, stats, err := render.RenderWithAssets(scrimDuotoneAssetBgDoc(contracts.AssetID(stored.ID)), s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets(scrim+duotone) error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets(scrim+duotone) returned empty bytes")
	}
	if stats.Shapes <= baseStats.Shapes {
		t.Errorf("scrim+duotone Shapes = %d, want > plain Shapes %d", stats.Shapes, baseStats.Shapes)
	}
}

// TestRenderCardImageFillRenders is the R14.1 product-level accept case for
// Card.ImageFill: a card with a photo ImageFill renders without error and
// resolves the asset.
func TestRenderCardImageFillRenders(t *testing.T) {
	t.Parallel()

	png := decodePNG(t)
	store := asset.NewMemoryStore()
	stored, err := store.Put("card.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}
	resolver := raster.NewStoreResolver(store)

	buf, stats, err := render.RenderWithAssets(cardImageFillDoc(contracts.AssetID(stored.ID)), soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes")
	}
	if stats.Assets < 1 {
		t.Errorf("stats.Assets = %d, want >= 1", stats.Assets)
	}
}

// TestRenderImageCornerRadiusElevationRenders is the R13.11 product-level
// accept case: an Image with CornerRadius + Elevation renders without error
// and resolves the asset.
func TestRenderImageCornerRadiusElevationRenders(t *testing.T) {
	t.Parallel()

	png := decodePNG(t)
	store := asset.NewMemoryStore()
	stored, err := store.Put("rounded.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}
	resolver := raster.NewStoreResolver(store)

	buf, stats, err := render.RenderWithAssets(roundedImageDoc(contracts.AssetID(stored.ID)), soul.DeckardWhite(), resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("RenderWithAssets() returned empty bytes")
	}
	if stats.Assets < 1 {
		t.Errorf("stats.Assets = %d, want >= 1", stats.Assets)
	}
}
