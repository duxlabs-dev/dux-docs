package gifs_handler

import (
	"context"
	"encoding/json"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
)

type ShareInput struct {
	Slug       string `path:"slug" required:"true" doc:"GIF slug or ID"`
	CustomerID string `json:"customer_id" required:"true" doc:"Unique user identifier"`
	Query      string `json:"q" required:"true" doc:"Search string that led to this share"`
}

type ShareOutput struct {
	Body json.RawMessage
}

func (h *HandlerGroup) Share(ctx context.Context, input *ShareInput) (*ShareOutput, error) {
	if h.client == nil {
		return nil, ErrClientNotConfigured
	}

	params := klipy.ShareParams{
		CustomerID: input.CustomerID,
		Query:      input.Query,
	}

	resp, err := h.client.Share(ctx, input.Slug, params)
	if err != nil {
		return nil, ErrUpstreamRequest
	}

	return &ShareOutput{Body: resp}, nil
}
