# Domain Map — TypeScript Reference (`_ref/ts-reference`)

> Source of truth for the Go rewrite ("Deckard"). This maps the *behavioral* reference only.
> The reference is the TS reference v4.22-era. Naming note: the existing design system (the
> reference's built-in white theme) becomes the **"Deckard White"** built-in soul in the rewrite.
> All KEEP/DROP/RETHINK classifications reflect the LOCKED PRODUCT DECISIONS.

Reference is read-only at `/Volumes/m2-extended-disk/Repos/go-slides-mcp/_ref/ts-reference/src`.

---

## 0. Architecture at a glance

The reference is **IR-first**: agents author a typed node tree (`SlideIR` / `SectionIR`), never HTML.
The server compiles IR → soul-themed HTML deterministically, stores both IR (source of truth) and
compiled HTML (derived snapshot), validates, and renders/exports.

Two authoring models coexist:
- `slides` (one HTML doc per page, 1920×1080) — **KEEP**.
- `document` (continuous A4/Letter sections, Chromium-paginated) — **DROP COMPLETELY** (A4/long-form).

The rewrite collapses to slides only and replaces HTML/Playwright rendering with native pptx-go.

Layering in the reference:
1. `types/` — entity & wire types (Deck, Slide, Section, DesignSoul, Asset, Comment, metadata).
2. `domain/ir/` — the IR node grammar (Zod schemas) + compiler to HTML.
3. `domain/souls/` — soul token generation, recipes, style guide, lookups.
4. `domain/rendering/` — HTML→PNG (Playwright), pptx/pdf/html exporters, SlideDocument shape inventory.
5. `domain/validation/` — Stage 0 (defensive defaults) / Stage 1 (static lint) / Stage 2 (render-truth).
6. `tools/` — MCP tool surface (decks, souls, assets, comments, export, session, validation, app-only).
7. `storage/` — store interfaces with `memory/` and `file/` implementations.
8. `resources/` + `prompts/` — `legacy://` docs/schema resources, the `ui://` App, onboarding prompts.

---

## 1. Slide IR / Node Model

### 1.1 Top-level containers

**`SlideIR`** (`domain/ir/slide-ir.ts`) — fields:
| Field | Type | Notes |
|---|---|---|
| `layout` | enum `default \| centered \| split` (opt) | compiler hint only |
| `background` | enum `canvas \| surface \| surface_alt \| accent` (opt) | semantic role |
| `background_color` | hex `#RRGGBB` (opt) | escape hatch; wins over `background` |
| `canvas` | `SlideCanvas` (opt) | outer rounded card wrapper |
| `body` | `SlideNode[]` | ordered node list |
| `chrome_override` | enum `inherit \| hide` (opt) | per-slide deck-chrome opt-out |

`SlideCanvas`: `{ background?: hex, padding?: css, radius?: css, shadow?: none\|soft\|medium\|elevated }`.

**`SectionIR`** (document mode — **DROP**): `{ layout?, background?, body: SlideNode[] }`.

### 1.2 RichText grammar (`domain/ir/rich-text.ts`)

`RichText = TextRun[]`. Token references are SEMANTIC, never literal hex.

**`TextRun`**: `{ text: string, bold?, italic?, code?, strike?, sup?, sub?: boolean, link?: url, color?: TextColor }`.
- `bold/italic/code/strike` compose freely; `sup`/`sub` mutually exclusive (sup wins).
- No underline, no highlight, no block elements inside RichText.

**`TextColor`** enum (inline run color role): `accent, accent_alt, accent_warm, success, warning, error, info, muted, inverse`.

### 1.3 Token-reference enums (semantic, resolved by compiler to `var(--...)`)

- **`ColorRole`** (node background): `canvas, surface, surface_alt, accent, accent_alt, accent_warm, success, warning, error, info`.
- **`TextColorRole`**: `primary, secondary, tertiary, inverse`.
- **`TextColor`** (above) — used by `accent` fields on card/chip/arrow/flow/decoration.

### 1.4 Node catalog — the full discriminated union

23 node types (`SLIDE_NODE_TYPES` in `domain/ir/nodes.ts`). Each schema is `.strict()`.

#### Leaf — text
| `type` | Fields |
|---|---|
| `hero` | `title: RichText`, `subtitle?`, `eyebrow?: RichText`, `align?: left\|center` |
| `heading` | `level: 1..6`, `text: RichText`, `align?: left\|center\|right` |
| `prose` | `body: RichText`, `align?: left\|center\|right` |
| `list` | `style: bullet\|numbered\|checklist`, `items: RichText[]` (min 1) |
| `quote` | `body: RichText`, `attribution?: RichText` |

#### Leaf — visual / data
| `type` | Fields |
|---|---|
| `image` | `asset_id: string`, `caption?: RichText`, `alt?: string`, `fit?: contain\|cover`, `frame?: none\|browser\|phone\|desktop\|laptop` |
| `callout` | `kind: note\|warning\|tip\|important`, `title?: RichText (string coerced)`, `body: RichText` |
| `table` | `headers?: RichText[]`, `rows: RichText[][]` (min 1), `caption?: RichText` (row/header length checked at render) |
| `chart` | `chart_type: bar\|stacked_bar\|line\|area\|scatter\|pie\|donut\|histogram\|heatmap\|radar`, `data: (number\|string)[][]`, `series_labels?`, `category_labels?`, `x_axis_title?`, `y_axis_title?`, `caption?`, `show_legend?`, `show_grid?`, `value_format?: number\|percent\|currency\|compact` |
| `divider` | `spacing?: sm\|md\|lg` |
| `code_block` | `code: string (1..4000)`, `language?: /^[a-z][a-z0-9+#-]*$/ (≤16)`, `caption?: RichText` |

#### Inline marks / structure
| `type` | Fields |
|---|---|
| `chip` | `label: RichText`, `accent?: TextColor`, `tone?: tint\|solid\|outline`, `dot?: bool`, `size?: xs\|sm\|md\|lg` |
| `arrow` | `direction?: right\|left\|up\|down`, `style?: solid\|dashed`, `label?: RichText`, `accent?: TextColor` |
| `flow` | `direction: horizontal\|vertical`, `connector: arrow\|arrow_dashed\|cycle\|plus`, `steps: FlowStep[]` (min 2) |
| `decoration` | `source: {asset_ref \| preset}`, `placement: {anchor, offset?, size?, rotation?, opacity?}`, `layer: background\|foreground`, `accent?: TextColor` |

`FlowStep`: `{ label: RichText, accent?: TextColor, icon?: IconName, badge?: string ≤16 }`.
`DecorationSource`: `{kind: asset_ref, asset_id}` or `{kind: preset, name: PresetOrnamentName}`.
`PresetOrnamentName`: `glow_ring, radial_glow, grid_dots, corner_bracket, chevron_arrow, noise_overlay`.
`DecorationAnchor`: 9 in-canvas (`top_left`…`bottom_right`) + 8 `bleed_*` anchors.

#### Card family (presentational wrappers)
| `type` | Fields |
|---|---|
| `card` | `accent?`, `icon?: IconName`, `eyebrow?: RichText`, `header_pill?: CardHeaderPill`, `body: LeafBlockNode[]` (LEAVES only), `body_layout?: column\|row`, `layout?: vertical\|horizontal`, `fill?: none\|tint\|solid`, `border_style?: solid\|dashed\|none`, `size?: compact\|default\|large`, `elevation?: flat\|raised` |
| `card_section` | same chrome as card, but `body: (LeafSlideNode \| grid \| two_column)[]` — accepts containers (cards-of-cards). NOT itself a leaf; cannot nest in card. |

`CardHeaderPill`: `{ label: RichText, accent?, tone?: solid\|tint\|outline, icon?, align?: left\|center\|right }`.

#### Layout containers (NON-recursive — single level)
| `type` | Fields |
|---|---|
| `two_column` | `ratio?: 1:1\|1:2\|2:1`, `gap?: sm\|md\|lg`, `left: LeafSlideNode[]`, `right: LeafSlideNode[]` |
| `grid` | `columns: 2\|3\|4`, `ratio?: "d:d(:d)"`, `gap?`, `align_items?: start\|center\|stretch`, `cells: LeafSlideNode[][]` (cells.length % columns == 0 at render) |

#### Mode-specific (DROP — doc/slide-divider only)
- `section_divider` (slide-only): `{ label?, ornament?: rule\|dot\|none }` — KEEP concept (slide section break), but reachable today via card/hero. RETHINK as a layout.
- `toc`, `bibliography`, `page_break` (doc-only) — **DROP** with document mode.

`LeafBlockNode` union (card body): hero, prose, image, callout, heading, list, divider, quote, table, chart, flow, chip, arrow, code_block.
`LeafSlideNode` union (two_column/grid cells): the above + `card` (cards allowed in cells; no nested grid/two_column).

### 1.5 Icon set

`IconName` — curated ~40-name lucide allowlist (inline SVG, no font dep): shield, lock, check, x, alert-triangle, info-circle, bell, arrow-right, refresh, rocket, zap, play, workflow, bar-chart, trending-up/down, target, gauge, eye, search, database, layers, grid, box, puzzle, network, globe, users, user, briefcase, building, file-text, gem, star, heart, sparkles, lightbulb. **KEEP**; pptx-go has an extendable curated icon set + `WithIconExtension`.

### 1.6 Deck chrome (`domain/ir/chrome.ts`)

`DeckChrome`: `{ header?: ChromeRegion, footer?: ChromeRegion, showOnCover?: bool }`.
`ChromeRegion`: `{ left?, center?, right?: ChromeSlot }` (CSS-grid `1fr auto 1fr`).
`ChromeSlot` (discriminated on `kind`): `logo {asset_id, height?: sm\|md\|lg}` | `text {content: RichText}` | `page_number {format?: 1\|1/N\|01}`.
Per-slide opt-out via `SlideIR.chrome_override = 'hide'`. **KEEP** — render natively via pptx-go slide masters.

### 1.7 Classification — IR model

- **KEEP**: every slide-relevant node (hero, heading, prose, list, quote, image, callout, table, chart, divider, code_block, chip, arrow, flow, decoration, card, card_section, two_column, grid), RichText grammar, semantic token roles, deck chrome, the curated icon set, the non-recursive single-level-nesting policy.
- **RETHINK**: `section_divider` (keep as a slide layout, not a doc construct); `image.frame` device chrome and `decoration` presets must be reproduced natively in pptx-go (the reference rasterized them — see §5).
- **DROP**: `toc`, `bibliography`, `page_break`, `SectionIR`, all `document` authoring-model concepts.

---

## 2. Design Soul / Token System

### 2.1 What a soul is (`types/design-soul.ts`)

A `DesignSoul` is the complete visual identity: a `name`, `slug`, `description`, `status`
(`draft|approved|archived`), the 7-layer `layers` object, and **derived** fields generated on
register/override: `cssTokens` (string), `tokenNames` (~73), `allowedFonts`, `utilityCss`, `styleGuide`.

### 2.2 Token taxonomy — 7 layers (`SoulLayers`)

| Layer | Fields → CSS tokens |
|---|---|
| **color** (`ColorLanguage`) | canvas, surface, surfaceAlt, border, textPrimary/Secondary/Tertiary/Inverse, accentPrimary/Secondary/Warm, success, warning, error, info → `--color-*` |
| **typography** (`Typography`) | fontDisplay/Body/Mono; sizeHero/H1/H2/H3/Body/Label/Caption; weightNormal/Medium/Bold; lineHeightHeading/Body; letterSpacingHeading/Body → `--font-*`, `--text-*`, `--weight-*`, `--line-height-*`, `--letter-spacing-*` |
| **spacing** (`Spacing`) | baseUnit, xs, sm, md, lg, xl, xxl, xxxl, safeAreaInset → `--space-*` |
| **shape** (`ShapeRadius`) | none, sm, md, lg, xl, full, buttonRadius, cardRadius, inputRadius, badgeRadius → `--radius-*` |
| **depth** (`DepthShadow`) | shadowNone/Soft/Medium/Elevated/Inner, borderWidth, borderOpacity → `--shadow-*`, `--border-*` |
| **components** (`ComponentTokens`) | cardPadding/Shadow/BorderWidth, button/input/badge padding etc. → `--card-*`, `--button-*`, `--input-*`, `--badge-*` |
| **motion** (`MotionTone`) | durationFast/Normal/Slow, easingDefault/Emphasized → `--duration-*`, `--easing-*`; PLUS non-CSS design voice: `northStar`, `doRules[]`, `dontRules[]` |

Plus **derived category colors** (`buildCategoryColorTokens`): 8 categorical chart/diagram colors
(`--color-category-a..h` + `-tint`) deterministically rotated/lightened from `accentPrimary`. **KEEP** (map to pptx-go theme palette).

### 2.3 Token generation & cascade (`domain/souls/token-generator.ts`, `soul-service.ts`)

- `generateTokens(layers)` → `{cssString, tokenNames, allowedFonts}`. camelCase→kebab, numeric→px where needed.
- Reference also emits a `[data-legacy-medium="print"]` scoped block (A4 typography) — **DROP** (print).
- **Cascade model**: IR nodes name roles; compiler maps role → `var(--token)`; the SAME IR re-renders
  under any soul. Token change propagates without per-slide rewrites.
- **Override (refine)**: `applyTokenOverride(soulId, layer, tokenName, value)` mutates one field,
  regenerates all derived fields, and if approved regenerates recipes. App-facing via `apply_token_override`,
  which then **cascade-recompiles every IR slide in every deck on that soul**.

### 2.4 Lifecycle & validation against the soul

- `register` → soul created `draft`, tokens generated.
- `approve` → status `approved`, generates layout **recipes** (built-in slide + print recipe families).
- Validation enforces token compliance: literal hex/rgb/hsl in CSS = error; literal spacing px = warning;
  fonts must be in `allowedFonts`. (This was an HTML-lint concern — see §5 RETHINK.)
- `save_as_template` captures a validated IR slide as a reusable `LayoutRecipe` (carries `ir`).

### 2.5 Recipes (`LayoutRecipe`)

`{ id, soulId, type, name, description, tags, source: built-in|user-saved, medium: slides|print, html, ir?, ... }`.
6 built-in slide recipes: title-slide, two-column, metrics, features-grid, closing-cta, blank-themed.
**KEEP** the IR-carrying recipe concept (instantiate via `apply_recipe`); **DROP** print recipes and HTML-only recipes.

### 2.6 Classification — souls

- **KEEP**: 7-layer taxonomy, ~73 semantic tokens, role→token cascade, category colors, draft→approved
  lifecycle, token-override refine, IR-carrying recipes, the `northStar`/`doRules`/`dontRules` design voice.
- **RETHINK (bootstrap + refine — LOCKED)**: the reference required the agent to author ALL 7 layers by hand
  in `register_design_soul`. The rewrite must seed a COMPLETE soul from natural language and/or a brand
  `.pptx` template (logo/colors/fonts via pptx-go `FromTemplate`), then refine via targeted overrides.
  Map soul layers → pptx-go `Theme` tokens (`define-a-theme`). Default built-in soul = **"Deckard White"**.
- **DROP**: print-scoped token block, `medium: print` recipes, `utilityCss`/`styleGuide` as HTML/CSS
  artifacts (regenerate equivalents as agent-facing docs, not CSS).

---

## 3. MCP Tool Surface

64 registered tools. Agent-facing unless marked **APP-ONLY** (`visibility:['app']`) or **MODEL→APP**
(`visibility:['model']`, opens UI). Tools that operate on `document`/`section`/print are **DROP**.

### 3.1 Souls (agent-facing) — KEEP, but reshape per §2.6
| Tool | Purpose / I/O |
|---|---|
| `register_design_soul` | in: `{name, description, layers: SoulLayers}`; out: `{soul_id, name, status, token_count}`. **RETHINK**: replace with NL/template bootstrap. |
| `approve_design_soul` | draft→approved; generates recipes. |
| `list_design_souls` | list souls (status filter). |
| `get_design_soul` | full soul; `include_recipes`, `include_style_guide`. |
| `get_design_tokens` | flat `{tokens:[{name,value,layer}]}` only. |
| `save_as_template` | capture validated slide IR as a recipe. |

### 3.2 Decks & slides (agent-facing) — KEEP (drop section/document siblings)
| Tool | Purpose / I/O |
|---|---|
| `create_deck` | `{soul_id, title?, author?, format?, authoring_model?}` → deck. **DROP** format/authoring_model (slides only). |
| `list_decks` / `get_deck_summary` | deck listing; summary = slides + meta + chrome (no HTML). |
| `add_slide` | `{deck_id, ir: SlideIR, metadata, position?}` → compiles+validates. metadata = title/type/narrative/keyPoints/dataPoints/tags/audience/confidentiality/sources. |
| `update_slide` | replace IR (+sourceKind/metadata/lastValidation). recompiles. |
| `get_slide` / `remove_slide` / `reorder_slides` | self-explanatory. |
| `apply_slide_node_edit` | replace ONE node by structural `path` (e.g. `["body",2,"right",1]`). |
| `apply_slide_field_edit` | edit a single field on a node by path. |
| `apply_slide_text_patch` | targeted text patch. |
| `insert_slide_node` / `remove_slide_node` / `duplicate_slide_node` / `move_slide_node` | structural ops (`slide-structural-ops.tool.ts`). |
| `apply_recipe` | instantiate an IR-carrying recipe as a new slide. |
| `compile_markdown` | markdown → IR leaf nodes; `{nodes,warnings}` or insert via `target`. |
| `compile_chart` | chart spec → `chart` IR node; `{node,warnings,svg_preview}` (RETHINK preview source). |
| `set_deck_chrome` | set deck-level header/footer chrome. |

**DROP (document/section)**: `add_section, update_section, get_section, remove_section, reorder_sections,
list_sections, update_document_meta, apply_section_node_edit, apply_section_field_edit,
apply_section_text_patch, insert/remove/duplicate/move_section_node`.

### 3.3 Assets (agent-facing) — KEEP
| Tool | Purpose |
|---|---|
| `upload_asset` | base64 in → returns `asset://UUID` ref (LLM never re-handles bytes). |
| `list_assets` / `get_asset` | metadata + ref only, never binary. |
| `delete_asset` | delete. |

### 3.4 Comments / collaboration (agent-facing) — KEEP
| Tool | Purpose |
|---|---|
| `list_comments` | between-turn feedback channel; filter resolved/targetKind. |
| `add_comment` | pin on slide / section / **ir_node by irPath** (kind: revision/question/approval/note). |
| `resolve_comment` | resolve with optional note. |

### 3.5 Session (agent-facing) — KEEP
| Tool | Purpose |
|---|---|
| `get_session` | active deck/soul/workflow + `open_panels` + `build_info`. Avoids re-asking "which deck?". |

### 3.6 Validation (agent-facing) — RETHINK (HTML→native)
| Tool | Purpose |
|---|---|
| `validate_slide_ir` / `validate_section_ir` | Zod shape pre-flight, no storage. (KEEP slide; DROP section.) |
| `validate_slide` / `validate_section` | full lint against soul (HTML-based). RETHINK as native IR/pptx checks. |
| `validate_deck_for_export` | pre-export gate. KEEP concept. |

### 3.7 Export (agent-facing) — major RETHINK/DROP
| Tool | Purpose | Verdict |
|---|---|---|
| `export_pptx` | `{deck_id, mode: editable_hybrid\|image, resolution, image_format, jpeg_quality, include_data}`; writes to `config.outputDir`, returns `file_path` + optional base64 blob resource `legacy://exports/<file>`. | **KEEP+RETHINK**: pptx-go native (no hybrid raster); fix delivery (see §4.4). |
| `export_pdf` | slides image/direct + document composer. | RETHINK (slides PDF only) / DROP document path. |
| `export_html` | self-contained HTML. | **DROP** (no HTML render stage). |
| `export_google_slides` | Google API. | DROP (out of scope; revisit later). |
| `render_preview` / `render_section_preview` | base64 thumbnails via Playwright + lint. | **RETHINK** native preview (no Playwright); DROP section variant. |

### 3.8 Resources access (agent-facing) — KEEP
`list_resources` / `get_resource` — wrap `legacy://` resources for tool-only clients.

### 3.9 App-only tools (`visibility:['app']`) — RETHINK around 3 surfaces
| Tool | Purpose |
|---|---|
| `get_editor_state` | full editor state for a deck/slide (thumbnails + selected slide + validation). |
| `apply_text_edit` | plain-text edit to an editable node; uses `expectedRevisionHash`. |
| `apply_block_edit` | structural block edit (document deck: section_kind/break_hints) — **DROP** document parts. |
| `apply_token_override` | live single-token edit + cascade-recompile all slides. |
| `get_thumbnail` | render PNG thumbnail for one slide/section. |
| `upload_asset_from_app` | base64 upload from UI (png/jpeg/svg/webp). |
| `set_active_workspace` | set active deck/soul/workflow in session. |
| `add_comment_from_app` | user-authored comment from UI. |
| `open_deck_editor` (`visibility:['model']`) | the model opens the editor App. |

**RETHINK (LOCKED 3-surface model)**: the reference had ONE fullscreen `ui://deck-editor` App that the
model opened by default — the original UX failure. The rewrite splits into THREE surfaces sharing one Go server:
`ui://widget/deck-preview` (default inline glanceable), `ui://app/deck-overview` (structure/reorder),
`ui://app/slide-editor` (opt-in single-slide deep edit). App tools map onto these; do NOT auto-open fullscreen.

---

## 4. Resources, Storage, Assets, Validation Stages

### 4.1 Resources exposed

- **Schema resources** (`legacy://`): `schema/slide-ir` (JSON Schema for the node grammar — the contract),
  `docs/ir-design-patterns` (composition cookbook), `docs/overview`, `docs/slide-format`, `docs/design-souls`,
  `docs/validation`, `docs/assets`, `docs/css-utilities`, `docs/recipes`, `docs/workflows`,
  `docs/print-mode`, `docs/document-mode`, `docs/charts-and-diagrams`, `docs/collaboration`.
  Registered in a `REGISTRY` map (`resources/registry.ts`) so both `resources/read` RPC and the
  `list_resources`/`get_resource` tools serve identical content. **KEEP** schema + slide docs;
  **DROP** print-mode/document-mode/css-utilities; rename `legacy://` → `deck://`.
- **`ui://` resource**: single `ui://deck-editor/index.html` (embedded Svelte bundle, `prefersBorder`).
  **RETHINK** → three `ui://` surfaces (above), still one embedded Go-served bundle.
- **Rewrite-locked resource**: export artifacts must be readable as MCP resources under `deck://export/<id>.pptx`.

### 4.2 Storage model (`storage/interfaces.ts`)

Interfaces (memory + file impls): `ISoulStore` (+recipes), `IDeckStore` (+revisions), `ISlideStore`,
`ISectionStore` (**DROP**), `IAssetStore` (metadata + data buffer split), `ICommentStore`.
- Entities carry `slug` (stable handle from name/title) resolved via a `SlugIndex` (UUID-or-slug refs).
- **Asset bytes never touched by LLM**: `asset://UUID` ref in IR; resolved to data URI only at render/export.
- Branded ID types (`SoulId`, `DeckId`, `SlideId`, `AssetId`, `CommentId`, `RevisionId`, `TemplateId`).

### 4.3 Revision hashing / optimistic concurrency (`domain/decks/revision-tracker.ts`)

- Every mutation creates an immutable `DeckRevision` `{id, deckId, type, description, slideIdsSnapshot,
  contentHash, createdAt, createdBy?}`. `contentHash = hashSlideContents(htmls)`.
- `DeckMutationType`: deck_created, slide_added/updated/removed, slides_reordered (+ doc variants to DROP).
- Slides/sections carry `metadata.revisionHash`; the App passes `expectedRevisionHash` on edits
  (`ApplyTextEditInput`) for optimistic-concurrency conflict detection. **KEEP** (hash over IR, not HTML).

### 4.4 The export delivery pain (RETHINK — LOCKED)

Reference `export_pptx` writes to `config.outputDir` and *optionally* returns a base64 blob resource only
when `include_data:true`. In Claude Code this caused download/persistence pain. **Rewrite contract**:
export ALWAYS writes to a deterministic workspace path, returns the absolute path, AND exposes it as a
readable MCP resource `deck://export/<id>.pptx`. "Just works" with no `include_data` flag dance.

### 4.5 Validation stages (`domain/validation/`)

- **Stage 0** (defensive defaults): injects hygiene CSS, emits info notes. HTML-specific — **DROP**.
- **Stage 1** (static lint, cheerio+postcss, always runs): `structural-check`, `network-isolation`,
  `safe-area-check`, `token-compliance` (no literal hex), `spacing-compliance` (no literal px),
  `font-compliance` (allowedFonts), `diagram-legibility`. Plus section-* lints (**DROP**).
- **Stage 2** (render-truth, Playwright/Chromium, `depth:'full'` only): `contrast-checker` (WCAG),
  `overflow-detector`, `color-sampler`, `legibility-check`. **DROP Playwright**; reproduce contrast/overflow
  natively against the IR + pptx-go layout where feasible.
- **StyleScore**: weighted 0–1 (token 30%, contrast 25%, typo 15%, spacing 15%, structural 15%);
  −0.20/error, −0.05/warning, clamped. `passed = errorCount===0`. **KEEP** the score concept;
  re-derive checks for an IR-native, hex-free model (most token/font/structural lints become moot because
  IR can't express literal hex except the typed `background_color`/`canvas.background` escape hatches).

### 4.6 SlideDocument shape inventory (`types/slide-document.ts`) — DROP

`SlideDocument` is the HTML→shape-inventory IR the editable-PPTX exporter built by measuring rendered HTML
with Playwright (`CURRENT_COMPILER_REVISION = 19`): elements (text/image/shape/table/group) with absolute
x/y/w/h, per-run styling, `exportDisposition: native|background|blocked`, composite-node PNG snapshots
(cards/decorations/flows rasterized + native text overlays). This entire intermediate layer — and the
"editable_hybrid" raster+overlay PPTX strategy — is **DROPPED**: pptx-go renders the IR natively to real
PPTX shapes (scene IR → renderer), no measure-the-DOM step, no Chromium, no PNG fallback.

---

## 5. Master KEEP / DROP / RETHINK

### KEEP (reproduce ~as-is)
- Full slide IR node grammar (§1.4) + RichText + semantic token roles + curated icons + deck chrome.
- 7-layer soul taxonomy + ~73 tokens + role→token cascade + category colors + draft→approved lifecycle.
- Agent-first deck/slide tools: create_deck, add/update/get/remove/reorder_slide, node/field/text edits,
  structural ops, apply_recipe, compile_markdown, compile_chart, set_deck_chrome, save_as_template.
- Assets (`asset://` ref, bytes hidden from LLM), comments (slide/ir_node pins, between-turn channel),
  session, schema/docs resources, revision history + optimistic concurrency (hash over IR).
- StyleScore concept + pre-flight `validate_slide_ir`.

### RETHINK (keep capability, fix the approach)
- **Soul creation**: NL + brand-`.pptx` bootstrap of a COMPLETE soul; targeted token refine; default
  "Deckard White". (Was: hand-author all 7 layers.)
- **Rendering/export**: native pptx-go (no Playwright, no HTML, no SlideDocument, no editable_hybrid raster).
- **Export delivery**: always-on path + `deck://export/<id>.pptx` resource.
- **Preview/thumbnails**: native render (no Chromium screenshots).
- **Validation Stage 1/2**: IR-native + pptx-layout checks; drop HTML/CSS lints made moot by IR-only authoring.
- **UI surfaces**: three `ui://` surfaces (preview widget default / overview / slide-editor), never auto-fullscreen.
- **`section_divider`**: keep as a slide layout primitive, decoupled from document mode.

### DROP (entirely)
- `document` authoring model + `SectionIR` + all section tools/stores/lints + DocumentComposer.
- Print formats (`print_a4_portrait`, `print_letter_portrait`), print-scoped tokens, print recipes, A4 geometry.
- Doc-only IR nodes: `toc`, `bibliography`, `page_break`.
- Playwright/Chromium everywhere: Stage 2 render-truth, render_preview screenshots, HTML→SlideDocument,
  editable_hybrid PNG fallback, html-exporter, pdf-exporter, playwright-pool.
- `export_html`, `export_google_slides` (and `export_pdf` document path); HTML-only recipes; `utilityCss`/CSS docs.
- The single fullscreen `ui://deck-editor` default-open UX.
- The word the legacy product name — rename scheme `legacy://` → `deck://`, theme → "Deckard White".

---

## 6. Notable file references (for the implementer)
- IR grammar: `src/domain/ir/nodes.ts` (819 lines — the authoritative node catalog), `rich-text.ts`, `slide-ir.ts`, `chrome.ts`.
- Soul: `src/types/design-soul.ts`, `src/domain/souls/token-generator.ts`, `soul-service.ts`.
- Tools: `src/tools/index.ts` (registry), per-tool files under `src/tools/{souls,decks,assets,comments,session,validation,export,app,resources}/`.
- Storage: `src/storage/interfaces.ts`, `src/domain/decks/revision-tracker.ts`.
- Validation: `src/domain/validation/validation-service.ts` + `stage1/*`, `stage2/*`.
- DROP-evidence: `src/types/slide-document.ts` (HTML-shape inventory), `src/domain/rendering/{playwright-pool,html-exporter,pdf-exporter,editable-pptx-exporter}.ts`, `src/types/section.ts`, `src/domain/documents/*`.
</content>
