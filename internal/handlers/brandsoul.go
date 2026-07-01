package handlers

import "github.com/hurtener/go-slides-mcp/internal/soul"

// brandSoulEstablished reports whether soulID names a real brand soul rather
// than the built-in Deckard White default (R8.8). Empty or the default id =>
// not established.
func brandSoulEstablished(soulID string) bool {
	return soulID != "" && soulID != soul.DeckardWhiteID
}

// noBrandSoulNotice is the non-fatal signal appended to a tool result when a
// deck builds/exports on the built-in default soul (R8.8).
const noBrandSoulNotice = "No brand soul established — this deck renders in built-in Deckard White. Run bootstrap_soul (or refine_soul) to set your brand palette, dark variant, gradients, and fonts before building."
