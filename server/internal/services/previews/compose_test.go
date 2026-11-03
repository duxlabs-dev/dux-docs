package previews

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

var testPlugin = models.Plugin{
	ID:          "test-plugin",
	Name:        "Test Plugin",
	Category:    "widgets",
	Description: "A plugin used for testing preview composition",
	Author:      "tester",
}

var sourceRed = color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}

func solidImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func assertPixel(t *testing.T, img image.Image, x, y int, want color.NRGBA) {
	t.Helper()
	r, g, b, _ := img.At(x, y).RGBA()
	got := color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xFF}
	if got.R != want.R || got.G != want.G || got.B != want.B {
		t.Fatalf("pixel (%d,%d) = #%02x%02x%02x, want #%02x%02x%02x", x, y, got.R, got.G, got.B, want.R, want.G, want.B)
	}
}

func assertPixelNear(t *testing.T, img image.Image, x, y int, want color.NRGBA, tol int) {
	t.Helper()
	r, g, b, _ := img.At(x, y).RGBA()
	got := color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xFF}
	near := func(a, b uint8) bool {
		d := int(a) - int(b)
		return d <= tol && d >= -tol
	}
	if !near(got.R, want.R) || !near(got.G, want.G) || !near(got.B, want.B) {
		t.Fatalf("pixel (%d,%d) = #%02x%02x%02x, want ~#%02x%02x%02x", x, y, got.R, got.G, got.B, want.R, want.G, want.B)
	}
}

var letterboxRed = color.NRGBA{R: 126, G: 10, B: 13, A: 0xFF}

func TestComposeScreenshotDimensions(t *testing.T) {
	cases := []struct {
		name string
		w, h int
	}{
		{"16x9", 1600, 900},
		{"4x3", 800, 600},
		{"21x9", 2100, 900},
		{"9x16", 540, 960},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := ComposeScreenshot(solidImage(tc.w, tc.h, sourceRed), testPlugin)
			if err != nil {
				t.Fatalf("ComposeScreenshot: %v", err)
			}
			b := img.Bounds()
			if b.Dx() != 960 || b.Dy() != 540 {
				t.Fatalf("output size %dx%d, want 960x540", b.Dx(), b.Dy())
			}
		})
	}
}

func TestComposeScreenshotContainBlurredLetterbox(t *testing.T) {
	img, err := ComposeScreenshot(solidImage(800, 600, sourceRed), testPlugin)
	if err != nil {
		t.Fatalf("ComposeScreenshot: %v", err)
	}
	assertPixelNear(t, img, 30, 220, letterboxRed, 8)
	assertPixelNear(t, img, 930, 220, letterboxRed, 8)
	assertPixel(t, img, 480, 220, sourceRed)
}

func TestComposeScreenshotPortraitBlurredLetterbox(t *testing.T) {
	img, err := ComposeScreenshot(solidImage(540, 960, sourceRed), testPlugin)
	if err != nil {
		t.Fatalf("ComposeScreenshot: %v", err)
	}
	assertPixelNear(t, img, 30, 220, letterboxRed, 8)
	assertPixel(t, img, 480, 220, sourceRed)
}

func TestComposeScreenshotCoverNearRegionRatio(t *testing.T) {
	img, err := ComposeScreenshot(solidImage(2100, 900, sourceRed), testPlugin)
	if err != nil {
		t.Fatalf("ComposeScreenshot: %v", err)
	}
	assertPixel(t, img, 100, 220, sourceRed)
	assertPixel(t, img, 860, 30, sourceRed)
}

func TestComposeScreenshotAccentBar(t *testing.T) {
	img, err := ComposeScreenshot(solidImage(1600, 900, sourceRed), testPlugin)
	if err != nil {
		t.Fatalf("ComposeScreenshot: %v", err)
	}
	assertPixel(t, img, 480, 538, colPrimary)
}

func TestComposeScreenshotUpscaleCapped(t *testing.T) {
	img, err := ComposeScreenshot(solidImage(120, 50, sourceRed), testPlugin)
	if err != nil {
		t.Fatalf("ComposeScreenshot: %v", err)
	}
	assertPixel(t, img, 480, 220, sourceRed)
	assertPixelNear(t, img, 480-160, 220, letterboxRed, 8)
	assertPixelNear(t, img, 480+160, 220, letterboxRed, 8)
}

func TestFooterLayoutShowsFullDescription(t *testing.T) {
	if err := loadFonts(); err != nil {
		t.Fatal(err)
	}
	dc := newCanvas()
	p := testPlugin
	p.Description = strings.TrimSpace(strings.Repeat("wide words flow across the card footer band ", 4))

	lines, regionH, err := footerLayout(dc, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 2 {
		t.Fatalf("expected description to wrap, got %d line(s)", len(lines))
	}
	if got := strings.Join(lines, " "); got != p.Description {
		t.Fatalf("description altered by wrapping:\n%q\nwant\n%q", got, p.Description)
	}
	want := baseRegionHeight - float64(len(lines)-1)*descLineHeight
	if regionH != want {
		t.Fatalf("region height %v, want %v", regionH, want)
	}
}

func TestComposeCardDimensions(t *testing.T) {
	img, err := ComposeCard(testPlugin)
	if err != nil {
		t.Fatalf("ComposeCard: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 960 || b.Dy() != 540 {
		t.Fatalf("output size %dx%d, want 960x540", b.Dx(), b.Dy())
	}
	assertPixel(t, img, 480, 538, colPrimary)
}

func TestComposeThumbFillsWithoutFooter(t *testing.T) {
	img := ComposeThumb(solidImage(1600, 1000, sourceRed))
	b := img.Bounds()
	if b.Dx() != 960 || b.Dy() != 600 {
		t.Fatalf("output size %dx%d, want 960x600", b.Dx(), b.Dy())
	}
	assertPixel(t, img, 2, 2, sourceRed)
	assertPixel(t, img, 957, 597, sourceRed)
}

func TestComposeThumbPortraitBlurredLetterbox(t *testing.T) {
	img := ComposeThumb(solidImage(540, 960, sourceRed))
	assertPixel(t, img, 480, 300, sourceRed)
	assertPixelNear(t, img, 40, 300, letterboxRed, 12)
}
