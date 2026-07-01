package autofit

import (
	"reflect"
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// overflowingUntil returns an OverflowFunc that reports slideID as
// overflowing until pred(doc) is true, then reports a clean render. Lets a
// test drive the ladder deterministically without an engine.
func overflowingUntil(slideID string, pred func(contracts.SlideDoc) bool) OverflowFunc {
	return func(d contracts.SlideDoc) (map[string]bool, error) {
		if pred(d) {
			return nil, nil
		}
		return map[string]bool{slideID: true}, nil
	}
}

// findSlide returns the slide with id from doc, or the zero value.
func findSlide(doc contracts.SlideDoc, id string) contracts.Slide {
	for _, s := range doc.Slides {
		if s.ID == id {
			return s
		}
	}
	return contracts.Slide{}
}

// TestRemediateAppliesLadderRungs asserts a fake OverflowFunc that clears
// once rung 1 (AutoFit) is applied drives exactly one rung, and asserts
// AutoFit landed on the Hero/Heading/Stat nodes.
func TestRemediateAppliesLadderRungs(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Big title"},
			&contracts.Heading{Level: 1, Text: contracts.RichText{{Text: "H"}}},
			&contracts.Stat{Value: "$1,000,000"},
		}},
	}}

	fits := func(d contracts.SlideDoc) bool {
		s := findSlide(d, "s1")
		hero := s.Nodes[0].(*contracts.Hero)
		heading := s.Nodes[1].(*contracts.Heading)
		stat := s.Nodes[2].(*contracts.Stat)
		return hero.AutoFit && heading.AutoFit && stat.AutoFit
	}

	got, rungs, err := Remediate(doc, overflowingUntil("s1", fits))
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if rungs != 1 {
		t.Fatalf("Remediate() rungs = %d, want 1", rungs)
	}
	s := findSlide(got, "s1")
	hero := s.Nodes[0].(*contracts.Hero)
	heading := s.Nodes[1].(*contracts.Heading)
	stat := s.Nodes[2].(*contracts.Stat)
	if !hero.AutoFit || !heading.AutoFit || !stat.AutoFit {
		t.Fatalf("Remediate() did not set AutoFit on all leaves: hero=%v heading=%v stat=%v", hero.AutoFit, heading.AutoFit, stat.AutoFit)
	}
}

// TestRemediateStepsCardSize asserts a fake OverflowFunc that only clears
// once every Card.Size on the slide has stepped down to sm drives 2 rungs
// (rung 1 AutoFit doesn't satisfy the predicate, rung 2 steps the card).
func TestRemediateStepsCardSize(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{
			&contracts.Card{Header: "A", Size: contracts.CardSizeLG},
		}},
	}}

	fits := func(d contracts.SlideDoc) bool {
		s := findSlide(d, "s1")
		card := s.Nodes[0].(*contracts.Card)
		return card.Size == contracts.CardSizeMD
	}

	got, rungs, err := Remediate(doc, overflowingUntil("s1", fits))
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if rungs != 2 {
		t.Fatalf("Remediate() rungs = %d, want 2", rungs)
	}
	s := findSlide(got, "s1")
	card := s.Nodes[0].(*contracts.Card)
	if card.Size != contracts.CardSizeMD {
		t.Fatalf("Remediate() card size = %q, want %q", card.Size, contracts.CardSizeMD)
	}
}

// TestRemediateCleanDocIsUnchanged asserts a doc that never overflows (the
// fake reports clean on the very first call) is returned unchanged, with 0
// rungs applied.
func TestRemediateCleanDocIsUnchanged(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Fine"}}},
	}}
	clean := func(contracts.SlideDoc) (map[string]bool, error) { return nil, nil }

	got, rungs, err := Remediate(doc, clean)
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if rungs != 0 {
		t.Fatalf("Remediate() rungs = %d, want 0", rungs)
	}
	if !reflect.DeepEqual(got, doc) {
		t.Fatalf("Remediate() changed a clean doc:\ngot =%#v\nwant=%#v", got, doc)
	}
}

// TestRemediateDoesNotMutateInput asserts the input doc's slide/nodes are not
// mutated even when the ladder runs to exhaustion (always-overflowing fake).
func TestRemediateDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	hero := &contracts.Hero{Title: "Big"}
	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{hero}},
	}}
	alwaysOverflowing := func(contracts.SlideDoc) (map[string]bool, error) {
		return map[string]bool{"s1": true}, nil
	}

	if _, _, err := Remediate(doc, alwaysOverflowing); err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if hero.AutoFit {
		t.Fatal("Remediate() mutated the input Hero node in place")
	}
	if doc.Slides[0].Nodes[0] != contracts.SlideNode(hero) {
		t.Fatal("Remediate() replaced a node pointer inside the input doc's slice")
	}
}

// TestRemediateIsDeterministic asserts two Remediate calls over the same
// input with the same fake overflow function produce equal docs.
func TestRemediateIsDeterministic(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{
			&contracts.Hero{Title: "Big"},
			&contracts.Card{Header: "A", Size: contracts.CardSizeLG},
		}},
	}}
	alwaysOverflowing := func(contracts.SlideDoc) (map[string]bool, error) {
		return map[string]bool{"s1": true}, nil
	}

	got1, rungs1, err := Remediate(doc, alwaysOverflowing)
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	got2, rungs2, err := Remediate(doc, alwaysOverflowing)
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if rungs1 != rungs2 {
		t.Fatalf("Remediate() rungs differ: %d vs %d", rungs1, rungs2)
	}
	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("Remediate() is not deterministic:\nfirst =%#v\nsecond=%#v", got1, got2)
	}
}

// TestRemediateCapsAtLadderLength asserts a fake that ALWAYS reports overflow
// stops after exactly len(ladderRungs) rungs — the ladder never loops forever.
func TestRemediateCapsAtLadderLength(t *testing.T) {
	t.Parallel()

	doc := contracts.SlideDoc{Slides: []contracts.Slide{
		{ID: "s1", Nodes: []contracts.SlideNode{&contracts.Hero{Title: "Never fits"}}},
	}}
	calls := 0
	alwaysOverflowing := func(contracts.SlideDoc) (map[string]bool, error) {
		calls++
		return map[string]bool{"s1": true}, nil
	}

	_, rungs, err := Remediate(doc, alwaysOverflowing)
	if err != nil {
		t.Fatalf("Remediate() error = %v", err)
	}
	if rungs != len(ladderRungs) {
		t.Fatalf("Remediate() rungs = %d, want %d (ladder exhausted)", rungs, len(ladderRungs))
	}
	if calls != len(ladderRungs) {
		t.Fatalf("Remediate() called overflowing %d times, want %d (one probe per rung, no trailing re-check)", calls, len(ladderRungs))
	}
}
