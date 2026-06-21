---
name: building-a-deck
description: "How to author a complete slide deck with the Deckard MCP server - the create to style to fill to validate to export loop. Use this whenever the user asks to make a presentation, slides, or a PowerPoint."
---

# Building a deck with Deckard

Deckard is an **agent-first** slides server: you build the whole deck by calling
tools — no UI required. The human can review/tweak in the inline surfaces, but
you never need them. Decks render to a real `.pptx` in pure Go.

## The loop

1. **Create the deck.** `create_deck { title, soulId? }` → returns a `deckId`.
   If you don't pass a `soulId`, the deck uses the built-in **Deckard White**
   soul (a warm, editorial look). To match a brand, see `styling-with-souls`.

2. **(Optional) Enable slide chrome.** For a designed, consistent deck, turn on
   the recurring furniture with `set_deck_chrome { deckId, chrome: { enabled: true, brandText: "Acme" } }`.
   This activates a **bottom footer** (brand slot + "N / total" page number) on
   every slide, and a **top section-eyebrow band** on any slide that carries a
   `section` label (see step 3 and `composing-a-slide`). Use `brandAssetId`
   instead of `brandText` to place a logo image from `upload_asset`. Chrome is
   opt-in: omit this step to skip it entirely — the deck is byte-identical to
   before.

3. **Plan the narrative, then add slides.** Decide the story first (cover →
   agenda → sections → close), then `add_slide { deckId, slide }` per slide.
   A slide is `{ layout, nodes: [ … ] }`. Pick the layout for intent
   (`cover`, `title_content`, `two_column`, `card_grid`, `full_bleed`, `blank`,
   `section_divider`) and compose nodes — see `composing-a-slide`. When chrome
   is enabled, set a `section` label on each content slide (e.g.
   `"section": "01 — Direction"`) so the eyebrow band shows the current chapter.

4. **Use the authoring helpers — they return ready-made nodes, so you never
   hand-encode rich text** (each tool's result text includes the node JSON to copy):
   - `compile_markdown { markdown }` → a `nodes` array (headings, lists, quotes,
     prose). **This is the reliable way to build text slides** — pass the returned
     nodes straight into `add_slide`. Prefer it over writing text nodes by hand.
   - `compile_chart { spec }` → a `chart` node (pure-Go image).
   - `compile_code { code, language }` → a `code_block` node.
   Drop the returned node(s) into a slide's `nodes`.

5. **Preview as you go.** `get_deck_preview { deckId }` renders the glanceable
   deck surface for the human. `open_slide_editor` / `get_deck_overview` open the
   editing surfaces — but you can also just keep building by tool.

6. **Validate before export.** `validate_deck_for_export { deckId }` returns a
   0–1 **StyleScore** plus structured findings (structural / contrast / overflow).
   Fix `error`-severity findings; `warning`s are advisory. See `validating-and-exporting`.

7. **Export.** `export_deck { deckId }` always writes a deterministic file AND
   exposes a downloadable `deck://export/<id>.pptx` resource. Hand the human that
   resource — no extra steps.

## Doing it well

- **One idea per slide.** If a slide has two arguments, it's two slides.
- **Lead with the point.** Heading states the takeaway; the body supports it.
- **Vary the rhythm.** Alternate text slides with a chart, a two-column compare,
  a card grid, a quote — don't stack ten bullet lists. A dark `variant` section
  divider between light content slides reads as designed; see `composing-a-slide`
  for the `variant` and `background` syntax.
- **Let it breathe — center sparse slides** via the slide `align` field. A
  cover or a single-stat slide with `{ "align": { "vertical": "center", "horizontal": "center" } }`
  uses whitespace as part of the design. See `composing-a-slide` for the full
  alignment syntax.
- **Let the soul carry the style.** Don't hand-set colors per node; pick/refine a
  soul and author with semantic nodes. Consistency comes for free.
- **Dress your cards.** A bare `card` (header + body only) looks plain. Give cards
  a `fill` + `elevation` + `eyebrow` (+ `headerPill`/`icon`) — see `composing-a-slide`
  → "Make cards look designed." This is the biggest lift from correct to designed.
- **Fill the frame.** Center content vertically (`align:{vertical:center}`) so slides
  aren't top-heavy with an empty bottom.
- **Validate, then export.** A clean StyleScore is the signal the deck is ready.

## Anti-patterns

- Dumping the entire outline onto one slide. Split it.
- Writing paragraphs where a list or a callout would land harder.
- Hand-tuning colors instead of choosing a soul.
- Exporting without validating — overflow and contrast issues ship silently otherwise.

## See also

- `composing-a-slide` — the node vocabulary and when to use each.
- `design-principles` — type, color, spacing, hierarchy choices.
- `styling-with-souls` — match a brand or pick a theme.
- `charts-and-code` — data visuals and code blocks.
- `validating-and-exporting` — the StyleScore and delivery.
