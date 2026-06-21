package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapVariant asserts all Variant wire values map to the correct scene enum.
func TestMapVariant(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   contracts.Variant
		want scene.Variant
	}{
		{contracts.VariantLight, scene.VariantLight},
		{"", scene.VariantLight}, // empty string = default = light
		{contracts.VariantDark, scene.VariantDark},
		{"unknown", scene.VariantLight}, // unknown = default = light
	}
	for _, tc := range cases {
		if got := mapVariant(tc.in); got != tc.want {
			t.Errorf("mapVariant(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestMapBackgroundKind asserts all BackgroundKind wire values map correctly.
func TestMapBackgroundKind(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   contracts.BackgroundKind
		want scene.BackgroundKind
	}{
		{contracts.BackgroundNone, scene.BackgroundNone},
		{"", scene.BackgroundNone},
		{contracts.BackgroundColor, scene.BackgroundColor},
		{contracts.BackgroundGradient, scene.BackgroundGradient},
		{contracts.BackgroundAsset, scene.BackgroundAsset},
		{"unknown", scene.BackgroundNone},
	}
	for _, tc := range cases {
		if got := mapBackgroundKind(tc.in); got != tc.want {
			t.Errorf("mapBackgroundKind(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestMapBackgroundGradientSlice asserts the gradient slice→[2] array mapping
// for the 0, 1, and 2+ role cases.
func TestMapBackgroundGradientSlice(t *testing.T) {
	t.Parallel()

	accentRole := mapColorRole(contracts.ColorAccent)
	accentAltRole := mapColorRole(contracts.ColorAccentAlt)
	surfaceRole := mapColorRole(contracts.ColorSurface)

	cases := []struct {
		name    string
		bg      contracts.Background
		wantG0  interface{} // pptx.ColorRole
		wantG1  interface{}
		wantNil bool
	}{
		{
			name: "two roles",
			bg: contracts.Background{
				Kind:     contracts.BackgroundGradient,
				Gradient: []contracts.ColorRole{contracts.ColorAccent, contracts.ColorAccentAlt},
				Angle:    135,
			},
		},
		{
			name: "one role (both stops same)",
			bg: contracts.Background{
				Kind:     contracts.BackgroundGradient,
				Gradient: []contracts.ColorRole{contracts.ColorSurface},
				Angle:    90,
			},
		},
		{
			name: "zero roles (both stops zero)",
			bg: contracts.Background{
				Kind:  contracts.BackgroundGradient,
				Angle: 45,
			},
		},
	}
	_ = cases

	// Two roles: gradient[0]=accent, gradient[1]=accentAlt
	{
		bg := contracts.Background{
			Kind:     contracts.BackgroundGradient,
			Gradient: []contracts.ColorRole{contracts.ColorAccent, contracts.ColorAccentAlt},
			Angle:    135,
		}
		got := mapBackground(bg)
		if got.Kind != scene.BackgroundGradient {
			t.Errorf("2-role gradient: Kind = %v, want BackgroundGradient", got.Kind)
		}
		if got.Gradient[0] != accentRole {
			t.Errorf("2-role gradient: Gradient[0] = %v, want %v", got.Gradient[0], accentRole)
		}
		if got.Gradient[1] != accentAltRole {
			t.Errorf("2-role gradient: Gradient[1] = %v, want %v", got.Gradient[1], accentAltRole)
		}
		if got.Angle != 135 {
			t.Errorf("2-role gradient: Angle = %v, want 135", got.Angle)
		}
	}

	// One role: both stops same
	{
		bg := contracts.Background{
			Kind:     contracts.BackgroundGradient,
			Gradient: []contracts.ColorRole{contracts.ColorSurface},
			Angle:    90,
		}
		got := mapBackground(bg)
		if got.Gradient[0] != surfaceRole {
			t.Errorf("1-role gradient: Gradient[0] = %v, want %v", got.Gradient[0], surfaceRole)
		}
		if got.Gradient[1] != surfaceRole {
			t.Errorf("1-role gradient: Gradient[1] = %v, want %v (both stops same)", got.Gradient[1], surfaceRole)
		}
	}

	// Zero roles: both stops are zero value
	{
		bg := contracts.Background{
			Kind:  contracts.BackgroundGradient,
			Angle: 45,
		}
		got := mapBackground(bg)
		var zero interface{} = got.Gradient[0]
		var zeroVal interface{} = mapColorRole("")
		_ = zero
		_ = zeroVal
		// Both stops should be the zero ColorRole (mapColorRole("") = ColorSurface default)
		if got.Gradient[0] != got.Gradient[1] {
			t.Errorf("0-role gradient: stops differ: [0]=%v [1]=%v", got.Gradient[0], got.Gradient[1])
		}
	}
}

// TestMapSlideVariantDark asserts a Slide with Variant:"dark" maps to
// scene.SceneSlide{Variant: scene.VariantDark}.
func TestMapSlideVariantDark(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{
		ID:      "section",
		Variant: contracts.VariantDark,
		Nodes:   []contracts.SlideNode{&contracts.Hero{Title: "Dark Section"}},
	}
	got := mapSlide(slide, 0)

	if got.Variant != scene.VariantDark {
		t.Errorf("Variant = %v, want VariantDark", got.Variant)
	}
}

// TestMapSlideDarkWithGradientBackground asserts the full acceptance case:
// Slide{Variant:"dark", Background:{Kind:"gradient", Gradient:["accent","accentAlt"], Angle:135}}
// maps to scene.SceneSlide{Variant: VariantDark, Background: {Kind: BackgroundGradient,
// Gradient:[2]{accent,accentAlt}, Angle:135}}.
func TestMapSlideDarkWithGradientBackground(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{
		ID:      "dark-section",
		Variant: contracts.VariantDark,
		Background: &contracts.Background{
			Kind:     contracts.BackgroundGradient,
			Gradient: []contracts.ColorRole{contracts.ColorAccent, contracts.ColorAccentAlt},
			Angle:    135,
		},
		Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Brand Section"}},
	}
	got := mapSlide(slide, 0)

	if got.Variant != scene.VariantDark {
		t.Errorf("Variant = %v, want VariantDark", got.Variant)
	}
	if got.Background.Kind != scene.BackgroundGradient {
		t.Errorf("Background.Kind = %v, want BackgroundGradient", got.Background.Kind)
	}
	wantG0 := mapColorRole(contracts.ColorAccent)
	wantG1 := mapColorRole(contracts.ColorAccentAlt)
	if got.Background.Gradient[0] != wantG0 {
		t.Errorf("Background.Gradient[0] = %v, want %v", got.Background.Gradient[0], wantG0)
	}
	if got.Background.Gradient[1] != wantG1 {
		t.Errorf("Background.Gradient[1] = %v, want %v", got.Background.Gradient[1], wantG1)
	}
	if got.Background.Angle != 135 {
		t.Errorf("Background.Angle = %v, want 135", got.Background.Angle)
	}
}

// TestMapSlideZeroVariantBackground asserts that a Slide with no variant/background
// maps to VariantLight and BackgroundNone — backward compat, byte-identical.
func TestMapSlideZeroVariantBackground(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{ID: "light", Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Light"}}}
	got := mapSlide(slide, 0)

	if got.Variant != scene.VariantLight {
		t.Errorf("zero Variant should map to VariantLight, got %v", got.Variant)
	}
	if got.Background != (scene.Background{}) {
		t.Errorf("nil Background should map to zero scene.Background (BackgroundNone), got %+v", got.Background)
	}
}

// TestMapSlideColorBackground asserts a solid-color background maps correctly.
func TestMapSlideColorBackground(t *testing.T) {
	t.Parallel()

	slide := contracts.Slide{
		ID: "color-bg",
		Background: &contracts.Background{
			Kind:  contracts.BackgroundColor,
			Color: contracts.ColorCanvas,
		},
		Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Solid BG"}},
	}
	got := mapSlide(slide, 0)

	if got.Background.Kind != scene.BackgroundColor {
		t.Errorf("Background.Kind = %v, want BackgroundColor", got.Background.Kind)
	}
	if got.Background.Color != mapColorRole(contracts.ColorCanvas) {
		t.Errorf("Background.Color = %v, want Canvas role", got.Background.Color)
	}
}
