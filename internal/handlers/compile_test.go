package handlers

import (
	"context"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
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
