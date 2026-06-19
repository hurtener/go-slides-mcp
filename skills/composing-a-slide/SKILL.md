---
name: composing-a-slide
description: The Deckard slide node vocabulary — what each node is for and how to combine them into a strong slide. Use when deciding how to lay out a slide's content with the Deckard MCP server.
---

# Composing a slide

A slide is `{ layout, nodes: [ … ] }`. The **layout** sets the structural intent;
the **nodes** are the content, stacked top-to-bottom (the renderer owns the exact
geometry — you author meaning, not pixel positions). Build each node as a typed
object with a `kind` discriminator.

## Layouts

`cover` (title slide) · `title_content` (the workhorse) · `two_column` ·
`card_grid` · `full_bleed` · `section_divider` · `blank`.

## Node vocabulary — pick by intent

| Node (`kind`) | Use it for |
|---|---|
| `hero` | The cover: `eyebrow`, `title`, `subtitle`. One per cover slide. |
| `heading` | A slide's headline. `text` (rich), `level` 1–6. State the takeaway. |
| `prose` | Short paragraphs (`paragraphs`). Keep to 1–3; prefer lists for points. |
| `list` | Points. `kind`: `bullet` \| `number` \| `checklist`; `items[].text`. ≤6 items. |
| `callout` | Highlight a fact. `kind`: `note`/`warning`/`tip`/`important`; `title`, `body`. |
| `quote` | A pull-quote. `text`, `attribution`. Great for emphasis or a closer. |
| `chip` | A small tag/label. `label`, `tone`, `color`. |
| `table` | Structured comparison. `headers`, `rows`, `caption`. Keep it small. |
| `two_column` | Side-by-side compare/contrast. `ratio` (1:1/1:2/2:1), `left`, `right`. |
| `grid` | A set of peers (features, segments). `columns` 2–4, `cells`. |
| `card` / `card_section` | A framed unit with a `header` + `body` nodes. |
| `flow` | A process/sequence. `orientation`, `steps[]`, `connector`. |
| `chart` | A data visual — produce with `compile_chart`, embed the returned node. |
| `code_block` | A code snippet — produce with `compile_code`, embed the returned node. |
| `image` | A picture (`assetId` from `upload_asset`), with an optional frame. |
| `divider` / `section_divider` | A rule, or a full-bleed section break. |

## How to make a slide land

- **Heading + one supporting block.** A `heading` that states the point, then ONE
  of: a `list`, a `callout`, a `chart`, a `two_column`, a `table`. Resist stacking
  many blocks — whitespace is part of the design.
- **Compare with `two_column`; enumerate peers with `grid`.** Before/after,
  problem/solution → two columns. Four features → a 2×2 grid of cards.
- **Use `callout` for the number that matters** ("38% lower latency") instead of
  burying it in prose.
- **Sequences are `flow`, not a numbered list**, when the order is the message.
- **Cover = `hero` only.** Let it breathe.

## Anti-patterns

- A slide with 4+ stacked nodes — it will feel cramped and may overflow.
- Lists with >6 items or items that are full sentences — tighten or split.
- A `table` used where a `chart` would communicate the trend faster.
- Putting body text in a `heading`, or a title in `prose`.

## See also

- `building-a-deck` — the overall loop. · `design-principles` — making it tasteful.
- `charts-and-code` — `chart` / `code_block` nodes.
