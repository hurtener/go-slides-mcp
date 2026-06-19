package deck

import (
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
	"github.com/hurtener/go-slides-mcp/internal/ir"
)

// computeRevision returns the deck's current slide-document content hash.
func computeRevision(d *Deck) (string, error) {
	rev, err := ir.DocHash(contracts.SlideDoc{
		Title:  d.Title,
		Slides: d.Slides,
	})
	if err != nil {
		return "", fmt.Errorf("deck: compute revision: %w", err)
	}
	return rev, nil
}
