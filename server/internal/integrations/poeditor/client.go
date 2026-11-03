// Package poeditor is a read-only client for the POEditor v2 API.
package poeditor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return NewClientWithBaseURL("https://api.poeditor.com/v2/", token)
}

func NewClientWithBaseURL(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/") + "/",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type Permission struct {
	Project struct {
		// POEditor sends the id as a string here but as a number elsewhere.
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
	} `json:"project"`
	Type        string   `json:"type"`
	Proofreader bool     `json:"proofreader"`
	Languages   []string `json:"languages"`
}

type Contributor struct {
	Name string `json:"name"`
	// Email is the only stable identifier POEditor returns. Never expose it.
	Email       string       `json:"email"`
	Permissions []Permission `json:"permissions"`
}

type Language struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Percentage float64 `json:"percentage"`
}

// Contributors lists everyone assigned to the project. POEditor has no
// per-person activity, so this is membership, not who translated what.
func (c *Client) Contributors(ctx context.Context, projectID string) ([]Contributor, error) {
	var out struct {
		Contributors []Contributor `json:"contributors"`
	}
	if err := c.call(ctx, "contributors/list", projectID, &out); err != nil {
		return nil, err
	}
	return out.Contributors, nil
}

func (c *Client) Languages(ctx context.Context, projectID string) ([]Language, error) {
	var out struct {
		Languages []Language `json:"languages"`
	}
	if err := c.call(ctx, "languages/list", projectID, &out); err != nil {
		return nil, err
	}
	return out.Languages, nil
}

func (c *Client) call(ctx context.Context, endpoint, projectID string, out any) error {
	form := url.Values{"api_token": {c.token}, "id": {projectID}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// The URL holds no secret (the token is in the body), so the error is safe to log.
		return fmt.Errorf("poeditor %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("poeditor %s: %w", endpoint, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("poeditor %s: status %d", endpoint, resp.StatusCode)
	}

	var envelope struct {
		Response struct {
			Status  string `json:"status"`
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"response"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("poeditor %s: %w", endpoint, err)
	}
	if envelope.Response.Status != "success" {
		return fmt.Errorf("poeditor %s: %s %s", endpoint, envelope.Response.Code, envelope.Response.Message)
	}
	return json.Unmarshal(envelope.Result, out)
}
