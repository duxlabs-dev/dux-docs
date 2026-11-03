package registry

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type ThemeCache struct {
	mu          sync.RWMutex
	themes      []models.Theme
	parser      *Parser
	lastUpdate  time.Time
	ready       bool
	persistPath string
}

type themeSnapshot struct {
	Themes     []models.Theme `json:"themes"`
	LastUpdate time.Time      `json:"last_update"`
}

func NewThemeCache(githubToken, persistPath string) *ThemeCache {
	return &ThemeCache{
		themes:      []models.Theme{},
		parser:      NewParser(githubToken),
		persistPath: persistPath,
	}
}

func (c *ThemeCache) Initialize(ctx context.Context) error {
	if c.persistPath != "" {
		if err := c.loadFromDisk(); err == nil {
			c.mu.RLock()
			n := len(c.themes)
			c.mu.RUnlock()
			log.Infof("Theme cache loaded from disk with %d themes; refreshing in background", n)
		} else if !errors.Is(err, fs.ErrNotExist) {
			log.Warn("Failed to load theme cache from disk", "err", err)
		}
	}
	return c.Refresh(ctx)
}

func (c *ThemeCache) Refresh(ctx context.Context) error {
	log.Info("Refreshing theme cache...")

	themes, err := c.parser.FetchThemes(ctx, c.GetThemes())
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.themes = themes
	c.lastUpdate = time.Now()
	c.ready = true
	c.mu.Unlock()

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist theme cache", "err", err)
	}

	log.Infof("Theme cache refreshed with %d themes", len(themes))
	return nil
}

func (c *ThemeCache) loadFromDisk() error {
	data, err := os.ReadFile(c.persistPath)
	if err != nil {
		return err
	}
	var snap themeSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	c.mu.Lock()
	c.themes = snap.Themes
	c.lastUpdate = snap.LastUpdate
	c.ready = true
	c.mu.Unlock()
	return nil
}

func (c *ThemeCache) saveToDisk() error {
	if c.persistPath == "" {
		return nil
	}
	c.mu.RLock()
	snap := themeSnapshot{
		Themes:     c.themes,
		LastUpdate: c.lastUpdate,
	}
	c.mu.RUnlock()

	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.persistPath), 0o755); err != nil {
		return err
	}
	tmp := c.persistPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.persistPath)
}

func (c *ThemeCache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

func (c *ThemeCache) GetThemes() []models.Theme {
	c.mu.RLock()
	defer c.mu.RUnlock()

	themesCopy := make([]models.Theme, len(c.themes))
	copy(themesCopy, c.themes)
	return themesCopy
}

func (c *ThemeCache) GetLastUpdate() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdate
}
