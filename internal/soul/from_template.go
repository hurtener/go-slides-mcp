package soul

import (
	"fmt"
	"strings"

	"github.com/hurtener/pptx-go/pptx"
)

// FromTemplate builds a complete brand soul from a theme extracted from a
// brand .pptx kit (R8.2). The theme's color scheme (theme1.xml dk1/lt1/lt2/
// accent1..3) is already mapped onto pptx.Theme.Colors by the engine codec;
// this wraps it as a soul: clones the theme, slugifies the id from name,
// re-derives a WCAG-legible accent text against the brand canvas (R8.6), and
// seeds deterministic brand-derived extension tokens. A nil theme or empty
// name is a typed error.
func FromTemplate(name, description string, theme *pptx.Theme) (*Soul, error) {
	if theme == nil {
		return nil, fmt.Errorf("soul: from-template requires a non-nil theme")
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("soul: from-template requires a non-empty name")
	}
	id := slugify(trimmedName)
	if id == "" {
		return nil, fmt.Errorf("soul: from-template produced an empty id from name %q", name)
	}

	t := theme.Clone()
	canvas := t.ResolveColor(pptx.ColorCanvas)
	t.Colors.Text[pptx.TextAccent] = legibleAccentText(t.ResolveColor(pptx.ColorAccent), canvas)
	t.Colors.Text[pptx.TextAccentAlt] = legibleAccentText(t.ResolveColor(pptx.ColorAccentAlt), canvas)

	extensions := map[string]string{
		"border":       string(t.ResolveColor(pptx.ColorSurfaceAlt)),
		"borderStrong": string(t.ResolveColor(pptx.ColorSurfaceAlt)),
		"accentSoft":   string(t.ResolveColor(pptx.ColorAccent)),
	}

	return &Soul{
		ID:          id,
		Name:        trimmedName,
		Description: description,
		Status:      "ready",
		Theme:       t,
		StyleGuide: StyleGuide{
			NorthStar: "Render in the brand's own palette and type.",
		},
		Extensions: extensions,
	}, nil
}
