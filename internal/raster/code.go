package raster

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

// code raster layout constants.
const (
	// codeSlotAspect matches pptx-go's top-level code_block slot: the full body
	// content width (slide 13.333" - 2*0.5" margin = 12.333") over the fixed
	// preferredHeight of 2.6". The slot is filled with FitFill (a stretch), so a
	// raster at this exact aspect scales uniformly — no horizontal distortion.
	codeSlotAspect = 12.333 / 2.6

	codeCanvasH    = 480 // raster height in px (width derives from the aspect)
	codePad        = 56  // inset from the canvas edge to the text block
	codeLineFactor = 1.5 // line height as a multiple of the font size
	codeMinFont    = 12.0
	codeMaxFont    = 44.0
	codeAscentFrac = 0.78 // baseline offset within a line, as a fraction of size
)

// code surface colors: a dark warm Deckard surface with off-white text.
var (
	codeBG   = color.RGBA{R: 0x2B, G: 0x27, B: 0x23, A: 0xFF} // #2B2723
	codeText = color.RGBA{R: 0xFA, G: 0xF7, B: 0xF2, A: 0xFF} // #FAF7F2
)

// RasterizeCode renders source code to a deterministic PNG using the Go Mono
// font (pure Go, CGo-free). The canvas matches the engine's code slot aspect so
// it is not distorted when placed, and the font size auto-fits so the code
// fills the fixed-height slot legibly at any line count. The language badge is
// NOT drawn here — the scene renderer overlays a native badge over the picture.
func RasterizeCode(code, language string) ([]byte, error) {
	_ = language // the engine draws the language badge natively over the image.
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("raster: code is empty")
	}
	tt, err := truetype.Parse(gomono.TTF)
	if err != nil {
		return nil, fmt.Errorf("raster: parse mono font: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(code, "\t", "    "), "\n")
	canvasW := int(math.Round(codeCanvasH * codeSlotAspect))
	textW := canvasW - 2*codePad
	textH := codeCanvasH - 2*codePad

	size := fitFontSize(tt, lines, textW, textH)
	lineH := int(size * codeLineFactor)

	img := image.NewRGBA(image.Rect(0, 0, canvasW, codeCanvasH))
	draw.Draw(img, img.Bounds(), image.NewUniform(codeBG), image.Point{}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDPI(72) // size in points == pixels
	ctx.SetFont(tt)
	ctx.SetFontSize(size)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(image.NewUniform(codeText))

	totalH := len(lines) * lineH
	startY := codePad + (textH-totalH)/2
	if startY < codePad {
		startY = codePad
	}
	ascent := int(size * codeAscentFrac)
	for i, line := range lines {
		baseline := startY + ascent + i*lineH
		if _, err := ctx.DrawString(line, freetype.Pt(codePad, baseline)); err != nil {
			return nil, fmt.Errorf("raster: draw line: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("raster: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// fitFontSize returns the largest font size (in px) at which every line fits the
// text width and all lines fit the text height. Measured against real Go Mono
// glyph advances so long lines never clip.
func fitFontSize(tt *truetype.Font, lines []string, textW, textH int) float64 {
	for size := codeMaxFont; size > codeMinFont; size -= 1.0 {
		lineH := int(size * codeLineFactor)
		if len(lines)*lineH > textH {
			continue
		}
		face := truetype.NewFace(tt, &truetype.Options{Size: size, DPI: 72, Hinting: font.HintingFull})
		d := &font.Drawer{Face: face}
		maxW := 0
		for _, line := range lines {
			if w := d.MeasureString(line).Round(); w > maxW {
				maxW = w
			}
		}
		_ = face.Close()
		if maxW <= textW {
			return size
		}
	}
	return codeMinFont
}
