package handlers

import (
	"log/slog"
	"sync"

	"github.com/hurtener/go-slides-mcp/internal/asset"
	"github.com/hurtener/go-slides-mcp/internal/comment"
	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
	"github.com/hurtener/go-slides-mcp/internal/recipe"
	"github.com/hurtener/go-slides-mcp/internal/soul"
)

// SessionState is the small in-memory workspace session shared by handlers.
type SessionState struct {
	mu           sync.RWMutex
	activeDeckID string
	activeSoulID string
	openPanels   []string
}

// Snapshot returns a copy of the current session state.
func (s *SessionState) Snapshot() (activeDeckID, activeSoulID string, openPanels []string) {
	if s == nil {
		return "", "", []string{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeDeckID, s.activeSoulID, append([]string{}, s.openPanels...)
}

// SetActive replaces the active deck and soul identifiers under the write lock.
func (s *SessionState) SetActive(deckID, soulID string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeDeckID = deckID
	s.activeSoulID = soulID
}

// ToolDeps are the concrete dependencies shared by the tool handlers.
type ToolDeps struct {
	// Store persists decks and slides.
	Store *deck.MemoryStore
	// Souls resolves design souls by ID.
	Souls *soul.MemoryRegistry
	// Assets persists uploaded binary assets in memory.
	Assets *asset.MemoryStore
	// Comments persists collaboration comments in memory.
	Comments *comment.MemoryStore
	// Recipes persists reusable slide templates in memory.
	Recipes *recipe.MemoryStore
	// Session is the small in-memory workspace session state.
	Session *SessionState
	// BuildInfo identifies the running Deckard build.
	BuildInfo contracts.BuildInfo
	// Workspace is the server workspace root for file-backed operations.
	Workspace string
	// Logger is the process logger.
	Logger *slog.Logger
}

type handlers struct{ deps ToolDeps }
