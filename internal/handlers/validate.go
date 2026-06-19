package handlers

import (
	"context"
	"fmt"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
	"github.com/hurtener/go-slides-mcp/internal/raster"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/go-slides-mcp/internal/validate"
)

func (h *handlers) validateSlideIR(_ context.Context, in contracts.ValidateSlideIRInput) (tool.Result[contracts.ValidateSlideIROutput], error) {
	s := h.resolveSoul(in.SoulID)
	report := validate.Slide(in.Slide, s.Theme, h.renderWarnings(contracts.SlideDoc{Slides: []contracts.Slide{in.Slide}}, s))
	out := contracts.ValidateSlideIROutput{
		OK:       report.Score.Passed,
		Score:    report.Score.Score,
		Issues:   report.Messages(),
		Findings: findings(report.Issues),
	}
	return tool.Result[contracts.ValidateSlideIROutput]{Text: scoreText("slide IR", out.OK, out.Score, out.Issues), Structured: out}, nil
}

func (h *handlers) validateSlide(_ context.Context, in contracts.ValidateSlideInput) (tool.Result[contracts.ValidateSlideOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.ValidateSlideOutput]{}, mapDeckError(in.DeckID, err)
	}
	slide, err := h.deps.Store.GetSlide(in.DeckID, in.SlideID)
	if err != nil {
		return tool.Result[contracts.ValidateSlideOutput]{}, mapDeckError(in.DeckID, err)
	}

	s := h.resolveSoul(stored.SoulID)
	report := validate.Slide(*slide, s.Theme, h.renderWarnings(contracts.SlideDoc{Slides: []contracts.Slide{*slide}}, s))
	out := contracts.ValidateSlideOutput{
		SlideID:  in.SlideID,
		OK:       report.Score.Passed,
		Score:    report.Score.Score,
		Issues:   report.Messages(),
		Findings: findings(report.Issues),
	}
	return tool.Result[contracts.ValidateSlideOutput]{Text: scoreText(fmt.Sprintf("slide %q", in.SlideID), out.OK, out.Score, out.Issues), Structured: out}, nil
}

func (h *handlers) validateDeckForExport(_ context.Context, in contracts.ValidateDeckForExportInput) (tool.Result[contracts.ValidateDeckForExportOutput], error) {
	stored, err := h.deps.Store.GetDeck(in.DeckID)
	if err != nil {
		return tool.Result[contracts.ValidateDeckForExportOutput]{}, mapDeckError(in.DeckID, err)
	}

	doc := contracts.SlideDoc{Title: stored.Title, Slides: append([]contracts.Slide(nil), stored.Slides...)}
	s := h.resolveSoul(stored.SoulID)
	deckReport, perSlide := validate.Deck(doc, s.Theme, [][]string{h.renderWarnings(doc, s)})

	out := contracts.ValidateDeckForExportOutput{
		OK:       deckReport.Score.Passed,
		Score:    deckReport.Score.Score,
		Blockers: deckReport.Messages(),
		Findings: findings(deckReport.Issues),
		PerSlide: make([]contracts.DeckSlideValidation, 0, len(stored.Slides)),
	}
	for i, slide := range stored.Slides {
		r := perSlide[i]
		out.PerSlide = append(out.PerSlide, contracts.DeckSlideValidation{
			SlideID: slide.ID,
			OK:      r.Score.Passed,
			Score:   r.Score.Score,
			Issues:  r.Messages(),
		})
	}
	return tool.Result[contracts.ValidateDeckForExportOutput]{Text: scoreText(fmt.Sprintf("deck %q", stored.ID), out.OK, out.Score, out.Blockers), Structured: out}, nil
}

// resolveSoul returns the named soul, or the built-in Deckard White default
// (matching export behavior) when the id is empty or unknown.
func (h *handlers) resolveSoul(soulID string) *soul.Soul {
	if soulID != "" {
		if s, ok := h.deps.Souls.Get(soulID); ok {
			return s
		}
	}
	return soul.DeckardWhite()
}

// renderWarnings renders a doc and returns the layout warnings (the render-truth
// overflow pass). A render error is reported as a warning rather than failing
// validation.
func (h *handlers) renderWarnings(doc contracts.SlideDoc, s *soul.Soul) []string {
	_, stats, err := render.RenderWithAssets(doc, s, raster.NewStoreResolver(h.deps.Assets))
	if err != nil {
		return []string{"render failed: " + err.Error()}
	}
	return stats.Warnings
}

func findings(issues []validate.Issue) []contracts.StyleFinding {
	if len(issues) == 0 {
		return nil
	}
	out := make([]contracts.StyleFinding, 0, len(issues))
	for _, is := range issues {
		out = append(out, contracts.StyleFinding{
			Category: string(is.Category),
			Severity: string(is.Severity),
			Message:  is.Message,
			Path:     is.Path,
		})
	}
	return out
}

func scoreText(target string, ok bool, score float64, issues []string) string {
	status := "OK"
	if !ok {
		status = fmt.Sprintf("%d issue(s)", len(issues))
	}
	return fmt.Sprintf("Validated %s: %s (StyleScore %.2f).", target, status, score)
}

// validateDoc remains for the inline structural validation used by the surface
// handlers (fast, no render/contrast pass).
func validateDoc(doc contracts.SlideDoc) contracts.ValidateDeckForExportOutput {
	err := ir.ValidateDoc(doc)
	if err == nil {
		return contracts.ValidateDeckForExportOutput{OK: true, Score: 1}
	}
	issues := collectIssues(err)
	blockers := make([]string, 0, len(issues))
	blockers = append(blockers, issues...)
	return contracts.ValidateDeckForExportOutput{OK: false, Blockers: blockers}
}
