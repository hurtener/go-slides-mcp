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

// SoulOverride is one targeted refine instruction.
type SoulOverride struct {
	// Category is the override family understood by the soul refiner.
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
