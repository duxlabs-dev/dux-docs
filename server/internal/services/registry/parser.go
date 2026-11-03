package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/gitlab"
	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type Parser struct {
	token   string
	mu      sync.Mutex
	clients map[string]*github.Client
	gitlab  *gitlab.Client
}

func NewParser(token string) *Parser {
	return &Parser{
		token:   token,
		clients: make(map[string]*github.Client),
	}
}

func (p *Parser) getGitLabClient() *gitlab.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.gitlab == nil {
		p.gitlab = gitlab.NewClient("")
	}
	return p.gitlab
}

func (p *Parser) getClient(host string) (*github.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, ok := p.clients[host]; ok {
		return client, nil
	}

	baseURL, err := getAPIBaseURL(host)
	if err != nil {
		return nil, err
	}

	token := ""
	if host == "github.com" {
		token = p.token
	}

	client := github.NewClientWithBaseURL(baseURL, token)
	p.clients[host] = client
	return client, nil
}

func (p *Parser) FetchPlugins(ctx context.Context, prev []models.Plugin) ([]models.Plugin, error) {
	registryPlugins, err := p.fetchRegistryPlugins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry plugins: %w", err)
	}

	carry := newCarryForward("plugin", pluginsByID(prev))
	for _, regPlugin := range registryPlugins {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		plugin, err := p.enrichPlugin(ctx, regPlugin)
		carry.add(regPlugin.ID, plugin, err)
	}

	return carry.result(), nil
}

func pluginsByID(plugins []models.Plugin) map[string]models.Plugin {
	byID := make(map[string]models.Plugin, len(plugins))
	for _, plugin := range plugins {
		byID[plugin.ID] = plugin
	}
	return byID
}

func (p *Parser) fetchRegistryPlugins(ctx context.Context) ([]models.RegistryPlugin, error) {
	client, err := p.getClient("github.com")
	if err != nil {
		return nil, err
	}

	contents, err := client.GetRepoContents(ctx, "duxlabs-dev", "dux-plugin-registry", "plugins")
	if err != nil {
		return nil, fmt.Errorf("failed to get plugins directory: %w", err)
	}

	var plugins []models.RegistryPlugin
	for _, content := range contents {
		if content.Type != "file" {
			continue
		}

		if !strings.HasSuffix(content.Name, ".json") {
			continue
		}

		fileData, err := client.GetFileContents(ctx, content.DownloadURL)
		switch {
		case errors.Is(err, github.ErrNotFound):
			log.Warnf("Skipping %s: %v", content.Name, err)
			continue
		case err != nil:
			return nil, fmt.Errorf("failed to fetch %s: %w", content.Name, err)
		}

		var plugin models.RegistryPlugin
		if err := json.Unmarshal(fileData, &plugin); err != nil {
			log.Warnf("Failed to parse %s: %v", content.Name, err)
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func (p *Parser) enrichPlugin(ctx context.Context, regPlugin models.RegistryPlugin) (models.Plugin, error) {
	host, owner, repo, err := parseRepoURL(regPlugin.Repo)
	if err != nil {
		return models.Plugin{}, fmt.Errorf("invalid repo URL: %w", err)
	}

	if host == "gitlab.com" {
		return p.enrichPluginGitLab(ctx, regPlugin, owner, repo)
	}

	client, err := p.getClient(host)
	if err != nil {
		return models.Plugin{}, err
	}

	metadataPath := "plugin.json"
	if regPlugin.Path != "" {
		metadataPath = regPlugin.Path + "/plugin.json"
	}

	contents, err := client.GetRepoContents(ctx, owner, repo, metadataPath)
	if err != nil {
		return models.Plugin{}, fmt.Errorf("plugin.json not found or inaccessible: %w", err)
	}

	if len(contents) == 0 {
		return models.Plugin{}, fmt.Errorf("plugin.json not found")
	}

	fileData, err := client.GetFileContents(ctx, contents[0].DownloadURL)
	if err != nil {
		return models.Plugin{}, fmt.Errorf("failed to fetch plugin.json: %w", err)
	}

	metadata, err := parseMetadata(fileData)
	if err != nil {
		return models.Plugin{}, err
	}

	lastCommit, err := client.GetLastCommit(ctx, owner, repo, regPlugin.Path)
	if err != nil {
		return models.Plugin{}, fmt.Errorf("failed to fetch last commit: %w", err)
	}

	return buildPlugin(regPlugin, metadata, lastCommit.Commit.Committer.Date), nil
}

func (p *Parser) enrichPluginGitLab(ctx context.Context, regPlugin models.RegistryPlugin, owner, repo string) (models.Plugin, error) {
	client := p.getGitLabClient()
	project := owner + "/" + repo

	filePath := "plugin.json"
	if regPlugin.Path != "" {
		filePath = regPlugin.Path + "/plugin.json"
	}

	fileData, err := client.GetRawFile(ctx, project, filePath, "HEAD")
	if err != nil {
		return models.Plugin{}, fmt.Errorf("plugin.json not found or inaccessible: %w", err)
	}

	metadata, err := parseMetadata(fileData)
	if err != nil {
		return models.Plugin{}, err
	}

	updatedAt, err := client.GetLastCommitDate(ctx, project, regPlugin.Path)
	if err != nil {
		return models.Plugin{}, fmt.Errorf("failed to fetch last commit: %w", err)
	}

	return buildPlugin(regPlugin, metadata, updatedAt), nil
}

func parseMetadata(data []byte) (models.PluginMetadata, error) {
	var metadata models.PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return models.PluginMetadata{}, fmt.Errorf("invalid plugin.json: %w", err)
	}

	if metadata.Version == "" {
		return models.PluginMetadata{}, fmt.Errorf("plugin.json missing version")
	}

	return metadata, nil
}

func buildPlugin(regPlugin models.RegistryPlugin, metadata models.PluginMetadata, updatedAt time.Time) models.Plugin {
	plugin := models.Plugin{
		ID:           regPlugin.ID,
		Name:         regPlugin.Name,
		Capabilities: withSurfaceCapabilities(regPlugin.Capabilities, metadata),
		Category:     regPlugin.Category,
		Repo:         regPlugin.Repo,
		Author:       regPlugin.Author,
		FirstParty:   regPlugin.FirstParty,
		Description:  regPlugin.Description,
		Dependencies: regPlugin.Dependencies,
		Compositors:  regPlugin.Compositors,
		Distro:       regPlugin.Distro,
		Screenshot:   regPlugin.Screenshot,
		RequiresDMS:  regPlugin.RequiresDMS,
		Version:      metadata.Version,
		Icon:         metadata.Icon,
		Permissions:  metadata.Permissions,
		UpdatedAt:    updatedAt,
	}

	if metadata.Author != "" {
		plugin.Author = metadata.Author
	}

	return plugin
}

var surfaceCapabilities = map[string]string{
	"launcher": "launcher",
	"daemon":   "daemon",
	"desktop":  "desktop-widget",
	"dash":     "dash-tab",
	"dashCard": "dash-card",
}

func manifestSurfaces(metadata models.PluginMetadata) []string {
	if len(metadata.Components) > 0 {
		surfaces := make([]string, 0, len(metadata.Components))
		for surface := range metadata.Components {
			surfaces = append(surfaces, surface)
		}
		sort.Strings(surfaces)
		return surfaces
	}
	if metadata.Type == "" {
		return nil
	}
	return []string{metadata.Type}
}

func withSurfaceCapabilities(capabilities []string, metadata models.PluginMetadata) []string {
	merged := append([]string(nil), capabilities...)
	for _, surface := range manifestSurfaces(metadata) {
		capability, ok := surfaceCapabilities[surface]
		if !ok || slices.Contains(merged, capability) {
			continue
		}
		merged = append(merged, capability)
	}
	return merged
}

func parseRepoURL(repoURL string) (host, owner, repo string, err error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", "", err
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid repo URL format")
	}

	return parsedURL.Host, parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

func getAPIBaseURL(host string) (string, error) {
	switch host {
	case "github.com":
		return "https://api.github.com", nil
	case "codeberg.org":
		return "https://codeberg.org/api/v1", nil
	default:
		return "", fmt.Errorf("unsupported git host: %s", host)
	}
}
