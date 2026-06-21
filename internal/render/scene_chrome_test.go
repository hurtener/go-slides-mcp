package render

import (
	"bytes"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapDocChromeEnabled verifies that a DeckChrome with Enabled==true maps
// to a scene.Chrome with Enabled==true, the correct Brand/BrandAsset fields,
// and Total==0 (so the engine resolves it to len(Scene.Slides) as required).
func TestMapDocChromeEnabled(t *testing.T) {
	t.Parallel()

	c := contracts.DeckChrome{Enabled: true, BrandText: "Acme Corp", BrandAssetID: "logo-01"}
	got := mapDocChrome(c)

	if !got.Enabled {
		t.Fatal("mapDocChrome: Enabled = false, want true")
	}
	if got.Brand != "Acme Corp" {
		t.Fatalf("mapDocChrome: Brand = %q, want %q", got.Brand, "Acme Corp")
	}
	if string(got.BrandAsset) != "logo-01" {
		t.Fatalf("mapDocChrome: BrandAsset = %q, want %q", got.BrandAsset, "logo-01")
	}
	if got.Total != 0 {
		t.Fatalf("mapDocChrome: Total = %d, want 0 (let engine auto-fill from slide count)", got.Total)
	}
}

// TestMapDocChromeDisabled verifies that a zero-value DeckChrome (Enabled==false)
// maps to an empty scene.Chrome — no fields set — so the engine draws nothing.
func TestMapDocChromeDisabled(t *testing.T) {
	t.Parallel()

	got := mapDocChrome(contracts.DeckChrome{})
	if got != (scene.Chrome{}) {
		t.Fatalf("mapDocChrome disabled: got %+v, want zero value", got)
	}
}

// TestMapSlidesSetsPageNumberAndSection verifies that mapSlides sets the
// 1-based PageNumber and the Section string for each slide.
func TestMapSlidesSetsPageNumberAndSection(t *testing.T) {
	t.Parallel()

	slides := []contracts.Slide{
		{ID: "s1", Section: "01 — Direction"},
		{ID: "s2", Section: ""},
		{ID: "s3", Section: "02 — Execution"},
	}
	got := mapSlides(slides)

	for i, ss := range got {
		wantPage := i + 1
		if ss.PageNumber != wantPage {
			t.Errorf("slide[%d] PageNumber = %d, want %d", i, ss.PageNumber, wantPage)
		}
	}
	if got[0].Section != "01 — Direction" {
		t.Errorf("slide[0] Section = %q, want %q", got[0].Section, "01 — Direction")
	}
	if got[1].Section != "" {
		t.Errorf("slide[1] Section = %q, want empty", got[1].Section)
	}
	if got[2].Section != "02 — Execution" {
		t.Errorf("slide[2] Section = %q, want %q", got[2].Section, "02 — Execution")
	}
}

// TestRenderChromeEnabledProducesValidPPTX renders a chrome-enabled deck and
// confirms a valid PPTX is produced without errors.
func TestRenderChromeEnabledProducesValidPPTX(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{
		Title:  "Chrome Test",
		Chrome: contracts.DeckChrome{Enabled: true, BrandText: "Acme Corp"},
		Slides: []contracts.Slide{
			{
				ID:      "cover",
				Layout:  contracts.LayoutCover,
				Section: "01 — Intro",
				Nodes:   []contracts.SlideNode{&contracts.Hero{Title: "Chrome Deck"}},
			},
			{
				ID:      "content",
				Layout:  contracts.LayoutTitleContent,
				Section: "02 — Detail",
				Nodes:   []contracts.SlideNode{&contracts.Heading{Level: 2, Text: rt("Body")}},
			},
		},
	}

	buf, stats, err := Render(doc, soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render chrome deck: %v", err)
	}
	if len(buf) == 0 {
		t.Fatal("Render returned empty bytes")
	}
	if stats.Slides != 2 {
		t.Fatalf("Render slides = %d, want 2", stats.Slides)
	}
	assertValidPPTX(t, buf)
}

// TestRenderChromeDisabledByteIdentical verifies that a deck with no chrome
// config renders byte-identical to the same deck explicitly setting
// Chrome.Enabled = false — the zero value and explicit false are equivalent.
func TestRenderChromeDisabledByteIdentical(t *testing.T) {
	t.Parallel()

	base := contracts.SlideDoc{
		Title: "No Chrome",
		Slides: []contracts.Slide{
			{ID: "s1", Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Cover"}}},
		},
	}
	explicitOff := contracts.SlideDoc{
		Title:  base.Title,
		Chrome: contracts.DeckChrome{Enabled: false},
		Slides: append([]contracts.Slide(nil), base.Slides...),
	}

	s := soul.DeckardWhite()
	a, _, err := Render(base, s)
	if err != nil {
		t.Fatalf("Render base: %v", err)
	}
	b, _, err := Render(explicitOff, s)
	if err != nil {
		t.Fatalf("Render explicitOff: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("chrome disabled vs zero-chrome bytes differ — regression")
	}
}
