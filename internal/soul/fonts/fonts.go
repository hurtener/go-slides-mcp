// Package fonts bundles the redistributable OFL faces the built-in Deckard
// White soul names (Playfair Display, Lora, Inter) and exposes them as a
// pptx.FontSource so the render/export path can physically embed them (R9.1,
// font-embedding-pipeline). Without embedding, a machine that lacks the brand
// serif substitutes a host sans and the deck loses its editorial identity; the
// bundled faces make the deck render with its own type on any machine, with no
// host install.
//
// # Licensing (SIL Open Font License 1.1)
//
// Every bundled .ttf ships alongside its OFL license text (OFL-*.txt). Lora and
// Playfair Display carry a Reserved Font Name, so a *modified* copy may not keep
// that family name — the bundled serif faces are therefore the UNMODIFIED
// upstream variable fonts (only the file was renamed, which the OFL permits;
// the font software and its name table are untouched). Inter has no Reserved
// Font Name, so its faces are static instances (weights 400/500/700 + italic)
// derived from the upstream variable font, which is permitted.
//
// Mono is deliberately NOT bundled: the default soul names "Consolas" (a system
// font, not OFL-redistributable) for its mono roles, and code blocks render as
// pure-Go rasters (P4), so the OOXML mono face only affects incidental inline
// mono and safely falls back to a host monospace.
package fonts

import (
	"embed"
	"sort"
	"strings"
	"sync"

	"github.com/hurtener/pptx-go/pptx"
)

// ttf embeds the bundled OFL faces. The OFL-*.txt license files sit beside them
// in the package directory (distributed with the source, per the OFL) but are
// not embedded — only the font bytes are needed at runtime.
//
//go:embed *.ttf
var ttf embed.FS

// face is one bundled font file: its family (as named in the font's name table,
// matching what a soul's type roles reference), whether it is the italic cut,
// its numeric weight, and the embed path to its bytes.
type face struct {
	family string // canonical family name, e.g. "Playfair Display"
	italic bool
	weight int
	path   string // path within the embedded FS
}

// bundled is the static manifest of shipped faces. It is hand-maintained to
// match the files in this directory (no runtime font parsing), keeping
// resolution deterministic and dependency-free. Weights reflect each file's
// OS/2 usWeightClass; the serif faces are variable fonts whose default instance
// is weight 400 (the weight the default soul uses for display/headings).
var bundled = []face{
	{family: "Playfair Display", italic: false, weight: 400, path: "PlayfairDisplay.ttf"},
	{family: "Playfair Display", italic: true, weight: 400, path: "PlayfairDisplay-Italic.ttf"},
	{family: "Lora", italic: false, weight: 400, path: "Lora.ttf"},
	{family: "Lora", italic: true, weight: 400, path: "Lora-Italic.ttf"},
	{family: "Inter", italic: false, weight: 400, path: "Inter-Regular.ttf"},
	{family: "Inter", italic: false, weight: 500, path: "Inter-Medium.ttf"},
	{family: "Inter", italic: false, weight: 700, path: "Inter-Bold.ttf"},
	{family: "Inter", italic: true, weight: 400, path: "Inter-Italic.ttf"},
}

// provider resolves (family, style, weight) to bundled font bytes. It is
// immutable after construction and therefore safe for concurrent use by the
// engine's save-time embedding pass across presentations saved in parallel.
type provider struct {
	// byKey groups faces by (lowercased family, italic) so Resolve can pick the
	// nearest weight within a cut. Each slice is sorted ascending by weight.
	byKey map[fontKey][]resolved
}

type fontKey struct {
	family string // lowercased
	italic bool
}

type resolved struct {
	weight int
	data   []byte
}

var (
	singleton *provider
	once      sync.Once
)

// Provider returns the shared FontSource backed by the bundled OFL faces. The
// built-in Deckard White soul registers it so its serif display/heading faces
// embed on export; any soul that names one of the bundled families (Playfair
// Display, Lora, Inter) resolves through it. A family the provider does not
// bundle yields ErrFontNotFound, which the engine's embedding pass treats as
// warn-don't-fail (the face simply is not embedded).
func Provider() pptx.FontSource {
	once.Do(func() {
		p := &provider{byKey: make(map[fontKey][]resolved)}
		for _, f := range bundled {
			data, err := ttf.ReadFile(f.path)
			if err != nil {
				// A missing embedded file is a build-time programming error (the
				// manifest and the //go:embed set have drifted); fail loudly rather
				// than silently shipping a soul that cannot embed its own faces.
				panic("soul/fonts: bundled face not embedded: " + f.path + ": " + err.Error())
			}
			k := fontKey{family: strings.ToLower(f.family), italic: f.italic}
			p.byKey[k] = append(p.byKey[k], resolved{weight: f.weight, data: data})
		}
		for k := range p.byKey {
			slice := p.byKey[k]
			sort.Slice(slice, func(i, j int) bool { return slice[i].weight < slice[j].weight })
		}
		singleton = p
	})
	return singleton
}

// Resolve implements pptx.FontSource. It matches name against a bundled family
// (case-insensitively), selects the italic or upright cut from style, and
// returns the bytes of the nearest available weight (ties resolve to the lower
// weight, deterministically). A family the provider does not bundle returns
// (nil, pptx.ErrFontNotFound) so the engine warns and skips the face rather than
// failing the export.
func (p *provider) Resolve(name, style string, weight int) ([]byte, error) {
	italic := strings.EqualFold(style, "italic") || strings.EqualFold(style, "oblique")
	k := fontKey{family: strings.ToLower(strings.TrimSpace(name)), italic: italic}
	candidates, ok := p.byKey[k]
	if !ok || len(candidates) == 0 {
		// No cut in the requested slant. If an italic run has no italic cut, fall
		// back to the upright cut of the same family (better a real serif upright
		// than a host sans); an upright request never borrows the italic cut.
		if italic {
			if up, ok2 := p.byKey[fontKey{family: k.family, italic: false}]; ok2 && len(up) > 0 {
				candidates = up
			}
		}
		if len(candidates) == 0 {
			return nil, pptx.ErrFontNotFound
		}
	}
	return nearest(candidates, weight).data, nil
}

// nearest returns the candidate whose weight is closest to want; on a tie it
// prefers the lower weight so selection is deterministic regardless of slice
// order (the slices are weight-sorted, so this scans ascending).
func nearest(candidates []resolved, want int) resolved {
	best := candidates[0]
	bestDelta := absInt(best.weight - want)
	for _, c := range candidates[1:] {
		d := absInt(c.weight - want)
		if d < bestDelta {
			best, bestDelta = c, d
		}
	}
	return best
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
