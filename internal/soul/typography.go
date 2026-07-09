package soul

import (
	"strings"

	"github.com/hurtener/go-slides-mcp/internal/soul/fonts"
	"github.com/hurtener/pptx-go/pptx"
)

const defaultAvgCharWidth = 0.5

// applyTypographyDefaults fills each type role's estimator and fallback fields
// from the resolved family. It preserves any explicit AvgCharWidth/Fallback a
// theme already carries.
func applyTypographyDefaults(t *pptx.Theme) {
	if t == nil {
		return
	}
	for role, fs := range t.Typography {
		if fs.AvgCharWidth <= 0 {
			fs.AvgCharWidth = avgCharWidthFor(role, fs.Family)
		}
		if len(fs.Fallback) == 0 {
			fs.Fallback = fallbackFamilies(role, fs.Family)
		}
		t.Typography[role] = fs
	}
}

func avgCharWidthFor(role pptx.TypeRole, family string) float64 {
	if avg, ok := fonts.AvgCharWidth(family); ok && avg > 0 {
		return avg
	}
	key := strings.ToLower(strings.TrimSpace(family))
	switch {
	case key == "":
		return defaultAvgCharWidth
	case isMonoFamily(key):
		return 0.60
	case strings.Contains(key, "playfair"):
		return 0.55
	case strings.Contains(key, "lora"):
		return 0.52
	case strings.Contains(key, "georgia"), strings.Contains(key, "garamond"), strings.Contains(key, "times"), strings.Contains(key, "serif"):
		return 0.53
	case strings.Contains(key, "inter"):
		return 0.50
	case strings.Contains(key, "arial"), strings.Contains(key, "helvetica"), strings.Contains(key, "roboto"), strings.Contains(key, "sans"):
		return 0.50
	case defaultRoleClass(role) == typeClassMono:
		return 0.60
	default:
		return defaultAvgCharWidth
	}
}

type typeClass int

const (
	typeClassSans typeClass = iota
	typeClassSerif
	typeClassMono
)

func fallbackFamilies(role pptx.TypeRole, family string) []string {
	want := classifyFamily(role, family)
	var candidates []string
	switch want {
	case typeClassMono:
		candidates = []string{"Consolas", "Menlo", "Courier New"}
	case typeClassSerif:
		candidates = []string{"Lora", "Playfair Display", "Georgia"}
	default:
		candidates = []string{"Inter", "Calibri", "Arial"}
	}
	return uniqueFallbacks(strings.TrimSpace(family), candidates)
}

func classifyFamily(role pptx.TypeRole, family string) typeClass {
	key := strings.ToLower(strings.TrimSpace(family))
	switch {
	case isMonoFamily(key):
		return typeClassMono
	case strings.Contains(key, "playfair"), strings.Contains(key, "lora"), strings.Contains(key, "georgia"), strings.Contains(key, "garamond"), strings.Contains(key, "times"), strings.Contains(key, "serif"):
		return typeClassSerif
	case strings.Contains(key, "inter"), strings.Contains(key, "arial"), strings.Contains(key, "helvetica"), strings.Contains(key, "roboto"), strings.Contains(key, "sans"):
		return typeClassSans
	default:
		return defaultRoleClass(role)
	}
}

func defaultRoleClass(role pptx.TypeRole) typeClass {
	switch role {
	case pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3:
		return typeClassSerif
	case pptx.TypeMono, pptx.TypeCode:
		return typeClassMono
	default:
		return typeClassSans
	}
}

func isMonoFamily(key string) bool {
	return strings.Contains(key, "mono") || strings.Contains(key, "consolas") || strings.Contains(key, "menlo") || strings.Contains(key, "courier")
}

func uniqueFallbacks(family string, candidates []string) []string {
	seen := make(map[string]bool, len(candidates)+1)
	current := strings.ToLower(strings.TrimSpace(family))
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if key == current || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	return out
}
