package contracts

// ColorRole names a surface fill color role (mirrors pptx-go's ColorRole;
// the soul/theme resolves it to concrete RGB at render time).
type ColorRole string

// Surface color roles (mirror the define-a-theme skill enum verbatim).
const (
	ColorCanvas     ColorRole = "canvas"
	ColorSurface    ColorRole = "surface"
	ColorSurfaceAlt ColorRole = "surfaceAlt"
	ColorAccent     ColorRole = "accent"
	ColorAccentAlt  ColorRole = "accentAlt"
	ColorAccentWarm ColorRole = "accentWarm"
	ColorSuccess    ColorRole = "success"
	ColorWarning    ColorRole = "warning"
	ColorError      ColorRole = "error"
	ColorInfo       ColorRole = "info"
)

// IsValid reports whether v is one of the closed ColorRole wire values
// (Phase 12 A4).
func (v ColorRole) IsValid() bool { return IsValidEnum(v, AllowedColorRole()) }

// SpaceRole names a spacing token role (mirrors pptx-go's SpaceRole).
type SpaceRole string

// Spacing roles (mirror the define-a-theme skill enum verbatim).
const (
	SpaceXS  SpaceRole = "xs"
	SpaceSM  SpaceRole = "sm"
	SpaceMD  SpaceRole = "md"
	SpaceLG  SpaceRole = "lg"
	SpaceXL  SpaceRole = "xl"
	Space2XL SpaceRole = "2xl"
)

// IsValid reports whether v is one of the closed SpaceRole wire values
// (Phase 12 A4).
func (v SpaceRole) IsValid() bool { return IsValidEnum(v, AllowedSpaceRole()) }

// ElevationRole names a shadow elevation role (mirrors pptx-go's
// ElevationRole).
type ElevationRole string

// Elevation roles (mirror the define-a-theme skill enum verbatim).
const (
	ElevationFlat     ElevationRole = "flat"
	ElevationRaised   ElevationRole = "raised"
	ElevationElevated ElevationRole = "elevated"
)

// IsValid reports whether v is one of the closed ElevationRole wire values
// (Phase 12 A4). Empty string is rejected here; the validator passes "" as
// legal in the optional-with-default semantics it actually uses.
func (v ElevationRole) IsValid() bool { return IsValidEnum(v, AllowedElevationRole()) }
