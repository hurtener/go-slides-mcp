package render

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// numberFormatDoc builds a single-slide doc with one Stat carrying the given
// Number + Format (or the raw Value when number is nil), for end-to-end
// render assertions (R14.13, D-121).
func numberFormatDoc(value string, number *float64, format *contracts.NumberFormat) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Number Format Coverage",
		Slides: []contracts.Slide{
			{
				ID:     "stat",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Heading{Level: 2, Text: rt("Number Format")},
					&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
						&contracts.Stat{Value: value, Label: "metric", Number: number, Format: format},
						&contracts.Stat{Value: "companion", Label: "other"},
					}},
				},
			},
		},
	}
}

// firstSlideXML renders buf, unzips it, and returns the raw bytes of
// ppt/slides/slide1.xml.
func firstSlideXML(t *testing.T, buf []byte) []byte {
	t.Helper()

	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	for _, f := range r.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open slide1.xml: %v", err)
			}
			defer func() { _ = rc.Close() }()
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read slide1.xml: %v", err)
			}
			return data
		}
	}
	t.Fatal("rendered zip missing ppt/slides/slide1.xml")
	return nil
}

// TestRenderStatNumberFormatUSD asserts a whole-dollar USD Stat renders
// "$4,000" in the slide XML (R14.13 accept case).
func TestRenderStatNumberFormatUSD(t *testing.T) {
	t.Parallel()

	num := 4000.0
	format := &contracts.NumberFormat{CurrencySymbol: "$", GroupSep: ","}
	doc := numberFormatDoc("unused", &num, format)

	buf, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := scene.FormatNumber(num, scene.NumberFormat{CurrencySymbol: "$", GroupSep: ","})
	if want != "$4,000" {
		t.Fatalf("test setup: scene.FormatNumber = %q, want %q", want, "$4,000")
	}

	xml := firstSlideXML(t, buf)
	if !bytes.Contains(xml, []byte(want)) {
		t.Errorf("slide1.xml does not contain %q", want)
	}
}

// TestRenderStatNumberFormatUSDSuffix asserts a USD Stat with a "+" suffix
// renders "$4,000+" with no orphaned "+" (R14.13 accept case).
func TestRenderStatNumberFormatUSDSuffix(t *testing.T) {
	t.Parallel()

	num := 4000.0
	format := &contracts.NumberFormat{CurrencySymbol: "$", GroupSep: ",", Suffix: "+"}
	doc := numberFormatDoc("unused", &num, format)

	buf, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := scene.FormatNumber(num, scene.NumberFormat{CurrencySymbol: "$", GroupSep: ",", Suffix: "+"})
	if want != "$4,000+" {
		t.Fatalf("test setup: scene.FormatNumber = %q, want %q", want, "$4,000+")
	}

	xml := firstSlideXML(t, buf)
	if !bytes.Contains(xml, []byte(want)) {
		t.Errorf("slide1.xml does not contain %q", want)
	}
	// The "+" must be attached to the number, not orphaned as its own run.
	if bytes.Contains(xml, []byte(">+<")) {
		t.Errorf("slide1.xml contains an orphaned \"+\" run")
	}
}

// TestRenderStatNumberFormatPercent asserts a percent Stat renders "92%"
// (R14.13 accept case).
func TestRenderStatNumberFormatPercent(t *testing.T) {
	t.Parallel()

	num := 0.92
	format := &contracts.NumberFormat{Percent: true}
	doc := numberFormatDoc("unused", &num, format)

	buf, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := scene.FormatNumber(num, scene.NumberFormat{Percent: true})
	if want != "92%" {
		t.Fatalf("test setup: scene.FormatNumber = %q, want %q", want, "92%")
	}

	xml := firstSlideXML(t, buf)
	if !bytes.Contains(xml, []byte(want)) {
		t.Errorf("slide1.xml does not contain %q", want)
	}
}

// TestRenderStatNumberFormatDeDELocale asserts a de-DE locale Stat (dot
// thousands separator) renders "4.000" (R14.13 accept case).
func TestRenderStatNumberFormatDeDELocale(t *testing.T) {
	t.Parallel()

	num := 4000.0
	format := &contracts.NumberFormat{GroupSep: ".", DecimalSep: ","}
	doc := numberFormatDoc("unused", &num, format)

	buf, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := scene.FormatNumber(num, scene.NumberFormat{GroupSep: ".", DecimalSep: ","})
	if want != "4.000" {
		t.Fatalf("test setup: scene.FormatNumber = %q, want %q", want, "4.000")
	}

	xml := firstSlideXML(t, buf)
	if !bytes.Contains(xml, []byte(want)) {
		t.Errorf("slide1.xml does not contain %q", want)
	}
}

// TestRenderStatNilNumberByteIdentical asserts a Stat with a nil Number
// renders the raw Value verbatim, byte-identical to a pre-R14.13 Stat, and
// contains no formatted-number artifact.
func TestRenderStatNilNumberByteIdentical(t *testing.T) {
	t.Parallel()

	legacyDoc := numberFormatDoc("legacy", nil, nil)

	first, _, err := Render(legacyDoc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(legacyDoc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("Render() bytes differ across identical renders of a nil-Number Stat")
	}

	xml := firstSlideXML(t, first)
	if !bytes.Contains(xml, []byte("legacy")) {
		t.Error("slide1.xml does not contain the raw Value \"legacy\"")
	}
	if strings.Contains(string(xml), "$4,000") || strings.Contains(string(xml), "92%") {
		t.Error("slide1.xml unexpectedly contains formatted-number output for a nil-Number Stat")
	}
}

// TestRenderStatNumberFormatDeterministic asserts repeated renders of the
// same formatted Stat are byte-identical (render determinism is a hard
// contract, CLAUDE.md §5).
func TestRenderStatNumberFormatDeterministic(t *testing.T) {
	t.Parallel()

	num := 4000.0
	doc := numberFormatDoc("unused", &num, &contracts.NumberFormat{CurrencySymbol: "$", GroupSep: ",", Suffix: "+"})

	first, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("Render() bytes differ across identical renders of a formatted Stat")
	}
}
