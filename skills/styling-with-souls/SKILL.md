---
name: styling-with-souls
description: "How to set a deck's look with Deckard \"souls\" (themes) - use the built-in Deckard White, bootstrap a soul from a brand description or a brand .pptx, and refine specific tokens. Use when the user wants a specific look or to match a brand."
---

# Styling with souls

A **soul** is a deck's complete visual identity (colors, type, spacing, shape) as
one typed theme. Author with semantic nodes and the soul renders them
consistently — and swapping the soul re-skins the whole deck. Never hand-set
colors per node; change the soul.

## Pick the right starting point

- **Default — do nothing.** A deck with no `soulId` uses **Deckard White** (a
  warm, editorial off-white look). It's a strong default for most decks.
- **List what's available:** `list_souls` → existing souls; `get_soul { soulId }`
  for one; `get_design_tokens` to inspect resolved tokens.

## Match a brand — bootstrap once

`bootstrap_soul` seeds a COMPLETE soul (all tokens) from natural language and/or a
brand file, in one call:

- From a description: `bootstrap_soul { description: "deep navy, warm orange
  accent, geometric sans headings, lots of air" }`.
- From a brand deck: pass the brand `.pptx` (via `upload_asset` → asset id) so the
  soul picks up the real palette/fonts.

Then create or point a deck at it: `create_deck { title, soulId }` (or set it on
an existing deck).

## Refine, don't rebuild

To nudge a soul, change just the tokens you mean: `refine_soul { soulId, … }`
(e.g. a different accent or heading font). Everything else stays. Re-validate
after refining — contrast is checked against the new colors.

## White-label

A deployment can ship its own brand tokens at startup; the UI surfaces re-skin to
the client's brand automatically. As the authoring agent you don't manage this —
just keep authoring with semantic nodes and souls and it all stays coherent.

## Doing it well

- **One soul per deck.** Consistency is the point.
- **Bootstrap from the real brand asset** when you have it — far better than
  guessing hex from a description.
- **Refine sparingly.** A couple of token overrides, not a full re-theme.

## Anti-patterns

- Setting colors on individual nodes to "match the brand" — change the soul instead.
- Creating a new soul per slide or per deck when an existing one fits.
- Refining a soul into low-contrast text — the validator will flag it.

## See also

- `design-principles` — using color/type well. · `building-a-deck` — the loop.
