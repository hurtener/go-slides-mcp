package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// tableStyleSample is a fully-populated comparison-matrix Style exercising
// every field, including a grouped header row whose spans sum to the column
// count (R14.3, D-118).
func tableStyleSample() *contracts.TableStyle {
	return &contracts.TableStyle{
		HeaderFill:   true,
		Zebra:        true,
		HighlightCol: 4,
		RowLabelCol:  true,
		HeaderGroups: []contracts.HeaderGroup{
			{Label: "Plan", Span: 1},
			{Label: "Paid tiers", Span: 3},
		},
	}
}

// TestMapTableStyleNilByteIdentical asserts mapTableStyle(nil) returns nil,
// so a Table with no Style maps to a scene.Table with a nil Style — the
// plain banded table stays byte-identical (R14.3, D-118).
func TestMapTableStyleNilByteIdentical(t *testing.T) {
	t.Parallel()

	if got := mapTableStyle(nil); got != nil {
		t.Errorf("mapTableStyle(nil) = %+v, want nil", got)
	}

	node := &contracts.Table{Headers: []contracts.RichText{rt("h")}}
	sn := mapNode(node)
	table, ok := sn.(scene.Table)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Table", sn)
	}
	if table.Style != nil {
		t.Errorf("Table with no contracts Style: mapped scene.Table.Style = %+v, want nil", table.Style)
	}
}

// TestMapNodeTableStyleAllFields asserts every TableStyle field (including
// the HeaderGroups slice) maps 1:1 into scene.TableStyle (R14.3, D-118).
func TestMapNodeTableStyleAllFields(t *testing.T) {
	t.Parallel()

	node := &contracts.Table{
		Headers: []contracts.RichText{rt("Feature"), rt("Free"), rt("Pro"), rt("Enterprise")},
		Rows:    [][]contracts.RichText{{rt("Seats"), rt("1"), rt("10"), rt("Unlimited")}},
		Caption: "Plans",
		Style:   tableStyleSample(),
	}
	sn := mapNode(node)
	table, ok := sn.(scene.Table)
	if !ok {
		t.Fatalf("mapNode returned %T, want scene.Table", sn)
	}
	if table.Style == nil {
		t.Fatal("Style: got nil, want non-nil")
	}
	want := scene.TableStyle{
		HeaderFill:   true,
		Zebra:        true,
		HighlightCol: 4,
		RowLabelCol:  true,
		HeaderGroups: []scene.HeaderGroup{
			{Label: "Plan", Span: 1},
			{Label: "Paid tiers", Span: 3},
		},
	}
	if table.Style.HeaderFill != want.HeaderFill || table.Style.Zebra != want.Zebra ||
		table.Style.HighlightCol != want.HighlightCol || table.Style.RowLabelCol != want.RowLabelCol {
		t.Errorf("Style scalar fields: got %+v, want %+v", *table.Style, want)
	}
	if len(table.Style.HeaderGroups) != len(want.HeaderGroups) {
		t.Fatalf("HeaderGroups length: got %d, want %d", len(table.Style.HeaderGroups), len(want.HeaderGroups))
	}
	for i, g := range table.Style.HeaderGroups {
		if g != want.HeaderGroups[i] {
			t.Errorf("HeaderGroups[%d]: got %+v, want %+v", i, g, want.HeaderGroups[i])
		}
	}
}

// matrixTable is a features x plans comparison table (headers + rows) shared
// by the plain and styled render-level fixtures below.
func matrixTable(style *contracts.TableStyle) *contracts.Table {
	return &contracts.Table{
		Headers: []contracts.RichText{rt("Feature"), rt("Free"), rt("Pro"), rt("Enterprise")},
		Rows: [][]contracts.RichText{
			{rt("Seats"), rt("1"), rt("10"), rt("Unlimited")},
			{rt("SSO"), rt("No"), rt("Yes"), rt("Yes")},
			{rt("SLA"), rt("No"), rt("99.9%"), rt("99.99%")},
		},
		Style: style,
	}
}

func matrixDoc(style *contracts.TableStyle) contracts.SlideDoc {
	return contracts.SlideDoc{
		Title: "Comparison Matrix",
		Slides: []contracts.Slide{{
			ID:     "matrix",
			Layout: contracts.LayoutTitleContent,
			Nodes:  []contracts.SlideNode{matrixTable(style)},
		}},
	}
}

// TestRenderTableStyleEmitsMoreThanPlain is the R14.3 product-level accept
// case: a styled comparison matrix (header band + zebra + highlight column +
// row-label column + grouped headers) renders without error and its slide
// XML carries strictly more fill markup than the same table left plain — the
// header band, zebra stripes, highlight tint, and grouped-header row are all
// additional <a:solidFill> elements the engine emits from soul tokens.
func TestRenderTableStyleEmitsMoreThanPlain(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	plainBuf, _, err := Render(matrixDoc(nil), s)
	if err != nil {
		t.Fatalf("Render(plain) error = %v", err)
	}
	styledBuf, _, err := Render(matrixDoc(tableStyleSample()), s)
	if err != nil {
		t.Fatalf("Render(styled) error = %v", err)
	}

	plainXML := string(firstSlideXML(t, plainBuf))
	styledXML := string(firstSlideXML(t, styledBuf))

	if !strings.Contains(styledXML, "<a:tbl>") {
		t.Fatalf("styled table did not emit a native table:\n%s", styledXML)
	}
	plainFills := strings.Count(plainXML, "<a:solidFill>")
	styledFills := strings.Count(styledXML, "<a:solidFill>")
	if styledFills <= plainFills {
		t.Errorf("styled table fill count = %d, want > plain table fill count %d", styledFills, plainFills)
	}
	// The grouped header row adds the group label as extra text content.
	if !strings.Contains(styledXML, "<a:t>Paid tiers</a:t>") {
		t.Errorf("styled table missing grouped header label:\n%s", styledXML)
	}
}

// TestRenderTableNilStyleByteIdentical asserts a Table whose Style field is
// left unset renders byte-identical to a Table with an explicit nil Style —
// the mapping introduces no drift for the pre-R14.3 plain-table shape.
func TestRenderTableNilStyleByteIdentical(t *testing.T) {
	t.Parallel()

	s := soul.DeckardWhite()

	implicit, _, err := Render(matrixDoc(nil), s)
	if err != nil {
		t.Fatalf("Render(implicit nil) error = %v", err)
	}
	var explicitNilStyle *contracts.TableStyle
	explicit, _, err := Render(matrixDoc(explicitNilStyle), s)
	if err != nil {
		t.Fatalf("Render(explicit nil) error = %v", err)
	}
	if !bytes.Equal(implicit, explicit) {
		t.Fatal("Table with nil Style is not byte-identical to Table with no Style set")
	}
}

// TestRenderTableStyleDeterministic guards that a styled comparison matrix
// renders byte-identically across repeated renders (the render-determinism
// hard contract, CLAUDE.md §5), mirroring TestRenderDeterministicAcrossRepeatedRenders.
func TestRenderTableStyleDeterministic(t *testing.T) {
	t.Parallel()

	doc := matrixDoc(tableStyleSample())
	s := soul.DeckardWhite()

	first, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, _, err := Render(doc, s)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("styled table Render() bytes differ across identical renders")
	}
}
