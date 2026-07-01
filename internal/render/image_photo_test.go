package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapRadiusRole asserts every wire RadiusRole value maps to its engine
// counterpart, and an unknown/empty value falls back to RadiusNone (R13.11).
func TestMapRadiusRole(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   contracts.RadiusRole
		want pptx.RadiusRole
	}{
		{contracts.RadiusNone, pptx.RadiusNone},
		{"", pptx.RadiusNone},
		{contracts.RadiusSM, pptx.RadiusSM},
		{contracts.RadiusMD, pptx.RadiusMD},
		{contracts.RadiusLG, pptx.RadiusLG},
		{contracts.RadiusFull, pptx.RadiusFull},
		{"bogus", pptx.RadiusNone},
	}
	for _, tc := range cases {
		if got := mapRadiusRole(tc.in); got != tc.want {
			t.Errorf("mapRadiusRole(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestMapNodeImageCornerRadiusElevation asserts a set CornerRadius/Elevation
// maps through to the engine Image (R13.11).
func TestMapNodeImageCornerRadiusElevation(t *testing.T) {
	t.Parallel()

	sn := mapNode(&contracts.Image{
		AssetID:      "a",
		CornerRadius: contracts.RadiusLG,
		Elevation:    contracts.ElevationRaised,
	})
	img, ok := sn.(scene.Image)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Image", sn)
	}
	if img.CornerRadius != pptx.RadiusLG {
		t.Errorf("CornerRadius = %v, want %v", img.CornerRadius, pptx.RadiusLG)
	}
	if img.Elevation != scene.ElevationRaised {
		t.Errorf("Elevation = %v, want %v", img.Elevation, scene.ElevationRaised)
	}
}

// TestMapNodeImageZeroByteIdentical asserts an Image without
// CornerRadius/Elevation set maps identically whether the fields are
// explicitly zeroed or simply omitted — the byte-identical opt-out
// (R13.11).
func TestMapNodeImageZeroByteIdentical(t *testing.T) {
	t.Parallel()

	withoutFields := mapNode(&contracts.Image{AssetID: "a"})
	explicitZero := mapNode(&contracts.Image{AssetID: "a", CornerRadius: contracts.RadiusNone, Elevation: contracts.ElevationFlat})

	imgWithout, ok := withoutFields.(scene.Image)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Image", withoutFields)
	}
	imgExplicit, ok := explicitZero.(scene.Image)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Image", explicitZero)
	}
	if imgWithout.CornerRadius != imgExplicit.CornerRadius {
		t.Errorf("CornerRadius (without) = %v, want == (explicit zero) %v", imgWithout.CornerRadius, imgExplicit.CornerRadius)
	}
	if imgWithout.Elevation != imgExplicit.Elevation {
		t.Errorf("Elevation (without) = %v, want == (explicit zero) %v", imgWithout.Elevation, imgExplicit.Elevation)
	}
	if imgWithout.CornerRadius != pptx.RadiusNone {
		t.Errorf("CornerRadius = %v, want RadiusNone", imgWithout.CornerRadius)
	}
	if imgWithout.Elevation != scene.ElevationFlat {
		t.Errorf("Elevation = %v, want ElevationFlat", imgWithout.Elevation)
	}
}

// TestMapNodeCardImageFill asserts Card.ImageFill maps through to the
// engine's Card.ImageFill (R14.1), and an unset field maps to the empty
// AssetID — byte-identical to a pre-R14.1 card.
func TestMapNodeCardImageFill(t *testing.T) {
	t.Parallel()

	withFill := mapNode(&contracts.Card{Header: "Photo", ImageFill: "asset://hero.jpg"})
	cWithFill, ok := withFill.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", withFill)
	}
	if cWithFill.ImageFill != scene.AssetID("asset://hero.jpg") {
		t.Errorf("ImageFill = %q, want %q", cWithFill.ImageFill, "asset://hero.jpg")
	}

	withoutFill := mapNode(&contracts.Card{Header: "Plain"})
	cWithoutFill, ok := withoutFill.(scene.Card)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Card", withoutFill)
	}
	if cWithoutFill.ImageFill != scene.AssetID("") {
		t.Errorf("ImageFill (unset) = %q, want empty", cWithoutFill.ImageFill)
	}
}

// TestMapBackgroundScrimDuotone asserts a set Scrim/Duotone maps through to
// the engine's scene.Background (R14.1), including the Opacity [0,1] ->
// OOXML [0,100000] scale conversion (mirrors mesh Alpha).
func TestMapBackgroundScrimDuotone(t *testing.T) {
	t.Parallel()

	b := contracts.Background{
		Kind:  contracts.BackgroundColor,
		Color: contracts.ColorAccent,
		Scrim: &contracts.Scrim{
			Color:         contracts.ColorCanvas,
			Opacity:       0.5,
			Gradient:      true,
			GradientAngle: 90,
		},
		Duotone: &contracts.Duotone{
			Shadow:    contracts.ColorAccent,
			Highlight: contracts.ColorSurface,
		},
	}
	sb := mapBackground(b)
	if sb.Scrim == nil {
		t.Fatal("Scrim = nil, want non-nil")
	}
	if sb.Scrim.Color != mapColorRole(contracts.ColorCanvas) {
		t.Errorf("Scrim.Color = %v, want %v", sb.Scrim.Color, mapColorRole(contracts.ColorCanvas))
	}
	if sb.Scrim.Opacity != 50000 {
		t.Errorf("Scrim.Opacity = %d, want 50000", sb.Scrim.Opacity)
	}
	if !sb.Scrim.Gradient {
		t.Error("Scrim.Gradient = false, want true")
	}
	if sb.Scrim.GradientAngle != 90 {
		t.Errorf("Scrim.GradientAngle = %d, want 90", sb.Scrim.GradientAngle)
	}
	if sb.Duotone == nil {
		t.Fatal("Duotone = nil, want non-nil")
	}
	if sb.Duotone.Shadow != mapColorRole(contracts.ColorAccent) {
		t.Errorf("Duotone.Shadow = %v, want %v", sb.Duotone.Shadow, mapColorRole(contracts.ColorAccent))
	}
	if sb.Duotone.Highlight != mapColorRole(contracts.ColorSurface) {
		t.Errorf("Duotone.Highlight = %v, want %v", sb.Duotone.Highlight, mapColorRole(contracts.ColorSurface))
	}
}

// TestMapBackgroundScrimDuotoneNilByteIdentical asserts a Background with
// neither Scrim nor Duotone set maps to a scene.Background with both nil —
// byte-identical to a pre-R14.1 background, whether the fields are left
// unset or explicitly nil.
func TestMapBackgroundScrimDuotoneNilByteIdentical(t *testing.T) {
	t.Parallel()

	withoutFields := mapBackground(contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorAccent})
	explicitNil := mapBackground(contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorAccent, Scrim: nil, Duotone: nil})

	if withoutFields.Scrim != nil {
		t.Errorf("Scrim (unset) = %v, want nil", withoutFields.Scrim)
	}
	if withoutFields.Duotone != nil {
		t.Errorf("Duotone (unset) = %v, want nil", withoutFields.Duotone)
	}
	if explicitNil.Scrim != nil {
		t.Errorf("Scrim (explicit nil) = %v, want nil", explicitNil.Scrim)
	}
	if explicitNil.Duotone != nil {
		t.Errorf("Duotone (explicit nil) = %v, want nil", explicitNil.Duotone)
	}
}
