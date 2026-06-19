package raster

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gomono"
)

// code raster layout constants.
const (
	codeFontSize = 28.0
	codeDPI      = 96.0
	codePadding  = 48
	codeLineH    = 40
	codeCharW    = 17 // approx advance width per Go Mono glyph at the size/DPI above
	codeMinW     = 640
	codeMaxW     = 2400
)

// code surface colors: a dark warm Deckard surface with off-white text (classic,
// legible code-block look; matches the Deckard palette).
var (
	codeBG   = color.RGBA{R: 0x2B, G: 0x27, B: 0x23, A: 0xFF} // #2B2723
	codeText = color.RGBA{R: 0xFA, G: 0xF7, B: 0xF2, A: 0xFF} // #FAF7F2
)

// RasterizeCode renders source code to a deterministic PNG using the Go Mono
// font (pure Go, CGo-free). language, if set, is drawn as a small header badge.
// No syntax highlighting in V1 — monospace, legible, on the Deckard dark surface.
func RasterizeCode(code, language string) ([]byte, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("raster: code is empty")
	}
	font, err := truetype.Parse(gomono.TTF)
	if err != nil {
		return nil, fmt.Errorf("raster: parse mono font: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(code, "\t", "    "), "\n")
	maxLen := len(language) + 2
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	width := codePadding*2 + maxLen*codeCharW
	if width < codeMinW {
		width = codeMinW
	}
	if width > codeMaxW {
		width = codeMaxW
	}
	header := 0
	if language != "" {
		header = codeLineH
	}
	height := codePadding*2 + header + len(lines)*codeLineH

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(codeBG), image.Point{}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDPI(codeDPI)
	ctx.SetFont(font)
	ctx.SetFontSize(codeFontSize)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)

	y := codePadding
	if language != "" {
		ctx.SetSrc(image.NewUniform(deckardTealUniform()))
		if _, err := ctx.DrawString(strings.ToUpper(language), freetype.Pt(codePadding, y+codeLineH-12)); err != nil {
			return nil, fmt.Errorf("raster: draw badge: %w", err)
		}
		y += codeLineH
	}

	ctx.SetSrc(image.NewUniform(codeText))
	for _, line := range lines {
		y += codeLineH
		if _, err := ctx.DrawString(line, freetype.Pt(codePadding, y-12)); err != nil {
			return nil, fmt.Errorf("raster: draw line: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("raster: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// deckardTealUniform is the Deckard accent teal for the language badge.
func deckardTealUniform() color.Color {
	return color.RGBA{R: 0x3B, G: 0x9C, B: 0x94, A: 0xFF} // #3B9C94
}
