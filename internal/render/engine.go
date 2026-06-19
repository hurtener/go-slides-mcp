package render

import (
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Stats is the render summary returned alongside the rendered PPTX bytes.
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []string
}

// Render maps doc+soul to a scene and renders deterministic .pptx bytes.
func Render(doc contracts.SlideDoc, s *soul.Soul) ([]byte, Stats, error) {
	return renderWithWorkers(doc, s, 0)
}

func renderWithWorkers(doc contracts.SlideDoc, s *soul.Soul, workers int) ([]byte, Stats, error) {
	if s == nil || s.Theme == nil {
		return nil, Stats{}, fmt.Errorf("render: nil soul theme")
	}

	pres := pptx.New(pptx.WithTheme(s.Theme))
	sc := scene.Scene{
		Theme:  s.Theme,
		Slides: mapSlides(doc.Slides),
		Meta: scene.Metadata{
			Title: doc.Title,
		},
	}

	sceneStats, err := scene.Render(pres, sc, scene.WithWorkers(workers))
	if err != nil {
		return nil, Stats{}, fmt.Errorf("render scene: %w", err)
	}

	buf, err := pres.WriteToBytes()
	if err != nil {
		return nil, Stats{}, fmt.Errorf("write pptx bytes: %w", err)
	}

	return buf, statsFromScene(sceneStats), nil
}

func statsFromScene(s scene.Stats) Stats {
	warnings := make([]string, 0, len(s.Warnings))
	for _, warning := range s.Warnings {
		warnings = append(warnings, fmt.Sprintf("slide=%s node=%s: %s", warning.SlideID, warning.Node, warning.Message))
	}
	return Stats{
		Slides:   s.Slides,
		Shapes:   s.Shapes,
		Assets:   s.Assets,
		Warnings: warnings,
	}
}
