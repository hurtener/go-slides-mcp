package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/markdown"
	"github.com/hurtener/go-slides-mcp/internal/raster"
)

// compileMarkdown parses markdown source into Deckard IR leaf nodes.
func (h *handlers) compileMarkdown(_ context.Context, in contracts.CompileMarkdownInput) (tool.Result[contracts.CompileMarkdownOutput], error) {
	nodes, warnings := markdown.Parse(in.Markdown)
	out := contracts.CompileMarkdownOutput{Nodes: nodes, Warnings: warnings}
	return tool.Result[contracts.CompileMarkdownOutput]{
		Text:       agentText(fmt.Sprintf("Compiled markdown into %d node(s). Pass these as a slide's \"nodes\" to add_slide:", len(nodes)), out.Nodes),
		Structured: out,
	}, nil
}

// compileCode rasterizes source code to a PNG (pure Go, Go Mono), stores it as
// an asset, and returns a code_block IR node referencing it by asset id.
func (h *handlers) compileCode(_ context.Context, in contracts.CompileCodeInput) (tool.Result[contracts.CompileCodeOutput], error) {
	png, err := raster.RasterizeCode(in.Code, in.Language)
	if err != nil {
		return tool.Result[contracts.CompileCodeOutput]{}, fmt.Errorf("compile_code: %w", err)
	}
	asset, err := h.deps.Assets.Put("code.png", "image/png", png)
	if err != nil {
		return tool.Result[contracts.CompileCodeOutput]{}, fmt.Errorf("compile_code: store asset: %w", err)
	}

	out := contracts.CompileCodeOutput{
		Node: contracts.CodeBlock{
			AssetID:  contracts.AssetID(asset.ID),
			Language: in.Language,
			Caption:  in.Caption,
		},
		AssetID: asset.ID,
	}
	return tool.Result[contracts.CompileCodeOutput]{
		Text:       agentText(fmt.Sprintf("Rasterized %s code. Embed this node in a slide's \"nodes\":", in.Language), out.Node),
		Structured: out,
	}, nil
}

// compileChart rasterizes a chart spec to a PNG (pure Go), stores it as an asset,
// and returns a chart IR node referencing it by asset id.
func (h *handlers) compileChart(_ context.Context, in contracts.CompileChartInput) (tool.Result[contracts.CompileChartOutput], error) {
	spec := raster.ChartSpec{
		Type:   raster.ChartType(in.Spec.Type),
		Title:  in.Spec.Title,
		Labels: in.Spec.Labels,
		Series: make([]raster.Series, len(in.Spec.Series)),
	}
	for i, s := range in.Spec.Series {
		spec.Series[i] = raster.Series{Name: s.Name, Values: s.Values}
	}

	png, err := raster.RasterizeChart(spec)
	if err != nil {
		return tool.Result[contracts.CompileChartOutput]{}, fmt.Errorf("compile_chart: %w", err)
	}
	asset, err := h.deps.Assets.Put("chart.png", "image/png", png)
	if err != nil {
		return tool.Result[contracts.CompileChartOutput]{}, fmt.Errorf("compile_chart: store asset: %w", err)
	}

	caption := in.Caption
	if caption == "" {
		caption = in.Spec.Title
	}
	out := contracts.CompileChartOutput{
		Node:    contracts.Chart{AssetID: contracts.AssetID(asset.ID), Caption: caption},
		AssetID: asset.ID,
	}
	return tool.Result[contracts.CompileChartOutput]{
		Text:       agentText(fmt.Sprintf("Compiled %s chart. Embed this node in a slide's \"nodes\":", spec.Type), out.Node),
		Structured: out,
	}, nil
}
