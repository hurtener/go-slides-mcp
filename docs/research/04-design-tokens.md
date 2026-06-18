# 04 — Design Tokens: the "Deckard White" built-in soul

**Status:** research / extraction
**Sources:** the legacy design-system export, gitignored under
`_ref/original-design-system/` (`TOKENS.md`, `design-system-reference.html`,
`ux-documentation-reference.html`). Cross-referenced against the pptx-go `*pptx.Theme` token model
(skill `define-a-theme`; authoritative catalog `docs/design/THEME.md`). The engine map
`docs/research/02-engine-map.md` does not exist yet, so the pptx-go mapping below is done
at the **token-category + concrete-field** level using the live theme taxonomy.

> Naming: the warm editorial theme formerly used as the product's default white theme is,
> for Deckard, the **"Deckard White"** built-in soul. It is **soul #1**, not a hardcode —
> every UI surface still reads `var(--app-*)` and every deck still renders through a
> swappable `*pptx.Theme`. Deckard White is the seed both layers ship with.

---

## 0. Two consumers of one token set

The same semantic tokens feed two render targets. Keep this split in mind throughout:

| Layer | Consumer | Token transport | Notes |
|---|---|---|---|
| **App UI** | The three Svelte `ui://` surfaces (deck-preview, deck-overview, slide-editor) | CSS custom properties `--app-*` | Human review/tweak chrome. Inherits host theme but defaults to Deckard White. |
| **Deck output** | The rendered `.pptx` (slides themselves) | `*pptx.Theme` resolved at apply time | What the agent actually authors. This is the "soul". |

A **soul** is primarily a `*pptx.Theme` (the deck's visual truth). The app surfaces reuse
the same palette/type/shape language so review chrome and the artifact it previews feel like
one product. Tokens below are tagged `[deck]`, `[app]`, or `[both]`.

---

## 1. Aesthetic (the soul's intent)

Warm, editorial, premium-productivity. Cream/beige "paper" backgrounds, a teal-green primary
accent, warm dark-brown near-black text, a serif display face (Playfair Display / Lora) for
hero/titles, and a system sans for body and dense UI. Not a Microsoft-Office clone; tasteful,
neutral, and easy to rebrand. Two type weights only (400 / 500). Everything left-aligned and
sentence-case by default.

---

## 2. Color tokens

### 2.1 Light mode — default `[both]`

| App token (`--app-*`) | Hex | Role / usage |
|---|---|---|
| `--app-bg` / canvas | `#FAF7F2` | App + slide background (warm off-white) |
| `--app-surface` | `#FFFFFF` | Cards, inputs, modals, panels |
| `--app-surface-raised` | `#F4EFE6` | Sidebar, chips, raised/inset, hover-on-canvas |
| `--app-surface-alt` | `#ECE6DC` | Secondary fill, selected rows, hover-on-surface |
| `--app-text` | `#2B2723` | Primary text (warm near-black); 13.87:1 on canvas |
| `--app-text-muted` | `#6A625B` | Secondary text, labels, metadata; 5.60:1 |
| `--app-text-subtle` | `#B8B0A4` | Disabled, placeholder (use only when disabled) |
| `--app-border` | `#E0D5CA` | Default borders (cards, inputs) |
| `--app-border-strong` | `#D8D0C4` | Dividers, section separators |
| `--app-accent` (fill) | `#3B9C94` | **Fills only** — CTA button bg, active-icon bg, send-on |
| `--app-accent-text` | `#2B7A73` | **Text/links/nav only** — AA-safe teal, 4.75:1; the Playfair italic accent word |
| `--app-accent-hover` | `#2B8378` | Primary button hover bg |
| `--app-accent-pressed` | `#2D7A73` | Pressed accent |
| `--app-accent-soft` | `rgba(59,156,148,0.12)` | Teal wash — user bubble, secondary button, selected radio |
| `--app-success` | `#3F8E6B` | Success toast/icon, "uploaded" badge |
| `--app-warning` | `#D97B1A` | Warning toast/icon, active star/featured |
| `--app-danger` | `#B64A4A` | Destructive button, error toast |
| `--app-tooltip` | `#FF9645` | **Tooltips ONLY**, max 2–3 instances; text on it is `#2B2723` (8.1:1) |

> **Critical rule (WCAG AA), carry into the soul:** the teal splits into a **fill token**
> (`#3B9C94`, backgrounds only) and a **text token** (`#2B7A73`, all teal text/links). This is
> exactly how pptx-go separates `ColorAccent` (surface) from `TextAccent` (text) — the mapping
> is 1:1 and must be preserved.

### 2.2 Dark mode — derived `[both]`

| App token | Hex |
|---|---|
| `--app-bg` | `#2B2723` |
| `--app-surface` | `#332E29` |
| `--app-surface-raised` | `#3D372F` |
| `--app-surface-alt` | `#46403A` |
| `--app-text` | `#FAF7F2` |
| `--app-text-muted` | `#D8D0C4` |
| `--app-text-subtle` | `#B8B0A4` |
| `--app-border` | `#46403A` |
| `--app-border-strong` | `#5A5249` |
| `--app-accent` (fill) | `#3B9C94` (brighter teal reads fine on dark) |
| `--app-accent-hover` | `#48B3AA` |
| `--app-accent-soft` | `#33413E` |
| `--app-success` | `#4FA67E` |
| `--app-warning` | `#FF9645` |
| `--app-danger` | `#C86A6A` |

### 2.3 Color → pptx-go theme field mapping `[deck]`

`*pptx.Theme.Colors.Surfaces[ColorRole]` and `.Colors.Text[TextColorRole]`. Deckard White
overrides the neutral/blue `DefaultTheme()` values with the warm palette:

| Soul token | Value | pptx-go field | Default it replaces |
|---|---|---|---|
| canvas | `#FAF7F2` | `Colors.Surfaces[ColorCanvas]` | `FFFFFF` |
| surface (white) | `#FFFFFF` | `Colors.Surfaces[ColorSurface]` | `FFFFFF` |
| surface-alt | `#F4EFE6` | `Colors.Surfaces[ColorSurfaceAlt]` | `F1F3F5` |
| accent fill | `#3B9C94` | `Colors.Surfaces[ColorAccent]` | `2563EB` |
| accent-alt (deep teal, secondary) | `#2B7A73` | `Colors.Surfaces[ColorAccentAlt]` | `7C3AED` |
| accent-warm | `#D97B1A` | `Colors.Surfaces[ColorAccentWarm]` | `EA580C` |
| success | `#3F8E6B` | `Colors.Surfaces[ColorSuccess]` | `16A34A` |
| warning | `#D97B1A` | `Colors.Surfaces[ColorWarning]` | `D97706` |
| error/danger | `#B64A4A` | `Colors.Surfaces[ColorError]` | `DC2626` |
| info (derive teal-tint) | `#2B7A73` | `Colors.Surfaces[ColorInfo]` | `0EA5E9` |
| text primary | `#2B2723` | `Colors.Text[TextPrimary]` | `111827` |
| text secondary | `#6A625B` | `Colors.Text[TextSecondary]` | `374151` |
| text tertiary/muted | `#6A625B` | `Colors.Text[TextTertiary]` | `6B7280` |
| text inverse | `#FAF7F2` | `Colors.Text[TextInverse]` | `FFFFFF` |
| text subtle/disabled | `#B8B0A4` | `Colors.Text[TextMuted]` | `9CA3AF` |
| accent text (AA-safe) | `#2B7A73` | `Colors.Text[TextAccent]` | `2563EB` |
| accent-alt text | `#2B7A73` | `Colors.Text[TextAccentAlt]` | `7C3AED` |
| success text | `#3F8E6B` | `Colors.Text[TextSuccess]` | `16A34A` |
| warning text | `#D97B1A` | `Colors.Text[TextWarning]` | `D97706` |
| error text | `#B64A4A` | `Colors.Text[TextError]` | `DC2626` |

Notes:
- pptx-go has **no border-color or tooltip token**. `--app-border` / `--app-border-strong`
  (`#E0D5CA` / `#D8D0C4`) live in the soul's extended metadata and are applied to slide shape
  outlines via literal stroke color; the tooltip orange (`#FF9645`) is **app-only**, never a
  deck token.
- `--app-accent-soft` (12% teal wash) is produced on the deck with `pptx.TokenColorAlpha(ColorAccent, 12000)`.
- Dark-mode surfaces are an **alternate soul variant**, not separate pptx fields — a second
  cloned `*pptx.Theme` with the §2.2 values swapped into the same roles.

---

## 3. Typography tokens

### 3.1 Faces `[both]`

| Token | Stack | Use |
|---|---|---|
| `--app-font-serif` | `'Playfair Display', 'Lora', Georgia, serif` | Display headings, hero/page titles, the italic accent word |
| `--app-font-sans` | `-apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Segoe UI', Inter, Roboto, Helvetica, Arial, sans-serif` | Body, dense UI, sub-headings |
| `--app-font-mono` | `ui-monospace, SFMono-Regular, Menlo, monospace` | Code, IDs, paths |

### 3.2 Rules (lock into the soul)

- **Two weights only: 400 (regular) and 500 (medium). Never 600/700/900.** This overrides the
  pptx-go default heading weights (which ship 600–700) down to 400/500.
- **Sentence case everywhere.** Never all-caps except the small uppercase eyebrow/section
  labels (10–11px, letter-spacing ~0.08–0.1em).
- **Italic accent rule:** exactly one word per screen/title set in **Playfair Display Italic +
  `#2B7A73`**; the rest of the title is **Lora 400**. Never in subtitles, labels, or body.
- **Left-aligned by default**, all components.

### 3.3 App type scale `[app]`

| Step | Face / weight / size | Use |
|---|---|---|
| Hero | Lora 400 · 42px (+ Playfair italic accent) | Home hero |
| Page title | Lora 400 · 28px (+ accent) | Screen titles |
| Section title | sans 500 · 22px | Section headers |
| Body lg / card name | sans 500 · 15px | Card titles |
| Body | sans 400 · 14px / lh 1.6 | Standard UI/body |
| Label / nav | sans 400 · 13px | Subtitles, field labels, nav items |
| Meta | sans 400 · 12px | Timestamps, counters |
| Eyebrow | sans 500 · 11px UPPERCASE, ls .1em | Section eyebrows |
| Sidebar label | sans 500 · 10px UPPERCASE, ls .07em | Sidebar group labels |

### 3.4 Typography → pptx-go `Theme.Typography[TypeRole]` mapping `[deck]`

`FontSpec = {Family, Size(pt), Weight, Italic}`. Deck sizes are presentation-scale (larger than
the px UI scale) and weights are clamped to the 400/500 rule:

| Soul role | Family | Size (pt) | Weight | pptx-go `TypeRole` | Default it replaces |
|---|---|---|---|---|---|
| Display / hero | Playfair Display (Lora fallback) | 40 | 400 | `TypeDisplay` | Calibri Light 40/700 |
| Title H1 | Lora | 32 | 400 | `TypeH1` | Calibri Light 32/700 |
| Title H2 | Lora | 28 | 400 | `TypeH2` | Calibri Light 28/600 |
| Title H3 | Lora | 24 | 400 | `TypeH3` | Calibri Light 24/600 |
| Subhead H4 | (sans) Inter | 20 | 500 | `TypeH4` | Calibri Light 20/600 |
| Subhead H5 | (sans) Inter | 16 | 500 | `TypeH5` | Calibri Light 16/600 |
| Body | (sans) Inter | 14 | 400 | `TypeBody` | Calibri 14/400 |
| Body small | (sans) Inter | 12 | 400 | `TypeBodySmall` | Calibri 12/400 |
| Caption / eyebrow | (sans) Inter | 10 | 500 | `TypeCaption` | Calibri 10/400 |
| Mono | mono (Consolas/JetBrains Mono) | 13 | 400 | `TypeMono` | Consolas 13/400 |
| Code | mono | 12 | 400 | `TypeCode` | Consolas 12/400 |

Implementation notes:
- `pptx.WithFonts("Playfair Display", "Inter")` sets `HeadingFont`/`BodyFont` and rewrites the
  heading vs body families in one call; then `Clone()` + per-role edits set the **400/500
  weights** (the option does not change weight) and the **Lora** face for H1–H3 (vs Playfair on
  Display).
- **Font embedding decision is open** (D-tbd): Playfair/Lora/Inter are not guaranteed on every
  viewer. The soul should either embed these faces in the `.pptx` or fall back to
  `Georgia, serif` / `Calibri, sans-serif` cleanly. Pure-Go pptx-go font embedding capability
  must be confirmed in the engine map. Until then, serif fallback = Georgia, sans fallback = Calibri.
- The **italic accent word** is not a type role — it is a per-run override
  (`RunStyle{TypeRole: TypeDisplay, Italic: true}` + `TokenTextColor(TextAccent)`) the renderer
  applies to one run inside the title.

---

## 4. Spacing tokens `[both]`

App rhythm (from gaps/padding in the CSS): 6 / 8 / 10 / 12 / 16 / 24 / 32 px; hard rule
**minimum 16px padding** in any container with content.

| Soul step | App value | pptx-go `SpaceRole` | pptx-go default |
|---|---|---|---|
| 2xs | 2px | — (use `SpaceXS`) | — |
| xs | 4px | `SpaceXS` = `Pt(2)` → override `Pt(4)` | `Pt(2)` |
| sm | 8px | `SpaceSM` = `Pt(4)` → override `Pt(8)` | `Pt(4)` |
| md | 12px | `SpaceMD` = `Pt(8)` → override `Pt(12)` | `Pt(8)` |
| lg (min container pad) | 16px | `SpaceLG` = `Pt(16)` | `Pt(16)` ✓ |
| xl | 24px | `SpaceXL` = `Pt(24)` | `Pt(24)` ✓ |
| 2xl | 40px | `Space2XL` = `Pt(40)` | `Pt(40)` ✓ |

The upper steps already match `DefaultTheme()`; only the small end is tightened to the warm,
airy 4/8/12 rhythm.

---

## 5. Radius / shape tokens `[both]`

App radii: micro 4–7px (chips, swatches, attach-x), sm 8px, md 10–12px (cards, inputs,
dropdowns), lg 14–16px (cards, chat input, bubbles), modal 20px, pill/avatar 50%.
Borders are thin and warm: `0.5px`–`1.5px solid var(--app-border)`.

| Soul step | App value | pptx-go `RadiusRole` | pptx-go default |
|---|---|---|---|
| none | 0 | `RadiusNone` = `0` | `0` ✓ |
| sm | 8px | `RadiusSM` = `Pt(2)` → override `Pt(8)` | `Pt(2)` |
| md | 12px | `RadiusMD` = `Pt(6)` → override `Pt(12)` | `Pt(6)` |
| lg | 16px | `RadiusLG` = `Pt(12)` → override `Pt(16)` | `Pt(12)` |
| full / pill | 50% / 20px modal | `RadiusFull` = `Pt(7200)` | `Pt(7200)` ✓ |

> The 4–7px micro radius and the 20px modal radius are **app-chrome only** (no slide shape
> needs them); the deck maps onto the four canonical pptx radii above.

---

## 6. Elevation / shadow tokens

App shadows are warm-tinted (shadow color `#2B2723` / rgba(43,39,35,...)) and soft:

| App component | Shadow |
|---|---|
| Toast | `0 4px 20px rgba(43,39,35,.12)` |
| Dropdown | `0 8px 24px rgba(0,0,0,.10)` |
| Modal | `0 8px 24px rgba(43,39,35,.12)` |
| Primary button hover | `0 4px 12px rgba(59,156,148,.20)` (teal glow) |
| Card hover | `0 0 0 1px rgba(59,156,148,.15), 0 6px 16px rgba(59,156,148,.08)` (teal glow + ring) |

### 6.1 Elevation → pptx-go `Theme.Elevations[ElevationRole]` `[deck]`

`Elevation = {Blur, OffsetX, OffsetY (EMU); Color RGB; Alpha 0–100000}` (px→pt ≈ ×0.75):

| Soul role | Value | pptx-go field | Default it replaces |
|---|---|---|---|
| flat | `{}` no shadow | `Elevations[ElevationFlat]` | `{}` ✓ |
| raised (toast/card) | `{Blur: Pt(12), OffsetY: Pt(3), Color: "2B2723", Alpha: 12000}` | `Elevations[ElevationRaised]` | `{Blur:Pt(4),OffsetY:Pt(1),Color:000000,Alpha:25000}` |
| elevated (modal/dropdown) | `{Blur: Pt(18), OffsetY: Pt(6), Color: "2B2723", Alpha: 15000}` | `Elevations[ElevationElevated]` | `{Blur:Pt(12),OffsetY:Pt(4),Color:000000,Alpha:35000}` |

Deckard White shadows are **warm (`#2B2723`) and lower-alpha** than the cool/black pptx defaults
— softer, more editorial. The teal-glow hover states are **app-only** (interaction feedback, no
slide equivalent).

---

## 7. Component styling tokens `[app]` (Svelte surfaces)

These do not map to pptx-go (they style the human review chrome) but are part of the soul so the
three surfaces feel native. Build from the primitives above.

| Component | Spec |
|---|---|
| **Button — primary** | bg `--app-accent` (`#3B9C94`), text `#FFF`, radius 10–12px, pad 10–12×20–28px, hover bg `#2B8378` + teal-glow shadow, pressed `#2D7A73` inset, disabled bg `#ECE6DC` text `#B8B0A4`. **One primary per screen.** |
| **Button — secondary** | bg `--app-accent-soft`, text `--app-accent-text`, no border, radius 10px |
| **Button — emphasis** | transparent, `1.5px solid #2B7A73`, text `#2B7A73`. Used for "Cancel" in destructive modals. |
| **Button — ghost** | transparent, text `--app-text-muted`, hover bg `rgba(43,39,35,.06)` |
| **Button — danger** | solid `#B64A4A` text `#FFF` (modals) / text-only `#B64A4A` (kebab) |
| **Input** | bg `#FFF`, `1px solid #E0D5CA`, radius 10px, pad 11×14px; focus `border #3B9C94` + `0 0 0 3px rgba(59,156,148,.12)`; error `border #B64A4A` + red ring |
| **Card** | bg `#FFF`, `1px solid #E0D5CA`, radius 16px, pad 20px; hover = teal glow + ring (see §6) |
| **Chat input** | bg `#FFF`, `1px solid #E0D5CA`, radius 16px; divider `0.5px #E0D5CA`; send button 32px circle, on=`#3B9C94`, off=`#ECE6DC` |
| **User bubble** | bg `--app-accent-soft`, text `#2B2723`, radius 14px (br-corner 4px), right-aligned |
| **Agent message** | no bubble, text on canvas, max-width 90%; preceded by italic "reasoning…" badge (`#6A625B` @60% → 100% hover), followed by "was this useful? yes/no" feedback |
| **Toast** | bg `#FFF`, radius 12–14px, raised shadow, width ~300–340px, bottom-right; semantic icon circle (success/warning/error) |
| **Dropdown / kebab** | bg `#FFF`, `0.5px–1px border`, radius 12px, elevated shadow; active item text `#2B7A73` + checkmark; danger item `#B64A4A`; divider `0.5px #E0D5CA` |
| **Modal** | bg `#FAF7F2`, `1px solid #DAD0C3`, radius 20px, elevated shadow, pad 24–28px; darkened overlay; left-aligned title/subtitle |
| **Tooltip** | bg `#FF9645`, text `#2B2723`, radius 7px, appears 300ms hover, rendered in `body` (never clipped) |
| **Badge** | radius 6px, 11px/500; tints at ~10% of teal/neutral/success/error/warning |
| **Toggle** | 36×20px track, off `#E0D5CA` / on `#3B9C94`, 14px white knob |
| **Stepper** | 28px circle bubbles; done/active `#3B9C94` (active +4px teal ring), pending `#ECE6DC`; always visible in multi-step flows |
| **Radio (selected)** | `border #3B9C94`, bg `rgba(59,156,148,.04)`, teal-filled circle |

---

## 8. Layout / motion tokens `[app]`

| Token | Value | Source |
|---|---|---|
| Sidebar width | 240px | nav |
| Reasoning panel width | 320px | chat |
| Split view (chat / artifact) | 55% / 45% | canvas |
| Min touch target | 44px | general rules |
| Min container padding | 16px | general rules |
| Transition | `0.2s ease` | global |
| Tooltip delay | 300ms | tooltips |
| Rename blur grace | 100ms | chat rename |

---

## 9. Assembling the Deckard White `*pptx.Theme` (deck soul)

Recommended construction — `NewTheme` for the faces/accent, then `Clone()` + edits for the warm
palette, the 400/500 weights, and the spacing/radius/elevation overrides:

```go
t := pptx.NewTheme(
    pptx.WithName("Deckard White"),
    pptx.WithAccent(pptx.RGB("3B9C94")),          // fill accent
    pptx.WithFonts("Playfair Display", "Inter"),  // heading / body
).Clone()

// Surfaces (warm paper)
t.Colors.Surfaces[pptx.ColorCanvas]     = pptx.RGB("FAF7F2")
t.Colors.Surfaces[pptx.ColorSurfaceAlt] = pptx.RGB("F4EFE6")
t.Colors.Surfaces[pptx.ColorAccentAlt]  = pptx.RGB("2B7A73")
t.Colors.Surfaces[pptx.ColorAccentWarm] = pptx.RGB("D97B1A")
t.Colors.Surfaces[pptx.ColorSuccess]    = pptx.RGB("3F8E6B")
t.Colors.Surfaces[pptx.ColorWarning]    = pptx.RGB("D97B1A")
t.Colors.Surfaces[pptx.ColorError]      = pptx.RGB("B64A4A")

// Text (warm near-black + AA-safe teal-text)
t.Colors.Text[pptx.TextPrimary]   = pptx.RGB("2B2723")
t.Colors.Text[pptx.TextSecondary] = pptx.RGB("6A625B")
t.Colors.Text[pptx.TextTertiary]  = pptx.RGB("6A625B")
t.Colors.Text[pptx.TextMuted]     = pptx.RGB("B8B0A4")
t.Colors.Text[pptx.TextInverse]   = pptx.RGB("FAF7F2")
t.Colors.Text[pptx.TextAccent]    = pptx.RGB("2B7A73") // NOT 3B9C94 — AA rule

// Type: clamp weights to 400/500, Lora on H1–H3
for _, r := range []pptx.TypeRole{pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3} {
    f := t.Typography[r]; f.Weight = 400
    if r != pptx.TypeDisplay { f.Family = "Lora" }
    t.Typography[r] = f
}
for _, r := range []pptx.TypeRole{pptx.TypeH4, pptx.TypeH5, pptx.TypeCaption} {
    f := t.Typography[r]; f.Weight = 500; t.Typography[r] = f
}

// Spacing (tighten small end), Radii (8/12/16), Elevations (warm shadow)
t.Spacing[pptx.SpaceXS] = pptx.Pt(4); t.Spacing[pptx.SpaceSM] = pptx.Pt(8); t.Spacing[pptx.SpaceMD] = pptx.Pt(12)
t.Radii[pptx.RadiusSM]  = pptx.Pt(8); t.Radii[pptx.RadiusMD]  = pptx.Pt(12); t.Radii[pptx.RadiusLG]  = pptx.Pt(16)
t.Elevations[pptx.ElevationRaised]   = pptx.Elevation{Blur: pptx.Pt(12), OffsetY: pptx.Pt(3), Color: pptx.RGB("2B2723"), Alpha: 12000}
t.Elevations[pptx.ElevationElevated] = pptx.Elevation{Blur: pptx.Pt(18), OffsetY: pptx.Pt(6), Color: pptx.RGB("2B2723"), Alpha: 15000}
```

(API names per the `define-a-theme` skill; verify exact `Elevation`/`Pt` signatures against the
engine map once `02-engine-map.md` lands.)

---

## 10. Soul = bootstrap + refine (how tokens are created/edited)

The locked decision: **one tool seeds a COMPLETE soul** (all tokens) from natural language
and/or a brand template; **targeted overrides refine it.** Map onto the token categories above:

| Stage | What it touches | Mechanism |
|---|---|---|
| **Bootstrap from NL** | All §2–§6 categories at once | Agent maps a prompt ("warm editorial, teal accent, serif titles") to a full token set; produces a complete `*pptx.Theme`. Deckard White is the literal default seed (skip = you get §9). |
| **Bootstrap from brand template** | Colors, fonts, (logo) | `pptx.FromTemplate(brand.pptx)` (skill `load-a-brand-template`) inherits the brand theme, masters, and layouts; remaining unset roles inherit Deckard White. |
| **Refine — token override** | Any single role | A targeted edit (e.g. accent → `#DB2777`) clones the active theme and rewrites one map entry; the theme-swap guarantee re-skins everything authored through tokens. |

This is **easier than the original** because (a) the seed is complete (no role can be left unset
— `Clone()` of the default guarantees it), and (b) refinement is one role at a time, not a
wizard. The original required walking a multi-step form.

---

## 11. UX-documentation insights for the agent-first redesign

### 11.1 What the original got RIGHT (keep)

- **Inline, in-place flows over modals for creation.** "The screen evolves in the same place."
  → deck-overview / deck-preview should mutate in place, not pop wizards.
- **Never native `confirm()`; destructive actions show a custom modal with the *specific
  impact* and an *active choice of destination*** (e.g. "move slides to … (recommended) /
  delete"). Nielsen control-and-freedom; never destroy work without explicit consent.
- **Always-visible stepper** with all steps clickable + checkmarks for multi-step flows (human
  path only — the agent skips it).
- **Toasts for every state change** (created / renamed / deleted / exported), with an action
  link ("view list", "retry") — strong, low-friction feedback.
- **Transparency into the agent:** the italic "reasoning…" badge (click → side panel) +
  "was this useful? yes/no" feedback after every agent answer.
- **Recommended option pre-selected** in destructive choices ("move to private — Recommended").
- **"Improve with AI"** affordance on free-text (prompt) fields.
- **Speak in user terms, never technical;** permissions framed as high-level capabilities.
- **Hard layout rules:** left-align everything; sentence case; min 16px air; 44px touch targets;
  dropdowns/tooltips rendered outside any `overflow:hidden` so they are never clipped.
- **Artifact split-view with 3 modes** (split / fullscreen / closed) and the artifact
  **persists when closed** (a "1 canvas" chip reopens it) — never silently lose generated work.

### 11.2 What HURT (fix in the rework)

- **Defaulting to a fullscreen editor was the core UX failure.** Deckard's default surface is
  `ui://widget/deck-preview` — inline, glanceable thumbnails + sorter + quick actions
  `[download] [edit this]`. Fullscreen `slide-editor` is **opt-in, one slide at a time**.
- **Download / persistence pain in Claude Code.** Export must "just work": write the `.pptx` to
  a deterministic workspace path, return the absolute path, AND expose it as a readable MCP
  resource under `deck://export/<id>.pptx`. No manual download dance.
- **Keyword-triggered artifact generation** (NLP guessing on words like "table/report") is
  fragile — replace with **explicit MCP tools** as the primary authoring path. The agent never
  has to guess intent from prose.
- **Heavy human-wizard authoring** (5-step forms, modals, manual uploads) does not fit
  agent-first. Reframe: the **agent authors the whole deck via tools without opening any UI**;
  the wizard-style affordances (stepper, inline forms, "improve with AI") become *optional human
  review/tweak* chrome, not the only path.

### 11.3 Mapping UX patterns → the three surfaces

| Surface | Inherited UX pattern |
|---|---|
| `ui://widget/deck-preview` (default, inline, glanceable) | Artifact cards + toast feedback + reasoning transparency; **glanceable, never forced fullscreen**; quick actions `[download][edit this]`. |
| `ui://app/deck-overview` (section/slide selector + reorder/structure) | Sidebar navigation + stepper structure + inline (non-modal) reorder; left-aligned lists; destructive reorder/delete uses the impact-modal pattern. |
| `ui://app/slide-editor` (opt-in deep edit of ONE slide) | The old fullscreen artifact mode — but **opt-in only**, scoped to one slide, with a clear "← back to deck" return; rename inline (Enter confirm / Esc cancel / blur-with-grace). |

All three must make the **human↔agent handoff flawless**: every human edit is expressible as the
same token/tool the agent uses, so a human tweak and an agent edit are the same operation under
the hood.

---

## 12. Open items for the engine map (`02-engine-map.md`)

1. Confirm pptx-go **font embedding** capability (Playfair/Lora/Inter) vs. relying on
   Georgia/Calibri fallbacks — affects whether the serif soul renders faithfully off-box.
2. Confirm exact `Elevation` struct field names + `Pt`/`EMU` constructors used above.
3. Confirm where **border/stroke color** for slide shapes is sourced (no border token exists in
   the theme model) — likely a soul-extension field applied as a literal outline.
4. Confirm the soul-bootstrap tool's storage shape (a serialized `*pptx.Theme` + extension
   metadata for border/tooltip/layout tokens that have no native theme field).
