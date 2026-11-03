package server

import (
	"github.com/duxlabs-dev/dux-docs/server/internal/services/registry"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/releases"
)

type EmptyInput struct{}

type Server struct {
	PluginCache   *registry.Cache
	ThemeCache    *registry.ThemeCache
	ReleasesCache *releases.Cache
}
