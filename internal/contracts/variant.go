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
	// BackgroundRadial fills the slide canvas with a center-out radial
	// gradient (spotlight/vignette) from Background.Stops (or the legacy
	// Background.Gradient pair when Stops is empty). The focal point is
	// centered — an offset is not yet exposed.
	BackgroundRadial BackgroundKind = "radial"
	// BackgroundMesh draws a soft mesh wash: the base canvas fill plus N
	// low-alpha radial glows from Background.Mesh, pooled at caller-chosen
	// anchors over the canvas.
	BackgroundMesh BackgroundKind = "mesh"
)

// GradientStop is one color stop in a multi-stop background gradient (R13.3).
type GradientStop struct {
	// Pos is the stop position along the gradient axis, in [0,1]
	// (0 = start, 1 = end).
	Pos float64 `json:"pos"`
	// Color is the surface color role at this stop. Resolves against the
	// active soul/theme.
	Color ColorRole `json:"color,omitempty"`
}

// MeshGlow is one pooled radial glow in a mesh background (R13.4).
type MeshGlow struct {
	// Anchor is where the glow pools on the slide (its center).
	Anchor Anchor `json:"anchor,omitempty"`
	// Color is the glow's surface color role. Resolves against the active
	// soul/theme.
	Color ColorRole `json:"color,omitempty"`
	// Radius is the glow circle's radius in POINTS; a value <= 0 is skipped.
	Radius float64 `json:"radius,omitempty"`
	// Alpha is the glow center's opacity in [0,1]; 0 = invisible. Keep it
	// low (~0.08-0.15) for a subtle pool; the edge fades to transparent.
	Alpha float64 `json:"alpha,omitempty"`
}

// Scrim is an optional darkening/tinting overlay drawn over a slide's
// background fill so text reads legibly over a photographic or busy
// background (R14.1). It applies over any drawn background kind. nil draws
// nothing — byte-identical to a pre-R14.1 background.
type Scrim struct {
	// Color is the overlay's surface color role (a dark role for a
	// darkening scrim). Resolves against the active soul/theme.
	Color ColorRole `json:"color,omitempty"`
	// Opacity is the overlay's dense-edge opacity, in [0,1]. For a solid
	// scrim the whole overlay carries Opacity; for a gradient scrim the
	// dense edge carries Opacity and the opposite edge is transparent.
	Opacity float64 `json:"opacity,omitempty"`
	// Gradient, when true, draws a transparent→Color linear gradient
	// overlay instead of a flat wash — the classic bottom-heavy caption
	// scrim.
	Gradient bool `json:"gradient,omitempty"`
	// GradientAngle orients a gradient scrim in degrees clockwise from the
	// positive x-axis; 0 defaults to 90° (top transparent, bottom dense).
	GradientAngle int `json:"gradientAngle,omitempty"`
}

// Duotone is an optional two-tone recolor applied to a photographic
// background (R14.1): the photo's shadows map to Shadow and its highlights
// to Highlight. Applies only to kind "asset". nil leaves the photo at its
// natural colors — byte-identical.
type Duotone struct {
	// Shadow is the role the photo's dark tones map to. Resolves against
	// the active soul/theme.
	Shadow ColorRole `json:"shadow,omitempty"`
	// Highlight is the role the photo's light tones map to. Resolves
	// against the active soul/theme.
	Highlight ColorRole `json:"highlight,omitempty"`
}

// Background is a slide's full-bleed background specification. It is drawn
// behind all body content — the lowest layer in z-order. The zero value
// (nil pointer, or Kind == "") draws nothing; all existing slides are
// byte-identical after this field is added to Slide.
type Background struct {
	// Kind selects the fill type: "" (no background), "color" (solid),
	// "gradient" (two-stop linear), "radial" (center-out radial gradient),
	// "mesh" (pooled radial glows), or "asset" (full-bleed picture).
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
	// Stops is an optional multi-stop gradient (2..8 ascending stops in [0,1]) for
	// kind "gradient" or "radial". When non-empty it supersedes Gradient (and is
	// required for a multi-hue "radial" vignette). Empty = the legacy Gradient/Angle
	// path, byte-identical to today. (R13.2/R13.3)
	Stops []GradientStop `json:"stops,omitempty"`
	// Mesh holds the pooled radial glows for kind "mesh" (R13.4), drawn over the
	// base canvas fill in order. Empty draws nothing (absent config).
	Mesh []MeshGlow `json:"mesh,omitempty"`
	// Scrim is an optional darkening/tinting overlay applied over any drawn
	// background kind (R14.1) — used to keep text legible over a photo or
	// busy background. nil draws nothing, byte-identical to today.
	Scrim *Scrim `json:"scrim,omitempty"`
	// Duotone is an optional two-tone recolor of a photographic background
	// (R14.1); applies only when Kind == "asset". nil = natural colors,
	// byte-identical to today.
	Duotone *Duotone `json:"duotone,omitempty"`
}
