package exportstore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// Export renders one deck and writes it to its deterministic workspace path.
func Export(workspace, deckID string, d contracts.SlideDoc, s *soul.Soul) (path string, stats render.Stats, err error) {
	buf, stats, err := render.Render(d, s)
	if err != nil {
		return "", render.Stats{}, fmt.Errorf("export render: %w", err)
	}
	path = ExportPath(workspace, deckID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", render.Stats{}, fmt.Errorf("mkdir exports dir: %w", err)
	}
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		return "", render.Stats{}, fmt.Errorf("write export file: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", render.Stats{}, fmt.Errorf("resolve export path: %w", err)
	}
	return absPath, stats, nil
}
