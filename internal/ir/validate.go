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
// descendants).
func ValidateSlide(s contracts.Slide) error {
	return childErr("nodes", s.Nodes)
}

// ValidateNode runs Stage-1 structural validation on a single node, mirroring
// pptx-go's scene.ValidateScene per-node rules, recursing into containers.
// Optional enum fields left empty are accepted (they default at render time);
// the substantive structural rules are enforced. Nodes are the pointer forms
// the codec produces.
func ValidateNode(n contracts.SlideNode) error {
	switch v := n.(type) {
	case *contracts.Heading:
		if v.Level < 1 || v.Level > 6 {
			return fmt.Errorf("heading: level %d out of range 1..6", v.Level)
		}
	case *contracts.List:
		var errs []error
		if !validListKind(v.Kind) {
			errs = append(errs, fmt.Errorf("list: invalid listKind %q", v.Kind))
		}
		if len(v.Items) == 0 {
			errs = append(errs, errors.New("list: needs at least one item"))
		}
		return errors.Join(errs...)
	case *contracts.Callout:
		if !validCalloutKind(v.Kind) {
			return fmt.Errorf("callout: invalid calloutKind %q", v.Kind)
		}
	case *contracts.Image:
		var errs []error
		if v.AssetID == "" {
			errs = append(errs, errors.New("image: empty assetId"))
		}
		errs = append(errs, cropErrs(v.Crop)...)
		return errors.Join(errs...)
	case *contracts.Chart:
		if v.AssetID == "" {
			return errors.New("chart: empty assetId")
		}
	case *contracts.CodeBlock:
		if v.AssetID == "" {
			return errors.New("code_block: empty assetId")
		}
	case *contracts.Flow:
		if len(v.Steps) == 0 {
			return errors.New("flow: needs at least one step")
		}
	case *contracts.Table:
		return validateTable(v)
	case *contracts.TwoColumn:
		return validateTwoColumn(v)
	case *contracts.Grid:
		return validateGrid(v)
	case *contracts.Card:
		return childErr("card.body", v.Body)
	case *contracts.CardSection:
		if len(v.Body) == 0 {
			return errors.New("card_section: body must be non-empty")
		}
		return childErr("card_section.body", v.Body)
	case *contracts.Decoration:
		return validateDecoration(v)
	}
	// Nodes with no structural constraints (hero, prose, quote, chip, arrow,
	// divider, section_divider) are always valid.
	return nil
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
	return errors.Join(errs...)
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
	default:
		errs = append(errs, fmt.Errorf("decoration: invalid decorationKind %q (want preset or asset)", d.Kind))
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

// validListKind accepts the known list kinds plus empty (defaults to bullet).
func validListKind(k contracts.ListKind) bool {
	switch k {
	case "", contracts.ListBullet, contracts.ListNumber, contracts.ListChecklist:
		return true
	}
	return false
}

// validCalloutKind accepts the known callout kinds plus empty (defaults to note).
func validCalloutKind(k contracts.CalloutKind) bool {
	switch k {
	case "", contracts.CalloutNote, contracts.CalloutWarning, contracts.CalloutTip, contracts.CalloutImportant:
		return true
	}
	return false
}
