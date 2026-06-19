package soul

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/hurtener/pptx-go/pptx"
)

var nonSlugRunes = regexp.MustCompile(`[^a-z0-9]+`)

// BootstrapParams seeds a complete soul from a small set of structured tokens.
type BootstrapParams struct {
	// Name is the required soul name and the source for the slugified ID.
	Name string
	// Description is an optional one-line summary.
	Description string
	// Accent overrides the accent surface color as a 6-digit hex string without '#'.
	Accent string
	// AccentAlt overrides the alternate accent surface and text colors.
	AccentAlt string
	// AccentWarm overrides the warm accent and warning surface colors.
	AccentWarm string
	// HeadingFont overrides the display and H1-H3 families.
	HeadingFont string
	// BodyFont overrides the H4-H5 and body families.
	BodyFont string
	// MonoFont overrides the mono and code families.
	MonoFont string
}

// Bootstrap seeds a complete soul from p, inheriting every unset token from
// DeckardWhite().
func Bootstrap(p BootstrapParams) (*Soul, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, fmt.Errorf("soul: bootstrap requires a non-empty name")
	}

	s := DeckardWhite().Clone()
	id := slugify(name)
	if id == "" {
		return nil, fmt.Errorf("soul: bootstrap produced an empty id from name %q", p.Name)
	}

	if p.Description != "" {
		s.Description = p.Description
	}
	if p.Accent != "" {
		accent, err := parseHexColor(p.Accent)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accent: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccent] = accent
		s.Theme.Colors.Text[pptx.TextAccent] = derivedTextAccent(accent)
	}
	if p.AccentAlt != "" {
		accentAlt, err := parseHexColor(p.AccentAlt)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accentAlt: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccentAlt] = accentAlt
		s.Theme.Colors.Text[pptx.TextAccentAlt] = derivedTextAccent(accentAlt)
	}
	if p.AccentWarm != "" {
		accentWarm, err := parseHexColor(p.AccentWarm)
		if err != nil {
			return nil, fmt.Errorf("soul: bootstrap accentWarm: %w", err)
		}
		s.Theme.Colors.Surfaces[pptx.ColorAccentWarm] = accentWarm
		s.Theme.Colors.Surfaces[pptx.ColorWarning] = accentWarm
	}
	if p.HeadingFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.HeadingFont, fs.Size, fs.Weight)
		}
		s.Theme.HeadingFont = p.HeadingFont
	}
	if p.BodyFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeH4, pptx.TypeH5, pptx.TypeBody, pptx.TypeBodySmall, pptx.TypeCaption} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.BodyFont, fs.Size, fs.Weight)
		}
		s.Theme.BodyFont = p.BodyFont
	}
	if p.MonoFont != "" {
		for _, role := range []pptx.TypeRole{pptx.TypeMono, pptx.TypeCode} {
			fs := s.Theme.Typography[role]
			setType(s.Theme, role, p.MonoFont, fs.Size, fs.Weight)
		}
	}

	s.ID = id
	s.Name = name
	s.Status = "ready"
	if name != "Deckard White" {
		s.StyleGuide.NorthStar = ""
		s.StyleGuide.Do = nil
		s.StyleGuide.Dont = nil
	}

	return s, nil
}

func slugify(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteByte('-')
	}
	return strings.Trim(nonSlugRunes.ReplaceAllString(b.String(), "-"), "-")
}
