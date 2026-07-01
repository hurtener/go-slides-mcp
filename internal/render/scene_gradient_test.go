package render

import (
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// gradientDoc builds a one-slide deck whose Background requests a named soul
// gradient.
func gradientDoc(name string) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Gradient render",
		Slides: []contracts.Slide{
			{
				ID:     "hero",
				Layout: contracts.LayoutTitleContent,
				Background: &contracts.Background{
					Kind:         contracts.BackgroundGradient,
					GradientName: name,
				},
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Title: "Named gradient"},
				},
			},
		},
	}
}

// TestRenderNamedGradientResolvesThroughSoulTheme is the R8.5 observability
// test: a soul bootstrapped with a "heroDark" gradient renders a slide that
// requests it by name with NO gradient-not-found warning, proving
// Background.GradientName reaches scene.Background and resolves through the
// soul's Theme.Gradients. A slide naming an undefined gradient DOES record a
// warning, proving the absence is also observable.
func TestRenderNamedGradientResolvesThroughSoulTheme(t *testing.T) {
	t.Parallel()

	s, err := soul.Bootstrap(soul.BootstrapParams{
		Name: "Acme",
		Gradients: []soul.GradientSpec{
			{
				Name: "heroDark",
				Stops: []soul.GradientStop{
					{Pos: 0, ColorHex: "0A1622"},
					{Pos: 1, ColorHex: "1B3350"},
				},
				Radial: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	t.Run("defined gradient name: no warning", func(t *testing.T) {
		t.Parallel()
		buf, stats, err := Render(gradientDoc("heroDark"), s)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}
		if len(buf) == 0 {
			t.Fatal("Render() returned empty bytes")
		}
		for _, w := range stats.Warnings {
			if strings.Contains(strings.ToLower(w), "gradient") {
				t.Errorf("unexpected gradient warning for a defined gradient name: %q", w)
			}
		}
	})

	t.Run("undefined gradient name: warning recorded", func(t *testing.T) {
		t.Parallel()
		buf, stats, err := Render(gradientDoc("nope"), s)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}
		if len(buf) == 0 {
			t.Fatal("Render() returned empty bytes")
		}
		found := false
		for _, w := range stats.Warnings {
			if strings.Contains(strings.ToLower(w), "gradient") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected a gradient-not-found warning for undefined name %q; got %v", "nope", stats.Warnings)
		}
	})
}

// TestMapBackgroundGradientName asserts GradientName passes through
// mapBackground unchanged.
func TestMapBackgroundGradientName(t *testing.T) {
	t.Parallel()

	bg := contracts.Background{
		Kind:         contracts.BackgroundGradient,
		GradientName: "heroDark",
	}
	got := mapBackground(bg)
	if got.GradientName != "heroDark" {
		t.Errorf("GradientName = %q, want %q", got.GradientName, "heroDark")
	}
}
