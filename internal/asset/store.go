// Package asset stores uploaded binary assets behind opaque asset URIs.
package asset

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// ErrNotFound reports a missing stored asset.
var ErrNotFound = errors.New("asset: not found")

// Asset is one stored binary plus its metadata.
type Asset struct {
	ID        string
	Filename  string
	MIME      string
	Bytes     []byte
	CreatedAt string
}

// Store is the asset storage seam used by the handlers.
type Store interface {
	Put(filename, mime string, data []byte) (*Asset, error)
	Get(id string) (*Asset, bool)
	List() []*Asset
	Delete(id string) error
}

// MemoryStore is a concurrency-safe in-memory asset store.
type MemoryStore struct {
	mu     sync.RWMutex
	assets map[string]*Asset
	order  []string
}

// NewMemoryStore returns an empty in-memory asset store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{assets: make(map[string]*Asset)}
}

// Put stores one binary snapshot with a generated opaque asset URI.
func (s *MemoryStore) Put(filename, mime string, data []byte) (*Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset := &Asset{
		ID:        newAssetID(),
		Filename:  filename,
		MIME:      mime,
		Bytes:     append([]byte(nil), data...),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.assets[asset.ID] = asset
	s.order = append(s.order, asset.ID)
	return cloneAsset(asset), nil
}

// Get resolves one asset by opaque asset URI.
func (s *MemoryStore) Get(id string) (*Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, ok := s.assets[id]
	if !ok {
		return nil, false
	}
	return cloneAsset(asset), true
}

// List returns snapshot copies of every stored asset in insertion order.
func (s *MemoryStore) List() []*Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Asset, 0, len(s.order))
	for _, id := range s.order {
		asset, ok := s.assets[id]
		if !ok {
			continue
		}
		out = append(out, cloneAsset(asset))
	}
	return out
}

// Delete removes one stored asset by opaque asset URI.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.assets[id]; !ok {
		return ErrNotFound
	}
	delete(s.assets, id)
	s.order = removeString(s.order, id)
	return nil
}

func newAssetID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "asset://" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return "asset://" + base64.RawURLEncoding.EncodeToString(b)
}

func cloneAsset(a *Asset) *Asset {
	if a == nil {
		return nil
	}
	return &Asset{
		ID:        a.ID,
		Filename:  a.Filename,
		MIME:      a.MIME,
		Bytes:     append([]byte(nil), a.Bytes...),
		CreatedAt: a.CreatedAt,
	}
}

func removeString(values []string, want string) []string {
	for i, value := range values {
		if value == want {
			return append(values[:i], values[i+1:]...)
		}
	}
	return values
}
