package themes_handler

import (
	"context"
	"math/rand"
	"reflect"
	"sort"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type ThemeSortBy string

const (
	SortByUpdatedAt ThemeSortBy = "updated_at"
	SortByNewest    ThemeSortBy = "newest"
	SortByOldest    ThemeSortBy = "oldest"
	SortByName      ThemeSortBy = "name"
	SortByRandom    ThemeSortBy = "random"
)

func (u ThemeSortBy) Schema(r huma.Registry) *huma.Schema {
	if r.Map()["ThemeSortBy"] != nil {
		return &huma.Schema{Ref: "#/components/schemas/ThemeSortBy"}
	}

	schemaRef := r.Schema(reflect.TypeOf(""), true, "ThemeSortBy")
	schemaRef.Title = "ThemeSortBy"
	schemaRef.Enum = append(schemaRef.Enum, []any{
		string(SortByUpdatedAt),
		string(SortByNewest),
		string(SortByOldest),
		string(SortByName),
		string(SortByRandom),
	}...)
	r.Map()["ThemeSortBy"] = schemaRef
	return &huma.Schema{Ref: "#/components/schemas/ThemeSortBy"}
}

type ListThemesInput struct {
	SortBy ThemeSortBy `query:"sortBy" doc:"Sort themes by field"`
}

type ListThemesResponse struct {
	Body struct {
		Themes []models.Theme `json:"themes"`
		Count  int            `json:"count"`
	}
}

func (h *HandlerGroup) GetThemes(ctx context.Context, input *ListThemesInput) (*ListThemesResponse, error) {
	if h.srv.ThemeCache == nil || !h.srv.ThemeCache.IsReady() {
		return nil, ErrCacheNotReady
	}

	themes := h.srv.ThemeCache.GetThemes()

	sortBy := input.SortBy
	if sortBy == "" {
		sortBy = SortByUpdatedAt
	}

	switch sortBy {
	case SortByName:
		sort.Slice(themes, func(i, j int) bool {
			return themes[i].Name < themes[j].Name
		})
	case SortByRandom:
		rand.Shuffle(len(themes), func(i, j int) {
			themes[i], themes[j] = themes[j], themes[i]
		})
	case SortByOldest:
		sort.Slice(themes, func(i, j int) bool {
			return themes[i].UpdatedAt.Before(themes[j].UpdatedAt)
		})
	default:
		sort.Slice(themes, func(i, j int) bool {
			return themes[i].UpdatedAt.After(themes[j].UpdatedAt)
		})
	}

	resp := &ListThemesResponse{}
	if themes == nil {
		themes = []models.Theme{}
	}
	resp.Body.Themes = themes
	resp.Body.Count = len(themes)

	return resp, nil
}
