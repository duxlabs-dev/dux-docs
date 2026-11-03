package releases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const (
	repoOwner    = "duxlabs-dev"
	repoName     = "dux-shell"
	masterBranch = "master"
	codenamePath = "quickshell/CODENAME"
	fetchCount   = 10
	keepCount    = 6
)

type Cache struct {
	mu sync.RWMutex
	// Serializes whole refreshes so a slow one cannot publish or persist over a newer result.
	refreshMu   sync.Mutex
	client      *github.Client
	blogDir     string
	persistPath string
	feed        models.ReleasesFeed
	feedETag    string
	latestETag  string
	codenames   map[string]string
	lastUpdate  time.Time
	ready       bool
}

type snapshot struct {
	Feed       models.ReleasesFeed `json:"feed"`
	Codenames  map[string]string   `json:"codenames"`
	LastUpdate time.Time           `json:"last_update"`
}

func NewCache(githubToken, blogDir, persistPath string) *Cache {
	return &Cache{
		client:      github.NewClient(githubToken),
		blogDir:     blogDir,
		persistPath: persistPath,
		codenames:   map[string]string{},
	}
}

func (c *Cache) Initialize(ctx context.Context) error {
	if c.persistPath != "" {
		if err := c.loadFromDisk(); err == nil {
			log.Info("Releases cache loaded from disk; refreshing in background")
		} else if !errors.Is(err, fs.ErrNotExist) {
			log.Warn("Failed to load releases cache from disk", "err", err)
		}
	}
	return c.Refresh(ctx)
}

func (c *Cache) Refresh(ctx context.Context) error {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()
	log.Info("Refreshing releases cache...")

	ghReleases, err := c.client.ListReleases(ctx, repoOwner, repoName, fetchCount)
	if err != nil {
		return err
	}
	head, commitCount, err := c.client.GetRefHead(ctx, repoOwner, repoName, masterBranch)
	if err != nil {
		return err
	}
	posts, err := scanBlog(c.blogDir)
	if err != nil {
		log.Warn("Failed to scan blog posts", "dir", c.blogDir, "err", err)
	}

	c.mu.RLock()
	codenames := maps.Clone(c.codenames)
	c.mu.RUnlock()

	for _, r := range ghReleases {
		if _, ok := codenames[r.TagName]; ok || r.Draft {
			continue
		}
		name, err := c.fetchCodename(ctx, r.TagName)
		if err != nil {
			log.Warn("Failed to fetch codename", "tag", r.TagName, "err", err)
			continue
		}
		codenames[r.TagName] = name
	}

	feed := buildFeed(ghReleases, codenames, posts)
	sha := head.SHA
	if len(sha) > 8 {
		sha = sha[:8]
	}
	feed.Master = &models.MasterInfo{
		SHA:         sha,
		CommitCount: commitCount,
		Date:        head.Commit.Committer.Date,
	}

	c.mu.Lock()
	c.setFeed(feed)
	c.codenames = codenames
	c.lastUpdate = time.Now()
	c.ready = true
	c.mu.Unlock()

	if err := c.saveToDisk(); err != nil {
		log.Warn("Failed to persist releases cache", "err", err)
	}

	log.Infof("Releases cache refreshed with %d releases", len(feed.Releases))
	return nil
}

// fetchCodename caches a missing CODENAME as "" since the file is immutable per tag.
func (c *Cache) fetchCodename(ctx context.Context, tag string) (string, error) {
	data, err := c.client.GetFileAtRef(ctx, repoOwner, repoName, codenamePath, tag)
	if errors.Is(err, github.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func buildFeed(ghReleases []github.Release, codenames map[string]string, posts map[string]blogPost) models.ReleasesFeed {
	all := make([]models.Release, 0, len(ghReleases))
	for _, r := range ghReleases {
		if r.Draft {
			continue
		}
		version := strings.TrimPrefix(r.TagName, "v")
		counts, highlights := parseNotes(r.Body)
		rel := models.Release{
			Tag:         r.TagName,
			Version:     version,
			Codename:    codenames[r.TagName],
			Prerelease:  r.Prerelease,
			PublishedAt: r.PublishedAt,
			URL:         r.HTMLURL,
			Counts:      counts,
			Highlights:  highlights,
		}
		if post, ok := blogFor(posts, version); ok {
			rel.BlogURL = blogBaseURL + post.Slug
			rel.Summary = post.Description
		}
		all = append(all, rel)
	}

	// GitHub lists by creation date, so a backported point release would otherwise outrank a newer minor.
	sort.SliceStable(all, func(i, j int) bool {
		return compareVersions(all[i].Version, all[j].Version) > 0
	})

	feed := models.ReleasesFeed{Releases: all[:min(len(all), keepCount)]}
	for i := range all {
		if !all[i].Prerelease {
			latest := all[i]
			feed.Latest = &latest
			break
		}
	}
	return feed
}

// compareVersions orders SemVer strings: build metadata is ignored, a release outranks its
// prereleases, and numeric prerelease identifiers compare as numbers (beta.10 > beta.2).
func compareVersions(a, b string) int {
	a, _, _ = strings.Cut(a, "+")
	b, _, _ = strings.Cut(b, "+")
	aCore, aPre, _ := strings.Cut(a, "-")
	bCore, bPre, _ := strings.Cut(b, "-")
	if d := compareIdentifiers(strings.Split(aCore, "."), strings.Split(bCore, "."), true); d != 0 {
		return d
	}
	switch {
	case aPre == bPre:
		return 0
	case aPre == "":
		return 1
	case bPre == "":
		return -1
	}
	return compareIdentifiers(strings.Split(aPre, "."), strings.Split(bPre, "."), false)
}

// core: a missing identifier counts as 0 (1.7 == 1.7.0); otherwise fewer identifiers sort first.
func compareIdentifiers(a, b []string, core bool) int {
	for i := 0; i < max(len(a), len(b)); i++ {
		if i >= len(a) {
			if core {
				return compareIdentifier("0", b[i])
			}
			return -1
		}
		if i >= len(b) {
			if core {
				return compareIdentifier(a[i], "0")
			}
			return 1
		}
		if d := compareIdentifier(a[i], b[i]); d != 0 {
			return d
		}
	}
	return 0
}

func compareIdentifier(a, b string) int {
	x, aErr := strconv.Atoi(a)
	y, bErr := strconv.Atoi(b)
	switch {
	case aErr == nil && bErr == nil:
		return x - y
	case aErr == nil:
		return -1
	case bErr == nil:
		return 1
	}
	return strings.Compare(a, b)
}

// setFeed must be called with c.mu held.
func (c *Cache) setFeed(feed models.ReleasesFeed) {
	c.feed = feed
	c.feedETag = etagOf(feed)
	c.latestETag = ""
	if feed.Latest != nil {
		c.latestETag = etagOf(feed.Latest)
	}
}

func etagOf(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

func (c *Cache) loadFromDisk() error {
	data, err := os.ReadFile(c.persistPath)
	if err != nil {
		return err
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	if snap.Codenames == nil {
		snap.Codenames = map[string]string{}
	}
	c.mu.Lock()
	c.setFeed(snap.Feed)
	c.codenames = snap.Codenames
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
	snap := snapshot{
		Feed:       c.feed,
		Codenames:  c.codenames,
		LastUpdate: c.lastUpdate,
	}
	data, err := json.Marshal(snap)
	c.mu.RUnlock()
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

func (c *Cache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

// The feed is replaced wholesale on refresh and never mutated, so callers may share it.
func (c *Cache) GetFeed() (models.ReleasesFeed, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feed, c.feedETag
}

func (c *Cache) GetLatest() (*models.Release, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feed.Latest, c.latestETag
}
