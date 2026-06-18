# Deckard

> **Deckard Slides** — an agent-first slides MCP App server. A pure-Go rewrite on the
> [Dockyard](https://github.com/hurtener/dockyard) framework that ships as one CGo-free
> static binary. Module `github.com/hurtener/go-slides-mcp`.

Deckard lets an agent author rich, highly-complex PowerPoint decks through typed MCP
tools — no UI required — and renders them to `.pptx` entirely in Go via
[pptx-go](https://github.com/hurtener/pptx-go) (no Playwright, no Chromium, no headless
browser). Humans review and tweak through three lightweight `ui://` surfaces that call
the *same* tools the agent uses.

## Principles

- **Agent-first.** MCP tools are the primary authoring path; the agent never needs a UI.
- **Pure Go, single binary.** pptx-go is the only `.pptx` producer; charts/code blocks are
  pure-Go rasters.
- **Three surfaces, never auto-fullscreen.** `deck-preview` (default inline glanceable),
  `deck-overview` (structure/reorder), `slide-editor` (opt-in, one slide).
- **Export just works.** Every export writes a deterministic workspace path **and** exposes
  it as a readable `deck://export/<id>.pptx` MCP resource.
- **Souls = bootstrap + refine.** Seed a complete design soul from natural language and/or a
  brand `.pptx`, then refine individual tokens. Default built-in soul: **Deckard White**.
- **Contract-first.** Typed Go structs are the source of truth; JSON Schema + TS are generated.

## Build & run

```bash
make preflight        # generate → validate → build → test (the CI gate)
make gate             # full local gate (green or it is NOT done)
dockyard dev          # watch + regenerate + rebuild + auto inspector
dockyard build        # one CGo-free static binary with all three surfaces embedded
```

## Where things live

- `docs/PLAN.md` — the phased implementation plan (binding acceptance criteria).
- `CLAUDE.md` (== `AGENTS.md`) — binding contributor & agent normatives.
- `docs/research/` — domain, engine, Dockyard-conventions, and design-token maps.
- `internal/` — contracts, IR, soul, render adapter, storage, handlers, the three apps.
- `web/` — the design system + the three Svelte surfaces.

## How it's built

Deckard is assembled with the **autonomous build loop** (`AUTONOMOUS_BUILD_LOOP.md`): a
capable orchestrator plans, writes reference units, and owns verification + git; a cheap
builder model running unattended in `.devcontainer/` fans out the clonable volume against
machine-checkable gates.
