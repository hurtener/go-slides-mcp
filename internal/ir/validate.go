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
	case *contracts.CardSection:
		if len(v.Body) == 0 {
			errs = append(errs, errors.New("card_section: body must be non-empty"))
		}
		errs = append(errs, childErr("card_section.body", v.Body))
	case *contracts.Decoration:
		errs = append(errs, validateDecoration(v))
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
	if err := childErr("grid.cells", g.Cells); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
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
