package render

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

func mapRichText(text contracts.RichText) scene.RichText {
	if text == nil {
		return nil
	}
	mapped := make(scene.RichText, len(text))
	for i, run := range text {
		mapped[i] = scene.TextRun{
			Text:  run.Text,
			Style: mapRunStyle(run.Style()),
			Color: mapTextColor(run.Color),
		}
	}
	return mapped
}

func mapRunStyle(style contracts.RunStyle) scene.RunStyle {
	return scene.RunStyle{
		TypeRole:  mapTypeRole(style.TypeRole),
		Bold:      style.Bold,
		Italic:    style.Italic,
		Underline: style.Underline,
		Strike:    style.Strike,
		Code:      style.Code,
		Link:      style.Link,
		Href:      style.Href,
	}
}

func mapTextColor(color contracts.TextColor) scene.TextColor {
	if color.Literal != "" {
		return scene.LiteralColor(color.Literal)
	}
	return scene.TokenTextColor(mapTextColorRole(color.Token))
}
