package render

import (
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/pptx-go/scene"
)

func mapSlides(slides []contracts.Slide) []scene.SceneSlide {
	if slides == nil {
		return nil
	}
	mapped := make([]scene.SceneSlide, len(slides))
	for i, slide := range slides {
		mapped[i] = mapSlide(slide)
	}
	return mapped
}

func mapSlide(slide contracts.Slide) scene.SceneSlide {
	return scene.SceneSlide{
		ID:     slide.ID,
		Layout: mapLayoutKind(slide.Layout),
		Nodes:  mapNodes(slide.Nodes),
		Notes:  mapRichText(slide.Notes),
	}
}

func mapNodes(nodes []contracts.SlideNode) []scene.SlideNode {
	if nodes == nil {
		return nil
	}
	mapped := make([]scene.SlideNode, 0, len(nodes))
	for _, node := range nodes {
		if mappedNode := mapNode(node); mappedNode != nil {
			mapped = append(mapped, mappedNode)
		}
	}
	return mapped
}

func mapNode(node contracts.SlideNode) scene.SlideNode {
	switch n := node.(type) {
	case *contracts.Hero:
		return scene.Hero{Eyebrow: n.Eyebrow, Title: n.Title, Subtitle: n.Subtitle}
	case *contracts.Heading:
		return scene.Heading{Text: mapRichText(n.Text), Level: n.Level}
	case *contracts.Prose:
		return scene.Prose{Paragraphs: mapParagraphs(n.Paragraphs)}
	case *contracts.List:
		return scene.List{Kind: mapListKind(n.Kind), Items: mapListItems(n.Items)}
	case *contracts.Quote:
		return scene.Quote{Text: mapRichText(n.Text), Attribution: n.Attribution}
	case *contracts.Callout:
		return scene.Callout{Kind: mapCalloutKind(n.Kind), Title: n.Title, Body: mapRichText(n.Body)}
	case *contracts.Table:
		return scene.Table{Headers: mapParagraphs(n.Headers), Rows: mapTableRows(n.Rows), Caption: n.Caption}
	case *contracts.TwoColumn:
		return scene.TwoColumn{Ratio: mapColumnRatio(n.Ratio), Left: mapNodes(n.Left), Right: mapNodes(n.Right)}
	case *contracts.Grid:
		return scene.Grid{Columns: n.Columns, Ratio: cloneInts(n.Ratio), Gap: mapSpaceRole(n.Gap), Cells: mapNodes(n.Cells)}
	case *contracts.Card:
		return scene.Card{
			Header:      n.Header,
			Eyebrow:     n.Eyebrow,
			Icon:        n.Icon,
			HeaderPill:  n.HeaderPill,
			Body:        mapNodes(n.Body),
			BodyLayout:  mapBodyLayout(n.BodyLayout),
			Fill:        mapColorRole(n.Fill),
			Outline:     n.Outline,
			BorderStyle: mapBorderStyle(n.BorderStyle),
			Size:        mapCardSize(n.Size),
			Layout:      mapCardLayout(n.Layout),
			Elevation:   mapElevationRole(n.Elevation),
		}
	case *contracts.CardSection:
		return scene.CardSection{Header: n.Header, Body: mapNodes(n.Body)}
	default:
		return nil
	}
}

func mapParagraphs(paragraphs []contracts.RichText) []scene.RichText {
	if paragraphs == nil {
		return nil
	}
	mapped := make([]scene.RichText, len(paragraphs))
	for i, paragraph := range paragraphs {
		mapped[i] = mapRichText(paragraph)
	}
	return mapped
}

func mapListItems(items []contracts.ListItem) []scene.ListItem {
	if items == nil {
		return nil
	}
	mapped := make([]scene.ListItem, len(items))
	for i, item := range items {
		mapped[i] = scene.ListItem{Text: mapRichText(item.Text), Level: item.Level, Checked: item.Checked}
	}
	return mapped
}

func mapTableRows(rows [][]contracts.RichText) [][]scene.RichText {
	if rows == nil {
		return nil
	}
	mapped := make([][]scene.RichText, len(rows))
	for i, row := range rows {
		mapped[i] = mapParagraphs(row)
	}
	return mapped
}

func cloneInts(values []int) []int {
	if values == nil {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}
