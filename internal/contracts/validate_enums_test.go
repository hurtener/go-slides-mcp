package contracts_test

import (
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// TestValidateNodeEnumsKnownGood confirms that well-formed nodes with explicit
// valid enum values produce no error.
func TestValidateNodeEnumsKnownGood(t *testing.T) {
	cases := []struct {
		name string
		node contracts.SlideNode
	}{
		{"callout-note", &contracts.Callout{Kind: contracts.CalloutNote}},
		{"callout-warning", &contracts.Callout{Kind: contracts.CalloutWarning}},
		{"callout-tip", &contracts.Callout{Kind: contracts.CalloutTip}},
		{"callout-important", &contracts.Callout{Kind: contracts.CalloutImportant}},
		{"list-bullet", &contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{{}}}},
		{"list-number", &contracts.List{Kind: contracts.ListNumber, Items: []contracts.ListItem{{}}}},
		{"list-checklist", &contracts.List{Kind: contracts.ListChecklist, Items: []contracts.ListItem{{}}}},
		{"chip-tone-tint", &contracts.Chip{Tone: contracts.ChipTint}},
		{"chip-tone-solid", &contracts.Chip{Tone: contracts.ChipSolid}},
		{"chip-color", &contracts.Chip{Color: contracts.ColorAccent}},
		{"chip-align", &contracts.Chip{Align: contracts.HAlignCenter}},
		{"arrow-right", &contracts.Arrow{Direction: contracts.ArrowRight}},
		{"arrow-left", &contracts.Arrow{Direction: contracts.ArrowLeft}},
		{"arrow-up", &contracts.Arrow{Direction: contracts.ArrowUp}},
		{"arrow-down", &contracts.Arrow{Direction: contracts.ArrowDown}},
		{"divider-spacing", &contracts.Divider{Spacing: contracts.SpaceMD}},
		{"grid-gap", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Hero{}, &contracts.Hero{}}, Gap: contracts.SpaceLG}},
		{"two-column-ratio-11", &contracts.TwoColumn{Ratio: contracts.Ratio11, Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"two-column-ratio-12", &contracts.TwoColumn{Ratio: contracts.Ratio12, Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"two-column-ratio-21", &contracts.TwoColumn{Ratio: contracts.Ratio21, Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"flow-horizontal", &contracts.Flow{Orientation: contracts.FlowHorizontal, Steps: []contracts.FlowStep{{}}}},
		{"flow-connector-arrow", &contracts.Flow{Connector: contracts.ConnectorArrow, Steps: []contracts.FlowStep{{}}}},
		{"flow-connector-dashed", &contracts.Flow{Connector: contracts.ConnectorArrowDashed, Steps: []contracts.FlowStep{{}}}},
		{"flow-connector-cycle", &contracts.Flow{Connector: contracts.ConnectorCycle, Steps: []contracts.FlowStep{{}}}},
		{"flow-connector-plus", &contracts.Flow{Connector: contracts.ConnectorPlus, Steps: []contracts.FlowStep{{}}}},
		{"card-body-layout", &contracts.Card{BodyLayout: contracts.BodyVertical}},
		{"card-border-style", &contracts.Card{BorderStyle: contracts.BorderSolid}},
		{"card-size-sm", &contracts.Card{Size: contracts.CardSizeSM}},
		{"card-layout-icon-top", &contracts.Card{Layout: contracts.CardLayoutIconTop}},
		{"card-elevation", &contracts.Card{Elevation: contracts.ElevationRaised}},
		{"card-fill", &contracts.Card{Fill: contracts.ColorSurface}},
		{"stat-delta-up", &contracts.Stat{Value: "$2,200", Label: "ARR", Delta: "+18%", DeltaTone: contracts.DeltaUp}},
		{"stat-delta-down", &contracts.Stat{Value: "14 days", Label: "cycle", Delta: "-2 days", DeltaTone: contracts.DeltaDown}},
		{"stat-delta-neutral", &contracts.Stat{Value: "98%", Label: "NPS", Delta: "±0", DeltaTone: contracts.DeltaNeutral}},
		{"decoration-preset", &contracts.Decoration{Kind: contracts.DecorationPreset}},
		{"decoration-asset", &contracts.Decoration{Kind: contracts.DecorationAsset}},
		{"decoration-layer-bg", &contracts.Decoration{Kind: contracts.DecorationPreset, Layer: contracts.LayerBackground}},
		{"decoration-layer-fg", &contracts.Decoration{Kind: contracts.DecorationPreset, Layer: contracts.LayerForeground}},
		{"decoration-anchor", &contracts.Decoration{Kind: contracts.DecorationPreset, Anchor: contracts.AnchorCenter}},
		{"image-frame-browser", &contracts.Image{AssetID: "a", Frame: contracts.FrameBrowser}},
		{"image-fit-fill", &contracts.Image{AssetID: "a", Fit: contracts.FitFill}},
		{"hero-align", &contracts.Hero{Align: contracts.HAlignLeft}},
		{"heading-align", &contracts.Heading{Level: 1, Align: contracts.HAlignRight}},
		{"prose-align", &contracts.Prose{Align: contracts.HAlignCenter}},
		{"section-divider-align", &contracts.SectionDivider{Align: contracts.HAlignLeft}},
		{"richtext-type-role", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", TypeRole: contracts.TypeH1}}}},
		{"richtext-color-token", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", Color: contracts.TextColor{Token: contracts.TextAccent}}}}},
		// R5 (D-055) — TwoColumn connector/badge.
		{"two-column-join-badge", &contracts.TwoColumn{Ratio: contracts.Ratio11, Join: contracts.JoinBadge, JoinLabel: "VS",
			Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"two-column-join-arrow", &contracts.TwoColumn{Ratio: contracts.Ratio11, Join: contracts.JoinArrow,
			Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := contracts.ValidateNodeEnums(tc.node); err != nil {
				t.Fatalf("want nil, got: %v", err)
			}
		})
	}
}

// TestValidateNodeEnumsOptionalEmpty confirms that optional enum fields left
// at the zero value (empty string) pass validation.
func TestValidateNodeEnumsOptionalEmpty(t *testing.T) {
	cases := []struct {
		name string
		node contracts.SlideNode
	}{
		{"callout-empty-kind", &contracts.Callout{}},
		{"list-empty-kind", &contracts.List{Items: []contracts.ListItem{{}}}},
		{"stat-empty-delta-tone", &contracts.Stat{Value: "42", Label: "units"}},
		{"chip-empty-tone", &contracts.Chip{}},
		{"chip-empty-color", &contracts.Chip{}},
		{"arrow-empty-direction", &contracts.Arrow{}},
		{"divider-empty-spacing", &contracts.Divider{}},
		{"grid-empty-gap", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Hero{}, &contracts.Hero{}}}},
		{"two-column-empty-ratio", &contracts.TwoColumn{Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"two-column-empty-join", &contracts.TwoColumn{Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}},
		{"flow-empty-orientation", &contracts.Flow{Steps: []contracts.FlowStep{{}}}},
		{"flow-empty-connector", &contracts.Flow{Steps: []contracts.FlowStep{{}}}},
		{"card-empty-fields", &contracts.Card{}},
		{"image-empty-frame", &contracts.Image{AssetID: "a"}},
		{"image-empty-fit", &contracts.Image{AssetID: "a"}},
		{"heading-empty-align", &contracts.Heading{Level: 1}},
		{"hero-empty-align", &contracts.Hero{}},
		{"richtext-empty-type-role", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x"}}}},
		{"richtext-empty-color-token", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := contracts.ValidateNodeEnums(tc.node); err != nil {
				t.Fatalf("optional-empty should pass, got: %v", err)
			}
		})
	}
}

// TestValidateNodeEnumsBadValues confirms that unknown wire values produce an
// error that names the field and lists the allowed set ("want one of ...").
func TestValidateNodeEnumsBadValues(t *testing.T) {
	cases := []struct {
		name string
		node contracts.SlideNode
		want string // substring expected in the error
	}{
		// CalloutKind — "info" is not a valid wire value
		{"callout-kind-bad", &contracts.Callout{Kind: "info"}, "want one of"},
		{"callout-kind-field", &contracts.Callout{Kind: "info"}, "calloutKind"},
		// ListKind — "ordered" is not valid
		{"list-kind-bad", &contracts.List{Kind: "ordered"}, "want one of"},
		{"list-kind-field", &contracts.List{Kind: "ordered"}, "listKind"},
		// ColumnRatio — "60:40" is not valid
		{"ratio-bad", &contracts.TwoColumn{Ratio: "60:40", Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}, "want one of"},
		{"ratio-field", &contracts.TwoColumn{Ratio: "60:40", Left: []contracts.SlideNode{&contracts.Hero{}}, Right: []contracts.SlideNode{&contracts.Hero{}}}, "ratio"},
		// ConnectorKind — "line" is not valid
		{"connector-bad", &contracts.Flow{Connector: "line", Steps: []contracts.FlowStep{{}}}, "want one of"},
		{"connector-field", &contracts.Flow{Connector: "line", Steps: []contracts.FlowStep{{}}}, "connector"},
		// DeltaTone — "positive" is not a valid wire value
		{"delta-tone-bad", &contracts.Stat{Value: "1", DeltaTone: "positive"}, "want one of"},
		{"delta-tone-field", &contracts.Stat{Value: "1", DeltaTone: "positive"}, "deltaTone"},
		// ColorRole — "blue" is not valid (e.g. as fill on card)
		{"fill-bad", &contracts.Card{Fill: "blue"}, "want one of"},
		{"fill-field", &contracts.Card{Fill: "blue"}, "fill"},
		// ChipTone
		{"chip-tone-bad", &contracts.Chip{Tone: "ghost"}, "want one of"},
		// ArrowDirection
		{"arrow-direction-bad", &contracts.Arrow{Direction: "diagonal"}, "want one of"},
		// SpaceRole (divider)
		{"spacing-bad", &contracts.Divider{Spacing: "huge"}, "want one of"},
		// SpaceRole (grid gap)
		{"gap-bad", &contracts.Grid{Columns: 2, Cells: []contracts.SlideNode{&contracts.Hero{}, &contracts.Hero{}}, Gap: "xxxl"}, "want one of"},
		// FlowOrientation
		{"flow-orientation-bad", &contracts.Flow{Orientation: "diagonal", Steps: []contracts.FlowStep{{}}}, "want one of"},
		// BodyLayout
		{"body-layout-bad", &contracts.Card{BodyLayout: "grid"}, "want one of"},
		// BorderStyle
		{"border-style-bad", &contracts.Card{BorderStyle: "dashed"}, "want one of"},
		// CardSize
		{"card-size-bad", &contracts.Card{Size: "xl"}, "want one of"},
		// CardLayout
		{"card-layout-bad", &contracts.Card{Layout: "iconLeft"}, "want one of"},
		// ElevationRole
		{"elevation-bad", &contracts.Card{Elevation: "sunken"}, "want one of"},
		// DecorationKind — empty string is also invalid (required)
		{"decoration-kind-empty", &contracts.Decoration{}, "want one of"},
		{"decoration-kind-bad", &contracts.Decoration{Kind: "pattern"}, "want one of"},
		// Layer
		{"layer-bad", &contracts.Decoration{Kind: contracts.DecorationPreset, Layer: "middle"}, "want one of"},
		// Anchor
		{"anchor-bad", &contracts.Decoration{Kind: contracts.DecorationPreset, Anchor: "top-center"}, "want one of"},
		// FrameKind
		{"frame-bad", &contracts.Image{AssetID: "a", Frame: "tablet"}, "want one of"},
		// Fit
		{"fit-bad", &contracts.Image{AssetID: "a", Fit: "contain"}, "want one of"},
		// HAlign on leaf
		{"halign-bad", &contracts.Hero{Align: "justify"}, "want one of"},
		// TypeRole in RichText
		{"type-role-bad", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", TypeRole: "huge"}}}, "want one of"},
		{"type-role-field", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", TypeRole: "huge"}}}, "typeRole"},
		// TextColorRole in RichText
		{"text-color-role-bad", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", Color: contracts.TextColor{Token: "pink"}}}}, "want one of"},
		{"text-color-role-field", &contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "x", Color: contracts.TextColor{Token: "pink"}}}}, "color.token"},
		// R5 (D-055) — ColumnJoin bad value
		{"column-join-bad", &contracts.TwoColumn{Join: "circle",
			Left:  []contracts.SlideNode{&contracts.Hero{}},
			Right: []contracts.SlideNode{&contracts.Hero{}}}, "want one of"},
		{"column-join-field", &contracts.TwoColumn{Join: "circle",
			Left:  []contracts.SlideNode{&contracts.Hero{}},
			Right: []contracts.SlideNode{&contracts.Hero{}}}, "join"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := contracts.ValidateNodeEnums(tc.node)
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got: %v", tc.want, err)
			}
		})
	}
}

// TestValidateSlideEnumsKnownGood confirms valid slide-level fields pass.
func TestValidateSlideEnumsKnownGood(t *testing.T) {
	cases := []struct {
		name  string
		slide contracts.Slide
	}{
		{"layout-cover", contracts.Slide{Layout: contracts.LayoutCover}},
		{"layout-blank", contracts.Slide{Layout: contracts.LayoutBlank}},
		{"valign", contracts.Slide{Align: contracts.Alignment{Vertical: contracts.VAlignCenter}}},
		{"halign", contracts.Slide{Align: contracts.Alignment{Horizontal: contracts.HAlignRight}}},
		{"variant-dark", contracts.Slide{Variant: contracts.VariantDark}},
		{"variant-light", contracts.Slide{Variant: contracts.VariantLight}},
		{"background-color", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundColor, Color: contracts.ColorAccent}}},
		{"background-gradient", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundGradient, Gradient: []contracts.ColorRole{contracts.ColorAccent, contracts.ColorSurface}}}},
		{"background-asset", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundAsset}}},
		{"background-radial", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{{Pos: 0, Color: contracts.ColorAccent}, {Pos: 1, Color: contracts.ColorSurface}}}}},
		{"background-mesh", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundMesh, Mesh: []contracts.MeshGlow{{Anchor: contracts.AnchorTopLeft, Color: contracts.ColorAccent, Radius: 120, Alpha: 0.1}}}}},
		{"empty-slide", contracts.Slide{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := contracts.ValidateSlideEnums(tc.slide); err != nil {
				t.Fatalf("want nil, got: %v", err)
			}
		})
	}
}

// TestValidateSlideEnumsBadValues confirms that bad slide-level enum values
// produce errors naming the field and the allowed set.
func TestValidateSlideEnumsBadValues(t *testing.T) {
	cases := []struct {
		name  string
		slide contracts.Slide
		want  string
	}{
		// LayoutKind
		{"layout-bad", contracts.Slide{Layout: "magazine"}, "want one of"},
		{"layout-field", contracts.Slide{Layout: "magazine"}, "layout"},
		// VAlign
		{"valign-bad", contracts.Slide{Align: contracts.Alignment{Vertical: "middle"}}, "want one of"},
		{"valign-field", contracts.Slide{Align: contracts.Alignment{Vertical: "middle"}}, "align.vertical"},
		// HAlign on slide
		{"halign-bad", contracts.Slide{Align: contracts.Alignment{Horizontal: "justify"}}, "want one of"},
		// Variant
		{"variant-bad", contracts.Slide{Variant: "sepia"}, "want one of"},
		{"variant-field", contracts.Slide{Variant: "sepia"}, "variant"},
		// BackgroundKind
		{"bg-kind-bad", contracts.Slide{Background: &contracts.Background{Kind: "pattern"}}, "want one of"},
		// Background ColorRole
		{"bg-color-bad", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundColor, Color: "red"}}, "want one of"},
		// Background Gradient ColorRole
		{"bg-gradient-bad", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundGradient, Gradient: []contracts.ColorRole{"red", contracts.ColorSurface}}}, "want one of"},
		// Background Stops ColorRole
		{"bg-stops-color-bad", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundRadial, Stops: []contracts.GradientStop{{Pos: 0, Color: "red"}, {Pos: 1, Color: contracts.ColorSurface}}}}, "want one of"},
		// Background Mesh Anchor
		{"bg-mesh-anchor-bad", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundMesh, Mesh: []contracts.MeshGlow{{Anchor: "middle", Color: contracts.ColorAccent}}}}, "want one of"},
		// Background Mesh ColorRole
		{"bg-mesh-color-bad", contracts.Slide{Background: &contracts.Background{Kind: contracts.BackgroundMesh, Mesh: []contracts.MeshGlow{{Anchor: contracts.AnchorCenter, Color: "red"}}}}, "want one of"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := contracts.ValidateSlideEnums(tc.slide)
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got: %v", tc.want, err)
			}
		})
	}
}

// TestValidateSlideEnumsFullyValidSlide confirms that a fully-valid complex
// slide (non-empty enums, nested content) produces no false rejections.
func TestValidateSlideEnumsFullyValidSlide(t *testing.T) {
	slide := contracts.Slide{
		Layout:  contracts.LayoutTwoColumn,
		Variant: contracts.VariantLight,
		Align: contracts.Alignment{
			Vertical:   contracts.VAlignTop,
			Horizontal: contracts.HAlignLeft,
		},
		Background: &contracts.Background{
			Kind:  contracts.BackgroundColor,
			Color: contracts.ColorCanvas,
		},
	}
	if err := contracts.ValidateSlideEnums(slide); err != nil {
		t.Fatalf("fully-valid slide should pass enum validation, got: %v", err)
	}
}

// TestValidateNodeEnumsFullyValidNode confirms that a complex node with
// multiple explicit valid enum values produces no error.
func TestValidateNodeEnumsFullyValidNode(t *testing.T) {
	card := &contracts.Card{
		BodyLayout:  contracts.BodyVertical,
		Fill:        contracts.ColorSurface,
		BorderStyle: contracts.BorderSolid,
		Size:        contracts.CardSizeMD,
		Layout:      contracts.CardLayoutDefault,
		Elevation:   contracts.ElevationFlat,
		Body: []contracts.SlideNode{
			&contracts.Prose{Align: contracts.HAlignLeft, Paragraphs: []contracts.RichText{
				{{Text: "hello", TypeRole: contracts.TypeBody, Color: contracts.TextColor{Token: contracts.TextPrimary}}},
			}},
		},
	}
	if err := contracts.ValidateNodeEnums(card); err != nil {
		t.Fatalf("fully-valid card should pass, got: %v", err)
	}
}

// TestAllowedVAlignCoversFill guards enum-const/allowed-set drift: R2 added
// VAlign "fill" without updating AllowedVAlign, so a valid value was rejected
// by enum validation until a deck actually used it. Adding an enum value must
// update its Allowed set.
func TestAllowedVAlignCoversFill(t *testing.T) {
	var found bool
	for _, v := range contracts.AllowedVAlign() {
		if v == contracts.VAlignFill {
			found = true
		}
	}
	if !found {
		t.Fatal("AllowedVAlign() must include VAlignFill")
	}
}
