---
name: composing-a-slide
description: "The Deckard slide node vocabulary - what each node is for and how to combine them into a strong slide. Use when deciding how to lay out a slide's content with the Deckard MCP server."
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

## Make cards look designed (don't ship bare cards)

A `card` with only `header` + `body` renders as a thin accent bar — functional but
plain. The card already supports a full design vocabulary; **use it**. These fields
are what turn a card grid from "a list in boxes" into something that looks built:

- `fill` — a surface color role (`"surface"`, `"surfaceAlt"`, `"accent"`, …). Gives the
  card a filled background instead of a bare rule. `"surfaceAlt"` reads as a clean
  panel on a light slide; `"accent"` makes one card pop.
- `elevation` — `"raised"` or `"elevated"` adds a soft drop shadow so the card lifts
  off the page.
- `eyebrow` — a small kicker label above the header (e.g. `"FOUNDATION"`).
- `headerPill` — a short badge to the right of the header (e.g. `"01"`, `"NEW"`).
- `icon` — a curated icon name; `layout: "iconTop"` stacks it above the header.

```json
{ "kind": "card", "fill": "surfaceAlt", "elevation": "raised",
  "eyebrow": "FOUNDATION", "header": "Contract-first", "headerPill": "P3",
  "body": [ { "kind": "prose", "paragraphs": [[ { "text": "One clear sentence." } ]] } ] }
```

**Default to filled + raised + eyebrow** for any card grid or two-column-of-cards —
it is the single biggest lift from "correct" to "designed."

## Fill the frame — don't leave content top-heavy

By default the body stacks from the top, so a slide with a heading + one block leaves
the bottom half empty and reads thin. Unless a slide is genuinely meant to hug the top,
**set the slide `align` to center the content vertically**:

```json
{ "layout": "card_grid", "align": { "vertical": "center" }, "nodes": [ … ] }
```

For a heading-plus-content slide, `"justify"` pushes the heading toward the top and the
block toward the bottom, using the whole frame. Centered or justified beats top-heavy.

## Getting the node encoding right

- **Rich-text fields** (a heading's `text`, a callout `body`, a quote `text`, a
  list item's `text`) are JSON arrays of runs: `[{"text":"…"}]` — not a bare
  string and not `{"item":…}`. **Plain string fields** (a hero's `title`/`eyebrow`/
  `subtitle`, a chip `label`) are just strings.
- **Easiest path: don't hand-encode.** Build text via `compile_markdown` (it
  returns correctly-shaped nodes) and only hand-write the simple nodes (hero,
  chip, callout). To see the exact shape of any node, read a `get_slide` result —
  its text includes the slide's real IR you can copy.

## Aligning slide content

Two independent alignment axes let you control how the body stack sits in the
slide frame. Both are optional; omitting them leaves the pre-existing top-left
layout unchanged.

**Slide-level alignment** — set `align` on the slide object (an object with
`vertical` and/or `horizontal` keys):

```json
{ "align": { "vertical": "center", "horizontal": "center" } }
```

- `vertical`: `"top"` (default) · `"center"` · `"bottom"` · `"justify"` (spreads inter-node gaps).
- `horizontal`: `"left"` (default) · `"center"` · `"right"` (narrows each leaf node to its natural text width).

**Per-node alignment** — set `align` on individual `hero`, `heading`,
`prose`, `quote`, `chip`, or `section_divider` nodes (a string, not an
object):

```json
{ "kind": "prose", "align": "right", "paragraphs": [...] }
```

The per-node `align` overrides the slide's `horizontal` for that block only.
Containers (`two_column`, `grid`, `card`, `table`, `flow`, …) always span the
full body width and are not affected by alignment.

**Center a cover** — a single `hero` node on a slide with vertical + horizontal
centering:

```json
{
  "layout": "cover",
  "align": { "vertical": "center", "horizontal": "center" },
  "nodes": [{ "kind": "hero", "title": "Q3 Review", "eyebrow": "All hands" }]
}
```

**Right-align a caption** — override alignment on a single `prose` block while
the rest of the slide defaults to left:

```json
{ "kind": "prose", "align": "right", "paragraphs": [[{ "text": "Source: internal data" }]] }
```

Note: slide `align` is an object (`{ "vertical": "…", "horizontal": "…" }`);
node `align` is a string (`"left"` | `"center"` | `"right"`).

## Dark & full-bleed slides

Two optional slide-level fields let you flip a slide to dark or paint a
full-bleed background behind the content:

**Dark section slide** — set `"variant": "dark"` on the slide. The engine
derives a legible dark palette (dark canvas, light text) from the active soul
automatically. Accent and semantic colors are preserved, so brand identity
survives. Use for section-divider slides to create a visual rhythm break
between content sections.

```json
{
  "layout": "title_content",
  "variant": "dark",
  "nodes": [{ "kind": "hero", "title": "Part II — Execution" }]
}
```

**Full-bleed gradient** — set `"background"` with `"kind": "gradient"`, a
`"gradient"` array of two soul color roles, and an `"angle"` in degrees. The
gradient is drawn behind all nodes (lowest layer, z-order 0). Color roles
resolve through the active soul, so a theme swap re-paints the background.

```json
{
  "layout": "title_content",
  "variant": "dark",
  "background": {
    "kind": "gradient",
    "gradient": ["accent", "accentAlt"],
    "angle": 135
  },
  "nodes": [{ "kind": "hero", "title": "Brand Section" }]
}
```

**Solid-color background** — set `"kind": "color"` and a single `"color"` role:

```json
{ "background": { "kind": "color", "color": "canvas" } }
```

**Full-bleed asset background** — set `"kind": "asset"` and an `"assetId"`
from `upload_asset`. The image fills the slide canvas; content nodes render on
top:

```json
{ "background": { "kind": "asset", "assetId": "brand-photo" } }
```

Background and variant are independent — you can have a dark variant without a
background (the engine fills the canvas automatically) or a light variant with
a background. Omitting both leaves the slide unchanged from before (backward
compatible, byte-identical output).

## Anti-patterns

- A slide with 4+ stacked nodes — it will feel cramped and may overflow.
- Lists with >6 items or items that are full sentences — tighten or split.
- A `table` used where a `chart` would communicate the trend faster.
- Putting body text in a `heading`, or a title in `prose`.

## See also

- `building-a-deck` — the overall loop. · `design-principles` — making it tasteful.
- `charts-and-code` — `chart` / `code_block` nodes.
