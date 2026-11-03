package stickers_handler

import (
	"context"
	"encoding/json"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
)

type TrendingInput struct {
	Page         int    `query:"page" minimum:"1" default:"1" doc:"Page number"`
	PerPage      int    `query:"per_page" minimum:"1" maximum:"50" default:"24" doc:"Items per page"`
	CustomerID   string `query:"customer_id" required:"true" doc:"Unique user identifier"`
	Locale       string `query:"locale" doc:"Country code (ISO 3166 Alpha-2)"`
	FormatFilter string `query:"format_filter" doc:"Comma-separated formats: gif, webp, jpg, mp4, webm"`
}

type TrendingOutput struct {
	Body json.RawMessage
}

func (h *HandlerGroup) GetTrending(ctx context.Context, input *TrendingInput) (*TrendingOutput, error) {
	if h.client == nil {
		return nil, ErrClientNotConfigured
	}

	params := klipy.TrendingParams{
		Page:         input.Page,
		PerPage:      input.PerPage,
		CustomerID:   input.CustomerID,
		Locale:       input.Locale,
		FormatFilter: input.FormatFilter,
	}

	resp, err := h.client.GetStickersTrending(ctx, params)
	if err != nil {
		return nil, ErrUpstreamRequest
	}

	return &TrendingOutput{Body: resp}, nil
}
