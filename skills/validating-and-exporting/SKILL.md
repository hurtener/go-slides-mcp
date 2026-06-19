---
name: validating-and-exporting
description: "How to validate a Deckard deck (StyleScore - contrast, overflow, structure) and export it to a downloadable .pptx. Use right before delivering a deck to the user."
---

# Validating & exporting

Before you hand a deck over, check it, then export it. Both are one tool call.

## Validate — `validate_deck_for_export`

`validate_deck_for_export { deckId }` returns:

- `ok` — true when there are **no errors** (warnings don't block).
- `score` — a 0–1 **StyleScore** (weighted: structure, contrast, typography,
  spacing). Higher is better; treat a high score as "ready".
- `findings` — structured issues, each with `category` + `severity`
  (`error` | `warning`) + `message`.
- `perSlide` — per-slide pass/score so you can find the weak slide.

What the checks mean and how to react:

- **structural `error`** — the IR is invalid (e.g. an empty list, a malformed
  node). **Fix it** with the edit tools; the deck won't be right otherwise.
- **contrast `error`/`warning`** — text-on-surface is below WCAG. If you refined a
  soul, adjust the colors (`refine_soul`); on the default soul this is usually
  fine.
- **spacing/overflow `warning`** — a slide's content may overflow (text could wrap
  or shrink in PowerPoint). Split the slide or trim content.

Iterate: fix `error`s, re-validate, and aim for a clean, high score. Warnings are
judgment calls — surface notable ones to the user rather than silently ignoring.

For a single slide while editing, `validate_slide { deckId, slideId }` (or
`validate_slide_ir` for an unsaved snapshot) gives the same shape for one slide.

## Export — `export_deck`

`export_deck { deckId }` ALWAYS:

1. writes a deterministic file to the workspace, AND
2. exposes a readable `deck://export/<id>.pptx` MCP resource.

Hand the user that resource to download — no `include_data` flags, no extra
steps. The render is byte-deterministic, so the same deck always exports the same
file.

## The delivery checklist

1. `validate_deck_for_export` → fix every `error`, note meaningful warnings.
2. `export_deck` → give the user the `deck://` resource.
3. Tell them what they're getting (N slides, the soul) and call out any warnings
   you chose to leave (e.g. "slide 6 is text-dense — say the word and I'll split it").

## Anti-patterns

- Exporting without validating — overflow/contrast issues ship silently.
- Chasing a perfect score by gutting content — a warning is sometimes the right call.
- Re-exporting and not telling the user the download changed.

## See also

- `building-a-deck` — the full loop. · `design-principles` — avoiding contrast/overflow up front.
