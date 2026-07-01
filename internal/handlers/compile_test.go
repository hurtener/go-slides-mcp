package handlers

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func TestCompileChartStoresAssetAndReturnsNode(t *testing.T) {
	h := testHandlers()
	got, err := h.compileChart(context.Background(), contracts.CompileChartInput{
		Spec: contracts.ChartSpec{
			Type:   "bar",
			Title:  "Quarterly",
			Labels: []string{"A", "B", "C"},
			Series: []contracts.ChartSeries{{Values: []float64{3, 7, 5}}},
		},
	})
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	if got.Structured.AssetID == "" {
		t.Fatal("compileChart: empty assetId")
	}
	if string(got.Structured.Node.AssetID) != got.Structured.AssetID {
		t.Fatalf("node AssetID %q != AssetID %q", got.Structured.Node.AssetID, got.Structured.AssetID)
	}
	if got.Structured.Node.Caption != "Quarterly" {
		t.Fatalf("caption = %q, want Quarterly (defaults to title)", got.Structured.Node.Caption)
	}
	a, ok := h.deps.Assets.Get(got.Structured.AssetID)
	if !ok {
		t.Fatal("rasterized chart not stored as an asset")
	}
	if a.MIME != "image/png" || len(a.Bytes) == 0 {
		t.Fatalf("stored asset wrong: mime=%q bytes=%d", a.MIME, len(a.Bytes))
	}
}

func TestCompileChartInvalidSpecErrors(t *testing.T) {
	h := testHandlers()
	if _, err := h.compileChart(context.Background(), contracts.CompileChartInput{
		Spec: contracts.ChartSpec{Type: "bogus", Series: []contracts.ChartSeries{{Values: []float64{1}}}},
	}); err == nil {
		t.Fatal("want error for invalid chart type")
	}
}

// brandSoul returns a Deckard-White-derived soul with a distinct accent
// palette, so brand-styled chart tests can assert on a resolvable, non-default
// palette without hand-rolling a whole soul.
func brandSoul(id string) *soul.Soul {
	s := soul.DeckardWhite()
	s.ID = id
	s.Theme = s.Theme.Clone()
	s.Theme.Colors.Surfaces[pptx.ColorAccent] = "2563EB"
	s.Theme.Colors.Surfaces[pptx.ColorAccentAlt] = "1D4ED8"
	s.Theme.Colors.Surfaces[pptx.ColorAccentWarm] = "F97316"
	s.Theme.Colors.Surfaces[pptx.ColorInfo] = "0EA5E9"
	s.Theme.Colors.Surfaces[pptx.ColorSuccess] = "16A34A"
	s.Theme.Colors.Surfaces[pptx.ColorWarning] = "EAB308"
	s.Theme.Colors.Surfaces[pptx.ColorError] = "DC2626"
	return s
}

func chartInput(soulID string) contracts.CompileChartInput {
	return contracts.CompileChartInput{
		Spec: contracts.ChartSpec{
			Type:   "bar",
			Title:  "Quarterly",
			Labels: []string{"A", "B", "C"},
			Series: []contracts.ChartSeries{{Values: []float64{3, 7, 5}}},
		},
		SoulID: soulID,
	}
}

// TestCompileChartEmptySoulIDByteIdentical asserts the pre-R14.2 path (no
// SoulID) renders byte-identical to the default rasterization, with no
// palette warning — the hard byte-identity boundary from the R14.2 spec.
func TestCompileChartEmptySoulIDByteIdentical(t *testing.T) {
	h := testHandlers()
	got, err := h.compileChart(context.Background(), chartInput(""))
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	if len(got.Structured.Warnings) != 0 {
		t.Fatalf("empty SoulID: got warnings %v, want none", got.Structured.Warnings)
	}
	a, ok := h.deps.Assets.Get(got.Structured.AssetID)
	if !ok {
		t.Fatal("chart asset not stored")
	}
	want, err := raster.RasterizeChart(raster.ChartSpec{
		Type:   raster.ChartBar,
		Title:  "Quarterly",
		Labels: []string{"A", "B", "C"},
		Series: []raster.Series{{Values: []float64{3, 7, 5}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Bytes, want) {
		t.Fatal("compileChart with empty SoulID is not byte-identical to the default (nil Style) rasterization")
	}
}

// TestCompileChartWithSoulAppliesBrandPalette asserts a resolved soul's
// accent palette drives the rasterized series colors, and that the palette
// the handler built equals the soul's resolved accent sequence.
func TestCompileChartWithSoulAppliesBrandPalette(t *testing.T) {
	h := testHandlers()
	brand := brandSoul("acme-brand")
	if err := h.deps.Souls.Put(brand); err != nil {
		t.Fatalf("seed brand soul: %v", err)
	}

	got, err := h.compileChart(context.Background(), chartInput("acme-brand"))
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	if len(got.Structured.Warnings) != 0 {
		t.Fatalf("known SoulID: got warnings %v, want none", got.Structured.Warnings)
	}
	a, ok := h.deps.Assets.Get(got.Structured.AssetID)
	if !ok || len(a.Bytes) == 0 {
		t.Fatal("brand-styled chart asset not stored")
	}

	wantPalette := []string{
		string(brand.Theme.ResolveColor(pptx.ColorAccent)),
		string(brand.Theme.ResolveColor(pptx.ColorAccentAlt)),
		string(brand.Theme.ResolveColor(pptx.ColorAccentWarm)),
		string(brand.Theme.ResolveColor(pptx.ColorInfo)),
		string(brand.Theme.ResolveColor(pptx.ColorSuccess)),
		string(brand.Theme.ResolveColor(pptx.ColorWarning)),
		string(brand.Theme.ResolveColor(pptx.ColorError)),
	}
	got2, ok := h.deps.Souls.Get("acme-brand")
	if !ok {
		t.Fatal("brand soul not resolvable")
	}
	if gotPalette := brandSeriesPalette(got2); !reflect.DeepEqual(gotPalette, wantPalette) {
		t.Fatalf("brandSeriesPalette = %v, want %v", gotPalette, wantPalette)
	}

	// Two different souls must resolve to two different palettes and thus
	// two different rasterized outputs.
	defaultGot, err := h.compileChart(context.Background(), chartInput(""))
	if err != nil {
		t.Fatalf("compileChart(default): %v", err)
	}
	defaultAsset, _ := h.deps.Assets.Get(defaultGot.Structured.AssetID)
	if bytes.Equal(a.Bytes, defaultAsset.Bytes) {
		t.Fatal("brand-styled chart bytes equal the default palette's bytes")
	}
}

// TestCompileChartTwoSoulsDifferentPalettes asserts two distinct souls
// produce palette-distinct rasterized charts for the identical spec.
func TestCompileChartTwoSoulsDifferentPalettes(t *testing.T) {
	h := testHandlers()
	soulA := brandSoul("brand-a")
	soulB := brandSoul("brand-b")
	soulB.Theme.Colors.Surfaces[pptx.ColorAccent] = "9333EA"
	if err := h.deps.Souls.Put(soulA); err != nil {
		t.Fatal(err)
	}
	if err := h.deps.Souls.Put(soulB); err != nil {
		t.Fatal(err)
	}

	gotA, err := h.compileChart(context.Background(), chartInput("brand-a"))
	if err != nil {
		t.Fatal(err)
	}
	gotB, err := h.compileChart(context.Background(), chartInput("brand-b"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := h.deps.Assets.Get(gotA.Structured.AssetID)
	b, _ := h.deps.Assets.Get(gotB.Structured.AssetID)
	if bytes.Equal(a.Bytes, b.Bytes) {
		t.Fatal("two different souls produced identical chart bytes")
	}
}

// TestCompileChartSameSoulDeterministic asserts the same spec + soul renders
// byte-identical output across two calls.
func TestCompileChartSameSoulDeterministic(t *testing.T) {
	h := testHandlers()
	if err := h.deps.Souls.Put(brandSoul("acme-brand")); err != nil {
		t.Fatal(err)
	}
	got1, err := h.compileChart(context.Background(), chartInput("acme-brand"))
	if err != nil {
		t.Fatal(err)
	}
	got2, err := h.compileChart(context.Background(), chartInput("acme-brand"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := h.deps.Assets.Get(got1.Structured.AssetID)
	b, _ := h.deps.Assets.Get(got2.Structured.AssetID)
	if !bytes.Equal(a.Bytes, b.Bytes) {
		t.Fatal("same spec + soul rendered different bytes across calls")
	}
}

// TestCompileChartUnknownSoulIDWarns asserts an unresolvable SoulID produces
// a non-fatal warning and still renders (default palette).
func TestCompileChartUnknownSoulIDWarns(t *testing.T) {
	h := testHandlers()
	got, err := h.compileChart(context.Background(), chartInput("does-not-exist"))
	if err != nil {
		t.Fatalf("compileChart: %v", err)
	}
	if len(got.Structured.Warnings) != 1 {
		t.Fatalf("unknown SoulID: got %d warnings, want 1: %v", len(got.Structured.Warnings), got.Structured.Warnings)
	}
	if got.Structured.AssetID == "" {
		t.Fatal("unknown SoulID: chart did not still render")
	}
}

func TestCompileCodeStoresAssetAndReturnsNode(t *testing.T) {
	h := testHandlers()
	got, err := h.compileCode(context.Background(), contracts.CompileCodeInput{
		Code:     "func main() {\n\tprintln(\"hi\")\n}",
		Language: "go",
		Caption:  "main.go",
	})
	if err != nil {
		t.Fatalf("compileCode: %v", err)
	}
	if got.Structured.AssetID == "" {
		t.Fatal("compileCode: empty assetId")
	}
	if string(got.Structured.Node.AssetID) != got.Structured.AssetID {
		t.Fatalf("node AssetID %q != AssetID %q", got.Structured.Node.AssetID, got.Structured.AssetID)
	}
	if got.Structured.Node.Language != "go" || got.Structured.Node.Caption != "main.go" {
		t.Fatalf("node = %#v, want language=go caption=main.go", got.Structured.Node)
	}
	a, ok := h.deps.Assets.Get(got.Structured.AssetID)
	if !ok {
		t.Fatal("rasterized code not stored as an asset")
	}
	if a.MIME != "image/png" || len(a.Bytes) == 0 {
		t.Fatalf("stored asset wrong: mime=%q bytes=%d", a.MIME, len(a.Bytes))
	}
}

func TestCompileCodeEmptyErrors(t *testing.T) {
	h := testHandlers()
	if _, err := h.compileCode(context.Background(), contracts.CompileCodeInput{Code: "   "}); err == nil {
		t.Fatal("want error for empty code")
	}
}

func TestCompileMarkdownReturnsNodes(t *testing.T) {
	h := testHandlers()
	got, err := h.compileMarkdown(context.Background(), contracts.CompileMarkdownInput{
		Markdown: "# Title\n\n- a\n- b",
	})
	if err != nil {
		t.Fatalf("compileMarkdown: %v", err)
	}
	if len(got.Structured.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2 (heading + list)", len(got.Structured.Nodes))
	}
}
