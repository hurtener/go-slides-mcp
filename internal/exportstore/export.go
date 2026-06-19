package exportstore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/render"
	"github.com/hurtener/go-slides-mcp/internal/soul"
	"github.com/hurtener/pptx-go/scene"
)

// Export renders one deck and writes it to its deterministic workspace path.
// Asset-free decks (no Image/Chart/CodeBlock/asset-Decoration references) are
// the canonical case; see ExportWithResolver when asset bytes must be resolved
// at render time. Export is equivalent to ExportWithResolver(_, _, _, _, nil).
func Export(workspace, deckID string, d contracts.SlideDoc, s *soul.Soul) (path string, stats render.Stats, err error) {
	return ExportWithResolver(workspace, deckID, d, s, nil)
}

// ExportWithResolver renders doc and writes it to its deterministic workspace
// path while threading resolver through to scene.WithAssetResolver so any
// asset-backed nodes (Image/Chart/CodeBlock/asset Decoration) can resolve
// their bytes from the asset store. A nil resolver renders the asset-free
// path; missing assets remain warn-don't-fail (D-036).
func ExportWithResolver(workspace, deckID string, d contracts.SlideDoc, s *soul.Soul, resolver scene.AssetResolver) (path string, stats render.Stats, err error) {
	buf, stats, err := render.RenderWithAssets(d, s, resolver)
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
