package stickers_handler

import (
	"context"
	"encoding/json"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
)

type SearchInput struct {
	Page          int    `query:"page" minimum:"1" default:"1" doc:"Page number"`
	PerPage       int    `query:"per_page" minimum:"8" maximum:"50" default:"24" doc:"Items per page"`
	Query         string `query:"q" required:"true" doc:"Search keyword"`
	CustomerID    string `query:"customer_id" required:"true" doc:"Unique user identifier"`
	Locale        string `query:"locale" doc:"Country code (ISO 3166 Alpha-2)"`
	ContentFilter string `query:"content_filter" enum:"off,low,medium,high" doc:"Content safety filter level"`
	FormatFilter  string `query:"format_filter" doc:"Comma-separated formats: gif, webp, jpg, mp4, webm"`
}

type SearchOutput struct {
	Body json.RawMessage
}

func (h *HandlerGroup) Search(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
	if h.client == nil {
		return nil, ErrClientNotConfigured
	}

	params := klipy.SearchParams{
		Page:          input.Page,
		PerPage:       input.PerPage,
		Query:         input.Query,
		CustomerID:    input.CustomerID,
		Locale:        input.Locale,
		ContentFilter: input.ContentFilter,
		FormatFilter:  input.FormatFilter,
	}

	resp, err := h.client.SearchStickers(ctx, params)
	if err != nil {
		return nil, ErrUpstreamRequest
	}

	return &SearchOutput{Body: resp}, nil
}
