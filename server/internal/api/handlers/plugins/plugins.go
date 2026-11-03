package plugins_handler

// Handles initial bootstrapping requirements

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/server"
)

type HandlerGroup struct {
	srv       *server.Server
	setupDone bool
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{
		srv: server,
	}

	huma.Register(
		grp,
		huma.Operation{
			OperationID: "get-plugins",
			Summary:     "Get All Plugins",
			Description: "Get All Plugins from the Dux Registry",
			Path:        "",
			Method:      http.MethodGet,
		},
		handlers.GetPlugins,
	)
}
