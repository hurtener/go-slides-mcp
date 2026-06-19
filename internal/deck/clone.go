package deck

import (
	"encoding/json"
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

func cloneDeck(d *Deck) (*Deck, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("deck: clone deck marshal: %w", err)
	}
	var out Deck
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("deck: clone deck unmarshal: %w", err)
	}
	return &out, nil
}

func cloneSlide(s contracts.Slide) (*contracts.Slide, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("deck: clone slide marshal: %w", err)
	}
	var out contracts.Slide
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("deck: clone slide unmarshal: %w", err)
	}
	return &out, nil
}

func cloneSections(sections []Section) []Section {
	if len(sections) == 0 {
		return nil
	}
	out := make([]Section, len(sections))
	for i, section := range sections {
		out[i] = Section{Name: section.Name}
		if len(section.SlideIDs) > 0 {
			out[i].SlideIDs = append([]string(nil), section.SlideIDs...)
		}
	}
	return out
}
