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
	case contracts.ColorSurface:
		fallthrough
	default:
		return scene.ColorSurface
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
