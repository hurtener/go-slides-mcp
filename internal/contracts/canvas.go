package contracts

// EMUBox is a rectangle in EMU (English Metric Units; 914400 per inch) — the
// coordinate space pptx-go lays out in. The canvas surface scales it to pixels.
type EMUBox struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
	W int64 `json:"w"`
	H int64 `json:"h"`
}

// NodePlacement is one node's resolved geometry on the slide canvas. Computed
// server-side by mirroring pptx-go's deterministic box-stack layout, so the
// editor canvas paints the SAME geometry the export uses (no second layout
// authority re-coded in the browser). Path is the IR path to the node (for
// click-to-select and the edit tools).
type NodePlacement struct {
	// Path is the structural IR path to the node ([] = top level, e.g. [2] or [2,"left",0]).
	Path IRPath `json:"path"`
	// Kind is the node kind discriminator.
	Kind string `json:"kind"`
	// Box is the node's resolved rectangle in EMU.
	Box EMUBox `json:"box"`
}

// SlideLayout is the full canvas geometry for one slide: the canvas size plus
// every node's placement, in deck (soul) coordinates.
type SlideLayout struct {
	// CanvasWidth / CanvasHeight are the slide dimensions in EMU (16:9 by default).
	CanvasWidth  int64 `json:"canvasWidth"`
	CanvasHeight int64 `json:"canvasHeight"`
	// Placements are the per-node boxes in render order (containers before children).
	Placements []NodePlacement `json:"placements,omitempty"`
	// Overflow is true when the stacked content exceeds the body region (the one
	// irreducible divergence — surfaced honestly on-canvas).
	Overflow bool `json:"overflow"`
}

// SoulPalette is the deck soul's resolved colors + fonts, so the canvas paints
// in the deck's actual visual language (matching the export) rather than the
// app-chrome theme.
type SoulPalette struct {
	Canvas        string `json:"canvas"`
	Surface       string `json:"surface"`
	SurfaceAlt    string `json:"surfaceAlt"`
	Accent        string `json:"accent"`
	AccentText    string `json:"accentText"`
	TextPrimary   string `json:"textPrimary"`
	TextSecondary string `json:"textSecondary"`
	TextInverse   string `json:"textInverse"`
	Border        string `json:"border"`
	// HeadingFont / BodyFont / MonoFont are CSS font-family hints from the soul.
	HeadingFont string `json:"headingFont"`
	BodyFont    string `json:"bodyFont"`
	MonoFont    string `json:"monoFont"`
}
