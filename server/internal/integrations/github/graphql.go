package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

type graphqlError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GraphQL posts a query to the v4 API and decodes its data into out.
func (c *Client) GraphQL(ctx context.Context, query string, vars map[string]any, out any) error {
	if c.token == "" {
		return errors.New("github graphql requires a token")
	}
	if err := c.rateLimitErr(); err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/graphql", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if reset, limited := rateLimitReset(response{status: resp.StatusCode, header: resp.Header, body: body}); limited {
		return c.markRateLimited(reset)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []graphqlError  `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	msgs := make([]string, len(envelope.Errors))
	for i, e := range envelope.Errors {
		// GraphQL reports an exhausted quota as a 200 with an error entry.
		if e.Type == "RATE_LIMITED" {
			return c.markRateLimited(quotaReset(resp.Header))
		}
		msgs[i] = e.Message
	}
	// Errors alongside data are partial failures (e.g. one sponsor org behind
	// SAML); the affected nodes come back null and callers skip them.
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		if len(msgs) == 0 {
			return errors.New("graphql: empty response")
		}
		return fmt.Errorf("graphql: %s", strings.Join(msgs, "; "))
	}
	if len(msgs) > 0 {
		log.Warn("GitHub GraphQL returned partial errors", "errors", strings.Join(msgs, "; "))
	}
	return json.Unmarshal(envelope.Data, out)
}
