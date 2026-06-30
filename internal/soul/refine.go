package soul

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TokenOverride targets one soul token category/name pair with a string value.
type TokenOverride struct {
	// Category is the token family: surface, text, space, radius, or extension.
	Category string
	// Token is the token name within the category.
	Token string
	// Value is the string form to apply: hex for colors/extensions, points for space/radius.
	Value string
}

// Refine returns a clone of s with the overrides applied.
func Refine(s *Soul, overrides []TokenOverride) (*Soul, error) {
	if s == nil {
		return nil, fmt.Errorf("soul: refine requires a non-nil soul")
	}

	clone := s.Clone()
	var errs []error
	for _, override := range overrides {
		if err := applyOverride(clone, override); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return clone, nil
}

func applyOverride(s *Soul, override TokenOverride) error {
	category := strings.TrimSpace(override.Category)
	token := strings.TrimSpace(override.Token)
	value := strings.TrimSpace(override.Value)

	switch category {
	case "surface":
		role, ok := surfaceRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown surface token %q", token)
		}
		rgb, err := parseHexColor(value)
		if err != nil {
			return fmt.Errorf("soul: surface %q: %w", token, err)
		}
		s.Theme.Colors.Surfaces[role] = rgb
		return nil
	case "text":
		role, ok := textRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown text token %q", token)
		}
		rgb, err := parseHexColor(value)
		if err != nil {
			return fmt.Errorf("soul: text %q: %w", token, err)
		}
		s.Theme.Colors.Text[role] = rgb
		return nil
	case "space":
		role, ok := spaceRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown space token %q", token)
		}
		pt, err := parsePointValue(value)
		if err != nil {
			return fmt.Errorf("soul: space %q: %w", token, err)
		}
		s.Theme.Spacing[role] = pptx.Pt(pt)
		return nil
	case "radius":
		role, ok := radiusRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown radius token %q", token)
		}
		pt, err := parsePointValue(value)
		if err != nil {
			return fmt.Errorf("soul: radius %q: %w", token, err)
		}
		s.Theme.Radii[role] = pptx.Pt(pt)
		return nil
	case "extension":
		if token == "" {
			return fmt.Errorf("soul: unknown extension token %q", token)
		}
		rgb, err := parseHexColor(value)
		if err != nil {
			return fmt.Errorf("soul: extension %q: %w", token, err)
		}
		if s.Extensions == nil {
			s.Extensions = make(map[string]string)
		}
		s.Extensions[token] = string(rgb)
		return nil
	case "darkSurface":
		role, ok := surfaceRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown dark surface token %q", token)
		}
		rgb, err := parseHexColor(value)
		if err != nil {
			return fmt.Errorf("soul: darkSurface %q: %w", token, err)
		}
		ensureDarkColors(s).Surfaces[role] = rgb
		return nil
	case "darkText":
		role, ok := textRole(token)
		if !ok {
			return fmt.Errorf("soul: unknown dark text token %q", token)
		}
		rgb, err := parseHexColor(value)
		if err != nil {
			return fmt.Errorf("soul: darkText %q: %w", token, err)
		}
		ensureDarkColors(s).Text[role] = rgb
		return nil
	default:
		return fmt.Errorf("soul: unknown override category %q", category)
	}
}

// ensureDarkColors lazily allocates s.Theme.DarkColors (and its maps) and
// returns it. The engine has its own unexported equivalent; the soul package
// defines its own since it cannot call the engine's.
func ensureDarkColors(s *Soul) *pptx.DarkPalette {
	if s.Theme.DarkColors == nil {
		s.Theme.DarkColors = &pptx.DarkPalette{
			Surfaces: make(map[pptx.ColorRole]pptx.RGB),
			Text:     make(map[pptx.TextColorRole]pptx.RGB),
		}
	}
	if s.Theme.DarkColors.Surfaces == nil {
		s.Theme.DarkColors.Surfaces = make(map[pptx.ColorRole]pptx.RGB)
	}
	if s.Theme.DarkColors.Text == nil {
		s.Theme.DarkColors.Text = make(map[pptx.TextColorRole]pptx.RGB)
	}
	return s.Theme.DarkColors
}

func parseHexColor(s string) (pptx.RGB, error) {
	value := strings.TrimSpace(s)
	if len(value) != 6 {
		return "", fmt.Errorf("invalid hex color %q", s)
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return "", fmt.Errorf("invalid hex color %q", s)
		}
	}
	return pptx.RGB(strings.ToUpper(value)), nil
}

func parsePointValue(s string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid point value %q", s)
	}
	return value, nil
}

// legibleAccentText derives a WCAG-contrast-aware accent text color: accent
// nudged (hue-preserving, deterministic) until it clears 4.5:1 against bg, or
// returned unchanged if it already does. See scene.LegibleTextOn (D-026): a
// mechanism, not an automatic render-path behavior — the soul calls it once
// per bootstrap to derive a legible accent text color per variant.
func legibleAccentText(accent, bg pptx.RGB) pptx.RGB {
	return scene.LegibleTextOn(accent, bg, 45)
}

func surfaceRole(token string) (pptx.ColorRole, bool) {
	switch token {
	case "canvas":
		return pptx.ColorCanvas, true
	case "surface":
		return pptx.ColorSurface, true
	case "surfaceAlt":
		return pptx.ColorSurfaceAlt, true
	case "accent":
		return pptx.ColorAccent, true
	case "accentAlt":
		return pptx.ColorAccentAlt, true
	case "accentWarm":
		return pptx.ColorAccentWarm, true
	case "success":
		return pptx.ColorSuccess, true
	case "warning":
		return pptx.ColorWarning, true
	case "error":
		return pptx.ColorError, true
	case "info":
		return pptx.ColorInfo, true
	default:
		return 0, false
	}
}

func textRole(token string) (pptx.TextColorRole, bool) {
	switch token {
	case "primary":
		return pptx.TextPrimary, true
	case "secondary":
		return pptx.TextSecondary, true
	case "tertiary":
		return pptx.TextTertiary, true
	case "inverse":
		return pptx.TextInverse, true
	case "muted":
		return pptx.TextMuted, true
	case "accent":
		return pptx.TextAccent, true
	case "accentAlt":
		return pptx.TextAccentAlt, true
	case "success":
		return pptx.TextSuccess, true
	case "warning":
		return pptx.TextWarning, true
	case "error":
		return pptx.TextError, true
	default:
		return 0, false
	}
}

func spaceRole(token string) (pptx.SpaceRole, bool) {
	switch token {
	case "xs":
		return pptx.SpaceXS, true
	case "sm":
		return pptx.SpaceSM, true
	case "md":
		return pptx.SpaceMD, true
	case "lg":
		return pptx.SpaceLG, true
	case "xl":
		return pptx.SpaceXL, true
	case "2xl":
		return pptx.Space2XL, true
	default:
		return 0, false
	}
}

func radiusRole(token string) (pptx.RadiusRole, bool) {
	switch token {
	case "none":
		return pptx.RadiusNone, true
	case "sm":
		return pptx.RadiusSM, true
	case "md":
		return pptx.RadiusMD, true
	case "lg":
		return pptx.RadiusLG, true
	case "full":
		return pptx.RadiusFull, true
	default:
		return 0, false
	}
}
