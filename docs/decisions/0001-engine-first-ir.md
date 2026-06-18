# ADR-0001 — The IR mirrors the pptx-go engine, not the TS reference

**Status:** accepted · **Date:** 2026-06-18 · **Owner:** orchestrator

## Context

Deckard rewrites a TypeScript slides MCP server. The TS server's slide-IR node set was
shaped — and *limited* — by its rendering stack (a JS PPTX library + HTML/Playwright):
shallow nesting, HTML-era token lints, raster fallbacks, and bespoke "special" blocks that
existed because the library couldn't compose them natively.

Deckard renders with **pptx-go**, whose Layer-2 `scene` package is a typed, deterministic,
token-driven IR renderer with a rich, fully-implemented node catalog and **recursive
container composition**. Designing Deckard's IR by porting the TS nodes would inherit
limitations we no longer have and would lose engine capability.

## Decision

**Deckard's slide IR is a contract-first, JSON-native, agent-ergonomic mirror of pptx-go's
`scene` node catalog.** The authoritative source for the node set, fields, and validation
rules is the `compose-a-scene` and `define-a-theme` skills + the pptx-go `scene`/`theme`
package source — NOT the TS reference and NOT the secondhand research summaries.

Concretely:

1. **Cover the engine's full ceiling.** Every `scene.SlideNode` kind is exposed: all leaf
   nodes (incl. Flow diagrams, device-framed Images, Decorations, checklist Lists) and all
   container nodes (`TwoColumn`, `Grid`, `Card`, `CardSection`).
2. **Recursive nesting is in.** Containers hold `[]SlideNode`; arbitrary composition depth
   is supported. (The TS IR's single-level-nesting cap is dropped.)
3. **Semantic token colors, soul-as-theme.** Nodes reference token *roles* (color/type/
   spacing/elevation); a soul is a `*pptx.Theme`. Literal hex is the explicit escape hatch.
4. **Composed patterns, not new node types.** TS "special" blocks (timeline, kpi-cards,
   comparison, …) are delivered as **recipes** built from `flow`/`grid`/`card`. We do NOT
   add IR node types the engine cannot render — the engine's sealed union is the ceiling and
   the floor.
5. **The render adapter stays ~1:1.** Phase 3 maps Deckard IR → `scene` node by node; every
   IR kind has a `scene` counterpart, keeping rendering deterministic and high-fidelity.
6. **The TS reference is a coverage checklist only** — it tells us which agent-facing
   *capabilities/ergonomics* to provide (markdown compile, path edits, optimistic
   concurrency, comments), never the node schema.

## Consequences

- The IR design reads the engine skills/source first; "what did the TS server do" is a
  secondary check for capability coverage, not a schema template.
- `validate_slide_ir` mirrors `scene.ValidateScene`'s per-node rules.
- New visual patterns are added as recipes or by extending pptx-go upstream (icons/frames/
  ornaments via the registered-extension path) — never by inventing un-renderable IR nodes.
- "Enhance, don't simplify the output ceiling" (P8) is satisfied structurally: Deckard can
  express strictly more than the TS server (recursive composition, frames, flows, decorations).
