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
		mapped[i] = mapSlide(slide, i)
	}
	return mapped
}

func mapSlide(slide contracts.Slide, idx int) scene.SceneSlide {
	ss := scene.SceneSlide{
		ID:         slide.ID,
		Layout:     mapLayoutKind(slide.Layout),
		Content:    mapAlignment(slide.Align),
		Variant:    mapVariant(slide.Variant),
		Nodes:      mapNodes(slide.Nodes),
		Notes:      mapRichText(slide.Notes),
		Section:    slide.Section,
		PageNumber: idx + 1, // 1-based; engine default (0) would also resolve here
	}
	if slide.Background != nil {
		ss.Background = mapBackground(*slide.Background)
	}
	return ss
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
		return scene.Hero{Eyebrow: n.Eyebrow, Title: n.Title, Subtitle: n.Subtitle, Align: mapHAlign(n.Align), AutoFit: n.AutoFit}
	case *contracts.Heading:
		return scene.Heading{Text: mapRichText(n.Text), Level: n.Level, Align: mapHAlign(n.Align), AutoFit: n.AutoFit}
	case *contracts.Prose:
		return scene.Prose{Paragraphs: mapParagraphs(n.Paragraphs), Align: mapHAlign(n.Align)}
	case *contracts.List:
		return scene.List{Kind: mapListKind(n.Kind), Items: mapListItems(n.Items)}
	case *contracts.Divider:
		return scene.Divider{Spacing: mapSpaceRole(n.Spacing)}
	case *contracts.Quote:
		return scene.Quote{Text: mapRichText(n.Text), Attribution: n.Attribution, Align: mapHAlign(n.Align)}
	case *contracts.Callout:
		return scene.Callout{Kind: mapCalloutKind(n.Kind), Title: n.Title, Body: mapRichText(n.Body)}
	case *contracts.Chip:
		return scene.Chip{Label: n.Label, Tone: mapChipTone(n.Tone), Color: mapColorRole(n.Color), Align: mapHAlign(n.Align)}
	case *contracts.Arrow:
		return scene.Arrow{Direction: mapArrowDirection(n.Direction), Label: n.Label}
	case *contracts.Table:
		return scene.Table{Headers: mapParagraphs(n.Headers), Rows: mapTableRows(n.Rows), Caption: n.Caption, Style: mapTableStyle(n.Style)}
	case *contracts.Flow:
		return scene.Flow{Orientation: mapFlowOrientation(n.Orientation), Steps: mapFlowSteps(n.Steps), Connector: mapConnectorKind(n.Connector)}
	case *contracts.SectionDivider:
		return scene.SectionDivider{Eyebrow: n.Eyebrow, Label: n.Label, Align: mapHAlign(n.Align)}
	case *contracts.Image:
		return scene.Image{
			AssetID:      scene.AssetID(n.AssetID),
			Alt:          n.Alt,
			Frame:        mapFrameKind(n.Frame),
			FrameName:    n.FrameName,
			Crop:         mapCrop(n.Crop),
			Fit:          mapFit(n.Fit),
			CornerRadius: mapRadiusRole(n.CornerRadius),
			Elevation:    mapElevationRole(n.Elevation),
		}
	case *contracts.CodeBlock:
		return scene.CodeBlock{AssetID: scene.AssetID(n.AssetID), Language: n.Language, Caption: n.Caption}
	case *contracts.Chart:
		return scene.Chart{AssetID: scene.AssetID(n.AssetID), Caption: n.Caption}
	case *contracts.Decoration:
		return mapDecoration(*n)
	case *contracts.TwoColumn:
		return scene.TwoColumn{
			Ratio:     mapColumnRatio(n.Ratio),
			Left:      mapNodes(n.Left),
			Right:     mapNodes(n.Right),
			Join:      mapColumnJoin(n.Join),
			JoinLabel: n.JoinLabel,
		}
	case *contracts.Bento:
		return mapBento(n)
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
			HeaderFill:  mapColorRolePtr(n.HeaderFill),
			StatusDot:   mapColorRolePtr(n.StatusDot),
			Watermark:   n.Watermark,
			Backdrop:    mapDecorationPtr(n.Backdrop),
			ImageFill:   scene.AssetID(n.ImageFill),
		}
	case *contracts.CardSection:
		return scene.CardSection{Header: n.Header, Body: mapNodes(n.Body)}
	case *contracts.Stat:
		return scene.Stat{Value: n.Value, Label: n.Label, Delta: n.Delta, DeltaTone: mapDeltaTone(n.DeltaTone), AutoFit: n.AutoFit, Number: n.Number, Format: mapNumberFormat(n.Format)}
	case *contracts.Timeline:
		return scene.Timeline{Milestones: mapMilestones(n.Milestones), Lanes: mapTimelineLanes(n.Lanes), Bands: mapTimelineBands(n.Bands)}
	default:
		return nil
	}
}

// mapMilestones maps a Timeline's milestone list, preserving nil vs. empty
// (R14.4, D-119).
func mapMilestones(milestones []contracts.Milestone) []scene.Milestone {
	if milestones == nil {
		return nil
	}
	mapped := make([]scene.Milestone, len(milestones))
	for i, m := range milestones {
		mapped[i] = scene.Milestone{
			Position:    m.Position,
			Label:       m.Label,
			Detail:      m.Detail,
			Icon:        m.Icon,
			AccentIndex: m.AccentIndex,
		}
	}
	return mapped
}

// mapTimelineLanes maps a Timeline's swimlanes (R14.4, D-119).
func mapTimelineLanes(lanes []contracts.TimelineLane) []scene.TimelineLane {
	if lanes == nil {
		return nil
	}
	mapped := make([]scene.TimelineLane, len(lanes))
	for i, ln := range lanes {
		mapped[i] = scene.TimelineLane{Label: ln.Label, Milestones: mapMilestones(ln.Milestones)}
	}
	return mapped
}

// mapTimelineBands maps a Timeline's phase/horizon bands (R14.4, D-119).
func mapTimelineBands(bands []contracts.TimelineBand) []scene.TimelineBand {
	if bands == nil {
		return nil
	}
	mapped := make([]scene.TimelineBand, len(bands))
	for i, b := range bands {
		mapped[i] = scene.TimelineBand{From: b.From, To: b.To, Label: b.Label, Fill: mapColorRole(b.Fill)}
	}
	return mapped
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

func mapFlowSteps(steps []contracts.FlowStep) []scene.FlowStep {
	if steps == nil {
		return nil
	}
	mapped := make([]scene.FlowStep, len(steps))
	for i, step := range steps {
		mapped[i] = scene.FlowStep{Label: mapRichText(step.Label), Detail: mapRichText(step.Detail), Icon: step.Icon}
	}
	return mapped
}

func mapBento(n *contracts.Bento) scene.Bento {
	rows := make([]scene.BentoRow, len(n.Rows))
	for ri, row := range n.Rows {
		cells := make([]scene.BentoCell, len(row.Cells))
		for ci, cell := range row.Cells {
			cells[ci] = scene.BentoCell{
				Span: cell.Span,
				Node: mapNode(cell.Node),
			}
		}
		rows[ri] = scene.BentoRow{Label: row.Label, Cells: cells}
	}
	return scene.Bento{Columns: n.Columns, Rows: rows}
}

func cloneInts(values []int) []int {
	if values == nil {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}
