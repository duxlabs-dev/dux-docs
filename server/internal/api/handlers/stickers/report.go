package stickers_handler

import (
	"context"
	"encoding/json"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
)

type ReportInput struct {
	Slug       string `path:"slug" required:"true" doc:"Sticker slug or ID"`
	CustomerID string `json:"customer_id" required:"true" doc:"Unique user identifier"`
	Reason     string `json:"reason" required:"true" enum:"nudity,violence,hate_speech,harassment,spam,misinformation,copyright,offensive,illegal,broken,low_quality,not_relevant,impersonation,other" doc:"Reason for reporting"`
}

type ReportOutput struct {
	Body json.RawMessage
}

func (h *HandlerGroup) Report(ctx context.Context, input *ReportInput) (*ReportOutput, error) {
	if h.client == nil {
		return nil, ErrClientNotConfigured
	}

	params := klipy.ReportParams{
		CustomerID: input.CustomerID,
		Reason:     input.Reason,
	}

	resp, err := h.client.ReportSticker(ctx, input.Slug, params)
	if err != nil {
		return nil, ErrUpstreamRequest
	}

	return &ReportOutput{Body: resp}, nil
}
