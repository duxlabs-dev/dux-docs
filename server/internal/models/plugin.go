package models

import "time"

type RegistryPlugin struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	Category     string   `json:"category"`
	Repo         string   `json:"repo"`
	Path         string   `json:"path,omitempty"`
	Author       string   `json:"author"`
	FirstParty   bool     `json:"firstParty,omitempty"`
	Featured     bool     `json:"featured,omitempty"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Compositors  []string `json:"compositors"`
	Distro       []string `json:"distro"`
	Screenshot   string   `json:"screenshot,omitempty"`
	RequiresDMS  string   `json:"requires_dms,omitempty"`
}

type PluginMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Icon        string            `json:"icon"`
	FirstParty  bool              `json:"firstParty"`
	Type        string            `json:"type,omitempty"`
	Component   string            `json:"component"`
	Components  map[string]string `json:"components,omitempty"`
	Settings    string            `json:"settings,omitempty"`
	Permissions []string          `json:"permissions"`
}

type Plugin struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Capabilities []string  `json:"capabilities"`
	Category     string    `json:"category"`
	Repo         string    `json:"repo"`
	Author       string    `json:"author"`
	FirstParty   bool      `json:"firstParty"`
	Featured     bool      `json:"featured"`
	Description  string    `json:"description"`
	Dependencies []string  `json:"dependencies"`
	Compositors  []string  `json:"compositors"`
	Distro       []string  `json:"distro"`
	Screenshot   string    `json:"screenshot"`
	PreviewURL   string    `json:"previewUrl,omitempty"`
	ThumbURL     string    `json:"thumbUrl,omitempty"`
	RequiresDMS  string    `json:"requires_dms,omitempty"`
	Version      string    `json:"version"`
	Icon         string    `json:"icon,omitempty"`
	Permissions  []string  `json:"permissions,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
	Upvotes      int       `json:"upvotes"`
	IssueURL     string    `json:"issueUrl,omitempty"`
	IssueNumber  int       `json:"issueNumber,omitempty"`
	Status       []string  `json:"status,omitempty"`
	Similar      []string  `json:"similar,omitempty"`
}

type ThemeVariantOption struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Dark  map[string]interface{} `json:"dark,omitempty"`
	Light map[string]interface{} `json:"light,omitempty"`
}

type ThemeFlavor struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Dark  map[string]interface{} `json:"dark,omitempty"`
	Light map[string]interface{} `json:"light,omitempty"`
}

type ThemeAccent struct {
	ID     string                            `json:"id"`
	Name   string                            `json:"name"`
	Colors map[string]map[string]interface{} `json:"-"`
}

type ThemeModeDefaults struct {
	Flavor string `json:"flavor"`
	Accent string `json:"accent"`
}

type ThemeVariants struct {
	Type     string                        `json:"type,omitempty"`
	Default  string                        `json:"default,omitempty"`
	Options  []ThemeVariantOption          `json:"options,omitempty"`
	Defaults map[string]*ThemeModeDefaults `json:"defaults,omitempty"`
	Flavors  []ThemeFlavor                 `json:"flavors,omitempty"`
	Accents  []map[string]interface{}      `json:"accents,omitempty"`
}

type ThemeWCAGGroup struct {
	Level     string   `json:"level"`
	MinRatio  float64  `json:"minRatio"`
	WorstPair []string `json:"worstPair,omitempty"`
}

type ThemeWCAGBreakdown struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Level     string `json:"level"`
	BodyLevel string `json:"bodyLevel"`
}

type ThemeWCAGMode struct {
	Level     string            `json:"level"`
	MinRatio  float64           `json:"minRatio"`
	WorstPair []string          `json:"worstPair,omitempty"`
	Body      *ThemeWCAGGroup   `json:"body,omitempty"`
	Accent    *ThemeWCAGGroup   `json:"accent,omitempty"`
	NonText   *ThemeWCAGGroup   `json:"nonText,omitempty"`
	Variants  map[string]string `json:"variants,omitempty"`

	Breakdown []ThemeWCAGBreakdown `json:"breakdown,omitempty"`
}

type ThemeWCAG struct {
	Level string         `json:"level"`
	Dark  *ThemeWCAGMode `json:"dark,omitempty"`
	Light *ThemeWCAGMode `json:"light,omitempty"`
}

type Theme struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Dark        map[string]interface{} `json:"dark"`
	Light       map[string]interface{} `json:"light"`
	Variants    *ThemeVariants         `json:"variants,omitempty"`
	WCAG        *ThemeWCAG             `json:"wcag,omitempty"`
	PreviewURL  string                 `json:"previewUrl"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
