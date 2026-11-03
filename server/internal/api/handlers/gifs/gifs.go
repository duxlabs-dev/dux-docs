package gifs_handler

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
		OperationID: "get-trending-gifs",
		Summary:     "Get Trending GIFs",
		Description: "Fetch the most popular and viral GIFs",
		Path:        "/trending",
		Method:      http.MethodGet,
	}, h.GetTrending)

	huma.Register(grp, huma.Operation{
		OperationID: "search-gifs",
		Summary:     "Search GIFs",
		Description: "Search GIFs by keyword or phrase",
		Path:        "/search",
		Method:      http.MethodGet,
	}, h.Search)

	huma.Register(grp, huma.Operation{
		OperationID: "get-gif-categories",
		Summary:     "Get GIF Categories",
		Description: "Retrieve curated categories for GIFs",
		Path:        "/categories",
		Method:      http.MethodGet,
	}, h.GetCategories)

	huma.Register(grp, huma.Operation{
		OperationID: "get-recent-gifs",
		Summary:     "Get Recent GIFs",
		Description: "Retrieve recently used GIFs for a user",
		Path:        "/recent/{customer_id}",
		Method:      http.MethodGet,
	}, h.GetRecent)

	huma.Register(grp, huma.Operation{
		OperationID: "get-gif-items",
		Summary:     "Get GIF Items",
		Description: "Retrieve specific GIFs by IDs or slugs",
		Path:        "/items",
		Method:      http.MethodGet,
	}, h.GetItems)

	huma.Register(grp, huma.Operation{
		OperationID: "delete-recent-gif",
		Summary:     "Hide GIF from Recent",
		Description: "Remove a GIF from a user's recent list",
		Path:        "/recent/{customer_id}",
		Method:      http.MethodDelete,
	}, h.DeleteRecent)

	huma.Register(grp, huma.Operation{
		OperationID: "share-gif",
		Summary:     "Share GIF",
		Description: "Log when a user shares a GIF",
		Path:        "/share/{slug}",
		Method:      http.MethodPost,
	}, h.Share)

	huma.Register(grp, huma.Operation{
		OperationID: "report-gif",
		Summary:     "Report GIF",
		Description: "Report a GIF for inappropriate content",
		Path:        "/report/{slug}",
		Method:      http.MethodPost,
	}, h.Report)
}
