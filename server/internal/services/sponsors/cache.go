package sponsors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type Cache struct {
	mu sync.RWMutex
	// Serializes GitHub refreshes so a slow one cannot publish over a newer result.
	refreshMu sync.Mutex
	kofiMu    sync.Mutex

	gql    GraphQL
	owners []string
	// kofiPath is the only record of Ko-fi supporters.
	githubPath string
	kofiPath   string

	github      []githubSponsor
	githubReady bool
	kofi        kofiState
	// kofiLoadErr blocks writes so an unreadable file is never overwritten.
	kofiLoadErr error
	// kofiDirty marks in-memory Ko-fi changes that failed to persist.
	kofiDirty bool
	feed      models.SponsorsFeed
	now       func() time.Time
}

type githubSnapshot struct {
	Sponsors  []githubSponsor `json:"sponsors"`
	FetchedAt time.Time       `json:"fetchedAt"`
}

// NewCache loads persisted state before returning, so an early webhook cannot
// overwrite it. gql may be nil (no token); an empty dataDir keeps Ko-fi in memory.
func NewCache(gql GraphQL, owners []string, dataDir string) *Cache {
	var cleaned []string
	for _, o := range owners {
		if o = strings.TrimSpace(o); o != "" {
			cleaned = append(cleaned, o)
		}
	}
	c := &Cache{
		gql:    gql,
		owners: cleaned,
		kofi:   newKofiState(),
		now:    time.Now,
	}
	if dataDir == "" {
		log.Warn("CACHE_DIR is not set; Ko-fi supporters will not survive a restart")
	} else {
		c.githubPath = filepath.Join(dataDir, "sponsors-github.json")
		c.kofiPath = filepath.Join(dataDir, "sponsors-kofi.json")
		c.load()
	}
	if gql == nil {
		log.Warn("No GitHub token; the sponsors feed will only list Ko-fi supporters")
		c.githubReady = true
	}
	c.feed = buildFeed(c.github, c.kofi, c.now())
	return c
}

func (c *Cache) load() {
	if err := readJSON(c.kofiPath, &c.kofi); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Error("Failed to load Ko-fi supporters; Ko-fi webhooks will fail until the file is fixed", "path", c.kofiPath, "err", err)
		c.kofiLoadErr = err
		c.kofi = newKofiState()
	}
	if c.kofi.Supporters == nil {
		c.kofi.Supporters = map[string]*kofiSupporter{}
	}
	if c.kofi.Private == nil {
		c.kofi.Private = map[string]time.Time{}
	}
	var snap githubSnapshot
	if err := readJSON(c.githubPath, &snap); err == nil {
		c.github = snap.Sponsors
		c.githubReady = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		log.Warn("Failed to load GitHub sponsors cache", "err", err)
	}
}

func (c *Cache) Refresh(ctx context.Context) error {
	if c.gql == nil {
		return nil
	}
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	sponsors, err := fetchGithubSponsors(ctx, c.gql, c.owners)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.github = sponsors
	c.githubReady = true
	c.feed = buildFeed(c.github, c.kofi, c.now())
	c.mu.Unlock()

	if c.githubPath != "" {
		if err := writeJSON(c.githubPath, githubSnapshot{Sponsors: sponsors, FetchedAt: c.now()}); err != nil {
			log.Warn("Failed to persist GitHub sponsors", "err", err)
		}
	}
	log.Infof("Sponsors cache refreshed with %d GitHub sponsors", len(sponsors))
	return nil
}

// RecordKofi applies a verified Ko-fi webhook and persists the result.
func (c *Cache) RecordKofi(p KofiPayload) error {
	// Held across apply and write so concurrent webhooks persist in order.
	c.kofiMu.Lock()
	defer c.kofiMu.Unlock()

	// Failing here makes Ko-fi retry the webhook once the file is repaired.
	if c.kofiLoadErr != nil {
		return fmt.Errorf("ko-fi store unreadable: %w", c.kofiLoadErr)
	}

	c.mu.Lock()
	changed := c.kofi.apply(p, c.now())
	// A retry after a failed write is a no-op for apply but must still persist.
	if !changed && !c.kofiDirty {
		c.mu.Unlock()
		return nil
	}
	data, err := json.Marshal(c.kofi)
	if changed {
		c.feed = buildFeed(c.github, c.kofi, c.now())
	}
	c.mu.Unlock()

	if err != nil || c.kofiPath == "" {
		return err
	}
	err = writeFile(c.kofiPath, data)
	c.mu.Lock()
	c.kofiDirty = err != nil
	c.mu.Unlock()
	return err
}

func (c *Cache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.githubReady
}

// The feed is replaced wholesale on change and never mutated, so callers may share it.
func (c *Cache) GetFeed() models.SponsorsFeed {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feed
}

// buildFeed never merges entries across platforms: names are user-controlled,
// so matching on them would let one person take over another's entry.
func buildFeed(gh []githubSponsor, kofi kofiState, now time.Time) models.SponsorsFeed {
	feed := models.SponsorsFeed{Current: []models.Sponsor{}, Past: []models.Sponsor{}, UpdatedAt: now.UTC()}
	add := func(s models.Sponsor, active bool) {
		if active {
			feed.Current = append(feed.Current, s)
		} else {
			feed.Past = append(feed.Past, s)
		}
	}

	for _, s := range gh {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			name = s.Login
		}
		add(models.Sponsor{Name: name, Login: s.Login, AvatarURL: s.AvatarURL, URL: s.URL, Source: "github"}, s.Active)
	}

	supporters := make([]*kofiSupporter, 0, len(kofi.Supporters))
	for _, s := range kofi.Supporters {
		supporters = append(supporters, s)
	}
	sort.Slice(supporters, func(i, j int) bool {
		if !supporters[i].LastPayment.Equal(supporters[j].LastPayment) {
			return supporters[i].LastPayment.After(supporters[j].LastPayment)
		}
		return supporters[i].Name < supporters[j].Name
	})
	for _, s := range supporters {
		add(models.Sponsor{Name: s.Name, Source: "kofi"}, now.Sub(s.LastPayment) < kofiActiveWindow)
	}
	return feed
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func writeJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

// writeFile replaces path atomically so a crash mid-write never truncates it.
func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		// Without this a power loss can leave a renamed but empty file.
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	if dir, err := os.Open(filepath.Dir(path)); err == nil {
		_ = dir.Sync()
		dir.Close()
	}
	return nil
}
