package contributors_handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/contributors"
)

var ErrCacheNotReady = huma.Error503ServiceUnavailable("contributors cache is warming up")

type GetContributorsResponse struct {
	CacheControl string `header:"Cache-Control"`
	Body         models.ContributorsFeed
}

func RegisterHandlers(cache *contributors.Cache, grp *huma.Group) {
	huma.Register(
		grp,
		huma.Operation{
			OperationID: "get-contributors",
			Summary:     "Get Contributors",
			Description: "GitHub code contributors across Dux repos and POEditor translators with their languages",
			Path:        "",
			Method:      http.MethodGet,
		},
		func(ctx context.Context, _ *struct{}) (*GetContributorsResponse, error) {
			if cache == nil || !cache.IsReady() {
				// The edge adds a 10-minute browser cache to responses without one.
				return nil, huma.ErrorWithHeaders(ErrCacheNotReady, http.Header{"Cache-Control": {"no-store"}})
			}
			return &GetContributorsResponse{
				CacheControl: "public, max-age=900",
				Body:         cache.GetFeed(),
			}, nil
		},
	)
}
