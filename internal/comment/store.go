// Package comment stores human and agent collaboration comments in memory.
package comment

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// ErrNotFound reports a missing stored comment.
var ErrNotFound = errors.New("comment: not found")

// Comment is one stored collaboration comment.
type Comment struct {
	ID          string
	DeckID      string
	Target      Target
	Body        string
	Kind        string
	Origin      string
	Resolved    bool
	ResolveNote string
	CreatedAt   string
}

// Target identifies what a comment points at.
type Target struct {
	Kind    string
	SlideID string
	IRPath  []any
}

// Store is the comment storage seam used by the handlers.
type Store interface {
	Add(c Comment) (*Comment, error)
	List(deckID string, resolved *bool, targetKind string) []*Comment
	Resolve(id, note string) (*Comment, error)
}

// MemoryStore is a concurrency-safe in-memory comment store.
type MemoryStore struct {
	mu       sync.RWMutex
	comments map[string]*Comment
	order    []string
}

// NewMemoryStore returns an empty in-memory comment store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{comments: make(map[string]*Comment)}
}

// Add stores one comment snapshot with a generated comment ID and timestamp.
func (s *MemoryStore) Add(c Comment) (*Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := cloneComment(&c)
	stored.ID = newCommentID()
	stored.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	s.comments[stored.ID] = stored
	s.order = append(s.order, stored.ID)
	return cloneComment(stored), nil
}

// List returns snapshot copies of comments for one deck, in insertion order.
func (s *MemoryStore) List(deckID string, resolved *bool, targetKind string) []*Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Comment, 0, len(s.order))
	for _, id := range s.order {
		stored, ok := s.comments[id]
		if !ok || stored.DeckID != deckID {
			continue
		}
		if resolved != nil && stored.Resolved != *resolved {
			continue
		}
		if targetKind != "" && stored.Target.Kind != targetKind {
			continue
		}
		out = append(out, cloneComment(stored))
	}
	return out
}

// Resolve marks one stored comment resolved and records the optional note.
func (s *MemoryStore) Resolve(id, note string) (*Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.comments[id]
	if !ok {
		return nil, ErrNotFound
	}
	stored.Resolved = true
	stored.ResolveNote = note
	return cloneComment(stored), nil
}

func newCommentID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "cmt_" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return "cmt_" + base64.RawURLEncoding.EncodeToString(b)
}

func cloneComment(c *Comment) *Comment {
	if c == nil {
		return nil
	}
	cloned := &Comment{
		ID:          c.ID,
		DeckID:      c.DeckID,
		Target:      Target{Kind: c.Target.Kind, SlideID: c.Target.SlideID, IRPath: append([]any(nil), c.Target.IRPath...)},
		Body:        c.Body,
		Kind:        c.Kind,
		Origin:      c.Origin,
		Resolved:    c.Resolved,
		ResolveNote: c.ResolveNote,
		CreatedAt:   c.CreatedAt,
	}
	return cloned
}
