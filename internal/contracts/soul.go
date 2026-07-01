package contracts

// SoulStatus filters or reports a soul lifecycle state.
type SoulStatus string

// TokenLayer identifies which theme layer produced one flattened design token.
type TokenLayer string

const (
	// TokenLayerSurface marks a surface color token.
	TokenLayerSurface TokenLayer = "surface"
	// TokenLayerText marks a text color token.
	TokenLayerText TokenLayer = "text"
	// TokenLayerTypography marks a typography token.
	TokenLayerTypography TokenLayer = "typography"
	// TokenLayerSpacing marks a spacing token.
	TokenLayerSpacing TokenLayer = "spacing"
	// TokenLayerRadius marks a radius token.
	TokenLayerRadius TokenLayer = "radius"
	// TokenLayerElevation marks an elevation token.
	TokenLayerElevation TokenLayer = "elevation"
	// TokenLayerExtension marks a non-native extension token.
	TokenLayerExtension TokenLayer = "extension"
)

// BootstrapSoulInput is the typed input for bootstrap_soul.
type BootstrapSoulInput struct {
	// Name is the required soul name and the source of the stored soul ID.
	Name string `json:"name"`
	// Description is an optional one-line soul summary.
	Description string `json:"description,omitempty"`
	// Accent overrides the primary accent surface token as a six-digit hex string.
	Accent string `json:"accent,omitempty"`
	// AccentAlt overrides the alternate accent token as a six-digit hex string.
	AccentAlt string `json:"accentAlt,omitempty"`
	// AccentWarm overrides the warm accent token as a six-digit hex string.
	AccentWarm string `json:"accentWarm,omitempty"`
	// HeadingFont overrides the display and heading font family.
	HeadingFont string `json:"headingFont,omitempty"`
	// BodyFont overrides the body and subheading font family.
	BodyFont string `json:"bodyFont,omitempty"`
	// MonoFont overrides the mono and code font family.
	MonoFont string `json:"monoFont,omitempty"`
	// Palette is an optional complete color palette covering every surface,
	// text, and extension token in one call. Unset keys inherit Deckard White
	// byte-for-byte; an unknown key is a typed error.
	Palette *BootstrapPalette `json:"palette,omitempty"`
	// DarkPalette is an optional soul-driven VariantDark color override set.
	// Unset leaves every VariantDark slide on the engine's pinned neutral-gray
	// dark default, byte-identical.
	DarkPalette *BootstrapDarkPalette `json:"darkPalette,omitempty"`
	// Gradients is an optional set of named brand gradients (R8.5). Each is
	// registered on the soul under its Name and requested at slide-authoring
	// time by Background.gradientName. Unset/empty leaves the soul with no
	// named gradients, byte-identical to today. Bootstrap-only: there is no
	// refine_soul path for gradients (a structured stop list does not fit the
	// flat category/token/value refine shape).
	Gradients []BootstrapGradient `json:"gradients,omitempty"`
}

// BootstrapGradient is one named brand gradient definition for bootstrap_soul
// (R8.5), requested at slide-authoring time by a Background's gradientName.
type BootstrapGradient struct {
	// Name is the gradient's stable identifier (e.g. "heroDark"), requested
	// by a slide Background's gradientName. Must be non-empty and unique
	// among the gradients in one bootstrap_soul call.
	Name string `json:"name"`
	// Stops is the ordered color-stop list (2..8 stops), each Pos strictly
	// ascending in [0,1] (0 = gradient start, 1 = gradient end).
	Stops []BootstrapGradientStop `json:"stops"`
	// Angle is the linear gradient angle in degrees clockwise from the
	// positive x-axis (0° = left-to-right, 90° = top-to-bottom). Ignored
	// when Radial is true.
	Angle int `json:"angle,omitempty"`
	// Radial selects a radial wash from the slide centre outward instead of
	// a linear gradient; when true, Angle is ignored.
	Radial bool `json:"radial,omitempty"`
}

// BootstrapGradientStop is one color stop within a BootstrapGradient. Exactly
// one of ColorHex or ColorRole must be set per stop.
type BootstrapGradientStop struct {
	// Pos is the stop position along the gradient axis, in [0,1].
	Pos float64 `json:"pos"`
	// ColorHex pins this stop to an exact six-digit hex color (no '#'),
	// unaffected by a light/dark variant swap. Mutually exclusive with
	// ColorRole — set exactly one per stop.
	ColorHex string `json:"colorHex,omitempty"`
	// ColorRole names a surface-role token (canvas, surface, surfaceAlt,
	// accent, accentAlt, accentWarm, success, warning, error, info) whose
	// resolved color follows the active theme/variant. Mutually exclusive
	// with ColorHex — set exactly one per stop.
	ColorRole string `json:"colorRole,omitempty"`
}

// BootstrapDarkPalette is an optional brand dark-mode color override set for
// bootstrap_soul (R8.3). Each map is keyed by the same token names
// refine_soul validates; an unset map leaves the corresponding VariantDark
// roles on the engine's pinned neutral-gray default.
type BootstrapDarkPalette struct {
	// DarkSurfaces maps surface-role tokens to six-digit hex strings for
	// VariantDark slides. Valid keys: canvas, surface, surfaceAlt, accent,
	// accentAlt, accentWarm, success, warning, error, info.
	DarkSurfaces map[string]string `json:"darkSurfaces,omitempty"`
	// DarkText maps text-role tokens to six-digit hex strings for VariantDark
	// slides. Valid keys: primary, secondary, tertiary, inverse, muted,
	// accent, accentAlt, success, warning, error.
	DarkText map[string]string `json:"darkText,omitempty"`
}

// BootstrapPalette is a complete optional brand color palette for
// bootstrap_soul. Each map is keyed by the same token names refine_soul
// validates.
type BootstrapPalette struct {
	// Surfaces maps surface-role tokens to six-digit hex strings. Valid keys:
	// canvas, surface, surfaceAlt, accent, accentAlt, accentWarm, success,
	// warning, error, info.
	Surfaces map[string]string `json:"surfaces,omitempty"`
	// Text maps text-role tokens to six-digit hex strings. Valid keys: primary,
	// secondary, tertiary, inverse, muted, accent, accentAlt, success, warning,
	// error.
	Text map[string]string `json:"text,omitempty"`
	// Extensions maps non-native extension tokens to six-digit hex strings.
	// Valid keys: border, borderStrong, accentSoft.
	Extensions map[string]string `json:"extensions,omitempty"`
}

// BootstrapSoulOutput is the structured result for bootstrap_soul.
type BootstrapSoulOutput struct {
	// SoulID is the stored soul identifier.
	SoulID string `json:"soulId"`
	// Name is the stored soul name.
	Name string `json:"name"`
	// Status is the stored soul lifecycle state.
	Status SoulStatus `json:"status,omitempty"`
	// TokenCount is the number of flattened resolved design tokens in the soul.
	TokenCount int `json:"tokenCount"`
}

// BootstrapSoulFromTemplateInput is the typed input for
// bootstrap_soul_from_template (R8.2): it extracts a complete brand soul from
// a brand .pptx kit's own theme (colors + fonts), so a deck can render in the
// brand's own palette without hand-typing every hex.
type BootstrapSoulFromTemplateInput struct {
	// Name is the required soul name and the source of the stored soul id.
	Name string `json:"name"`
	// Description is an optional one-line soul summary.
	Description string `json:"description,omitempty"`
	// Path is the filesystem path to the brand .pptx kit whose theme (colors +
	// fonts) seeds the soul. The file must exist and be a .pptx.
	Path string `json:"path"`
}

// BootstrapSoulFromTemplateOutput is the structured result for
// bootstrap_soul_from_template.
type BootstrapSoulFromTemplateOutput struct {
	// SoulID is the stored soul identifier.
	SoulID string `json:"soulId"`
	// Name is the stored soul name.
	Name string `json:"name"`
	// Status is the stored soul lifecycle state.
	Status SoulStatus `json:"status,omitempty"`
	// TokenCount is the number of flattened resolved design tokens in the soul.
	TokenCount int `json:"tokenCount"`
	// ExtractedColors summarizes the key resolved brand colors pulled from the
	// template theme (role -> 6-digit hex), for agent review before building.
	ExtractedColors map[string]string `json:"extractedColors,omitempty"`
}

// SoulOverride is one targeted refine instruction.
type SoulOverride struct {
	// Category is the override family understood by the soul refiner: surface,
	// text, space, radius, extension, darkSurface, or darkText. darkSurface and
	// darkText target the same token names as surface/text but apply to the
	// VariantDark color override set (R8.3) instead of the light theme.
	Category string `json:"category"`
	// Token is the token name within the selected category.
	Token string `json:"token"`
	// Value is the string form to apply to the selected token.
	Value string `json:"value"`
}

// RefineSoulInput is the typed input for refine_soul.
type RefineSoulInput struct {
	// SoulID addresses the stored soul to refine.
	SoulID string `json:"soulId"`
	// Overrides is the ordered set of token overrides to apply.
	Overrides []SoulOverride `json:"overrides,omitempty"`
	// Icons is an optional set of brand glyphs to bind to the soul (R14.16):
	// glyph-name -> single-path SVG string. Each SVG is validated against the
	// icon translator's single-path/solid-fill constraints (D-040/D-005); a
	// bad glyph is rejected with a typed error naming it and the call fails
	// before any change is persisted. A bound glyph resolves ahead of the
	// curated set for every Card/Flow/Milestone/etc. icon reference across
	// this soul's renders. Only single-color icons are supported — a
	// duotone/two-tone icon is a pptx-go engine gap, not modeled here.
	Icons map[string]string `json:"icons,omitempty"`
}

// RefineSoulOutput is the structured result for refine_soul.
type RefineSoulOutput struct {
	// SoulID is the refined soul identifier.
	SoulID string `json:"soulId"`
	// Changed is the ordered set of category.token pairs overridden by this call.
	Changed []string `json:"changed,omitempty"`
	// TokenCount is the number of flattened resolved design tokens in the refined soul.
	TokenCount int `json:"tokenCount"`
}

// ListSoulsInput is the typed input for list_souls.
type ListSoulsInput struct {
	// Status filters the list to one lifecycle state when set.
	Status SoulStatus `json:"status,omitempty"`
}

// SoulSummary is the list payload for one stored soul.
type SoulSummary struct {
	// SoulID is the stable soul identifier.
	SoulID string `json:"soulId"`
	// Name is the human-facing soul name.
	Name string `json:"name"`
	// Status is the soul lifecycle state.
	Status SoulStatus `json:"status,omitempty"`
	// TokenCount is the number of flattened resolved design tokens in the soul.
	TokenCount int `json:"tokenCount"`
}

// ListSoulsOutput is the structured result for list_souls.
type ListSoulsOutput struct {
	// Souls is every stored soul summary matching the filter.
	Souls []SoulSummary `json:"souls,omitempty"`
}

// GetSoulInput is the typed input for get_soul.
type GetSoulInput struct {
	// SoulID addresses the stored soul to load.
	SoulID string `json:"soulId"`
	// IncludeStyleGuide requests the soul voice guidance in the response.
	IncludeStyleGuide bool `json:"includeStyleGuide,omitempty"`
}

// SoulStyleGuide is the model-facing soul voice guidance.
type SoulStyleGuide struct {
	// NorthStar is the one-line design intent.
	NorthStar string `json:"northStar,omitempty"`
	// Do lists encouraged authoring behaviors.
	Do []string `json:"do,omitempty"`
	// Dont lists discouraged authoring behaviors.
	Dont []string `json:"dont,omitempty"`
}

// TokenEntry is one flattened design token from a soul theme.
type TokenEntry struct {
	// Name is the stable token name within its layer.
	Name string `json:"name"`
	// Value is the string form of the resolved token value.
	Value string `json:"value"`
	// Layer identifies which token family produced this entry.
	Layer TokenLayer `json:"layer"`
}

// GetSoulOutput is the structured result for get_soul.
type GetSoulOutput struct {
	// SoulID is the loaded soul identifier.
	SoulID string `json:"soulId"`
	// Name is the loaded soul name.
	Name string `json:"name"`
	// Status is the loaded soul lifecycle state.
	Status SoulStatus `json:"status,omitempty"`
	// Description is the optional one-line soul summary.
	Description string `json:"description,omitempty"`
	// Tokens is the flattened token list for the soul theme.
	Tokens []TokenEntry `json:"tokens,omitempty"`
	// StyleGuide is the optional voice guidance for the soul.
	StyleGuide *SoulStyleGuide `json:"styleGuide,omitempty"`
}

// GetDesignTokensInput is the typed input for get_design_tokens.
type GetDesignTokensInput struct {
	// SoulID addresses the stored soul whose flattened tokens should be returned.
	SoulID string `json:"soulId"`
}

// GetDesignTokensOutput is the structured result for get_design_tokens.
type GetDesignTokensOutput struct {
	// Tokens is the flattened token list for the soul theme.
	Tokens []TokenEntry `json:"tokens,omitempty"`
}
