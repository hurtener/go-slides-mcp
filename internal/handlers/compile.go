package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
)

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
		Text:       fmt.Sprintf("Compiled %s chart -> %s (%d bytes).", spec.Type, asset.ID, len(png)),
		Structured: out,
	}, nil
}
