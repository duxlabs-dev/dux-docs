package previews

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

func TestStoreReuseAndRegenerate(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	p := models.Plugin{ID: "foo", Name: "Foo", Category: "bar", Description: "desc", Author: "me"}
	key := SourceKey("https://example.com/a.png", p)

	if !s.NeedsUpdate("foo", key) {
		t.Fatal("expected NeedsUpdate before first Put")
	}

	if err := s.Put("foo", "screenshot", key, "jpg", []byte("data")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if s.NeedsUpdate("foo", key) {
		t.Fatal("expected no regeneration for unchanged sourceKey")
	}

	changed := SourceKey("https://example.com/b.png", p)
	if !s.NeedsUpdate("foo", changed) {
		t.Fatal("expected regeneration for changed sourceKey")
	}
}

func TestSourceKeyCoversVersionAndStatus(t *testing.T) {
	p := models.Plugin{ID: "foo", Name: "Foo", Version: "1.0.0"}
	base := SourceKey("u", p)

	p.Version = "1.1.0"
	if SourceKey("u", p) == base {
		t.Fatal("expected sourceKey to change with version")
	}

	p.Version = "1.0.0"
	p.Status = []string{"broken"}
	if SourceKey("u", p) == base {
		t.Fatal("expected sourceKey to change with status")
	}

	p.Status = []string{"reviewed", "broken"}
	reordered := models.Plugin{ID: "foo", Name: "Foo", Version: "1.0.0", Status: []string{"broken", "reviewed"}}
	if SourceKey("u", p) != SourceKey("u", reordered) {
		t.Fatal("expected sourceKey to be order-insensitive for statuses")
	}
}

func TestStoreManifestPersists(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	key := SourceKey("", models.Plugin{ID: "foo", Name: "Foo"})
	if err := s.Put("foo", "card", key, "png", []byte("data")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore reload: %v", err)
	}
	if reloaded.NeedsUpdate("foo", key) {
		t.Fatal("expected manifest to persist across reload")
	}

	path, etag, ok := reloaded.Lookup("foo")
	if !ok {
		t.Fatal("expected Lookup to find entry")
	}
	if filepath.Base(path) != "foo.png" {
		t.Fatalf("unexpected file %q", path)
	}
	if etag != key[:16] {
		t.Fatalf("etag %q, want %q", etag, key[:16])
	}
}

func TestStoreNeedsUpdateWhenFileMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	key := SourceKey("", models.Plugin{ID: "foo", Name: "Foo"})
	if err := s.Put("foo", "card", key, "png", []byte("data")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if err := os.Remove(filepath.Join(dir, "previews", "foo.png")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !s.NeedsUpdate("foo", key) {
		t.Fatal("expected regeneration when file missing")
	}
	if _, _, ok := s.Lookup("foo"); ok {
		t.Fatal("expected Lookup miss when file missing")
	}
}

func TestStoreExtensionChangeRemovesStaleFile(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	p := models.Plugin{ID: "foo", Name: "Foo"}
	if err := s.Put("foo", "screenshot", SourceKey("https://example.com/a.png", p), "jpg", []byte("a")); err != nil {
		t.Fatalf("Put jpg: %v", err)
	}
	if err := s.Put("foo", "card", SourceKey("", p), "png", []byte("b")); err != nil {
		t.Fatalf("Put png: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "previews", "foo.jpg")); !os.IsNotExist(err) {
		t.Fatal("expected stale jpg to be removed")
	}
	path, _, ok := s.Lookup("foo")
	if !ok || filepath.Base(path) != "foo.png" {
		t.Fatalf("expected foo.png, got %q ok=%v", path, ok)
	}
}
