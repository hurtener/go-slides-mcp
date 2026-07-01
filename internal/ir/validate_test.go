package ir

import (
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// validDeck is a structurally valid deck exercising several node kinds + nesting.
func validDeck() contracts.SlideDoc {
	return contracts.SlideDoc{Title: "ok", Slides: []contracts.Slide{{
		ID:     "s1",
		Layout: contracts.LayoutTwoColumn,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Text: contracts.RichText{{Text: "T"}}, Level: 1},
			&contracts.TwoColumn{
				Left:  []contracts.SlideNode{&contracts.List{Items: []contracts.ListItem{{Text: contracts.RichText{{Text: "a"}}}}}},
				Right: []contracts.SlideNode{&contracts.Card{Header: "C", Body: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "p"}}}}}}},
			},
			&contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{
				&contracts.Callout{Kind: contracts.CalloutTip, Body: contracts.RichText{{Text: "x"}}},
				&contracts.Image{AssetID: "asset://1"},
			}},
			&contracts.Table{Headers: []contracts.RichText{{{Text: "h1"}}, {{Text: "h2"}}}, Rows: [][]contracts.RichText{
				{{{Text: "a"}}, {{Text: "b"}}},
			}},
		},
	}}}
}

func TestValidateDocOK(t *testing.T) {
	if err := ValidateDoc(validDeck()); err != nil {
		t.Fatalf("valid deck should pass, got: %v", err)
	}
}

func TestValidateNodeRules(t *testing.T) {
	cases := []struct {
		name string
		node contracts.SlideNode
		want string // substring expected in the error
	}{
		{"heading-level-low", &contracts.Heading{Level: 0}, "out of range"},
		{"heading-level-high", &contracts.Heading{Level: 7}, "out of range"},
		{"list-empty", &contracts.List{}, "at least one item"},
		{"list-bad-kind", &contracts.List{Kind: "bogus", Items: []contracts.ListItem{{}}}, "want one of"},
		{"callout-bad-kind", &contracts.Callout{Kind: "bogus"}, "want one of"},
		{"image-no-asset", &contracts.Image{}, "empty assetId"},
		{"image-crop-oob", &contracts.Image{AssetID: "a", Crop: contracts.Crop{Left: 1.5}}, "out of [0,1]"},
		{"image-crop-sum", &contracts.Image{AssetID: "a", Crop: contracts.Crop{Left: 0.6, Right: 0.6}}, "left+right"},
		{"chart-no-asset", &contracts.Chart{}, "empty assetId"},
		{"code-no-asset", &contracts.CodeBlock{}, "empty assetId"},
		{"flow-no-steps", &contracts.Flow{}, "at least one step"},
		{"table-no-headers", &contracts.Table{}, "at least one header"},
		{"table-row-width", &contracts.Table{Headers: []contracts.RichText{{{Text: "h"}}}, Rows: [][]contracts.RichText{{{{Text: "a"}}, {{Text: "b"}}}}}, "row[0] width"},
		{"table-style-highlight-oob", &contracts.Table{Headers: []contracts.RichText{{{Text: "h1"}}, {{Text: "h2"}}}, Style: &contracts.TableStyle{HighlightCol: 3}}, "highlightCol"},
		{"table-style-group-span-zero", &contracts.Table{Headers: []contracts.RichText{{{Text: "h1"}}, {{Text: "h2"}}}, Style: &contracts.TableStyle{HeaderGroups: []contracts.HeaderGroup{{Label: "g", Span: 0}}}}, "must be >= 1"},
		{"table-style-group-span-sum", &contracts.Table{Headers: []contracts.RichText{{{Text: "h1"}}, {{Text: "h2"}}}, Style: &contracts.TableStyle{HeaderGroups: []contracts.HeaderGroup{{Label: "g", Span: 1}}}}, "span sum"},
		{"twocol-empty", &contracts.TwoColumn{Left: []contracts.SlideNode{&contracts.Hero{}}}, "right must be non-empty"},
		{"grid-cols", &contracts.Grid{Columns: 5, Cells: []contracts.SlideNode{&contracts.Hero{}}}, "out of range 2..4"},
		{"grid-multiple", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Hero{}}}, "not a multiple"},
		{"cardsection-empty", &contracts.CardSection{Header: "h"}, "must be non-empty"},
		{"decoration-preset", &contracts.Decoration{Kind: contracts.DecorationPreset}, "needs a preset"},
		{"decoration-asset", &contracts.Decoration{Kind: contracts.DecorationAsset}, "needs an assetId"},
		{"decoration-kind", &contracts.Decoration{}, "want one of"},
		{"decoration-opacity", &contracts.Decoration{Kind: contracts.DecorationPreset, Preset: "p", Opacity: 2}, "opacity"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateNode(c.node)
			if err == nil {
				t.Fatalf("want error containing %q, got nil", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want error containing %q, got: %v", c.want, err)
			}
		})
	}
}

// TestValidateBackgroundRules exercises the Stage-1 structural checks on a
// slide's Background.Stops/Mesh (R13.2/R13.3/R13.4), mirroring the
// table-style-* cases in TestValidateNodeRules above but at the slide level
// (validateBackground is called from ValidateSlide, not ValidateNode).
func TestValidateBackgroundRules(t *testing.T) {
	cases := []struct {
		name  string
		slide contracts.Slide
		want  string
	}{
		{"bg-nil", contracts.Slide{}, ""},
		{"bg-stops-empty", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundGradient}}, ""},
		{"bg-stops-too-few", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{{Pos: 0, Color: contracts.ColorAccent}}}}, "need 2..8 entries"},
		{"bg-stops-too-many", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{
			{Pos: 0}, {Pos: 0.1}, {Pos: 0.2}, {Pos: 0.3}, {Pos: 0.4}, {Pos: 0.5}, {Pos: 0.6}, {Pos: 0.7}, {Pos: 0.8},
		}}}, "need 2..8 entries"},
		{"bg-stops-pos-oob", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{{Pos: -0.1}, {Pos: 1.5}}}}, "out of [0,1]"},
		{"bg-stops-not-ascending", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{{Pos: 0.5}, {Pos: 0.2}}}}, "not strictly ascending"},
		{"bg-mesh-radius-negative", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundMesh, Mesh: []contracts.MeshGlow{{Radius: -1}}}}, "out of range, must be >= 0"},
		{"bg-mesh-alpha-oob", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundMesh, Mesh: []contracts.MeshGlow{{Radius: 100, Alpha: 1.5}}}}, "out of [0,1]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSlide(c.slide)
			if c.want == "" {
				if err != nil {
					t.Fatalf("want nil, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want error containing %q, got nil", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want error containing %q, got: %v", c.want, err)
			}
		})
	}
}

// TestValidateRecursesIntoContainers proves a bad node nested deep in a
// container surfaces through ValidateSlide.
func TestValidateRecursesIntoContainers(t *testing.T) {
	s := contracts.Slide{ID: "s", Nodes: []contracts.SlideNode{
		&contracts.TwoColumn{
			Left:  []contracts.SlideNode{&contracts.Card{Body: []contracts.SlideNode{&contracts.Heading{Level: 99}}}},
			Right: []contracts.SlideNode{&contracts.Hero{Title: "ok"}},
		},
	}}
	err := ValidateSlide(s)
	if err == nil {
		t.Fatal("want nested validation error, got nil")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("want nested heading error, got: %v", err)
	}
}
