# ADR-0002 — Surfaces are orchestrator-authored to a White-glove bar; multi-theme + selector

**Status:** accepted · **Date:** 2026-06-18 · **Owner:** orchestrator (Claude)

## Context

The product owner's explicit directive: "everything enters through the eyes." Both the **app
surfaces** and the **generated `.pptx` output** must impress on visual quality — this is a
headline acceptance criterion, not polish. The three `ui://` surfaces (Phases 8–9) are the
most judgment- and taste-sensitive work in the project.

## Decision

1. **Phases 8–9 (the three Svelte surfaces + the shared design system) are authored by the
   orchestrator (Claude), NOT delegated to the cheap builder.** The builder may do mechanical
   shell/clone scaffolding only after the orchestrator sets the reference; all visual design,
   interaction, motion, and polish are orchestrator-owned.
2. **White-glove quality bar.** Use the `frontend-design` skill. Distinctive, production-grade,
   intentional — never a generic AI-template aesthetic. Considered typography, spacing rhythm,
   color, micro-interactions, empty/loading/error states that feel designed, and motion that is
   purposeful. The bar is "impressive," explicitly.
3. **Multiple built-in themes + a theme/soul selector.** The design system ships **several**
   built-in souls — **"Deckard White"** (the warm-editorial design system, the renamed legacy
   aesthetic) as the default, plus at least a dark variant and 1–2 distinct alternates (e.g. a
   bold/high-contrast and a mono/minimal). Every surface exposes a **theme selector** that
   re-skins the surface live (via the soul tokens / `--app-*` CSS custom properties) AND can set
   the active deck soul. This expands Phase 2: ship >1 built-in soul + a dark-variant mechanism.
4. **Output-quality mandate.** The rendered deck must look premium too. After the render adapter
   (Phase 3) and rasterizers (Phase 7) land, the orchestrator renders representative decks under
   each built-in soul, **opens/screenshots them**, and iterates the soul tokens + render mapping
   until the output is genuinely impressive — verified by eye, not just "it renders."
5. **Naming.** The product uses no "pengui" anywhere; the aesthetic the owner calls "pengui" is
   the **Deckard White** soul. (ADR-0001 / the forbidden-name rule still hold.)

## Consequences

- Phase 2 gains a follow-on unit: additional built-in souls + the dark-variant approach.
- Phases 8–9 acceptance adds: the theme selector works and re-skins live; each surface meets the
  White-glove bar reviewed by the orchestrator in the inspector (and, where useful, Playwright
  screenshots); host-theme is still honored, with the soul selector layered on top.
- A visual-QA pass on the generated `.pptx` is a named deliverable, not optional.
- The orchestrator budgets real effort (and the `frontend-design` skill) for the surfaces rather
  than treating them as clone volume.
