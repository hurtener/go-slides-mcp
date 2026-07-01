package render

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// annotatedPNG is a 1x1 transparent PNG, reused from the render_assets_test
// pattern for asset-bearing image fixtures.
const annotatedPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

// imageAnnotationsSample is a fully-populated overlay exercising 3 pins
// (one with a caption/leader) and 1 highlight rectangle (R14.17).
func imageAnnotationsSample() *contracts.ImageAnnotations {
	return &contracts.ImageAnnotations{
		Pins: []contracts.ImagePin{
			{X: 0.1, Y: 0.1, Label: "1", Caption: "The widget", AccentIndex: 0},
			{X: 0.5, Y: 0.5, Label: "2", AccentIndex: 1},
			{X: 0.9, Y: 0.9, Label: "3", AccentIndex: 2},
		},
		Highlights: []contracts.ImageHighlight{
			{X: 0.2, Y: 0.2, W: 0.3, H: 0.2, AccentIndex: 0},
		},
	}
}

func annotatedImageDoc(id contracts.AssetID, annotations *contracts.ImageAnnotations) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Image annotations",
		Slides: []contracts.Slide{
			{
				ID:     "annotated",
				Layout: contracts.LayoutTitleContent,
				Nodes: []contracts.SlideNode{
					&contracts.Hero{Title: "Image annotations"},
					&contracts.Image{AssetID: id, Alt: "annotated image", Annotations: annotations},
				},
			},
		},
	}
}

func annotatedStoreAndDoc(t *testing.T, annotations *contracts.ImageAnnotations) (contracts.SlideDoc, raster.StoreResolver) {
	t.Helper()

	png, err := base64.StdEncoding.DecodeString(annotatedPNG)
	if err != nil {
		t.Fatalf("base64 decode PNG: %v", err)
	}
	store := asset.NewMemoryStore()
	stored, err := store.Put("annotated.png", "image/png", png)
	if err != nil {
		t.Fatalf("store.Put() error = %v", err)
	}
	return annotatedImageDoc(contracts.AssetID(stored.ID), annotations), raster.NewStoreResolver(store)
}

// TestRenderImageAnnotationsEmitsMoreShapesThanPlain is the R14.17
// product-level accept case: an Image with 3 pins + 1 highlight renders
// without error and emits strictly more native shapes than the same image
// left plain — the pins, leaders/captions, and highlight rectangle are all
// additional <p:sp> shapes the engine draws over the picture.
func TestRenderImageAnnotationsEmitsMoreShapesThanPlain(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	plainDoc, plainResolver := annotatedStoreAndDoc(t, nil)
	plainBuf, _, err := RenderWithAssets(plainDoc, s, plainResolver)
	if err != nil {
		t.Fatalf("RenderWithAssets(plain) error = %v", err)
	}

	annotatedDoc, annotatedResolver := annotatedStoreAndDoc(t, imageAnnotationsSample())
	annotatedBuf, _, err := RenderWithAssets(annotatedDoc, s, annotatedResolver)
	if err != nil {
		t.Fatalf("RenderWithAssets(annotated) error = %v", err)
	}

	plainXML := string(firstSlideXML(t, plainBuf))
	annotatedXML := string(firstSlideXML(t, annotatedBuf))

	plainShapes := strings.Count(plainXML, "<p:sp>")
	annotatedShapes := strings.Count(annotatedXML, "<p:sp>")
	if annotatedShapes <= plainShapes {
		t.Errorf("annotated image shape count = %d, want > plain image shape count %d", annotatedShapes, plainShapes)
	}
	if !strings.Contains(annotatedXML, "<a:t>1</a:t>") {
		t.Errorf("annotated image missing pin label text:\n%s", annotatedXML)
	}
}

// TestRenderImageNilAnnotationsByteIdentical asserts an Image whose
// Annotations field is nil renders byte-identical across repeated renders —
// mapImageAnnotations(nil) must map to nil, so the R14.17 field is inert.
func TestRenderImageNilAnnotationsByteIdentical(t *testing.T) {
	t.Parallel()

	if got := mapImageAnnotations(nil); got != nil {
		t.Errorf("mapImageAnnotations(nil) = %+v, want nil", got)
	}

	s := soul.DeckardWhite()
	doc, resolver := annotatedStoreAndDoc(t, nil)

	first, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() error = %v", err)
	}
	second, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("RenderWithAssets() second call error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("RenderWithAssets() bytes differ across identical nil-Annotations renders (determinism broken)")
	}
}

// TestRenderImageAnnotationsDeterministic asserts an annotated Image renders
// byte-identically across repeated renders and across worker counts (the
// render determinism contract, CLAUDE.md §5).
func TestRenderImageAnnotationsDeterministic(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()
	doc, resolver := annotatedStoreAndDoc(t, imageAnnotationsSample())

	first, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("first RenderWithAssets() error = %v", err)
	}
	second, _, err := RenderWithAssets(doc, s, resolver)
	if err != nil {
		t.Fatalf("second RenderWithAssets() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("RenderWithAssets() bytes differ across identical annotated renders (determinism broken)")
	}

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1, resolver)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("annotated image render bytes differ across worker counts")
	}
}
