# Deckard — Phased Implementation Plan

> Product: **Deckard** (UI title **Deckard Slides**). Repo `go-slides-mcp`.
> Module `github.com/hurtener/go-slides-mcp`. Full Go rewrite on Dockyard, single
> CGo-free static binary. Agent-first; three `ui://` surfaces; export = path +
> `deck://` resource; souls = bootstrap + refine; slides-only (A4 dropped).
>
> This plan is written to be executed largely by a cheap autonomous **builder**
> model that clones an established reference unit. Every phase: a correct
> **reference unit** lands first (orchestrator-owned), then the builder fans out
> N near-identical units against a green machine-checkable gate. Ownership is
> tagged **[ORCH]** (high-judgment: novel UI, contracts, the pptx-go adapter) or
> **[BUILD]** (clonable: storage, per-node renderers, per-tool handlers).

Sources this plan is derived from (binding): `docs/research/01-domain-map.md`
(KEEP/DROP/RETHINK), `docs/research/02-engine-map.md` (pptx-go surface + gaps),
`docs/research/03-dockyard-conventions.md` (layout, contract-first, gates),
`docs/research/04-design-tokens.md` (Deckard White soul).

---

## 1. Target architecture

### 1.1 Project layout (single module, single binary, three surfaces)

```text
go-slides-mcp/
├── CLAUDE.md  AGENTS.md            # binding normatives (verbatim mirror)
├── Makefile  .golangci.yml  go.mod  go.sum
├── dockyard.app.yaml               # ONE manifest: apps[]×3 + tools[] + quality{}
├── main.go                         # server.New → RegisterAll(apps) → RegisterResources(deck://) → RegisterTools → serve
│
├── internal/
│   ├── contracts/                  # CONTRACT-FIRST source of truth (Design A)
│   │   ├── contracts.go            #   typed In/Out structs for every tool
│   │   ├── ir.go                   #   the slide IR node grammar (Go types, sealed-ish union)
│   │   ├── *.schema.json *.ts      #   GENERATED — never hand-edit
│   ├── ir/                         # IR validation/normalization, structural-path ops, hashing
│   ├── soul/                       # soul model + Deckard White built-in + bootstrap/refine → *pptx.Theme
│   ├── render/                     # pptx-go adapter: IR → scene.Scene → .pptx bytes ([ORCH] core, [BUILD] per-node)
│   ├── raster/                     # pure-Go chart + code rasterizers (AssetResolver feed)
│   ├── deck/                       # Deck/Slide/Section store, branded IDs, revisions, slug index
│   ├── exportstore/                # deterministic workspace paths + deck:// resource handler
│   ├── handlers/                   # one file per tool group; ToolDeps DI; RegisterTools(srv, deps)
│   ├── apps/                       # RegisterAll for the three surfaces
│   ├── validate/                   # native IR + layout validation, StyleScore
│   └── fixtures/                   # shared sample decks (incl. a high-complexity "ceiling" deck)
│
├── web/
│   ├── design-system/              # Deckard White tokens (--app-*) + shared Svelte primitives
│   └── apps/
│       ├── deck-preview/           # ui://widget/deck-preview   (inline, DEFAULT, glanceable)
│       ├── deck-overview/          # ui://app/deck-overview     (structure / reorder)
│       └── slide-editor/           # ui://app/slide-editor      (opt-in single-slide deep edit)
│
├── fixtures/<tool>/{happy,empty,error,permission,slow,large}.json
└── docs/{contracts/CONVENTIONS.md, ui/SURFACES.md, decisions/, research/, glossary.md}
```

### 1.2 Package decomposition (responsibility per package)

| Package | Owns | Depends on | Ownership |
|---|---|---|---|
| `internal/contracts` | typed tool In/Out + IR node grammar; **single source of truth** | — | **[ORCH]** |
| `internal/ir` | IR normalize, `.strict` shape checks, structural-path edits, content hash | contracts | **[ORCH]** seed, **[BUILD]** per-node |
| `internal/soul` | soul persistence, Deckard White default, NL/template bootstrap, token-override refine → `*pptx.Theme` | contracts, pptx-go | **[ORCH]** |
| `internal/render` | IR → `scene.Scene` mapping + `scene.Render` driver; per-node mappers | ir, soul, pptx-go, raster | **[ORCH]** core, **[BUILD]** mappers |
| `internal/raster` | pure-Go chart rasterizer + code-highlight rasterizer (PNG/SVG bytes) | stdlib only | **[ORCH]** |
| `internal/deck` | Deck/Slide/Section store (memory + file), branded IDs, revision tracker, slug index | contracts, ir | **[BUILD]** (clone storage iface) |
| `internal/exportstore` | deterministic `<ws>/exports/<id>.pptx`; `deck://export/{id}.pptx` template handler | render, deck | **[ORCH]** |
| `internal/handlers` | one handler per tool over a contract pair; `ToolDeps`; `RegisterTools` | all domain pkgs | **[ORCH]** reference, **[BUILD]** rest |
| `internal/apps` | `RegisterAll(srv, preview, overview, editor []byte)` | dockyard runtime/apps | **[ORCH]** |
| `internal/validate` | native StyleScore: token/contrast/overflow/structural against IR + layout | ir, soul, render | **[ORCH]** |
| `web/design-system` | Deckard White `--app-*` tokens + ThumbStrip/SlideSorter/EmptyState/ErrorState/shells | — | **[ORCH]** |
| `web/apps/*` | the three Svelte surfaces (bridge handshake, callTool handoff) | design-system, dockyard-bridge | **[ORCH]** novel, **[BUILD]** clone |

**Engine boundary (locked):** pptx-go is the *only* PPTX byte producer. It owns
rendering and nothing else. Charts and code blocks are pre-rasterized by
`internal/raster` (pure Go, no Chromium) and fed through an `AssetResolver`. No
HTML stage, no Playwright, no `SlideDocument` measure-the-DOM layer.

---

## 2. Tool contract surface (the complete rewrite surface)

47 tools. Agent-facing unless tagged **[APP]** (`tool.VisibilityApp`, surface-only)
or **[→UI]** (agent-callable AND attached to a surface via `.UI(app)`). All
input/output are `internal/contracts.<Name>` typed structs (Design A). Types below
are the wire shape; `…` = elided optional fields. **Human↔agent handoff rule:** the
surfaces call the *same* agent tools (reorder/edit/export/refine), so there are no
`*_from_app` duplicates — surfaces pass an optional `origin:"app"` field instead.

### Group A — Souls & themes (5) — *bootstrap + refine, easier than the reference*
| Tool | Input | Output | Notes |
|---|---|---|---|
| `bootstrap_soul` | `{name, description?, prompt?, brandTemplateAssetId?, logoAssetId?}` | `{soulId, name, status, tokenCount, previewTokens[]}` | seeds a **complete** soul from NL prompt and/or a brand `.pptx` (`pptx.FromTemplate`); unset roles inherit Deckard White. Replaces the reference's hand-author-all-7-layers. |
| `refine_soul` | `{soulId, overrides:[{layer, token, value}]}` | `{soulId, changed[], tokenCount}` | targeted token override(s); clones theme, rewrites entries; cascade-recompiles affected decks. |
| `list_souls` | `{status?}` | `{souls:[{soulId,name,status,tokenCount}]}` | |
| `get_soul` | `{soulId, includeStyleGuide?}` | `{soul:{…layers}, styleGuide?}` | full soul + design voice (northStar/do/don't). |
| `get_design_tokens` | `{soulId}` | `{tokens:[{name,value,layer}]}` | flat token list. |

### Group B — Templates & recipes (3)
| Tool | Input | Output | Notes |
|---|---|---|---|
| `save_as_template` | `{deckId, slideId, name, description?, tags?}` | `{recipeId, name}` | capture a validated slide IR as an IR-carrying recipe. |
| `list_recipes` | `{soulId?, tag?}` | `{recipes:[{recipeId,name,tags,source}]}` | built-in (6) + user-saved; **no print recipes**. |
| `apply_recipe` | `{deckId, recipeId, position?}` | `{slideId, slide}` | instantiate a recipe as a new slide. |

### Group C — Decks (6)
| Tool | Input | Output | Surface |
|---|---|---|---|
| `create_deck` | `{title?, author?, soulId?}` | `{kind:"deck", deckId, slides[]}` | **[→UI deck-preview]** (slides-only; no `format`/`authoring_model`). |
| `list_decks` | `{}` | `{decks:[{deckId,title,slideCount,soulId}]}` | |
| `get_deck` | `{deckId}` | `{kind:"deck", deckId, title, soulId, chrome, slides[]}` | **[→UI deck-preview]** glanceable summary (IR thumbs, no HTML). |
| `delete_deck` | `{deckId}` | `{deckId, deleted:true}` | |
| `set_deck_chrome` | `{deckId, chrome:{header?,footer?,showOnCover?}}` | `{deckId, chrome}` | native chrome via masters. |
| `set_deck_sections` | `{deckId, sections:[{name, slideIds[]}]}` | `{deckId, sections[]}` | sorter groups (Layer-1 `pres.AddSection`; no scene-IR field). |

### Group D — Slides (6)
| Tool | Input | Output | Surface |
|---|---|---|---|
| `add_slide` | `{deckId, ir:SlideIR, metadata?, position?}` | `{slideId, slide, validation}` | **[→UI deck-preview]** compiles + validates. |
| `update_slide` | `{deckId, slideId, ir:SlideIR, metadata?, expectedRevisionHash?}` | `{slideId, slide, validation}` | full IR replace; optimistic concurrency. |
| `get_slide` | `{deckId, slideId}` | `{slideId, ir, metadata, validation}` | |
| `remove_slide` | `{deckId, slideId}` | `{deckId, removed:true}` | |
| `reorder_slides` | `{deckId, order:[slideId…]}` | `{kind:"deck", deckId, slides[]}` | **[→UI deck-preview]**. |
| `duplicate_slide` | `{deckId, slideId, position?}` | `{slideId, slide}` | |

### Group E — Fine-grained slide editing (7) — *node-level, path-addressed*
| Tool | Input | Output |
|---|---|---|
| `edit_slide_node` | `{deckId, slideId, path:[…], node:SlideNode, expectedRevisionHash?}` | `{slide, validation}` |
| `edit_slide_field` | `{deckId, slideId, path:[…], field, value, expectedRevisionHash?}` | `{slide, validation}` |
| `patch_slide_text` | `{deckId, slideId, path:[…], text, expectedRevisionHash?}` | `{slide, validation}` |
| `insert_slide_node` | `{deckId, slideId, path:[…], node:SlideNode}` | `{slide, validation}` |
| `remove_slide_node` | `{deckId, slideId, path:[…]}` | `{slide, validation}` |
| `duplicate_slide_node` | `{deckId, slideId, path:[…]}` | `{slide, validation}` |
| `move_slide_node` | `{deckId, slideId, from:[…], to:[…]}` | `{slide, validation}` |

### Group F — Authoring helpers (2)
| Tool | Input | Output | Notes |
|---|---|---|---|
| `compile_markdown` | `{markdown, target?:{deckId,slideId,path}}` | `{nodes[], warnings[]}` | markdown → IR leaf nodes; optional insert. |
| `compile_chart` | `{spec:{chartType,data,…}, target?}` | `{node:ChartNode, assetId, warnings[]}` | chart spec → `chart` node; **pure-Go raster** bytes registered as an asset. |

### Group G — Assets (4)
| Tool | Input | Output | Notes |
|---|---|---|---|
| `upload_asset` | `{filename, mime, dataBase64, origin?}` | `{assetId:"asset://…", mime, bytes}` | LLM never re-handles bytes after upload. |
| `list_assets` | `{}` | `{assets:[{assetId,filename,mime}]}` | metadata only. |
| `get_asset` | `{assetId}` | `{assetId,filename,mime}` | never returns binary. |
| `delete_asset` | `{assetId}` | `{assetId, deleted:true}` | |

### Group H — Comments / between-turn collaboration (3)
| Tool | Input | Output |
|---|---|---|
| `add_comment` | `{deckId, target:{kind, slideId?, irPath?}, body, kind?, origin?}` | `{commentId}` |
| `list_comments` | `{deckId, resolved?, targetKind?}` | `{comments[]}` |
| `resolve_comment` | `{commentId, note?}` | `{commentId, resolved:true}` |

### Group I — Session (1)
| Tool | Input | Output |
|---|---|---|
| `get_session` | `{}` | `{activeDeckId?, activeSoulId?, openPanels[], buildInfo}` |

### Group J — Validation (3) — *native, no Chromium*
| Tool | Input | Output | Notes |
|---|---|---|---|
| `validate_slide_ir` | `{ir:SlideIR, soulId?}` | `{ok, issues[]}` | shape pre-flight, no storage. |
| `validate_slide` | `{deckId, slideId}` | `{styleScore, passed, issues[]}` | native token/contrast/overflow/structural against IR + pptx layout. |
| `validate_deck_for_export` | `{deckId}` | `{ok, perSlide[], blockers[]}` | pre-export gate. |

### Group K — Export & resources (3) — *path + downloadable resource*
| Tool | Input | Output | Surface |
|---|---|---|---|
| `export_deck` | `{deckId}` | `{path, resourceUri:"deck://export/<id>.pptx", stats:{slides,shapes,warnings[]}}` | **[→UI deck-preview]** ALWAYS writes deterministic path + registers `deck://` resource. No `include_data` flag. |
| `list_resources` | `{}` | `{resources:[{uri,mime,title}]}` | wraps `deck://` docs/schema for tool-only clients. |
| `get_resource` | `{uri}` | `{uri, mime, text? , blobBase64?}` | |

### Group L — Surface tools (4) — *the three-surface model*
| Tool | Input | Output | Visibility |
|---|---|---|---|
| `get_deck_overview` | `{deckId}` | `{kind:"overview", deckId, sections[], slides[]}` | **[→UI deck-overview]** agent+app; structure/reorder view. |
| `open_slide_editor` | `{deckId, slideId}` | `{kind:"editor", slideId, ir, soulId, validation}` | **[→UI slide-editor]** model→app; **opt-in** deep edit of ONE slide. |
| `get_deck_state` | `{deckId, selectedSlideId?}` | `{slides[], selected?, souls[], validation}` | **[APP]** rich hydration the surfaces fetch on mount. |
| `set_active_workspace` | `{deckId?, soulId?}` | `{activeDeckId?, activeSoulId?}` | **[APP]** session write. |

**Dropped vs reference (locked):** all `*_section` / document / `update_document_meta`
tools, `export_html`, `export_google_slides`, `export_pdf` document path,
`render_preview`/`render_section_preview` (surfaces render IR natively in Svelte;
no server-side raster preview), `apply_block_edit` document parts, the single
`open_deck_editor` fullscreen-default tool, and the `*_from_app` duplicates.

---

## 3. The three surfaces (what each renders / does)

| Surface | URI | Default? | Renders | Drives (via `callTool`, same agent tools) |
|---|---|---|---|---|
| **deck-preview** | `ui://widget/deck-preview` | **YES (inline, glanceable)** | slide thumbnails (IR rendered natively in Svelte) + sorter + quick actions `[download]` `[edit this]` | `export_deck`, `reorder_slides`, then `open_slide_editor` on `[edit this]` |
| **deck-overview** | `ui://app/deck-overview` | opt-in (inline+fullscreen) | section/slide selector, reorder/structure, impact-modal on destructive ops | `set_deck_sections`, `reorder_slides`, `remove_slide`, `duplicate_slide` |
| **slide-editor** | `ui://app/slide-editor` | **opt-in only**, one slide | deep edit of a single slide IR with "← back to deck"; inline rename | `edit_slide_node`, `edit_slide_field`, `patch_slide_text`, `refine_soul`, `update_slide` |

Locked UX rules: **never auto-open fullscreen**; preview is the default attach
target for the common authoring tools; the slide-editor opens only on an explicit
"edit this slide" action. Document in `docs/ui/SURFACES.md`. `[download]` resolves
the `deck://export/<id>.pptx` resource. All three honor host theme and ship the
four UI states (loading/empty/error/permission) each exercised by a fixture; the
`large.json` fixture is a **high-complexity ceiling deck**, not a toy.

---

## 4. Phase breakdown

Each phase: **goal · files · reference pattern to clone · acceptance checks · gate
commands · ownership.** Sequencing guarantees a correct reference unit exists
before any fan-out. The universal per-unit gate (every unit, every phase) is:

```bash
gofmt -l .                                   # MUST print nothing
GOFLAGS="" dockyard generate
git diff --exit-code internal/contracts      # no stale codegen
GOFLAGS="" dockyard validate                 # 0 blockers
GOFLAGS="" CGO_ENABLED=0 go build ./...
GOFLAGS="" go vet ./...
GOFLAGS="" go test -race ./...
GOFLAGS="" dockyard test                      # contract+spec+capability
# UI units add:
cd web/apps/<surface> && npx svelte-check --tsconfig ./tsconfig.json && npx vite build
```

Below, "gate" = the universal gate above plus the phase-specific checks named.

---

### Phase 0 — Foundation / scaffold **[ORCH]**
- **Goal:** a buildable, validating, serving blank Dockyard server with the Deckard
  layout, manifest skeleton (3 apps + ≥1 tool), Makefile, CI, `.golangci.yml`.
- **Files:** `go.mod`, `main.go`, `dockyard.app.yaml`, `Makefile`, `.golangci.yml`,
  `.github/workflows/ci.yml`, `internal/contracts/contracts.go` (one example tool),
  `internal/handlers/`, `internal/apps/register.go`, empty `web/apps/*` shells.
- **Clone:** `scaffold-a-server` (blank), then the WorkBridge repo shape from
  `docs/research/03`. Use `dockyard new go-slides-mcp --module github.com/hurtener/go-slides-mcp`.
- **Acceptance:** `dockyard validate` exits 0; `CGO_ENABLED=0 go build ./...`
  compiles; `make preflight` green; `main.go` registers apps BEFORE tools.
- **Gate:** universal gate + `make preflight` + `diff -q AGENTS.md CLAUDE.md`.

### Phase 1 — IR model & contract conventions **[ORCH]** — ENGINE-FIRST (see ADR-0001)
- **Design principle (binding, ADR-0001):** the Deckard IR is a contract-first,
  JSON-native, agent-ergonomic **mirror of pptx-go's full `scene` node catalog** — NOT a
  port of the TS reference. The TS nodes are a *coverage checklist* only; we were limited
  there by the TS render library and are not limited now. Expose the engine's full ceiling.
- **Goal:** the slide IR node grammar as Go types that map ~1:1 onto `scene.SlideNode`,
  covering the ENTIRE engine catalog, + RichText + normalize/validate-shape + structural-path
  edits + IR content hash; locked `CONVENTIONS.md`.
- **The node set (from the `compose-a-scene` skill — authoritative):**
  - Leaf: `hero, prose, heading(1..6), list(bullet|number|checklist, levels, checked),
    divider, quote, callout(note|warning|tip|important), chip(tint|solid|outline),
    arrow, section_divider, table, flow(orientation, connector, steps{label,detail,icon}),
    image(frame: none|browser|phone|desktop|laptop, crop, fit), code_block, chart,
    decoration(preset|asset, layer, opacity, rotation, bleed)`.
  - Container (RECURSIVELY NESTABLE — the TS cap on nesting is GONE): `two_column(ratio,
    left[], right[])`, `grid(columns 2..4, ratio[], gap, cells[])`,
    `card(header,eyebrow,icon,headerPill,body[],bodyLayout,fill,outline,borderStyle,size,
    layout,elevation)`, `card_section(header, body[])`.
  - RichText = ordered runs `{text, style{typeRole,bold,italic,underline,strike,code,link,
    href}, color: token-role | literal}`. Colors are SEMANTIC token roles (soul re-skins);
    literal hex is the explicit escape hatch only.
- **Composed patterns, NOT new node types:** TS "special" blocks (timeline, kpi_cards,
  comparison, etc.) are delivered as **recipes** built from `flow`/`grid`/`card` — never as
  lossy custom IR types the engine can't render. Stay inside the engine's sealed union.
- **Files:** `internal/contracts/ir.go` (the node union + RichText + token-role enums +
  slide `Layout` kinds + chrome), split across small files if needed; `internal/ir/{normalize,
  path,hash}.go`; `docs/contracts/CONVENTIONS.md`; `docs/decisions/0001-engine-first-ir.md`.
- **Reference (clone source):** the `compose-a-scene` and `define-a-theme` SKILLS + the
  pptx-go `scene` package — copy real struct/field shapes; do NOT invent fields the engine
  lacks, and do NOT reach for the TS schema. Node-as-Go-struct + a `kind` JSON discriminator.
- **Acceptance:** every engine node + RichText round-trips JSON (table tests); containers
  nest recursively and validate children; path-edit ops (`["body",2,"left",1]`) have table
  tests; `validate_slide_ir` mirrors `scene.ValidateScene`'s per-node rules and rejects a
  malformed node; **no** `toc`/`bibliography`/`page_break`/`SectionIR`/A4 types exist; no IR
  node lacks a `scene` counterpart (every kind has a render path in Phase 3).
- **Gate:** universal + `! grep -rE 'toc|bibliography|page_break|SectionIR' internal/contracts/ir.go`.
- **High-judgment:** the discriminated-union JSON encoding, recursive container nesting, and
  the path-edit semantics — the reference that `internal/ir` per-node code, the Phase 3
  render mappers, and the Group-E handlers all clone.

### Phase 2 — Soul/token engine + Deckard White **[ORCH]**
- **Goal:** soul model, the built-in **Deckard White** soul as a `*pptx.Theme`,
  bootstrap (NL + brand-template) and refine (token override) producing themes.
- **Files:** `internal/soul/{soul.go,deckard_white.go,bootstrap.go,refine.go,
  store.go}`.
- **Clone:** `define-a-theme` + the exact construction in `docs/research/04` §9
  (warm palette, 400/500 weights, Lora H1–H3, tightened spacing, warm shadows);
  `load-a-brand-template` for `pptx.FromTemplate`.
- **Acceptance:** `DefaultSoul()` returns a complete theme (no unset role, by
  `Clone()`); accent splits fill `#3B9C94` / text `#2B7A73` (AA rule); refine of one
  token re-skins a sample render; bootstrap from a brand `.pptx` inherits its
  palette/fonts. Soul persists as serialized theme + extension metadata
  (border/tooltip tokens that have no native field).
- **Gate:** universal + a golden test asserting Deckard White token values.
- **High-judgment:** the soul↔`*pptx.Theme` mapping and the persistence shape (engine
  gap: code-authored themes don't round-trip through `theme1.xml` — soul is our truth).

### Phase 3 — pptx-go render adapter **[ORCH] core, [BUILD] per-node**
- **Goal:** IR → `scene.Scene` → `.pptx` bytes, deterministic, via `scene.Render`.
- **Files:** `internal/render/{engine.go,scene.go}` (driver, [ORCH]) +
  `internal/render/node_<kind>.go` (one per node, [BUILD] fan-out).
- **Clone:** ONE reference mapper (`node_hero.go`) hand-written by [ORCH] with a
  golden `.pptx` test; the builder clones it per node against the engine-map §4
  policy table. Asset-bearing nodes (image/chart/code/decoration) carry `AssetID`.
- **Acceptance:** the ceiling fixture deck renders to a valid `.pptx` that opens
  without a repair prompt; byte-identical output across `WithWorkers(1)` vs N;
  every node kind has a mapper + golden; overflow surfaces as a warning, not a fail.
- **Gate:** universal + golden `.pptx` byte-equality test + worker-count determinism test.
- **High-judgment:** the `scene.Render` driver, theme application, asset-resolver
  wiring, determinism contract. **Clonable:** the per-node mappers once the hero
  reference exists.

### Phase 4 — Storage layer **[BUILD]**
- **Goal:** Deck/Slide/Section store (memory + file impls), branded IDs, slug index,
  revision tracker (hash over IR), optimistic concurrency.
- **Files:** `internal/deck/{store.go,memory.go,file.go,ids.go,revision.go,slug.go}`.
- **Clone:** the reference store interface in `docs/research/01` §4.2–4.3; one
  interface + memory impl written by [ORCH], file impl + remaining entity stores
  cloned by [BUILD].
- **Acceptance:** CRUD + reorder round-trip; every mutation creates a `DeckRevision`;
  `expectedRevisionHash` mismatch returns a typed conflict; slug↔UUID resolution works.
- **Gate:** universal + store table tests with `-race`.

### Phase 5 — Tool handlers fan-out **[ORCH] reference, [BUILD] rest**
- **Goal:** all 47 tools registered, each handler CALLS the real store/renderer/soul
  (no stubs); `ToolDeps` DI; `RegisterTools(srv, deps)`.
- **Files:** `internal/handlers/{deck.go,slide.go,edit.go,soul.go,recipe.go,asset.go,
  comment.go,session.go,validate.go,export.go,surface.go,deps.go}` + per-tool
  contracts in `internal/contracts/contracts.go` + `fixtures/<tool>/*.json`.
- **Clone:** `add-a-tool` + `define-contracts`. [ORCH] writes ONE reference per
  shape — a read tool (`get_deck`), a mutate tool (`add_slide`), a UI-attached tool
  (`export_deck`) — each with contract + handler + 6 fixtures + contract test. [BUILD]
  clones the remaining ~44 by analogy, group by group, in dependency order
  (decks → slides → edits → souls/recipes → assets/comments/session → validation →
  surface tools).
- **Acceptance:** every tool in `dockyard.app.yaml` is registered and its handler
  invokes a dep (grep-proof: no handler returns a literal/empty struct without a
  Store/Renderer/Souls call); `RegisterTools` receives `ToolDeps`; each UI-bearing
  tool has all six fixtures incl. a complex `large.json`.
- **Gate:** universal + fixture presence check + the stub-trap grep from
  `docs/research/03` §5.5.
- **High-judgment:** the three reference handlers + the contract conventions.
  **Clonable:** the ~44 remaining handlers.

### Phase 6 — Export + `deck://` resources **[ORCH]**
- **Goal:** `export_deck` writes `<ws>/exports/<id>.pptx` (deterministic) and returns
  `path` + `resourceUri`; `deck://export/{id}.pptx` reads the same bytes back.
- **Files:** `internal/exportstore/{paths.go,resource.go}`, wired in `main.go`.
- **Clone:** `docs/research/03` §4 (`server.AddResourceTemplate`, `Blob` bytes,
  scheme-bearing URI).
- **Acceptance:** export returns absolute path AND `deck://…`; reading the resource
  returns byte-identical content to the file; re-export is idempotent (same path).
- **Gate:** universal + a test that exports then reads `deck://` and asserts byte-equality.

### Phase 7 — Asset resolver + pure-Go rasterizers **[ORCH]**
- **Goal:** `AssetResolver` over `asset://<uuid>`; pure-Go chart + code rasterizers
  (no Chromium) feeding chart/code-block nodes.
- **Files:** `internal/raster/{chart.go,code.go,resolver.go}`.
- **Clone:** `register-an-asset`, `embed-a-chart-raster`, `embed-a-code-block-raster`.
- **Acceptance:** `compile_chart` produces PNG bytes resolved at render; a code block
  renders as a raster + native language badge; missing asset → warn-don't-fail.
- **Gate:** universal + a render test placing a chart and a code block.

### Phase 8 — UI surface 1: deck-preview + design system **[ORCH]**
- **Goal:** the default inline glanceable surface + the shared Deckard White design
  system; the bridge handshake + `callTool` handoff pattern proven.
- **Files:** `web/design-system/**` (tokens `--app-*`, ThumbStrip, SlideSorter, the
  four state components, shells), `web/apps/deck-preview/**` (vite.config iife+strip,
  `App.svelte` with top-level `createBridge`, `main.ts` `mount(App,{target})`).
- **Clone:** `attach-a-ui-resource` + the WorkBridge `vite.config.ts` verbatim
  (`docs/research/03` §3.5) + Deckard White tokens (`docs/research/04`).
- **Acceptance:** renders in `dockyard inspect` against fixtures (not standalone),
  four states each, host theme honored; thumbnails render IR natively; `[download]`
  resolves the `deck://` resource; `[edit this]` calls `open_slide_editor`. NEVER
  auto-opens fullscreen.
- **Gate:** universal UI gate + `dockyard inspect` Apps-tab render verified per fixture.
- **High-judgment:** the design system + the first surface (the reference shell the
  other two clone).

### Phase 9 — UI surfaces 2 & 3: deck-overview + slide-editor **[ORCH] novel, [BUILD] shell**
- **Goal:** the structure/reorder surface and the opt-in single-slide editor.
- **Files:** `web/apps/deck-overview/**`, `web/apps/slide-editor/**`.
- **Clone:** the deck-preview shell + bridge + vite config from Phase 8; [ORCH] owns
  the novel interactions (drag-reorder + impact-modal; node-level edit form); [BUILD]
  clones the shell/handshake/state-machine.
- **Acceptance:** overview reorders via `reorder_slides`/`set_deck_sections`,
  destructive ops show the impact-modal (never native `confirm()`); editor edits ONE
  slide via the node-edit tools with "← back to deck"; both render four states in the
  inspector; human edit == the same tool the agent calls.
- **Gate:** universal UI gate for both surfaces + inspector render per fixture.

### Phase 10 — Native validation engine **[ORCH]**
- **Goal:** `validate_slide` / `validate_deck_for_export` as native IR + pptx-layout
  checks; StyleScore; no Chromium.
- **Files:** `internal/validate/{score.go,contrast.go,overflow.go,structural.go}`.
- **Clone:** `docs/research/01` §4.5 (StyleScore weights, error/warn deltas); drop
  the HTML/CSS lints made moot by IR-only authoring.
- **Acceptance:** contrast checked against resolved theme colors; overflow detected
  against the pptx layout; `passed = errorCount==0`; weighted score 0–1.
- **Gate:** universal + scoring table tests.

### Phase 11 — Polish, packaging, gate hardening **[ORCH] + [BUILD]**
- **Goal:** ship a single CGo-free binary embedding all three surfaces; CI green; docs
  complete; no stubs/TODOs.
- **Files:** `release.yml`, `docs/ui/SURFACES.md`, `docs/glossary.md`, ADRs.
- **Clone:** `package` (`dockyard build --cross-compile`), `validate`/`test`.
- **Acceptance:** `dockyard build --cross-compile` produces per-platform binaries +
  `.sha256`; `dockyard install claude` boot-checks; clean-build before embed (no
  stale `dist`); zero TODO/FIXME/"not implemented"/empty-stub returns; the word
  formerly used for the white theme appears nowhere.
- **Gate:** full machine gate (`docs/research/03` §5.4) + `grep -ri` forbidden-name
  scan returns nothing + four-pass polish (UI / wiring / no-stubs / inspector live-test).

---

## 5. Ownership summary

- **[ORCH] high-judgment (write the reference, own verification):** IR grammar &
  contracts (P1), soul↔theme engine (P2), the `scene.Render` adapter core +
  determinism (P3), export/`deck://` resource wiring (P6), pure-Go rasterizers (P7),
  the design system + all three novel surfaces (P8–P9), the validation engine (P10),
  the three reference tool handlers (P5).
- **[BUILD] clonable (fan-out against the green gate):** per-node renderers (P3),
  the storage interface impls (P4), the ~44 non-reference tool handlers (P5), the
  surface shells/handshake for overview + editor (P9).

Builder advances only on the **green gate output**, never on a stop-token claim
(`docs/research/03` §5.5). A handler that returns an empty struct without calling a
dep is the stub trap and fails orchestrator verification.
