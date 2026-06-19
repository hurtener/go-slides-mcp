# Deckard — Contributor & Agent Normatives

> **Binding** on every human and every agent (Claude, the GLM-5.2 builder loop, any
> subagent) that touches this repo. Mirrored **verbatim** to `AGENTS.md`
> (`make check-mirror` = `diff -q AGENTS.md CLAUDE.md`). On a conflict between two
> artifacts, fix the **lower-priority** one (see §2/§12) — never diverge silently.

**Deckard** (UI title *Deckard Slides*) is an agent-first slides MCP App server: a
full Go rewrite on the **Dockyard** framework, shipping as one CGo-free static binary.
Repo `go-slides-mcp`, module `github.com/hurtener/go-slides-mcp`.

## Orientation — read first, every iteration

Read in order: §1 (what Deckard is), §2 (sources + priority), §3 (layout), §8 (phase
workflow). Then the live references:

- `docs/PLAN.md` — the phased plan; **its acceptance criteria are binding**.
- `docs/research/{01-domain-map,02-engine-map,03-dockyard-conventions,04-design-tokens}.md`
  — the domain, the pptx-go engine surface + gaps, the Dockyard conventions, the
  Deckard White tokens. Cite these instead of reasoning from memory.
- `docs/contracts/CONVENTIONS.md`, `docs/ui/SURFACES.md`, `docs/decisions/*` (ADRs),
  `docs/glossary.md`.

You are a fresh, stateless session each loop iteration: re-orient from `git log`,
`git status`, `.devcontainer/TASK.md`, this file, and the plan. Assume nothing carried.

## §1 — What Deckard is (product invariants — non-negotiable)

- **P1 Agent-first.** MCP tools are the primary authoring path; the agent can build a
  whole deck without ever opening a UI surface. UI is for human review/tweak only.
- **P2 Slides only.** No A4 / document / long-form / print mode — ever. (Dropped.)
- **P3 Contract-first (Design A).** The typed Go struct is the single source of truth;
  JSON Schema + TypeScript are **generated** by `dockyard generate`; `dockyard validate`
  fails the build on drift. Never hand-write a schema, a `.ts`, a `_meta` block, or a
  vite config.
- **P4 Pure Go, single binary.** **pptx-go** is the *only* PPTX byte producer; CGo-free
  static binary; **no Playwright/Chromium**; no HTML→render stage; no measure-the-DOM
  layer. Charts and code blocks are pure-Go rasters fed through an `AssetResolver`.
- **P5 Three surfaces, never auto-fullscreen.** `ui://widget/deck-preview` (DEFAULT,
  inline, glanceable), `ui://app/deck-overview` (structure/reorder), `ui://app/slide-editor`
  (OPT-IN, one slide). The slide-editor opens only on an explicit "edit this slide".
- **P6 Export just works.** `export_deck` ALWAYS writes a deterministic workspace path
  AND exposes `deck://export/<id>.pptx` as a readable MCP resource. No `include_data` dance.
- **P7 Souls = bootstrap + refine.** One tool seeds a COMPLETE soul (all tokens) from
  natural language and/or a brand `.pptx`; targeted overrides refine it. The default
  built-in soul is **Deckard White**. (Easier than the reference's hand-author-7-layers.)
- **P8 Enhance, don't simplify, the output ceiling.** Authoring is easier to *drive*; the
  deck is no less capable. The `large.json` fixtures must exercise a high-complexity deck.

## §2 — Authoritative sources (priority chain)

```
RFC / design source of truth > docs/PLAN.md > CLAUDE.md (== AGENTS.md)
  > docs/contracts/CONVENTIONS.md > docs/ui/SURFACES.md > docs/decisions/*.md (ADRs) > code comments
```

When two disagree, fix the lower-priority artifact. Never silently ignore the conflict.

## §3 — Repository layout

Single Go module, single binary, three UI surfaces (see `docs/PLAN.md` §1.1 for the full
tree):

```
internal/{contracts,ir,soul,render,raster,deck,exportstore,handlers,apps,validate,fixtures}
web/{design-system, apps/{deck-preview,deck-overview,slide-editor}}
fixtures/<tool>/{happy,empty,error,permission,slow,large}.json
docs/{contracts/CONVENTIONS.md, ui/SURFACES.md, decisions/, research/, glossary.md}
main.go  dockyard.app.yaml  Makefile  .golangci.yml
```

- `internal/contracts/` is the contract-first source of truth; its `*.schema.json` / `.ts`
  are **generated** and never hand-edited.
- A new top-level directory requires a `docs/PLAN.md` (or RFC) update first.

## §4 — Build, test, lint, dev (canonical commands)

- `make {generate,validate,build,test,vet,lint,web,check-mirror,preflight}`.
- `dockyard {generate,validate,test,dev,inspect,build,install}`.
- `make preflight = generate validate build test` is the CI gate.
- `GOFLAGS` MUST stay empty (never `-mod=mod` — it breaks workspace mode silently).
  Shipped builds are `CGO_ENABLED=0`; tests run `-race`.
- Dev loop: `dockyard dev` (watch → regenerate → rebuild → restart, auto inspector).

## §5 — Code conventions (Go)

- Go pinned in `go.mod` (1.26); `gofmt -s` clean; `go vet` clean; `golangci-lint` (v2,
  pinned **v2.12.2**) clean.
- Errors: wrap with `%w`, compare with `errors.Is/As`, define sentinels; **never `panic`
  for control flow or across the MCP boundary** — return a typed `Result` with a UI-state
  field for domain "empty"/"permission".
- `context.Context` is the first param on I/O; honor cancellation (render respects it).
- `log/slog` only; **never log user content or asset bytes**.
- Concurrency: `-race` mandatory; **render determinism is a hard contract** — byte-identical
  output regardless of worker count.
- Stdlib `encoding/json`; replace `time.Time` in contracts with an ISO-8601 `string` or `int64`.

## §6 — Contract-first rules (Design A — non-negotiable)

- Never hand-write JSON Schema, TS types, `_meta` blocks, or the vite config.
- Document **every** contract field with a leading `//` comment → it becomes the schema
  `description` + TS JSDoc; a model reads these, they are product surface.
- `json:",omitempty"` for optionals; named scalar types for constrained values; a `Kind`
  discriminator on outputs that drive a multi-renderer surface.
- After `dockyard generate`, `git diff --exit-code internal/contracts` MUST be clean
  (no stale codegen). CrossCheck catches silent server↔UI drift at `validate`.
- A new tool ⇒ Go structs + `dockyard generate` + 6 fixtures + a contract test, same PR.

## §7 — The three-surface UI rules

- Three `apps[]` entries + three `apps.Register` calls; **register Apps BEFORE tools** in
  `main.go` so `.UI(name)` resolves (a `.UI()` naming no App is a loud error).
- URIs honored verbatim: `ui://widget/deck-preview`, `ui://app/deck-overview`,
  `ui://app/slide-editor`.
- **Default = deck-preview inline.** `slide-editor` opens ONLY on an explicit "edit this slide".
- Human↔agent handoff: surfaces call the SAME agent tools via `bridge.callTool` — **no
  `*_from_app` duplicates**; a human tweak and an agent edit are one operation.
- `createBridge` at the TOP LEVEL of `App.svelte` (never inside `onMount`); `main.ts` uses
  Svelte 5 `mount(App,{target})`; **iife + `stripModuleType` + singlefile** vite build
  (an ES-module build silently breaks in the sandboxed iframe); deny-by-default CSP
  (`connect:[], resource:[]`); host theme via `hostContext.styles.variables`; four UI states
  (loading/empty/error/permission) each backed by a fixture; design tokens come from
  `web/design-system` (Deckard White) — no hardcoded hex or copy.

## §8 — Phase workflow

- Per phase: read the phase in `docs/PLAN.md` (acceptance criteria binding) → read the
  cited research maps → read `CONVENTIONS.md` + `SURFACES.md` → **define contracts FIRST**
  → author fixtures → run the full gate (§11) before every commit.
- **USE THE SKILLS.** Do not reverse-engineer the toolchain or an engine API from memory:
  `scaffold-a-server`, `define-contracts`, `add-a-tool`, `attach-a-ui-resource`,
  `run-the-dev-loop`, `validate`, `test-with-the-inspector`, `package`; and the pptx-go
  engine skills `compose-a-scene`, `define-a-theme`, `load-a-brand-template`,
  `register-an-asset`, `embed-a-chart-raster`, `embed-a-code-block-raster`, `extend-the-icon-set`.
- **Reference-unit-before-fan-out:** a correct, clonable reference unit lands before the
  builder replicates it across N units.
- Definition of done = four polish passes: UI polish / wiring / no-stubs / inspector live-test.

## §9 — Tool-use discipline (the #1 cause of wasted builder iterations)

A single large `write` truncates mid-content and fails silently. Therefore:
- To CHANGE a file: small targeted `edit` calls, one block at a time; never rewrite a whole file.
- To CREATE a file: write a small skeleton (package + imports + one type/func), then GROW
  it with successive small edits. Keep every payload well under ~150 lines.
- Keep files small; split a big unit into several small files. If a write fails with a
  JSON/truncation error, STOP retrying the big payload — split it smaller.

## §10 — Forbidden practices

- The word formerly naming the white theme **anywhere** (the built-in soul is "Deckard White").
- **Playwright / Chromium / any headless browser**; any HTML→render or measure-the-DOM stage.
- **A4 / document / long-form / print** mode, formats, tokens, recipes, or nodes
  (`toc` / `bibliography` / `page_break` / `SectionIR`).
- Hand-written JSON Schema / TS / `_meta` / vite config (violates contract-first).
- **Large single writes**; god-files; god-commits.
- A stub handler that returns an empty/literal struct without calling a real dep (the stub
  trap); a `RegisterTools` / `Register(srv)` that receives no `ToolDeps`.
- Auto-opening the fullscreen editor by default; inventing a fourth surface or a new
  interaction paradigm.
- `panic` across the MCP boundary; CGo in shipped binaries; `GOFLAGS=-mod=mod`.
- Merging with `dockyard validate` failing or `internal/contracts` codegen stale.
- Logging user content or asset bytes; a surface missing empty/error states.
- Editing a generated `*.schema.json` / `.ts` by hand.
- Committing with the work account or its signing key; committing a token/secret or `_ref/`.

## §11 — Gate commands (green or it is NOT done)

```bash
gofmt -l .                                   # MUST print nothing
GOFLAGS="" dockyard generate                 # regenerate from Go contracts
GOFLAGS="" dockyard generate                 # idempotence = the stale-codegen check (2nd run: "no changes")
GOFLAGS="" dockyard validate                 # 0 blockers
GOFLAGS="" CGO_ENABLED=0 go build ./...
GOFLAGS="" go vet ./...
GOFLAGS="" go test -race ./...
GOFLAGS="" dockyard test                     # contract + spec + capability
golangci-lint run                            # pinned v2.12.2
diff -q AGENTS.md CLAUDE.md
# UI surfaces (when touched):
cd web/design-system && npx svelte-check --tsconfig ./tsconfig.json
for s in deck-preview deck-overview slide-editor; do \
  (cd web/apps/$s && npx svelte-check --tsconfig ./tsconfig.json && npx vite build); done
```

Orchestrator-verified (never builder-claimed): every manifest tool is registered AND its
handler calls a real dep; each touched surface renders in `dockyard inspect` against its
fixtures; `export_deck` proves path + `deck://` byte-equality. A `[goal:complete]` stop
token is an untrusted claim — advance only on the green gate.

## §12 — Priority chain on conflict (restate)

`RFC > docs/PLAN.md > CLAUDE.md (== AGENTS.md) > CONVENTIONS.md > SURFACES.md > ADRs > code comments`.
Fix the lower-priority artifact; never silently ignore the conflict.
