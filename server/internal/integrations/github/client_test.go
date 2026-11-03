package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConditionalRequestReturnsCachedResponse(t *testing.T) {
	cases := []struct {
		name  string
		fetch func(c *Client, baseURL string) (string, error)
	}{
		{
			name: "api call",
			fetch: func(c *Client, _ string) (string, error) {
				commit, err := c.GetLastCommit(context.Background(), "o", "r", "")
				if err != nil {
					return "", err
				}
				return commit.SHA, nil
			},
		},
		{
			name: "raw download",
			fetch: func(c *Client, baseURL string) (string, error) {
				body, err := c.GetFileContents(context.Background(), baseURL+"/raw/plugin.json")
				return string(body), err
			},
		},
		{
			name: "cached headers survive",
			fetch: func(c *Client, _ string) (string, error) {
				head, count, err := c.GetRefHead(context.Background(), "o", "r", "master")
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%s/%d", head.SHA, count), nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var conditional []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				conditional = append(conditional, r.Header.Get("If-None-Match"))
				mu.Unlock()
				if r.Header.Get("If-None-Match") == `W/"v1"` {
					w.WriteHeader(http.StatusNotModified)
					return
				}
				w.Header().Set("ETag", `W/"v1"`)
				w.Header().Set("Link", `<https://api.github.com/repos/o/r/commits?per_page=1&page=42>; rel="last"`)
				fmt.Fprint(w, `[{"sha":"abc123"}]`)
			}))
			defer srv.Close()

			c := NewClientWithBaseURL(srv.URL, "token")
			first, err := tc.fetch(c, srv.URL)
			if err != nil {
				t.Fatalf("first fetch: %v", err)
			}
			second, err := tc.fetch(c, srv.URL)
			if err != nil {
				t.Fatalf("second fetch: %v", err)
			}

			if first != second {
				t.Fatalf("304 returned %q, want cached %q", second, first)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(conditional) != 2 || conditional[0] != "" || conditional[1] != `W/"v1"` {
				t.Fatalf("If-None-Match headers = %q, want none then the cached ETag", conditional)
			}
		})
	}
}

type cannedResponse struct {
	status int
	header map[string]string
	body   string
}

func TestRetryPolicy(t *testing.T) {
	resetAt := time.Now().Add(time.Hour).Truncate(time.Second)

	cases := []struct {
		name         string
		responses    []cannedResponse
		wantBody     string
		wantErr      error
		wantRequests int32
	}{
		{
			name:         "server errors are retried",
			responses:    []cannedResponse{{status: 502}, {status: 502}, {status: 200, body: "ok"}},
			wantBody:     "ok",
			wantRequests: 3,
		},
		{
			name:         "server errors give up after max retries",
			responses:    []cannedResponse{{status: 503}},
			wantErr:      errors.New("unexpected status code: 503"),
			wantRequests: 3,
		},
		{
			name: "forbidden without rate limit is not retried",
			responses: []cannedResponse{{
				status: 403,
				header: map[string]string{"X-RateLimit-Remaining": "4999"},
				body:   `{"message":"Resource not accessible by personal access token"}`,
			}},
			wantErr:      errors.New("unexpected status code: 403"),
			wantRequests: 1,
		},
		{
			name:         "not found is not retried",
			responses:    []cannedResponse{{status: 404}},
			wantErr:      ErrNotFound,
			wantRequests: 1,
		},
		{
			name: "primary rate limit",
			responses: []cannedResponse{{
				status: 403,
				header: map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": strconv.FormatInt(resetAt.Unix(), 10)},
				body:   `{"message":"API rate limit exceeded for user ID 1."}`,
			}},
			wantErr:      ErrRateLimited,
			wantRequests: 1,
		},
		{
			name:         "secondary rate limit with retry-after",
			responses:    []cannedResponse{{status: 403, header: map[string]string{"Retry-After": "60"}}},
			wantErr:      ErrRateLimited,
			wantRequests: 1,
		},
		{
			name:         "secondary rate limit from body",
			responses:    []cannedResponse{{status: 403, body: `{"message":"You have exceeded a secondary rate limit."}`}},
			wantErr:      ErrRateLimited,
			wantRequests: 1,
		},
		{
			name:         "too many requests",
			responses:    []cannedResponse{{status: 429}},
			wantErr:      ErrRateLimited,
			wantRequests: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, requests := cannedServer(tc.responses)
			defer srv.Close()

			body, err := NewClientWithBaseURL(srv.URL, "token").GetFileContents(context.Background(), srv.URL+"/file")

			if !sameError(err, tc.wantErr) {
				t.Fatalf("got error %v, want %v", err, tc.wantErr)
			}
			if string(body) != tc.wantBody {
				t.Fatalf("got body %q, want %q", body, tc.wantBody)
			}
			if requests.Load() != tc.wantRequests {
				t.Fatalf("got %d requests, want %d", requests.Load(), tc.wantRequests)
			}
		})
	}
}

func TestRateLimitBlocksRequestsUntilReset(t *testing.T) {
	resetAt := time.Now().Add(time.Hour).Truncate(time.Second)
	srv, requests := cannedServer([]cannedResponse{{
		status: 403,
		header: map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": strconv.FormatInt(resetAt.Unix(), 10)},
	}})
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL, "token")
	for range 3 {
		_, err := c.ListReleases(context.Background(), "o", "r", 10)
		if !errors.Is(err, ErrRateLimited) {
			t.Fatalf("got %v, want ErrRateLimited", err)
		}
		if !strings.Contains(err.Error(), resetAt.UTC().Format(time.RFC3339)) {
			t.Fatalf("error %q does not carry the reset time", err)
		}
	}

	if requests.Load() != 1 {
		t.Fatalf("got %d requests, want 1", requests.Load())
	}
}

func sameError(got, want error) bool {
	if got == nil || want == nil {
		return got == want
	}
	return errors.Is(got, want) || got.Error() == want.Error()
}

func cannedServer(responses []cannedResponse) (*httptest.Server, *atomic.Int32) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(requests.Add(1))
		resp := responses[min(n, len(responses))-1]
		for k, v := range resp.header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(resp.status)
		fmt.Fprint(w, resp.body)
	}))
	return srv, &requests
}
