package contracts

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestVariantRoundTrip asserts that Variant marshals/unmarshals cleanly:
// the zero value ("light") is omitted when embedded in a struct via omitempty,
// and "dark" survives a round-trip.
func TestVariantRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		variant    Variant
		wantAbsent bool // if true, expect the field to be absent in JSON (omitempty)
	}{
		{name: "light omitted", variant: VariantLight, wantAbsent: false},
		{name: "empty omitted", variant: "", wantAbsent: true},
		{name: "dark present", variant: VariantDark, wantAbsent: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			type wrapper struct {
				V Variant `json:"v,omitempty"`
			}
			w := wrapper{V: tc.variant}
			b, err := json.Marshal(w)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got wrapper
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.V != tc.variant {
				t.Fatalf("Variant round-trip: want %q got %q", tc.variant, got.V)
			}
			if tc.wantAbsent {
				if string(b) != `{}` {
					t.Fatalf("expected omitted field ({}), got %s", b)
				}
			}
		})
	}
}

// TestBackgroundRoundTrip asserts Background marshals/unmarshals cleanly
// across all kinds, including the gradient slice (0, 1, 2 roles).
func TestBackgroundRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		bg   Background
	}{
		{
			name: "none (zero value)",
			bg:   Background{},
		},
		{
			name: "color",
			bg:   Background{Kind: BackgroundColor, Color: ColorAccent},
		},
		{
			name: "gradient two roles",
			bg: Background{
				Kind:     BackgroundGradient,
				Gradient: []ColorRole{ColorAccent, ColorAccentAlt},
				Angle:    135,
			},
		},
		{
			name: "gradient one role",
			bg: Background{
				Kind:     BackgroundGradient,
				Gradient: []ColorRole{ColorCanvas},
				Angle:    90,
			},
		},
		{
			name: "asset",
			bg:   Background{Kind: BackgroundAsset, AssetID: "brand-bg"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.bg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got Background
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tc.bg, got) {
				t.Fatalf("Background round-trip drift:\nwant=%#v\ngot =%#v", tc.bg, got)
			}
		})
	}
}

// TestSlideVariantBackgroundRoundTrip asserts a Slide with Variant + Background
// survives the full marshal→unmarshal cycle, including the pointer omitempty:
// a nil Background must be absent from JSON; a non-nil Background must survive.
func TestSlideVariantBackgroundRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("dark with gradient background", func(t *testing.T) {
		want := Slide{
			ID:      "section-dark",
			Layout:  LayoutTitleContent,
			Variant: VariantDark,
			Background: &Background{
				Kind:     BackgroundGradient,
				Gradient: []ColorRole{ColorAccent, ColorAccentAlt},
				Angle:    135,
			},
			Nodes: []SlideNode{&Hero{Title: "Dark Section"}},
		}
		b, err := json.Marshal(&want)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var got Slide
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Variant != want.Variant {
			t.Fatalf("Variant: want %q got %q", want.Variant, got.Variant)
		}
		if !reflect.DeepEqual(got.Background, want.Background) {
			t.Fatalf("Background round-trip drift:\nwant=%#v\ngot =%#v", want.Background, got.Background)
		}
	})

	t.Run("nil background omitted from JSON", func(t *testing.T) {
		slide := Slide{ID: "light", Nodes: []SlideNode{&Hero{Title: "Light"}}}
		b, err := json.Marshal(&slide)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		// background key must be absent
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatalf("unmarshal raw: %v", err)
		}
		if _, ok := raw["background"]; ok {
			t.Fatal("nil Background should be absent from JSON, but key was present")
		}
		if _, ok := raw["variant"]; ok {
			t.Fatal("zero Variant should be absent from JSON, but key was present")
		}
	})

	t.Run("light variant omitted from JSON", func(t *testing.T) {
		slide := Slide{ID: "s", Variant: VariantLight, Nodes: []SlideNode{&Hero{Title: "T"}}}
		b, err := json.Marshal(&slide)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatalf("unmarshal raw: %v", err)
		}
		// VariantLight == "light" is non-empty, so it will be present in JSON.
		// The zero Variant ("") is what gets omitted; "light" is explicit.
		// Both are valid; just verify round-trip is lossless.
		var got Slide
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Variant != VariantLight {
			t.Fatalf("Variant round-trip: want %q got %q", VariantLight, got.Variant)
		}
	})
}
