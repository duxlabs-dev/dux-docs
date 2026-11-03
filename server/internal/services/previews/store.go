package previews

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type manifestEntry struct {
	SourceKind  string    `json:"sourceKind"`
	SourceKey   string    `json:"sourceKey"`
	File        string    `json:"file"`
	ETag        string    `json:"etag"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type Store struct {
	dir      string
	mu       sync.Mutex
	manifest map[string]manifestEntry
}

func NewStore(cacheDir string) (*Store, error) {
	return newStoreAt(filepath.Join(cacheDir, "previews"))
}

func NewThumbStore(cacheDir string) (*Store, error) {
	return newStoreAt(filepath.Join(cacheDir, "previews", "thumbs"))
}

func newStoreAt(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create previews directory: %w", err)
	}

	s := &Store{dir: dir, manifest: map[string]manifestEntry{}}
	if err := s.loadManifest(); err != nil {
		return nil, err
	}
	return s, nil
}

const composeVersion = "v3"

func SourceKey(sourceURL string, p models.Plugin) string {
	statuses := slices.Clone(p.Status)
	slices.Sort(statuses)
	parts := []string{composeVersion, sourceURL, p.Name, p.Category, p.Description, p.Author, p.Version, strings.Join(statuses, ",")}
	h := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h[:])
}

func (s *Store) loadManifest() error {
	data, err := os.ReadFile(s.manifestPath())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read preview manifest: %w", err)
	}
	if err := json.Unmarshal(data, &s.manifest); err != nil {
		return fmt.Errorf("failed to parse preview manifest: %w", err)
	}
	return nil
}

func (s *Store) manifestPath() string {
	return filepath.Join(s.dir, "manifest.json")
}

func (s *Store) PlaceholderPath() string {
	return filepath.Join(s.dir, "placeholder.png")
}

func (s *Store) NeedsUpdate(id, sourceKey string) bool {
	s.mu.Lock()
	entry, ok := s.manifest[id]
	s.mu.Unlock()

	if !ok {
		return true
	}
	if entry.SourceKey != sourceKey {
		return true
	}
	if _, err := os.Stat(filepath.Join(s.dir, entry.File)); err != nil {
		return true
	}
	return false
}

func (s *Store) Put(id, sourceKind, sourceKey, ext string, data []byte) error {
	file := id + "." + ext
	if err := atomicWrite(filepath.Join(s.dir, file), data); err != nil {
		return fmt.Errorf("failed to write preview %s: %w", file, err)
	}

	etag := sourceKey
	if len(etag) > 16 {
		etag = etag[:16]
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if prev, ok := s.manifest[id]; ok && prev.File != file {
		_ = os.Remove(filepath.Join(s.dir, prev.File))
	}
	s.manifest[id] = manifestEntry{
		SourceKind:  sourceKind,
		SourceKey:   sourceKey,
		File:        file,
		ETag:        etag,
		GeneratedAt: time.Now().UTC(),
	}
	return s.saveManifestLocked()
}

func (s *Store) Lookup(id string) (string, string, bool) {
	s.mu.Lock()
	entry, ok := s.manifest[id]
	s.mu.Unlock()

	if !ok {
		return "", "", false
	}
	path := filepath.Join(s.dir, entry.File)
	if _, err := os.Stat(path); err != nil {
		return "", "", false
	}
	return path, entry.ETag, true
}

func (s *Store) EnsurePlaceholder(render func() ([]byte, error)) error {
	path := s.PlaceholderPath()
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	data, err := render()
	if err != nil {
		return err
	}
	if err := atomicWrite(path, data); err != nil {
		return fmt.Errorf("failed to write placeholder: %w", err)
	}
	return nil
}

func (s *Store) saveManifestLocked() error {
	data, err := json.Marshal(s.manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal preview manifest: %w", err)
	}
	if err := atomicWrite(s.manifestPath(), data); err != nil {
		return fmt.Errorf("failed to write preview manifest: %w", err)
	}
	return nil
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
