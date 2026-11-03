package registry

import (
	"math"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

func TestContrastRatioBlackWhite(t *testing.T) {
	black, _ := parseHexColor("#000000")
	white, _ := parseHexColor("#FFFFFF")

	ratio := contrastRatio(black, white)
	if math.Abs(ratio-21.0) > 0.01 {
		t.Fatalf("expected 21.0, got %f", ratio)
	}
}

func TestContrastRatioAABoundaryGray(t *testing.T) {
	// #767676 on white is the canonical 4.54:1 AA boundary gray
	gray, _ := parseHexColor("#767676")
	white, _ := parseHexColor("#FFFFFF")

	ratio := contrastRatio(gray, white)
	if math.Abs(ratio-4.54) > 0.01 {
		t.Fatalf("expected 4.54, got %f", ratio)
	}
}

func TestParseHexColorRejectsInvalid(t *testing.T) {
	invalid := []interface{}{nil, 42, "", "#fff", "#GGGGGG", "#+12345", "123456#"}
	for _, value := range invalid {
		if _, ok := parseHexColor(value); ok {
			t.Fatalf("expected %v to be rejected", value)
		}
	}
}

func TestWCAGLevelThresholds(t *testing.T) {
	cases := map[float64]string{21.0: "AAA", 7.0: "AAA", 6.99: "AA", 4.5: "AA", 4.49: "fail"}
	for ratio, expected := range cases {
		if got := wcagLevel(ratio); got != expected {
			t.Fatalf("ratio %f: expected %s, got %s", ratio, expected, got)
		}
	}
}

func TestComputeThemeWCAGSimple(t *testing.T) {
	theme := &models.Theme{
		Dark: map[string]interface{}{
			"surfaceText": "#FFFFFF",
			"surface":     "#000000",
		},
		Light: map[string]interface{}{
			"surfaceText": "#767676",
			"surface":     "#FFFFFF",
		},
	}

	wcag := computeThemeWCAG(theme)
	if wcag == nil {
		t.Fatal("expected report, got nil")
	}
	if wcag.Level != "AA" {
		t.Fatalf("expected overall AA, got %s", wcag.Level)
	}
	if wcag.Dark.Level != "AAA" {
		t.Fatalf("expected dark AAA, got %s", wcag.Dark.Level)
	}
	if wcag.Light.Level != "AA" {
		t.Fatalf("expected light AA, got %s", wcag.Light.Level)
	}
	if wcag.Dark.Variants != nil {
		t.Fatal("expected no variants map for variant-less theme")
	}
}

func TestComputeThemeWCAGVariantsUseDefault(t *testing.T) {
	theme := &models.Theme{
		Dark: map[string]interface{}{
			"surfaceText": "#FFFFFF",
			"surface":     "#000000",
		},
		Light: map[string]interface{}{
			"surfaceText": "#FFFFFF",
			"surface":     "#000000",
		},
		Variants: &models.ThemeVariants{
			Default: "good",
			Options: []models.ThemeVariantOption{
				{ID: "good"},
				{
					ID:    "bad",
					Dark:  map[string]interface{}{"surfaceText": "#777777", "surface": "#888888"},
					Light: map[string]interface{}{"surfaceText": "#777777", "surface": "#888888"},
				},
			},
		},
	}

	wcag := computeThemeWCAG(theme)
	if wcag == nil {
		t.Fatal("expected report, got nil")
	}
	if wcag.Level != "AAA" {
		t.Fatalf("expected default variant to drive overall level AAA, got %s", wcag.Level)
	}
	if wcag.Dark.Variants["bad"] != "fail" {
		t.Fatalf("expected bad variant to fail, got %s", wcag.Dark.Variants["bad"])
	}
	if wcag.Dark.Variants["good"] != "AAA" {
		t.Fatalf("expected good variant AAA, got %s", wcag.Dark.Variants["good"])
	}
}

func TestComputeThemeWCAGMultiVariants(t *testing.T) {
	theme := &models.Theme{
		Dark: map[string]interface{}{
			"surfaceText": "#FFFFFF",
			"surface":     "#000000",
		},
		Light: map[string]interface{}{
			"surfaceText": "#000000",
			"surface":     "#FFFFFF",
		},
		Variants: &models.ThemeVariants{
			Type: "multi",
			Defaults: map[string]*models.ThemeModeDefaults{
				"dark":  {Flavor: "mocha", Accent: "blue"},
				"light": {Flavor: "latte", Accent: "blue"},
			},
			Flavors: []models.ThemeFlavor{
				{ID: "mocha", Dark: map[string]interface{}{"primary": "#89B4FA"}},
				{ID: "latte", Light: map[string]interface{}{"primary": "#1E66F5"}},
			},
			Accents: []map[string]interface{}{
				{
					"id":    "blue",
					"mocha": map[string]interface{}{"primaryText": "#11111B"},
					"latte": map[string]interface{}{"primaryText": "#EFF1F5"},
				},
			},
		},
	}

	wcag := computeThemeWCAG(theme)
	if wcag == nil {
		t.Fatal("expected report, got nil")
	}
	if wcag.Dark == nil || wcag.Light == nil {
		t.Fatal("expected reports for both modes")
	}
	if wcag.Dark.Level != "AAA" {
		t.Fatalf("expected dark AAA, got %s", wcag.Dark.Level)
	}
}

func TestComputeThemeWCAGNonTextGate(t *testing.T) {
	scheme := map[string]interface{}{
		"surfaceText":      "#FFFFFF",
		"surface":          "#000000",
		"surfaceContainer": "#000000",
		"error":            "#1a1a1a",
	}

	report := schemeWCAG(scheme)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Level != "fail" {
		t.Fatalf("expected non-text failure to fail the scheme, got %s", report.Level)
	}
	if report.NonText == nil || report.NonText.WorstPair[0] != "error" {
		t.Fatalf("expected error to be the worst non-text pair, got %+v", report.NonText)
	}
	if report.MinRatio != 21 {
		t.Fatalf("expected text ratio to stay 21, got %f", report.MinRatio)
	}
}

func TestSchemeWCAGSplitsBodyFromAccent(t *testing.T) {
	// Readable body text with an accent too light to read on the bar.
	scheme := map[string]interface{}{
		"surfaceText":      "#FFFFFF",
		"surface":          "#000000",
		"surfaceContainer": "#000000",
		"primary":          "#3A3A3A",
		"primaryText":      "#FFFFFF",
		"error":            "#FF5555",
	}

	report := schemeWCAG(scheme)
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.Body == nil || report.Body.Level != "AAA" {
		t.Fatalf("expected body AAA, got %+v", report.Body)
	}
	if report.Accent == nil || report.Accent.Level != "fail" {
		t.Fatalf("expected accent fail, got %+v", report.Accent)
	}
	if report.Level != "fail" {
		t.Fatalf("expected combined level to follow the accent failure, got %s", report.Level)
	}
}

func TestBreakdownRollsAccentsIntoFlavors(t *testing.T) {
	// Two accents on one flavor collapse to a single row carrying the worse level.
	theme := &models.Theme{
		Dark: map[string]interface{}{"surfaceText": "#FFFFFF", "surface": "#000000"},
		Variants: &models.ThemeVariants{
			Type:     "multi",
			Defaults: map[string]*models.ThemeModeDefaults{"dark": {Flavor: "mocha", Accent: "good"}},
			Flavors: []models.ThemeFlavor{
				{ID: "mocha", Name: "Mocha", Dark: map[string]interface{}{"surfaceContainer": "#000000"}},
			},
			Accents: []map[string]interface{}{
				{"id": "good", "mocha": map[string]interface{}{"primary": "#FFFFFF", "primaryText": "#000000"}},
				{"id": "bad", "mocha": map[string]interface{}{"primary": "#111111", "primaryText": "#000000"}},
			},
		},
	}

	wcag := computeThemeWCAG(theme)
	if wcag == nil || wcag.Dark == nil {
		t.Fatal("expected a dark report")
	}
	if len(wcag.Dark.Breakdown) != 1 {
		t.Fatalf("expected accents to roll into one flavor row, got %+v", wcag.Dark.Breakdown)
	}

	row := wcag.Dark.Breakdown[0]
	if row.Name != "Mocha" || row.Mode != "dark" {
		t.Fatalf("expected the flavor display name, got %+v", row)
	}
	if row.Level != "fail" {
		t.Fatalf("expected the worst accent to drive the row, got %s", row.Level)
	}
	if row.BodyLevel != "AAA" {
		t.Fatalf("expected body text to stay AAA despite the accent, got %s", row.BodyLevel)
	}
}

func TestBreakdownUsesModeNameWithoutVariants(t *testing.T) {
	theme := &models.Theme{
		Dark:  map[string]interface{}{"surfaceText": "#FFFFFF", "surface": "#000000"},
		Light: map[string]interface{}{"surfaceText": "#000000", "surface": "#FFFFFF"},
	}

	wcag := computeThemeWCAG(theme)
	if wcag.Dark.Breakdown[0].Name != "Dark" {
		t.Fatalf("expected row named Dark, got %+v", wcag.Dark.Breakdown)
	}
	if wcag.Light.Breakdown[0].Name != "Light" {
		t.Fatalf("expected row named Light, got %+v", wcag.Light.Breakdown)
	}
}

func TestComputeThemeWCAGNoColors(t *testing.T) {
	if wcag := computeThemeWCAG(&models.Theme{}); wcag != nil {
		t.Fatalf("expected nil for theme without colors, got %+v", wcag)
	}
}
