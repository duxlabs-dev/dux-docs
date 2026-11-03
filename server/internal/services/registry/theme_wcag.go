package registry

import (
	"math"
	"strconv"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

// Thresholds from WCAG 2.2 SC 1.4.3 / 1.4.6
// https://www.w3.org/TR/WCAG22/#contrast-minimum
const (
	wcagAARatio  = 4.5
	wcagAAARatio = 7.0
)

// Pairs mirror what DMS/quickshell actually renders: bars, popouts, and modals
// fill with surfaceContainer, nested cards with surfaceContainerHigh (see
// DankMaterialShell Common/Theme.qml nestedSurface), window bases with surface.
// Body covers the text read constantly; accent covers primary, which DMS draws
// as bar text (Clock widget) and as filled button labels.
var wcagBodyPairs = [][2]string{
	{"surfaceText", "surface"},
	{"surfaceText", "surfaceContainer"},
	{"surfaceText", "surfaceContainerHigh"},
	{"surfaceText", "surfaceContainerHighest"},
	{"surfaceVariantText", "surface"},
	{"surfaceVariantText", "surfaceContainer"},
	{"surfaceVariantText", "surfaceContainerHigh"},
}

var wcagAccentPairs = [][2]string{
	{"primaryText", "primary"},
	{"primary", "surfaceContainer"},
}

var wcagTextPairs = append(append([][2]string{}, wcagBodyPairs...), wcagAccentPairs...)

// Status colors render as standalone icons and badges, so they get the 3:1
// non-text minimum from WCAG 2.2 SC 1.4.11
// https://www.w3.org/TR/WCAG22/#non-text-contrast
// Outline is excluded: it is a divider color that DMS draws at 12% alpha
// (Theme.outlineMedium), which SC 1.4.11 exempts as decorative.
var wcagNonTextPairs = [][2]string{
	{"error", "surfaceContainer"},
	{"warning", "surfaceContainer"},
	{"info", "surfaceContainer"},
}

const wcagNonTextRatio = 3.0

var wcagLevelRank = map[string]int{"fail": 0, "AA": 1, "AAA": 2}

var wcagModeLabels = map[string]string{"dark": "Dark", "light": "Light"}

type wcagRGB struct {
	r, g, b float64
}

func parseHexColor(value interface{}) (wcagRGB, bool) {
	s, ok := value.(string)
	if !ok {
		return wcagRGB{}, false
	}
	if len(s) != 7 || s[0] != '#' {
		return wcagRGB{}, false
	}

	n, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return wcagRGB{}, false
	}

	return wcagRGB{
		r: float64(n >> 16 & 0xff),
		g: float64(n >> 8 & 0xff),
		b: float64(n & 0xff),
	}, true
}

// https://www.w3.org/TR/WCAG22/#dfn-relative-luminance
func relativeLuminance(c wcagRGB) float64 {
	linearize := func(channel float64) float64 {
		channel /= 255
		if channel <= 0.03928 {
			return channel / 12.92
		}
		return math.Pow((channel+0.055)/1.055, 2.4)
	}
	return 0.2126*linearize(c.r) + 0.7152*linearize(c.g) + 0.0722*linearize(c.b)
}

// https://www.w3.org/TR/WCAG22/#dfn-contrast-ratio
func contrastRatio(a, b wcagRGB) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if lb > la {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func wcagLevel(ratio float64) string {
	if ratio >= wcagAAARatio {
		return "AAA"
	}
	if ratio >= wcagAARatio {
		return "AA"
	}
	return "fail"
}

func worstRatio(scheme map[string]interface{}, pairs [][2]string) (float64, []string) {
	minRatio := math.Inf(1)
	var worstPair []string
	for _, pair := range pairs {
		fg, ok := parseHexColor(scheme[pair[0]])
		if !ok {
			continue
		}
		bg, ok := parseHexColor(scheme[pair[1]])
		if !ok {
			continue
		}

		ratio := contrastRatio(fg, bg)
		if ratio >= minRatio {
			continue
		}
		minRatio = ratio
		worstPair = []string{pair[0], pair[1]}
	}

	return minRatio, worstPair
}

func groupWCAG(scheme map[string]interface{}, pairs [][2]string, level func(float64) string) *models.ThemeWCAGGroup {
	ratio, pair := worstRatio(scheme, pairs)
	if pair == nil {
		return nil
	}

	return &models.ThemeWCAGGroup{
		Level:     level(ratio),
		MinRatio:  math.Round(ratio*100) / 100,
		WorstPair: pair,
	}
}

func nonTextLevel(ratio float64) string {
	if ratio >= wcagNonTextRatio {
		return "AA"
	}
	return "fail"
}

func schemeWCAG(scheme map[string]interface{}) *models.ThemeWCAGMode {
	minRatio, worstPair := worstRatio(scheme, wcagTextPairs)
	if worstPair == nil {
		return nil
	}

	report := &models.ThemeWCAGMode{
		Level:     wcagLevel(minRatio),
		MinRatio:  math.Round(minRatio*100) / 100,
		WorstPair: worstPair,
		Body:      groupWCAG(scheme, wcagBodyPairs, wcagLevel),
		Accent:    groupWCAG(scheme, wcagAccentPairs, wcagLevel),
		NonText:   groupWCAG(scheme, wcagNonTextPairs, nonTextLevel),
	}

	// SC 1.4.11 is itself a Level AA criterion, so failing it fails AA outright.
	if report.NonText != nil && report.NonText.Level == "fail" {
		report.Level = "fail"
	}
	return report
}

func mergeSchemes(layers ...map[string]interface{}) map[string]interface{} {
	merged := map[string]interface{}{}
	for _, layer := range layers {
		for key, value := range layer {
			merged[key] = value
		}
	}
	return merged
}

// wcagConfig is one thing a user can actually select. Themes without variants
// offer just the mode itself; flavor themes offer every flavor/accent combo.
// Group is the unit the breakdown rolls up to.
type wcagConfig struct {
	key    string
	group  string
	label  string
	scheme map[string]interface{}
}

func modeConfigs(theme *models.Theme, mode string) ([]wcagConfig, string) {
	base := theme.Dark
	if mode == "light" {
		base = theme.Light
	}

	plain := []wcagConfig{{label: wcagModeLabels[mode], scheme: base}}
	variants := theme.Variants
	if variants == nil {
		return plain, ""
	}
	if variants.Type == "multi" {
		return multiVariantConfigs(variants, base, mode)
	}
	if len(variants.Options) == 0 {
		return plain, ""
	}

	configs := make([]wcagConfig, 0, len(variants.Options))
	for _, option := range variants.Options {
		if option.ID == "" {
			continue
		}
		override := option.Dark
		if mode == "light" {
			override = option.Light
		}
		configs = append(configs, wcagConfig{
			key:    option.ID,
			group:  option.ID,
			label:  wcagLabel(option.Name, option.ID),
			scheme: mergeSchemes(base, override),
		})
	}
	return configs, variants.Default
}

// Accents roll up into their flavor: authors document flavors as the unit a
// user picks, and listing every combo would run to dozens of rows.
func multiVariantConfigs(variants *models.ThemeVariants, base map[string]interface{}, mode string) ([]wcagConfig, string) {
	configs := []wcagConfig{}
	for _, flavor := range variants.Flavors {
		flavorColors := flavor.Dark
		if mode == "light" {
			flavorColors = flavor.Light
		}
		if flavor.ID == "" || flavorColors == nil {
			continue
		}

		for _, accent := range variants.Accents {
			aid, _ := accent["id"].(string)
			if aid == "" {
				continue
			}
			accentColors, _ := accent[flavor.ID].(map[string]interface{})
			configs = append(configs, wcagConfig{
				key:    flavor.ID + "-" + aid,
				group:  flavor.ID,
				label:  wcagLabel(flavor.Name, flavor.ID),
				scheme: mergeSchemes(base, flavorColors, accentColors),
			})
		}
	}

	defaults := variants.Defaults[mode]
	if defaults == nil {
		return configs, ""
	}
	return configs, defaults.Flavor + "-" + defaults.Accent
}

func wcagLabel(name, id string) string {
	if name != "" {
		return name
	}
	return id
}

func worstModeWCAG(reports map[string]*models.ThemeWCAGMode) *models.ThemeWCAGMode {
	var worst *models.ThemeWCAGMode
	for _, report := range reports {
		if worst == nil {
			worst = report
			continue
		}
		if wcagLevelRank[report.Level] < wcagLevelRank[worst.Level] {
			worst = report
			continue
		}
		if report.Level == worst.Level && report.MinRatio < worst.MinRatio {
			worst = report
		}
	}
	return worst
}

type wcagGroupLevels struct {
	name      string
	level     string
	bodyLevel string
}

func modeWCAG(theme *models.Theme, mode string) *models.ThemeWCAGMode {
	configs, defaultKey := modeConfigs(theme, mode)
	reports := map[string]*models.ThemeWCAGMode{}
	order := []string{}
	groups := map[string]*wcagGroupLevels{}

	for _, config := range configs {
		report := schemeWCAG(config.scheme)
		if report == nil {
			continue
		}
		reports[config.key] = report

		bodyLevel := "fail"
		if report.Body != nil {
			bodyLevel = report.Body.Level
		}

		group, seen := groups[config.group]
		if !seen {
			groups[config.group] = &wcagGroupLevels{
				name:      config.label,
				level:     report.Level,
				bodyLevel: bodyLevel,
			}
			order = append(order, config.group)
			continue
		}
		if wcagLevelRank[report.Level] < wcagLevelRank[group.level] {
			group.level = report.Level
		}
		if wcagLevelRank[bodyLevel] < wcagLevelRank[group.bodyLevel] {
			group.bodyLevel = bodyLevel
		}
	}

	if len(reports) == 0 {
		return nil
	}

	// The headline stays the default config, which is what a user gets on first
	// apply; breakdown carries every other config so nothing is over-promised.
	primary := reports[defaultKey]
	if primary == nil {
		primary = worstModeWCAG(reports)
	}

	result := *primary
	if len(reports) > 1 {
		result.Variants = make(map[string]string, len(reports))
		for key, report := range reports {
			result.Variants[key] = report.Level
		}
	}

	result.Breakdown = make([]models.ThemeWCAGBreakdown, 0, len(order))
	for _, key := range order {
		group := groups[key]
		result.Breakdown = append(result.Breakdown, models.ThemeWCAGBreakdown{
			Name:      group.name,
			Mode:      mode,
			Level:     group.level,
			BodyLevel: group.bodyLevel,
		})
	}
	return &result
}

func computeThemeWCAG(theme *models.Theme) *models.ThemeWCAG {
	dark := modeWCAG(theme, "dark")
	light := modeWCAG(theme, "light")
	if dark == nil && light == nil {
		return nil
	}

	level := "AAA"
	for _, mode := range []*models.ThemeWCAGMode{dark, light} {
		if mode == nil {
			continue
		}
		if wcagLevelRank[mode.Level] < wcagLevelRank[level] {
			level = mode.Level
		}
	}

	return &models.ThemeWCAG{Level: level, Dark: dark, Light: light}
}
