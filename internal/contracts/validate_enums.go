package contracts

import (
	"errors"
	"fmt"
)

// ValidateNodeEnums validates every enum-typed field on n (including inline
// RichText fields). It is called by the Stage-1 validator in internal/ir
// after structural checks. Returns nil when all values are within their
// closed sets; returns errors.Join of all violations otherwise.
//
// Optional fields (omitempty) accept the empty string (acceptEmpty=true).
// DecorationKind is NOT optional — an empty kind is a hard error.
func ValidateNodeEnums(n SlideNode) error {
	var errs []error
	switch v := n.(type) {
	case *Hero:
		errs = append(errs,
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
	case *Heading:
		errs = append(errs,
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
		errs = append(errs, validateRichTextEnums("text", v.Text))
	case *Prose:
		errs = append(errs,
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
		for i, p := range v.Paragraphs {
			errs = append(errs, validateRichTextEnums(fmt.Sprintf("paragraphs[%d]", i), p))
		}
	case *Quote:
		errs = append(errs,
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
		errs = append(errs, validateRichTextEnums("text", v.Text))
	case *List:
		errs = append(errs,
			checkEnum("listKind", "ListKind", v.Kind, AllowedListKind(), true),
		)
		for i, item := range v.Items {
			errs = append(errs, validateRichTextEnums(fmt.Sprintf("items[%d].text", i), item.Text))
		}
	case *Callout:
		errs = append(errs,
			checkEnum("calloutKind", "CalloutKind", v.Kind, AllowedCalloutKind(), true),
		)
		errs = append(errs, validateRichTextEnums("body", v.Body))
	case *Chip:
		errs = append(errs,
			checkEnum("tone", "ChipTone", v.Tone, AllowedChipTone(), true),
			checkEnum("color", "ColorRole", v.Color, AllowedColorRole(), true),
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
	case *Arrow:
		errs = append(errs,
			checkEnum("direction", "ArrowDirection", v.Direction, AllowedArrowDirection(), true),
		)
	case *SectionDivider:
		errs = append(errs,
			checkEnum("align", "HAlign", v.Align, AllowedHAlign(), true),
		)
	case *Divider:
		errs = append(errs,
			checkEnum("spacing", "SpaceRole", v.Spacing, AllowedSpaceRole(), true),
		)
	case *Grid:
		errs = append(errs,
			checkEnum("gap", "SpaceRole", v.Gap, AllowedSpaceRole(), true),
		)
	case *TwoColumn:
		errs = append(errs,
			checkEnum("ratio", "ColumnRatio", v.Ratio, AllowedColumnRatio(), true),
			checkEnum("join", "ColumnJoin", v.Join, AllowedColumnJoin(), true),
		)
	case *Flow:
		errs = append(errs,
			checkEnum("orientation", "FlowOrientation", v.Orientation, AllowedFlowOrientation(), true),
			checkEnum("connector", "ConnectorKind", v.Connector, AllowedConnectorKind(), true),
		)
		for i, step := range v.Steps {
			errs = append(errs,
				validateRichTextEnums(fmt.Sprintf("steps[%d].label", i), step.Label),
				validateRichTextEnums(fmt.Sprintf("steps[%d].detail", i), step.Detail),
			)
		}
	case *Card:
		errs = append(errs,
			checkEnum("bodyLayout", "BodyLayout", v.BodyLayout, AllowedBodyLayout(), true),
			checkEnum("fill", "ColorRole", v.Fill, AllowedColorRole(), true),
			checkEnum("borderStyle", "BorderStyle", v.BorderStyle, AllowedBorderStyle(), true),
			checkEnum("size", "CardSize", v.Size, AllowedCardSize(), true),
			checkEnum("layout", "CardLayout", v.Layout, AllowedCardLayout(), true),
			checkEnum("elevation", "ElevationRole", v.Elevation, AllowedElevationRole(), true),
		)
	case *Decoration:
		// DecorationKind is NOT optional: acceptEmpty=false.
		errs = append(errs,
			checkEnum("decorationKind", "DecorationKind", v.Kind, AllowedDecorationKind(), false),
			checkEnum("layer", "Layer", v.Layer, AllowedLayer(), true),
			checkEnum("anchor", "Anchor", v.Anchor, AllowedAnchor(), true),
			checkEnum("color", "ColorRole", v.Color, AllowedColorRole(), true),
		)
	case *Image:
		errs = append(errs,
			checkEnum("frame", "FrameKind", v.Frame, AllowedFrameKind(), true),
			checkEnum("fit", "Fit", v.Fit, AllowedFit(), true),
			checkEnum("cornerRadius", "RadiusRole", v.CornerRadius, AllowedRadiusRole(), true),
			checkEnum("elevation", "ElevationRole", v.Elevation, AllowedElevationRole(), true),
		)
	case *Table:
		for i, h := range v.Headers {
			errs = append(errs, validateRichTextEnums(fmt.Sprintf("headers[%d]", i), h))
		}
		for i, row := range v.Rows {
			for j, cell := range row {
				errs = append(errs, validateRichTextEnums(fmt.Sprintf("rows[%d][%d]", i, j), cell))
			}
		}
	case *Stat:
		errs = append(errs,
			checkEnum("deltaTone", "DeltaTone", v.DeltaTone, AllowedDeltaTone(), true),
		)
	}
	return errors.Join(errs...)
}

// ValidateSlideEnums validates the enum-typed fields on a Slide (layout,
// alignment, variant, background). Called from ir.ValidateSlide so that
// LayoutKind, VAlign, HAlign, Variant, BackgroundKind, and ColorRole are all
// checked at Stage-1 rather than silently falling through to the renderer.
func ValidateSlideEnums(s Slide) error {
	var errs []error
	errs = append(errs,
		checkEnum("layout", "LayoutKind", s.Layout, AllowedLayoutKind(), true),
		checkEnum("align.vertical", "VAlign", s.Align.Vertical, AllowedVAlign(), true),
		checkEnum("align.horizontal", "HAlign", s.Align.Horizontal, AllowedHAlign(), true),
		checkEnum("variant", "Variant", s.Variant, AllowedVariant(), true),
		checkEnum("archetype", "SlideArchetype", s.Archetype, AllowedSlideArchetype(), true),
	)
	if s.Background != nil {
		errs = append(errs,
			checkEnum("background.kind", "BackgroundKind", s.Background.Kind, AllowedBackgroundKind(), true),
			checkEnum("background.color", "ColorRole", s.Background.Color, AllowedColorRole(), true),
		)
		for i, gr := range s.Background.Gradient {
			errs = append(errs,
				checkEnum(
					fmt.Sprintf("background.gradient[%d]", i),
					"ColorRole", gr, AllowedColorRole(), true,
				),
			)
		}
		for i, st := range s.Background.Stops {
			errs = append(errs, checkEnum(fmt.Sprintf("background.stops[%d].color", i), "ColorRole", st.Color, AllowedColorRole(), true))
		}
		for i, mg := range s.Background.Mesh {
			errs = append(errs,
				checkEnum(fmt.Sprintf("background.mesh[%d].anchor", i), "Anchor", mg.Anchor, AllowedAnchor(), true),
				checkEnum(fmt.Sprintf("background.mesh[%d].color", i), "ColorRole", mg.Color, AllowedColorRole(), true))
		}
		if s.Background.Scrim != nil {
			errs = append(errs,
				checkEnum("background.scrim.color", "ColorRole", s.Background.Scrim.Color, AllowedColorRole(), true))
		}
		if s.Background.Duotone != nil {
			errs = append(errs,
				checkEnum("background.duotone.shadow", "ColorRole", s.Background.Duotone.Shadow, AllowedColorRole(), true),
				checkEnum("background.duotone.highlight", "ColorRole", s.Background.Duotone.Highlight, AllowedColorRole(), true))
		}
	}
	return errors.Join(errs...)
}

// validateRichTextEnums checks the TypeRole and TextColorRole enums on every
// run in rt. The field prefix is included in each sub-error path so the caller
// can see exactly which run is invalid.
func validateRichTextEnums(field string, rt RichText) error {
	var errs []error
	for i, run := range rt {
		if err := checkEnum(
			fmt.Sprintf("%s[%d].typeRole", field, i),
			"TypeRole", run.TypeRole, AllowedTypeRole(), true,
		); err != nil {
			errs = append(errs, err)
		}
		if err := checkEnum(
			fmt.Sprintf("%s[%d].color.token", field, i),
			"TextColorRole", run.Color.Token, AllowedTextColorRole(), true,
		); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
