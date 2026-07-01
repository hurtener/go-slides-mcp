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
		agendaRecipe(),
		pricingTiersRecipe(),
		featureCardRecipe(),
		comparisonMatrixRecipe(),
	}
}

// agendaRecipe (R14.6) composes a 4-up numbered section index: a card_grid
// of 4 Cards, each Eyebrow "01".."04" with a short Header/Prose sub-line.
// Demonstrates R14.18's extensible-by-composition seam — no new NodeKind.
func agendaRecipe() Recipe {
	return Recipe{
		ID:          "rcp_agenda",
		Name:        "Agenda",
		Description: "A four-part numbered section index for opening a deck.",
		Source:      "builtin",
		Tags:        []string{"agenda", "toc", "index"},
		Slide: contracts.Slide{Layout: contracts.LayoutCardGrid, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Agenda"}}},
			&contracts.Grid{Columns: 4, Gap: contracts.SpaceMD, Cells: []contracts.SlideNode{
				&contracts.Card{Eyebrow: "01", Header: "Context", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Where we are today"}}}},
				}},
				&contracts.Card{Eyebrow: "02", Header: "Strategy", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Where we're headed"}}}},
				}},
				&contracts.Card{Eyebrow: "03", Header: "Roadmap", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "How we get there"}}}},
				}},
				&contracts.Card{Eyebrow: "04", Header: "Ask", Body: []contracts.SlideNode{
					&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "What we need from you"}}}},
				}},
			}},
		}},
	}
}

// pricingTiersRecipe (R14.20 offer-card family, pricing variant) composes 3
// Cards — Eyebrow the plan name, a Stat with Number+Format USD (R14.13) for
// the price, and a feature List. The middle card carries a HeaderFill +
// StatusDot accent to signal "recommended".
func pricingTiersRecipe() Recipe {
	tier := func(eyebrow string, price float64, accent bool, features ...string) *contracts.Card {
		items := make([]contracts.ListItem, 0, len(features))
		for _, f := range features {
			items = append(items, contracts.ListItem{Text: contracts.RichText{{Text: f}}})
		}
		card := &contracts.Card{
			Eyebrow: eyebrow,
			Body: []contracts.SlideNode{
				&contracts.Stat{
					Label:  "per seat / month",
					Number: numberPtr(price),
					Format: &contracts.NumberFormat{CurrencySymbol: "$"},
				},
				&contracts.List{Kind: contracts.ListBullet, Items: items},
			},
		}
		if accent {
			card.HeaderFill = contracts.ColorAccent
			card.StatusDot = contracts.ColorAccent
		}
		return card
	}
	return Recipe{
		ID:          "rcp_pricing_tiers",
		Name:        "Pricing Tiers",
		Description: "Three pricing tiers with the middle plan highlighted as recommended.",
		Source:      "builtin",
		Tags:        []string{"pricing", "offer", "comparison"},
		Slide: contracts.Slide{Layout: contracts.LayoutCardGrid, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Pricing"}}},
			&contracts.Grid{Columns: 3, Gap: contracts.SpaceMD, Cells: []contracts.SlideNode{
				tier("STARTER", 29, false, "5 seats", "Community support", "Core features"),
				tier("GROWTH", 79, true, "25 seats", "Priority support", "Advanced analytics"),
				tier("SCALE", 199, false, "Unlimited seats", "Dedicated support", "Custom SLAs"),
			}},
		}},
	}
}

// featureCardRecipe (R14.20 offer-card family, non-price variant) reuses the
// SAME Card+List composition shape as pricingTiersRecipe but with no Stat —
// Header is a capability name, body a feature List. Proves the family
// renders both a pricing tier and a plain feature card from one shape.
func featureCardRecipe() Recipe {
	feature := func(header string, details ...string) *contracts.Card {
		items := make([]contracts.ListItem, 0, len(details))
		for _, d := range details {
			items = append(items, contracts.ListItem{Text: contracts.RichText{{Text: d}}})
		}
		return &contracts.Card{
			Header: header,
			Body: []contracts.SlideNode{
				&contracts.List{Kind: contracts.ListBullet, Items: items},
			},
		}
	}
	return Recipe{
		ID:          "rcp_feature_card",
		Name:        "Feature Card",
		Description: "Three capability cards, each a header and a feature list — no price.",
		Source:      "builtin",
		Tags:        []string{"feature", "offer", "comparison"},
		Slide: contracts.Slide{Layout: contracts.LayoutCardGrid, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Capabilities"}}},
			&contracts.Grid{Columns: 3, Gap: contracts.SpaceMD, Cells: []contracts.SlideNode{
				feature("Automation", "Scheduled builds", "One-click rollback"),
				feature("Collaboration", "Shared workspaces", "Inline comments"),
				feature("Insight", "Usage dashboards", "Export to CSV"),
			}},
		}},
	}
}

// comparisonMatrixRecipe (R14.3/R14.20) composes a styled comparison Table —
// header band, zebra striping, and a highlighted "recommended" column — from
// the R14.3 TableStyle, 4 rows x 3 cols kept modest for zero-overflow.
func comparisonMatrixRecipe() Recipe {
	row := func(feature, planA, planB string) []contracts.RichText {
		return []contracts.RichText{{{Text: feature}}, {{Text: planA}}, {{Text: planB}}}
	}
	return Recipe{
		ID:          "rcp_comparison_matrix",
		Name:        "Comparison Matrix",
		Description: "A styled comparison table with one highlighted plan column.",
		Source:      "builtin",
		Tags:        []string{"comparison", "matrix", "table"},
		Slide: contracts.Slide{Layout: contracts.LayoutTitleContent, Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Feature Comparison"}}},
			&contracts.Table{
				Headers: []contracts.RichText{{{Text: "Feature"}}, {{Text: "Growth"}}, {{Text: "Scale"}}},
				Rows: [][]contracts.RichText{
					row("Seats", "25", "Unlimited"),
					row("Support", "Priority", "Dedicated"),
					row("Analytics", "Advanced", "Custom"),
					row("SLA", "99.9%", "99.99%"),
				},
				Style: &contracts.TableStyle{HeaderFill: true, Zebra: true, HighlightCol: 2},
			},
		}},
	}
}

// numberPtr returns a pointer to v — the Stat.Number field is a pointer so
// 0 is a real, distinguishable value (D-054 nil-means-unset pattern).
func numberPtr(v float64) *float64 { return &v }
