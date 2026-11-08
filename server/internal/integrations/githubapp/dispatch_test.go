package githubapp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v69/github"
)

func TestDispatchPostsRepositoryDispatch(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewToken("pat-123", "duxlabs-dev", "dux-repo")
	base, _ := url.Parse(srv.URL + "/")
	c.base = github.NewClient(nil)
	c.base.BaseURL = base

	payload := map[string]any{"release": "F44", "builds": []string{"qt6-qtbase-6.11.2-2.fc44"}}
	if err := c.Dispatch(context.Background(), "duxlabs-dev", "dux-repo", "fedora-qt-update", payload); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if gotPath != "POST /repos/duxlabs-dev/dux-repo/dispatches" {
		t.Fatalf("unexpected request: %s", gotPath)
	}
	if gotAuth != "Bearer pat-123" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotBody["event_type"] != "fedora-qt-update" {
		t.Fatalf("unexpected event_type: %v", gotBody["event_type"])
	}
	cp, ok := gotBody["client_payload"].(map[string]any)
	if !ok || cp["release"] != "F44" {
		t.Fatalf("unexpected client_payload: %v", gotBody["client_payload"])
	}
}
