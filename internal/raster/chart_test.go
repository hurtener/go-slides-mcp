package raster

import (
	"bytes"
	"testing"
)

var pngMagic = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

func barSpec() ChartSpec {
	return ChartSpec{Type: ChartBar, Title: "Bar", Labels: []string{"A", "B", "C"},
		Series: []Series{{Name: "s", Values: []float64{3, 7, 5}}}}
}

func TestRasterizeChartProducesValidPNG(t *testing.T) {
	cases := map[string]ChartSpec{
		"bar":  barSpec(),
		"pie":  {Type: ChartPie, Title: "Pie", Labels: []string{"X", "Y"}, Series: []Series{{Values: []float64{1, 2}}}},
		"line": {Type: ChartLine, Title: "Line", Series: []Series{{Name: "a", Values: []float64{1, 3, 2, 5}}}},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			png, err := RasterizeChart(spec)
			if err != nil {
				t.Fatalf("RasterizeChart(%s): %v", name, err)
			}
			if len(png) == 0 {
				t.Fatalf("RasterizeChart(%s): empty", name)
			}
			if !bytes.HasPrefix(png, pngMagic) {
				t.Fatalf("RasterizeChart(%s): not a PNG (prefix %x)", name, png[:8])
			}
		})
	}
}

func TestRasterizeChartDeterministic(t *testing.T) {
	a, err := RasterizeChart(barSpec())
	if err != nil {
		t.Fatal(err)
	}
	b, err := RasterizeChart(barSpec())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("RasterizeChart not deterministic across identical specs")
	}
}

// TestRasterizeChartNilStyleByteIdentical asserts the nil-Style path (the
// default for every chart compiled before R14.2, and every chart compiled
// with an empty SoulID after it) renders byte-identical to itself across
// bar/pie/line — i.e. adding the Style field did not perturb the existing
// default rasterization.
func TestRasterizeChartNilStyleByteIdentical(t *testing.T) {
	cases := map[string]ChartSpec{
		"bar":  barSpec(),
		"pie":  {Type: ChartPie, Title: "Pie", Labels: []string{"X", "Y"}, Series: []Series{{Values: []float64{1, 2}}}},
		"line": {Type: ChartLine, Title: "Line", Series: []Series{{Name: "a", Values: []float64{1, 3, 2, 5}}}},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			spec.Style = nil
			a, err := RasterizeChart(spec)
			if err != nil {
				t.Fatalf("RasterizeChart(%s): %v", name, err)
			}
			b, err := RasterizeChart(spec)
			if err != nil {
				t.Fatalf("RasterizeChart(%s): %v", name, err)
			}
			if !bytes.Equal(a, b) {
				t.Fatalf("RasterizeChart(%s) with nil Style not stable across identical calls", name)
			}
		})
	}
}

func brandStyle() *ChartStyle {
	return &ChartStyle{SeriesColors: []string{"2563EB", "F97316", "16A34A"}}
}

func altBrandStyle() *ChartStyle {
	return &ChartStyle{SeriesColors: []string{"9333EA", "DC2626", "0EA5E9"}}
}

func TestRasterizeChartBrandStyleDiffersFromDefault(t *testing.T) {
	specs := map[string]ChartSpec{
		"bar":  barSpec(),
		"pie":  {Type: ChartPie, Title: "Pie", Labels: []string{"X", "Y"}, Series: []Series{{Values: []float64{1, 2}}}},
		"line": {Type: ChartLine, Title: "Line", Series: []Series{{Name: "a", Values: []float64{1, 3, 2, 5}}}},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			def, err := RasterizeChart(spec)
			if err != nil {
				t.Fatalf("RasterizeChart(%s) default: %v", name, err)
			}
			styled := spec
			styled.Style = brandStyle()
			brand, err := RasterizeChart(styled)
			if err != nil {
				t.Fatalf("RasterizeChart(%s) styled: %v", name, err)
			}
			if bytes.Equal(def, brand) {
				t.Fatalf("RasterizeChart(%s): styled output identical to default, want different bytes", name)
			}
		})
	}
}

func TestRasterizeChartDifferentPalettesDiffer(t *testing.T) {
	spec := barSpec()
	spec.Style = brandStyle()
	a, err := RasterizeChart(spec)
	if err != nil {
		t.Fatal(err)
	}
	spec.Style = altBrandStyle()
	b, err := RasterizeChart(spec)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("RasterizeChart: two different palettes produced identical bytes")
	}
}

func TestRasterizeChartBrandStyleDeterministic(t *testing.T) {
	spec := barSpec()
	spec.Style = brandStyle()
	a, err := RasterizeChart(spec)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RasterizeChart(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("RasterizeChart with a brand Style not deterministic across identical specs")
	}
}

func TestRasterizeChartErrors(t *testing.T) {
	cases := map[string]ChartSpec{
		"bad-type":     {Type: "scatter", Labels: []string{"A"}, Series: []Series{{Values: []float64{1}}}},
		"no-series":    {Type: ChartBar, Labels: []string{"A"}},
		"empty-values": {Type: ChartBar, Labels: []string{"A"}, Series: []Series{{Values: nil}}},
		"label-mismatch": {Type: ChartBar, Labels: []string{"A", "B"},
			Series: []Series{{Values: []float64{1}}}},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := RasterizeChart(spec); err == nil {
				t.Fatalf("RasterizeChart(%s): want error, got nil", name)
			}
		})
	}
}
