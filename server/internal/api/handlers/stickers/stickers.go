package stickers_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/server"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
)

type HandlerGroup struct {
	srv    *server.Server
	client *klipy.Client
}

func RegisterHandlers(srv *server.Server, client *klipy.Client, grp *huma.Group) {
	h := &HandlerGroup{
		srv:    srv,
		client: client,
	}

	huma.Register(grp, huma.Operation{
		OperationID: "get-trending-stickers",
		Summary:     "Get Trending Stickers",
		Description: "Fetch the most popular stickers",
		Path:        "/trending",
		Method:      http.MethodGet,
	}, h.GetTrending)

	huma.Register(grp, huma.Operation{
		OperationID: "search-stickers",
		Summary:     "Search Stickers",
		Description: "Search stickers by keyword or phrase",
		Path:        "/search",
		Method:      http.MethodGet,
	}, h.Search)

	huma.Register(grp, huma.Operation{
		OperationID: "get-sticker-categories",
		Summary:     "Get Sticker Categories",
		Description: "Retrieve curated categories for stickers",
		Path:        "/categories",
		Method:      http.MethodGet,
	}, h.GetCategories)

	huma.Register(grp, huma.Operation{
		OperationID: "get-recent-stickers",
		Summary:     "Get Recent Stickers",
		Description: "Retrieve recently used stickers for a user",
		Path:        "/recent/{customer_id}",
		Method:      http.MethodGet,
	}, h.GetRecent)

	huma.Register(grp, huma.Operation{
		OperationID: "get-sticker-items",
		Summary:     "Get Sticker Items",
		Description: "Retrieve specific stickers by IDs or slugs",
		Path:        "/items",
		Method:      http.MethodGet,
	}, h.GetItems)

	huma.Register(grp, huma.Operation{
		OperationID: "delete-recent-sticker",
		Summary:     "Hide Sticker from Recent",
		Description: "Remove a sticker from a user's recent list",
		Path:        "/recent/{customer_id}",
		Method:      http.MethodDelete,
	}, h.DeleteRecent)

	huma.Register(grp, huma.Operation{
		OperationID: "share-sticker",
		Summary:     "Share Sticker",
		Description: "Log when a user shares a sticker",
		Path:        "/share/{slug}",
		Method:      http.MethodPost,
	}, h.Share)

	huma.Register(grp, huma.Operation{
		OperationID: "report-sticker",
		Summary:     "Report Sticker",
		Description: "Report a sticker for inappropriate content",
		Path:        "/report/{slug}",
		Method:      http.MethodPost,
	}, h.Report)
}
