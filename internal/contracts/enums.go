// Package contracts owns the contract enums for every named-string type
// whose wire value is constrained to a closed set. Each type lives in
// its node file (e.g. CalloutKind in ir_nodes_callout.go) but the
// validation helpers — IsValid, Allowed, EnumError — live here so a
// decoder, validator, or rendering pass can ALL consult the same
// closed-set source of truth. References to the wire values are kept in
// sync with the doc-comments on each typed string.
package contracts

import "strings"

// EnumError is the typed error returned when a wire value does not name a
// known constant of a named enum. It names the offending field, the bad
// value, and the closed set of allowed values — so a wrong enum reaches
// the model as a loud, correctable error message (Phase 12 A4).
type EnumError struct {
	// Field is the dotted path to the offending field on its node
	// (e.g. "calloutKind", "steps[0].connector", "align.horizontal").
	Field string
	// Got is the received wire value.
	Got string
	// Allowed is the closed set of wire values the type accepts; kept
	// in the order declared so the message is stable + diff-friendly.
	Allowed []string
	// Type is the named enum type (e.g. "CalloutKind"), included in
	// the message so the error is self-describing.
	Type string
}

// Error implements the error interface. Format:
//
//	<field>: <type> "<got>" is not valid; want one of a|b|c
func (e *EnumError) Error() string {
	return enumMsg(e.Field, e.Type, e.Got, e.Allowed)
}

// enumMsg is the message body used by Error() and by the surrogate errors
// that do not flow through EnumError. Kept as a package function so a
// validator that builds a plain error (e.g. embedding the message into a
// larger errors.Join) can stay byte-identical to a typed EnumError.
func enumMsg(field, typeName, got string, allowed []string) string {
	return fmtEnumErr(field, typeName, got, allowed)
}

// stringsFrom converts a typed-slice of named string enums to a string
// slice — used at the single boundary that hands wire values to a
// human-facing message (formatEnumErr / EnumError.Allowed). Kept tiny and
// inlinable; no per-enum conversion helpers.
func stringsFrom[T ~string](vs []T) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = string(v)
	}
	return out
}

// AllowedTypeRole returns the closed set of TypeRole wire values, in
// canonical (declaration) order. The set is the single source of truth for
// the render mapper's default branch and the validator's IsValid check.
func AllowedTypeRole() []TypeRole {
	return []TypeRole{
		TypeDisplay, TypeH1, TypeH2, TypeH3, TypeH4, TypeH5,
		TypeBody, TypeBodySmall, TypeCaption, TypeMono, TypeCode,
	}
}

// AllowedString returns the closed set of TypeRole wire values as plain
// strings — for inclusion in an error message.
func (TypeRole) allowedStrings() []string { return stringsFrom(AllowedTypeRole()) }

// AllowedTextColorRole returns the closed set of TextColorRole wire values.
func AllowedTextColorRole() []TextColorRole {
	return []TextColorRole{
		TextPrimary, TextSecondary, TextTertiary, TextInverse, TextMuted,
		TextAccent, TextAccentAlt, TextSuccess, TextWarning, TextError,
	}
}

// allowedStrings returns the closed set of TextColorRole wire values as
// plain strings — for inclusion in an error message.
func (TextColorRole) allowedStrings() []string { return stringsFrom(AllowedTextColorRole()) }

// AllowedColorRole returns the closed set of ColorRole wire values.
func AllowedColorRole() []ColorRole {
	return []ColorRole{
		ColorCanvas, ColorSurface, ColorSurfaceAlt, ColorAccent,
		ColorAccentAlt, ColorAccentWarm, ColorSuccess, ColorWarning,
		ColorError, ColorInfo, ColorPaper,
	}
}

// allowedStrings returns the closed set of ColorRole wire values as plain
// strings — for inclusion in an error message.
func (ColorRole) allowedStrings() []string { return stringsFrom(AllowedColorRole()) }

// AllowedSpaceRole returns the closed set of SpaceRole wire values.
func AllowedSpaceRole() []SpaceRole {
	return []SpaceRole{
		SpaceXS, SpaceSM, SpaceMD, SpaceLG, SpaceXL, Space2XL,
	}
}

// allowedStrings returns the closed set of SpaceRole wire values as plain
// strings — for inclusion in an error message.
func (SpaceRole) allowedStrings() []string { return stringsFrom(AllowedSpaceRole()) }

// AllowedElevationRole returns the closed set of ElevationRole wire values.
func AllowedElevationRole() []ElevationRole {
	return []ElevationRole{
		ElevationFlat, ElevationRaised, ElevationElevated,
	}
}

// allowedStrings returns the closed set of ElevationRole wire values as
// plain strings — for inclusion in an error message.
func (ElevationRole) allowedStrings() []string { return stringsFrom(AllowedElevationRole()) }

// AllowedRadiusRole returns the closed set of RadiusRole wire values. The
// zero value (RadiusNone, "") is also legal — a rectangular picture corner.
func AllowedRadiusRole() []RadiusRole {
	return []RadiusRole{
		RadiusNone, RadiusSM, RadiusMD, RadiusLG, RadiusFull,
	}
}

// allowedStrings returns the closed set of RadiusRole wire values as plain
// strings — for inclusion in an error message.
func (RadiusRole) allowedStrings() []string { return stringsFrom(AllowedRadiusRole()) }

// AllowedLayoutKind returns the closed set of LayoutKind wire values.
func AllowedLayoutKind() []LayoutKind {
	return []LayoutKind{
		LayoutCover, LayoutTitleContent, LayoutTwoColumn,
		LayoutCardGrid, LayoutFullBleed, LayoutBlank,
	}
}

// allowedStrings returns the closed set of LayoutKind wire values as plain
// strings — for inclusion in an error message.
func (LayoutKind) allowedStrings() []string { return stringsFrom(AllowedLayoutKind()) }

// AllowedListKind returns the closed set of ListKind wire values.
func AllowedListKind() []ListKind {
	return []ListKind{
		ListBullet, ListNumber, ListChecklist,
	}
}

// allowedStrings returns the closed set of ListKind wire values as plain
// strings — for inclusion in an error message.
func (ListKind) allowedStrings() []string { return stringsFrom(AllowedListKind()) }

// AllowedCalloutKind returns the closed set of CalloutKind wire values.
func AllowedCalloutKind() []CalloutKind {
	return []CalloutKind{
		CalloutNote, CalloutWarning, CalloutTip, CalloutImportant,
	}
}

// allowedStrings returns the closed set of CalloutKind wire values as plain
// strings — for inclusion in an error message.
func (CalloutKind) allowedStrings() []string { return stringsFrom(AllowedCalloutKind()) }

// AllowedChipTone returns the closed set of ChipTone wire values.
func AllowedChipTone() []ChipTone {
	return []ChipTone{
		ChipTint, ChipSolid, ChipOutline,
	}
}

// allowedStrings returns the closed set of ChipTone wire values as plain
// strings — for inclusion in an error message.
func (ChipTone) allowedStrings() []string { return stringsFrom(AllowedChipTone()) }

// AllowedArrowDirection returns the closed set of ArrowDirection wire values.
func AllowedArrowDirection() []ArrowDirection {
	return []ArrowDirection{
		ArrowRight, ArrowLeft, ArrowUp, ArrowDown,
	}
}

// allowedStrings returns the closed set of ArrowDirection wire values as
// plain strings — for inclusion in an error message.
func (ArrowDirection) allowedStrings() []string { return stringsFrom(AllowedArrowDirection()) }

// AllowedColumnRatio returns the closed set of ColumnRatio wire values.
func AllowedColumnRatio() []ColumnRatio {
	return []ColumnRatio{
		Ratio11, Ratio12, Ratio21,
	}
}

// allowedStrings returns the closed set of ColumnRatio wire values as plain
// strings — for inclusion in an error message.
func (ColumnRatio) allowedStrings() []string { return stringsFrom(AllowedColumnRatio()) }

// AllowedColumnJoin returns the closed set of ColumnJoin wire values (D-055).
// The empty string (JoinNone) is excluded from this list but is accepted by
// the validator (acceptEmpty=true).
func AllowedColumnJoin() []ColumnJoin {
	return []ColumnJoin{JoinBadge, JoinArrow}
}

// allowedStrings returns the closed set of ColumnJoin wire values as plain
// strings — for inclusion in an error message.
func (ColumnJoin) allowedStrings() []string { return stringsFrom(AllowedColumnJoin()) }

// AllowedBodyLayout returns the closed set of BodyLayout wire values.
func AllowedBodyLayout() []BodyLayout {
	return []BodyLayout{
		BodyVertical, BodyHorizontal,
	}
}

// allowedStrings returns the closed set of BodyLayout wire values as plain
// strings — for inclusion in an error message.
func (BodyLayout) allowedStrings() []string { return stringsFrom(AllowedBodyLayout()) }

// AllowedBorderStyle returns the closed set of BorderStyle wire values.
func AllowedBorderStyle() []BorderStyle {
	return []BorderStyle{
		BorderDefault, BorderNone, BorderSolid, BorderAccent,
	}
}

// allowedStrings returns the closed set of BorderStyle wire values as plain
// strings — for inclusion in an error message.
func (BorderStyle) allowedStrings() []string { return stringsFrom(AllowedBorderStyle()) }

// AllowedCardSize returns the closed set of CardSize wire values.
func AllowedCardSize() []CardSize {
	return []CardSize{
		CardSizeMD, CardSizeSM, CardSizeLG,
	}
}

// allowedStrings returns the closed set of CardSize wire values as plain
// strings — for inclusion in an error message.
func (CardSize) allowedStrings() []string { return stringsFrom(AllowedCardSize()) }

// AllowedCardLayout returns the closed set of CardLayout wire values.
func AllowedCardLayout() []CardLayout {
	return []CardLayout{
		CardLayoutDefault, CardLayoutIconTop,
	}
}

// allowedStrings returns the closed set of CardLayout wire values as plain
// strings — for inclusion in an error message.
func (CardLayout) allowedStrings() []string { return stringsFrom(AllowedCardLayout()) }

// AllowedFlowOrientation returns the closed set of FlowOrientation values.
func AllowedFlowOrientation() []FlowOrientation {
	return []FlowOrientation{
		FlowHorizontal, FlowVertical,
	}
}

// allowedStrings returns the closed set of FlowOrientation wire values as
// plain strings — for inclusion in an error message.
func (FlowOrientation) allowedStrings() []string {
	return stringsFrom(AllowedFlowOrientation())
}

// AllowedConnectorKind returns the closed set of ConnectorKind values.
func AllowedConnectorKind() []ConnectorKind {
	return []ConnectorKind{
		ConnectorArrow, ConnectorArrowDashed, ConnectorCycle, ConnectorPlus,
	}
}

// allowedStrings returns the closed set of ConnectorKind wire values as
// plain strings — for inclusion in an error message.
func (ConnectorKind) allowedStrings() []string { return stringsFrom(AllowedConnectorKind()) }

// AllowedDecorationKind returns the closed set of DecorationKind values.
func AllowedDecorationKind() []DecorationKind {
	return []DecorationKind{
		DecorationPreset, DecorationAsset, DecorationText,
	}
}

// allowedStrings returns the closed set of DecorationKind wire values as
// plain strings — for inclusion in an error message.
func (DecorationKind) allowedStrings() []string {
	return stringsFrom(AllowedDecorationKind())
}

// AllowedDataMarkKind returns the closed set of DataMarkKind wire values
// (R14.8, D-122). ENUM DRIFT GUARD: every DataMarkKind const above must be
// listed here — a missing value silently rejects a valid wire value.
func AllowedDataMarkKind() []DataMarkKind {
	return []DataMarkKind{
		DataMarkBar, DataMarkBars, DataMarkSparkline, DataMarkDonut, DataMarkGauge,
	}
}

// allowedStrings returns the closed set of DataMarkKind wire values as plain
// strings — for inclusion in an error message.
func (DataMarkKind) allowedStrings() []string { return stringsFrom(AllowedDataMarkKind()) }

// AllowedButtonTone returns the closed set of ButtonTone wire values
// (R12.1, D-094). ENUM DRIFT GUARD: every ButtonTone const above must be
// listed here — a missing value silently rejects a valid wire value. The
// empty string is also legal (ButtonPrimary, the zero value) and is accepted
// by the validator (acceptEmpty=true), mirroring the existing ChipTone
// pattern.
func AllowedButtonTone() []ButtonTone {
	return []ButtonTone{ButtonPrimary, ButtonAccentAlt, ButtonGhost, ButtonNeutral}
}

// allowedStrings returns the closed set of ButtonTone wire values as plain
// strings — for inclusion in an error message.
func (ButtonTone) allowedStrings() []string { return stringsFrom(AllowedButtonTone()) }

// AllowedButtonSize returns the closed set of ButtonSize wire values
// (R12.1, D-094). The empty string is also legal (ButtonSizeMD, the zero
// value) and is accepted by the validator (acceptEmpty=true).
func AllowedButtonSize() []ButtonSize {
	return []ButtonSize{ButtonSizeMD, ButtonSizeSM, ButtonSizeLG}
}

// allowedStrings returns the closed set of ButtonSize wire values as plain
// strings — for inclusion in an error message.
func (ButtonSize) allowedStrings() []string { return stringsFrom(AllowedButtonSize()) }

// AllowedCheckState returns the closed set of CheckState wire values
// (R12.2, D-095). The empty string is also legal (CheckDone, the zero
// value) and is accepted by the validator (acceptEmpty=true).
func AllowedCheckState() []CheckState {
	return []CheckState{CheckDone, CheckNo, CheckNeutral}
}

// allowedStrings returns the closed set of CheckState wire values as plain
// strings — for inclusion in an error message.
func (CheckState) allowedStrings() []string { return stringsFrom(AllowedCheckState()) }

// AllowedRowTone returns the closed set of RowTone wire values (R12.7,
// D-100). The empty string is also legal (RowPlain, the zero value) and is
// accepted by the validator (acceptEmpty=true).
func AllowedRowTone() []RowTone {
	return []RowTone{RowPlain, RowPill}
}

// allowedStrings returns the closed set of RowTone wire values as plain
// strings — for inclusion in an error message.
func (RowTone) allowedStrings() []string { return stringsFrom(AllowedRowTone()) }

// AllowedLogoToneKind returns the closed set of LogoToneKind wire values
// (R14.7, D-125). The empty string (LogoToneNone) is excluded from this list
// but is accepted by the validator (acceptEmpty=true), mirroring the
// AllowedColumnJoin pattern.
func AllowedLogoToneKind() []LogoToneKind {
	return []LogoToneKind{LogoToneNone, LogoToneMono, LogoToneBrand}
}

// allowedStrings returns the closed set of LogoToneKind wire values as plain
// strings — for inclusion in an error message.
func (LogoToneKind) allowedStrings() []string { return stringsFrom(AllowedLogoToneKind()) }

// AllowedLayer returns the closed set of Layer wire values.
func AllowedLayer() []Layer {
	return []Layer{
		LayerBackground, LayerForeground,
	}
}

// allowedStrings returns the closed set of Layer wire values as plain
// strings — for inclusion in an error message.
func (Layer) allowedStrings() []string { return stringsFrom(AllowedLayer()) }

// AllowedFrameKind returns the closed set of FrameKind wire values.
func AllowedFrameKind() []FrameKind {
	return []FrameKind{
		FrameNone, FrameBrowser, FramePhone, FrameDesktop, FrameLaptop,
	}
}

// allowedStrings returns the closed set of FrameKind wire values as plain
// strings — for inclusion in an error message.
func (FrameKind) allowedStrings() []string { return stringsFrom(AllowedFrameKind()) }

// AllowedFit returns the closed set of Fit wire values.
func AllowedFit() []Fit { return []Fit{FitFill, FitNone} }

// allowedStrings returns the closed set of Fit wire values as plain
// strings — for inclusion in an error message.
func (Fit) allowedStrings() []string { return stringsFrom(AllowedFit()) }

// AllowedAnchor returns the closed set of Anchor wire values.
func AllowedAnchor() []Anchor {
	return []Anchor{
		AnchorTopLeft, AnchorTop, AnchorTopRight,
		AnchorLeft, AnchorCenter, AnchorRight,
		AnchorBottomLeft, AnchorBottom, AnchorBottomRight,
	}
}

// allowedStrings returns the closed set of Anchor wire values as plain
// strings — for inclusion in an error message.
func (Anchor) allowedStrings() []string { return stringsFrom(AllowedAnchor()) }

// AllowedVAlign returns the closed set of VAlign wire values. Note: the
// empty string is also legal (meaning VAlignTop, the default) — the
// validator accepts both; this list names only the named wire values.
func AllowedVAlign() []VAlign {
	return []VAlign{
		VAlignTop, VAlignCenter, VAlignBottom, VAlignJustify, VAlignFill,
	}
}

// allowedStrings returns the closed set of VAlign wire values as plain
// strings — for inclusion in an error message.
func (VAlign) allowedStrings() []string { return stringsFrom(AllowedVAlign()) }

// AllowedHAlign returns the closed set of HAlign wire values. Empty string
// is also legal (HAlignLeft).
func AllowedHAlign() []HAlign {
	return []HAlign{
		HAlignLeft, HAlignCenter, HAlignRight,
	}
}

// allowedStrings returns the closed set of HAlign wire values as plain
// strings — for inclusion in an error message.
func (HAlign) allowedStrings() []string { return stringsFrom(AllowedHAlign()) }

// AllowedDeltaTone returns the closed set of DeltaTone wire values (D-057).
func AllowedDeltaTone() []DeltaTone {
	return []DeltaTone{DeltaNeutral, DeltaUp, DeltaDown}
}

// allowedStrings returns the closed set of DeltaTone wire values as plain
// strings — for inclusion in an error message.
func (DeltaTone) allowedStrings() []string { return stringsFrom(AllowedDeltaTone()) }

// AllowedVariant returns the closed set of Variant wire values. Empty
// string means VariantLight (default) and is also legal.
func AllowedVariant() []Variant { return []Variant{VariantLight, VariantDark} }

// allowedStrings returns the closed set of Variant wire values as plain
// strings — for inclusion in an error message.
func (Variant) allowedStrings() []string { return stringsFrom(AllowedVariant()) }

// AllowedBackgroundKind returns the closed set of BackgroundKind wire
// values. The zero value (empty string) is BackgroundNone and is legal.
func AllowedBackgroundKind() []BackgroundKind {
	return []BackgroundKind{
		BackgroundNone, BackgroundColor, BackgroundGradient, BackgroundAsset,
		BackgroundRadial, BackgroundMesh,
	}
}

// allowedStrings returns the closed set of BackgroundKind wire values as
// plain strings — for inclusion in an error message.
func (BackgroundKind) allowedStrings() []string { return stringsFrom(AllowedBackgroundKind()) }

// IsValidEnum reports whether v is a known constant of a closed enum.
// Compares against the typed-slice allowed set returned by AllowedX();
// uses a generic constraint so callers can stay typed end-to-end (an
// unknown string is a HARD error at Stage-1 validation rather than a
// silent fall-through to the render mapper's default branch).
func IsValidEnum[T ~string, S ~[]T](v T, allowed S) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// checkEnum builds a typed EnumError if v is not in the closed set. Used
// by the Stage-1 validator: returns nil on a valid (including empty-for-
// optional) value, or returns a fully-formed error listing the offending
// field + allowed values. typeName is the Go type name (e.g.
// "CalloutKind"), used in the message. T must implement allowedStrings()
// so every per-enum method is exercised through the error path and the
// allowed-set message is sourced from the type itself.
func checkEnum[T interface {
	~string
	allowedStrings() []string
}, S ~[]T](field, typeName string, v T, allowed S, acceptEmpty bool) error {
	if string(v) == "" {
		if acceptEmpty {
			return nil
		}
	}
	if IsValidEnum(v, allowed) {
		return nil
	}
	return &EnumError{
		Field:   field,
		Type:    typeName,
		Got:     string(v),
		Allowed: v.allowedStrings(),
	}
}

// formatEnumErr formats a one-line enum-failure message shared by every
// named enum and by the validator's plain-error path. The format is
// stable so snapshot / golden tests can match without drift:
//
//	<field>: <type> "<got>" is not valid; want one of a|b|c
//
// When field is empty, the leading "<field>:" segment is dropped.
func formatEnumErr(field, typeName, got string, allowed []string) string {
	pipe := strings.Join(allowed, "|")
	body := typeName + " \"" + got + "\" is not valid; want one of " + pipe
	if field == "" {
		return body
	}
	return field + ": " + body
}

// fmtEnumErr is a tiny alias kept so callers can use the shorter name
// without an import dance.
func fmtEnumErr(field, typeName, got string, allowed []string) string {
	return formatEnumErr(field, typeName, got, allowed)
}

// _ keeps strings in used-imports — guard against accidental trim.
var _ = strings.Join
