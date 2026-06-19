---
name: design-principles
description: How to make a Deckard deck look genuinely good — hierarchy, typography, color, spacing, and contrast within Deckard's soul/token model. Use alongside composing-a-slide to lift output from correct to polished.
---

# Design principles for Deckard decks

Deckard renders through a **soul** (a typed theme) so most visual choices are
already made consistently. Your job is to make good *structural* choices and let
the soul carry the finish. The principles below are the difference between a deck
that's merely correct and one that looks designed.

## Hierarchy — the most important thing

- **One focal point per slide.** The eye should land on the heading, then the
  single supporting element. If everything is bold, nothing is.
- **Say the takeaway in the heading**, not a topic label. "Latency dropped 38%"
  beats "Performance".
- **Demote support.** Detail goes in `prose`/`list`/`caption`, not in headings.

## Typography

- Deckard White pairs a **serif display** face (titles) with a **system sans**
  (body) and a mono for code. Don't fight it — use `heading`/`hero` for display
  text and `prose`/`list` for body. The type scale is the soul's; trust it.
- **Two weights, sentence case, left-aligned** by default. Avoid ALL CAPS except
  tiny eyebrows/labels.
- Keep headings to one line where you can; long headings read as paragraphs.

## Color

- **Use semantic intent, not literal colors.** A `callout` of kind `tip` or an
  accent `chip` pulls the soul's accent automatically. Never try to hand-set hex
  per node — the soul owns the palette so the deck stays coherent and re-skinnable.
- **One accent does the pointing.** Reserve the accent for the thing you want
  noticed (a key stat, the active step). Overusing it flattens the hierarchy.

## Spacing & rhythm

- **Whitespace is a feature.** A slide with a heading and one chart, lots of air,
  outperforms a dense one. Don't fill space just because it's there.
- **Vary slide types across the deck** — text, chart, two-column compare, card
  grid, quote, section divider. Monotony (ten bullet slides) reads as a wall.

## Contrast & legibility (it's scored)

- Deckard validates **WCAG contrast** against the resolved soul colors. If you
  refine a soul, keep body text well above 4.5:1 on its surface.
- Avoid long text on a busy/low-contrast background. The validator flags it; so
  will the audience.

## Voice & tone

- Concrete and specific. Numbers, verbs, outcomes. Cut hedging and filler.
- Consistent terminology across slides. Title-case the deck title; sentence-case
  everything else.

## Anti-patterns (what NOT to do)

- Rainbow accents / per-node hex. Pick a soul; refine it once if needed.
- Center-aligned walls of text. Left-align; break into points.
- Five fonts. The soul ships exactly the faces you need.
- Cramming — if it doesn't fit, it's two slides.

## See also

- `composing-a-slide` — node choices. · `styling-with-souls` — pick/refine the palette.
- `validating-and-exporting` — the StyleScore checks contrast/overflow for you.
