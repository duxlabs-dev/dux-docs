package contributors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/poeditor"
)

type fakeGitHub struct {
	repos map[string][]github.Contributor
	err   error
}

func (f *fakeGitHub) ListContributors(_ context.Context, owner, repo string) ([]github.Contributor, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.repos[owner+"/"+repo], nil
}

type fakePOEditor struct {
	people map[string][]poeditor.Contributor
	langs  map[string][]poeditor.Language
	err    error
}

func (f *fakePOEditor) Contributors(_ context.Context, id string) ([]poeditor.Contributor, error) {
	return f.people[id], f.err
}

func (f *fakePOEditor) Languages(_ context.Context, id string) ([]poeditor.Language, error) {
	return f.langs[id], f.err
}

func perm(project, name string, proofreader bool, langs ...string) poeditor.Permission {
	var p poeditor.Permission
	p.Project.ID = json.Number(project)
	p.Project.Name = name
	p.Type = "contributor"
	p.Proofreader = proofreader
	p.Languages = langs
	return p
}

func TestFetchCodeMergesAndFilters(t *testing.T) {
	gh := &fakeGitHub{repos: map[string][]github.Contributor{
		"o/a": {
			{Login: "Ada", Contributions: 10, Type: "User"},
			{Login: "dependabot[bot]", Contributions: 50, Type: "Bot"},
			{Login: "renovate", Contributions: 40, Type: "Bot"},
			{Login: "owner", Contributions: 99, Type: "User"},
		},
		"o/b": {
			{Login: "ada", Contributions: 5, Type: "User"},
			{Login: "Bob", Contributions: 15, Type: "User"},
		},
	}}
	got, err := fetchCode(context.Background(), gh, []string{"o/a", "o/b"}, map[string]bool{"owner": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d contributors: %+v", len(got), got)
	}
	// Ties break by login so the order is stable between refreshes.
	if got[0].Login != "Ada" || got[0].Contributions != 15 || strings.Join(got[0].Repos, ",") != "a,b" {
		t.Errorf("merged = %+v", got[0])
	}
	if got[1].Login != "Bob" {
		t.Errorf("second = %+v", got[1])
	}

	if _, err := fetchCode(context.Background(), gh, []string{"no-slash"}, nil); err == nil {
		t.Error("malformed repo accepted")
	}
}

func TestFetchLanguagesCountsPeopleWithoutNames(t *testing.T) {
	po := &fakePOEditor{
		langs: map[string][]poeditor.Language{
			"1": {{Code: "ru", Name: "Russian"}, {Code: "de", Name: "German"}},
			"2": {{Code: "ru", Name: "Russian"}},
		},
		people: map[string][]poeditor.Contributor{
			"1": {
				{Name: "Ivan Real-Name", Email: "ivan@x.org", Permissions: []poeditor.Permission{perm("1", "DuxShell", false, "ru", "de")}},
				{Name: "Olga", Email: "olga@x.org", Permissions: []poeditor.Permission{perm("1", "DuxShell", false, "ru")}},
				{Name: "Admin", Email: "admin@x.org", Permissions: []poeditor.Permission{{Type: "administrator"}}},
				// Permissions for other projects in the org must not leak in.
				{Name: "Other", Email: "other@x.org", Permissions: []poeditor.Permission{perm("9", "Elsewhere", false, "fr")}},
			},
			"2": {
				// The same person in a second project counts once.
				{Name: "Ivan Real-Name", Email: "IVAN@x.org", Permissions: []poeditor.Permission{perm("2", "DuxCalendar", true, "ru", "xx")}},
			},
		},
	}
	got, people, err := fetchLanguages(context.Background(), po, []string{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if people != 2 {
		t.Errorf("people = %d, want 2", people)
	}
	var summary []string
	for _, l := range got {
		summary = append(summary, fmt.Sprintf("%s=%s:%d", l.Code, l.Name, l.Translators))
	}
	// Most translators first; an unknown code falls back to itself as the name.
	if strings.Join(summary, ",") != "ru=Russian:2,de=German:1,xx=xx:1" {
		t.Errorf("languages = %v", summary)
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "@") || strings.Contains(string(raw), "Ivan") {
		t.Errorf("feed leaks a name or email: %s", raw)
	}
}

func TestCacheKeepsLastGoodSourceAndPersists(t *testing.T) {
	dir := t.TempDir()
	gh := &fakeGitHub{repos: map[string][]github.Contributor{"o/a": {{Login: "ada", Contributions: 3}}}}
	po := &fakePOEditor{
		langs:  map[string][]poeditor.Language{"1": {{Code: "ru", Name: "Russian"}}},
		people: map[string][]poeditor.Contributor{"1": {{Name: "Ivan", Email: "i@x", Permissions: []poeditor.Permission{perm("1", "DuxShell", false, "ru")}}}},
	}
	cfg := Config{GitHub: gh, Repos: []string{"o/a"}, POEditor: po, ProjectIDs: []string{"1"}, DataDir: dir}

	c := NewCache(cfg)
	if c.IsReady() {
		t.Fatal("ready before any fetch")
	}

	// GitHub alone is not enough: the feed waits for every configured source.
	po.err = errors.New("poeditor down")
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("expected POEditor error")
	}
	if c.IsReady() {
		t.Fatal("ready without translators")
	}

	po.err = nil
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	feed := c.GetFeed()
	if !c.IsReady() || len(feed.Code) != 1 || len(feed.Languages) != 1 || feed.TranslatorCount != 1 || feed.UpdatedAt.IsZero() {
		t.Fatalf("feed = %+v", feed)
	}

	gh.err = errors.New("github down")
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatal("expected GitHub error")
	}
	if got := c.GetFeed(); len(got.Code) != 1 {
		t.Errorf("failed GitHub refresh dropped code contributors: %+v", got)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "contributors.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "i@x") || strings.Contains(string(raw), "Ivan") {
		t.Errorf("snapshot stores a name or email: %s", raw)
	}

	// A restart serves the snapshot before the first fetch completes.
	restarted := NewCache(cfg)
	if !restarted.IsReady() || restarted.GetFeed().TranslatorCount != 1 {
		t.Errorf("restart did not load snapshot: %+v", restarted.GetFeed())
	}
}

func TestCacheWithoutPOEditorIsReadyFromGitHubAlone(t *testing.T) {
	gh := &fakeGitHub{repos: map[string][]github.Contributor{"o/a": {{Login: "ada", Contributions: 1}}}}
	c := NewCache(Config{GitHub: gh, Repos: []string{"o/a"}, ProjectIDs: []string{"1"}})
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !c.IsReady() || c.GetFeed().Languages == nil {
		t.Errorf("feed = %+v", c.GetFeed())
	}
}

func TestPOEditorClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.PostForm.Get("api_token") != "tok" || r.PostForm.Get("id") != "805178" {
			t.Errorf("bad form: %v", r.PostForm)
		}
		switch r.URL.Path {
		case "/contributors/list":
			// The project id arrives as a string here, unlike everywhere else.
			w.Write([]byte(`{"response":{"status":"success","code":"200","message":"OK"},"result":{"contributors":[
				{"name":"Translator","email":"t@example.com","permissions":[{"project":{"id":"805178","name":"DuxShell"},"type":"contributor","proofreader":false,"languages":["ru"]}]}]}}`))
		case "/languages/list":
			w.Write([]byte(`{"response":{"status":"fail","code":"4011","message":"Invalid API key"}}`))
		}
	}))
	defer srv.Close()

	client := poeditor.NewClientWithBaseURL(srv.URL, "tok")
	people, err := client.Contributors(context.Background(), "805178")
	if err != nil {
		t.Fatal(err)
	}
	if len(people) != 1 || people[0].Permissions[0].Project.ID.String() != "805178" {
		t.Errorf("people = %+v", people)
	}
	if _, err := client.Languages(context.Background(), "805178"); err == nil || !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("err = %v", err)
	}
}
