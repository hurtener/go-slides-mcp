// Package recipe stores reusable slide templates in memory.
package recipe

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/deck"
)

// ErrNotFound reports a missing stored recipe.
var ErrNotFound = errors.New("recipe: not found")

// Recipe is one reusable slide template.
type Recipe struct {
	ID          string
	Name        string
	Description string
	Source      string
	Tags        []string
	Slide       contracts.Slide
}

// Store is the recipe storage seam used by the handlers.
type Store interface {
	Save(r Recipe) (*Recipe, error)
	List(tag string) []*Recipe
	Get(id string) (*Recipe, error)
}

// MemoryStore is a concurrency-safe in-memory recipe store.
type MemoryStore struct {
	mu       sync.RWMutex
	recipes  map[string]*Recipe
	builtins []string
	users    []string
}

// NewMemoryStore returns an in-memory recipe store seeded with built-ins.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{recipes: make(map[string]*Recipe)}
	for _, builtin := range builtinRecipes() {
		stored := cloneRecipe(&builtin)
		s.recipes[stored.ID] = stored
		s.builtins = append(s.builtins, stored.ID)
	}
	return s
}

// Save stores one user recipe snapshot with a generated recipe ID.
func (s *MemoryStore) Save(r Recipe) (*Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := cloneRecipe(&r)
	stored.ID = deck.NewDeckID()
	stored.ID = "rcp_" + stored.ID[len("deck_"):]
	stored.Source = "user"
	stored.Slide.ID = ""
	s.recipes[stored.ID] = stored
	s.users = append(s.users, stored.ID)
	return cloneRecipe(stored), nil
}

// List returns built-ins first, then user recipes, optionally filtered by tag.
func (s *MemoryStore) List(tag string) []*Recipe {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := append(append([]string{}, s.builtins...), s.users...)
	out := make([]*Recipe, 0, len(ids))
	for _, id := range ids {
		stored, ok := s.recipes[id]
		if !ok || (tag != "" && !slices.Contains(stored.Tags, tag)) {
			continue
		}
		out = append(out, cloneRecipe(stored))
	}
	return out
}

// Get resolves one recipe by stable recipe ID.
func (s *MemoryStore) Get(id string) (*Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stored, ok := s.recipes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneRecipe(stored), nil
}

func cloneRecipe(r *Recipe) *Recipe {
	if r == nil {
		return nil
	}
	return &Recipe{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Source:      r.Source,
		Tags:        append([]string(nil), r.Tags...),
		Slide:       cloneSlide(r.Slide),
	}
}

func cloneSlide(slide contracts.Slide) contracts.Slide {
	data, err := json.Marshal(slide)
	if err != nil {
		return slide
	}
	var cloned contracts.Slide
	if err := json.Unmarshal(data, &cloned); err != nil {
		return slide
	}
	return cloned
}

func builtinRecipes() []Recipe {
	return []Recipe{
		{
			ID:          "rcp_title_cover",
			Name:        "Title Cover",
			Description: "A cover slide with eyebrow, title, and subtitle.",
			Source:      "builtin",
			Tags:        []string{"cover", "hero"},
			Slide: contracts.Slide{Layout: contracts.LayoutCover, Nodes: []contracts.SlideNode{
				&contracts.Hero{Eyebrow: "Deckard Slides", Title: "Presentation Title", Subtitle: "Clear setup for the story ahead"},
			}},
		},
		{
			ID:          "rcp_bulleted_content",
			Name:        "Bulleted Content",
			Description: "A heading with a short checklist or bullet list.",
			Source:      "builtin",
			Tags:        []string{"content", "list"},
			Slide: contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
				&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Key Points"}}},
				&contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{{Text: contracts.RichText{{Text: "First point"}}}, {Text: contracts.RichText{{Text: "Second point"}}}, {Text: contracts.RichText{{Text: "Third point"}}}}},
			}},
		},
		{
			ID:          "rcp_two_column",
			Name:        "Two Column",
			Description: "Balanced two-column comparison for parallel ideas.",
			Source:      "builtin",
			Tags:        []string{"layout", "comparison"},
			Slide: contracts.Slide{Layout: contracts.LayoutTwoColumn, Nodes: []contracts.SlideNode{
				&contracts.TwoColumn{Ratio: contracts.Ratio11, Left: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Left column content"}}}}}, Right: []contracts.SlideNode{&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Right column content"}}}}}},
			}},
		},
		{
			ID:          "rcp_section_break",
			Name:        "Section Break",
			Description: "A section divider slide for pacing transitions.",
			Source:      "builtin",
			Tags:        []string{"section", "transition"},
			Slide: contracts.Slide{Layout: contracts.LayoutFullBleed, Nodes: []contracts.SlideNode{
				&contracts.SectionDivider{Eyebrow: "Section", Label: "Next Chapter"},
			}},
		},
	}
}
