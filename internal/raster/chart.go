package raster

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// deckardAccent is the Deckard White teal applied to bars/lines for on-brand output.
var deckardAccent = drawing.ColorFromHex("3B9C94")

// ChartType selects the chart family rasterized by RasterizeChart.
type ChartType string

// Supported chart types (matches the compile_chart input contract).
const (
	ChartBar  ChartType = "bar"
	ChartLine ChartType = "line"
	ChartPie  ChartType = "pie"
)

// Series is one named data series with parallel numeric values.
type Series struct {
	// Name is the series label; used as a legend entry for line charts.
	Name string
	// Values is the numeric data for this series.
	Values []float64
}

// ChartSpec is the typed input for RasterizeChart. The compile_chart tool
// mirrors this struct on its input contract; ToolDeps.Assets holds the
// produced PNG bytes, and an IR Chart node references its asset ID.
type ChartSpec struct {
	// Type selects the chart family: bar, line, or pie.
	Type ChartType `json:"type"`
	// Title is the chart title rendered above the plot (optional).
	Title string `json:"title,omitempty"`
	// Labels are the per-data-point category labels (X axis for bar/line;
	// slice labels for pie).
	Labels []string `json:"labels,omitempty"`
	// Series is one or more parallel data series. For bar/pie, only the first
	// series is plotted (chart.Value{Label, Value}); for line, every series
	// renders as its own continuous line.
	Series []Series `json:"series,omitempty"`
}

// rasterWidth / rasterHeight pin the canvas to a deterministic 1200x720 PNG.
// Phase 7B (the reference unit) — the V2 renderer can re-skin, scale, and
// vectorize; V1 is one fixed raster size per spec (V1 charts are image-shapes
// — D-004).
const (
	rasterWidth  = 1200
	rasterHeight = 720
)

// RasterizeChart turns a ChartSpec into a deterministic PNG via the pure-Go
// go-chart library (CGo-free). It returns an error for an invalid spec
// rather than panicking. Validation rules:
//   - Type must be one of bar, line, pie (case-sensitive).
//   - Series must contain at least one series with at least one value.
//   - For bar/pie, len(Labels) must match len(Series[0].Values).
//
// All chart types render deterministically to the same PNG for the same
// spec: the library is pure-Go, the canvas is fixed, no time-derived IDs.
func RasterizeChart(spec ChartSpec) ([]byte, error) {
	if err := validateChartSpec(spec); err != nil {
		return nil, err
	}
	r, err := buildChartRenderer(spec)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := r.Render(chart.PNG, &buf); err != nil {
		return nil, fmt.Errorf("render chart: %w", err)
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("chart renderer produced empty output")
	}
	return buf.Bytes(), nil
}

func validateChartSpec(spec ChartSpec) error {
	switch spec.Type {
	case ChartBar, ChartLine, ChartPie:
	default:
		return fmt.Errorf("chart type %q: must be one of bar, line, pie", spec.Type)
	}
	if len(spec.Series) == 0 {
		return fmt.Errorf("chart spec must include at least one series")
	}
	for i, s := range spec.Series {
		if len(s.Values) == 0 {
			return fmt.Errorf("series %d (%q) must include at least one value", i, s.Name)
		}
	}
	if spec.Type == ChartBar || spec.Type == ChartPie {
		first := spec.Series[0].Values
		if len(spec.Labels) != len(first) {
			return fmt.Errorf("chart spec: labels (%d) must match first series values (%d)", len(spec.Labels), len(first))
		}
	}
	return nil
}

func buildChartRenderer(spec ChartSpec) (interface {
	Render(chart.RendererProvider, io.Writer) error
}, error) {
	switch spec.Type {
	case ChartBar:
		return barChartFor(spec), nil
	case ChartPie:
		return pieChartFor(spec), nil
	case ChartLine:
		c, err := lineChartFor(spec)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return nil, fmt.Errorf("chart type %q: unsupported", spec.Type)
}

// barChartFor, pieChartFor, lineChartFor are unexported builders; the
// dispatch in buildChartRenderer is what callers see. Each returns a type
// with a Render method that accepts (chart.RendererProvider, *bytes.Buffer).

func barChartFor(spec ChartSpec) *chart.BarChart {
	values := make([]chart.Value, 0, len(spec.Labels))
	for i, label := range spec.Labels {
		values = append(values, chart.Value{
			Label: label,
			Value: spec.Series[0].Values[i],
			Style: chart.Style{FillColor: deckardAccent, StrokeColor: deckardAccent},
		})
	}
	return &chart.BarChart{
		Title:    spec.Title,
		Width:    rasterWidth,
		Height:   rasterHeight,
		BarWidth: 60,
		Bars:     values,
	}
}

func pieChartFor(spec ChartSpec) *chart.PieChart {
	values := make([]chart.Value, 0, len(spec.Labels))
	for i, label := range spec.Labels {
		values = append(values, chart.Value{Label: label, Value: spec.Series[0].Values[i]})
	}
	return &chart.PieChart{
		Title:  spec.Title,
		Width:  rasterWidth,
		Height: rasterHeight,
		Values: values,
	}
}

func lineChartFor(spec ChartSpec) (*chart.Chart, error) {
	lines := make([]chart.Series, 0, len(spec.Series))
	for _, s := range spec.Series {
		xv := make([]float64, len(s.Values))
		for i := range xv {
			xv[i] = float64(i)
		}
		lines = append(lines, chart.ContinuousSeries{
			Name:    strings.TrimSpace(s.Name),
			XValues: xv,
			YValues: append([]float64(nil), s.Values...),
		})
	}
	return &chart.Chart{
		Title:  spec.Title,
		Width:  rasterWidth,
		Height: rasterHeight,
		Series: lines,
	}, nil
}
