package plugins_handler

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/registry"
)

type PluginSortBy string

const (
	SortByUpvotes   PluginSortBy = "upvotes"
	SortByUpdatedAt PluginSortBy = "updated_at"
	SortByCreatedAt PluginSortBy = "created_at"
	SortByName      PluginSortBy = "name"
	SortByRandom    PluginSortBy = "random"
)

type PluginSortOrder string

const (
	OrderAscending  PluginSortOrder = "asc"
	OrderDescending PluginSortOrder = "desc"
)

func (u PluginSortBy) Schema(r huma.Registry) *huma.Schema {
	if r.Map()["PluginSortBy"] == nil {
		schemaRef := r.Schema(reflect.TypeOf(""), true, "PluginSortBy")
		schemaRef.Title = "PluginSortBy"
		schemaRef.Enum = append(schemaRef.Enum, []any{
			string(SortByUpvotes),
			string(SortByUpdatedAt),
			string(SortByCreatedAt),
			string(SortByName),
			string(SortByRandom),
		}...)
		r.Map()["PluginSortBy"] = schemaRef
	}
	return &huma.Schema{Ref: "#/components/schemas/PluginSortBy"}
}

func (u PluginSortOrder) Schema(r huma.Registry) *huma.Schema {
	if r.Map()["PluginSortOrder"] == nil {
		schemaRef := r.Schema(reflect.TypeOf(""), true, "PluginSortOrder")
		schemaRef.Title = "PluginSortOrder"
		schemaRef.Enum = append(schemaRef.Enum, []any{
			string(OrderAscending),
			string(OrderDescending),
		}...)
		r.Map()["PluginSortOrder"] = schemaRef
	}
	return &huma.Schema{Ref: "#/components/schemas/PluginSortOrder"}
}

type ListPluginsInput struct {
	Category      string          `query:"category" doc:"Filter by category, comma-separated for any of several (utilities, monitoring, system, appearance, productivity, media, networking)"`
	Compositor    string          `query:"compositor" doc:"Filter by compositor, comma-separated for any of several (niri, hyprland, sway, any)"`
	FirstParty    bool            `query:"firstParty" doc:"Only show first-party plugins"`
	Capability    string          `query:"capability" doc:"Filter by capability, comma-separated for any of several. Plugin types: dankbar-widget, control-center, launcher, daemon, desktop-widget, dash-tab, dash-card. Types declared by the plugin manifest (type or components) are merged into each plugin's capabilities"`
	ExcludeStatus []string        `query:"excludeStatus" doc:"Exclude plugins with these status labels (e.g. broken, deprecated)"`
	SortBy        PluginSortBy    `query:"sortBy" doc:"Sort plugins by field"`
	Order         PluginSortOrder `query:"order" doc:"Sort direction (asc or desc); defaults to descending, ascending for name"`
}

type ListPluginsResponse struct {
	Body struct {
		Plugins []models.Plugin `json:"plugins"`
		Count   int             `json:"count"`
	}
}

func (self *HandlerGroup) GetPlugins(ctx context.Context, input *ListPluginsInput) (*ListPluginsResponse, error) {
	if self.srv.PluginCache == nil || !self.srv.PluginCache.IsReady() {
		return nil, ErrCacheNotReady
	}

	filterOpts := registry.FilterOptions{
		Categories:    splitList(input.Category),
		Compositors:   splitList(input.Compositor),
		FirstParty:    input.FirstParty,
		Capabilities:  splitList(input.Capability),
		ExcludeStatus: input.ExcludeStatus,
	}

	plugins := self.srv.PluginCache.FilterPlugins(filterOpts)

	sortBy := input.SortBy
	if sortBy == "" {
		sortBy = SortByUpvotes
	}

	if sortBy == SortByRandom {
		rand.Shuffle(len(plugins), func(i, j int) {
			plugins[i], plugins[j] = plugins[j], plugins[i]
		})
	} else {
		descending := resolveOrder(sortBy, input.Order)
		sort.SliceStable(plugins, func(i, j int) bool {
			cmp := comparePlugins(plugins[i], plugins[j], sortBy)
			if cmp != 0 {
				if descending {
					return cmp > 0
				}
				return cmp < 0
			}
			return strings.ToLower(plugins[i].Name) < strings.ToLower(plugins[j].Name)
		})
	}

	resp := &ListPluginsResponse{}
	if plugins == nil {
		plugins = []models.Plugin{}
	}
	resp.Body.Plugins = plugins
	resp.Body.Count = len(plugins)

	return resp, nil
}

// comparePlugins returns the ascending ordering of a and b for the given sort field
// (negative when a sorts first). Direction is applied by the caller.
func comparePlugins(a, b models.Plugin, sortBy PluginSortBy) int {
	switch sortBy {
	case SortByUpdatedAt:
		return compareTime(a.UpdatedAt, b.UpdatedAt)
	case SortByCreatedAt:
		return compareTime(a.CreatedAt, b.CreatedAt)
	case SortByName:
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	default:
		if a.Upvotes != b.Upvotes {
			return compareInt(a.Upvotes, b.Upvotes)
		}
		if ar, br := isReviewed(a), isReviewed(b); ar != br {
			if ar {
				return 1
			}
			return -1
		}
		return compareTime(a.UpdatedAt, b.UpdatedAt)
	}
}

// resolveOrder maps an explicit asc/desc to a descending flag, falling back to each
// field's natural default: name reads best A→Z, everything else newest/highest first.
func resolveOrder(sortBy PluginSortBy, order PluginSortOrder) bool {
	switch order {
	case OrderAscending:
		return false
	case OrderDescending:
		return true
	default:
		return sortBy != SortByName
	}
}

func splitList(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		values = append(values, item)
	}
	return values
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareTime(a, b time.Time) int {
	switch {
	case a.Before(b):
		return -1
	case a.After(b):
		return 1
	default:
		return 0
	}
}

func isReviewed(plugin models.Plugin) bool {
	for _, status := range plugin.Status {
		if status == "reviewed" {
			return true
		}
	}
	return false
}
