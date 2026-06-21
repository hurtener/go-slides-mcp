package render

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapNodeStatAllFields asserts that a fully-populated Stat maps to a
// scene.Stat with all fields set (R6, D-057).
func TestMapNodeStatAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Stat{
		Value:     "$2,200",
		Label:     "per month",
		Delta:     "+18%",
		DeltaTone: contracts.DeltaUp,
	}
	sn := mapNode(node)
	s, ok := sn.(scene.Stat)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Stat", sn)
	}

	if s.Value != "$2,200" {
		t.Errorf("Value: got %q, want %q", s.Value, "$2,200")
	}
	if s.Label != "per month" {
		t.Errorf("Label: got %q, want %q", s.Label, "per month")
	}
	if s.Delta != "+18%" {
		t.Errorf("Delta: got %q, want %q", s.Delta, "+18%")
	}
	if s.DeltaTone != scene.DeltaUp {
		t.Errorf("DeltaTone: got %v, want scene.DeltaUp", s.DeltaTone)
	}
}

// TestMapNodeStatDeltaDown asserts that DeltaDown maps correctly (error color).
func TestMapNodeStatDeltaDown(t *testing.T) {
	t.Parallel()

	node := &contracts.Stat{
		Value:     "14 days",
		Label:     "avg cycle",
		Delta:     "-2 days",
		DeltaTone: contracts.DeltaDown,
	}
	sn := mapNode(node)
	s, ok := sn.(scene.Stat)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Stat", sn)
	}
	if s.DeltaTone != scene.DeltaDown {
		t.Errorf("DeltaTone: got %v, want scene.DeltaDown", s.DeltaTone)
	}
}

// TestMapNodeStatNoDelta asserts that an omitted Delta and DeltaTone map to
// empty string and DeltaNeutral respectively.
func TestMapNodeStatNoDelta(t *testing.T) {
	t.Parallel()

	node := &contracts.Stat{Value: "98%", Label: "NPS score"}
	sn := mapNode(node)
	s, ok := sn.(scene.Stat)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Stat", sn)
	}
	if s.Delta != "" {
		t.Errorf("Delta: got %q, want empty", s.Delta)
	}
	if s.DeltaTone != scene.DeltaNeutral {
		t.Errorf("DeltaTone: got %v, want scene.DeltaNeutral (zero)", s.DeltaTone)
	}
}

// TestMapNodeStatNeutralToneEmptyString asserts that an empty DeltaTone string
// maps to DeltaNeutral (the zero value / default).
func TestMapNodeStatNeutralToneEmptyString(t *testing.T) {
	t.Parallel()

	node := &contracts.Stat{Value: "42", Label: "units", Delta: "±0"}
	// DeltaTone is "" (zero value — not set)
	sn := mapNode(node)
	s, ok := sn.(scene.Stat)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Stat", sn)
	}
	if s.DeltaTone != scene.DeltaNeutral {
		t.Errorf("DeltaTone: got %v, want scene.DeltaNeutral for empty wire value", s.DeltaTone)
	}
}
