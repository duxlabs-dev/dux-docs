package registry

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

func TestFetchPluginsCarriesForwardOnTransientFailure(t *testing.T) {
	gh := newFakeGitHub(t)

	gh.registry("fresh", "flaky", "gone")
	gh.plugin("fresh", "2.0.0")
	gh.status("/repos/dev/flaky/contents/plugin.json", http.StatusBadGateway, nil)
	gh.status("/repos/dev/gone/contents/plugin.json", http.StatusNotFound, nil)

	prev := []models.Plugin{
		{ID: "fresh", Version: "1.0.0"},
		{ID: "flaky", Version: "1.0.0"},
		{ID: "gone", Version: "1.0.0"},
		{ID: "removed", Version: "1.0.0"},
	}

	plugins, err := gh.parser().FetchPlugins(context.Background(), prev)
	if err != nil {
		t.Fatalf("FetchPlugins: %v", err)
	}

	var got []string
	for _, p := range plugins {
		got = append(got, p.ID+"@"+p.Version)
	}
	want := []string{"fresh@2.0.0", "flaky@1.0.0"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFetchThemesStopsCallingGitHubWhenRateLimited(t *testing.T) {
	gh := newFakeGitHub(t)

	gh.json(registryAPI+"/contents/themes", []github.RepoContent{
		{Name: "fresh", Type: "dir"},
		{Name: "limited", Type: "dir"},
		{Name: "untouched", Type: "dir"},
		{Name: "new", Type: "dir"},
	})
	gh.json(registryAPI+"/contents/themes/fresh/theme.json", gh.file("/themes/fresh/theme.json", map[string]string{"id": "freshTheme", "version": "2"}))
	gh.commit(registryAPI + "/commits?path=themes/fresh")
	gh.status(registryAPI+"/contents/themes/limited/theme.json", http.StatusForbidden, rateLimitHeader())

	prev := []models.Theme{
		{ID: "freshTheme", Version: "1", PreviewURL: themePreviewURL("fresh")},
		{ID: "limitedTheme", Version: "1", PreviewURL: themePreviewURL("limited")},
		{ID: "untouchedTheme", Version: "1", PreviewURL: themePreviewURL("untouched")},
	}

	themes, err := gh.parser().FetchThemes(context.Background(), prev)
	if err != nil {
		t.Fatalf("FetchThemes: %v", err)
	}

	var got []string
	for _, theme := range themes {
		got = append(got, theme.ID+"@"+theme.Version)
	}
	want := []string{"freshTheme@2", "limitedTheme@1", "untouchedTheme@1"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for _, dir := range []string{"untouched", "new"} {
		if gh.wasRequested(registryAPI + "/contents/themes/" + dir) {
			t.Fatalf("theme %s was requested after the rate limit", dir)
		}
	}
}

func TestThemeRefreshKeepsCacheWhenListingFails(t *testing.T) {
	gh := newFakeGitHub(t)
	gh.status(registryAPI+"/contents/themes", http.StatusForbidden, rateLimitHeader())

	prev := []models.Theme{{ID: "kept", Version: "1", PreviewURL: themePreviewURL("kept")}}
	cache := NewThemeCache("", "")
	cache.parser = gh.parser()
	cache.themes = prev
	cache.ready = true

	if err := cache.Refresh(context.Background()); err == nil {
		t.Fatal("Refresh succeeded, want the listing error")
	}

	if !cache.IsReady() {
		t.Fatal("cache marked not ready after a failed refresh")
	}
	if got := cache.GetThemes(); len(got) != 1 || got[0].ID != "kept" {
		t.Fatalf("got %v, want the previous themes", got)
	}
}
