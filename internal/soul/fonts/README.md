# Bundled fonts (R9.1 font-embedding-pipeline)

These faces back the built-in **Deckard White** soul's type roles so exported
decks embed their own brand type and render the editorial serif on any machine
(no host install). They are exposed to the render/export path via
[`fonts.Provider()`](fonts.go), which implements `pptx.FontSource`.

All faces are licensed under the **SIL Open Font License 1.1** (see the
`OFL-*.txt` files, distributed with this source per the license).

| File(s) | Family | Form | Weight(s) | Upstream |
| --- | --- | --- | --- | --- |
| `PlayfairDisplay.ttf`, `PlayfairDisplay-Italic.ttf` | Playfair Display | **unmodified** variable font | 400 default (400–900 axis) | [google/fonts ofl/playfairdisplay](https://github.com/google/fonts/tree/main/ofl/playfairdisplay) |
| `Lora.ttf`, `Lora-Italic.ttf` | Lora | **unmodified** variable font | 400 default (400–700 axis) | [google/fonts ofl/lora](https://github.com/google/fonts/tree/main/ofl/lora) |
| `Inter-Regular/Medium/Bold/Italic.ttf` | Inter | static instances | 400 / 500 / 700 / 400-italic | [google/fonts ofl/inter](https://github.com/google/fonts/tree/main/ofl/inter) |

## Reserved Font Name compliance

`Lora` and `Playfair Display` each declare a **Reserved Font Name** in their OFL
(`with Reserved Font Name "Lora"` / `"Playfair Display"`). Under the OFL a
*modified* copy may not keep the reserved name — so the bundled serif faces are
the **unmodified upstream variable fonts**; only the file was renamed (the OFL
permits renaming the file; the font software and its `name` table are untouched).
Their default variable instance is weight 400, which is the weight the default
soul uses for display and H1–H3.

`Inter` declares **no** Reserved Font Name, so its faces are static weight
instances (400/500/700 + italic) produced from the upstream variable font with
`fontTools.varLib.instancer` — permitted, and smaller than shipping the ~880 KB
variable file. Each instance keeps the family name `Inter`.

## Mono is intentionally not bundled

The default soul names `Consolas` (a system font, not OFL-redistributable) for
its mono roles, and code blocks render as pure-Go rasters (P4), so the OOXML mono
face only affects incidental inline mono. It is left as a host-monospace fallback
rather than bundling a redistributable mono the default soul does not name.

## Adding or changing a face

The face manifest in [`fonts.go`](fonts.go) (`bundled`) is hand-maintained to
match the files here — there is no runtime font parsing. When you add a `.ttf`,
add its `//go:embed`-covered file **and** a `bundled` entry (family/italic/weight
must match the file's `name` table and OS/2 `usWeightClass`), and drop the OFL
license text beside it.
