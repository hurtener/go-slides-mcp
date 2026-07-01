package soul

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestBootstrapGradientsPopulateThemeGradients(t *testing.T) {
	s, err := Bootstrap(BootstrapParams{
		Name: "Acme",
		Gradients: []GradientSpec{
			{
				Name: "heroDark",
				Stops: []GradientStop{
					{Pos: 0, ColorHex: "0A1622"},
					{Pos: 0.6, ColorHex: "12233A"},
					{Pos: 1, ColorHex: "1B3350"},
				},
				Radial: true,
			},
			{
				Name: "accentSweep",
				Stops: []GradientStop{
					{Pos: 0, ColorRole: "accent"},
					{Pos: 1, ColorRole: "accentAlt"},
				},
				Angle: 135,
			},
		},
	})
	if err != nil {
		fatalBootstrap(t, err)
	}
	if s.Theme.Gradients == nil {
		t.Fatal("Theme.Gradients = nil, want populated")
	}

	hero, ok := s.Theme.Gradients["heroDark"]
	if !ok {
		t.Fatal(`Theme.Gradients["heroDark"] missing`)
	}
	if !hero.Radial {
		t.Error("heroDark.Radial = false, want true")
	}
	if len(hero.Stops) != 3 {
		t.Fatalf("heroDark stops = %d, want 3", len(hero.Stops))
	}
	wantHex := []pptx.RGB{"0A1622", "12233A", "1B3350"}
	for i, want := range wantHex {
		var c pptx.Color = want
		if hero.Stops[i].Color != c {
			t.Errorf("heroDark stop %d Color = %v, want literal %v", i, hero.Stops[i].Color, want)
		}
		if hero.Stops[i].Pos != []float64{0, 0.6, 1}[i] {
			t.Errorf("heroDark stop %d Pos = %v", i, hero.Stops[i].Pos)
		}
	}

	sweep, ok := s.Theme.Gradients["accentSweep"]
	if !ok {
		t.Fatal(`Theme.Gradients["accentSweep"] missing`)
	}
	if sweep.Radial {
		t.Error("accentSweep.Radial = true, want false")
	}
	if sweep.Angle != 135 {
		t.Errorf("accentSweep.Angle = %d, want 135", sweep.Angle)
	}
	if len(sweep.Stops) != 2 {
		t.Fatalf("accentSweep stops = %d, want 2", len(sweep.Stops))
	}
	if sweep.Stops[0].Color != pptx.TokenColor(pptx.ColorAccent) {
		t.Errorf("accentSweep stop 0 Color = %v, want TokenColor(ColorAccent)", sweep.Stops[0].Color)
	}
	if sweep.Stops[1].Color != pptx.TokenColor(pptx.ColorAccentAlt) {
		t.Errorf("accentSweep stop 1 Color = %v, want TokenColor(ColorAccentAlt)", sweep.Stops[1].Color)
	}
}

func TestBootstrapNilOrEmptyGradientsLeavesThemeGradientsNil(t *testing.T) {
	for _, g := range [][]GradientSpec{nil, {}} {
		s, err := Bootstrap(BootstrapParams{Name: "x", Gradients: g})
		if err != nil {
			fatalBootstrap(t, err)
		}
		if s.Theme.Gradients != nil {
			t.Fatalf("Theme.Gradients = %+v, want nil for %+v", s.Theme.Gradients, g)
		}
	}
}

func TestBootstrapGradientValidationErrors(t *testing.T) {
	base := func() GradientSpec {
		return GradientSpec{
			Name: "g1",
			Stops: []GradientStop{
				{Pos: 0, ColorHex: "112233"},
				{Pos: 1, ColorHex: "445566"},
			},
		}
	}

	cases := []struct {
		name  string
		specs []GradientSpec
	}{
		{
			name: "duplicate name",
			specs: []GradientSpec{
				base(),
				base(),
			},
		},
		{
			name: "1 stop",
			specs: []GradientSpec{{
				Name:  "g1",
				Stops: []GradientStop{{Pos: 0, ColorHex: "112233"}},
			}},
		},
		{
			name: "9 stops",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0.0, ColorHex: "112233"},
					{Pos: 0.1, ColorHex: "112233"},
					{Pos: 0.2, ColorHex: "112233"},
					{Pos: 0.3, ColorHex: "112233"},
					{Pos: 0.4, ColorHex: "112233"},
					{Pos: 0.5, ColorHex: "112233"},
					{Pos: 0.6, ColorHex: "112233"},
					{Pos: 0.7, ColorHex: "112233"},
					{Pos: 0.8, ColorHex: "112233"},
				},
			}},
		},
		{
			name: "non-ascending pos",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0.5, ColorHex: "112233"},
					{Pos: 0.5, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "pos out of [0,1]",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: -0.1, ColorHex: "112233"},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "pos above 1",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0, ColorHex: "112233"},
					{Pos: 1.1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "both ColorHex and ColorRole set",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0, ColorHex: "112233", ColorRole: "accent"},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "neither ColorHex nor ColorRole set",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "bad hex",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0, ColorHex: "ZZZZZZ"},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "unknown role",
			specs: []GradientSpec{{
				Name: "g1",
				Stops: []GradientStop{
					{Pos: 0, ColorRole: "notARole"},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
		{
			name: "empty name",
			specs: []GradientSpec{{
				Name: "",
				Stops: []GradientStop{
					{Pos: 0, ColorHex: "112233"},
					{Pos: 1, ColorHex: "445566"},
				},
			}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Bootstrap(BootstrapParams{Name: "x", Gradients: tc.specs}); err == nil {
				t.Fatal("expected a typed error, got nil")
			}
		})
	}
}
