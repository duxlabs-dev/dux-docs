package sponsors_handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/sponsors"
)

var ErrCacheNotReady = huma.Error503ServiceUnavailable("sponsors cache is warming up")

type GetSponsorsResponse struct {
	CacheControl string `header:"Cache-Control"`
	Body         models.SponsorsFeed
}

func RegisterHandlers(cache *sponsors.Cache, grp *huma.Group) {
	huma.Register(
		grp,
		huma.Operation{
			OperationID: "get-sponsors",
			Summary:     "Get Sponsors",
			Description: "Public GitHub Sponsors and Ko-fi supporters, split into current and past",
			Path:        "",
			Method:      http.MethodGet,
		},
		func(ctx context.Context, _ *struct{}) (*GetSponsorsResponse, error) {
			if cache == nil || !cache.IsReady() {
				// The edge adds a 10-minute browser cache to responses without one.
				return nil, huma.ErrorWithHeaders(ErrCacheNotReady, http.Header{"Cache-Control": {"no-store"}})
			}
			return &GetSponsorsResponse{
				CacheControl: "public, max-age=300",
				Body:         cache.GetFeed(),
			}, nil
		},
	)
}

const maxKofiBody = 64 << 10

// KofiWebhook handles Ko-fi's form-encoded webhook (JSON in the "data" field).
// Ko-fi retries on non-200, which RecordKofi relies on after a failed write.
func KofiWebhook(token string, cache *sponsors.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxKofiBody)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		var payload sponsors.KofiPayload
		if err := json.Unmarshal([]byte(r.PostForm.Get("data")), &payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if subtle.ConstantTimeCompare([]byte(payload.VerificationToken), []byte(token)) != 1 {
			http.Error(w, "invalid verification token", http.StatusUnauthorized)
			return
		}
		if err := cache.RecordKofi(payload); err != nil {
			log.Error("Failed to record Ko-fi supporter", "err", err)
			http.Error(w, "failed to record", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
