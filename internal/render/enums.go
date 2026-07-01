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

func mapConnectorKind(kind contracts.ConnectorKind) scene.ConnectorKind {
	switch kind {
	case contracts.ConnectorArrowDashed:
		return scene.ConnectorArrowDashed
	case contracts.ConnectorCycle:
		return scene.ConnectorCycle
	case contracts.ConnectorPlus:
		return scene.ConnectorPlus
	case contracts.ConnectorArrow:
		fallthrough
	default:
		return scene.ConnectorArrow
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

// mapDecorationKind converts the wire-level DecorationKind string to the scene
// enum (preset vs asset).
func mapDecorationKind(kind contracts.DecorationKind) scene.DecorationKind {
	if kind == contracts.DecorationAsset {
		return scene.DecorationAsset
	}
	return scene.DecorationPreset
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

// mapBackground converts a contracts.Background to scene.Background.
// The gradient slice is mapped to the engine's [2]pptx.ColorRole:
//   - 0 roles → both stops are the zero ColorRole
//   - 1 role  → both stops use the same role
//   - 2+ roles → first two roles are used as start and end stops
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
	return scene.Background{
		Kind:         mapBackgroundKind(b.Kind),
		Color:        mapColorRole(b.Color),
		Gradient:     grad,
		Angle:        b.Angle,
		AssetID:      scene.AssetID(b.AssetID),
		GradientName: b.GradientName,
	}
}
