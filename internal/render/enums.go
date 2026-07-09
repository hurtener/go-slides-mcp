package render

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func mapLayoutKind(kind contracts.LayoutKind) scene.LayoutKind {
	switch kind {
	case contracts.LayoutCover:
		return scene.LayoutCover
	case contracts.LayoutTwoColumn:
		return scene.LayoutTwoColumn
	case contracts.LayoutCardGrid:
		return scene.LayoutCardGrid
	case contracts.LayoutFullBleed:
		return scene.LayoutFullBleed
	case contracts.LayoutBlank:
		return scene.LayoutBlank
	case contracts.LayoutTitleContent:
		fallthrough
	default:
		return scene.LayoutTitleContent
	}
}

func mapListKind(kind contracts.ListKind) scene.ListKind {
	switch kind {
	case contracts.ListNumber:
		return scene.ListNumber
	case contracts.ListChecklist:
		return scene.ListChecklist
	case contracts.ListBullet:
		fallthrough
	default:
		return scene.ListBullet
	}
}

func mapCalloutKind(kind contracts.CalloutKind) scene.CalloutKind {
	switch kind {
	case contracts.CalloutWarning:
		return scene.CalloutWarning
	case contracts.CalloutTip:
		return scene.CalloutTip
	case contracts.CalloutImportant:
		return scene.CalloutImportant
	case contracts.CalloutNote:
		fallthrough
	default:
		return scene.CalloutNote
	}
}

func mapChipTone(tone contracts.ChipTone) scene.ChipTone {
	switch tone {
	case contracts.ChipSolid:
		return scene.ChipSolid
	case contracts.ChipOutline:
		return scene.ChipOutline
	case contracts.ChipTint:
		fallthrough
	default:
		return scene.ChipTint
	}
}

func mapArrowDirection(direction contracts.ArrowDirection) scene.ArrowDirection {
	switch direction {
	case contracts.ArrowLeft:
		return scene.ArrowLeft
	case contracts.ArrowUp:
		return scene.ArrowUp
	case contracts.ArrowDown:
		return scene.ArrowDown
	case contracts.ArrowRight:
		fallthrough
	default:
		return scene.ArrowRight
	}
}

func mapColumnRatio(ratio contracts.ColumnRatio) scene.ColumnRatio {
	switch ratio {
	case contracts.Ratio12:
		return scene.Ratio12
	case contracts.Ratio21:
		return scene.Ratio21
	case contracts.Ratio11:
		fallthrough
	default:
		return scene.Ratio11
	}
}

func mapColumnJoin(join contracts.ColumnJoin) scene.ColumnJoin {
	switch join {
	case contracts.JoinBadge:
		return scene.JoinBadge
	case contracts.JoinArrow:
		return scene.JoinArrow
	default: // JoinNone ("") and any unknown value
		return scene.JoinNone
	}
}

func mapBodyLayout(layout contracts.BodyLayout) scene.BodyLayout {
	if layout == contracts.BodyHorizontal {
		return scene.BodyHorizontal
	}
	return scene.BodyVertical
}

func mapBorderStyle(style contracts.BorderStyle) scene.BorderStyle {
	switch style {
	case contracts.BorderNone:
		return scene.BorderNone
	case contracts.BorderSolid:
		return scene.BorderSolid
	case contracts.BorderAccent:
		return scene.BorderAccent
	case contracts.BorderDefault:
		fallthrough
	default:
		return scene.BorderDefault
	}
}

func mapCardSize(size contracts.CardSize) scene.CardSize {
	switch size {
	case contracts.CardSizeSM:
		return scene.CardSizeSM
	case contracts.CardSizeLG:
		return scene.CardSizeLG
	case contracts.CardSizeMD:
		fallthrough
	default:
		return scene.CardSizeMD
	}
}

func mapCardLayout(layout contracts.CardLayout) scene.CardLayout {
	if layout == contracts.CardLayoutIconTop {
		return scene.CardLayoutIconTop
	}
	return scene.CardLayoutDefault
}

// mapButtonTone maps the product ButtonTone string enum to the engine's int
// ButtonTone enum (R12.1, D-094). The empty string ("" — unset) maps to the
// zero value ButtonPrimary, mirroring the engine default (a solid accent pill).
func mapButtonTone(t contracts.ButtonTone) scene.ButtonTone {
	switch t {
	case contracts.ButtonAccentAlt:
		return scene.ButtonAccentAlt
	case contracts.ButtonGhost:
		return scene.ButtonGhost
	case contracts.ButtonNeutral:
		return scene.ButtonNeutral
	case contracts.ButtonPrimary:
		fallthrough
	default:
		return scene.ButtonPrimary
	}
}

// mapButtonSize maps the product ButtonSize string enum to the engine's int
// ButtonSize enum (R12.1, D-094). The empty string ("" — unset) maps to the
// zero value ButtonMD (the default geometry).
func mapButtonSize(s contracts.ButtonSize) scene.ButtonSize {
	switch s {
	case contracts.ButtonSizeSM:
		return scene.ButtonSM
	case contracts.ButtonSizeLG:
		return scene.ButtonLG
	case contracts.ButtonSizeMD:
		fallthrough
	default:
		return scene.ButtonMD
	}
}

// mapCheckState maps the product CheckState string enum to the engine's
// int CheckState enum (R12.2, D-095). The empty string ("" — unset) maps
// to the zero value CheckDone, mirroring the engine default (a filled
// accent-tinted check glyph).
func mapCheckState(s contracts.CheckState) scene.CheckState {
	switch s {
	case contracts.CheckNo:
		return scene.CheckNo
	case contracts.CheckNeutral:
		return scene.CheckNeutral
	case contracts.CheckDone:
		fallthrough
	default:
		return scene.CheckDone
	}
}

// mapRowTone maps the product RowTone string enum to the engine's int
// RowTone enum (R12.7, D-100). The empty string ("" — unset) maps to the
// zero value RowPlain, mirroring the engine default (no frame).
func mapRowTone(t contracts.RowTone) scene.RowTone {
	switch t {
	case contracts.RowPill:
		return scene.RowPill
	case contracts.RowPlain:
		fallthrough
	default:
		return scene.RowPlain
	}
}

// mapBannerFill maps a product ColorRole to the engine's ColorRole for
// the Banner.Fill field (R12.6, D-097). The product empty string maps to
// the engine's zero value ColorCanvas — NOT the generic ColorSurface
// default — because the Banner renderer interprets engine-ColorCanvas as
// "promote to ColorAccent" (a banner is always a filled strip and a
// canvas-colored one would be invisible). A non-empty value delegates to
// the generic mapper.
func mapBannerFill(r contracts.ColorRole) pptx.ColorRole {
	if r == "" {
		return scene.ColorCanvas
	}
	return mapColorRole(r)
}

// mapIconRowsGlyphColor maps a product ColorRole to the engine's
// ColorRole for the IconRows.GlyphColor field (R12.7, D-100). Same pattern
// as mapBannerFill: an empty product value preserves the engine-ColorCanvas
// zero so the renderer promotes it to ColorAccent (a canvas-colored glyph
// would be invisible against any slide background).
func mapIconRowsGlyphColor(r contracts.ColorRole) pptx.ColorRole {
	if r == "" {
		return scene.ColorCanvas
	}
	return mapColorRole(r)
}

// mapAssetSide maps the product AssetSide string enum to the engine's
// int enum (R12.9, D-102). The empty string ("" — unset) maps to the
// zero value LeadCaption, mirroring the engine default (caption leads).
func mapAssetSide(s contracts.AssetSide) scene.AssetSide {
	switch s {
	case contracts.TrailCaption:
		return scene.TrailCaption
	case contracts.LeadCaption:
		fallthrough
	default:
		return scene.LeadCaption
	}
}

func mapColorRole(role contracts.ColorRole) pptx.ColorRole {
	switch role {
	case contracts.ColorCanvas:
		return scene.ColorCanvas
	case contracts.ColorSurfaceAlt:
		return scene.ColorSurfaceAlt
	case contracts.ColorAccent:
		return scene.ColorAccent
	case contracts.ColorAccentAlt:
		return scene.ColorAccentAlt
	case contracts.ColorAccentWarm:
		return scene.ColorAccentWarm
	case contracts.ColorSuccess:
		return scene.ColorSuccess
	case contracts.ColorWarning:
		return scene.ColorWarning
	case contracts.ColorError:
		return scene.ColorError
	case contracts.ColorInfo:
		return scene.ColorInfo
	case contracts.ColorPaper:
		return pptx.ColorPaper
	case contracts.ColorSurface:
		fallthrough
	default:
		return scene.ColorSurface
	}
}

func mapFlowOrientation(orientation contracts.FlowOrientation) scene.FlowOrientation {
	if orientation == contracts.FlowVertical {
		return scene.FlowVertical
	}
	return scene.FlowHorizontal
}

// mapDataMarkKind maps the product DataMarkKind string enum to the engine's
// int enum (R14.8, D-122). The empty string ("" — unset) maps to the
// engine's zero value DataMarkBar, mirroring the engine default.
func mapDataMarkKind(kind contracts.DataMarkKind) scene.DataMarkKind {
	switch kind {
	case contracts.DataMarkBars:
		return scene.DataMarkBars
	case contracts.DataMarkSparkline:
		return scene.DataMarkSparkline
	case contracts.DataMarkDonut:
		return scene.DataMarkDonut
	case contracts.DataMarkGauge:
		return scene.DataMarkGauge
	case contracts.DataMarkBar:
		fallthrough
	default:
		return scene.DataMarkBar
	}
}

// mapLogoToneKind maps the product LogoToneKind string enum to the engine's
// int enum (R14.7, D-125). The empty string ("" — unset) maps to the
// engine's zero value LogoToneNone, mirroring the engine default.
func mapLogoToneKind(tone contracts.LogoToneKind) scene.LogoToneKind {
	switch tone {
	case contracts.LogoToneMono:
		return scene.LogoToneMono
	case contracts.LogoToneBrand:
		return scene.LogoToneBrand
	case contracts.LogoToneNone:
		fallthrough
	default:
		return scene.LogoToneNone
	}
}

func mapConnectorKind(kind contracts.ConnectorKind) scene.ConnectorKind {
	switch kind {
	case contracts.ConnectorArrowDashed:
		return scene.ConnectorArrowDashed
	case contracts.ConnectorCycle:
		return scene.ConnectorCycle
	case contracts.ConnectorPlus:
		return scene.ConnectorPlus
	case contracts.ConnectorBiArrow:
		return scene.ConnectorBiArrow
	case contracts.ConnectorArrow:
		fallthrough
	default:
		return scene.ConnectorArrow
	}
}

// mapRibbonPos maps the product RibbonPos string enum to the engine's int
// RibbonPos enum (R12.3, D-098). The empty string maps to RibbonTopBar, the
// engine zero value.
func mapRibbonPos(p contracts.RibbonPos) scene.RibbonPos {
	switch p {
	case contracts.RibbonCornerTL:
		return scene.RibbonCornerTL
	case contracts.RibbonCornerTR:
		return scene.RibbonCornerTR
	case contracts.RibbonCornerStar:
		return scene.RibbonCornerStar
	case contracts.RibbonTopBar:
		fallthrough
	default:
		return scene.RibbonTopBar
	}
}

// mapJoinPosition maps the product JoinPosition string enum to the engine's
// int JoinPosition enum (R12.8, D-101). The empty string maps to JoinSeam,
// the engine zero value.
func mapJoinPosition(p contracts.JoinPosition) scene.JoinPosition {
	switch p {
	case contracts.JoinTopBridge:
		return scene.JoinTopBridge
	case contracts.JoinBottomBridge:
		return scene.JoinBottomBridge
	case contracts.JoinSeam:
		fallthrough
	default:
		return scene.JoinSeam
	}
}

func mapSpaceRole(role contracts.SpaceRole) pptx.SpaceRole {
	switch role {
	case contracts.SpaceXS:
		return scene.SpaceXS
	case contracts.SpaceSM:
		return scene.SpaceSM
	case contracts.SpaceMD:
		return scene.SpaceMD
	case contracts.SpaceLG:
		return scene.SpaceLG
	case contracts.SpaceXL:
		return scene.SpaceXL
	case contracts.Space2XL:
		return scene.Space2XL
	default:
		return scene.SpaceMD
	}
}

func mapElevationRole(role contracts.ElevationRole) pptx.ElevationRole {
	switch role {
	case contracts.ElevationRaised:
		return scene.ElevationRaised
	case contracts.ElevationElevated:
		return scene.ElevationElevated
	case contracts.ElevationFlat:
		fallthrough
	default:
		return scene.ElevationFlat
	}
}

// mapRadiusRole converts the wire-level RadiusRole string to the engine's
// pptx.RadiusRole enum (R13.11). The empty string (RadiusNone) maps to the
// zero value, leaving a picture rectangular — byte-identical.
func mapRadiusRole(role contracts.RadiusRole) pptx.RadiusRole {
	switch role {
	case contracts.RadiusSM:
		return pptx.RadiusSM
	case contracts.RadiusMD:
		return pptx.RadiusMD
	case contracts.RadiusLG:
		return pptx.RadiusLG
	case contracts.RadiusFull:
		return pptx.RadiusFull
	case contracts.RadiusNone:
		fallthrough
	default:
		return pptx.RadiusNone
	}
}

func mapTextColorRole(role contracts.TextColorRole) pptx.TextColorRole {
	switch role {
	case contracts.TextSecondary:
		return scene.TextSecondary
	case contracts.TextTertiary:
		return scene.TextTertiary
	case contracts.TextInverse:
		return scene.TextInverse
	case contracts.TextMuted:
		return scene.TextMuted
	case contracts.TextAccent:
		return scene.TextAccent
	case contracts.TextAccentAlt:
		return scene.TextAccentAlt
	case contracts.TextSuccess:
		return scene.TextSuccess
	case contracts.TextWarning:
		return scene.TextWarning
	case contracts.TextError:
		return scene.TextError
	case contracts.TextPrimary:
		fallthrough
	default:
		return scene.TextPrimary
	}
}

func mapTypeRole(role contracts.TypeRole) pptx.TypeRole {
	switch role {
	case contracts.TypeDisplay:
		return scene.TypeDisplay
	case contracts.TypeH1:
		return scene.TypeH1
	case contracts.TypeH2:
		return scene.TypeH2
	case contracts.TypeH3:
		return scene.TypeH3
	case contracts.TypeH4:
		return scene.TypeH4
	case contracts.TypeH5:
		return scene.TypeH5
	case contracts.TypeBodySmall:
		return scene.TypeBodySmall
	case contracts.TypeCaption:
		return scene.TypeCaption
	case contracts.TypeMono:
		return scene.TypeMono
	case contracts.TypeCode:
		return scene.TypeCode
	case contracts.TypeBody:
		fallthrough
	default:
		return scene.TypeBody
	}
}

// mapFrameKind converts the wire-level FrameKind string to the scene enum.
func mapFrameKind(kind contracts.FrameKind) scene.FrameKind {
	switch kind {
	case contracts.FrameBrowser:
		return scene.FrameBrowser
	case contracts.FramePhone:
		return scene.FramePhone
	case contracts.FrameDesktop:
		return scene.FrameDesktop
	case contracts.FrameLaptop:
		return scene.FrameLaptop
	case contracts.FrameNone:
		fallthrough
	default:
		return scene.FrameNone
	}
}

// mapFit converts the wire-level Fit string to the scene/pptx Fit enum.
func mapFit(fit contracts.Fit) scene.Fit {
	if fit == contracts.FitNone {
		return scene.FitNone
	}
	return scene.FitFill
}

// mapCrop copies a Crop's fields. The scene Crop type is a re-export of
// pptx.Crop and uses the same field set.
func mapCrop(c contracts.Crop) scene.Crop {
	return scene.Crop{Left: c.Left, Top: c.Top, Right: c.Right, Bottom: c.Bottom}
}

// mapImageAnnotations converts a product ImageAnnotations to the engine's
// scene.ImageAnnotations 1:1 (R14.17). nil in -> nil out, so an Image with no
// Annotations stays byte-identical to a pre-R14.17 Image.
func mapImageAnnotations(a *contracts.ImageAnnotations) *scene.ImageAnnotations {
	if a == nil {
		return nil
	}
	out := &scene.ImageAnnotations{}
	for _, p := range a.Pins {
		out.Pins = append(out.Pins, scene.ImagePin{
			X:           p.X,
			Y:           p.Y,
			Label:       p.Label,
			Caption:     p.Caption,
			AccentIndex: p.AccentIndex,
		})
	}
	for _, h := range a.Highlights {
		out.Highlights = append(out.Highlights, scene.ImageHighlight{
			X:           h.X,
			Y:           h.Y,
			W:           h.W,
			H:           h.H,
			AccentIndex: h.AccentIndex,
		})
	}
	return out
}

// mapDecorationKind converts the wire-level DecorationKind string to the scene
// enum (preset, asset, or text — R13.9).
func mapDecorationKind(kind contracts.DecorationKind) scene.DecorationKind {
	switch kind {
	case contracts.DecorationAsset:
		return scene.DecorationAsset
	case contracts.DecorationText:
		return scene.DecorationText
	default:
		return scene.DecorationPreset
	}
}

// mapDecoration converts a product Decoration to the engine's scene.Decoration
// 1:1. Extracted so both the *contracts.Decoration node case in scene.go and
// mapDecorationPtr (Card.Backdrop, R13.10) share one mapping — no logic
// change, byte-identical to the previously-inline version.
func mapDecoration(n contracts.Decoration) scene.Decoration {
	return scene.Decoration{
		Kind:     mapDecorationKind(n.Kind),
		Preset:   n.Preset,
		AssetID:  scene.AssetID(n.AssetID),
		Layer:    mapLayer(n.Layer),
		Anchor:   mapAnchor(n.Anchor),
		Offset:   mapPosition(n.Offset),
		Size:     mapSize(n.Size),
		Bleed:    n.Bleed,
		Opacity:  n.Opacity,
		Rotation: n.Rotation,
		Color:    mapColorRolePtr(n.Color),
		Pitch:    pptx.Pt(n.Pitch),
		Text:     n.Text,
		FontSize: n.FontSize,
	}
}

// mapDecorationPtr maps an optional product Decoration to the engine,
// nil-safe (R13.10, Card.Backdrop). nil stays nil — byte-identical to a card
// with no backdrop.
func mapDecorationPtr(d *contracts.Decoration) *scene.Decoration {
	if d == nil {
		return nil
	}
	m := mapDecoration(*d)
	return &m
}

// mapLayer converts the wire-level Layer string to the scene enum
// (background vs foreground).
func mapLayer(layer contracts.Layer) scene.Layer {
	if layer == contracts.LayerForeground {
		return scene.LayerForeground
	}
	return scene.LayerBackground
}

// mapAnchor translates the wire 9-point compass string to the scene/pptx
// Anchor enum (the scene enum spells its middle row as "…CenterLeft" /
// "…Center" / "…CenterRight", while the contracts spell it as "left" / "_" /
// "right"). Unknown values fall back to AnchorTopLeft.
func mapAnchor(anchor contracts.Anchor) scene.Anchor {
	switch anchor {
	case contracts.AnchorTopLeft:
		return scene.AnchorTopLeft
	case contracts.AnchorTop:
		return scene.AnchorTopCenter
	case contracts.AnchorTopRight:
		return scene.AnchorTopRight
	case contracts.AnchorLeft:
		return scene.AnchorCenterLeft
	case contracts.AnchorCenter:
		return scene.AnchorCenter
	case contracts.AnchorRight:
		return scene.AnchorCenterRight
	case contracts.AnchorBottomLeft:
		return scene.AnchorBottomLeft
	case contracts.AnchorBottom:
		return scene.AnchorBottomCenter
	case contracts.AnchorBottomRight:
		return scene.AnchorBottomRight
	default:
		return scene.AnchorTopLeft
	}
}

// mapVAlign converts the wire-level VAlign string to the scene enum.
// The empty string and "top" both map to VAlignTop (the zero value).
func mapVAlign(v contracts.VAlign) scene.VAlign {
	switch v {
	case contracts.VAlignCenter:
		return scene.VAlignCenter
	case contracts.VAlignBottom:
		return scene.VAlignBottom
	case contracts.VAlignJustify:
		return scene.VAlignJustify
	case contracts.VAlignFill:
		return scene.VAlignFill
	case contracts.VAlignBalanced:
		return scene.VAlignBalanced
	default:
		// VAlignTop ("top") and empty string both map to the zero value (top).
		return scene.VAlignTop
	}
}

// mapHAlign converts the wire-level HAlign string to the scene enum.
// The empty string and "left" both map to HAlignLeft (the zero value).
func mapHAlign(h contracts.HAlign) scene.HAlign {
	switch h {
	case contracts.HAlignCenter:
		return scene.HAlignCenter
	case contracts.HAlignRight:
		return scene.HAlignRight
	default:
		// HAlignLeft ("left") and empty string both map to the zero value (left).
		return scene.HAlignLeft
	}
}

// mapAlignment converts a contracts.Alignment to scene.Alignment.
func mapAlignment(a contracts.Alignment) scene.Alignment {
	return scene.Alignment{
		Vertical:   mapVAlign(a.Vertical),
		Horizontal: mapHAlign(a.Horizontal),
	}
}

// mapPosition converts a points-based contracts.Position into the EMU-based
// scene.Position. pptx.Pt performs the integer rounding.
func mapPosition(p contracts.Position) scene.Position {
	return scene.Position{X: pptx.Pt(p.X), Y: pptx.Pt(p.Y)}
}

// mapSize converts a points-based contracts.Size into the EMU-based
// scene.Size.
func mapSize(s contracts.Size) scene.Size {
	return scene.Size{W: pptx.Pt(s.W), H: pptx.Pt(s.H)}
}

// mapColorRolePtr maps an optional ColorRole to a *pptx.ColorRole.
// An empty role returns nil — the engine's zero/none sentinel for optional
// colored card elements such as HeaderFill and StatusDot (D-054). A non-empty
// role delegates to mapColorRole.
func mapColorRolePtr(role contracts.ColorRole) *pptx.ColorRole {
	if role == "" {
		return nil
	}
	r := mapColorRole(role)
	return &r
}

// mapVariant converts the wire-level Variant string to the scene enum.
// The empty string and "light" both map to VariantLight (the zero value).
func mapVariant(v contracts.Variant) scene.Variant {
	if v == contracts.VariantDark {
		return scene.VariantDark
	}
	return scene.VariantLight
}

// mapBackgroundKind converts the wire-level BackgroundKind string to the
// scene integer enum. The empty string maps to BackgroundNone (zero value).
func mapBackgroundKind(k contracts.BackgroundKind) scene.BackgroundKind {
	switch k {
	case contracts.BackgroundColor:
		return scene.BackgroundColor
	case contracts.BackgroundGradient:
		return scene.BackgroundGradient
	case contracts.BackgroundAsset:
		return scene.BackgroundAsset
	case contracts.BackgroundRadial:
		return scene.BackgroundRadial
	case contracts.BackgroundMesh:
		return scene.BackgroundMesh
	default:
		// BackgroundNone ("") and any unknown value map to BackgroundNone.
		return scene.BackgroundNone
	}
}

// mapDeltaTone converts the wire-level DeltaTone string to the scene integer
// enum. The empty string and "neutral" both map to DeltaNeutral (zero value).
func mapDeltaTone(t contracts.DeltaTone) scene.DeltaTone {
	switch t {
	case contracts.DeltaUp:
		return scene.DeltaUp
	case contracts.DeltaDown:
		return scene.DeltaDown
	default: // DeltaNeutral ("neutral") and empty string
		return scene.DeltaNeutral
	}
}

// mapNumberFormat converts a contracts.NumberFormat pointer to a
// scene.NumberFormat pointer, copying all 10 fields 1:1 (R14.13, D-121).
// nil maps to nil, keeping a Stat with no Format byte-identical.
func mapNumberFormat(f *contracts.NumberFormat) *scene.NumberFormat {
	if f == nil {
		return nil
	}
	return &scene.NumberFormat{
		Decimals:         f.Decimals,
		GroupSep:         f.GroupSep,
		DecimalSep:       f.DecimalSep,
		CurrencySymbol:   f.CurrencySymbol,
		SymbolAfter:      f.SymbolAfter,
		Percent:          f.Percent,
		Compact:          f.Compact,
		CompactThreshold: f.CompactThreshold,
		Prefix:           f.Prefix,
		Suffix:           f.Suffix,
	}
}

// mapTableStyle converts a contracts.TableStyle pointer to a
// scene.TableStyle pointer, copying the scalar fields 1:1 and mapping
// HeaderGroups element-wise (R14.3, D-118). nil maps to nil, keeping a Table
// with no Style byte-identical to the plain banded table.
func mapTableStyle(s *contracts.TableStyle) *scene.TableStyle {
	if s == nil {
		return nil
	}
	return &scene.TableStyle{
		HeaderFill:   s.HeaderFill,
		Zebra:        s.Zebra,
		HighlightCol: s.HighlightCol,
		RowLabelCol:  s.RowLabelCol,
		HeaderGroups: mapHeaderGroups(s.HeaderGroups),
	}
}

// mapHeaderGroups converts contracts.HeaderGroup slices to scene.HeaderGroup
// slices 1:1 (R14.3, D-118).
func mapHeaderGroups(groups []contracts.HeaderGroup) []scene.HeaderGroup {
	if groups == nil {
		return nil
	}
	mapped := make([]scene.HeaderGroup, len(groups))
	for i, g := range groups {
		mapped[i] = scene.HeaderGroup{Label: g.Label, Span: g.Span}
	}
	return mapped
}

// mapBackground converts a contracts.Background to scene.Background.
// The gradient slice is mapped to the engine's [2]pptx.ColorRole:
//   - 0 roles → both stops are the zero ColorRole
//   - 1 role  → both stops use the same role
//   - 2+ roles → first two roles are used as start and end stops
//
// Stops and Mesh (R13.2/R13.3/R13.4) map element-wise and stay nil when the
// product slice is empty.
func mapBackground(b contracts.Background) scene.Background {
	var grad [2]pptx.ColorRole
	switch len(b.Gradient) {
	case 0:
		// both stops remain zero
	case 1:
		grad[0] = mapColorRole(b.Gradient[0])
		grad[1] = mapColorRole(b.Gradient[0])
	default:
		grad[0] = mapColorRole(b.Gradient[0])
		grad[1] = mapColorRole(b.Gradient[1])
	}
	// Stops/Mesh stay nil when the product slices are empty — that keeps a
	// background that sets neither byte-identical to pre-R13.2/13.3/13.4
	// output (CLAUDE.md byte-identity contract).
	var stops []scene.GradientStop
	if len(b.Stops) > 0 {
		stops = make([]scene.GradientStop, len(b.Stops))
		for i, s := range b.Stops {
			stops[i] = scene.GradientStop{Pos: s.Pos, Color: mapColorRole(s.Color)}
		}
	}
	var mesh []scene.MeshGlow
	if len(b.Mesh) > 0 {
		mesh = make([]scene.MeshGlow, len(b.Mesh))
		for i, m := range b.Mesh {
			mesh[i] = scene.MeshGlow{
				Anchor: mapAnchor(m.Anchor),
				Color:  mapColorRole(m.Color),
				Radius: pptx.Pt(m.Radius),
				Alpha:  int(m.Alpha * 100000),
			}
		}
	}
	// Scrim/Duotone stay nil when the product field is unset — that keeps a
	// background that sets neither byte-identical to pre-R14.1 output
	// (CLAUDE.md byte-identity contract).
	var scrim *scene.Scrim
	if b.Scrim != nil {
		scrim = &scene.Scrim{
			Color:         mapColorRole(b.Scrim.Color),
			Opacity:       int(b.Scrim.Opacity * 100000),
			Gradient:      b.Scrim.Gradient,
			GradientAngle: b.Scrim.GradientAngle,
		}
	}
	var duo *scene.Duotone
	if b.Duotone != nil {
		duo = &scene.Duotone{
			Shadow:    mapColorRole(b.Duotone.Shadow),
			Highlight: mapColorRole(b.Duotone.Highlight),
		}
	}
	return scene.Background{
		Kind:         mapBackgroundKind(b.Kind),
		Color:        mapColorRole(b.Color),
		Gradient:     grad,
		Stops:        stops,
		Angle:        b.Angle,
		AssetID:      scene.AssetID(b.AssetID),
		GradientName: b.GradientName,
		Mesh:         mesh,
		Scrim:        scrim,
		Duotone:      duo,
	}
}
