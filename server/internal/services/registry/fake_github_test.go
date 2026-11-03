package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const registryAPI = "/repos/duxlabs-dev/dux-plugin-registry"

type fakeGitHub struct {
	*httptest.Server
	t         *testing.T
	mu        sync.Mutex
	routes    map[string]http.HandlerFunc
	requested []string
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	f := &fakeGitHub{t: t, routes: map[string]http.HandlerFunc{}}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		if p := r.URL.Query().Get("path"); p != "" {
			key += "?path=" + p
		}

		f.mu.Lock()
		f.requested = append(f.requested, key)
		route, ok := f.routes[key]
		f.mu.Unlock()

		if !ok {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		route(w, r)
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeGitHub) json(path string, v any) {
	f.routes[path] = func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			f.t.Errorf("encode %s: %v", path, err)
		}
	}
}

func (f *fakeGitHub) status(path string, status int, header map[string]string) {
	f.routes[path] = func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
	}
}

func (f *fakeGitHub) file(path string, v any) []github.RepoContent {
	f.json("/raw"+path, v)
	return []github.RepoContent{{Type: "file", DownloadURL: f.URL + "/raw" + path}}
}

func (f *fakeGitHub) commit(path string) {
	f.json(path, []map[string]any{{"sha": "abc", "commit": map[string]any{"committer": map[string]any{"date": "2026-10-01T00:00:00Z"}}}})
}

func (f *fakeGitHub) wasRequested(prefix string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.ContainsFunc(f.requested, func(r string) bool { return strings.HasPrefix(r, prefix) })
}

func (f *fakeGitHub) parser() *Parser {
	p := NewParser("")
	p.clients["github.com"] = github.NewClientWithBaseURL(f.URL, "")
	return p
}

func rateLimitHeader() map[string]string {
	return map[string]string{
		"X-RateLimit-Remaining": "0",
		"X-RateLimit-Reset":     strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
	}
}

func (f *fakeGitHub) registry(ids ...string) {
	var listing []github.RepoContent
	for _, id := range ids {
		entry := f.file("/registry/"+id+".json", models.RegistryPlugin{ID: id, Repo: "https://github.com/dev/" + id})
		entry[0].Name = id + ".json"
		listing = append(listing, entry...)
	}
	f.json(registryAPI+"/contents/plugins", listing)
}

func (f *fakeGitHub) plugin(id, version string) {
	f.json("/repos/dev/"+id+"/contents/plugin.json", f.file("/dev/"+id+"/plugin.json", map[string]string{"version": version}))
	f.commit("/repos/dev/" + id + "/commits")
}

func (f *fakeGitHub) issues(issues ...map[string]any) {
	f.routes[registryAPI+"/issues"] = func(w http.ResponseWriter, r *http.Request) {
		page := issues
		if r.URL.Query().Get("page") != "1" {
			page = []map[string]any{}
		}
		if err := json.NewEncoder(w).Encode(page); err != nil {
			f.t.Errorf("encode issues: %v", err)
		}
	}
}
