package soul

import (
	"fmt"
	"sort"
	"sync"
)

// Registry stores design souls. It is the clonable storage seam — a file-backed
// impl can wrap the same interface later. Implementations return CLONES so a
// caller can never mutate a stored soul in place.
type Registry interface {
	// Get returns a clone of the soul with id, and whether it exists.
	Get(id string) (*Soul, bool)
	// List returns clones of every soul, ordered by id (built-ins first by id).
	List() []*Soul
	// Put stores a clone of s (insert or replace). An empty ID is an error.
	Put(s *Soul) error
}

// MemoryRegistry is an in-memory, concurrency-safe Registry.
type MemoryRegistry struct {
	mu    sync.RWMutex
	souls map[string]*Soul
}

// NewMemoryRegistry returns a registry seeded with the built-in Deckard White soul.
func NewMemoryRegistry() *MemoryRegistry {
	r := &MemoryRegistry{souls: make(map[string]*Soul)}
	_ = r.Put(DeckardWhite())
	return r
}

// Get returns a clone of the soul with id.
func (r *MemoryRegistry) Get(id string) (*Soul, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.souls[id]
	if !ok {
		return nil, false
	}
	return s.Clone(), true
}

// List returns clones of every soul, ordered by id.
func (r *MemoryRegistry) List() []*Soul {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.souls))
	for id := range r.souls {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]*Soul, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.souls[id].Clone())
	}
	return out
}

// Put stores a clone of s.
func (r *MemoryRegistry) Put(s *Soul) error {
	if s == nil || s.ID == "" {
		return fmt.Errorf("soul: Put requires a non-empty soul ID")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.souls[s.ID] = s.Clone()
	return nil
}
