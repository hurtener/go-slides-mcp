package render

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// TestMapBackgroundStopsMeshNilByteIdentical asserts mapBackground leaves
// scene.Background.Stops and .Mesh nil when the product Background carries
// no Stops/Mesh — the byte-identical opt-out invariant (R13.2/R13.3/R13.4).
func TestMapBackgroundStopsMeshNilByteIdentical(t *testing.T) {
	t.Parallel()

	got := mapBackground(contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorAccent})
	if got.Stops != nil {
		t.Errorf("Stops = %+v, want nil", got.Stops)
	}
	if got.Mesh != nil {
		t.Errorf("Mesh = %+v, want nil", got.Mesh)
	}

	want := scene.Background{
		Kind:  scene.BackgroundColor,
		Color: mapColorRole(contracts.ColorAccent),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mapBackground() = %+v, want %+v", got, want)
	}
}

// TestMapBackgroundStopsAndMesh asserts every Stops/Mesh field maps 1:1 into
// the scene types (R13.2/R13.3/R13.4).
func TestMapBackgroundStopsAndMesh(t *testing.T) {
	t.Parallel()

	b := contracts.Background{
		Kind: contracts.BackgroundRadial,
		Stops: []contracts.GradientStop{
			{Pos: 0, Color: contracts.ColorAccent},
			{Pos: 0.5, Color: contracts.ColorAccentAlt},
			{Pos: 1, Color: contracts.ColorSurface},
		},
		Mesh: []contracts.MeshGlow{
			{Anchor: contracts.AnchorTopLeft, Color: contracts.ColorAccent, Radius: 120, Alpha: 0.12},
			{Anchor: contracts.AnchorBottomRight, Color: contracts.ColorAccentAlt, Radius: 90, Alpha: 0.08},
		},
	}
	got := mapBackground(b)

	if len(got.Stops) != 3 {
		t.Fatalf("Stops length = %d, want 3", len(got.Stops))
	}
	for i, s := range b.Stops {
		if got.Stops[i].Pos != s.Pos || got.Stops[i].Color != mapColorRole(s.Color) {
			t.Errorf("Stops[%d] = %+v, want Pos=%v Color=%v", i, got.Stops[i], s.Pos, mapColorRole(s.Color))
		}
	}

	if len(got.Mesh) != 2 {
		t.Fatalf("Mesh length = %d, want 2", len(got.Mesh))
	}
	for i, m := range b.Mesh {
		wantAlpha := int(m.Alpha * 100000)
		if got.Mesh[i].Anchor != mapAnchor(m.Anchor) || got.Mesh[i].Color != mapColorRole(m.Color) ||
			got.Mesh[i].Alpha != wantAlpha {
			t.Errorf("Mesh[%d] = %+v, want Anchor=%v Color=%v Alpha=%d", i, got.Mesh[i], mapAnchor(m.Anchor), mapColorRole(m.Color), wantAlpha)
		}
	}
}

// radialBackgroundDoc builds a one-slide cover doc whose Background is a
// 3-stop radial gradient (R13.2/R13.3).
func radialBackgroundDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Radial Background",
		Slides: []contracts.Slide{{
			ID:     "cover",
			Layout: contracts.LayoutCover,
			Background: &contracts.Background{
				Kind: contracts.BackgroundRadial,
				Stops: []contracts.GradientStop{
					{Pos: 0, Color: contracts.ColorAccent},
					{Pos: 0.5, Color: contracts.ColorAccentAlt},
					{Pos: 1, Color: contracts.ColorSurface},
				},
			},
			Nodes: []contracts.SlideNode{
				&contracts.Hero{Title: "Radial Spotlight"},
			},
		}},
	}
}

// meshBackgroundDoc builds a one-slide cover doc whose Background is a mesh
// wash with 2 pooled glows (R13.4).
func meshBackgroundDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Mesh Background",
		Slides: []contracts.Slide{{
			ID:     "cover",
			Layout: contracts.LayoutCover,
			Background: &contracts.Background{
				Kind: contracts.BackgroundMesh,
				Mesh: []contracts.MeshGlow{
					{Anchor: contracts.AnchorTopLeft, Color: contracts.ColorAccent, Radius: 240, Alpha: 0.12},
					{Anchor: contracts.AnchorBottomRight, Color: contracts.ColorAccentAlt, Radius: 200, Alpha: 0.08},
				},
			},
			Nodes: []contracts.SlideNode{
				&contracts.Hero{Title: "Mesh Wash"},
			},
		}},
	}
}

// noBackgroundDoc mirrors radialBackgroundDoc/meshBackgroundDoc's shape with
// Background left nil (BackgroundNone) — the shape-count baseline.
func noBackgroundDoc() contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "No Background",
		Slides: []contracts.Slide{{
			ID:     "cover",
			Layout: contracts.LayoutCover,
			Nodes: []contracts.SlideNode{
				&contracts.Hero{Title: "Plain Cover"},
			},
		}},
	}
}

// TestRenderRadialBackgroundEmitsMoreShapesThanNone is the R13.2 product-level
// accept case: a radial background with 3 Stops renders without error and
// emits more shapes than the same slide with no background (proves effect,
// not dead infra).
func TestRenderRadialBackgroundEmitsMoreShapesThanNone(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	baseBuf, baseStats, err := Render(noBackgroundDoc(), s)
	if err != nil {
		t.Fatalf("Render(none) error = %v", err)
	}
	if len(baseBuf) == 0 {
		t.Fatal("Render(none) returned empty bytes")
	}
	radialBuf, radialStats, err := Render(radialBackgroundDoc(), s)
	if err != nil {
		t.Fatalf("Render(radial) error = %v", err)
	}
	if len(radialBuf) == 0 {
		t.Fatal("Render(radial) returned empty bytes")
	}
	if radialStats.Shapes <= baseStats.Shapes {
		t.Errorf("radial Shapes = %d, want > none Shapes %d", radialStats.Shapes, baseStats.Shapes)
	}
}

// TestRenderMeshBackgroundEmitsShapesAndIsDeterministic is the R13.4
// product-level accept case: a mesh background with 2 glows renders without
// error, emits more shapes than a no-background slide, and produces
// byte-identical output across repeated renders and worker counts (the
// render-determinism hard contract, CLAUDE.md §5), mirroring
// TestRenderDeterministicAcrossRepeatedRenders / TestRenderDeterministicAcrossWorkerCounts.
func TestRenderMeshBackgroundEmitsShapesAndIsDeterministic(t *testing.T) {
	t.Parallel()

	doc := meshBackgroundDoc()
	s := soul.DeckardWhite()

	baseBuf, baseStats, err := Render(noBackgroundDoc(), s)
	if err != nil {
		t.Fatalf("Render(none) error = %v", err)
	}
	if len(baseBuf) == 0 {
		t.Fatal("Render(none) returned empty bytes")
	}

	first, firstStats, err := Render(doc, s)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	if firstStats.Shapes <= baseStats.Shapes {
		t.Errorf("mesh Shapes = %d, want > none Shapes %d", firstStats.Shapes, baseStats.Shapes)
	}

	second, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("mesh background Render() bytes differ across identical renders")
	}

	defaultWorkers, _, err := renderWithWorkers(doc, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(doc, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("mesh background render bytes differ across worker counts")
	}
}
