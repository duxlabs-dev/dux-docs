package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string

	mu           sync.Mutex
	responses    map[string]cachedResponse
	limitedUntil time.Time
}

type cachedResponse struct {
	etag   string
	header http.Header
	body   []byte
}

type response struct {
	status int
	header http.Header
	body   []byte
}

func NewClient(token string) *Client {
	return NewClientWithBaseURL("https://api.github.com", token)
}

func NewClientWithBaseURL(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:     token,
		responses: make(map[string]cachedResponse),
	}
}

const (
	maxRetries        = 3
	retryBaseDelay    = 250 * time.Millisecond
	rateLimitFallback = time.Minute
	apiAccept         = "application/vnd.github.v3+json"
)

var (
	ErrNotFound    = errors.New("unexpected status code: 404")
	ErrRateLimited = errors.New("github rate limit exceeded")
)

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	body, _, err := c.fetch(ctx, c.baseURL+path, apiAccept)
	return body, err
}

func (c *Client) fetch(ctx context.Context, rawURL, accept string) ([]byte, http.Header, error) {
	if err := c.rateLimitErr(); err != nil {
		return nil, nil, err
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			case <-time.After(retryBaseDelay << (attempt - 1)):
			}
		}

		cached, hasCached := c.cachedResponse(rawURL)
		resp, err := c.fetchOnce(ctx, rawURL, accept, cached.etag)
		if err != nil {
			lastErr = err
			continue
		}

		switch {
		case resp.status == http.StatusOK:
			c.storeResponse(rawURL, resp)
			return resp.body, resp.header, nil
		case resp.status == http.StatusNoContent:
			return nil, resp.header, nil
		case resp.status == http.StatusNotModified && hasCached:
			return cached.body, cached.header, nil
		case resp.status == http.StatusNotFound:
			return nil, nil, ErrNotFound
		}

		if reset, limited := rateLimitReset(resp); limited {
			return nil, nil, c.markRateLimited(reset)
		}

		lastErr = fmt.Errorf("unexpected status code: %d", resp.status)
		if !retryableStatus(resp.status) {
			return nil, nil, lastErr
		}
	}

	return nil, nil, lastErr
}

func (c *Client) fetchOnce(ctx context.Context, rawURL, accept, etag string) (response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return response{}, fmt.Errorf("failed to create request: %w", err)
	}

	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response{}, fmt.Errorf("failed to read response body: %w", err)
	}

	return response{status: resp.StatusCode, header: resp.Header, body: body}, nil
}

func (c *Client) cachedResponse(rawURL string) (cachedResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cached, ok := c.responses[rawURL]
	return cached, ok
}

func (c *Client) storeResponse(rawURL string, resp response) {
	etag := resp.header.Get("ETag")
	if etag == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.responses[rawURL] = cachedResponse{etag: etag, header: resp.header, body: resp.body}
}

func (c *Client) rateLimitErr() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !time.Now().Before(c.limitedUntil) {
		return nil
	}
	return rateLimitedUntil(c.limitedUntil)
}

func (c *Client) markRateLimited(reset time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.limitedUntil = reset
	return rateLimitedUntil(reset)
}

func rateLimitedUntil(reset time.Time) error {
	return fmt.Errorf("%w until %s", ErrRateLimited, reset.UTC().Format(time.RFC3339))
}

func rateLimitReset(resp response) (time.Time, bool) {
	if resp.status != http.StatusForbidden && resp.status != http.StatusTooManyRequests {
		return time.Time{}, false
	}

	if seconds, err := strconv.Atoi(resp.header.Get("Retry-After")); err == nil {
		return time.Now().Add(time.Duration(seconds) * time.Second), true
	}

	if resp.header.Get("X-RateLimit-Remaining") == "0" {
		return quotaReset(resp.header), true
	}

	if resp.status == http.StatusTooManyRequests || bytes.Contains(bytes.ToLower(resp.body), []byte("rate limit")) {
		return time.Now().Add(rateLimitFallback), true
	}

	return time.Time{}, false
}

func quotaReset(header http.Header) time.Time {
	epoch, err := strconv.ParseInt(header.Get("X-RateLimit-Reset"), 10, 64)
	if err != nil {
		return time.Now().Add(rateLimitFallback)
	}
	return time.Unix(epoch, 0)
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

type RepoContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
	URL         string `json:"url"`
}

func (c *Client) GetRepoContents(ctx context.Context, owner, repo, path string) ([]RepoContent, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)

	body, err := c.get(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	var contents []RepoContent
	if err := json.Unmarshal(body, &contents); err != nil {
		var singleContent RepoContent
		if err := json.Unmarshal(body, &singleContent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return []RepoContent{singleContent}, nil
	}

	return contents, nil
}

func (c *Client) GetFileContents(ctx context.Context, url string) ([]byte, error) {
	body, _, err := c.fetch(ctx, url, "")
	return body, err
}

type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

type Issue struct {
	Number      int       `json:"number"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct{} `json:"pull_request,omitempty"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Reactions struct {
		PlusOne int `json:"+1"`
	} `json:"reactions"`
}

func (c *Client) ListIssues(ctx context.Context, owner, repo, label string) ([]Issue, error) {
	var issues []Issue

	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/%s/issues?labels=%s&state=all&per_page=100&page=%d", owner, repo, label, page)

		body, err := c.get(ctx, path)
		if err != nil {
			return nil, err
		}

		var batch []Issue
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal issues: %w", err)
		}

		if len(batch) == 0 {
			break
		}

		issues = append(issues, batch...)
	}

	return issues, nil
}

func (c *Client) GetLastCommit(ctx context.Context, owner, repo, path string) (*Commit, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/commits?per_page=1", owner, repo)
	if path != "" {
		apiPath += fmt.Sprintf("&path=%s", path)
	}

	body, err := c.get(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	if err := json.Unmarshal(body, &commits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commits: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return &commits[0], nil
}

type Release struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
}

func (c *Client) ListReleases(ctx context.Context, owner, repo string, perPage int) ([]Release, error) {
	body, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/releases?per_page=%d", owner, repo, perPage))
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal releases: %w", err)
	}
	return releases, nil
}

type Contributor struct {
	Login         string `json:"login"`
	AvatarURL     string `json:"avatar_url"`
	HTMLURL       string `json:"html_url"`
	Type          string `json:"type"`
	Contributions int    `json:"contributions"`
}

const maxContributorPages = 20

// ListContributors returns a repo's contributors by commit count. Anonymous
// (email-only) authors are left out; they have no profile to link.
func (c *Client) ListContributors(ctx context.Context, owner, repo string) ([]Contributor, error) {
	var all []Contributor
	for page := 1; page <= maxContributorPages; page++ {
		body, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/contributors?per_page=100&page=%d", owner, repo, page))
		if err != nil {
			return nil, err
		}
		// An empty repository answers 204 with no body.
		if len(body) == 0 {
			break
		}
		var batch []Contributor
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contributors: %w", err)
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

// GetRefHead returns the newest commit on ref and the ref's total commit count.
func (c *Client) GetRefHead(ctx context.Context, owner, repo, ref string) (*Commit, int, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/commits?sha=%s&per_page=1", owner, repo, url.QueryEscape(ref))

	body, header, err := c.fetch(ctx, c.baseURL+apiPath, apiAccept)
	if err != nil {
		return nil, 0, err
	}

	var commits []Commit
	if err := json.Unmarshal(body, &commits); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal commits: %w", err)
	}
	if len(commits) == 0 {
		return nil, 0, fmt.Errorf("no commits found")
	}

	// With per_page=1 the rel="last" page number is the commit count; no Link means a single page.
	count := lastPage(header.Get("Link"))
	if count == 0 {
		count = len(commits)
	}
	return &commits[0], count, nil
}

func lastPage(link string) int {
	for _, part := range strings.Split(link, ",") {
		target, params, ok := strings.Cut(part, ";")
		if !ok || !strings.Contains(params, `rel="last"`) {
			continue
		}
		u, err := url.Parse(strings.Trim(strings.TrimSpace(target), "<>"))
		if err != nil {
			return 0
		}
		n, _ := strconv.Atoi(u.Query().Get("page"))
		return n
	}
	return 0
}

func (c *Client) GetFileAtRef(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, url.QueryEscape(ref))

	body, err := c.get(ctx, apiPath)
	if err != nil {
		return nil, err
	}

	var file struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}
	if file.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding %q", file.Encoding)
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
}
