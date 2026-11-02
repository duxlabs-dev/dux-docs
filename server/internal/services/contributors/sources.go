package contributors

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/poeditor"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

// GitHub is the slice of the GitHub client this package needs.
type GitHub interface {
	ListContributors(ctx context.Context, owner, repo string) ([]github.Contributor, error)
}

// POEditor is the slice of the POEditor client this package needs.
type POEditor interface {
	Contributors(ctx context.Context, projectID string) ([]poeditor.Contributor, error)
	Languages(ctx context.Context, projectID string) ([]poeditor.Language, error)
}

func isBot(c github.Contributor) bool {
	return strings.EqualFold(c.Type, "Bot") || strings.HasSuffix(strings.ToLower(c.Login), "[bot]")
}

// fetchCode merges contributors across repos ("owner/name"). One failed repo
// fails the whole fetch so a partial result never shrinks anyone's count.
func fetchCode(ctx context.Context, gh GitHub, repos []string, exclude map[string]bool) ([]models.CodeContributor, error) {
	var all []models.CodeContributor
	index := map[string]int{}
	for _, full := range repos {
		owner, repo, ok := strings.Cut(full, "/")
		if !ok {
			return nil, fmt.Errorf("repo %q must be owner/name", full)
		}
		list, err := gh.ListContributors(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", full, err)
		}
		for _, c := range list {
			key := strings.ToLower(c.Login)
			if key == "" || c.Contributions <= 0 || isBot(c) || exclude[key] {
				continue
			}
			i, seen := index[key]
			if !seen {
				i = len(all)
				index[key] = i
				all = append(all, models.CodeContributor{Login: c.Login, AvatarURL: c.AvatarURL, URL: c.HTMLURL})
			}
			all[i].Contributions += c.Contributions
			all[i].Repos = append(all[i].Repos, repo)
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Contributions != all[j].Contributions {
			return all[i].Contributions > all[j].Contributions
		}
		return strings.ToLower(all[i].Login) < strings.ToLower(all[j].Login)
	})
	return all, nil
}

const maxNameLen = 64

// displayName cleans a POEditor name, or returns "" for one that must not be
// published: names are free text and some people use their email address.
func displayName(raw string) string {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" || strings.Contains(name, "@") {
		return ""
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		name = string([]rune(name)[:maxNameLen])
	}
	return name
}

// fetchTranslators lists everyone assigned to a language in projects. People
// are merged across projects by email, which is used only as a key.
func fetchTranslators(ctx context.Context, po POEditor, projects []string) ([]models.Translator, error) {
	var all []models.Translator
	index := map[string]int{}
	for _, project := range projects {
		langs, err := po.Languages(ctx, project)
		if err != nil {
			return nil, fmt.Errorf("project %s: %w", project, err)
		}
		langNames := make(map[string]string, len(langs))
		for _, l := range langs {
			langNames[l.Code] = l.Name
		}
		people, err := po.Contributors(ctx, project)
		if err != nil {
			return nil, fmt.Errorf("project %s: %w", project, err)
		}
		for _, p := range people {
			name := displayName(p.Name)
			if name == "" {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(p.Email))
			if key == "" {
				key = "name:" + strings.ToLower(name)
			}
			for _, perm := range p.Permissions {
				if perm.Project.ID.String() != project || len(perm.Languages) == 0 {
					continue
				}
				i, seen := index[key]
				if !seen {
					i = len(all)
					index[key] = i
					all = append(all, models.Translator{Name: name})
				}
				all[i].Proofreader = all[i].Proofreader || perm.Proofreader
				for _, code := range perm.Languages {
					langName := langNames[code]
					if langName == "" {
						langName = code
					}
					all[i].Languages = append(all[i].Languages, models.TranslatorLanguage{
						ProjectID: project,
						Project:   perm.Project.Name,
						Code:      code,
						Name:      langName,
					})
				}
			}
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		return strings.ToLower(all[i].Name) < strings.ToLower(all[j].Name)
	})
	return all, nil
}
