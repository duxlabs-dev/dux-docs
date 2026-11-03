package registry

import (
	"slices"
	"sync"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/gitlab"
)

func TestParserSharesOneClientPerHostAcrossGoroutines(t *testing.T) {
	p := NewParser("token")
	hosts := []string{"github.com", "codeberg.org"}

	githubClients := make([][]*github.Client, 16)
	gitlabClients := make([]*gitlab.Client, len(githubClients))
	var wg sync.WaitGroup
	for i := range githubClients {
		wg.Go(func() {
			for _, host := range hosts {
				client, err := p.getClient(host)
				if err != nil {
					t.Errorf("getClient(%s): %v", host, err)
					return
				}
				githubClients[i] = append(githubClients[i], client)
			}
			gitlabClients[i] = p.getGitLabClient()
		})
	}
	wg.Wait()

	for i := range githubClients {
		if !slices.Equal(githubClients[i], githubClients[0]) || gitlabClients[i] != gitlabClients[0] {
			t.Fatalf("goroutine %d got its own client instead of the shared one", i)
		}
	}
}
