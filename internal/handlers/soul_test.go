package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/pptx"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

func TestListSoulsIncludesDeckardWhite(t *testing.T) {
	h := testHandlers()

	got, err := h.listSouls(context.Background(), contracts.ListSoulsInput{})
	if err != nil {
		t.Fatalf("listSouls: %v", err)
	}
	if len(got.Structured.Souls) == 0 {
		t.Fatal("listSouls returned no souls")
	}
	if got.Structured.Souls[0].SoulID != soul.DeckardWhiteID {
		t.Fatalf("first soul id = %q, want %q", got.Structured.Souls[0].SoulID, soul.DeckardWhiteID)
	}
}

func TestBootstrapSoulAndGetDesignTokensRoundTrip(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Teal Variant", Accent: "112233"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}
	if bootstrapped.Structured.TokenCount == 0 {
		t.Fatal("bootstrapSoul token count = 0, want > 0")
	}

	tokens, err := h.getDesignTokens(ctx, contracts.GetDesignTokensInput{SoulID: bootstrapped.Structured.SoulID})
	if err != nil {
		t.Fatalf("getDesignTokens: %v", err)
	}
	if got := tokenValue(tokens.Structured.Tokens, contracts.TokenLayerSurface, "accent"); got != "112233" {
		t.Fatalf("accent token = %q, want 112233", got)
	}
	if len(tokens.Structured.Tokens) != bootstrapped.Structured.TokenCount {
		t.Fatalf("getDesignTokens len = %d, want %d", len(tokens.Structured.Tokens), bootstrapped.Structured.TokenCount)
	}
}

func TestRefineSoulChangesToken(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Refine Me"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}

	refined, err := h.refineSoul(ctx, contracts.RefineSoulInput{SoulID: bootstrapped.Structured.SoulID, Overrides: []contracts.SoulOverride{{Category: "surface", Token: "accent", Value: "ABCDEF"}}})
	if err != nil {
		t.Fatalf("refineSoul: %v", err)
	}
	if len(refined.Structured.Changed) != 1 || refined.Structured.Changed[0] != "surface.accent" {
		t.Fatalf("refineSoul changed = %+v, want [surface.accent]", refined.Structured.Changed)
	}

	tokens, err := h.getDesignTokens(ctx, contracts.GetDesignTokensInput{SoulID: bootstrapped.Structured.SoulID})
	if err != nil {
		t.Fatalf("getDesignTokens: %v", err)
	}
	if got := tokenValue(tokens.Structured.Tokens, contracts.TokenLayerSurface, "accent"); got != "ABCDEF" {
		t.Fatalf("accent token = %q, want ABCDEF", got)
	}
}

// validIconSVG mirrors pptx-go's known-good icon fixture
// (scene/icons_validate_test.go): a single-path triangle that satisfies the
// icon translator subset (D-040).
const validIconSVG = `<svg viewBox="0 0 24 24"><path d="M12 2 L22 22 L2 22 Z"/></svg>`

// arcIconSVG uses an elliptical arc path command, which the translator
// rejects — the known-bad fixture from the same pptx-go test file.
const arcIconSVG = `<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`

// TestRefineSoulBindsIcon asserts that refine_soul's Icons field binds a
// valid brand glyph to the stored soul (R14.16) and surfaces a note in Text.
func TestRefineSoulBindsIcon(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Icon Soul"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}

	refined, err := h.refineSoul(ctx, contracts.RefineSoulInput{
		SoulID: bootstrapped.Structured.SoulID,
		Icons:  map[string]string{"brandmark": validIconSVG},
	})
	if err != nil {
		t.Fatalf("refineSoul: %v", err)
	}
	if refined.Text == "" {
		t.Error("refineSoul with icons returned no Text note")
	}

	stored, ok := h.deps.Souls.Get(bootstrapped.Structured.SoulID)
	if !ok {
		t.Fatal("stored soul not found after refineSoul")
	}
	if stored.IconSet["brandmark"] != validIconSVG {
		t.Fatalf("stored soul IconSet[brandmark] = %q, want %q", stored.IconSet["brandmark"], validIconSVG)
	}
}

// TestRefineSoulRejectsInvalidIcon asserts that refine_soul returns a typed
// error (no panic) when an Icons entry fails scene.ValidateIcon, and does
// not persist the change.
func TestRefineSoulRejectsInvalidIcon(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	bootstrapped, err := h.bootstrapSoul(ctx, contracts.BootstrapSoulInput{Name: "Bad Icon Soul"})
	if err != nil {
		t.Fatalf("bootstrapSoul: %v", err)
	}

	_, err = h.refineSoul(ctx, contracts.RefineSoulInput{
		SoulID: bootstrapped.Structured.SoulID,
		Icons:  map[string]string{"bad-glyph": arcIconSVG},
	})
	if err == nil {
		t.Fatal("refineSoul with an invalid icon: error = nil, want a typed validation error")
	}

	stored, ok := h.deps.Souls.Get(bootstrapped.Structured.SoulID)
	if !ok {
		t.Fatal("stored soul not found after refineSoul")
	}
	if len(stored.IconSet) != 0 {
		t.Fatalf("stored soul IconSet = %v, want unchanged/empty after a rejected refine", stored.IconSet)
	}
}

func TestGetSoulIncludesDeckardWhiteStyleGuide(t *testing.T) {
	h := testHandlers()

	got, err := h.getSoul(context.Background(), contracts.GetSoulInput{SoulID: soul.DeckardWhiteID, IncludeStyleGuide: true})
	if err != nil {
		t.Fatalf("getSoul: %v", err)
	}
	if got.Structured.StyleGuide == nil {
		t.Fatal("getSoul style guide = nil, want value")
	}
	if got.Structured.StyleGuide.NorthStar == "" {
		t.Fatal("getSoul north star empty, want Deckard White voice")
	}
	if len(got.Structured.StyleGuide.Do) == 0 || len(got.Structured.StyleGuide.Dont) == 0 {
		t.Fatalf("getSoul style guide = %+v, want do/dont guidance", got.Structured.StyleGuide)
	}
}

func TestBootstrapSoulFromTemplateExtractsBrandAccent(t *testing.T) {
	h := testHandlers()
	ctx := context.Background()

	theme := pptx.NewTheme(pptx.WithAccent("DB2777"), pptx.WithFonts("Georgia", "Verdana"))
	path := writeBrandPPTXFixture(t, theme)

	got, err := h.bootstrapSoulFromTemplate(ctx, contracts.BootstrapSoulFromTemplateInput{Name: "Acme", Path: path})
	if err != nil {
		t.Fatalf("bootstrapSoulFromTemplate: %v", err)
	}
	if got.Structured.SoulID != "acme" {
		t.Fatalf("SoulID = %q, want acme", got.Structured.SoulID)
	}
	if got.Structured.ExtractedColors["accent"] != "DB2777" {
		t.Fatalf("ExtractedColors[accent] = %q, want DB2777", got.Structured.ExtractedColors["accent"])
	}
	if _, ok := h.deps.Souls.Get("acme"); !ok {
		t.Fatal("soul \"acme\" not found in store after bootstrap_soul_from_template")
	}
}

func TestBootstrapSoulFromTemplateMissingPath(t *testing.T) {
	h := testHandlers()
	_, err := h.bootstrapSoulFromTemplate(context.Background(), contracts.BootstrapSoulFromTemplateInput{Name: "Acme", Path: filepath.Join(t.TempDir(), "missing.pptx")})
	if err == nil {
		t.Fatal("expected error for missing brand template path")
	}
}

func TestBootstrapSoulFromTemplateRejectsNonPPTX(t *testing.T) {
	h := testHandlers()
	path := filepath.Join(t.TempDir(), "brand.txt")
	if err := os.WriteFile(path, []byte("not a pptx"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := h.bootstrapSoulFromTemplate(context.Background(), contracts.BootstrapSoulFromTemplateInput{Name: "Acme", Path: path})
	if err == nil {
		t.Fatal("expected error for non-.pptx brand template path")
	}
}

func TestBootstrapSoulFromTemplateRejectsEmptyName(t *testing.T) {
	h := testHandlers()
	path := filepath.Join(t.TempDir(), "brand.pptx")
	if err := os.WriteFile(path, []byte("irrelevant"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := h.bootstrapSoulFromTemplate(context.Background(), contracts.BootstrapSoulFromTemplateInput{Name: "  ", Path: path})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// brandThemePartName is the package-relative entry pptx-go writes the active
// theme's OOXML scheme to (ppt/theme/theme1.xml); see writeBrandPPTXFixture.
const brandThemePartName = "ppt/theme/theme1.xml"

// writeBrandPPTXFixture builds a minimal, valid .pptx in t.TempDir() whose
// theme1.xml carries theme's color scheme + fonts, mimicking a real brand kit
// authored in PowerPoint (R8.2's source material). pptx.New(pptx.WithTheme)
// alone is not enough: WithTheme only drives in-process token resolution for
// rendering, it does not persist into the written theme1.xml part (that part
// is seeded once, at New(), from the engine's fixed scaffold). So this builds
// a normal scaffolded deck, then swaps the theme1.xml zip entry for the bytes
// pptx.Theme.ThemeXML() produces for the caller's theme — the same OOXML
// shape NewFromBytes' theme codec reads on open (see themecodec_test.go's
// TestThemeRoundTripOOXML for the same round-trip via theme.ThemePart
// directly). Returns the written file's path.
func writeBrandPPTXFixture(t *testing.T, theme *pptx.Theme) string {
	t.Helper()

	pres := pptx.New()
	pres.AddSlide()
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	themeXML, err := theme.ThemeXML()
	if err != nil {
		t.Fatalf("ThemeXML: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", f.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip entry %s: %v", f.Name, err)
		}
		if f.Name == brandThemePartName {
			content = themeXML
		}
		fw, err := w.Create(f.Name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", f.Name, err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatalf("write zip entry %s: %v", f.Name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip.Writer.Close: %v", err)
	}

	path := filepath.Join(t.TempDir(), "brand.pptx")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func tokenValue(tokens []contracts.TokenEntry, layer contracts.TokenLayer, name string) string {
	for _, token := range tokens {
		if token.Layer == layer && token.Name == name {
			return token.Value
		}
	}
	return ""
}
