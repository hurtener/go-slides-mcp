package deck

import (
	"fmt"
	"sync"
	"time"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// MemoryStore is a concurrency-safe in-memory deck store.
type MemoryStore struct {
	mu       sync.RWMutex
	decks    map[string]*Deck
	slugToID map[string]string
	order    []string
}

// NewMemoryStore returns an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		decks:    make(map[string]*Deck),
		slugToID: make(map[string]string),
	}
}

// CreateDeck stores a new deck with generated identifiers and timestamps.
func (s *MemoryStore) CreateDeck(in CreateDeckInput) (*Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestampNow()
	deck := &Deck{
		ID:        NewDeckID(),
		Slug:      s.uniqueSlugLocked(Slugify(in.Title)),
		Title:     in.Title,
		Author:    in.Author,
		SoulID:    in.SoulID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	rev, err := computeRevision(deck)
	if err != nil {
		return nil, err
	}
	deck.Revision = rev
	s.decks[deck.ID] = deck
	s.slugToID[deck.Slug] = deck.ID
	s.order = append(s.order, deck.ID)
	return cloneDeck(deck)
}

// ListDecks returns snapshot copies of every stored deck.
func (s *MemoryStore) ListDecks() []*Deck {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Deck, 0, len(s.order))
	for _, id := range s.order {
		deck, ok := s.decks[id]
		if !ok {
			continue
		}
		copyDeck, err := cloneDeck(deck)
		if err != nil {
			continue
		}
		out = append(out, copyDeck)
	}
	return out
}

// GetDeck resolves a deck by ID or slug and returns a snapshot copy.
func (s *MemoryStore) GetDeck(idOrSlug string) (*Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	return cloneDeck(deck)
}

// DeleteDeck removes a deck addressed by ID or slug.
func (s *MemoryStore) DeleteDeck(idOrSlug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return err
	}
	delete(s.decks, deck.ID)
	delete(s.slugToID, deck.Slug)
	s.order = removeString(s.order, deck.ID)
	return nil
}

// SetChrome replaces a deck's chrome configuration.
func (s *MemoryStore) SetChrome(idOrSlug string, c Chrome) (*Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	deck.Chrome = c
	if err := s.touchLocked(deck); err != nil {
		return nil, err
	}
	return cloneDeck(deck)
}

// SetSections replaces a deck's section grouping.
func (s *MemoryStore) SetSections(idOrSlug string, sections []Section) (*Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	deck.Sections = cloneSections(sections)
	if err := s.touchLocked(deck); err != nil {
		return nil, err
	}
	return cloneDeck(deck)
}

// AddSlide inserts a new slide snapshot at the requested position.
func (s *MemoryStore) AddSlide(idOrSlug string, slide contracts.Slide, position *int) (*Deck, *contracts.Slide, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, nil, err
	}
	copySlide, err := cloneSlide(slide)
	if err != nil {
		return nil, nil, err
	}
	copySlide.ID = NewSlideID()
	idx, err := normalizePosition(position, len(deck.Slides))
	if err != nil {
		return nil, nil, err
	}
	deck.Slides = insertSlide(deck.Slides, idx, *copySlide)
	if err := s.touchLocked(deck); err != nil {
		return nil, nil, err
	}
	deckCopy, err := cloneDeck(deck)
	if err != nil {
		return nil, nil, err
	}
	slideCopy, err := cloneSlide(*copySlide)
	if err != nil {
		return nil, nil, err
	}
	return deckCopy, slideCopy, nil
}

// UpdateSlide replaces one stored slide, enforcing optimistic concurrency when requested.
func (s *MemoryStore) UpdateSlide(idOrSlug, slideID string, slide contracts.Slide, expectedRevision string) (*Deck, *contracts.Slide, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, nil, err
	}
	if expectedRevision != "" && expectedRevision != deck.Revision {
		return nil, nil, fmt.Errorf("deck: expected revision %q, current %q: %w", expectedRevision, deck.Revision, ErrRevisionConflict)
	}
	idx := slideIndex(deck.Slides, slideID)
	if idx < 0 {
		return nil, nil, fmt.Errorf("deck: slide %q: %w", slideID, ErrNotFound)
	}
	copySlide, err := cloneSlide(slide)
	if err != nil {
		return nil, nil, err
	}
	if copySlide.ID == "" {
		copySlide.ID = NewSlideID()
	}
	deck.Slides[idx] = *copySlide
	if err := s.touchLocked(deck); err != nil {
		return nil, nil, err
	}
	deckCopy, err := cloneDeck(deck)
	if err != nil {
		return nil, nil, err
	}
	slideCopy, err := cloneSlide(*copySlide)
	if err != nil {
		return nil, nil, err
	}
	return deckCopy, slideCopy, nil
}

// GetSlide returns a snapshot copy of one slide by deck and slide identifier.
func (s *MemoryStore) GetSlide(idOrSlug, slideID string) (*contracts.Slide, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	idx := slideIndex(deck.Slides, slideID)
	if idx < 0 {
		return nil, fmt.Errorf("deck: slide %q: %w", slideID, ErrNotFound)
	}
	return cloneSlide(deck.Slides[idx])
}

// RemoveSlide deletes one slide from the deck and prunes section references to it.
func (s *MemoryStore) RemoveSlide(idOrSlug, slideID string) (*Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	idx := slideIndex(deck.Slides, slideID)
	if idx < 0 {
		return nil, fmt.Errorf("deck: slide %q: %w", slideID, ErrNotFound)
	}
	deck.Slides = append(deck.Slides[:idx], deck.Slides[idx+1:]...)
	deck.Sections = removeSlideFromSections(deck.Sections, slideID)
	if err := s.touchLocked(deck); err != nil {
		return nil, err
	}
	return cloneDeck(deck)
}

// ReorderSlides replaces the deck's slide order with the provided complete order.
func (s *MemoryStore) ReorderSlides(idOrSlug string, order []string) (*Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, err
	}
	if len(order) != len(deck.Slides) {
		return nil, fmt.Errorf("deck: reorder needs %d slide ids, got %d", len(deck.Slides), len(order))
	}
	byID := make(map[string]contracts.Slide, len(deck.Slides))
	for _, slide := range deck.Slides {
		byID[slide.ID] = slide
	}
	reordered := make([]contracts.Slide, len(order))
	seen := make(map[string]struct{}, len(order))
	for i, id := range order {
		slide, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("deck: slide %q: %w", id, ErrNotFound)
		}
		if _, dup := seen[id]; dup {
			return nil, fmt.Errorf("deck: duplicate slide id %q in reorder", id)
		}
		seen[id] = struct{}{}
		reordered[i] = slide
	}
	deck.Slides = reordered
	if err := s.touchLocked(deck); err != nil {
		return nil, err
	}
	return cloneDeck(deck)
}

// DuplicateSlide deep-copies one slide, assigns a new ID, and inserts it.
func (s *MemoryStore) DuplicateSlide(idOrSlug, slideID string, position *int) (*Deck, *contracts.Slide, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, err := s.resolveDeckLocked(idOrSlug)
	if err != nil {
		return nil, nil, err
	}
	idx := slideIndex(deck.Slides, slideID)
	if idx < 0 {
		return nil, nil, fmt.Errorf("deck: slide %q: %w", slideID, ErrNotFound)
	}
	copySlide, err := cloneSlide(deck.Slides[idx])
	if err != nil {
		return nil, nil, err
	}
	copySlide.ID = NewSlideID()
	insertAt, err := normalizePosition(position, len(deck.Slides))
	if err != nil {
		return nil, nil, err
	}
	deck.Slides = insertSlide(deck.Slides, insertAt, *copySlide)
	if err := s.touchLocked(deck); err != nil {
		return nil, nil, err
	}
	deckCopy, err := cloneDeck(deck)
	if err != nil {
		return nil, nil, err
	}
	slideCopy, err := cloneSlide(*copySlide)
	if err != nil {
		return nil, nil, err
	}
	return deckCopy, slideCopy, nil
}

func (s *MemoryStore) resolveDeckLocked(idOrSlug string) (*Deck, error) {
	if deck, ok := s.decks[idOrSlug]; ok {
		return deck, nil
	}
	if id, ok := s.slugToID[idOrSlug]; ok {
		if deck, ok := s.decks[id]; ok {
			return deck, nil
		}
	}
	return nil, fmt.Errorf("deck %q: %w", idOrSlug, ErrNotFound)
}

func (s *MemoryStore) uniqueSlugLocked(base string) string {
	if _, ok := s.slugToID[base]; !ok {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, ok := s.slugToID[candidate]; !ok {
			return candidate
		}
	}
}

func (s *MemoryStore) touchLocked(deck *Deck) error {
	deck.UpdatedAt = timestampNow()
	rev, err := computeRevision(deck)
	if err != nil {
		return err
	}
	deck.Revision = rev
	return nil
}

func timestampNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func normalizePosition(position *int, length int) (int, error) {
	if position == nil {
		return length, nil
	}
	if *position < 0 || *position > length {
		return 0, fmt.Errorf("deck: position %d out of range [0,%d]", *position, length)
	}
	return *position, nil
}

func slideIndex(slides []contracts.Slide, slideID string) int {
	for i, slide := range slides {
		if slide.ID == slideID {
			return i
		}
	}
	return -1
}

func insertSlide(slides []contracts.Slide, idx int, slide contracts.Slide) []contracts.Slide {
	slides = append(slides, contracts.Slide{})
	copy(slides[idx+1:], slides[idx:])
	slides[idx] = slide
	return slides
}

func removeSlideFromSections(sections []Section, slideID string) []Section {
	if len(sections) == 0 {
		return nil
	}
	out := make([]Section, len(sections))
	for i, section := range sections {
		out[i].Name = section.Name
		for _, id := range section.SlideIDs {
			if id != slideID {
				out[i].SlideIDs = append(out[i].SlideIDs, id)
			}
		}
	}
	return out
}

func removeString(items []string, needle string) []string {
	for i, item := range items {
		if item == needle {
			return append(items[:i], items[i+1:]...)
		}
	}
	return items
}
