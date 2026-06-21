package deck

import (
	"errors"
	"strings"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func TestMemoryStoreCreateGetAndSlugLookup(t *testing.T) {
	store := NewMemoryStore()
	created, err := store.CreateDeck(CreateDeckInput{Title: "Quarterly Review", Author: "Deckard", SoulID: "soul_1"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	if created.ID == "" || !strings.HasPrefix(created.ID, "deck_") {
		t.Fatalf("CreateDeck ID = %q, want deck_*", created.ID)
	}
	if created.Slug != "quarterly-review" {
		t.Fatalf("CreateDeck slug = %q, want quarterly-review", created.Slug)
	}
	if created.Revision == "" {
		t.Fatal("CreateDeck revision is empty")
	}

	byID, err := store.GetDeck(created.ID)
	if err != nil {
		t.Fatalf("GetDeck by ID: %v", err)
	}
	bySlug, err := store.GetDeck(created.Slug)
	if err != nil {
		t.Fatalf("GetDeck by slug: %v", err)
	}
	if byID.ID != bySlug.ID || byID.Slug != bySlug.Slug {
		t.Fatalf("GetDeck mismatch by ID=%+v by slug=%+v", byID, bySlug)
	}
}

func TestMemoryStoreSlideLifecycleAndRevisionChanges(t *testing.T) {
	store := NewMemoryStore()
	deck, err := store.CreateDeck(CreateDeckInput{Title: "Roadmap"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}

	rev0 := deck.Revision
	deck, added, err := store.AddSlide(deck.ID, testSlide("Intro"), nil)
	if err != nil {
		t.Fatalf("AddSlide: %v", err)
	}
	if added.ID == "" || !strings.HasPrefix(added.ID, "slide_") {
		t.Fatalf("AddSlide ID = %q, want slide_*", added.ID)
	}
	assertRevisionChanged(t, rev0, deck.Revision, "AddSlide")

	gotSlide, err := store.GetSlide(deck.Slug, added.ID)
	if err != nil {
		t.Fatalf("GetSlide: %v", err)
	}
	if headingText(gotSlide) != "Intro" {
		t.Fatalf("GetSlide heading = %q, want Intro", headingText(gotSlide))
	}

	rev1 := deck.Revision
	updatedInput := testSlide("Updated Intro")
	updatedInput.ID = added.ID
	deck, updated, err := store.UpdateSlide(deck.ID, added.ID, updatedInput, rev1)
	if err != nil {
		t.Fatalf("UpdateSlide: %v", err)
	}
	if updated.ID != added.ID {
		t.Fatalf("UpdateSlide ID = %q, want %q", updated.ID, added.ID)
	}
	assertRevisionChanged(t, rev1, deck.Revision, "UpdateSlide")

	rev2 := deck.Revision
	_, second, err := store.AddSlide(deck.ID, testSlide("Second"), nil)
	if err != nil {
		t.Fatalf("AddSlide second: %v", err)
	}
	deck, err = store.ReorderSlides(deck.ID, []string{second.ID, updated.ID})
	if err != nil {
		t.Fatalf("ReorderSlides: %v", err)
	}
	assertRevisionChanged(t, rev2, deck.Revision, "ReorderSlides")
	if deck.Slides[0].ID != second.ID || deck.Slides[1].ID != updated.ID {
		t.Fatalf("ReorderSlides order = [%s %s]", deck.Slides[0].ID, deck.Slides[1].ID)
	}

	rev3 := deck.Revision
	position := 1
	deck, dup, err := store.DuplicateSlide(deck.ID, second.ID, &position)
	if err != nil {
		t.Fatalf("DuplicateSlide: %v", err)
	}
	assertRevisionChanged(t, rev3, deck.Revision, "DuplicateSlide")
	if dup.ID == second.ID {
		t.Fatalf("DuplicateSlide ID = %q, want new ID", dup.ID)
	}

	rev4 := deck.Revision
	deck, err = store.RemoveSlide(deck.ID, dup.ID)
	if err != nil {
		t.Fatalf("RemoveSlide: %v", err)
	}
	assertRevisionChanged(t, rev4, deck.Revision, "RemoveSlide")
	if len(deck.Slides) != 2 {
		t.Fatalf("RemoveSlide len = %d, want 2", len(deck.Slides))
	}
}

func TestMemoryStoreChromeSectionsConflictAndDuplicateDeepCopy(t *testing.T) {
	store := NewMemoryStore()
	deck, err := store.CreateDeck(CreateDeckInput{Title: "Launch"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	deck, added, err := store.AddSlide(deck.ID, testSlide("Original"), nil)
	if err != nil {
		t.Fatalf("AddSlide: %v", err)
	}

	deck, err = store.SetChrome(deck.ID, Chrome{Enabled: true, BrandText: "Deckard", BrandAssetID: "logo-123"})
	if err != nil {
		t.Fatalf("SetChrome: %v", err)
	}
	if !deck.Chrome.Enabled || deck.Chrome.BrandText != "Deckard" || deck.Chrome.BrandAssetID != "logo-123" {
		t.Fatalf("SetChrome got %+v", deck.Chrome)
	}

	deck, err = store.SetSections(deck.Slug, []Section{{Name: "Opening", SlideIDs: []string{added.ID}}})
	if err != nil {
		t.Fatalf("SetSections: %v", err)
	}
	if len(deck.Sections) != 1 || deck.Sections[0].SlideIDs[0] != added.ID {
		t.Fatalf("SetSections got %+v", deck.Sections)
	}

	_, _, err = store.UpdateSlide(deck.ID, added.ID, testSlide("Wrong Revision"), "stale")
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("UpdateSlide conflict err = %v, want ErrRevisionConflict", err)
	}

	_, dup, err := store.DuplicateSlide(deck.ID, added.ID, nil)
	if err != nil {
		t.Fatalf("DuplicateSlide: %v", err)
	}
	dup.Nodes = []contracts.SlideNode{&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: "Mutated Copy"}}}}

	original, err := store.GetSlide(deck.ID, added.ID)
	if err != nil {
		t.Fatalf("GetSlide original: %v", err)
	}
	if headingText(original) != "Original" {
		t.Fatalf("original heading after duplicate mutation = %q, want Original", headingText(original))
	}
	duplicated, err := store.GetSlide(deck.ID, dup.ID)
	if err != nil {
		t.Fatalf("GetSlide duplicate: %v", err)
	}
	if headingText(duplicated) != "Original" {
		t.Fatalf("stored duplicate heading = %q, want Original", headingText(duplicated))
	}
}

func TestMemoryStoreUniqueSlugAndDelete(t *testing.T) {
	store := NewMemoryStore()
	first, err := store.CreateDeck(CreateDeckInput{Title: "Same Title"})
	if err != nil {
		t.Fatalf("CreateDeck first: %v", err)
	}
	second, err := store.CreateDeck(CreateDeckInput{Title: "Same Title"})
	if err != nil {
		t.Fatalf("CreateDeck second: %v", err)
	}
	if first.Slug != "same-title" {
		t.Fatalf("first slug = %q, want same-title", first.Slug)
	}
	if second.Slug != "same-title-2" {
		t.Fatalf("second slug = %q, want same-title-2", second.Slug)
	}
	if err := store.DeleteDeck(second.Slug); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}
	if _, err := store.GetDeck(second.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetDeck after delete err = %v, want ErrNotFound", err)
	}
}

func testSlide(title string) contracts.Slide {
	return contracts.Slide{
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Level: 2, Text: contracts.RichText{{Text: title}}},
			&contracts.Prose{Paragraphs: []contracts.RichText{{{Text: "Body"}}}},
		},
	}
}

func headingText(slide *contracts.Slide) string {
	heading, ok := slide.Nodes[0].(*contracts.Heading)
	if !ok || len(heading.Text) == 0 {
		return ""
	}
	return heading.Text[0].Text
}

func assertRevisionChanged(t *testing.T, before, after, op string) {
	t.Helper()
	if before == after {
		t.Fatalf("%s revision did not change: %q", op, after)
	}
}
