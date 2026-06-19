package handlers

import (
	"log/slog"

	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// ToolDeps are the concrete dependencies shared by the tool handlers.
type ToolDeps struct {
	// Store persists decks and slides.
	Store *deck.MemoryStore
	// Souls resolves design souls by ID.
	Souls *soul.MemoryRegistry
	// Workspace is the server workspace root for file-backed operations.
	Workspace string
	// Logger is the process logger.
	Logger *slog.Logger
}

type handlers struct{ deps ToolDeps }
