# Phase 13 — Engine richness (pptx-go): make agent-authored decks look designed

> Binding plan. The north star is an **agency-grade investor deck**: centered covers,
> dark full-bleed section slides, rich cards (icon chip + eyebrow + status dot + watermark
> + colored band), composition primitives (a VS badge between two cards, connector arrows
> between columns, a row-labeled bento grid), big-stat / pricing treatments, and consistent
> chrome (section eyebrow + rule, footer logo + `02 / 11` page number). Scope: **full push**.

## Where the work lives (multi-repo flow)

The engine is **`github.com/hurtener/pptx-go`** at `~/Repos/pptx-go` (go-slides-mcp pins it as
a module; `_engine/` is a read-only copy). Each engine unit:
1. Implement + test in `~/Repos/pptx-go` (own golden/determinism tests; byte-identical output
   regardless of worker count — a hard contract).
2. Tag a new pptx-go version; in go-slides-mcp `go get github.com/hurtener/pptx-go@<tag>`.
3. Mirror any layout math into `internal/layout/layout.go` (the canvas snapshot — kept in
   lockstep, it already documents the pin).
4. Expose the new capability in `internal/contracts` (+ `dockyard generate`, fixtures, tests)
   and teach it in `skills/`.

The engine stays an **engine, not a product** (D-026): it renders a typed scene faithfully and
deterministically. "What looks good" is the caller's (Deckard's) opinion, encoded in the soul +
the agent guidance — but the engine must be *capable* of the richness.

## The foundational layout model (the one-way-door API — design once)

Today `scene/render.go:layout()` stacks top-level nodes from `box.Y` downward at fixed
`preferredHeight`s, full body width, no centering. The new model (all additive; **zero values
reproduce today's render**, so it is backward-compatible and determinism is preserved):

- **`SceneSlide.Content Alignment`** — `Alignment{ Vertical VAlign; Horizontal HAlign }`.
  - `VAlign`: `VAlignTop` (zero/default), `VAlignCenter`, `VAlignBottom`, `VAlignJustify`
    (distribute slack into inter-node gaps), `VAlignFill` (grow flexible nodes to fit). Computed
    from the measured total stack height vs the body region — this is what centers a cover and
    stops sparse slides reading thin.
  - `HAlign`: `HAlignLeft` (zero/default), `HAlignCenter`, `HAlignRight` — the default horizontal
    alignment for every block, overridable per node.
- **Per-node override** — leaf nodes that commonly need independent alignment (`Hero`, `Heading`,
  `Prose`, `Quote`, `Chip`, `SectionDivider`, the new `Stat`) carry an optional `Align HAlign`
  (zero = inherit the slide default). Horizontal alignment requires a block **intrinsic width**:
  introduce a deterministic **text-metrics** estimate (chars × per-role font metrics → wrapped
  line count + natural width) — no DOM, pinned constants. This same estimate fixes the
  undercounted heights (Phase 12 E2) and the under-reported overflow (E2/C7).
- **`SceneSlide.Background Background`** — a full-bleed fill behind all content:
  `Background{ Kind: None|Color|Gradient|Asset; Color ColorRole; Gradient [2]ColorRole; AssetID }`.
  Drawn first, full-slide Box, before the body stack. This is what makes the dark/gradient
  section slides possible.
- **Variant** — implement `VariantDark` / `VariantPrint` (today `render.go:67` only warns): a
  variant selects an alternate token resolution (dark surfaces + inverse text) so the SAME IR
  re-renders dark. Pairs with `Background` for full-bleed dark slides.

Determinism rule holds throughout: every new computation is integer-EMU and pinned; the
text-metrics estimate is a pure function of (text, role, width) — no measurement, no RNG.

## Workstreams (sequenced; each is one or more loop units)

**A — Alignment & layout system (the reference unit).** `Alignment` (vertical + horizontal),
the deterministic text-metrics estimate (natural width + wrapped line count), per-node `Align`,
and content-aware `preferredHeight`. Vertical centering/justify/fill of the body stack.
*Accept:* a cover with one `Hero` renders centered both axes; a long paragraph gets ≥ its
wrapped line-count of height (no overlap with the next node); a two-node slide centers instead
of clinging to the top; existing golden renders that used top-left default are unchanged
(zero-value alignment ⇒ byte-identical).

**B — Variants & full-bleed backgrounds.** Implement `VariantDark`/`VariantPrint` token
resolution + `SceneSlide.Background` (color/gradient/asset, full-bleed). *Accept:* a slide with
`Variant: Dark, Background: gradient` renders a dark full-bleed slide with inverse text;
`VariantLight` unchanged.

**C — Slide chrome.** Optional auto-chrome driven by slide metadata: a section eyebrow + hairline
rule (`01 — DIRECTION`), and a footer (logo slot + `NN / TOTAL` page number). Modeled as
render-time chrome bands outside the body region, opt-in via fields on `SceneSlide`/`Scene`.
*Accept:* a deck renders consistent footer page numbers + section eyebrows; opting out yields
today's bare slide.

**D — Rich card chrome.** Finish `Card`'s already-declared fields into the render: icon chip,
eyebrow caps, top-right **status dot**, ghosted **watermark** label/number, colored **header
band** (a `Card.HeaderFill`), footer pill row. *Accept:* a `Card` with icon + eyebrow + status +
watermark + header band renders all of them (matches the reference "Three ways" / "Vision/Mission"
cards).

**E — Composition primitives.** A centered **connector badge** between two columns (the "VS"),
**inter-column connector arrows** (3-column architecture), and a **row-labeled bento grid** (a
Grid with per-row left labels + variable column spans). *Accept:* the "Convert your company"
two-card-with-VS slide and the row-labeled platform grid render.

**F — Stat & pricing treatments.** A **`Stat`** node (big number + label + delta) and a
**pricing-card** composition (header, big price, feature checklist, CTA, a "MOST POPULAR" ribbon
via existing `HeaderPill`/badge). *Accept:* a 4-column pricing row with big-$ stats + a featured
ribbon renders (matches the reference pricing slide).

**G — Deckard integration (per engine unit, not last).** After each engine unit lands + is
tagged: bump the dependency, mirror layout math in `internal/layout`, add the contract fields
(`Alignment`/`Background`/`Variant`/`Stat`/card chrome) with `dockyard generate` + fixtures +
tests, refresh the **Deckard White** soul (add the dark-variant palette), and teach the new
capability in `skills/` (composing-a-slide: "center a cover", "dark section divider", "a stat",
"a pricing row"). This is the "change the model to generate something like this" half — the
engine makes it *possible*, the soul + skills + descriptions make the agent *do it*.

## Sequencing & gate

1. **A (alignment)** — the reference unit; the API above is designed for the whole phase so it
   does not churn. Land + tag + integrate (G) end-to-end on ONE deck before fan-out (§8).
2. **B → C → D → E → F**, each engine-unit then its integration (G).
3. Definition of done for the phase: rebuild a Deckard deck that visually approaches the
   reference — centered cover, a dark full-bleed section slide, a rich card grid, a two-column
   compare with a center badge, a stat/pricing row, consistent footer chrome.

Engine units gate on `~/Repos/pptx-go` green (`go test -race ./...`, golden determinism).
Integration units gate on the full §11 Deckard gate. Builder model: minimax / gpt via the loop,
or **Sonnet via Claude Code** (never Sonnet via OpenRouter) for the high-judgment API units.
