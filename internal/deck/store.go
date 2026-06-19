package deck

import (
	"errors"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// ErrRevisionConflict reports an optimistic concurrency mismatch.
var ErrRevisionConflict = errors.New("deck: revision conflict")

// ErrNotFound reports a missing deck or slide.
var ErrNotFound = errors.New("deck: not found")

// CreateDeckInput is the create-deck payload.
type CreateDeckInput struct {
	// Title is the initial deck title.
	Title string
	// Author is the initial deck author.
	Author string
	// SoulID is the initial design soul.
	SoulID string
}

// Store is the deck storage seam used by the handlers.
type Store interface {
	CreateDeck(in CreateDeckInput) (*Deck, error)
	ListDecks() []*Deck
	GetDeck(idOrSlug string) (*Deck, error)
	DeleteDeck(idOrSlug string) error
	SetChrome(idOrSlug string, c Chrome) (*Deck, error)
	SetSections(idOrSlug string, s []Section) (*Deck, error)
	AddSlide(idOrSlug string, slide contracts.Slide, position *int) (*Deck, *contracts.Slide, error)
	UpdateSlide(idOrSlug, slideID string, slide contracts.Slide, expectedRevision string) (*Deck, *contracts.Slide, error)
	GetSlide(idOrSlug, slideID string) (*contracts.Slide, error)
	RemoveSlide(idOrSlug, slideID string) (*Deck, error)
	ReorderSlides(idOrSlug string, order []string) (*Deck, error)
	DuplicateSlide(idOrSlug, slideID string, position *int) (*Deck, *contracts.Slide, error)
}
