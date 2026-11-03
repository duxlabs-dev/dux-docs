package previews

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"sync"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

const (
	cardWidth        = 960
	cardHeight       = 540
	thumbWidth       = 960
	thumbHeight      = 600
	regionInset      = 16.0
	regionWidth      = 928.0
	baseRegionHeight = 392.0
	descLineHeight   = 30.0
	maxDescLines     = 4
	regionRadius     = 16.0
	accentHeight     = 4.0
	coverTolerance   = 0.04
	maxUpscale       = 2.0
	blurSampleWidth  = 58
	blurSampleHeight = 26
	blurOverlayAlpha = 0.55
	chipHeight       = 28.0
	chipPadX         = 12.0
	chipGap          = 8.0
)

var (
	colPrimary          = color.NRGBA{R: 0xD0, G: 0xBC, B: 0xFF, A: 0xFF}
	colSurface          = color.NRGBA{R: 0x14, G: 0x12, B: 0x18, A: 0xFF}
	colSurfaceText      = color.NRGBA{R: 0xE6, G: 0xE0, B: 0xE9, A: 0xFF}
	colSurfaceContainer = color.NRGBA{R: 0x21, G: 0x1F, B: 0x24, A: 0xFF}
	colOutline          = color.NRGBA{R: 0x94, G: 0x8F, B: 0x99, A: 0xFF}
	colDescription      = color.NRGBA{R: 0xC4, G: 0xC7, B: 0xC5, A: 0xFF}
)

var statusChipColors = map[string]color.NRGBA{
	"broken":       {R: 0xFF, G: 0xB4, B: 0xAB, A: 0xFF},
	"unmaintained": {R: 0xF9, G: 0xD6, B: 0x8A, A: 0xFF},
	"deprecated":   {R: 0xCF, G: 0xCF, B: 0xCF, A: 0xFF},
	"reviewed":     {R: 0xA9, G: 0xD8, B: 0x9B, A: 0xFF},
}

var (
	fontOnce    sync.Once
	fontErr     error
	regularFont *opentype.Font
	boldFont    *opentype.Font
)

func loadFonts() error {
	fontOnce.Do(func() {
		regularFont, fontErr = opentype.Parse(goregular.TTF)
		if fontErr != nil {
			return
		}
		boldFont, fontErr = opentype.Parse(gobold.TTF)
	})
	if fontErr != nil {
		return fmt.Errorf("failed to parse embedded fonts: %w", fontErr)
	}
	return nil
}

func newFace(f *opentype.Font, size float64) (font.Face, error) {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}
	return face, nil
}

func withAlpha(c color.NRGBA, alpha float64) color.NRGBA {
	c.A = uint8(math.Round(alpha * 255))
	return c
}

func lighten(c color.NRGBA, frac float64) color.NRGBA {
	return color.NRGBA{
		R: lightenChannel(c.R, frac),
		G: lightenChannel(c.G, frac),
		B: lightenChannel(c.B, frac),
		A: c.A,
	}
}

func lightenChannel(v uint8, frac float64) uint8 {
	return uint8(math.Round(float64(v) + (255-float64(v))*frac))
}

func newCanvas() *gg.Context {
	return newCanvasSized(cardWidth, cardHeight)
}

func newCanvasSized(w, h int) *gg.Context {
	dc := gg.NewContext(w, h)
	grad := gg.NewLinearGradient(0, 0, 0, float64(h))
	grad.AddColorStop(0, colSurface)
	grad.AddColorStop(1, lighten(colSurface, 0.06))
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()
	return dc
}

func ellipsize(dc *gg.Context, s string, maxW float64) string {
	if w, _ := dc.MeasureString(s); w <= maxW {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := strings.TrimRight(string(runes), " …") + "…"
		if w, _ := dc.MeasureString(candidate); w <= maxW {
			return candidate
		}
	}
	return "…"
}

type chipSpec struct {
	label string
	text  color.NRGBA
	fill  color.NRGBA
}

func drawChipRow(dc *gg.Context, chips []chipSpec, right, y float64) (float64, error) {
	face, err := newFace(boldFont, 14)
	if err != nil {
		return 0, err
	}
	dc.SetFontFace(face)

	x := right
	for i := len(chips) - 1; i >= 0; i-- {
		chip := chips[i]
		textW, _ := dc.MeasureString(chip.label)
		chipW := textW + 2*chipPadX
		x -= chipW

		dc.SetColor(chip.fill)
		dc.DrawRoundedRectangle(x, y, chipW, chipHeight, chipHeight/2)
		dc.Fill()

		dc.SetColor(chip.text)
		dc.DrawStringAnchored(chip.label, x+chipW/2, y+chipHeight/2, 0.5, 0.35)
		x -= chipGap
	}
	return x + chipGap, nil
}

func footerChips(p models.Plugin) []chipSpec {
	var chips []chipSpec
	if p.Version != "" {
		label := p.Version
		if !strings.HasPrefix(label, "v") {
			label = "v" + label
		}
		chips = append(chips, chipSpec{label: label, text: colSurfaceText, fill: withAlpha(colSurfaceText, 0.10)})
	}
	if p.Category != "" {
		chips = append(chips, chipSpec{label: strings.ToUpper(p.Category), text: colPrimary, fill: withAlpha(colPrimary, 0.15)})
	}
	return chips
}

func wrapLines(dc *gg.Context, s string, maxW float64, maxLines int) []string {
	if s == "" {
		return nil
	}
	lines := dc.WordWrap(s, maxW)
	if len(lines) <= maxLines {
		return lines
	}
	lines = lines[:maxLines]
	lines[maxLines-1] = ellipsize(dc, lines[maxLines-1]+"…", maxW)
	return lines
}

func footerLayout(dc *gg.Context, p models.Plugin) ([]string, float64, error) {
	descFace, err := newFace(regularFont, 22)
	if err != nil {
		return nil, 0, err
	}
	dc.SetFontFace(descFace)

	lines := wrapLines(dc, p.Description, cardWidth-2*regionInset, maxDescLines)
	extra := max(len(lines), 1) - 1
	return lines, baseRegionHeight - float64(extra)*descLineHeight, nil
}

func drawStatusChips(dc *gg.Context, statuses []string) error {
	var chips []chipSpec
	for _, status := range statuses {
		tint, ok := statusChipColors[status]
		if !ok {
			continue
		}
		chips = append(chips, chipSpec{label: strings.ToUpper(status), text: tint, fill: withAlpha(colSurface, 0.78)})
	}
	if len(chips) == 0 {
		return nil
	}
	_, err := drawChipRow(dc, chips, regionInset+regionWidth-12, regionInset+12)
	return err
}

func drawFooter(dc *gg.Context, p models.Plugin, descLines []string, regionHeight float64) error {
	dc.SetColor(colPrimary)
	dc.DrawRectangle(0, cardHeight-accentHeight, cardWidth, accentHeight)
	dc.Fill()

	regionBottom := regionInset + regionHeight
	chipLeft := float64(cardWidth - regionInset)
	chips := footerChips(p)
	if len(chips) > 0 {
		left, err := drawChipRow(dc, chips, cardWidth-regionInset, regionBottom+19)
		if err != nil {
			return err
		}
		chipLeft = left
	}

	nameFace, err := newFace(boldFont, 40)
	if err != nil {
		return err
	}
	dc.SetFontFace(nameFace)
	dc.SetColor(colSurfaceText)
	nameMax := chipLeft - 16 - regionInset
	dc.DrawString(ellipsize(dc, p.Name, nameMax), regionInset, regionBottom+48)

	if len(descLines) == 0 {
		return nil
	}

	descFace, err := newFace(regularFont, 22)
	if err != nil {
		return err
	}
	dc.SetFontFace(descFace)
	dc.SetColor(colDescription)
	for i, line := range descLines {
		dc.DrawString(line, regionInset, regionBottom+88+float64(i)*descLineHeight)
	}
	return nil
}

func ComposeScreenshot(src image.Image, p models.Plugin) (image.Image, error) {
	if err := loadFonts(); err != nil {
		return nil, err
	}

	dc := newCanvas()
	descLines, regionHeight, err := footerLayout(dc, p)
	if err != nil {
		return nil, err
	}

	dc.SetColor(colSurfaceContainer)
	dc.DrawRoundedRectangle(regionInset, regionInset, regionWidth, regionHeight, regionRadius)
	dc.Fill()

	dc.DrawRoundedRectangle(regionInset, regionInset, regionWidth, regionHeight, regionRadius)
	dc.Clip()
	drawFitted(dc, src, regionInset, regionInset, regionWidth, regionHeight)
	dc.ResetClip()

	if err := drawStatusChips(dc, p.Status); err != nil {
		return nil, err
	}
	if err := drawFooter(dc, p, descLines, regionHeight); err != nil {
		return nil, err
	}
	return dc.Image(), nil
}

// drawFitted letterboxes src into the box at (x0,y0), covering only when the ratios nearly match.
// Letterbox bands are a dimmed blur of src so portrait screenshots don't sit on flat color.
func drawFitted(dc *gg.Context, src image.Image, x0, y0, w, h float64) {
	scaled, x, y := fitImage(src, w, h)
	if scaled.Bounds().Dx() < int(w) || scaled.Bounds().Dy() < int(h) {
		dc.DrawImage(blurFill(src, w, h), int(x0), int(y0))
		dc.SetColor(withAlpha(colSurface, blurOverlayAlpha))
		dc.DrawRectangle(x0, y0, w, h)
		dc.Fill()
	}
	dc.DrawImage(scaled, int(x0)+x, int(y0)+y)
}

func blurFill(src image.Image, w, h float64) image.Image {
	small := image.NewRGBA(image.Rect(0, 0, blurSampleWidth, blurSampleHeight))
	draw.ApproxBiLinear.Scale(small, small.Bounds(), src, src.Bounds(), draw.Src, nil)
	full := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	draw.CatmullRom.Scale(full, full.Bounds(), small, small.Bounds(), draw.Src, nil)
	return full
}

func fitImage(src image.Image, w, h float64) (image.Image, int, int) {
	b := src.Bounds()
	sw, sh := float64(b.Dx()), float64(b.Dy())
	boxAR := w / h
	srcAR := sw / sh

	scale := math.Min(w/sw, h/sh)
	if math.Abs(srcAR-boxAR)/boxAR <= coverTolerance {
		scale = math.Max(w/sw, h/sh)
	}
	scale = math.Min(scale, maxUpscale)

	dw := max(int(math.Round(sw*scale)), 1)
	dh := max(int(math.Round(sh*scale)), 1)

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	return dst, (int(w) - dw) / 2, (int(h) - dh) / 2
}
