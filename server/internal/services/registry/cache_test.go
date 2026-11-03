package registry

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

func pluginIssue(id string, upvotes int, labels ...string) map[string]any {
	names := []map[string]string{{"name": "plugin"}}
	for _, label := range labels {
		names = append(names, map[string]string{"name": label})
	}
	return map[string]any{
		"body":      "<!-- dms-plugin-id: " + id + " -->",
		"labels":    names,
		"reactions": map[string]int{"+1": upvotes},
	}
}

func TestRefreshTakesFreshFeedbackAndKeepsRecentModeration(t *testing.T) {
	gh := newFakeGitHub(t)
	gh.registry("alpha", "beta")
	gh.plugin("alpha", "1.0.0")
	gh.plugin("beta", "1.0.0")
	gh.issues(
		pluginIssue("alpha", 5),
		pluginIssue("beta", 7, "status:unmaintained"),
	)

	cache := NewCache("", "")
	cache.parser = gh.parser()
	cache.plugins = []models.Plugin{
		{ID: "alpha", Version: "1.0.0", Upvotes: 1},
		{ID: "beta", Version: "1.0.0", Upvotes: 1, Status: []string{"broken"}},
	}
	cache.moderatedAt["beta"] = time.Now().Add(-2 * moderationGrace)
	cache.ApplyStatus("alpha", "reviewed", true)

	if err := cache.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	alpha, _ := cache.PluginByID("alpha")
	beta, _ := cache.PluginByID("beta")
	if alpha.Upvotes != 5 || beta.Upvotes != 7 {
		t.Fatalf("upvotes = %d, %d, want the fresh 5, 7", alpha.Upvotes, beta.Upvotes)
	}
	if !slices.Equal(alpha.Status, []string{"reviewed"}) {
		t.Fatalf("alpha status = %v, want the just-applied [reviewed] to survive a stale read", alpha.Status)
	}
	if !slices.Equal(beta.Status, []string{"unmaintained"}) {
		t.Fatalf("beta status = %v, want the fresh [unmaintained] once moderation has settled", beta.Status)
	}
}
