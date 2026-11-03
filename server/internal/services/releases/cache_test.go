package releases

import (
	"testing"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
)

func TestBuildFeedOrdering(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 9, d, 0, 0, 0, 0, time.UTC) }
	// GitHub order: creation date, newest first.
	gh := []github.Release{
		{TagName: "v1.8.0-beta.1", Prerelease: true, PublishedAt: day(20)},
		{TagName: "v1.6.4", PublishedAt: day(19)},
		{TagName: "v1.7.0-rc", Draft: true, PublishedAt: day(18)},
		{TagName: "v1.7.1", PublishedAt: day(17)},
		{TagName: "v1.7.0", PublishedAt: day(10)},
		{TagName: "v1.6.3", PublishedAt: day(9)},
		{TagName: "v1.6.2", PublishedAt: day(8)},
		{TagName: "v1.6.10", PublishedAt: day(7)},
	}

	feed := buildFeed(gh, map[string]string{"v1.7.1": "Heisenberg"}, map[string]blogPost{
		"1.7": {Slug: "v1-7-release", Description: "Seven"},
	})

	want := []string{"v1.8.0-beta.1", "v1.7.1", "v1.7.0", "v1.6.10", "v1.6.4", "v1.6.3"}
	if len(feed.Releases) != len(want) {
		t.Fatalf("got %d releases, want %d", len(feed.Releases), len(want))
	}
	for i, tag := range want {
		if feed.Releases[i].Tag != tag {
			t.Errorf("releases[%d] = %s, want %s", i, feed.Releases[i].Tag, tag)
		}
	}

	latest := feed.Latest
	if latest == nil || latest.Tag != "v1.7.1" || latest.Version != "1.7.1" || latest.Codename != "Heisenberg" {
		t.Fatalf("latest = %+v, want v1.7.1 Heisenberg", latest)
	}
	if latest.BlogURL != "https://danklinux.com/blog/v1-7-release" || latest.Summary != "Seven" {
		t.Errorf("latest blog = %q %q", latest.BlogURL, latest.Summary)
	}
	if !feed.Releases[0].Prerelease {
		t.Error("beta not flagged as prerelease")
	}
}

func TestCompareVersionsSemver(t *testing.T) {
	for _, tt := range []struct{ a, b string }{
		{"1.7.0-beta.10", "1.7.0-beta.2"},
		{"1.8.2+build.5", "1.8.1"},
		{"1.7.0", "1.7.0-rc.1"},
		{"1.7.0-rc.1", "1.7.0-beta.9"},
		{"1.7.0-beta.2", "1.7.0-beta"},
		{"1.7.0-beta", "1.7.0-1"},
	} {
		if compareVersions(tt.a, tt.b) <= 0 {
			t.Errorf("compareVersions(%q, %q) <= 0, want >", tt.a, tt.b)
		}
		if compareVersions(tt.b, tt.a) >= 0 {
			t.Errorf("compareVersions(%q, %q) >= 0, want <", tt.b, tt.a)
		}
	}
	if compareVersions("1.7", "1.7.0") != 0 || compareVersions("1.7.0+a", "1.7.0+b") != 0 {
		t.Error("equal versions did not compare equal")
	}
}
