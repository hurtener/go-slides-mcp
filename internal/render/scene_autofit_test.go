package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeAutoFit asserts that AutoFit maps through unchanged (true/false)
// for Hero, Heading, and Stat — the three display-text nodes that expose
// engine shrink-to-fit (R10.5, D-074).
func TestMapNodeAutoFit(t *testing.T) {
	t.Parallel()

	t.Run("Hero true", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Hero{Title: "Long Title", AutoFit: true})
		h, ok := sn.(scene.Hero)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Hero", sn)
		}
		if !h.AutoFit {
			t.Errorf("AutoFit = false, want true")
		}
	})

	t.Run("Hero unset", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Hero{Title: "Title"})
		h, ok := sn.(scene.Hero)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Hero", sn)
		}
		if h.AutoFit {
			t.Errorf("AutoFit = true, want false (unset = byte-identical default)")
		}
	})

	t.Run("Heading true", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Heading{Text: contracts.RichText{{Text: "Heading"}}, Level: 1, AutoFit: true})
		h, ok := sn.(scene.Heading)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Heading", sn)
		}
		if !h.AutoFit {
			t.Errorf("AutoFit = false, want true")
		}
	})

	t.Run("Heading unset", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Heading{Level: 1})
		h, ok := sn.(scene.Heading)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Heading", sn)
		}
		if h.AutoFit {
			t.Errorf("AutoFit = true, want false (unset = byte-identical default)")
		}
	})

	t.Run("Stat true", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Stat{Value: "$2,200", Label: "per month", AutoFit: true})
		s, ok := sn.(scene.Stat)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Stat", sn)
		}
		if !s.AutoFit {
			t.Errorf("AutoFit = false, want true")
		}
	})

	t.Run("Stat unset", func(t *testing.T) {
		t.Parallel()
		sn := mapNode(&contracts.Stat{Value: "98%", Label: "NPS score"})
		s, ok := sn.(scene.Stat)
		if !ok {
			t.Fatalf("mapNode returned %T, want scene.Stat", sn)
		}
		if s.AutoFit {
			t.Errorf("AutoFit = true, want false (unset = byte-identical default)")
		}
	})
}
