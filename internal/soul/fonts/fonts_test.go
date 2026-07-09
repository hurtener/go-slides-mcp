package fonts

import (
	"bytes"
	"errors"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// sfntMagic is the TrueType outline signature (0x00010000) every bundled face
// starts with; a resolved face must be real font bytes, not an empty slice.
var sfntMagic = []byte{0x00, 0x01, 0x00, 0x00}

func TestProviderSingleton(t *testing.T) {
	if Provider() == nil {
		t.Fatal("Provider() = nil")
	}
	if Provider() != Provider() {
		t.Fatal("Provider() is not a stable singleton")
	}
}

func TestResolveBundledFamilies(t *testing.T) {
	p := Provider()
	cases := []struct {
		name, style string
		weight      int
	}{
		{"Playfair Display", "", 400},
		{"playfair display", "", 400}, // case-insensitive
		{"Playfair Display", "italic", 400},
		{"Lora", "", 400},
		{"Lora", "italic", 400},
		{"Inter", "", 400},
		{"Inter", "", 500},
		{"Inter", "", 700},
		{"Inter", "italic", 400},
	}
	for _, c := range cases {
		data, err := p.Resolve(c.name, c.style, c.weight)
		if err != nil {
			t.Errorf("Resolve(%q,%q,%d) err = %v", c.name, c.style, c.weight, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Resolve(%q,%q,%d) returned empty bytes", c.name, c.style, c.weight)
			continue
		}
		if !bytes.HasPrefix(data, sfntMagic) {
			t.Errorf("Resolve(%q,%q,%d) not a TrueType font (bad magic %x)", c.name, c.style, c.weight, data[:4])
		}
	}
}

func TestResolveItalicDiffersFromUpright(t *testing.T) {
	p := Provider()
	up, err := p.Resolve("Playfair Display", "", 400)
	if err != nil {
		t.Fatalf("upright err = %v", err)
	}
	it, err := p.Resolve("Playfair Display", "italic", 400)
	if err != nil {
		t.Fatalf("italic err = %v", err)
	}
	if bytes.Equal(up, it) {
		t.Fatal("italic cut resolved to the same bytes as the upright cut")
	}
}

func TestResolveNearestWeightDeterministic(t *testing.T) {
	p := Provider()
	w400, _ := p.Resolve("Inter", "", 400)
	w500, _ := p.Resolve("Inter", "", 500)
	w700, _ := p.Resolve("Inter", "", 700)

	// 600 is equidistant to 500 and 700; the tie resolves to the lower weight.
	w600, err := p.Resolve("Inter", "", 600)
	if err != nil {
		t.Fatalf("Resolve Inter 600 err = %v", err)
	}
	if !bytes.Equal(w600, w500) {
		t.Error("Inter@600 should tie-break to the 500 (lower) weight")
	}
	// 450 is equidistant to 400 and 500; tie resolves to 400.
	w450, _ := p.Resolve("Inter", "", 450)
	if !bytes.Equal(w450, w400) {
		t.Error("Inter@450 should tie-break to the 400 (lower) weight")
	}
	// A far-out weight snaps to the nearest available (700).
	w900, _ := p.Resolve("Inter", "", 900)
	if !bytes.Equal(w900, w700) {
		t.Error("Inter@900 should snap to the nearest available weight (700)")
	}
	// Deterministic across calls.
	again, _ := p.Resolve("Inter", "", 600)
	if !bytes.Equal(again, w600) {
		t.Error("Resolve is not deterministic across calls")
	}
}

func TestResolveItalicFallsBackToUpright(t *testing.T) {
	// Inter bundles only a 400 italic; an italic request at a heavier weight
	// snaps to the nearest italic cut (400), never borrows the upright file.
	p := Provider()
	it700, err := p.Resolve("Inter", "italic", 700)
	if err != nil {
		t.Fatalf("Resolve Inter italic 700 err = %v", err)
	}
	it400, _ := p.Resolve("Inter", "italic", 400)
	if !bytes.Equal(it700, it400) {
		t.Error("Inter italic@700 should snap to the only italic cut (400)")
	}
	upright400, _ := p.Resolve("Inter", "", 400)
	if bytes.Equal(it700, upright400) {
		t.Error("italic request must not borrow the upright cut when an italic cut exists")
	}
}

func TestResolveUnknownFamilyNotFound(t *testing.T) {
	p := Provider()
	for _, name := range []string{"Consolas", "Arial", "Helvetica", ""} {
		if _, err := p.Resolve(name, "", 400); !errors.Is(err, pptx.ErrFontNotFound) {
			t.Errorf("Resolve(%q) err = %v, want ErrFontNotFound", name, err)
		}
	}
}

func TestEveryBundledFileValid(t *testing.T) {
	// Guards manifest/embed drift: every declared face must embed and be a real
	// TrueType font.
	for _, f := range bundled {
		data, err := ttf.ReadFile(f.path)
		if err != nil {
			t.Errorf("bundled face %q not embedded: %v", f.path, err)
			continue
		}
		if !bytes.HasPrefix(data, sfntMagic) {
			t.Errorf("bundled face %q is not TrueType (magic %x)", f.path, data[:4])
		}
	}
}

func TestAvgCharWidthMeasuredPerBundledFamily(t *testing.T) {
	tests := []struct {
		family string
		min    float64
		max    float64
	}{
		{family: "Inter", min: 0.45, max: 0.65},
		{family: "Lora", min: 0.45, max: 0.65},
		{family: "Playfair Display", min: 0.45, max: 0.65},
	}
	for _, tc := range tests {
		got, ok := AvgCharWidth(tc.family)
		if !ok {
			t.Fatalf("AvgCharWidth(%q) missing", tc.family)
		}
		if got < tc.min || got > tc.max {
			t.Fatalf("AvgCharWidth(%q) = %.4f, want in [%.2f, %.2f]", tc.family, got, tc.min, tc.max)
		}
	}
	inter, _ := AvgCharWidth("Inter")
	lora, _ := AvgCharWidth("Lora")
	playfair, _ := AvgCharWidth("Playfair Display")
	if inter == lora || lora == playfair || inter == playfair {
		t.Fatalf("expected family-specific measurements, got Inter=%.4f Lora=%.4f Playfair=%.4f", inter, lora, playfair)
	}
	if playfair == 0.5 && lora == 0.5 && inter == 0.5 {
		t.Fatal("expected measured values, got the default factor 0.5 for every bundled family")
	}
}
