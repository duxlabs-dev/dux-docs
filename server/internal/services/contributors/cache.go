// Package contributors keeps a public feed of GitHub code contributors and
// POEditor translators, refreshed on a schedule and persisted across restarts.
package contributors

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type Config struct {
	GitHub GitHub
	// Repos are "owner/name"; empty disables code contributors.
	Repos   []string
	Exclude []string
	// POEditor may be nil (no token), which disables translators.
	POEditor   POEditor
	ProjectIDs []string
	// DataDir holds the last good snapshot; empty keeps it in memory only.
	DataDir string
}

type Cache struct {
	mu sync.RWMutex
	// Serializes refreshes so a slow one cannot publish over a newer result.
	refreshMu sync.Mutex

	gh       GitHub
	repos    []string
	exclude  map[string]bool
	po       POEditor
	projects []string
	path     string

	snap  snapshot
	feed  models.ContributorsFeed
	ready bool
	now   func() time.Time
}

// snapshot holds only what the feed publishes: no POEditor emails.
type snapshot struct {
	Code          []models.CodeContributor `json:"code"`
	CodeAt        time.Time                `json:"codeAt"`
	Translators   []models.Translator      `json:"translators"`
	TranslatorsAt time.Time                `json:"translatorsAt"`
}

func NewCache(cfg Config) *Cache {
	c := &Cache{
		gh:       cfg.GitHub,
		repos:    clean(cfg.Repos),
		exclude:  map[string]bool{},
		po:       cfg.POEditor,
		projects: clean(cfg.ProjectIDs),
		now:      time.Now,
	}
	for _, login := range clean(cfg.Exclude) {
		c.exclude[strings.ToLower(login)] = true
	}
	if c.gh == nil {
		c.repos = nil
	}
	if c.po == nil || len(c.projects) == 0 {
		log.Warn("POEDITOR_API_TOKEN or POEDITOR_PROJECT_IDS not set; the contributors feed will not list translators")
		c.po, c.projects = nil, nil
	}
	if cfg.DataDir != "" {
		c.path = filepath.Join(cfg.DataDir, "contributors.json")
		if err := readJSON(c.path, &c.snap); err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Warn("Failed to load contributors cache", "err", err)
			c.snap = snapshot{}
		}
	}
	c.publish()
	return c
}

func clean(in []string) []string {
	var out []string
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// publish rebuilds the feed from snap. Callers hold mu, or own c exclusively.
func (c *Cache) publish() {
	codeReady := len(c.repos) == 0 || !c.snap.CodeAt.IsZero()
	translatorsReady := c.po == nil || !c.snap.TranslatorsAt.IsZero()
	c.ready = codeReady && translatorsReady

	feed := models.ContributorsFeed{Code: []models.CodeContributor{}, Translators: []models.Translator{}}
	if len(c.repos) > 0 && c.snap.Code != nil {
		feed.Code = c.snap.Code
	}
	if c.po != nil && c.snap.Translators != nil {
		feed.Translators = c.snap.Translators
	}
	feed.UpdatedAt = c.snap.CodeAt
	if c.snap.TranslatorsAt.After(feed.UpdatedAt) {
		feed.UpdatedAt = c.snap.TranslatorsAt
	}
	c.feed = feed
}

// Refresh fetches each source independently; a failed source keeps its last
// good data so one outage never blanks the other half of the page.
func (c *Cache) Refresh(ctx context.Context) error {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	c.mu.RLock()
	next := c.snap
	c.mu.RUnlock()

	var errs []error
	if len(c.repos) > 0 {
		code, err := fetchCode(ctx, c.gh, c.repos, c.exclude)
		if err != nil {
			errs = append(errs, err)
		} else {
			next.Code, next.CodeAt = code, c.now().UTC()
		}
	}
	if c.po != nil {
		translators, err := fetchTranslators(ctx, c.po, c.projects)
		if err != nil {
			errs = append(errs, err)
		} else {
			next.Translators, next.TranslatorsAt = translators, c.now().UTC()
		}
	}

	c.mu.Lock()
	changed := !next.CodeAt.Equal(c.snap.CodeAt) || !next.TranslatorsAt.Equal(c.snap.TranslatorsAt)
	c.snap = next
	c.publish()
	c.mu.Unlock()

	if changed && c.path != "" {
		if err := writeJSON(c.path, next); err != nil {
			log.Warn("Failed to persist contributors", "err", err)
		}
	}
	if changed {
		log.Infof("Contributors cache refreshed: %d code contributors, %d translators", len(next.Code), len(next.Translators))
	}
	return errors.Join(errs...)
}

func (c *Cache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

// The feed is replaced wholesale on change and never mutated, so callers may share it.
func (c *Cache) GetFeed() models.ContributorsFeed {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feed
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// writeJSON replaces path atomically so a crash mid-write never truncates it.
func writeJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
