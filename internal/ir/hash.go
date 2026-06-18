package ir

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// SlideHash returns the lowercase-hex SHA-256 of a slide's canonical JSON. It
// backs optimistic concurrency (expectedRevisionHash). Determinism: the node
// codec emits sorted-key objects and Go marshals struct fields in declaration
// order, so identical IR yields identical bytes — and identical hashes — across
// processes and runs. CONVENTIONS §6.
func SlideHash(s contracts.Slide) (string, error) {
	return hashJSON(s)
}

// DocHash returns the lowercase-hex SHA-256 of a whole deck's canonical JSON.
func DocHash(d contracts.SlideDoc) (string, error) {
	return hashJSON(d)
}

func hashJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("ir: hash marshal: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
