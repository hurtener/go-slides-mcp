package render

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// validPNG is a 1x1 transparent PNG (same fixture used by render_assets_test).
// The decode-config layer reads IHDR width/height so the asset path places it
// with the correct aspect.
const validPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

func decodePNG(t *testing.T) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(validPNG)
	if err != nil {
		t.Fatalf("base64 decode PNG: %v", err)
	}
	return b
}

func TestMapNodeLockupIconAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Lockup{
		Caption:   "POWERED BY",
		Icon:      "star",
		AssetSide: contracts.TrailCaption,
		MaxHeight: 18,
		Align:     contracts.HAlignCenter,
	}
	sn := mapNode(node)
	l, ok := sn.(scene.Lockup)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Lockup", sn)
	}
	if l.Caption != "POWERED BY" {
		t.Errorf("Caption: got %q, want %q", l.Caption, "POWERED BY")
	}
	if l.Icon != "star" {
		t.Errorf("Icon: got %q, want %q", l.Icon, "star")
	}
	if l.AssetID != "" {
		t.Errorf("AssetID: got %q, want empty on icon-path", l.AssetID)
	}
	if l.AssetSide != scene.TrailCaption {
		t.Errorf("AssetSide: got %v, want scene.TrailCaption", l.AssetSide)
	}
	if l.MaxHeight != pptx.Pt(18) {
		t.Errorf("MaxHeight: got %v, want %v", l.MaxHeight, pptx.Pt(18))
	}
	if l.Align != scene.HAlignCenter {
		t.Errorf("Align: got %v, want scene.HAlignCenter", l.Align)
	}
}

func TestMapNodeLockupAssetAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Lockup{
		Caption:   "IN PARTNERSHIP WITH",
		AssetID:   "asset://logo-acme",
		AssetSide: contracts.LeadCaption,
		MaxHeight: 20,
		Align:     contracts.HAlignLeft,
	}
	sn := mapNode(node)
	l, ok := sn.(scene.Lockup)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Lockup", sn)
	}
	if l.AssetID != scene.AssetID("asset://logo-acme") {
		t.Errorf("AssetID: got %q, want %q", l.AssetID, scene.AssetID("asset://logo-acme"))
	}
	if l.Icon != "" {
		t.Errorf("Icon: got %q, want empty on asset-path", l.Icon)
	}
	if l.AssetSide != scene.LeadCaption {
		t.Errorf("AssetSide: got %v, want scene.LeadCaption", l.AssetSide)
	}
	if l.MaxHeight != pptx.Pt(20) {
		t.Errorf("MaxHeight: got %v, want %v", l.MaxHeight, pptx.Pt(20))
	}
	if l.Align != scene.HAlignLeft {
		t.Errorf("Align: got %v, want scene.HAlignLeft", l.Align)
	}
}

func TestRenderLockupIconEffect(t *testing.T) {
	t.Parallel()

	empty := contracts.SlideDoc{
		Title: "Lockup baseline",
		Slides: []contracts.Slide{{
			ID:     "s",
			Layout: contracts.LayoutTitleContent,
			Nodes:  []contracts.SlideNode{&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Partnered With"}}}},
		}},
	}
	withIcon := contracts.SlideDoc{
		Title: "Lockup icon effect",
		Slides: []contracts.Slide{{
			ID:     "s",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Partnered With"}}},
				&contracts.Lockup{Caption: "POWERED BY", Icon: "star", AssetSide: contracts.LeadCaption, MaxHeight: 18, Align: contracts.HAlignCenter},
			},
		}},
	}
	s := soul.DeckardWhite()

	base, baseStats, err := Render(empty, s)
	if err != nil {
		t.Fatalf("baseline Render() error = %v", err)
	}
	assertValidPPTX(t, base)

	got, gotStats, err := Render(withIcon, s)
	if err != nil {
		t.Fatalf("withIcon Render() error = %v", err)
	}
	assertValidPPTX(t, got)
	if gotStats.Shapes <= baseStats.Shapes {
		t.Errorf("Lockup(icon) Shapes = %d, want > baseline %d (node renders no shapes?)", gotStats.Shapes, baseStats.Shapes)
	}
	if _, err := pptx.NewFromBytes(got); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}

	again, _, err := Render(withIcon, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(got, again) {
		t.Fatal("Render() bytes differ across identical icon-path renders")
	}
	defW, _, err := renderWithWorkers(withIcon, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneW, _, err := renderWithWorkers(withIcon, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defW, oneW) {
		t.Fatal("render bytes differ across worker counts on icon-path lockup")
	}
}

func TestRenderLockupAssetResolver(t *testing.T) {
	t.Parallel()

	png := decodePNG(t)
	store := asset.NewMemoryStore()
	stored, err := store.Put("logo.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}
	resolver := raster.NewStoreResolver(store)
	doc := contracts.SlideDoc{
		Title: "Lockup asset effect",
		Slides: []contracts.Slide{{
			ID:     "s",
			Layout: contracts.LayoutTitleContent,
			Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Partnered With"}}},
				&contracts.Lockup{Caption: "POWERED BY", AssetID: contracts.AssetID(stored.ID), AssetSide: contracts.TrailCaption, MaxHeight: 18},
			},
		}},
	}
	s := soul.DeckardWhite()

	got, stats, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	assertValidPPTX(t, got)
	if stats.Assets < 1 {
		t.Errorf("stats.Assets = %d, want >= 1 (resolved media path)", stats.Assets)
	}
	if _, err := pptx.NewFromBytes(got); err != nil {
		t.Fatalf("pptx.NewFromBytes() error = %v", err)
	}
	again, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("second RenderWithAssets() error = %v", err)
	}
	if !bytes.Equal(got, again) {
		t.Fatal("RenderWithAssets() bytes differ across identical asset-path renders")
	}
	defW, _, err := renderWithWorkers(doc, s, 0, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneW, _, err := renderWithWorkers(doc, s, 1, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defW, oneW) {
		t.Fatal("render bytes differ across worker counts on asset-path lockup")
	}
}
