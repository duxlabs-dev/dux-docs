package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

func (p *Parser) FetchThemes(ctx context.Context, prev []models.Theme) ([]models.Theme, error) {
	client, err := p.getClient("github.com")
	if err != nil {
		return nil, err
	}

	contents, err := client.GetRepoContents(ctx, "duxlabs-dev", "dux-plugin-registry", "themes")
	if err != nil {
		return nil, fmt.Errorf("failed to get themes directory: %w", err)
	}

	carry := newCarryForward("theme", themesByDir(prev))
	for _, content := range contents {
		if content.Type != "dir" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		theme, err := p.fetchTheme(ctx, content.Name)
		carry.add(content.Name, theme, err)
	}

	return carry.result(), nil
}

// Inverts themePreviewURL, the only persisted trace of a theme's registry dir.
func themesByDir(themes []models.Theme) map[string]models.Theme {
	byDir := make(map[string]models.Theme, len(themes))
	for _, theme := range themes {
		byDir[path.Base(path.Dir(theme.PreviewURL))] = theme
	}
	return byDir
}

func themePreviewURL(themeName string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/duxlabs-dev/dux-plugin-registry/main/themes/%s/preview.svg", themeName)
}

func (p *Parser) fetchTheme(ctx context.Context, themeName string) (models.Theme, error) {
	client, err := p.getClient("github.com")
	if err != nil {
		return models.Theme{}, err
	}

	themeJSONPath := fmt.Sprintf("themes/%s/theme.json", themeName)

	contents, err := client.GetRepoContents(ctx, "duxlabs-dev", "dux-plugin-registry", themeJSONPath)
	if err != nil {
		return models.Theme{}, fmt.Errorf("failed to get theme.json: %w", err)
	}

	if len(contents) == 0 {
		return models.Theme{}, fmt.Errorf("theme.json not found")
	}

	fileData, err := client.GetFileContents(ctx, contents[0].DownloadURL)
	if err != nil {
		return models.Theme{}, fmt.Errorf("failed to fetch theme.json: %w", err)
	}

	var theme models.Theme
	if err := json.Unmarshal(fileData, &theme); err != nil {
		return models.Theme{}, fmt.Errorf("invalid theme.json: %w", err)
	}

	if theme.ID == "" {
		return models.Theme{}, fmt.Errorf("theme.json missing id")
	}

	lastCommit, err := client.GetLastCommit(ctx, "duxlabs-dev", "dux-plugin-registry", fmt.Sprintf("themes/%s", themeName))
	if err != nil {
		return models.Theme{}, fmt.Errorf("failed to fetch last commit: %w", err)
	}

	theme.PreviewURL = themePreviewURL(themeName)
	theme.UpdatedAt = lastCommit.Commit.Committer.Date
	theme.WCAG = computeThemeWCAG(&theme)

	return theme, nil
}
