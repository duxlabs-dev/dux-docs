package registry

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const statusLabelPrefix = "status:"

var markerRe = regexp.MustCompile(`<!--\s*dms-plugin-id:\s*([A-Za-z0-9]+)\s*-->`)
var similarRe = regexp.MustCompile(`<!--\s*dms-similar:\s*([^>]*?)\s*-->`)

type Feedback struct {
	Upvotes     int
	IssueURL    string
	IssueNumber int
	CreatedAt   time.Time
	Status      []string
	Similar     []string
}

func (p *Parser) FetchFeedback(ctx context.Context) (map[string]Feedback, error) {
	client, err := p.getClient("github.com")
	if err != nil {
		return nil, err
	}

	issues, err := client.ListIssues(ctx, "duxlabs-dev", "dux-plugin-registry", "plugin")
	if err != nil {
		return nil, err
	}

	feedback := make(map[string]Feedback, len(issues))
	for _, issue := range issues {
		if issue.PullRequest != nil {
			continue
		}

		match := markerRe.FindStringSubmatch(issue.Body)
		if match == nil {
			continue
		}

		feedback[match[1]] = Feedback{
			Upvotes:     issue.Reactions.PlusOne,
			IssueURL:    issue.HTMLURL,
			IssueNumber: issue.Number,
			CreatedAt:   issue.CreatedAt,
			Status:      extractStatus(issue),
			Similar:     extractSimilar(issue.Body),
		}
	}

	return feedback, nil
}

func mergeFeedback(plugins []models.Plugin, feedback map[string]Feedback) {
	for i := range plugins {
		fb, ok := feedback[plugins[i].ID]
		if !ok {
			continue
		}
		plugins[i].Upvotes = fb.Upvotes
		plugins[i].IssueURL = fb.IssueURL
		plugins[i].IssueNumber = fb.IssueNumber
		plugins[i].CreatedAt = fb.CreatedAt
		plugins[i].Status = fb.Status
		plugins[i].Similar = fb.Similar
	}
}

// carryFeedback preserves moderation state across registry refreshes, which would
// otherwise reset it until the next feedback cycle and churn the preview cards.
func carryFeedback(fresh, prev []models.Plugin) {
	byID := make(map[string]*models.Plugin, len(prev))
	for i := range prev {
		byID[prev[i].ID] = &prev[i]
	}
	for i := range fresh {
		old, ok := byID[fresh[i].ID]
		if !ok {
			continue
		}
		fresh[i].Upvotes = old.Upvotes
		fresh[i].IssueURL = old.IssueURL
		fresh[i].IssueNumber = old.IssueNumber
		fresh[i].CreatedAt = old.CreatedAt
		fresh[i].Status = old.Status
		fresh[i].Similar = old.Similar
	}
}

// extractSimilar reads the moderator-managed `dms-similar` marker, whose payload is a
// comma-separated list of `id=issueNumber` pairs, and returns the related plugin ids.
func extractSimilar(body string) []string {
	match := similarRe.FindStringSubmatch(body)
	if match == nil {
		return nil
	}

	var ids []string
	for _, part := range strings.Split(match[1], ",") {
		id := strings.TrimSpace(part)
		if idx := strings.Index(id, "="); idx != -1 {
			id = strings.TrimSpace(id[:idx])
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func extractStatus(issue github.Issue) []string {
	var status []string
	for _, label := range issue.Labels {
		if !strings.HasPrefix(label.Name, statusLabelPrefix) {
			continue
		}
		status = append(status, strings.TrimPrefix(label.Name, statusLabelPrefix))
	}
	return status
}
