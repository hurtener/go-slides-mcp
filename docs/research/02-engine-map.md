# 02 — Engine Map: pptx-go (the Deckard PPTX engine)

Source studied: clone at `/tmp/pptx-go-clone`, import path
`github.com/hurtener/pptx-go`. RFC: `RFC-001-pptx-go.md`. Skills:
`skills/{scaffold-a-presentation,compose-a-scene,define-a-theme,load-a-brand-template,register-an-asset,embed-a-chart-raster,embed-a-code-block-raster,extend-the-icon-set}/SKILL.md`.

This document maps what pptx-go can do so the Deckard (go-slides-mcp) rewrite can
treat it as the single rendering engine behind all MCP tools. **pptx-go is the
engine, not the product** (RFC D-026): it converts a typed description into valid
OOXML and nothing else — no render modes, no legibility heuristics, no markdown
ingestion, no rasterization, no HTML stage. That maps cleanly onto our pure-Go,
no-Chromium constraint: every product behavior lives in Deckard, every PPTX byte
comes from pptx-go.

---

## 0. The two-layer shape (what we call from Deckard)

```
scene  (Layer 2, PUBLIC) — typed IR → PPTX. Composes pptx only.
   │
pptx   (Layer 1, PUBLIC) — token-aware builder: Presentation/Slide/Shape/
   │                        TextFrame/Table/Image/Theme. General-purpose.
internal/opc + internal/ooxml (PRIVATE) — raw OOXML/OPC wire types.
```

Binding properties relevant to us (RFC §4):
- **P2 Tokens, not literals.** Every visual property resolves through a `*Theme`
  at *apply time*. A theme swap re-skins identical input. This is exactly the
  substrate for our "design souls."
- **P4 No CGo, stdlib-only runtime.** Zero third-party Go deps; curated assets
  embedded via `go:embed`. Single static CGo-free binary is achievable.

Deckard will primarily drive **Layer 2 (`scene.Render`)** — it is the IR-shaped
surface. We drop to **Layer 1 (`pptx`)** only for things scene doesn't express
(e.g. placing a brand logo image per slide, direct shape work, font embedding).

---

## 1. High-level authoring API surface

### 1.1 Scaffold a presentation (Layer 1 — `pptx`)

Real signatures (`pptx/presentation.go`, `pptx/template.go`):

```go
func New(opts ...Option) *Presentation
func NewWithTemplate(name TemplateType) (*Presentation, error)
func NewFromBytes(data []byte, opts ...Option) (*Presentation, error)
func NewFromFile(path string, opts ...Option) (*Presentation, error)
func OpenStream(path string, opts ...Option) (*Presentation, error) // lazy per-part read

func (p *Presentation) AddSlide(layout ...string) *Slide
func (p *Presentation) AddSlideAt(index int, layout ...string) (*Slide, error)
func (p *Presentation) Slides() []*Slide
func (p *Presentation) SlideCount() int
func (p *Presentation) GetSlide(index int) (*Slide, error)
func (p *Presentation) RemoveSlide(index int) error
func (p *Presentation) Theme() *Theme
func (p *Presentation) SetTheme(t *Theme)
func (p *Presentation) AddSection(name string) *Section
func (p *Presentation) Sections() []*Section
func (p *Presentation) Save(path string) error
func (p *Presentation) Write(w io.Writer) error      // e.g. an HTTP response
func (p *Presentation) WriteToBytes() ([]byte, error)
func (p *Presentation) SaveStream(path string) error // streaming OPC writer
func (p *Presentation) Close() error
```

`Option`s (`pptx`): `WithFormat(Slides16x9 | Slides4x3)`, `WithTheme(*Theme)`,
`WithLogger(*slog.Logger)`, `WithFontSource(FontSource)`,
`WithReadPartLimit(int64)`, `FromTemplate(brand *Presentation)`.

`New()` with no options yields a complete, valid 16:9 deck (master + blank layout
+ theme, relationships wired) that opens in PowerPoint without a repair prompt.
Named built-in templates: `TemplateBlank`, `TemplateDefault`, `TemplateWide`,
`TemplateStandard`.

Note: there is **no A4/document/portrait format** — only `Slides16x9` and
`Slides4x3`. This aligns with the locked decision to drop long-form/document mode
entirely. (RFC §5: print formats are explicitly out of scope.)

### 1.2 Primitives (Layer 1)

All geometry is `pptx.Box{X, Y, W, H pptx.EMU}` from the slide top-left. Units:
`pptx.In/Cm/Pt/Px(float64) EMU`. Canvas constants `pptx.Slide16x9Width/Height`,
`pptx.Slide4x3Width/Height`.

**Shapes** (`pptx/shape.go`):
```go
func (s *Slide) AddShape(geom ShapeGeometry, box Box, opts ...ShapeOption) *Shape
```
`ShapeGeometry` (a string enum): `ShapeRect`, `ShapeRoundRect`, `ShapeEllipse`,
`ShapeTriangle`, `ShapeDiamond`, `ShapeParallelogram`, `ShapeHexagon`,
`ShapeChevron`, `ShapeRightArrow`, `ShapeLine`. (The RFC mentions a far larger
OOXML preset list and a reserved `CustomGeom`; only the above are shipped today —
see Gaps.)
`ShapeOption`: `WithFill(Fill)`, `WithLine(Line)`, `WithRadius(RadiusRole)`,
`WithElevation(ElevationRole)`, `WithShadow(Elevation)`, `WithRotation(float64)`.
Fills: `SolidFill(Color)`, `NoFill()`, `LinearGradient(angleDeg, stops...)`,
`RadialGradient(stops...)`.

**Text / rich text** (`pptx/text.go`):
```go
func (s *Slide) AddTextFrame(box Box) *TextFrame
func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph
func (tf *TextFrame) AutoFit(mode AutoFitMode) *TextFrame  // None | Normal | Shape
func (tf *TextFrame) Anchor(v TextAnchor) *TextFrame       // Top | Middle | Bottom
func (tf *TextFrame) Margins(top, right, bottom, left EMU) *TextFrame
func (p *Paragraph) AddRun(text string, style RunStyle) *Run
func (p *Paragraph) AddBreak()
func (p *Paragraph) Align(a Alignment) *Paragraph          // Left/Center/Right/Justify
func (p *Paragraph) Indent(level int) *Paragraph
func (p *Paragraph) Bullet(kind BulletKind) *Paragraph     // None/Disc/Number/Checkbox
```
`RunStyle{TypeRole TypeRole; Color Color; Bold, Italic bool; Underline; Strike;
BaselineRel; Code bool}` — fully token-typed. (Hyperlinks: scene `RunStyle` adds
`Link bool` + `Href string`; builder-level hyperlink support lives in
`pptx/text_hyperlink.go`.)

**Tables** (`pptx/table.go`):
```go
func (s *Slide) AddTable(box Box, rows, cols int) *Table
func (t *Table) Cell(row, col int) *Cell
func (t *Table) SetHeaderRow(on bool) *Table
func (t *Table) SetBanding(rowBand, colBand bool) *Table
func (t *Table) SetColumnWidths(widths ...EMU) *Table
func (c *Cell) SetText(text string) *Cell
func (c *Cell) TextFrame() *TextFrame
func (c *Cell) SetFill(f Fill) *Cell
func (c *Cell) SetBorders(line Line) *Cell
func (c *Cell) MergeRight(n int) *Cell
func (c *Cell) MergeDown(n int) *Cell
```
Header rows, banding, merges are first-class and emit concrete alternating fills
(no dependency on a table-style part).

**Images** (`pptx/media.go`):
```go
func (s *Slide) AddImage(src ImageSource, box Box) (*Image, error)
func ImageBytes(data []byte, mime string) ImageSource
func ImageFile(path string) ImageSource
func ImageReader(r io.Reader, mime string) ImageSource
func (im *Image) SetAltText(text string) *Image
func (im *Image) SetCrop(c Crop) *Image          // per-edge fractions 0..1
func (im *Image) SetFit(f Fit) *Image            // FitFill | FitNone
func (im *Image) SetRotation(deg float64) *Image
func (im *Image) SetOpacity(alpha int) *Image    // 0..100000
```
PNG/JPEG/GIF/BMP/WebP recognized by magic bytes; malformed/mismatched bytes
rejected (`ErrUnknownImageFormat`, `ErrImageMIMEMismatch`). Identical bytes are
deduplicated (written once). Pixel data is never parsed.

**Icons / ornaments / frames** — not Layer-1 primitives; they are scene-layer
curated registries (see §1.4 and §4). Icons render as **native PPTX path
geometry** (single-path SVG → OOXML), not rasters.

**Speaker notes & sections** (`pptx/notes.go`, `pptx/section.go`):
```go
func (s *Slide) SpeakerNotes() *TextFrame
func (s *Slide) SetSpeakerNotes(text string)
func (s *Slide) HasSpeakerNotes() bool
func (sec *Section) Include(s *Slide)
func (sec *Section) Name() string
```

**Charts / code blocks** — no native Layer-1 chart in V1. Charts and code are
caller-rasterized images (see §4). Native `c:chart` is a V2 backlog item.

### 1.3 Compose a scene (Layer 2 — `scene`)

Entry point (`scene/scene.go`):
```go
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error)

type Scene struct {
    Theme  *pptx.Theme   // optional; nil = builder default theme
    Slides []SceneSlide
    Meta   Metadata      // Title, Author, Subject → docProps/core.xml
}
type SceneSlide struct {
    ID      string        // label used in warnings/timings
    Layout  LayoutKind    // structural intent → master layout
    Nodes   []SlideNode   // top-level node tree
    Notes   RichText      // speaker notes
    Variant Variant       // VariantLight | VariantDark | VariantPrint
}
```
`LayoutKind`: `LayoutCover`, `LayoutTitleContent`, `LayoutTwoColumn`,
`LayoutCardGrid`, `LayoutFullBleed`, `LayoutBlank`.

`RenderOption`s (`scene/scene.go`): `WithTheme`, `WithWorkers(int)`,
`WithLogger`, `WithContext`, `WithLayoutMap(LayoutMap)`,
`WithAssetResolver(AssetResolver)`, `WithIconExtension(name, svg)`,
`WithFrameExtension(name, recipe)`, `WithOrnamentExtension(name, recipe)`.

`Render` is **deterministic**: same scene + theme = byte-identical PPTX
regardless of worker count (a hard requirement — supports our snapshot/diff
tooling). It validates Stage 1 first, then lays out and composes.

The scene renderer **has a layout engine**; Layer 1 does not. In `scene` you give
a node tree and the renderer places boxes (priority-ordered: decorations →
hero → body in IR order → foreground decorations → section dividers). Overflow is
a non-fatal `LayoutWarning`, not an error. There is no constraint solver and no
strict mode (RFC §10.2).

### 1.4 Scene node catalog (the IR primitives) — `scene/nodes.go`

`SlideNode` is a **sealed union** (unexported marker `isSlideNode()`), so the set
is closed to the package — we construct concrete structs and cannot inject
arbitrary nodes. `NodeKind` discriminates. **Exactly the nodes that render as a
picture carry an `AssetID`** — `Image`, `Chart`, `CodeBlock`, asset-kind
`Decoration`; every other node renders as native PPTX shapes (RFC §12, D-018).

Leaf nodes: `Hero{Eyebrow,Title,Subtitle}`, `Prose{Paragraphs []RichText}`,
`Heading{Text RichText; Level int}`, `List{Kind; Items []ListItem}`,
`Divider{Spacing SpaceRole}`, `Quote{Text RichText; Attribution string}`,
`Callout{Kind; Title; Body RichText}`, `Image{AssetID; Alt; Frame; FrameName;
Crop; Fit}`, `Chip{Label; Tone; Color ColorRole}`, `Arrow{Direction; Label}`,
`CodeBlock{AssetID; Language; Caption}`, `Chart{AssetID; Caption}`,
`Table{Headers []RichText; Rows [][]RichText; Caption}`,
`Flow{Orientation; Steps []FlowStep; Connector}`,
`Decoration{Kind; Preset; AssetID; Layer; Anchor; Offset; Size; Bleed; Opacity;
Rotation}`, `SectionDivider{Eyebrow; Label}`.

Container nodes: `TwoColumn{Ratio; Left, Right []SlideNode}`,
`Grid{Columns int; Ratio []int; Gap SpaceRole; Cells []SlideNode}`,
`Card{Header,Eyebrow,Icon,HeaderPill string; Body []SlideNode; BodyLayout; Fill
ColorRole; Outline bool; BorderStyle; Size; Layout; Elevation}`,
`CardSection{Header string; Body []SlideNode}`.

**Implementation status:** every node kind is wired in `scene/render.go`'s
dispatch to a real renderer (`render_leaves.go`, `render_card.go`,
`render_card_section.go`, `render_container.go`, `render_table.go`,
`render_chart.go`, `render_code_block.go`, `render_image.go`, `render_flow.go`,
`render_decoration.go`). The default "not yet implemented; node skipped" warn
branch only fires for an unrecognized kind. **The catalog is fully rendered
today** — this is a high output ceiling we can drive immediately.

RichText (`scene/richtext.go`):
```go
type RichText []TextRun
type TextRun struct { Text string; Style RunStyle; Color TextColor }
type RunStyle struct { TypeRole TypeRole; Bold,Italic,Underline,Strike,Code,Link bool; Href string }
func TokenTextColor(role TextColorRole) TextColor // theme-bound (default path)
func LiteralColor(hex string) TextColor           // escape hatch
```
The scene token enums (`scene/tokens.go`) are **type aliases** of the `pptx`
enums (`type ColorRole = pptx.ColorRole`, etc.), so the IR and builder share one
vocabulary.

---

## 2. Themes, brand templates, and the soul bootstrap+refine path

### 2.1 The Theme object (`pptx/theme.go`)

```go
type Theme struct {
    Name        string
    HeadingFont string       // theme1.xml "major" face
    BodyFont    string       // theme1.xml "minor" face
    Colors      ColorPalette // Surfaces map[ColorRole]RGB + Text map[TextColorRole]RGB
    Typography  Typography    // map[TypeRole]FontSpec{Family,Size,Weight,Italic}
    Spacing     Spacing       // map[SpaceRole]EMU
    Radii       Radii         // map[RadiusRole]EMU
    Elevations  Elevations    // map[ElevationRole]Elevation{Blur,OffsetX,OffsetY,Color,Alpha}
}
```
Token taxonomy (the full set of resolvable roles):
- **Surfaces (10):** Canvas, Surface, SurfaceAlt, Accent, AccentAlt, AccentWarm,
  Success, Warning, Error, Info.
- **Text (10):** Primary, Secondary, Tertiary, Inverse, Muted, Accent, AccentAlt,
  Success, Warning, Error.
- **Type (11):** Display, H1–H5, Body, BodySmall, Caption, Mono, Code.
- **Spacing (6):** XS, SM, MD, LG, XL, 2XL.
- **Radius (5):** None, SM, MD, LG, Full.
- **Elevation (3):** Flat, Raised, Elevated.

Resolution is **deterministic and at apply time**; unset roles fall back to safe
neutrals (surfaces → FFFFFF, text → 000000, type → Calibri 14/400, spacing/radius
→ 0, elevation → flat) — `Resolve*` never panics. Public resolvers:
`ResolveColor`, `ResolveTextColor`, `ResolveType`, `ResolveSpace`,
`ResolveRadius`, `ResolveElevation`.

`Color` is a sealed interface: `TokenColor(role)`, `TokenColorAlpha(role, alpha)`,
`TokenTextColor(role)`, and the literal escape hatches `RGB("2563EB")` /
`RGBA(rgb, alpha)`.

### 2.2 define-a-theme (powers soul **refine**)

Three constructors + one mutator (`pptx/theme.go`):
```go
func DefaultTheme() *Theme            // complete legible light theme, no embedding
func NewTheme(opts ...ThemeOption) *Theme
func (t *Theme) Clone() *Theme        // deep copy; every map reallocated
```
`ThemeOption`: `WithName(string)`, `WithAccent(RGB)`,
`WithFonts(heading, body string)` (rewrites Typography families: heading roles
get `heading`, mono roles untouched, the rest get `body`).

Refinement model for souls: `DefaultTheme().Clone()` then mutate any token map
directly (e.g. `t.Colors.Surfaces[pptx.ColorCanvas] = pptx.RGB("0B1220")`,
`h1 := t.Typography[pptx.TypeH1]; h1.Size = 36; t.Typography[pptx.TypeH1] = h1`).
This is the engine substrate for "targeted token overrides refine the soul."

`DefaultTheme()` ships: heading `Calibri Light`, body `Calibri`, mono `Consolas`;
accent `2563EB`; full palette/type/spacing/radius/elevation maps. We replace this
default with our **"Deckard White"** built-in soul (same structure, our values).

### 2.3 load-a-brand-template (powers soul **bootstrap** from a .pptx)

```go
func pptx.NewFromFile(path string, opts ...Option) (*pptx.Presentation, error)
func pptx.NewFromBytes(data []byte, opts ...Option) (*pptx.Presentation, error)
func pptx.FromTemplate(brand *Presentation) Option   // seed theme+masters+layouts
func (p *Presentation) Theme() *Theme
func (p *Presentation) Masters() []*Master           // m.Name(), m.Layouts() -> []*Layout
func (p *Presentation) HasLayout(name string) bool
```
On open of a brand `.pptx`, pptx-go extracts `theme1.xml` (colors + major/minor
fonts) and the master/layout registry. `pptx.New(pptx.FromTemplate(brand))`
**clones** the brand package (theme + masters + layouts + auxiliary parts) and
strips slides, so a new deck starts slide-free with the brand look. Cloning the
already-valid relationship graph is what avoids the repair-prompt bug class. The
brand presentation is not retained or mutated (close it after `New`).

**Inputs a brand template provides:** colors (theme1.xml palette), fonts (major
/minor face names), named slide layouts/masters. **It does NOT carry** spacing,
radius, or elevation tokens, and **logo placement is not part of the Theme**.

### 2.4 Mapping our soul bootstrap+refine onto this

- **Bootstrap from natural language:** Deckard synthesizes a `*pptx.Theme` in Go
  (Clone the Deckard White default, set palette/fonts/spacing/radius/elevation
  from the NL spec). No engine call needed beyond `NewTheme`/`Clone`.
- **Bootstrap from a brand .pptx:** `NewFromFile` + read `Theme()` to seed
  colors/fonts; then enrich the missing token families (spacing/radius/elevation)
  from our defaults. Logo: extract/accept separately and place it as an `image`
  node (or Layer-1 `AddImage`) — it is not a Theme field.
- **Refine:** Clone + targeted map writes. One swapped `*Theme` re-skins the whole
  deck because resolution is at apply time and everything authored through tokens.

### 2.5 Font embedding (mechanism, caller-driven)

`FontSource` interface + `pres.EmbedFont(name, style, weight)` (RFC §7.6). No
auto-embed: Deckard registers a `FontSource` and calls `EmbedFont` for each
soul-referenced face it wants embedded. CGo-free; full-face embed in V1
(subsetting is V1.x).

---

## 3. Concurrency / streaming (fast generation)

From RFC §17 and `scene/scene.go`:
- **Parallel slide compose.** `scene.Render` creates all slides in scene order
  (fixing slideN.xml numbering and order under the presentation lock), then fans
  media-free slides across a worker pool sized to `runtime.GOMAXPROCS(0)`
  (`WithWorkers(n)`; `n<=1` forces sequential). **Media-bearing slides render
  sequentially in scene order** so media part numbering — and therefore the bytes
  — stay deterministic. Output is byte-identical regardless of worker count.
- **Cancellation.** `WithContext(ctx)` — the `AssetResolver` receives it and
  `Render` honors cancellation between slides (returns `ctx.Err()`).
- **Lazy asset resolution.** Resolver called per asset on first reference;
  failures surface as `LayoutWarning` (unless required → render fails).
- **Streaming I/O.** `OpenStream` / `SaveStream` are first-class (lazy per-part);
  large decks (hundreds of slides, >50MB) work without full in-memory load.
  Underlying `StreamPackage`, `ConcurrentZipCollector`, `ConcurrentStreamSave`,
  and the `sync.Map`-based media dedup pool live in `internal/opc`.
- **Reusable artifacts** (themes, asset/icon/ornament/frame registries, masters)
  are read-only after construction and safe for concurrent use; a `*Presentation`
  is single-writer.
- **Observability.** `WithLogger(*slog.Logger)` (zero-cost when nil) + the `Stats`
  return (`Slides`, `Shapes`, `Assets`, `Warnings []LayoutWarning`,
  `Timings []SlideTiming` per-slide wall-clock).

Implication for Deckard: one MCP "render/export" call can render a large deck
in parallel deterministically, stream it to a workspace path, and hand back
`Stats` for preview/telemetry — no external process, no Chromium.

---

## 4. IR-node → scene-primitive mapping sketch

Assume Deckard's slide IR nodes map to `scene.SlideNode` values (the scene IR is a
strict superset of the prior product's v4 IR — RFC §21, one-to-one per node). The
engine renders as follows (RFC §12 policy table, confirmed against the wired
renderers in `scene/render.go`):

| Deckard IR node | scene node(s) | Renders as | Asset? | Notes |
|---|---|---|---|---|
| hero / title | `Hero` | native text shapes | — | eyebrow + title + subtitle |
| paragraph / body | `Prose` | native text | — | `Paragraphs []RichText` |
| heading | `Heading{Level 1..6}` | native text | — | |
| bullets / numbered / checklist | `List{Kind}` | native text + bullet props | — | `ListBullet/ListNumber/ListChecklist` |
| divider / rule | `Divider{Spacing}` | native thin shape | — | |
| quote / pullquote | `Quote` | native text shapes | — | + attribution |
| callout / admonition | `Callout{Kind}` | native rect + icon + text | — | Note/Warning/Tip/Important |
| chip / tag / pill | `Chip{Tone,Color}` | native roundrect + text | — | Tint/Solid/Outline |
| arrow | `Arrow{Direction,Label}` | native preset arrow | — | |
| image / screenshot | `Image{AssetID,Frame,Crop,Fit}` | **pic** | yes | optional device frame chrome |
| chart | `Chart{AssetID,Caption}` | **pic** (contain-fit) | yes | caller rasterizes; native c:chart is V2 |
| code block | `CodeBlock{AssetID,Language,Caption}` | **pic** + native lang badge/caption | yes | caller rasterizes (D-014) |
| table | `Table{Headers,Rows,Caption}` | native `tbl` | — | header row + caption shape |
| flow / process / pipeline | `Flow{Orientation,Steps,Connector}` | native step pills + connectors | — | Arrow/Dashed/Cycle/Plus connectors; steps carry icons |
| timeline / KPI / comparison | — | — | — | **GAP** (RFC §11.3 extensions, not shipped) |
| decoration / ornament | `Decoration{Kind=Preset}` | native preset SVG geom | — | curated ornament set |
| decoration (image bg) | `Decoration{Kind=Asset}` | **pic** w/ bleed offsets | yes | |
| section break / chapter | `SectionDivider{Eyebrow,Label}` | native full-bleed shape | — | overrides body to full-bleed |
| two-column layout | `TwoColumn{Ratio,Left,Right}` | container (children render per policy) | — | 1:1 / 1:2 / 2:1 |
| grid layout | `Grid{Columns 2..4,Ratio,Gap,Cells}` | container | — | weighted column ratios |
| card | `Card{...}` | native roundrect + accent stripe + icon/header/pill + body | — | Fill/Border/Size/Elevation/Layout |
| card section | `CardSection{Header,Body}` | native card accepting grid/two_column/nested cards | — | |
| device frame around image | `Image.Frame` / `FrameName` | native group around the pic | via inner | browser/phone/desktop/laptop |
| inline icon (in card/flow) | `Card.Icon` / `FlowStep.Icon` | native path geometry | — | closed-name curated set + `WithIconExtension` |
| speaker notes | `SceneSlide.Notes` | notesSlide part | — | RichText |
| deck sections (sorter groups) | Layer-1 `pres.AddSection` | OOXML sectionLst | — | **no scene-IR field** — drive via Layer 1 |

Curated asset registries (RFC §14): **icons** ≈60 lucide-style, native path
geometry (SVG translator subset: single path, solid fill, no gradients, no
elliptical arcs — `pptx.ValidateIcon`); **ornaments** glow_ring, radial_glow,
grid_dots, corner_bracket, chevron_arrow, noise_overlay; **frames** browser,
phone, desktop, laptop. All extensible per-render by name registration.

The rasterization contract (RFC §12.3): for every `AssetID`-bearing node, Deckard
pre-rasterizes (our chart renderer, our code highlighter — both **pure Go**, no
Playwright) and serves bytes through an `AssetResolver`
(`Resolve(ctx, id) ([]byte, contentType, error)`). pptx-go never invokes a
rasterizer. `URIAssetResolver(fn)` handles `asset://<uuid>` ids.

---

## 5. Gaps the engine cannot yet express (handle in the rewrite)

1. **Native charts.** V1 charts are rasters only. Editable `c:chart` is V2
   backlog. Deckard must own a pure-Go chart rasterizer (SVG/PNG) and feed bytes.
   Acceptable for our "no Chromium" constraint; flag as a future capability.
2. **Code blocks are rasters.** No native monospace text block; Deckard must
   rasterize code (pure-Go highlighter → PNG). Whitespace/metrics can't survive as
   native text by design (D-014).
3. **Theme variants (dark/print) not implemented.** A non-`VariantLight` slide
   renders with the active theme and emits a `LayoutWarning`. If souls need
   light/dark per slide, Deckard must render with the appropriate full `*Theme`,
   not rely on `Variant`. (Per-slide theme override is also V2.)
4. **In-code themes don't persist to theme1.xml.** A brand built in code with
   `WithTheme`, saved and re-opened, adopts the default scaffold theme, not the
   in-memory tokens (token emission to theme1.xml is pending). Consequence: soul
   bootstrap from a *PowerPoint-authored* brand .pptx works fully; round-tripping
   a *code-authored* brand does not yet preserve custom tokens. Keep the soul as
   our own source of truth and re-apply it per render rather than relying on a
   saved .pptx to carry it.
5. **Logo is not a Theme field.** Brand logo must be placed as an `image` node or
   Layer-1 `AddImage` per slide/layout. Soul bootstrap must store the logo asset
   separately and inject it; the engine won't auto-place it.
6. **Image fit is FitFill/FitNone only.** No aspect-aware cover/contain (needs
   pixel dims, forbidden by P4 — no image decoding). `Chart` does an internal
   contain-fit from sniffed dims, but generic `Image` cover/contain is on us
   (pre-crop/size before handing bytes).
7. **§11.3 "and more" nodes not shipped:** `timeline`, `kpi_cards`, `quote_card`,
   `comparison`, standalone `frame`. If our IR/templates need these, either
   compose them from existing nodes (grids of cards, etc.) or land them upstream.
8. **Shape geometry set is small.** Only ~10 `ShapeGeometry` presets are exposed
   (rect/roundRect/ellipse/triangle/diamond/parallelogram/hexagon/chevron/
   rightArrow/line); the RFC's broader OOXML preset list and `CustomGeom` are not
   shipped. Custom diagram work must stay within these or be composed.
9. **Deck sections have no scene-IR field.** Sorter sections exist only at Layer 1
   (`pres.AddSection`/`Section.Include`). Our deck-overview "structure/reorder"
   surface must call Layer 1 directly (or we add a scene-side wrapper).
10. **Spacing scale is 6 steps.** A richer soul spacing scale (8–12 steps) must
    collapse into XS..2XL (documented `xxs..xxxl → xs..2xl` mapping). Lossy but
    fine for slides.
11. **No JSON/YAML theme load.** `LoadThemeFile` is V1.1+. Souls must be
    constructed in Go from our own persisted representation.
12. **Authoring-only soul layers dropped by design.** Motion/tone/voice/do-don't
    have no PPTX semantics and are not in `Theme` (expected) — they live entirely
    in Deckard's authoring/agent layer.

---

## 6. Net assessment for the rewrite

pptx-go gives Deckard a complete, deterministic, pure-Go, zero-dep PPTX engine
with the exact two-property foundation we need: **tokens (souls) and parallel
deterministic render**. The full scene node catalog is implemented today, so the
output ceiling is high out of the box. The engine deliberately owns *only*
rendering; everything agent-facing (IR construction, rasterization of
charts/code, validation, soul persistence, export delivery, the three ui://
surfaces) is ours to build on top — which is exactly the agent-first, no-Chromium
product we are specifying. The principal things we must supply around it:
pure-Go chart + code rasterizers, soul persistence + re-apply (don't trust saved
theme1.xml round-trip for code-authored themes), logo injection, and (optionally)
the §11.3 richer nodes if our templates demand them.
