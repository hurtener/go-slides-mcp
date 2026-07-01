package contracts

// ChartSeries is one named data series for compile_chart.
type ChartSeries struct {
	// Name labels the series (shown for line charts).
	Name string `json:"name,omitempty"`
	// Values are the numeric data points.
	Values []float64 `json:"values"`
}

// ChartSpec is the declarative chart description for compile_chart. The server
// rasterizes it to a PNG (pure Go, no Chromium) and stores it as an asset.
type ChartSpec struct {
	// Type selects the chart family: "bar", "line", or "pie".
	Type string `json:"type"`
	// Title is the chart title (optional).
	Title string `json:"title,omitempty"`
	// Labels are per-point category labels (required for bar and pie).
	Labels []string `json:"labels,omitempty"`
	// Series are the data series (bar/pie use the first; line plots each).
	Series []ChartSeries `json:"series"`
}

// CompileChartInput is the model-facing input for compile_chart.
type CompileChartInput struct {
	// Spec is the chart to rasterize.
	Spec ChartSpec `json:"spec"`
	// Caption overrides the chart node caption (defaults to the spec title).
	Caption string `json:"caption,omitempty"`
	// SoulID resolves this soul's accent palette to brand-style the chart's
	// series colors (R14.2). Empty = the built-in Deckard White default
	// rasterization (byte-identical to a chart compiled before soul-aware
	// styling).
	SoulID string `json:"soulId,omitempty"`
}

// CompileChartOutput returns a ready-to-use chart IR node plus its asset id.
type CompileChartOutput struct {
	// Node is the chart IR node referencing the rasterized image by asset id.
	Node Chart `json:"node"`
	// AssetID is the stored PNG's id ("asset://...").
	AssetID string `json:"assetId"`
	// Warnings are non-fatal rasterization notes.
	Warnings []string `json:"warnings,omitempty"`
}

// CompileCodeInput is the model-facing input for compile_code. The server
// rasterizes the source to a PNG (pure-Go Go Mono font, no Chromium) and stores
// it as an asset, returning a ready-to-use code_block IR node.
type CompileCodeInput struct {
	// Code is the source text to rasterize.
	Code string `json:"code"`
	// Language labels the snippet and is drawn as a small header badge (optional).
	Language string `json:"language,omitempty"`
	// Caption overrides the code_block node caption (optional).
	Caption string `json:"caption,omitempty"`
}

// CompileCodeOutput returns a ready-to-use code_block IR node plus its asset id.
type CompileCodeOutput struct {
	// Node is the code_block IR node referencing the rasterized image by asset id.
	Node CodeBlock `json:"node"`
	// AssetID is the stored PNG's id ("asset://...").
	AssetID string `json:"assetId"`
	// Warnings are non-fatal rasterization notes.
	Warnings []string `json:"warnings,omitempty"`
}

// CompileMarkdownInput is the model-facing input for compile_markdown.
type CompileMarkdownInput struct {
	// Markdown is the source text to parse into IR leaf nodes.
	Markdown string `json:"markdown"`
}

// CompileMarkdownOutput returns the parsed IR leaf nodes and any parse warnings.
type CompileMarkdownOutput struct {
	// Nodes are the parsed slide IR leaf nodes (headings, lists, quotes, prose).
	Nodes []SlideNode `json:"nodes"`
	// Warnings are non-fatal parse notes.
	Warnings []string `json:"warnings,omitempty"`
}
