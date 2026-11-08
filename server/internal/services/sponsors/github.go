package sponsors

import (
	"context"
	"fmt"
	"strings"
)

// GraphQL is the slice of the GitHub client this package needs.
type GraphQL interface {
	GraphQL(ctx context.Context, query string, vars map[string]any, out any) error
}

// Only public sponsorships are requested; private sponsors never leave GitHub.
const sponsorshipsQuery = `query($login: String!, $active: Boolean!, $after: String) {
  repositoryOwner(login: $login) {
    ... on Sponsorable {
      sponsorshipsAsMaintainer(first: 100, after: $after, activeOnly: $active, includePrivate: false) {
        pageInfo { hasNextPage endCursor }
        nodes {
          sponsorEntity {
            ... on User { login name avatarUrl(size: 160) url }
            ... on Organization { login name avatarUrl(size: 160) url }
          }
        }
      }
    }
  }
}`

const maxSponsorPages = 50

type githubSponsor struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	URL       string `json:"url"`
	Active    bool   `json:"active"`
}

type sponsorshipsPage struct {
	RepositoryOwner *struct {
		SponsorshipsAsMaintainer *struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []struct {
				SponsorEntity *githubSponsor `json:"sponsorEntity"`
			} `json:"nodes"`
		} `json:"sponsorshipsAsMaintainer"`
	} `json:"repositoryOwner"`
}

// fetchGithubSponsors lists all public sponsors of owners. Active status comes
// from a second activeOnly query because isActive needs the read:user scope.
func fetchGithubSponsors(ctx context.Context, gql GraphQL, owners []string) ([]githubSponsor, error) {
	var all []githubSponsor
	index := map[string]int{}
	for _, owner := range owners {
		everyone, err := listSponsorships(ctx, gql, owner, false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", owner, err)
		}
		active, err := listSponsorships(ctx, gql, owner, true)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", owner, err)
		}
		activeSet := make(map[string]bool, len(active))
		for _, s := range active {
			activeSet[strings.ToLower(s.Login)] = true
		}
		// Merge both lists in case they disagree between calls.
		for _, s := range append(everyone, active...) {
			key := strings.ToLower(s.Login)
			s.Active = activeSet[key]
			if i, ok := index[key]; ok {
				all[i].Active = all[i].Active || s.Active
				continue
			}
			index[key] = len(all)
			all = append(all, s)
		}
	}
	return all, nil
}

func listSponsorships(ctx context.Context, gql GraphQL, owner string, activeOnly bool) ([]githubSponsor, error) {
	var out []githubSponsor
	var after any
	// A page cap guards against a cursor that never advances.
	for range maxSponsorPages {
		var page sponsorshipsPage
		vars := map[string]any{"login": owner, "active": activeOnly, "after": after}
		if err := gql.GraphQL(ctx, sponsorshipsQuery, vars, &page); err != nil {
			return nil, err
		}
		if page.RepositoryOwner == nil || page.RepositoryOwner.SponsorshipsAsMaintainer == nil {
			return out, nil
		}
		conn := page.RepositoryOwner.SponsorshipsAsMaintainer
		for _, n := range conn.Nodes {
			// Deleted or suspended accounts come back as null or empty entities.
			if n.SponsorEntity == nil || n.SponsorEntity.Login == "" {
				continue
			}
			out = append(out, *n.SponsorEntity)
		}
		if !conn.PageInfo.HasNextPage || conn.PageInfo.EndCursor == "" {
			return out, nil
		}
		after = conn.PageInfo.EndCursor
	}
	return out, nil
}
