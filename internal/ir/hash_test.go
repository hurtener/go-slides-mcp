package ir

import (
	"testing"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func sampleSlide() contracts.Slide {
	return contracts.Slide{
		ID:     "s1",
		Layout: contracts.LayoutTitleContent,
		Nodes: []contracts.SlideNode{
			&contracts.Heading{Text: contracts.RichText{{Text: "Hi"}}, Level: 2},
			&contracts.List{Kind: contracts.ListBullet, Items: []contracts.ListItem{
				{Text: contracts.RichText{{Text: "a"}}},
			}},
		},
	}
}

func TestSlideHashDeterministic(t *testing.T) {
	a, err := SlideHash(sampleSlide())
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	b, err := SlideHash(sampleSlide())
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if a != b {
		t.Fatalf("hash not deterministic: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("want 64-hex sha256, got %d chars: %s", len(a), a)
	}
}

func TestSlideHashChangesOnEdit(t *testing.T) {
	base, _ := SlideHash(sampleSlide())

	edited := sampleSlide()
	edited.Nodes[0].(*contracts.Heading).Level = 3
	h2, _ := SlideHash(edited)
	if base == h2 {
		t.Fatal("hash unchanged after editing heading level")
	}

	retitled := sampleSlide()
	retitled.ID = "s2"
	h3, _ := SlideHash(retitled)
	if base == h3 {
		t.Fatal("hash unchanged after editing slide id")
	}
}

func TestDocHashDeterministic(t *testing.T) {
	d := contracts.SlideDoc{Title: "Deck", Slides: []contracts.Slide{sampleSlide()}}
	a, err := DocHash(d)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	b, _ := DocHash(d)
	if a != b {
		t.Fatalf("doc hash not deterministic: %s != %s", a, b)
	}
}
