package sponsors

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var now = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func donation(name, tx string, public bool, at time.Time) KofiPayload {
	return KofiPayload{Type: "Tip", FromName: name, KofiTransactionID: tx, IsPublic: public, Timestamp: at.Format(time.RFC3339)}
}

func TestKofiApply(t *testing.T) {
	s := newKofiState()

	if !s.apply(donation("  Ada   Lovelace ", "tx1", true, now.Add(-48*time.Hour)), now) {
		t.Fatal("first donation not applied")
	}
	if s.apply(donation("Ada Lovelace", "tx1", true, now), now) {
		t.Error("retried transaction applied twice")
	}
	if got := s.Supporters["ada lovelace"]; got == nil || got.Name != "Ada Lovelace" {
		t.Fatalf("supporter = %+v", got)
	}

	s.apply(donation("ada lovelace", "tx2", true, now), now)
	if got := s.Supporters["ada lovelace"]; !got.LastPayment.Equal(now) || !got.FirstSeen.Equal(now.Add(-48*time.Hour)) {
		t.Errorf("dates = first %v last %v", got.FirstSeen, got.LastPayment)
	}

	s.apply(donation("Hidden", "tx3", false, now), now)
	if _, ok := s.Supporters["hidden"]; ok {
		t.Error("private supporter stored")
	}

	s.apply(donation("Ada Lovelace", "tx4", false, now), now)
	if _, ok := s.Supporters["ada lovelace"]; ok {
		t.Error("supporter who went private still listed")
	}

	// A delayed retry of an older public payment must not undo going private.
	s.apply(donation("Ada Lovelace", "tx-old", true, now.Add(-time.Hour)), now)
	if _, ok := s.Supporters["ada lovelace"]; ok {
		t.Error("stale public retry re-listed a private supporter")
	}
	later := now.Add(time.Hour)
	s.apply(donation("Ada Lovelace", "tx-new", true, later), later)
	if _, ok := s.Supporters["ada lovelace"]; !ok {
		t.Error("newer public payment did not re-list supporter")
	}

	if s.apply(KofiPayload{Type: "Shop Order", FromName: "Buyer", KofiTransactionID: "tx5", IsPublic: true}, now) {
		t.Error("shop order counted as sponsorship")
	}
}

func TestBuildFeed(t *testing.T) {
	gh := []githubSponsor{
		{Login: "active", Name: "Active One", Active: true},
		{Login: "lapsed", Active: false},
		{Login: "both", Name: "Both Places", Active: false},
	}
	kofi := newKofiState()
	kofi.apply(donation("Recent", "a", true, now.Add(-24*time.Hour)), now)
	kofi.apply(donation("Old", "b", true, now.Add(-90*24*time.Hour)), now)
	kofi.apply(donation("both places", "c", true, now.Add(-24*time.Hour)), now)

	feed := buildFeed(gh, kofi, now)

	got := func(which string) []string {
		var out []string
		list := feed.Current
		if which == "past" {
			list = feed.Past
		}
		for _, s := range list {
			out = append(out, s.Source+":"+s.Name)
		}
		return out
	}
	// Never merged across platforms: names are user-controlled on both.
	assertList(t, "current", got("current"), []string{"github:Active One", "kofi:Recent", "kofi:both places"})
	assertList(t, "past", got("past"), []string{"github:lapsed", "github:Both Places", "kofi:Old"})

	empty := buildFeed(nil, newKofiState(), now)
	data, _ := json.Marshal(empty)
	if string(data) != `{"current":[],"past":[],"updatedAt":"2026-11-02T12:00:00Z"}` {
		t.Errorf("empty feed = %s", data)
	}
}

type fakeGQL struct {
	pages map[string][]sponsorshipsPage // key: owner + active flag
	calls int
}

func (f *fakeGQL) GraphQL(_ context.Context, _ string, vars map[string]any, out any) error {
	f.calls++
	key := vars["login"].(string)
	if vars["active"].(bool) {
		key += ":active"
	}
	idx := 0
	if after, ok := vars["after"].(string); ok {
		idx = int(after[0] - '0')
	}
	pages := f.pages[key]
	if idx >= len(pages) {
		return json.Unmarshal([]byte(`{"repositoryOwner":null}`), out)
	}
	*(out.(*sponsorshipsPage)) = pages[idx]
	return nil
}

func page(next string, logins ...string) sponsorshipsPage {
	raw := map[string]any{}
	var nodes []map[string]any
	for _, l := range logins {
		if l == "" {
			nodes = append(nodes, map[string]any{"sponsorEntity": nil})
			continue
		}
		nodes = append(nodes, map[string]any{"sponsorEntity": map[string]any{"login": l}})
	}
	raw["repositoryOwner"] = map[string]any{"sponsorshipsAsMaintainer": map[string]any{
		"pageInfo": map[string]any{"hasNextPage": next != "", "endCursor": next},
		"nodes":    nodes,
	}}
	data, _ := json.Marshal(raw)
	var p sponsorshipsPage
	_ = json.Unmarshal(data, &p)
	return p
}

func TestFetchGithubSponsors(t *testing.T) {
	f := &fakeGQL{pages: map[string][]sponsorshipsPage{
		"Org":         {page("1", "a", "", "b"), page("", "c")},
		"Org:active":  {page("", "b")},
		"Next":        {page("", "B", "d")},
		"Next:active": {page("", "d")},
	}}
	got, err := fetchGithubSponsors(context.Background(), f, []string{"Org", "Next", "Missing"})
	if err != nil {
		t.Fatal(err)
	}
	var flat []string
	for _, s := range got {
		state := "past"
		if s.Active {
			state = "active"
		}
		flat = append(flat, s.Login+"="+state)
	}
	assertList(t, "sponsors", flat, []string{"a=past", "b=active", "c=past", "d=active"})
}

func TestRecordKofiPersists(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(nil, nil, dir)
	c.now = func() time.Time { return now }
	if err := c.RecordKofi(donation("Persisted", "tx", true, now)); err != nil {
		t.Fatal(err)
	}
	if feed := c.GetFeed(); len(feed.Current) != 1 || feed.Current[0].Name != "Persisted" {
		t.Fatalf("feed = %+v", feed)
	}

	reloaded := NewCache(nil, nil, dir)
	reloaded.now = func() time.Time { return now }
	if !reloaded.IsReady() {
		t.Error("cache without a token should be ready")
	}
	if err := reloaded.RecordKofi(donation("Persisted", "tx", true, now)); err != nil {
		t.Fatal(err)
	}
	if got := reloaded.kofi.Supporters["persisted"]; got == nil {
		t.Fatalf("supporter not reloaded from %s", filepath.Join(dir, "sponsors-kofi.json"))
	}
	if len(reloaded.kofi.Seen) != 1 {
		t.Errorf("seen = %v, want the retried tx deduplicated after reload", reloaded.kofi.Seen)
	}
}

func assertList(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}

func TestRecordKofiRefusesUnreadableStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sponsors-kofi.json")
	if err := os.WriteFile(path, []byte(`{"supporters":{"a":`), 0o600); err != nil {
		t.Fatal(err)
	}
	c := NewCache(nil, nil, dir)
	if err := c.RecordKofi(donation("New", "tx", true, now)); err == nil {
		t.Fatal("wrote over an unreadable Ko-fi store")
	}
	if data, _ := os.ReadFile(path); string(data) != `{"supporters":{"a":` {
		t.Errorf("store modified: %s", data)
	}
}

func TestRecordKofiRetryPersistsAfterFailedWrite(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(nil, nil, dir)
	c.now = func() time.Time { return now }
	// Make the target path a directory so the rename fails.
	if err := os.Mkdir(c.kofiPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := c.RecordKofi(donation("Retried", "tx", true, now)); err == nil {
		t.Fatal("expected write failure")
	}
	if err := os.Remove(c.kofiPath); err != nil {
		t.Fatal(err)
	}
	if err := c.RecordKofi(donation("Retried", "tx", true, now)); err != nil {
		t.Fatal(err)
	}
	reloaded := NewCache(nil, nil, dir)
	if _, ok := reloaded.kofi.Supporters["retried"]; !ok {
		t.Error("retry acknowledged without persisting")
	}
}
