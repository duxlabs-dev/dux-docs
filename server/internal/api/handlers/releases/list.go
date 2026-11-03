package releases_handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type ConditionalInput struct {
	IfNoneMatch string `header:"If-None-Match" doc:"ETag from a previous response"`
}

type GetReleasesResponse struct {
	ETag string `header:"ETag"`
	Body models.ReleasesFeed
}

type GetLatestResponse struct {
	ETag string `header:"ETag"`
	Body *models.Release
}

func (h *HandlerGroup) GetReleases(ctx context.Context, input *ConditionalInput) (*GetReleasesResponse, error) {
	if h.srv.ReleasesCache == nil || !h.srv.ReleasesCache.IsReady() {
		return nil, ErrCacheNotReady
	}

	feed, etag := h.srv.ReleasesCache.GetFeed()
	quoted := `"` + etag + `"`
	if matchesETag(input.IfNoneMatch, quoted) {
		return nil, notModified(quoted)
	}
	return &GetReleasesResponse{ETag: quoted, Body: feed}, nil
}

func (h *HandlerGroup) GetLatest(ctx context.Context, input *ConditionalInput) (*GetLatestResponse, error) {
	if h.srv.ReleasesCache == nil || !h.srv.ReleasesCache.IsReady() {
		return nil, ErrCacheNotReady
	}

	latest, etag := h.srv.ReleasesCache.GetLatest()
	if latest == nil {
		return nil, ErrNoStable
	}
	quoted := `"` + etag + `"`
	if matchesETag(input.IfNoneMatch, quoted) {
		return nil, notModified(quoted)
	}
	return &GetLatestResponse{ETag: quoted, Body: latest}, nil
}

func notModified(etag string) error {
	return huma.ErrorWithHeaders(huma.Status304NotModified(), http.Header{"ETag": {etag}})
}

// Proxies weaken ETags (W/) when they compress the body.
func matchesETag(header, quoted string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == quoted {
			return true
		}
	}
	return false
}
