package registry

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const moderationGrace = time.Minute

type PreviewSyncer interface {
	Sync(ctx context.Context, plugins []models.Plugin) []models.Plugin
}

type Cache struct {
	mu          sync.RWMutex
	plugins     []models.Plugin
	moderatedAt map[string]time.Time
	parser      *Parser
	lastUpdate  time.Time
	ready       bool
	persistPath string
	previews    PreviewSyncer
}

type pluginSnapshot struct {
	Plugins    []models.Plugin `json:"plugins"`
	LastUpdate time.Time       `json:"last_update"`
}

func NewCache(githubToken, persistPath string) *Cache {
	return &Cache{
		plugins:     []models.Plugin{},
		moderatedAt: make(map[string]time.Time),
		parser:      NewParser(githubToken),
		persistPath: persistPath,
	}
}

func (c *Cache) SetPreviewSyncer(s PreviewSyncer) {
	c.previews = s
}

func (c *Cache) Initialize(ctx context.Context) error {
	if c.persistPath != "" {
		if err := c.loadFromDisk(); err == nil {
			c.mu.RLock()
			n := len(c.plugins)
			c.mu.RUnlock()
			log.Infof("Plugin cache loaded from disk with %d plugins; refreshing in background", n)
		} else if !errors.Is(err, fs.ErrNotExist) {
			log.Warn("Failed to load plugin cache from disk", "err", err)
		}
	}
	return c.Refresh(ctx)
}

func (c *Cache) Refresh(ctx context.Context) error {
	log.Info("Refreshing plugin cache...")

	plugins, err := c.parser.FetchPlugins(ctx, c.GetPlugins())
	if err != nil {
		return err
	}

	feedbackStarted := time.Now()
	feedback, err := c.parser.FetchFeedback(ctx)
	if err != nil {
		log.Warnf("Failed to fetch plugin feedback: %v", err)
	}

	c.mu.Lock()
	carryFeedback(plugins, c.plugins)
	mergeFeedback(plugins, feedback)
	c.keepRecentModeration(plugins, feedbackStarted.Add(-moderationGrace))
	c.mu.Unlock()

	if c.previews != nil {
		plugins = c.previews.Sync(ctx, plugins)
	}

	c.mu.Lock()
	c.plugins = plugins
	c.lastUpdate = time.Now()
	c.ready = true
	c.mu.Unlock()

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist plugin cache", "err", err)
	}

	log.Infof("Plugin cache refreshed with %d plugins", len(plugins))
	return nil
}

func (c *Cache) RefreshFeedback(ctx context.Context) error {
	feedback, err := c.parser.FetchFeedback(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	mergeFeedback(c.plugins, feedback)
	plugins := make([]models.Plugin, len(c.plugins))
	copy(plugins, c.plugins)
	c.mu.Unlock()

	if c.previews != nil {
		plugins = c.previews.Sync(ctx, plugins)
		c.mu.Lock()
		c.plugins = plugins
		c.mu.Unlock()
	}

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist plugin cache", "err", err)
	}

	log.Info("Plugin feedback refreshed")
	return nil
}

func (c *Cache) loadFromDisk() error {
	data, err := os.ReadFile(c.persistPath)
	if err != nil {
		return err
	}

	var snap pluginSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}

	c.mu.Lock()
	c.plugins = snap.Plugins
	c.lastUpdate = snap.LastUpdate
	c.ready = true
	c.mu.Unlock()
	return nil
}

func (c *Cache) saveToDisk() error {
	if c.persistPath == "" {
		return nil
	}

	c.mu.RLock()
	snap := pluginSnapshot{
		Plugins:    c.plugins,
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

// ApplyStatus updates a plugin's status labels in place so a moderation action is
// reflected immediately, without waiting for the next GitHub re-fetch (which can lag
// behind a just-applied label due to API eventual consistency).
func (c *Cache) ApplyStatus(pluginID, status string, add bool) {
	c.mu.Lock()
	for i := range c.plugins {
		if c.plugins[i].ID == pluginID {
			c.plugins[i].Status = upsertStatus(c.plugins[i].Status, status, add)
			break
		}
	}
	c.moderatedAt[pluginID] = time.Now()
	c.mu.Unlock()

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist plugin cache after status update", "err", err)
	}
}

// ApplySimilar updates a plugin's related-plugin list in place so a /similar moderation
// action is reflected immediately, without waiting for the next GitHub re-fetch.
func (c *Cache) ApplySimilar(pluginID, similarID string, add bool) {
	c.mu.Lock()
	for i := range c.plugins {
		if c.plugins[i].ID == pluginID {
			c.plugins[i].Similar = upsertStatus(c.plugins[i].Similar, similarID, add)
			break
		}
	}
	c.moderatedAt[pluginID] = time.Now()
	c.mu.Unlock()

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist plugin cache after similar update", "err", err)
	}
}

func (c *Cache) keepRecentModeration(plugins []models.Plugin, since time.Time) {
	current := pluginsByID(c.plugins)
	for i := range plugins {
		cached, ok := current[plugins[i].ID]
		if !ok || c.moderatedAt[plugins[i].ID].Before(since) {
			continue
		}
		plugins[i].Status = cached.Status
		plugins[i].Similar = cached.Similar
	}
}

func (c *Cache) PluginByID(id string) (models.Plugin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, plugin := range c.plugins {
		if plugin.ID == id {
			return plugin, true
		}
	}
	return models.Plugin{}, false
}

func (c *Cache) PluginByIssue(number int) (models.Plugin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, plugin := range c.plugins {
		if plugin.IssueNumber == number {
			return plugin, true
		}
	}
	return models.Plugin{}, false
}

func upsertStatus(statuses []string, status string, add bool) []string {
	var out []string
	for _, s := range statuses {
		if s != status {
			out = append(out, s)
		}
	}
	if add {
		out = append(out, status)
	}
	return out
}

func (c *Cache) RepoOwner(pluginID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, plugin := range c.plugins {
		if plugin.ID != pluginID {
			continue
		}
		_, owner, _, err := parseRepoURL(plugin.Repo)
		if err != nil {
			return "", false
		}
		return owner, true
	}
	return "", false
}

func (c *Cache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

func (c *Cache) GetPlugins() []models.Plugin {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pluginsCopy := make([]models.Plugin, len(c.plugins))
	copy(pluginsCopy, c.plugins)
	return pluginsCopy
}

func (c *Cache) GetLastUpdate() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdate
}

type FilterOptions struct {
	Categories    []string
	Compositors   []string
	FirstParty    bool
	Capabilities  []string
	ExcludeStatus []string
}

func (c *Cache) FilterPlugins(opts FilterOptions) []models.Plugin {
	plugins := c.GetPlugins()

	var filtered []models.Plugin
	for _, plugin := range plugins {
		if !matchesFilter(plugin, opts) {
			continue
		}
		filtered = append(filtered, plugin)
	}

	return filtered
}

func matchesFilter(plugin models.Plugin, opts FilterOptions) bool {
	if len(opts.Categories) > 0 && !slices.Contains(opts.Categories, plugin.Category) {
		return false
	}

	if opts.FirstParty && !plugin.FirstParty {
		return false
	}

	if len(opts.Compositors) > 0 && !slices.Contains(plugin.Compositors, "any") && !anyOverlap(plugin.Compositors, opts.Compositors) {
		return false
	}

	if len(opts.Capabilities) > 0 && !anyOverlap(plugin.Capabilities, opts.Capabilities) {
		return false
	}

	if anyOverlap(plugin.Status, opts.ExcludeStatus) {
		return false
	}

	return true
}

func anyOverlap(values, wanted []string) bool {
	return slices.ContainsFunc(values, func(value string) bool {
		return slices.Contains(wanted, value)
	})
}
