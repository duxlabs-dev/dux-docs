package registry

import (
	"errors"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/gitlab"
	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

type carryForward[T any] struct {
	kind         string
	previous     map[string]T
	entries      []T
	rateLimitErr error
	rateLimited  int
}

func newCarryForward[T any](kind string, previous map[string]T) *carryForward[T] {
	return &carryForward[T]{kind: kind, previous: previous}
}

func (c *carryForward[T]) add(key string, entry T, err error) {
	switch {
	case err == nil:
		c.entries = append(c.entries, entry)
		return
	case errors.Is(err, github.ErrNotFound), errors.Is(err, gitlab.ErrNotFound):
		log.Warnf("Dropping %s %s: %v", c.kind, key, err)
		return
	case errors.Is(err, github.ErrRateLimited):
		c.rateLimitErr = err
		c.rateLimited++
	default:
		log.Warnf("Failed to refresh %s %s: %v", c.kind, key, err)
	}

	if prev, ok := c.previous[key]; ok {
		c.entries = append(c.entries, prev)
	}
}

func (c *carryForward[T]) result() []T {
	if c.rateLimitErr != nil {
		log.Warnf("Rate limited, %d %ss not refreshed: %v", c.rateLimited, c.kind, c.rateLimitErr)
	}
	return c.entries
}
