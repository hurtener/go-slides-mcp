---
name: charts-and-code
description: "How to add data charts and code blocks to a Deckard deck with compile_chart and compile_code (pure-Go rasters, no browser). Use when a slide needs a chart, graph, or a code snippet."
---

# Charts & code

Deckard renders charts and code to images in **pure Go** (no browser, no external
service). You produce the image + node with one tool call, then drop the returned
node into a slide.

## Charts — `compile_chart`

```
compile_chart {
  spec: {
    type: "bar" | "line" | "pie",
    title: "p99 latency by quarter",
    labels: ["Q1","Q2","Q3","Q4"],          // categories (bar/pie)
    series: [ { name: "p99 (ms)", values: [120, 98, 61, 47] } ]
  },
  caption?: "Quarterly p99"
}
```

Returns a `chart` node (referencing the rendered image) + the asset id. Put the
node in a slide's `nodes`. Guidance:

- **Pick the type for the message:** trend over time → `line`; compare categories
  → `bar`; parts of a whole → `pie` (only with a few slices).
- **One chart per slide**, with a `heading` that states what it shows.
- Keep series/labels few and legible — a chart is a glance, not a spreadsheet.
- Prefer a `chart` over a big `table` when the point is the *trend* or *comparison*.

## Code — `compile_code`

```
compile_code {
  code: "func main() {\n\tprintln(\"hi\")\n}",
  language: "go",
  caption?: "main.go"
}
```

Returns a `code_block` node + asset id. Guidance:

- **Show the smallest snippet that makes the point** — trim imports/boilerplate.
- The renderer auto-fits the font to the slide, but very long lines still shrink;
  keep lines short and the snippet to a handful of lines.
- Use the `caption` for the filename or context.

## Anti-patterns

- A 12-series chart or a pie with 9 slices — illegible at slide size.
- Pasting a whole source file into one `code_block` — excerpt it.
- A chart with no heading saying what it proves.

## See also

- `composing-a-slide` — placing the `chart` / `code_block` node.
- `building-a-deck` — the authoring loop.
