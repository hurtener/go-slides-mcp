package ir

import (
	"errors"
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// ValidateDoc validates every slide in a deck, joining all violations so the
// caller sees every problem at once (like the engine's scene.ValidateScene).
func ValidateDoc(d contracts.SlideDoc) error {
	var errs []error
	for i, s := range d.Slides {
		if err := ValidateSlide(s); err != nil {
			errs = append(errs, fmt.Errorf("slide[%d] %q: %w", i, s.ID, err))
		}
	}
	return errors.Join(errs...)
}

// ValidateSlide validates a slide's top-level nodes (and, via recursion, their
// descendants) plus the slide's own enum-typed fields (layout, alignment,
// variant, background kind/color).
func ValidateSlide(s contracts.Slide) error {
	return errors.Join(
		contracts.ValidateSlideEnums(s),
		errors.Join(validateBackground(s.Background)...),
		childErr("nodes", s.Nodes),
	)
}

// validateBackground checks the structural constraints of a background's
// multi-stop gradient, mesh glows (R13.2/R13.3/R13.4), and scrim overlay
// (R14.1), mirroring the engine's degrade-to-warning rules but enforced as a
// hard Stage-1 error (precedent: validateTableStyle). A nil background, or
// one with empty Stops/Mesh/Scrim, produces no error — the byte-identical
// opt-out.
func validateBackground(b *contracts.Background) []error {
	if b == nil {
		return nil
	}
	var errs []error
	if n := len(b.Stops); n > 0 {
		if n < 2 || n > 8 {
			errs = append(errs, fmt.Errorf("background: stops need 2..8 entries, got %d", n))
		}
		prev := -1.0
		for i, st := range b.Stops {
			if st.Pos < 0 || st.Pos > 1 {
				errs = append(errs, fmt.Errorf("background: stops[%d] pos %v out of [0,1]", i, st.Pos))
			}
			if st.Pos <= prev {
				errs = append(errs, fmt.Errorf("background: stops not strictly ascending at [%d]", i))
			}
			prev = st.Pos
		}
	}
	for i, mg := range b.Mesh {
		if mg.Radius < 0 {
			errs = append(errs, fmt.Errorf("background: mesh[%d] radius %v out of range, must be >= 0", i, mg.Radius))
		}
		if mg.Alpha < 0 || mg.Alpha > 1 {
			errs = append(errs, fmt.Errorf("background: mesh[%d] alpha %v out of [0,1]", i, mg.Alpha))
		}
	}
	if b.Scrim != nil {
		if b.Scrim.Opacity < 0 || b.Scrim.Opacity > 1 {
			errs = append(errs, fmt.Errorf("background: scrim opacity %v out of [0,1]", b.Scrim.Opacity))
		}
	}
	return errs
}

// ValidateNode runs Stage-1 structural validation on a single node, mirroring
// pptx-go's scene.ValidateScene per-node rules, recursing into containers.
// After structural checks, enum validation is applied via ValidateNodeEnums:
// optional enum fields left empty are accepted (they default at render time);
// unknown wire values produce an error naming the field and the allowed set.
func ValidateNode(n contracts.SlideNode) error {
	var errs []error
	switch v := n.(type) {
	case *contracts.Heading:
		if v.Level < 1 || v.Level > 6 {
			errs = append(errs, fmt.Errorf("heading: level %d out of range 1..6", v.Level))
		}
	case *contracts.List:
		if len(v.Items) == 0 {
			errs = append(errs, errors.New("list: needs at least one item"))
		}
	case *contracts.Image:
		if v.AssetID == "" {
			errs = append(errs, errors.New("image: empty assetId"))
		}
		errs = append(errs, cropErrs(v.Crop)...)
		errs = append(errs, imageAnnotationErrs(v.Annotations)...)
	case *contracts.Chart:
		if v.AssetID == "" {
			errs = append(errs, errors.New("chart: empty assetId"))
		}
	case *contracts.CodeBlock:
		if v.AssetID == "" {
			errs = append(errs, errors.New("code_block: empty assetId"))
		}
	case *contracts.Flow:
		if len(v.Steps) == 0 {
			errs = append(errs, errors.New("flow: needs at least one step"))
		}
	case *contracts.Table:
		errs = append(errs, validateTable(v))
	case *contracts.TwoColumn:
		errs = append(errs, validateTwoColumn(v))
	case *contracts.Grid:
		errs = append(errs, validateGrid(v))
	case *contracts.Card:
		errs = append(errs, childErr("card.body", v.Body))
		errs = append(errs, validateCard(v)...)
	case *contracts.CardSection:
		if len(v.Body) == 0 {
			errs = append(errs, errors.New("card_section: body must be non-empty"))
		}
		errs = append(errs, childErr("card_section.body", v.Body))
	case *contracts.Decoration:
		errs = append(errs, validateDecoration(v))
	case *contracts.Timeline:
		errs = append(errs, validateTimeline(v)...)
	case *contracts.DataMark:
		errs = append(errs, validateDataMark(v)...)
	case *contracts.Quadrant:
		errs = append(errs, validateQuadrant(v)...)
	case *contracts.Tree:
		errs = append(errs, validateTree(v)...)
	case *contracts.Funnel:
		if len(v.Stages) == 0 {
			errs = append(errs, errors.New("funnel: needs at least one stage"))
		}
	case *contracts.Cycle:
		if len(v.Stages) == 0 {
			errs = append(errs, errors.New("cycle: needs at least one stage"))
		}
	case *contracts.LogoWall:
		errs = append(errs, validateLogoWall(v))
	case *contracts.Button:
		errs = append(errs, validateButton(v)...)
	case *contracts.ChipRow:
		errs = append(errs, validateChipRow(v)...)
	case *contracts.Checklist:
		errs = append(errs, validateChecklist(v)...)
	case *contracts.Banner:
		errs = append(errs, validateBanner(v)...)
	case *contracts.IconRows:
		errs = append(errs, validateIconRows(v)...)
	case *contracts.Lockup:
		errs = append(errs, validateLockup(v)...)
	}
	// Enum validation applies to every node type; optional empty fields pass.
	errs = append(errs, contracts.ValidateNodeEnums(n))
	return errors.Join(errs...)
}

// childErr validates each child node under label, joining violations.
func childErr(label string, nodes []contracts.SlideNode) error {
	var errs []error
	for i, n := range nodes {
		if err := ValidateNode(n); err != nil {
			errs = append(errs, fmt.Errorf("%s[%d]: %w", label, i, err))
		}
	}
	return errors.Join(errs...)
}

func validateTable(t *contracts.Table) error {
	if len(t.Headers) == 0 {
		return errors.New("table: needs at least one header column")
	}
	w := len(t.Headers)
	var errs []error
	for i, row := range t.Rows {
		if len(row) != w {
			errs = append(errs, fmt.Errorf("table: row[%d] width %d != header width %d", i, len(row), w))
		}
	}
	errs = append(errs, validateTableStyle(t.Style, w)...)
	return errors.Join(errs...)
}

// validateTableStyle checks the structural constraints of a comparison-matrix
// Style against the table's column count w (R14.3, D-118): HighlightCol must
// be a real column (or 0 = none); every HeaderGroup must span at least one
// column; and, when groups are present, their spans must sum to the column
// count so the merged header row lines up with the body.
func validateTableStyle(s *contracts.TableStyle, w int) []error {
	if s == nil {
		return nil
	}
	var errs []error
	if s.HighlightCol < 0 || s.HighlightCol > w {
		errs = append(errs, fmt.Errorf("table: style.highlightCol %d out of range 0..%d", s.HighlightCol, w))
	}
	if len(s.HeaderGroups) > 0 {
		sum := 0
		for i, g := range s.HeaderGroups {
			if g.Span < 1 {
				errs = append(errs, fmt.Errorf("table: style.headerGroups[%d] span %d must be >= 1", i, g.Span))
			}
			sum += g.Span
		}
		if sum != w {
			errs = append(errs, fmt.Errorf("table: style.headerGroups span sum %d != header width %d", sum, w))
		}
	}
	return errs
}

func validateTwoColumn(tc *contracts.TwoColumn) error {
	var errs []error
	if len(tc.Left) == 0 {
		errs = append(errs, errors.New("two_column: left must be non-empty"))
	}
	if len(tc.Right) == 0 {
		errs = append(errs, errors.New("two_column: right must be non-empty"))
	}
	if err := childErr("two_column.left", tc.Left); err != nil {
		errs = append(errs, err)
	}
	if err := childErr("two_column.right", tc.Right); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func validateGrid(g *contracts.Grid) error {
	var errs []error
	if g.Columns < 2 || g.Columns > 4 {
		errs = append(errs, fmt.Errorf("grid: columns %d out of range 2..4", g.Columns))
	}
	if len(g.Ratio) != 0 && len(g.Ratio) != g.Columns {
		errs = append(errs, fmt.Errorf("grid: ratio length %d must be 0 or == columns %d", len(g.Ratio), g.Columns))
	}
	if len(g.Cells) == 0 {
		errs = append(errs, errors.New("grid: needs at least one cell"))
	} else if g.Columns >= 2 && g.Columns <= 4 && len(g.Cells)%g.Columns != 0 {
		errs = append(errs, fmt.Errorf("grid: cell count %d not a multiple of columns %d", len(g.Cells), g.Columns))
	}
	for i, c := range g.Connectors {
		if c.Between[0] < 0 || c.Between[1] < 0 {
			errs = append(errs, fmt.Errorf("grid: connectors[%d] between indices must be >= 0", i))
			continue
		}
		if c.Between[1] != c.Between[0]+1 {
			errs = append(errs, fmt.Errorf("grid: connectors[%d] between %v must name adjacent columns {c,c+1}", i, c.Between))
		}
		if g.Columns > 0 && c.Between[1] >= g.Columns {
			errs = append(errs, fmt.Errorf("grid: connectors[%d] between %v out of range for columns %d", i, c.Between, g.Columns))
		}
	}
	if err := childErr("grid.cells", g.Cells); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// validateCard adds Stage-1 structural checks for the additive Ribbon field
// (R12.3, D-098). A corner-star ribbon ignores Text; every other RibbonPos
// requires non-empty Text so a highlighted card surfaces a real label rather
// than an invisible badge.
func validateCard(c *contracts.Card) []error {
	if c.Ribbon == nil {
		return nil
	}
	if c.Ribbon.Position != contracts.RibbonCornerStar && c.Ribbon.Text == "" {
		return []error{errors.New("card: ribbon text must be non-empty unless position == corner_star")}
	}
	return nil
}

// validateDecoration checks structural constraints for Decoration nodes.
// The enum check for DecorationKind itself is handled by ValidateNodeEnums.
func validateDecoration(d *contracts.Decoration) error {
	var errs []error
	switch d.Kind {
	case contracts.DecorationPreset:
		if d.Preset == "" {
			errs = append(errs, errors.New("decoration: preset kind needs a preset name"))
		}
	case contracts.DecorationAsset:
		if d.AssetID == "" {
			errs = append(errs, errors.New("decoration: asset kind needs an assetId"))
		}
	case contracts.DecorationText:
		if d.Text == "" {
			errs = append(errs, errors.New("decoration: text kind needs non-empty text"))
		}
	}
	if d.Opacity < 0 || d.Opacity > 1 {
		errs = append(errs, fmt.Errorf("decoration: opacity %.3f out of [0,1]", d.Opacity))
	}
	return errors.Join(errs...)
}

// validateTimeline checks structural constraints for the Timeline node
// (R14.4, D-119), mirroring the engine's scene.ValidateScene rules for
// Timeline: at least one milestone (across top-level Milestones and every
// Lane), every Milestone.Position in [0,1], and every Band's [From,To] span
// within [0,1] with From <= To.
func validateTimeline(t *contracts.Timeline) []error {
	var errs []error
	total := len(t.Milestones)
	for _, ln := range t.Lanes {
		total += len(ln.Milestones)
	}
	if total == 0 {
		errs = append(errs, errors.New("timeline: needs at least one milestone or lane"))
	}
	checkMilestones := func(where string, ms []contracts.Milestone) {
		for i, m := range ms {
			if m.Position < 0 || m.Position > 1 {
				errs = append(errs, fmt.Errorf("timeline: %s milestone[%d] position %g out of [0,1]", where, i, m.Position))
			}
		}
	}
	checkMilestones("top-level", t.Milestones)
	for li, ln := range t.Lanes {
		checkMilestones(fmt.Sprintf("lane[%d]", li), ln.Milestones)
	}
	for i, b := range t.Bands {
		if b.From < 0 || b.From > 1 || b.To < 0 || b.To > 1 || b.From > b.To {
			errs = append(errs, fmt.Errorf("timeline: band[%d] span [%g,%g] invalid (need 0<=from<=to<=1)", i, b.From, b.To))
		}
	}
	return errs
}

// validateDataMark applies the DataMark node's structural Stage-1 rules
// (R14.8, D-122), mirroring the engine's scene.ValidateScene rules: an
// empty Kind defaults to "bar" (the engine's zero value). Bar/donut/gauge
// use Value, which must be in [0,1]; bars/sparkline use Values, which must
// have at least one entry, each in [0,1].
func validateDataMark(d *contracts.DataMark) []error {
	var errs []error
	switch d.Kind {
	case contracts.DataMarkBars, contracts.DataMarkSparkline:
		if len(d.Values) == 0 {
			errs = append(errs, errors.New("data_mark: bars/sparkline requires at least one value"))
		}
		for i, val := range d.Values {
			if val < 0 || val > 1 {
				errs = append(errs, fmt.Errorf("data_mark: values[%d] (%g) out of [0,1]", i, val))
			}
		}
	default: // "" (default), DataMarkBar, DataMarkDonut, DataMarkGauge
		if d.Value < 0 || d.Value > 1 {
			errs = append(errs, fmt.Errorf("data_mark: value %g out of [0,1]", d.Value))
		}
	}
	return errs
}

// validateQuadrant checks structural constraints for the Quadrant node
// (R14.9, D-124), mirroring the engine's scene.ValidateScene rules: at
// least one plotted item, and every Item's X/Y in [0,1].
func validateQuadrant(q *contracts.Quadrant) []error {
	var errs []error
	if len(q.Items) == 0 {
		errs = append(errs, errors.New("quadrant: needs at least one item"))
	}
	for i, it := range q.Items {
		if it.X < 0 || it.X > 1 {
			errs = append(errs, fmt.Errorf("quadrant: items[%d] x %g out of [0,1]", i, it.X))
		}
		if it.Y < 0 || it.Y > 1 {
			errs = append(errs, fmt.Errorf("quadrant: items[%d] y %g out of [0,1]", i, it.Y))
		}
	}
	return errs
}

// validateTree checks structural constraints for the Tree node (R14.10,
// D-127), mirroring the engine's scene.ValidateScene rules: the root needs
// a label. Depth/breadth clamping past the safe area is the engine's
// render-time warn, not a Stage-1 hard error.
func validateTree(t *contracts.Tree) []error {
	var errs []error
	if t.Root.Label == "" {
		errs = append(errs, errors.New("tree: root needs a label"))
	}
	return errs
}

// validateLogoWall enforces LogoWall's structural rules (R14.7, D-125): at
// least one logo, and every logo needs an AssetID (mirroring the engine's
// warn-don't-fail behavior only for a resolvable-but-missing asset — an
// empty AssetID is a Stage-1 hard error, not a render-time warning).
func validateLogoWall(l *contracts.LogoWall) error {
	var errs []error
	if len(l.Logos) == 0 {
		errs = append(errs, errors.New("logo_wall: needs at least one logo"))
	}
	for i, logo := range l.Logos {
		if logo.AssetID == "" {
			errs = append(errs, fmt.Errorf("logo_wall: logo[%d] empty assetId", i))
		}
	}
	return errors.Join(errs...)
}

// validateButton enforces Button's structural rule (R12.1, D-094): a button
// needs a non-empty Label (it is shape-only — no label means no affordance).
// Tone/Size/Align enum checks run separately via ValidateNodeEnums.
func validateButton(b *contracts.Button) []error {
	if b.Label == "" {
		return []error{errors.New("button: empty label")}
	}
	return nil
}

// validateChipRow enforces ChipRow's structural rule (R12.5, D-096): at least
// one chip (an empty strip renders nothing and is not a real "tag strip").
// Per-chip Tone/Color enum checks run separately via ValidateNodeEnums.
func validateChipRow(c *contracts.ChipRow) []error {
	if len(c.Chips) == 0 {
		return []error{errors.New("chip_row: needs at least one chip")}
	}
	return nil
}

// validateChecklist enforces Checklist's structural rules (R12.2, D-095): at
// least one item, and Columns in [1..3] (the engine clamps past 3 to 3 at
// render — we lift it to a Stage-1 hard error so an agent gets a loud
// correction rather than a silent reflow). The engine allows Columns==0 to
// mean "1 column" but rejects negative values, so we accept the zero value.
func validateChecklist(c *contracts.Checklist) []error {
	if len(c.Items) == 0 {
		return []error{errors.New("checklist: needs at least one item")}
	}
	if c.Columns < 0 || c.Columns > 3 {
		return []error{fmt.Errorf("checklist: columns %d out of range 0..3", c.Columns)}
	}
	return nil
}

// validateBanner enforces Banner's structural rules (R12.6, D-097): at
// least one of Lead or Body (an invisible strip carries no message) and
// every Trailing child itself validates. Recurses into Trailing so the
// Stage-1 cascade catches a malformed trailing node.
func validateBanner(b *contracts.Banner) []error {
	var errs []error
	if len(b.Lead) == 0 && len(b.Body) == 0 {
		errs = append(errs, errors.New("banner: needs at least one of lead or body"))
	}
	for i, tw := range b.Trailing {
		if err := ValidateNode(tw); err != nil {
			errs = append(errs, fmt.Errorf("banner.trailing[%d]: %w", i, err))
		}
	}
	return errs
}

// validateIconRows enforces IconRows' structural rule (R12.7, D-100): at
// least one row (an empty list renders nothing and is not a real
// "icon-label" strip).
func validateIconRows(ir *contracts.IconRows) []error {
	if len(ir.Rows) == 0 {
		return []error{errors.New("icon_rows: needs at least one row")}
	}
	return nil
}

// validateLockup enforces Lockup's structural rules (R12.9, D-102): exactly
// one of AssetID or Icon is set (asset path OR media-free icon path), and
// MaxHeight is non-negative. Enum checks for AssetSide/Align run separately
// via ValidateNodeEnums.
func validateLockup(l *contracts.Lockup) []error {
	var errs []error
	hasAsset := l.AssetID != ""
	hasIcon := l.Icon != ""
	if hasAsset == hasIcon {
		errs = append(errs, errors.New("lockup: exactly one of assetId or icon must be set"))
	}
	if l.MaxHeight < 0 {
		errs = append(errs, fmt.Errorf("lockup: maxHeight %v must be >= 0", l.MaxHeight))
	}
	return errs
}

func cropErrs(c contracts.Crop) []error {
	var errs []error
	for _, e := range []struct {
		name string
		v    float64
	}{{"left", c.Left}, {"top", c.Top}, {"right", c.Right}, {"bottom", c.Bottom}} {
		if e.v < 0 || e.v > 1 {
			errs = append(errs, fmt.Errorf("image: crop %s %.3f out of [0,1]", e.name, e.v))
		}
	}
	if c.Left+c.Right >= 1 {
		errs = append(errs, errors.New("image: crop left+right must be < 1"))
	}
	if c.Top+c.Bottom >= 1 {
		errs = append(errs, errors.New("image: crop top+bottom must be < 1"))
	}
	return errs
}

// imageAnnotationErrs validates an Image's optional R14.17 annotation
// overlay: every pin/highlight coordinate must be a fraction in [0,1] of the
// image box. A nil Annotations is valid (no annotations to check).
func imageAnnotationErrs(a *contracts.ImageAnnotations) []error {
	if a == nil {
		return nil
	}
	var errs []error
	frac := func(v float64) bool { return v >= 0 && v <= 1 }
	for i, p := range a.Pins {
		if !frac(p.X) || !frac(p.Y) {
			errs = append(errs, fmt.Errorf("image: annotation pin[%d] x/y out of [0,1]", i))
		}
	}
	for i, h := range a.Highlights {
		if !frac(h.X) || !frac(h.Y) || !frac(h.W) || !frac(h.H) {
			errs = append(errs, fmt.Errorf("image: annotation highlight[%d] x/y/w/h out of [0,1]", i))
		}
	}
	return errs
}
