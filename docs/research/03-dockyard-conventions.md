# Dockyard Conventions for Deckard (`go-slides-mcp`)

> **Status:** Research / binding-conventions draft.
> **Audience:** the rewrite plan and the autonomous builder loop.
> **Scope:** how to build the Deckard MCP App server on the Dockyard framework —
> project layout, the contract-first Design A workflow, adding tools, attaching the
> three `ui://` surfaces, the dev loop / inspector / packaging, and the exact
> machine-checkable gate commands.

Sources: the installed Dockyard skills (`scaffold-a-server`, `add-a-tool`,
`attach-a-ui-resource`, `define-contracts`, `run-the-dev-loop`, `validate`,
`test-with-the-inspector`, `package`, `orchestrate-autonomous-build`,
`dockyard-study-mcp`), the WorkBridge production repo, and the Dockyard runtime
module source (`github.com/hurtener/dockyard@v1.7.1`, packages `runtime/server`,
`runtime/apps`, `runtime/tool`, `runtime/tasks`).

Product decisions are LOCKED upstream (see the brief). This doc shows *how* to
realize them on Dockyard; it does not re-litigate them.

- Product: **Deckard** (UI title **Deckard Slides**). Repo `go-slides-mcp`. Module
  `github.com/hurtener/go-slides-mcp`.
- Full Go rewrite on Dockyard, single CGo-free static binary.
- Agent-first: MCP tools are the primary authoring path.
- THREE distinct `ui://` surfaces served by ONE server.
- Export = deterministic workspace path **and** a readable MCP resource under
  `deck://`.
- Pure Go PPTX via pptx-go. No Playwright/Chromium. No HTML→render export stage.
- Built-in default soul: **Deckard White**.

---

## 0. The priority chain (so drift resolves deterministically)

WorkBridge's `CLAUDE.md` is the model. Lock a chain and make it binding on humans
and agents alike. Recommended for Deckard:

```
RFC / design source of truth  >  phased plan (01-plan.md)  >  CLAUDE.md (== AGENTS.md)
  >  docs/contracts/CONVENTIONS.md  >  docs/decisions/*.md (ADRs)  >  code comments
```

Mirror `CLAUDE.md` verbatim to `AGENTS.md` (WorkBridge has a `make check-mirror`
target: `diff -q AGENTS.md CLAUDE.md`). When two artifacts disagree, fix the
lower-priority one — never diverge silently.

---

## 1. Canonical project layout

### 1.1 Single-module, single-binary (Deckard's shape)

WorkBridge is a **12-binary `go.work` workspace** (`servers/<name>` handler
packages + `cmd/ms365-<name>-mcp` mains + `web/apps/<name>` UIs + a shared
`web/design-system`). Deckard is **one server with three UI surfaces**, so it
collapses to a single Go module and a single `main`. Keep WorkBridge's *internal
discipline* (typed contracts package, handlers package, shared design-system,
fixtures) but drop the multi-module workspace.

The blank `dockyard new` scaffold (the first-class path) produces:

```text
my-server/
├── README.md
├── dockyard.app.yaml          # the manifest (RFC §4.2)
├── go.mod
├── greet.go                   # registerTools + example handler
├── greet_test.go              # contract test
├── main.go                    # stdio | http serve
└── internal/
    └── contracts/
        └── contracts.go       # typed Input/Output structs — source of truth
```

Deckard grows that into:

```text
go-slides-mcp/
├── README.md
├── CLAUDE.md  AGENTS.md            # binding normatives (verbatim mirror)
├── Makefile                        # canonical gate commands (see §5)
├── go.mod                          # module github.com/hurtener/go-slides-mcp
├── go.sum
├── go.work  go.work.sum            # OPTIONAL — only to develop pptx-go locally (§1.4)
├── .golangci.yml                   # pinned linter config (§5.4)
├── .editorconfig  .gitignore
├── .github/
│   └── workflows/
│       ├── ci.yml                  # preflight gate: generate→validate→build→test→lint
│       └── release.yml             # dockyard build --cross-compile on tag
│
├── main.go                         # server.New + register souls/apps BEFORE tools + serve
├── dockyard.app.yaml               # ONE manifest: apps[] (×3) + tools[] + quality{}
│
├── internal/
│   ├── contracts/                  # CONTRACT-FIRST source of truth
│   │   ├── contracts.go            #   typed In/Out structs (Deck, Slide, Soul, Export…)
│   │   ├── *_input.schema.json     #   GENERATED — never hand-edit
│   │   ├── *_output.schema.json    #   GENERATED
│   │   └── contracts.ts            #   GENERATED TS (if codegen emits here; see §1.3)
│   ├── handlers/                   # typed tool handlers (one file per tool group)
│   │   ├── deck.go                 #   create/add/reorder/remove slide, get deck
│   │   ├── soul.go                 #   bootstrap_soul + refine_soul (token overrides)
│   │   ├── export.go               #   export_deck → writes .pptx + registers deck://
│   │   └── deps.go                 #   ToolDeps struct (DI: store, renderer, workspace)
│   ├── deck/                       # domain model: Deck/Section/Slide store + IDs
│   ├── soul/                       # soul model + Deckard White built-in + token merge
│   ├── render/                     # pptx-go bridge: scene IR → .pptx bytes
│   ├── exportstore/                # deterministic workspace paths + deck:// resource fn
│   ├── apps/                       # RegisterUI helpers (one per surface) — see §3
│   └── fixtures/                   # shared fake/sample decks for tests + UI fixtures
│
├── web/
│   ├── design-system/              # "Deckard White" soul tokens + shared Svelte primitives
│   │   ├── package.json            #   name @deckard/design-system, private, file: dep
│   │   ├── theme/                  #   base.css / default.css (CSS custom properties)
│   │   ├── components/             #   ThumbStrip, SlideSorter, EmptyState, ErrorState…
│   │   ├── layouts/                #   PreviewShell, OverviewShell, EditorShell
│   │   └── src/index.ts
│   └── apps/                       # ONE Vite project per ui:// surface (§3.4)
│       ├── deck-preview/           #   ui://widget/deck-preview   (inline, default)
│       │   ├── index.html          #     entry at ROOT (not src/)
│       │   ├── package.json        #     deps: dockyard-bridge, @deckard/design-system
│       │   ├── vite.config.ts      #     iife + stripModuleType + singlefile (§3.5)
│       │   ├── tsconfig.json
│       │   ├── src/main.ts         #     mount(App,{target}) — NOT new App()
│       │   ├── src/App.svelte      #     createBridge() at TOP LEVEL
│       │   └── dist/index.html     #     BUILT — embedded by go:embed
│       ├── deck-overview/          #   ui://app/deck-overview
│       └── slide-editor/           #   ui://app/slide-editor
│
├── fixtures/                       # per-tool inspector/validate fixtures (§2.4)
│   ├── create_deck/{happy,empty,error,permission,slow,large}.json
│   ├── export_deck/…
│   └── …
│
└── docs/
    ├── contracts/CONVENTIONS.md    # locked tool-contract rules
    ├── ui/SURFACES.md              # the three-surface rules (when each opens)
    ├── decisions/                  # numbered ADRs
    ├── research/                   # this file
    └── glossary.md
```

### 1.2 Where each kind of artifact lives (the rules)

- **Tool contracts (source of truth):** `internal/contracts/contracts.go` — plain Go
  structs with `json:` tags and a leading `//` comment per field (the comment is
  lifted into the JSON Schema `description`). Keep them in `internal/` so they are
  not a public API.
- **Generated artifacts:** `internal/contracts/*.schema.json` and the TypeScript
  types. **Never hand-edit a `*.gen.*` / `*.schema.json` / generated `.ts`** —
  `dockyard validate` rejects hand-edited generated files (stale-codegen / CrossCheck).
- **Handlers:** `internal/handlers/` — typed functions over the contract pair.
- **Manifest:** one `dockyard.app.yaml` at the repo root (RFC §4.2).
- **Svelte UI:** `web/apps/<surface>/` — one Vite single-file project per `ui://`
  surface; shared tokens/components in `web/design-system/`.
- **The embed bundle:** each surface's built `web/apps/<surface>/dist/index.html` is
  pulled into the Go binary with `//go:embed`. WorkBridge embeds per-surface bytes:
  `//go:embed web/apps/identity/dist/index.html` → `var uiIndexHTML []byte`. Deckard
  uses three such embeds (one per surface) or one `//go:embed all:web/apps` FS read.

### 1.3 The manifest (`dockyard.app.yaml`)

One manifest declares the runtime, all three apps, every tool, and the quality
block. Skeleton (Deckard-shaped — names illustrative):

```yaml
name: go-slides-mcp
title: Deckard Slides
version: 0.1.0

runtime:
  transports: [stdio, http]

apps:
  - id: deck-preview                       # DEFAULT inline glanceable surface
    uri: ui://widget/deck-preview          # framework treats ui:// as OPAQUE (D-178)
    entry: web/apps/deck-preview/src/App.svelte
    display_modes: [inline]
    csp: { connect: [], resource: [] }     # deny-by-default; single-file bundle
    visibility: [model, app]
  - id: deck-overview                       # section/slide selector + reorder
    uri: ui://app/deck-overview
    entry: web/apps/deck-overview/src/App.svelte
    display_modes: [inline, fullscreen]
    csp: { connect: [], resource: [] }
    visibility: [model, app]
  - id: slide-editor                        # OPT-IN deep edit of ONE slide
    uri: ui://app/slide-editor
    entry: web/apps/slide-editor/src/App.svelte
    display_modes: [inline, fullscreen]
    csp: { connect: [], resource: [] }
    visibility: [model, app]

tools:
  - name: create_deck
    description: Create a new deck (agent-first authoring entry point).
    input: internal/contracts.CreateDeckInput
    output: internal/contracts.CreateDeckOutput
    ui: deck-preview                         # default surface for glanceable result
    task_support: forbidden
  # … add_slide, reorder_slides, edit_slide, get_deck, bootstrap_soul,
  #    refine_soul, export_deck …  (input/output are <pkg-path>.<TypeName> refs)

quality:
  require_loading_state: true
  require_empty_state: true
  require_error_state: true
  require_permission_state: true
  require_fixtures: true
  require_contract_tests: true
  require_spec_compliance: true
```

Notes:
- The `ui://` string is an **opaque identifier** to the framework. The reference
  MCP Apps convention is `ui://<server>/<app>/index.html` (D-178); the LOCKED
  Deckard URIs (`ui://widget/deck-preview`, `ui://app/deck-overview`,
  `ui://app/slide-editor`) are honored verbatim — only the documentation convention
  moved, existing URIs keep working.
- `input:`/`output:` are Go type references in `<package-path>.<TypeName>` form,
  resolved against the module (`internal/contracts.CreateDeckInput`).
- The `quality{}` block turns on the §20 four-state page rule and the v1.3
  fixtures/contract-test gates. `require_fixtures` is UI-scoped (only UI-bearing
  tools need fixtures). WorkBridge sets the four `require_*_state` to `false` and
  relies on `require_fixtures: true` + `require_contract_tests: true`; for
  Deckard's review surfaces, turn the four-state flags ON.

### 1.4 `go.work` — optional, only for local pptx-go development

pptx-go is a published module resolved from the Go proxy; Deckard needs **no
workspace** for a normal build. Add a `go.work` only if you are co-developing
pptx-go from a local checkout (mirrors WorkBridge developing Dockyard locally):

```
go 1.26
use (
    .
    /path/to/pptx-go   // local replace via go.work — never commit a machine-local path
)
```

Otherwise pin pptx-go (and `github.com/hurtener/dockyard`) by version in `go.mod`
and let `go mod tidy` resolve from the proxy. Do **not** set `GOFLAGS=-mod=mod` —
WorkBridge documents that it conflicts with workspace mode and silently breaks
every build (a "builder wrote broken code" red herring that was actually env
sabotage).

---

## 2. Contract-first "Design A" workflow

Design A (RFC §6, P1): the **typed Go struct is the single source of truth**. JSON
Schema (what the host sees) and TypeScript types (what the Svelte App imports) are
**generated**. You never hand-write either; if you try, `dockyard validate` rejects
it. This is Dockyard's headline differentiator — lean on it.

### 2.1 Write the contract

```go
// internal/contracts/contracts.go
package contracts

// CreateDeckInput is the model-facing input for create_deck.
type CreateDeckInput struct {
    // Title is the deck title shown on the cover slide. Required.
    Title string `json:"title"`
    // Soul is the design-soul id to apply; defaults to "deckard-white".
    Soul string `json:"soul,omitempty"`
}

// CreateDeckOutput is the typed, UI-facing output for create_deck.
type CreateDeckOutput struct {
    // Kind is the renderer discriminator ("deck") for the multi-surface App.
    Kind string `json:"kind"`
    // DeckID is the opaque id used by subsequent tools and the deck:// scheme.
    DeckID string `json:"deckId"`
    // Slides are the slide thumbnails/metadata the preview surface renders.
    Slides []SlideThumb `json:"slides"`
}
```

Conventions that pay off downstream (from `define-contracts` + WorkBridge
`docs/contracts/CONVENTIONS.md`):

- Document **every** field with a leading `//` comment → JSON Schema `description`
  + TS JSDoc. A model reads these; they are part of the product surface.
- `json:"name,omitempty"` for optional fields → codegen marks them optional.
- Prefer **named scalar types** for constrained values (e.g. `type SoulToken
  string`) → typed TS union, guides the model.
- Use a **`Kind` discriminator** on outputs that drive a multi-renderer App so the
  Svelte dispatcher switches without sniffing shape.
- Replace JSON-incompatible Go types: `time.Time` → ISO-8601 `string` or `int64`.
  The codegen errors with "unsupported field type" otherwise.

### 2.2 Generate → JSON Schema + TS types

```bash
dockyard generate            # in the project dir (or: dockyard generate --dir .)
```

- Reads each tool's `input:`/`output:` Go types named in the manifest, emits the
  JSON Schema and the TypeScript types.
- **Byte-deterministic / idempotent**: rerun with no source change → no diff (P1,
  RFC §6.2).
- Failure modes: `"unknown type ref"` (manifest path doesn't resolve to the
  contracts package) and `"unsupported field type"` (use a JSON-compatible type).

### 2.3 Drift detection — what `validate` and `test` check

`dockyard validate` (fast, build-blocking — RFC §9.4):

| Category            | What it catches |
|---------------------|-----------------|
| manifest            | malformed `dockyard.app.yaml`, missing required fields |
| schemas             | generated JSON Schema invalid against draft-2020-12 |
| tool↔UI mappings    | a tool's `ui: <id>` matches no `apps[].id` |
| MIME                | an App resource MIME ≠ `text/html;profile=mcp-app` |
| spec compliance     | Apps/Tasks shapes deviate from vendored MCP specs |
| four-state UI rule  | a fixture missing for a required UI state (§20) |
| fixtures (D-169)    | `require_fixtures` on + UI tool ships no `fixtures/<tool>/*.json` |
| contract tests (D-168) | `require_contract_tests` on + no `*_test.go` |
| **stale-codegen**   | generated `*.schema.json`/`.ts` no longer match the Go source |
| **CrossCheck (D-113)** | the TS that `generate` *would* produce now differs from disk — catches silent server↔UI drift |

`dockyard test` (full gate) categories:

| Category          | Runs |
|-------------------|------|
| `go-test`         | the project's `go test ./...` |
| `contract`        | generated schema + TS still match the Go contracts |
| `golden`          | fixture/golden snapshots present and coherent |
| `spec-compliance` | Apps/Tasks shapes conform to vendored MCP specs |
| `capability`      | degrades gracefully across host capability sets |

Both exit non-zero on a regression / blocker, 0 when clean; **warnings do not change
the exit code**. Report shape: `<level>: <message>`, diagnostics → warnings →
one-line verdict (e.g. `validate: 1 blocker, 1 warning`). `dockyard test
--skip-go-test` keeps contract/spec/capability gates when the slow tests ran
elsewhere.

> **Why this matters for Deckard:** the three surfaces import the same generated
> TS. If a contract field is renamed and only one surface is updated, CrossCheck
> + the App's `svelte-check` build fail at `validate`/`web-check` — *before* ship.

### 2.4 Fixtures

Each UI-bearing tool ships six fixtures under `fixtures/<tool>/`:
`happy.json`, `empty.json`, `error.json`, `permission.json`, `slow.json`,
`large.json` (D-130). On-disk fixtures are preferred over schema-derived synthetic
ones. They drive both the inspector's Fixtures switcher and the four-state
`validate` rule. The `analytics-widgets` template ships all six per tool as a
reference. For Deckard, `large.json` must exercise a **highly complex deck** (the
"enhance, don't simplify, the output ceiling" decision) so the surfaces are proven
against the real ceiling, not a toy deck.

---

## 3. Adding a tool + attaching the THREE `ui://` surfaces

### 3.1 Add a tool (four steps + regenerate)

1. **Contracts** in `internal/contracts/`.
2. **Handler** in `internal/handlers/`. Signature: `func(ctx, In) (tool.Result[Out],
   error)`. `tool.Result[Out]` splits `Text` (model-facing) from `Structured`
   (UI-facing). Return a non-nil `error` only for transport-level failure; return a
   `Result` with a UI-state field for domain "empty"/"permission". **Never panic
   across the MCP boundary.**
3. **Register** with the builder chain (typed generics):

   ```go
   tool.New[contracts.CreateDeckInput, contracts.CreateDeckOutput]("create_deck").
       Describe("Create a new deck (agent-first authoring entry point).").
       UI(appNamePreview).              // attaches result to the deck-preview surface
       Handler(h.createDeck).
       Register(srv)
   ```

   - `.UI(appName)` resolves the App name → its `ui://` URI and emits
     `_meta.ui.resourceUri` on the tool def (RFC §7.1). You never hand-build `_meta`.
   - For a UI-only action tool the model should not call directly (e.g. the editor's
     "save edits"), pass `tool.VisibilityApp`: `.UI(appName, tool.VisibilityApp)`.
   - **A `.UI("name")` naming no registered App is a loud error at `Register`** — a
     typo surfaces immediately, never a silent no-op.

4. **Declare** in `dockyard.app.yaml` `tools[]` (name, description, input, output,
   `ui:`, `task_support`).

Then `dockyard generate && dockyard validate && go test ./...`.

WorkBridge injects dependencies via a `ToolDeps` struct and a `RegisterTools(srv,
deps)` function (handlers hang off a `*handlers{deps}` receiver). Deckard does the
same: `ToolDeps{ Store *deck.Store; Renderer *render.Engine; Souls *soul.Registry;
Workspace string; Logger *slog.Logger }`. **A `Register(srv)` that receives no deps
is the stub-trap tell** (§5.5) — Deckard's registration must take its deps.

### 3.2 Three surfaces = three `apps[]` entries + three `apps.Register` calls

A Dockyard server serves **multiple `ui://` resources** simply by registering each
one. Each is an independent App with its own URI, bundle, CSP, and display modes;
tools point at whichever surface they drive via `.UI(appName)`.

The `runtime/apps` registration helper (confirmed in WorkBridge + the runtime
source):

```go
// internal/apps/register.go
package apps

import (
    dkapps "github.com/hurtener/dockyard/runtime/apps"
    "github.com/hurtener/dockyard/runtime/server"
)

const (
    NamePreview  = "deck-preview"
    NameOverview = "deck-overview"
    NameEditor   = "slide-editor"
)

// RegisterAll registers the three Deckard surfaces. MUST run BEFORE registerTools
// so .UI(name) can resolve each name to its ui:// URI.
func RegisterAll(srv *server.Server, preview, overview, editor []byte) error {
    if err := dkapps.Register(srv, dkapps.App{
        URI:   "ui://widget/deck-preview",
        Name:  NamePreview,
        Title: "Deckard Slides — Preview",
        HTML:  preview,
    }); err != nil { return err }
    if err := dkapps.Register(srv, dkapps.App{
        URI:   "ui://app/deck-overview",
        Name:  NameOverview,
        Title: "Deckard Slides — Overview",
        HTML:  overview,
    }); err != nil { return err }
    return dkapps.Register(srv, dkapps.App{
        URI:   "ui://app/slide-editor",
        Name:  NameEditor,
        Title: "Deckard Slides — Slide Editor",
        HTML:  editor,
    })
}
```

`apps.App` fields used: `URI`, `Name`, `Title`, `HTML`. Leave `Domain` **empty**
(host mints a per-conversation sandbox origin); only set it to a host-documented
verbatim string for a verified REMOTE/HTTP deployment — a stdio server ignores it
and logs a loud warning. `App.HostProfile`/`App.ServerURL` are deprecated; don't set
them.

### 3.3 `main.go` wiring — register Apps BEFORE tools

```go
//go:embed web/apps/deck-preview/dist/index.html
var previewHTML []byte
//go:embed web/apps/deck-overview/dist/index.html
var overviewHTML []byte
//go:embed web/apps/slide-editor/dist/index.html
var editorHTML []byte

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    srv, err := server.New(server.Info{
        Name: "go-slides-mcp", Title: "Deckard Slides", Version: "0.1.0",
    }, &server.Options{Logger: logger})
    // … construct deps (store, renderer, souls, workspace) …

    // 1) Apps FIRST (so .UI(name) resolves):
    must(deckardapps.RegisterAll(srv, previewHTML, overviewHTML, editorHTML))
    // 2) deck:// export resources (§4):
    must(exportstore.RegisterResources(srv, deps.Workspace))
    // 3) Tools LAST:
    must(handlers.RegisterTools(srv, deps))

    // 4) serve (stdio default; http when DOCKYARD_TRANSPORT=http)
    must(serve(ctx, srv, logger))
}
```

WorkBridge's `serve` reads `DOCKYARD_TRANSPORT` (`""`/`stdio` → `srv.ServeStdio(ctx)`;
`http` → `srv.HTTPHandler(nil)` on `127.0.0.1:8080`). The dev loop pins HTTP for the
inspector. `main.go` owns transport; the CLI never reimplements it.

> **`all:web/apps` embed gotcha.** If you instead embed the whole tree with
> `//go:embed all:web/apps` and `fs.ReadFile`, the `all:` prefix is **required** —
> without it, dotfiles/underscore files are skipped (RFC §14). Per-surface
> `//go:embed .../dist/index.html []byte` (the WorkBridge shape) sidesteps this and
> is clearer for exactly three surfaces.

### 3.4 The bridge handshake + host-theme propagation (Svelte side)

Each surface is a Svelte 5 single-file app. Hard-won rules (from
`dockyard-study-mcp` + WorkBridge):

- **`createBridge(...)` at the TOP LEVEL of `App.svelte`**, not inside `onMount` —
  `$state` effects get orphaned otherwise (`effect_orphan`, blank App). Only
  `await bridge.connect()` goes inside `onMount`.
- **`main.ts` uses Svelte 5 `mount(App, { target })`**, never `new App({target})`.
- `index.html` lives at the surface ROOT with `<div id="app"></div>` and
  `<script type="module" src="/src/main.ts">`.
- **Receiving data two ways:**
  - `bridge.onToolResult((payload) => …)` — the structured output of the tool that
    opened the surface (the glanceable preview path).
  - `bridge.callTool<I,O>(name, args)` — the surface calls tools itself. This is the
    **human↔agent handoff**: the overview/editor surfaces drive `reorder_slides`,
    `edit_slide`, `export_deck` directly so a human tweak goes through the exact
    same typed tools the agent uses. WorkBridge's `App.svelte` wraps `callTool` in a
    `{ok,data}|{ok:false,code,message}` helper that reads `res.isError` /
    `res.structuredContent` — copy that pattern.
- **Host theme:** read `hostContext.styles.variables`; the bridge auto-propagates it
  on the handshake (no per-call wiring). If a contract carries an explicit `theme`
  field, resolve `"auto"` server-side against `hostContext`. Deckard's design-system
  CSS reads these custom properties so the surfaces match the host's light/dark.
- **Four UI states** (loading/empty/error/permission/ready) on every surface, each
  exercised by a fixture, theme honored in each.
- **`dockyard-bridge` ≥ 1.6.1** (WorkBridge pins `^1.7.3`). Older builds spoke a
  non-spec handshake a strict host rejected and never reported size → host collapsed
  the iframe to ~0px (the blank-App saga, D-179/180/181).

### 3.5 Deny-by-default CSP + the iife/sandbox build

- **CSP:** `csp: { connect: [], resource: [] }` per surface. Empty lists = single-file
  bundle, no external origins — the deny-by-default sweet spot (RFC §7.4). The
  `csp` block models domain *allowlists* (`connect`→`connect-src`,
  `resource`→img/media/script/style/font-src, `frame`→`frame-src`); the literal CSP
  string is built by the host. There is **no manifest knob for `data:`/`blob:`
  media** — a single-file bundle inlines small assets as `data:` URIs but a host's
  deny-by-default CSP may block them, so **design slide thumbnails to degrade** to a
  placeholder/empty state when a `data:` image can't load.
- **Vite build (do NOT "tidy" it):** the App renders in a sandboxed iframe with
  `sandbox="allow-scripts"` and **no `allow-same-origin`**, where browsers refuse
  `<script type="module">`. The one shape that runs is an **iife** bundle with the
  `type="module"` attribute stripped. WorkBridge's `vite.config.ts` (copy verbatim):

  ```ts
  import { defineConfig, type Plugin } from 'vite';
  import { svelte } from '@sveltejs/vite-plugin-svelte';
  import { viteSingleFile } from 'vite-plugin-singlefile';

  function stripModuleType(): Plugin {            // rewrites <script type="module"> → <script>
    return { name: 'dockyard-strip-module-type', enforce: 'post',
      generateBundle(_o, bundle) {
        for (const fn of Object.keys(bundle)) {
          const f = bundle[fn];
          if (f.type !== 'asset' || !fn.endsWith('.html')) continue;
          const src = typeof f.source === 'string' ? f.source
            : new TextDecoder().decode(f.source as Uint8Array);
          f.source = src.replace(/<script([^>]*?)\stype="module"([^>]*)>/g, '<script$1$2>');
        }
      }};
  }
  export default defineConfig({
    plugins: [svelte(), viteSingleFile(), stripModuleType()],
    base: './',
    build: {
      outDir: 'dist', emptyOutDir: true,
      assetsInlineLimit: 100_000_000, cssCodeSplit: false, target: 'es2020',
      rollupOptions: { output: { format: 'iife', inlineDynamicImports: true } },
    },
  });
  ```

  Switching back to an ES-module build silently breaks the App in the host with **no
  build error** (note: `dockyard-study-mcp` shows the simpler module form — for the
  sandboxed-host case the iife+strip form is correct; prefer it).

### 3.6 Surface-opening discipline (the original UX failure to avoid)

The LOCKED decision: **do not default to opening a fullscreen editor.** Encode it:

- `create_deck`, `add_slide`, `reorder_slides`, `export_deck` → `ui: deck-preview`
  (inline, glanceable: thumbnails + sorter + `[download]` `[edit this]`).
- A structure/reorder request → `deck-overview` (inline+fullscreen).
- `slide-editor` (fullscreen) is **opt-in** — opened only by an explicit "edit this
  slide" action (the `[edit this]` quick action on the preview, or an explicit tool
  call). Document this in `docs/ui/SURFACES.md` and make the agent prompt + tool
  descriptions steer to the inline preview by default.

---

## 4. The `deck://` export resource scheme

Export delivery = **deterministic path + readable MCP resource**. `export_deck`
writes the `.pptx` to a deterministic workspace path, returns the absolute path in
its structured output, AND the bytes are exposed as a readable MCP resource under
`deck://`. This solves the original Claude-Code download/persistence pain.

The runtime API is `server.AddResource` / `server.AddResourceTemplate` (verified in
`runtime/server/resource.go`):

```go
// internal/exportstore/resource.go
func RegisterResources(srv *server.Server, workspace string) error {
    // A family of exports addressed by id — one template covers all decks.
    return srv.AddResourceTemplate(server.ResourceTemplateDef{
        URITemplate: "deck://export/{id}.pptx",     // RFC 6570; MUST carry a scheme
        Name:        "deck-export",
        Title:       "Deckard export (.pptx)",
        Description: "The exported PowerPoint for a deck, by deck id.",
        MIMEType:    "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    }, func(ctx context.Context, uri string) (server.ResourceContent, error) {
        id := parseDeckID(uri)                       // handler receives the concrete URI
        b, err := os.ReadFile(deterministicPath(workspace, id))
        if err != nil { return server.ResourceContent{}, err }
        return server.ResourceContent{
            MIMEType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
            Blob:     b,                             // Blob (binary) takes precedence over Text
        }, nil
    })
}
```

Key facts from the runtime source:
- `ResourceDef.URI` (or `ResourceTemplateDef.URITemplate`) **must be absolute (carry
  a scheme)** — the runtime validates the scheme and rejects a scheme-less URI; the
  `deck://` scheme satisfies this.
- `ResourceContent` carries **`Blob []byte`** (binary; precedence over `Text`) — the
  right field for `.pptx` bytes.
- `AddResource` is for a single fixed URI; `AddResourceTemplate` (RFC 6570) is for a
  family like `deck://export/{id}.pptx`. The handler receives the concrete URI the
  host requested.
- `export_deck`'s structured output should carry both `path` (absolute workspace
  path) and `resourceUri` (`deck://export/<id>.pptx`) so the preview surface shows a
  `[download]` that resolves the resource and the agent can cite the path. The
  workspace path must be **deterministic** (e.g. `<workspace>/exports/<id>.pptx`) so
  re-export is idempotent and the resource handler can find the file.
- The inspector's Verdicts/RPC tabs and `dockyard validate` cover resources too;
  resources also flow through the Logbook.

> Note: WorkBridge registered no custom resources (apps only), so `deck://` is
> Deckard-new — but it uses the public, documented `runtime/server` resource API,
> not a workaround.

---

## 5. Dev loop, inspector, packaging, and the GATE commands

### 5.1 Dev loop — `dockyard dev`

One process supervising a tree (embedded fsnotify; no `air`/`wgo`):

```bash
dockyard dev                       # from the project dir
# flags: --dir <path> --debounce 250ms --no-inspector --inspector-addr 127.0.0.1:0
```

- `.go` change → rebuild → restart server.
- `internal/contracts/*.go` or `dockyard.app.yaml` change → **regenerate** → rebuild
  → restart (generated types are live before the restart).
- `web/src/**` → Vite HMR (no server restart). `web/dist/` → ignored.
- **Auto-attaches the inspector** as a third supervised child; prints its URL to
  stdout. Pins the supervised server to HTTP `127.0.0.1:8080` (sets
  `DOCKYARD_TRANSPORT=http` + `DOCKYARD_HTTP_ADDR` as defaults you can override).
  `--no-inspector` keeps the stdio default (CI/headless).
- One Ctrl-C tears the whole tree down cleanly.

WorkBridge adds per-server `make dev-<name>` targets; Deckard (one server) just runs
`dockyard dev`. **Caveat for three Vite projects:** `dockyard dev` supervises the
project's `web/` Vite server; with three separate `web/apps/*` projects you iterate
one surface at a time, or run additional `vite build --watch` per surface (the
WorkBridge per-app `package.json` uses `"dev": "vite build --watch"`).

### 5.2 Inspector — `dockyard inspect`

Dev-mode-gated, localhost-only, operator-initiated (D-144). Either auto-attached by
`dockyard dev`, or standalone against a running HTTP server:

```bash
DOCKYARD_TRANSPORT=http DOCKYARD_HTTP_ADDR=127.0.0.1:8080 ./go-slides-mcp &
dockyard inspect --url http://127.0.0.1:8080 --dir .
# flags: --url (required) --dir (Verdicts+Fixtures) --port (loopback) --no-open
```

Tabs: **Tools** (Operator-Invoke: schema form + fixture switcher + Invoke →
`tools/call`), **Apps** (each `ui://` surface in a sandboxed iframe under its CSP),
**Tasks**, **Events** (live Logbook), **Analytics**, **Fixtures**, **Verdicts**
(re-runs `dockyard validate`), **RPC**, **Prompts**. Use it to fire each Deckard
tool against each fixture and watch the right surface render. Capability emulation
(toggle Apps/Tasks off) proves the server degrades to a working text-only response.
A non-loopback `--port` is refused before the listener opens.

> "Renders in a standalone serve" ≠ "works in the host." Validate each surface in
> the **inspector with fixtures** (the real MCP path), per the orchestrate-build
> live-validation rule.

### 5.3 Packaging — `dockyard build` / `dockyard run` / `dockyard install`

```bash
(cd web/apps/deck-preview  && npm install)   # one-time per surface
(cd web/apps/deck-overview && npm install)
(cd web/apps/slide-editor  && npm install)
(cd web/design-system      && npm install)

dockyard build                     # generate → validate → vite build → go build (host)
dockyard build --cross-compile     # darwin/linux/windows × amd64/arm64 + .sha256
dockyard build --output dist
dockyard run                       # build + run (stdio); --transport http --addr …
dockyard install claude            # write host MCP config + boot-check via initialize
dockyard install cursor
```

`dockyard build` order: `dockyard generate` → `dockyard validate` (a blocker fails
the build, so a stale contract never ships) → Vite `npm run build` per `web/` project
→ `go build` one CGo-free static binary (`CGO_ENABLED=0` always) with the bundles
embedded. **Clean-build before embedding** — a stale `web/apps/*/dist` cache serves
an old surface (orchestrate-build gotcha).

### 5.4 The GATE commands (Makefile + CI)

WorkBridge's canonical `make preflight` = `validate build test`. Deckard's
single-module equivalent (drop the `for s in $(SERVERS)` loop):

```make
generate:                ## regenerate contracts
	GOFLAGS="" dockyard generate

validate:                ## quality gates — fails on stale contracts / blockers
	GOFLAGS="" dockyard validate

build:                   ## CGo-free static binary with embedded UI
	GOFLAGS="" CGO_ENABLED=0 go build -o bin/go-slides-mcp .

test:                    ## go test -race ./...
	GOFLAGS="" go test -race ./...

vet:                     ## go vet ./...
	GOFLAGS="" go vet ./...

lint:                    ## pinned golangci-lint
	golangci-lint run

web:                     ## type-check + build every surface + design-system
	cd web/design-system     && npx svelte-check --tsconfig ./tsconfig.json && npx vite build
	cd web/apps/deck-preview && npx svelte-check --tsconfig ./tsconfig.json && npx vite build
	cd web/apps/deck-overview&& npx svelte-check --tsconfig ./tsconfig.json && npx vite build
	cd web/apps/slide-editor && npx svelte-check --tsconfig ./tsconfig.json && npx vite build

check-mirror:            ## AGENTS.md == CLAUDE.md
	diff -q AGENTS.md CLAUDE.md

preflight: generate validate build test  ## the CI gate
```

**Exact machine-checkable gate (the bottom line — green or it is NOT done):**

```bash
gofmt -l .                                   # MUST print nothing
GOFLAGS="" dockyard generate                 # idempotent — must produce no diff
git diff --exit-code internal/contracts      # MUST be clean (no stale codegen)
GOFLAGS="" dockyard validate                 # MUST exit 0 (0 blockers)
GOFLAGS="" CGO_ENABLED=0 go build ./...       # MUST compile
GOFLAGS="" go vet ./...                        # MUST be clean
GOFLAGS="" go test -race ./...                 # MUST pass with the race detector
GOFLAGS="" dockyard test                       # full contract+spec+capability gate, exit 0
golangci-lint run                              # pinned version (advisory→blocking, your call)
# UI surfaces:
cd web/design-system && npx svelte-check --tsconfig ./tsconfig.json
for s in deck-preview deck-overview slide-editor; do \
  (cd web/apps/$s && npx svelte-check --tsconfig ./tsconfig.json && npx vite build); done
```

**CI-pinned linter pattern.** WorkBridge does **not** ship a `.golangci.yml` and its
`make lint` runs `golangci-lint run 2>/dev/null || true` (advisory, per-module); its
*hard* gate is `gofmt -s` + `go vet` + `dockyard validate` + `go test -race`
(`make preflight`). For Deckard, harden this: commit a `.golangci.yml` and **pin the
linter by version in CI** so the gate is reproducible across the host and the
container:

```yaml
# .github/workflows/ci.yml (sketch)
- uses: actions/setup-go@v5
  with: { go-version-file: go.mod }   # pin the Go toolchain from go.mod
- run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8  # PIN the version
- run: make preflight                 # generate → validate → build → test
- run: golangci-lint run
```

Pinning the version (not `@latest`) is the pattern: a floating linter version
silently changes the gate between runs. Keep `GOFLAGS` empty (never `-mod=mod`).
`-race` needs `CGO_ENABLED=1`; shipped binaries are `CGO_ENABLED=0` (test binaries
are not shipped).

### 5.5 Per-unit `TASK.md` / `BUILD_PROMPT` gate block (for the autonomous builder)

Drop this verbatim into the builder's per-unit `TASK.md` and the global
`BUILD_PROMPT`. It encodes "verify, don't trust" (orchestrate-autonomous-build §7):

```text
USE THE DOCKYARD SKILLS. Do NOT reverse-engineer the toolchain or hand-write
schemas, TS types, _meta blocks, or the vite config. For task → skill:
  scaffold-a-server · define-contracts · add-a-tool · attach-a-ui-resource ·
  run-the-dev-loop · validate · test-with-the-inspector · package

A UNIT IS DONE ONLY WHEN ALL OF THESE EXIT 0 (paste the output):
  gofmt -l .                                  # empty
  dockyard generate && git diff --exit-code internal/contracts
  dockyard validate                           # 0 blockers
  CGO_ENABLED=0 go build ./...
  go vet ./...
  go test -race ./...
  dockyard test
  cd web/apps/<surface> && npx svelte-check --tsconfig ./tsconfig.json && npx vite build

THEN (orchestrator-verified, not builder-claimed):
  - Every tool in dockyard.app.yaml is registered AND its handler CALLS the real
    deck store / pptx-go renderer (grep for handlers returning literal/empty structs
    with no Store/Renderer call — that is the stub trap).
  - RegisterTools RECEIVES ToolDeps (a deps-less Register(srv) is a stub).
  - Each ui:// surface RENDERS in `dockyard inspect` against its fixtures (not a
    standalone serve), four states each, host theme honored.
  - export_deck writes the deterministic path AND deck://export/<id>.pptx reads back
    the same bytes.
Emit [goal:complete] ONLY after pasting the green gate output. The stop token is an
untrusted claim; the orchestrator advances on the gate, never the token.
```

The container/loop mechanics (from WorkBridge `.devcontainer/`): Go 1.26 + Node 22 +
`gh` + opencode image; `dockyard` installed at runtime from source with `GOPRIVATE`
+ retry; **Dockyard skills copied into the agent's skills dir** so the `skill` tool
can reach them; secrets injected read-only at `/run/secrets` (never baked); a
stateless fresh agent per iteration with a `[goal:complete]`/`[goal:blocked]` stop
token, primary+fallback model with rate-limit failover, and `run.sh` that
**recreates** the container per task (never `docker start` a stopped one). Publish
inspector + server ports to host loopback (`-p 127.0.0.1:7100:7100 -p
127.0.0.1:8080:8080`) for orchestrator-side Playwright validation against the live
inspector.

---

## 6. Deckard-specific adaptation summary (what differs from WorkBridge)

| Concern | WorkBridge | Deckard |
|--------|------------|---------|
| Modules | 12-binary `go.work` workspace | single module, single binary |
| `ui://` per server | one App, `display_modes:[inline,fullscreen]` | **three Apps**, three `apps.Register` calls |
| URIs | `ui://<server>/<app>/index.html` | `ui://widget/deck-preview`, `ui://app/deck-overview`, `ui://app/slide-editor` (opaque, honored verbatim) |
| Default surface | inline card | inline **deck-preview** (glanceable); fullscreen editor is opt-in only |
| Resources | apps only | adds `deck://export/{id}.pptx` via `server.AddResourceTemplate` (Blob bytes) |
| Backend | Microsoft Graph | pptx-go renderer + deck store + soul registry; no network, no Chromium |
| Design tokens | "warm-editorial" design-system | **Deckard White** built-in soul in `web/design-system` |
| Lint gate | advisory `|| true` | committed `.golangci.yml`, version-pinned in CI |

---

## Appendix — command quick reference

```bash
# scaffold (if starting fresh):  dockyard new go-slides-mcp --module github.com/hurtener/go-slides-mcp
dockyard generate            # contracts → JSON Schema + TS (idempotent)
dockyard validate            # fast build-blocker gate (drift, mappings, four-state)
dockyard test                # full contract+golden+spec+capability gate
dockyard dev                 # watch+rebuild+regenerate+vite HMR+auto inspector
dockyard inspect --url http://127.0.0.1:8080 --dir .
dockyard build [--cross-compile] [--output dist]
dockyard run [--transport http --addr 127.0.0.1:8080]
dockyard install claude | cursor
```
