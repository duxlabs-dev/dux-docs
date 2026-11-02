package contributors

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

// fetchLanguages counts the people assigned to each language across projects,
// and the distinct people overall. People are told apart by email (or name
// without one), which is used only as a key and never returned.
func fetchLanguages(ctx context.Context, po POEditor, projects []string) ([]models.TranslationLanguage, int, error) {
	type language struct {
		name   string
		people map[string]bool
	}
	byCode := map[string]*language{}
	everyone := map[string]bool{}
	for _, project := range projects {
		langs, err := po.Languages(ctx, project)
		if err != nil {
			return nil, 0, fmt.Errorf("project %s: %w", project, err)
		}
		langNames := make(map[string]string, len(langs))
		for _, l := range langs {
			langNames[l.Code] = l.Name
		}
		people, err := po.Contributors(ctx, project)
		if err != nil {
			return nil, 0, fmt.Errorf("project %s: %w", project, err)
		}
		for _, p := range people {
			key := strings.ToLower(strings.TrimSpace(p.Email))
			if key == "" {
				key = "name:" + strings.ToLower(strings.Join(strings.Fields(p.Name), " "))
			}
			if key == "name:" {
				continue
			}
			for _, perm := range p.Permissions {
				if perm.Project.ID.String() != project {
					continue
				}
				for _, code := range perm.Languages {
					lang := byCode[code]
					if lang == nil {
						lang = &language{name: code, people: map[string]bool{}}
						byCode[code] = lang
					}
					if lang.name == code && langNames[code] != "" {
						lang.name = langNames[code]
					}
					lang.people[key] = true
					everyone[key] = true
				}
			}
		}
	}
	out := make([]models.TranslationLanguage, 0, len(byCode))
	for code, lang := range byCode {
		out = append(out, models.TranslationLanguage{Code: code, Name: lang.name, Translators: len(lang.people)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Translators != out[j].Translators {
			return out[i].Translators > out[j].Translators
		}
		return out[i].Name < out[j].Name
	})
	return out, len(everyone), nil
}
