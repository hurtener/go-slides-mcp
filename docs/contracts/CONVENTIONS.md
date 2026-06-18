# Deckard contract & IR conventions (binding)

Source of truth, in order: this file → `docs/PLAN.md` → `CLAUDE.md`. The IR is
**engine-first** (ADR-0001): it mirrors pptx-go's `scene` catalog. The AUTHORITATIVE node
list, fields, and validation rules come from the **`compose-a-scene` skill** and the pptx-go
`scene` package — read the skill before touching IR code; do not invent fields the engine
lacks, and do not port the TS reference's schema.

## 1. Contract-first (Design A)

- Typed Go structs in `internal/contracts` are the source of truth. Run `dockyard generate`;
  never hand-write JSON Schema / TS / `_meta`. Use the `define-contracts` skill.
- Every exported field has a leading `//` comment (→ schema description + TS JSDoc).
- `json:"name,omitempty"` for optionals; named scalar string types for closed enums.
- No `time.Time` in contracts (use ISO-8601 `string` / `int64`).

## 2. The slide IR — JSON shape

A deck is slides; a slide is `{ id, layout, nodes[], notes? }`; a node is a JSON object with a
**`"kind"` discriminator** plus that kind's fields. `kind` values are `snake_case` and match
the engine node (e.g. `hero`, `heading`, `list`, `two_column`, `card`, `card_section`,
`section_divider`, `code_block`, `flow`, `decoration`). Field names are `camelCase`.

```json
{
  "id": "s1",
  "layout": "title_content",
  "nodes": [
    { "kind": "heading", "level": 2, "text": [{ "text": "Highlights" }] },
    { "kind": "two_column", "ratio": "1:1",
      "left":  [{ "kind": "card", "header": "A", "body": [ /* nodes */ ] }],
      "right": [{ "kind": "list", "listKind": "checklist",
                  "items": [{ "text": [{ "text": "ship it" }], "checked": true }] }] }
  ]
}
```

- `layout` ∈ `cover | title_content | two_column | card_grid | full_bleed | blank`
  (mirrors `scene.LayoutKind`).
- Container nodes (`two_column`, `grid`, `card`, `card_section`) hold child node arrays and
  **nest recursively** — depth is not capped.

## 3. The Go union codec (the one tricky mechanism — get it right)

`SlideNode` is an interface implemented by each concrete node struct:

```go
type SlideNode interface{ slideNodeKind() Kind } // sealed-ish: unexported marker method
```

- Each node struct has its fields + a `slideNodeKind()` returning its `Kind` constant.
- **Marshal:** a node marshals to its fields PLUS the injected `"kind"`. Implement one helper
  (e.g. `marshalNode(n SlideNode)`) used by every node's `MarshalJSON`, or a wrapper type —
  pick ONE consistent approach and document it; do not duplicate per node ad hoc.
- **Unmarshal:** a `kind → func() SlideNode` **registry** drives a single dispatch:
  `UnmarshalSlideNode([]byte) (SlideNode, error)` peeks `{"kind":…}`, constructs the concrete
  type, then unmarshals into it. Container fields are `[]SlideNode`; their `UnmarshalJSON`
  calls the SAME dispatch per child (recursion is the whole point).
- An **unknown `kind` is a hard error** (never silently dropped). Round-trip MUST be lossless
  for every kind, including deeply nested containers.

## 4. RichText

```json
"text": [
  { "text": "see ", "color": { "token": "secondary" } },
  { "text": "the docs", "link": true, "href": "https://…", "code": false }
]
```

- `RichText` = ordered `TextRun[]`. A run = `{ text, typeRole?, bold?, italic?, underline?,
  strike?, code?, link?, href?, color? }` (mirror `scene.TextRun`/`RunStyle`).
- `color` is `{ "token": "<textColorRole>" }` (semantic, soul-bound — the default) OR
  `{ "literal": "RRGGBB" }` (explicit escape hatch). Omitted = the token `primary`.
- Token roles mirror pptx-go: text color roles `primary|secondary|tertiary|inverse|muted|
  accent|accentAlt|success|warning|error`; type roles `body|h1|h2|h3|code|…` (use the
  `define-a-theme`/`compose-a-scene` skill enums verbatim — re-export, don't rename).

## 5. Structural-path addressing (for the Group-E edit tools)

A path is a JSON array mixing field legs (string) and indices (int), addressing into the node
tree from a slide's `nodes`:

- `["nodes", 2]` → the 3rd top-level node.
- `["nodes", 0, "left", 1]` → in node 0 (a `two_column`), the 2nd node of `left`.
- `["nodes", 1, "body", 0, "cells", 3]` → nested container descent.
- A trailing leaf field addresses a scalar/RichText (`[…, "text"]`, `[…, "header"]`).

Container child-field names per kind: `two_column → left|right`, `grid → cells`,
`card → body`, `card_section → body`. Define `Resolve(path)`, `Set`, `Insert`, `Remove`,
`Move` in `internal/ir/path.go` with table tests covering nesting + out-of-range errors.

## 6. Content hashing (optimistic concurrency)

`internal/ir/hash.go`: a slide/deck content hash = SHA-256 over its **canonical JSON**
(struct → `json.Marshal` with map keys sorted / deterministic field order; no transient
fields). Used as `expectedRevisionHash`. Determinism is a hard requirement: identical IR →
identical hash across processes.

## 7. Validation (mirror the engine)

`validate_slide_ir` / `internal/ir` shape checks mirror `scene.ValidateScene` per-node rules
(read `compose-a-scene` → "Validation"): heading level 1..6; list ≥1 item; callout valid
kind; image/chart/code_block non-empty `assetId` (+ image crop bounds); flow ≥1 step; table
row width == header width; two_column non-empty both sides; grid columns 2..4 + cell count a
multiple of columns; card/card_section children valid + `card_section.body` non-empty. Return
a JOINED error (report every problem at once), like the engine.

## 8. Forbidden in IR

No `toc` / `bibliography` / `page_break` / `section` / `SectionIR` / A4 / print / document
types. No node kind without a `scene` counterpart (every kind must render in Phase 3). No
literal hex except via the RichText `literal` escape hatch and node color fields the engine
itself exposes literally.
