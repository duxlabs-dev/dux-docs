package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

var ErrNotFound = errors.New("unexpected status code: 404")

func NewClient(token string) *Client {
	return &Client{
		baseURL: "https://gitlab.com/api/v4",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (c *Client) GetRawFile(ctx context.Context, project, filePath, ref string) ([]byte, error) {
	path := fmt.Sprintf("/projects/%s/repository/files/%s/raw?ref=%s",
		url.PathEscape(project), url.PathEscape(filePath), url.QueryEscape(ref))
	return c.get(ctx, path)
}

type commit struct {
	CommittedDate time.Time `json:"committed_date"`
}

func (c *Client) GetLastCommitDate(ctx context.Context, project, path string) (time.Time, error) {
	apiPath := fmt.Sprintf("/projects/%s/repository/commits?per_page=1", url.PathEscape(project))
	if path != "" {
		apiPath += "&path=" + url.QueryEscape(path)
	}

	body, err := c.get(ctx, apiPath)
	if err != nil {
		return time.Time{}, err
	}

	var commits []commit
	if err := json.Unmarshal(body, &commits); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal commits: %w", err)
	}

	if len(commits) == 0 {
		return time.Time{}, fmt.Errorf("no commits found")
	}

	return commits[0].CommittedDate, nil
}
