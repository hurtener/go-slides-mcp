# Phase 12 — Deck fidelity & discoverability hardening

> Binding punch-list. Acceptance criteria here are **binding** (CLAUDE.md §8). Each
> item cites the code that proves the defect. Build in workstream order A→E; run the
> full gate (§11) before every commit. IDs are stable — reference them in commits.

## Why this phase exists

The *El Mate* dogfood deck (built by a real agent over MCP) came out poor: empty card
grids, skeleton lists, dropped flow-step content, stripped bold/italic. A 5-agent
code audit proved the **engine is sound** — `internal/render/scene.go:mapNode` forwards
every field of all 20 node kinds; nothing is lost in the contract→scene mapping. The
failure is a four-part disease at the **MCP↔engine boundary and the surfaces**:

1. **Silent decode loss.** A wrong-shaped node key (`{title,body}` on a flow step, a
   nested `{style:{Italic:true}}` on a run) is dropped by lenient `json.Unmarshal`; the
   node stores empty and the engine faithfully renders nothing.
2. **A validator blind to it.** StyleScore seeds every category at 1.0 and only
   subtracts on an emitted issue; no issue ever inspects whether RichText carries text,
   so 5 empty flow steps scored **1.00 / OK**.
3. **Zero shape discoverability.** The `SlideNode` union is opaque in every generated
   schema (`items: true`), and where RichText *does* surface (speaker notes) the schema
   teaches the *wrong* nested shape. The agent learned every field by trial-and-error
   rejection.
4. **A lossy preview.** `nodeToThumb` flattens containers to a count and never recurses;
   the surface paints empty frames / skeleton bars even when content exists.

The cure: **make wrong input loud, make correct shapes discoverable, make the validator
catch dropped content, and make the preview render real content.** Not "change the
contracts" — the contracts are correct.

---

## Workstream A — Loud decode (stop the loss at the source) · **highest leverage**

**A1 `silent-drop-at-node-decode-boundary` — CRITICAL.** All node/run decode goes
through lenient `json.Unmarshal`; unknown keys vanish. (`internal/contracts/ir_node.go:55`;
`richtext.go:138-169`; `ir_nodes_card.go:100`, grid:35, two_column:43, card_section:31,
ir_slide.go:49.) **Fix:** add a shared strict-decode helper in `internal/contracts`
(`strictUnmarshal(data, v, allowExtra...)`) that map-diffs the object keys against the
target's json-tag set and returns a typed `*UnknownFieldError{Kind,Unknown,Allowed}`.
Apply at `UnmarshalSlideNode` (allow `"kind"`), inside each custom `UnmarshalJSON` raw
struct, and in `TextRun.UnmarshalJSON` (no extras). Flow/List use default decode → give
them a custom `UnmarshalJSON` that strict-decodes each step/item, or recurse the helper
into known slice fields. **Accept:** decoding `{"kind":"flow","steps":[{"title":"a","body":"b"}]}`
errors naming `title`+`body`; a run `{"text":"x","style":{"italic":true}}` errors naming
`style`; all six fixtures + existing `ir_test.go` round-trips stay green.

**A2 `disallowunknownfields-feasible-with-caveats` — HIGH.** Don't flip one global flag.
The injected `"kind"` is not a struct field (List even uses `listKind`), and custom
`UnmarshalJSON` types bypass an outer decoder's strictness — it must live *inside* each
method. (`ir_node.go:67-84`; `ir_nodes_card.go:84-100`; `richtext.go:139-150`.) Use the
A1 helper (or per-method `Decoder.DisallowUnknownFields` with a `Kind json.RawMessage`
ignore-field). **Accept:** strictness regresses no valid decode; a node carrying the
legit injected `"kind"` still decodes clean.

**A3 `surface-unknown-key-error-through-result-text` — HIGH.** The error must reach the
**model**, which reads `Result.Text` (`agenttext.go:1-16`), not `structuredContent`.
`edit_slide_node`/`insert_slide_node` return the decode error as a *bare* Go error
(`edit.go:16-44,101-111`); `add_slide`/`update_slide` decode during SDK input-binding
*before* the handler runs, so the error never reaches `Result.Text`. **Fix:** give
`UnknownFieldError.Error()` a self-describing message (offending key + valid keys + a
one-line correct-shape example, e.g. flow step → `{"label":<RichText>,"detail":<RichText>}`,
run → flat `{"text":"x","bold":true}`). In the edit handlers, `errors.As` it and return
`tool.Result{Text: msg}`. For `add_slide`/`update_slide`, take the slide node(s) as
`json.RawMessage`/`map[string]any` and decode in-handler via the same path so a hint can
be crafted. **Accept:** `edit_slide_node` with a `{title,body}` step returns a result
whose `Result.Text` contains `title` and the `{label,detail}` shape; a test asserts the
key string appears in model-facing content.

**A4 `enum-silent-coercion-to-default` — HIGH.** Bare-`string` enums + a `default:` in
~12 mappers silently coerce `calloutKind:"info"`, `listKind:"ordered"`, `ratio:"60:40"`
to the first value. (`internal/render/enums.go:41-205`; `richtext.go:68-101`.) **Fix:**
add an `IsValid()`/allowed-set check per named enum, invoked at decode/Stage-1 validate,
returning a typed error listing legal values; keep the render `default` as a safety net.
**Accept:** validating a node with `calloutKind:"info"` errors enumerating valid values;
a table-driven test covers every enum type.

> A5 `round-trip-mapnode-complete-decode-is-the-lever` confirms the mapper is field-complete —
> **no mapper change is needed**; A1 is the whole lever. (Recorded so no one "fixes" `mapNode`.)

---

## Workstream B — Discoverability (teach the correct shapes)

**B1 `node-introspection-tool` — HIGH (primary discoverability fix).** Add an MCP tool
`describe_node {kind?}` → per kind returns `{kind, summary, fields:[{name,jsonType,required,
isRichText,note}], example}`, where `example` is a canonical, schema-valid object built
from the real contract structs (cannot drift). Source it from `nodeRegistry`
(`ir_node.go:28-59`). Register before tools; ship 6 fixtures + a contract test (§6).
**Accept:** `describe_node{kind:"flow"}` returns a `label`/`detail` step example;
`describe_node{kind:"callout"}` returns `calloutKind`; no-arg lists all kinds; every
returned example round-trips through `UnmarshalSlideNode` dropping nothing.

**B2 `richtext-schema-teaches-wrong-shape` / `generated-richtext-schema-misleads` — CRITICAL.**
The only place RichText surfaces in a schema (`notes`) shows the reflected struct —
capitalized `Text`/`Style`/`Color` with a **nested `Style`**, required, the exact wrong
shape the agent guessed. (`validate_slide_ir_input.schema.json:24-90`; `richtext.go:105-170`.)
**Fix:** make the flat wire shape authoritative — add a codegen-visible shape (jsonschema
tag / companion type) so `dockyard generate` emits flat lowercase `{text,bold,italic,…}`,
and add a doc example to `RichText`/`TextRun`. **Accept:** no generated `*.schema.json`
contains a nested `Style` or capitalized required `Text`; the notes schema shows flat
`{text,bold,italic}`; `dockyard validate` reports no drift.

**B3 `node-union-opaque-in-schema` — CRITICAL.** Node-carrying inputs are `map[string]any`
/ the `SlideNode` interface → schema is `additionalProperties:true` / `items:true`; all
node doc-comments are dead to the agent. (`edit.go:15`; `insert_slide_node_input.schema.json:18-21`.)
**Fix (cheap, immediate, pairs with B1):** enrich the `Node`/`nodes` field `jsonschema`
tag (as `edit.go` already does for `IRPath`) naming every valid `kind`, the flat run
shape, and flow/list/callout examples, ending with "call `describe_node` for the full
shape." **Accept:** after `dockyard generate`, the `node`/`nodes` description names every
kind + the flat run shape + ≥3 examples; an agent reading only the schema can build a
valid non-empty flow first try.

**B4 skills — HIGH/MEDIUM.** Fix the agent-facing skills (all confirmed wrong/missing):
- `skill-callout-list-wrong-variant-key` (HIGH): composing-a-slide says the variant is
  `kind` — it's `listKind` / `calloutKind` (`kind` is the node discriminator). Fix lines
  24, 26 + add a note distinguishing discriminator from variant.
- `skill-flow-no-step-fields` (HIGH): name the step fields — flat `{label:RichText,
  detail:RichText, icon?}`; warn against `title/body` (line 33 + worked example).
- `skill-no-flat-styled-run-example` (HIGH): show a flat styled run
  `[{"text":"Latency "},{"text":"38% lower","bold":true}]`; state there is **no nested
  `style`** (lines 52-60; mirror in editing-a-slide).
- `skill-table-card-shapes-missing` (MED): table headers/cells are RichText; card
  `header` is a plain **string**, `body` is child nodes (lines 29, 32 + examples).
- `skill-editing-no-node-example` (MED): add one complete `edit_slide_node` payload
  (listKind + items[].text).
**Accept:** each skill names the correct field shapes with a concrete example; re-sync to
`~/Documents/Deckard/skills/`.

---

## Workstream C — Validator fidelity (catch dropped/empty content)

**C1 `stylescore-blind-to-empty-content` — CRITICAL.** `validate.Score` seeds every
category at 1.0 and only subtracts per issue; the only issue sources are Structural /
contrast / overflow — none inspect text. (`internal/validate/score.go:65-97`;
`report.go:26-56`.) **Fix:** add a content-fidelity issue source
(`internal/validate/fidelity.go` → `Fidelity(slide)`/`FidelityDoc(doc)`), append in
`validate.Slide`/`validate.Deck`, under a weighted category so the score drops. **Accept:**
a slide with 5 empty `FlowStep{}` scores `< 1.0` and `Passed == false`.

**C2 `ir-structural-checks-presence-only` — HIGH.** `ir.ValidateNode` checks shape, never
RichText emptiness. (`internal/ir/validate.go:36-89`.) **Fix:** in the Fidelity walk add
`richTextEmpty(rt) := len(rt)==0 || strings.TrimSpace(rt.PlainText())==""` (`PlainText`
exists, `richtext.go:12-21`); emit an issue for every empty content-bearing leaf
(Heading.Text, Prose paragraphs, Quote.Text, Callout.Body, List items, Table headers/cells,
FlowStep.Label), recursing into Card/Grid/TwoColumn/CardSection; set `Path` (e.g.
`nodes[2].steps[3].label`). **Accept:** empty heading/list-item/callout-body/table-cell
each yield a finding referencing the node Path.

**C3 `empty-flow-step-finding` — HIGH.** Per empty step → warning; **all** steps empty →
error (flips OK=false). The error message names the correct flat `{label,detail}` shape
(doubles as discoverability). **Accept:** `Flow{5× FlowStep{}}` → 5 warnings + 1 error,
`Passed==false`, a message contains `label` and `detail`; 1 empty of 4 filled → 1 warning,
`Passed` stays true.

**C4 `empty-card-grid-cell-finding` — HIGH.** `nodeIsEmpty` per kind; empty Card / Grid
cell / column → warning; a wholesale-empty repeating container → error. **Accept:** an
empty Card in a Grid → warning with cell Path; an all-empty Grid → error + `Passed==false`.

**C5 `decoded-to-zero-value-finding` — MEDIUM.** Catch-all: any non-decorative node with
no renderable content → warning "node decoded to an empty value — likely an unrecognized
JSON shape; check field names." Safety net behind C2-C4. **Accept:** a Heading whose JSON
used a wrong nested shape decodes empty and yields a warning; a populated slide yields
none (no false positives).

**C6 `surface-fidelity-before-export` — HIGH.** Plumbing already exists
(`handlers/validate.go:101-115` maps issues; `agenttext.go:11-17` appends findings to
model text). Make the wholesale-empty case Severity error so it lands in `Blockers` and
flips OK in `validate_deck_for_export`. **Accept:** a deck with all-empty flow steps →
`validate_deck_for_export` returns `OK=false`, `Score<1.0`, Blockers with the empty-flow
error, Findings listing each step Path; the model-facing Text includes the findings JSON.

**C7 `overflow-under-reported` — HIGH (depends on E2).** Overflow uses the same optimistic
estimate, so wrapped content that runs off-slide computes `Overflow=false`. (`internal/layout/layout.go:74-92`;
engine `scene/render.go:185`.) **Fix:** tie overflow to the content-aware height from E2 +
a 95%-of-body margin; emit per-node overflow warnings into `validate/OverflowIssues`.
**Accept:** an over-long prose/list slide sets `Overflow=true` and fails validation.

---

## Workstream D — Preview surface (render real content)

**D1 `preview-container-nodes-flattened-to-empty-frames` — CRITICAL.** `nodeToThumb`
reduces Grid→count, Card→header, CardSection→header, and has **no TwoColumn case**;
`ThumbNode` has no nested array; `SlideThumb.svelte` paints empty `<div>`s.
(`internal/handlers/preview.go:114-149`; `internal/contracts/preview.go:22-33`;
`SlideThumb.svelte:84-91`.) **Fix:** add `Children []ThumbNode` to `ThumbNode` (regenerate);
recurse in `nodeToThumb` for Grid/Card/CardSection/TwoColumn; render children recursively
in `SlideThumb.svelte` (`<svelte:self>`). **Accept:** a Grid of Cards (header + body)
shows each card's header **and** body text; a two_column shows its column children —
verified live in `dockyard inspect`.

**D2 `preview-prose-text-ignored` — HIGH.** Server sends prose text
(`preview.go:121-125`) but the surface paints empty `n-line` spans (`SlideThumb.svelte:56-57`).
**Fix:** render `node.text` in the prose branch. **Accept:** a Prose node shows its first
paragraph; skeletons only when truly empty.

**D3 `preview-list-items-no-text` — HIGH.** List sends only count; items never on the
wire. **Fix:** add `Items []string` to `ThumbNode`; populate from `Items[].Text.PlainText()`
(cap ~4); render each item next to its bullet. **Accept:** a List shows real item text.

**D4 `preview-callout-body-dropped` — HIGH.** `Detail` field exists but Callout.Body is
never copied. **Fix:** `t.Detail = v.Body.PlainText()`; render title + body in the surface.
**Accept:** a Callout shows title **and** body snippet.

**D5 `preview-flow-steps-dropped` — HIGH.** Flow sends only step count. **Fix:** carry step
labels via the D3 `Items` array; render label text in each step box. **Accept:** a Flow
shows step labels; doubles as an early warning for the empty-step bug.

**D6 `preview-table-cells-skeleton-only` — LOW.** Optional: carry header-row strings, or
record an ADR that tables are intentionally skeleton-only in the glance.

**D7 `preview-toolbar-buttons-verified-wired` — LOW (verify-only, do NOT rewrite).** The
three buttons already call `bridge.callTool('export_deck'|'get_deck_overview'|'open_slide_editor')`
with correct args (`App.svelte:106-122`). If they "do nothing" in Claude Desktop, the cause
is the host not surfacing the secondary call / not opening another app-surface from an
inline widget — **investigate the host round-trip in `dockyard inspect` first**; only then
consider a handler-side fix. **Accept:** in `dockyard inspect`, clicking each button issues
the expected `tools/call` (visible in the Logbook). If it fires, the surface is correct.

---

## Workstream E — Engine layout & charts (faithful decks still look weak)

> ⚠️ **E2/E3 live in the separate `github.com/hurtener/pptx-go` repo** (`scene/render.go`),
> mirrored in `internal/layout/layout.go`. They require an engine change + release, and the
> owner believes the engine is sound for hand-authored decks — these only bite *naive
> agent stacking*. **Hold E2/E3 for explicit go-ahead.** E1 is in-repo and unblocked.

**E1 `chart-fixed-raster-pillarbox` — HIGH (in-repo).** Charts pin to 1200×720 (AR 1.667)
while the body slot is ~4.74 AR → heavy pillarbox (~185% divergence), and the warning never
reaches the author. (`internal/raster/chart.go:51-58`; `compile.go:13-39`; `handlers/compile.go:63-83`.)
**Fix:** add `Width/Height` (or an `Aspect` enum) to `ChartSpec`, thread into
`RasterizeChart`, default to a wide-short raster (~1600×520) near the slot AR; populate
`CompileChartOutput.Warnings` with the aspect-divergence advisory at author time. **Accept:**
`compile_chart` accepts a dimension/aspect input and warns when AR diverges >15%; an
exported default chart fills most of the slot width, not a centered strip.

**E2 `preferredheight-undercounts-wrapped-text` — HIGH (engine).** Prose=`In(0.4)`/paragraph,
List=`In(0.32)`/item regardless of length → wrapped text overruns into the next node →
"poorly located" overlap. (`internal/layout/layout.go:133-136`; engine `scene/render.go:176-189`.)
**Fix:** content-aware line-count estimate (chars-per-line at the type-role metrics),
deterministic, updated in **both** files in lockstep. **Accept:** an N-line paragraph gets
≥N line-heights in both `layout.Compute` and export; a golden test asserts no overlap.

**E3 `no-vertical-fill-sparse-slides-thin` — MEDIUM (engine).** The stack never centers/
justifies/fills → sparse slides read "too simple." (`scene/render.go:141-189`.) **Fix:** a
layout finishing pass (center / distribute / grow flexible nodes) + an under-fill ratio in
`SlideLayout` the validator can flag. **Accept:** a two-node slide no longer leaves the
bottom 80% blank; `SlideLayout` reports an under-fill ratio.

---

## Build order & gate

1. **A** (loud decode) — foundational; stops loss at source.
2. **B** (discoverability) — agent sends right shapes; B3 jsonschema tag is a 1-line quick win.
3. **C** (validator fidelity) — safety net; C7 waits on E2.
4. **D** (preview surface) — human-facing content.
5. **E1** (chart dims) in-repo; **E2/E3** held for go-ahead (separate engine repo).

Reference-unit-before-fan-out (§8): land **A1 + one strict node** and **C1+C3 (flow)** as
the reference pair, prove the El Mate flow slide now (a) errors on `{title,body}` and (b)
fails validation when empty — then replicate the pattern across the other nodes.

Every commit runs the §11 gate green: `gofmt -l .` empty · `dockyard generate` ×2 idempotent
· `dockyard validate` 0 blockers · `CGO_ENABLED=0 go build ./...` · `go vet` · `go test -race ./...`
· `dockyard test` · `golangci-lint run` · `diff -q AGENTS.md CLAUDE.md` · touched surfaces
`svelte-check` + `vite build`. **Definition of done:** rebuild the El Mate deck end-to-end
over MCP — flow steps land, bold/italic survive, the preview shows card/list/prose/callout
content, and a wrong shape returns a loud, correctable error.
