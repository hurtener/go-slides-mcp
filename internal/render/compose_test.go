package render

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// decoredSoul returns a Bootstrap-produced soul, which R13-D wires to always
// carry a non-nil DecorPolicy (soul.DefaultDecorPolicy).
func decoredSoul(t *testing.T) *soul.Soul {
	t.Helper()
	s, err := soul.Bootstrap(soul.BootstrapParams{Name: "Compose Test"})
	if err != nil {
		t.Fatalf("soul.Bootstrap() error = %v", err)
	}
	return s
}

func TestApplyDecorPolicy_NilPolicyIsIdentityRender(t *testing.T) {
	t.Parallel()

	doc := testDoc()
	white := soul.DeckardWhite()

	direct, _, err := Render(doc, white)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	composed := applyDecorPolicy(doc, white)
	viaCompose, _, err := Render(composed, white)
	if err != nil {
		t.Fatalf("Render(composed) error = %v", err)
	}
	if !bytes.Equal(direct, viaCompose) {
		t.Fatal("applyDecorPolicy changed render bytes for a nil-Decor soul")
	}
}

func TestApplyDecorPolicy_EmptyByArchetypeIsIdentity(t *testing.T) {
	t.Parallel()

	doc := testDoc()
	s := soul.DeckardWhite()
	s.Decor = &contracts.DecorPolicy{}

	got := applyDecorPolicy(doc, s)
	if !reflect.DeepEqual(got, doc) {
		t.Fatal("applyDecorPolicy mutated doc for an empty ByArchetype policy")
	}
}

func TestApplyDecorPolicy_ContentSlideGetsPaperAndTexture(t *testing.T) {
	t.Parallel()

	s := decoredSoul(t)
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Cover"}}},
		{ID: "body", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.Heading{Text: rt("Body")}}},
	}}

	got := applyDecorPolicy(doc, s)
	body := got.Slides[1]
	if body.Background == nil || body.Background.Color != contracts.ColorPaper {
		t.Fatalf("body.Background = %+v, want Color=paper", body.Background)
	}
	if len(body.Nodes) == 0 {
		t.Fatal("body.Nodes is empty, want the texture decoration prepended")
	}
	if _, ok := body.Nodes[0].(*contracts.Decoration); !ok {
		t.Fatalf("body.Nodes[0] = %T, want *contracts.Decoration", body.Nodes[0])
	}
}

func TestApplyDecorPolicy_ExplicitBackgroundNotOverwritten(t *testing.T) {
	t.Parallel()

	s := decoredSoul(t)
	explicit := &contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorAccent}
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover},
		{ID: "body", Layout: contracts.LayoutTitleContent, Background: explicit},
	}}

	got := applyDecorPolicy(doc, s)
	if got.Slides[1].Background != explicit {
		t.Fatalf("Background = %+v, want the untouched explicit pointer %+v", got.Slides[1].Background, explicit)
	}
}

func TestApplyDecorPolicy_ExplicitDecorationNotPrepended(t *testing.T) {
	t.Parallel()

	s := decoredSoul(t)
	own := &contracts.Decoration{Kind: contracts.DecorationPreset, Preset: "glow_ring"}
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover},
		{ID: "body", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{own}},
	}}

	got := applyDecorPolicy(doc, s)
	if len(got.Slides[1].Nodes) != 1 || got.Slides[1].Nodes[0] != contracts.SlideNode(own) {
		t.Fatalf("Nodes = %+v, want unchanged single-node slice with the caller's own Decoration", got.Slides[1].Nodes)
	}
}

func TestApplyDecorPolicy_DeterministicAcrossWorkerCounts(t *testing.T) {
	t.Parallel()

	// The nil-decor determinism-across-workers test (render_test.go) exercises
	// only the flat path. A decor policy adds background gradients + prepended
	// ornament nodes; assert THOSE render byte-identically regardless of worker
	// count too (render determinism is a hard contract, CLAUDE.md §5).
	s := decoredSoul(t)
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Cover"}}},
		{ID: "dark", Variant: contracts.VariantDark, Nodes: []contracts.SlideNode{&contracts.Heading{Level: 2, Text: rt("Dark")}}},
		{ID: "body", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.Heading{Level: 2, Text: rt("Body")}}},
	}}
	composed := applyDecorPolicy(doc, s)

	defaultWorkers, _, err := renderWithWorkers(composed, s, 0, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(default) error = %v", err)
	}
	oneWorker, _, err := renderWithWorkers(composed, s, 1, nil)
	if err != nil {
		t.Fatalf("renderWithWorkers(1) error = %v", err)
	}
	if !bytes.Equal(defaultWorkers, oneWorker) {
		t.Fatal("decorated render differs across worker counts (determinism contract broken)")
	}
}

func TestInferArchetype(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		s    contracts.Slide
		idx  int
		n    int
		want contracts.SlideArchetype
	}{
		{"dark variant", contracts.Slide{Variant: contracts.VariantDark}, 3, 5, contracts.ArchetypeDark},
		{"cover layout", contracts.Slide{Layout: contracts.LayoutCover}, 2, 5, contracts.ArchetypeCover},
		{"index zero", contracts.Slide{Layout: contracts.LayoutTitleContent}, 0, 5, contracts.ArchetypeCover},
		{"full bleed", contracts.Slide{Layout: contracts.LayoutFullBleed}, 2, 5, contracts.ArchetypeSection},
		{"else content", contracts.Slide{Layout: contracts.LayoutTitleContent}, 2, 5, contracts.ArchetypeContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := inferArchetype(tc.s, tc.idx, tc.n); got != tc.want {
				t.Errorf("inferArchetype() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInferArchetype_ExplicitWinsOverEveryInferenceSignal(t *testing.T) {
	t.Parallel()

	// applyDecorPolicy only calls inferArchetype when Slide.Archetype == "";
	// an explicit archetype is read straight off the slide instead. Assert
	// that end-to-end: a dark-variant, cover-layout, index-0 slide explicitly
	// marked "content" resolves to the content recipe, not dark/cover.
	s := decoredSoul(t)
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s0", Layout: contracts.LayoutCover, Variant: contracts.VariantDark, Archetype: contracts.ArchetypeContent},
	}}
	got := applyDecorPolicy(doc, s)
	if got.Slides[0].Background == nil || got.Slides[0].Background.Color != contracts.ColorPaper {
		t.Fatalf("Background = %+v, want the content archetype's paper fill", got.Slides[0].Background)
	}
}

func TestApplyDecorPolicy_DeterministicAcrossRepeatedCalls(t *testing.T) {
	t.Parallel()

	s := decoredSoul(t)
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "cover", Layout: contracts.LayoutCover},
		{ID: "body", Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{&contracts.Heading{Level: 2, Text: rt("Body")}}},
	}}

	first := applyDecorPolicy(doc, s)
	second := applyDecorPolicy(doc, s)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("applyDecorPolicy is not deterministic across repeated calls on the same input")
	}

	r1, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	r2, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !bytes.Equal(r1, r2) {
		t.Fatal("Render() with a decor policy is not byte-deterministic across repeated calls")
	}
}
