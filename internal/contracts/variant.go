package contracts

// Variant selects a named theme variant for a slide. VariantDark renders the
// slide with a dark canvas and light text — the engine derives a legible dark
// palette from the active soul automatically. Omitting the field (or using the
// empty string) is identical to "light"; byte-identical output is guaranteed
// for all pre-existing slides.
type Variant string

const (
	// VariantLight (default) renders the slide with the soul's base palette.
	// This is the backward-compatible default when the field is absent.
	VariantLight Variant = "light"
	// VariantDark renders the slide with a dark canvas and light text. The
	// engine derives a deterministic dark palette from the active soul —
	// accent and semantic roles (error, warning, success, info, accent*) are
	// preserved so brand identity survives the swap. Use for section-divider
	// slides that need a visual break between content sections.
	VariantDark Variant = "dark"
)

// BackgroundKind selects the fill type for a slide's full-bleed background.
// The zero value (empty string) draws nothing — preserving byte-identical
// output for all slides that do not explicitly set a background.
type BackgroundKind string

const (
	// BackgroundNone draws no explicit background; the slide inherits the
	// presentation's default. This is the zero/default value.
	BackgroundNone BackgroundKind = ""
	// BackgroundColor fills the entire slide canvas with a solid color
	// resolved from the active soul via Background.Color.
	BackgroundColor BackgroundKind = "color"
	// BackgroundGradient fills the slide canvas with a two-stop linear
	// gradient between the color roles in Background.Gradient at
	// Background.Angle degrees.
	BackgroundGradient BackgroundKind = "gradient"
	// BackgroundAsset fills the slide canvas with a full-bleed picture
	// resolved via Background.AssetID from the registered asset store.
	BackgroundAsset BackgroundKind = "asset"
)

// Background is a slide's full-bleed background specification. It is drawn
// behind all body content — the lowest layer in z-order. The zero value
// (nil pointer, or Kind == "") draws nothing; all existing slides are
// byte-identical after this field is added to Slide.
type Background struct {
	// Kind selects the fill type: "" (no background), "color" (solid),
	// "gradient" (two-stop linear), or "asset" (full-bleed picture).
	Kind BackgroundKind `json:"kind,omitempty"`
	// Color is the surface color role for a solid-color background
	// (kind == "color"). Resolves against the active soul/theme.
	Color ColorRole `json:"color,omitempty"`
	// Gradient is an ordered list of 0–2 surface color roles for a linear
	// gradient (kind == "gradient"). The first role is the start stop; the
	// second is the end stop. If only one role is given, both stops use it.
	// If empty, both stops resolve to the zero role. All roles resolve
	// against the active soul/theme.
	Gradient []ColorRole `json:"gradient,omitempty"`
	// Angle is the linear gradient angle in degrees clockwise from the
	// positive x-axis (kind == "gradient"). 0° = left-to-right,
	// 90° = top-to-bottom. Common values: 45, 90, 135, 180.
	Angle int `json:"angle,omitempty"`
	// AssetID is the asset reference for a full-bleed picture background
	// (kind == "asset"). Resolved via the registered asset store.
	AssetID string `json:"assetId,omitempty"`
	// GradientName, when set (kind == "gradient"), requests a named brand
	// gradient registered on the active soul (bootstrap_soul's "gradients",
	// R8.5) instead of the legacy Gradient role pair: its own stop list,
	// angle, and linear/radial flag win, and its stops may pin exact brand
	// hues or follow the active light/dark variant. A name not found on the
	// soul renders without a background fill (a warning is recorded) rather
	// than failing. Empty (the default) uses the legacy Gradient/Angle path,
	// byte-identical to today.
	GradientName string `json:"gradientName,omitempty"`
}
