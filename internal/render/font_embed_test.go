package render

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
)

// readZipPart returns the bytes of a named part inside a .pptx (a zip), or nil.
func readZipPart(t *testing.T, buf []byte, name string) []byte {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %q: %v", name, err)
		}
		defer func() { _ = rc.Close() }()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		return data
	}
	return nil
}

// fontParts returns the names of every embedded font-data part in a .pptx.
func fontParts(t *testing.T, buf []byte) []string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	var out []string
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/fonts/") {
			out = append(out, f.Name)
		}
	}
	return out
}

// TestFontEmbeddingShipsBrandFaces proves the R9.1 acceptance: exporting a deck
// whose soul names bundled families yields /ppt/fonts/*.fntdata parts and a
// presentation.xml embeddedFontLst covering every family used on a slide.
func TestFontEmbeddingShipsBrandFaces(t *testing.T) {
	t.Parallel()

	buf, _, err := Render(testDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	parts := fontParts(t, buf)
	if len(parts) == 0 {
		t.Fatal("no ppt/fonts/*.fntdata parts embedded; the brand faces did not ship")
	}
	for _, p := range parts {
		if !strings.HasSuffix(p, ".fntdata") {
			t.Errorf("font part %q is not a .fntdata part", p)
		}
	}

	pres := string(readZipPart(t, buf, "ppt/presentation.xml"))
	if !strings.Contains(pres, "embeddedFontLst") {
		t.Fatal("presentation.xml has no embeddedFontLst")
	}
	// testDoc uses TypeDisplay (Playfair Display), TypeH2 (Lora) and body (Inter);
	// every one must be listed as an embedded typeface.
	for _, family := range []string{"Playfair Display", "Lora", "Inter"} {
		if !strings.Contains(pres, `typeface="`+family+`"`) {
			t.Errorf("embeddedFontLst missing used family %q", family)
		}
	}
}

// TestFontEmbeddingDeterministic renders the same deck twice and asserts the
// bytes — including the font parts and their relationship ids — are identical.
func TestFontEmbeddingDeterministic(t *testing.T) {
	t.Parallel()

	first, _, err := Render(testDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(testDoc(), soul.DeckardWhite())
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("font-embedding render is not byte-identical across runs")
	}
}

// neverFontSource resolves nothing — it models a soul whose named families are
// all system fonts the provider cannot supply.
type neverFontSource struct{}

func (neverFontSource) Resolve(string, string, int) ([]byte, error) {
	return nil, pptx.ErrFontNotFound
}

// TestFontEmbeddingByteIdenticalWhenNothingResolves proves the byte-identity
// clause: enabling embedding with a provider that resolves no face embeds
// nothing and is byte-identical to the no-provider (pre-embedding) path.
func TestFontEmbeddingByteIdenticalWhenNothingResolves(t *testing.T) {
	t.Parallel()

	noProvider := soul.DeckardWhite()
	noProvider.FontProvider = nil

	unresolvable := soul.DeckardWhite()
	unresolvable.FontProvider = neverFontSource{}

	base, _, err := Render(testDoc(), noProvider)
	if err != nil {
		t.Fatalf("no-provider Render() error = %v", err)
	}
	withEmbedding, _, err := Render(testDoc(), unresolvable)
	if err != nil {
		t.Fatalf("unresolvable-provider Render() error = %v", err)
	}

	if parts := fontParts(t, base); len(parts) != 0 {
		t.Fatalf("nil-provider render embedded fonts: %v", parts)
	}
	if parts := fontParts(t, withEmbedding); len(parts) != 0 {
		t.Fatalf("unresolvable-provider render embedded fonts: %v", parts)
	}
	if !bytes.Equal(base, withEmbedding) {
		t.Fatal("enabling embedding with an all-system-font soul is not byte-identical to the pre-embedding path")
	}
}
