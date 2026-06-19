---
name: editing-a-slide
description: "How to refine an existing slide in a Deckard deck - change text, fix a field, reorder or replace a node - using small, targeted edits instead of resending the whole slide. Use whenever you're tweaking a slide that already exists (e.g. fixing overflow or a typo)."
---

# Editing an existing slide

**Golden rule: for a small change, make a small edit.** Don't reach for
`update_slide` to fix one line — `update_slide` replaces the ENTIRE slide, so you
have to resend every node correctly (a large, slow, error-prone argument). Use it
only to swap a whole slide wholesale. For everything else, use the targeted tools
below.

## Step 1 — see the slide and find the path

`get_slide { deckId, slideId }` returns the slide's IR (the result text includes
it). Read it to locate the node you want and work out its **path**.

A path is an array of legs from the slide root:

- The **first leg is always `"nodes"`**, followed by a 0-based node index.
- Nested container legs: `"left"` / `"right"` (two_column), `"cells"` (grid),
  `"body"` (card / card_section).
- Examples: `["nodes", 2]` (the 3rd top-level node) · `["nodes", 1, "left", 0]`
  (first node in a two_column's left column) · `["nodes", 0, "cells", 3]`.
- **List items and prose paragraphs are NOT legs** — they're fields of their node.
  To change them, replace the whole list / prose node (see below).

## Step 2 — make the targeted edit

- **Rich-text field** (a heading's `text`, a callout `body`, a quote `text`):
  `patch_slide_text { deckId, slideId, path, field, text }` — sets it to plain text.
- **Plain string field** (a hero `title`/`eyebrow`/`subtitle`, a callout `title`,
  a chip `label`): `edit_slide_field { deckId, slideId, path, field, value }`.
- **A list's items, a prose's paragraphs, or any structural change to one node:**
  `edit_slide_node { deckId, slideId, path, node }` — `node` is the full replacement
  node object (with its `kind`). E.g. to shorten a bullet, send the whole list node
  with the edited `items`.
- **Reorder / add / remove / copy nodes:** `move_slide_node { from, to }`,
  `insert_slide_node { path, node }`, `remove_slide_node { path }`,
  `duplicate_slide_node { path }`.

Each returns the updated slide + its validation, so you can confirm the change.

## Common task: fixing overflow

`validate_*` flagged a slide as overflowing? Don't rebuild it — shorten in place:
`get_slide` → for each long line, `patch_slide_text` (or `edit_slide_node` for a
list) with tighter copy → re-validate.

## Anti-patterns

- `update_slide` for a one-line fix. It makes you resend the whole slide.
- Guessing the path — read `get_slide` first and count the legs.
- Trying to path into a list item (`…, "items", 0`) — that won't resolve; edit the
  whole list node instead.

## See also

- `composing-a-slide` — the node vocabulary. · `validating-and-exporting` — the checks.
