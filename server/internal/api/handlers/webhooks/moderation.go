package webhooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

type Moderator interface {
	IsOrgTeamMember(ctx context.Context, org, team, user string) (bool, error)
	EnsureLabel(ctx context.Context, owner, repo, name, color, description string) error
	AddLabel(ctx context.Context, owner, repo string, issue int, label string) error
	RemoveLabel(ctx context.Context, owner, repo string, issue int, label string) error
	CreateCommentReaction(ctx context.Context, owner, repo string, commentID int64, content string) error
	AppendAudit(ctx context.Context, owner, repo string, issue int, line string) error
	GetIssueBody(ctx context.Context, owner, repo string, issue int) (string, error)
	UpdateIssueBody(ctx context.Context, owner, repo string, issue int, body string) error
}

// PluginLookup resolves plugins by id, author handle, and issue number so the handler can
// block self-review and wire up bidirectional /similar relationships.
type PluginLookup interface {
	RepoOwner(pluginID string) (string, bool)
	PluginByID(id string) (models.Plugin, bool)
	PluginByIssue(number int) (models.Plugin, bool)
}

var pluginIDMarker = regexp.MustCompile(`<!--\s*dms-plugin-id:\s*([A-Za-z0-9]+)\s*-->`)

// selfRestricted commands cannot be used by a plugin's own author (only an Owner can).
var selfRestrictedLabels = map[string]bool{"status:reviewed": true}

// authorLabels are the status flags a plugin's own author may set on their own plugin,
// even without a moderator or owner role.
var authorLabels = map[string]bool{
	"status:deprecated":   true,
	"status:unmaintained": true,
}

type command struct {
	add   bool
	label string
}

type labelMeta struct {
	color       string
	description string
}

var commands = map[string]command{
	"/broken":       {add: true, label: "status:broken"},
	"/working":      {add: false, label: "status:broken"},
	"/unmaintained": {add: true, label: "status:unmaintained"},
	"/deprecated":   {add: true, label: "status:deprecated"},
	"/review":       {add: true, label: "status:reviewed"},
	"/unreview":     {add: false, label: "status:reviewed"},
}

var statusLabels = map[string]labelMeta{
	"status:broken":       {color: "b60205", description: "Reported broken"},
	"status:unmaintained": {color: "fbca04", description: "No longer maintained"},
	"status:deprecated":   {color: "cccccc", description: "Deprecated / retired"},
	"status:reviewed":     {color: "0e8a16", description: "Reviewed by catalog moderators"},
}

func parseCommands(body string) []command {
	var actions []command
	seen := map[string]bool{}

	for _, word := range strings.Fields(body) {
		token := strings.TrimRight(strings.ToLower(word), ".,;:!?")
		cmd, ok := commands[token]
		if !ok || seen[token] {
			continue
		}
		seen[token] = true
		actions = append(actions, cmd)
	}

	return actions
}

func hasPluginLabel(labels []label) bool {
	for _, l := range labels {
		if l.Name == pluginLabel {
			return true
		}
	}
	return false
}

func (h *HandlerGroup) handleComment(p eventPayload) {
	if h.moderator == nil || !hasPluginLabel(p.Issue.Labels) {
		return
	}

	labelActions := parseCommands(p.Comment.Body)
	similarActions := parseSimilarCommands(p.Comment.Body)
	if len(labelActions) == 0 && len(similarActions) == 0 {
		return
	}

	go h.processComment(p, labelActions, similarActions)
}

func (h *HandlerGroup) processComment(p eventPayload, labelActions []command, similarActions []similarCommand) {
	ctx := context.Background()
	user := p.Comment.User.Login
	pluginID := extractPluginID(p.Issue.Body)

	isOwner, isModerator, err := h.membership(ctx, user)
	if err != nil {
		log.Error("Failed to check team membership", "err", err)
		return
	}
	isAuthor := h.isPluginAuthor(pluginID, user)
	if !isOwner && !isModerator && !isAuthor {
		h.react(ctx, p.Comment.ID, "confused")
		return
	}

	switch {
	case isOwner:
		// Owners have no restrictions.
	case isModerator:
		labelActions = h.filterSelfModeration(pluginID, user, labelActions)
	default:
		// Plugin author without a moderator role: limited to self-service status flags.
		labelActions = filterAuthorLabels(labelActions)
		similarActions = nil
	}

	if len(labelActions) == 0 && len(similarActions) == 0 {
		h.react(ctx, p.Comment.ID, "confused")
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	var auditLines []string

	for _, action := range labelActions {
		h.applyCommand(ctx, p.Issue.Number, action)
		if pluginID != "" {
			h.cache.ApplyStatus(pluginID, strings.TrimPrefix(action.label, "status:"), action.add)
		}
		verb := "added"
		if !action.add {
			verb = "removed"
		}
		auditLines = append(auditLines, fmt.Sprintf("- %s · @%s %s `%s`", timestamp, user, verb, action.label))
	}

	for _, action := range similarActions {
		if line := h.applySimilar(ctx, pluginID, user, action, timestamp); line != "" {
			auditLines = append(auditLines, line)
		}
	}

	h.react(ctx, p.Comment.ID, "+1")

	if len(auditLines) == 0 {
		return
	}
	if err := h.moderator.AppendAudit(ctx, h.owner, h.repo, p.Issue.Number, strings.Join(auditLines, "\n")); err != nil {
		log.Error("Failed to append moderation audit log", "err", err)
	}
}

func (h *HandlerGroup) membership(ctx context.Context, user string) (isOwner, isModerator bool, err error) {
	isOwner, err = h.moderator.IsOrgTeamMember(ctx, h.org, h.ownersTeam, user)
	if err != nil {
		return false, false, err
	}
	if isOwner {
		return true, false, nil
	}

	isModerator, err = h.moderator.IsOrgTeamMember(ctx, h.org, h.team, user)
	if err != nil {
		return false, false, err
	}
	return false, isModerator, nil
}

func extractPluginID(body string) string {
	match := pluginIDMarker.FindStringSubmatch(body)
	if match == nil {
		return ""
	}
	return match[1]
}

func (h *HandlerGroup) isPluginAuthor(pluginID, user string) bool {
	if h.authors == nil || pluginID == "" {
		return false
	}
	owner, ok := h.authors.RepoOwner(pluginID)
	return ok && strings.EqualFold(owner, user)
}

// filterSelfModeration drops review/unreview actions when the commenter is the plugin's
// own author, so a moderator can't mark their own plugin reviewed. Owners bypass this
// entirely (checked earlier).
func (h *HandlerGroup) filterSelfModeration(pluginID, user string, actions []command) []command {
	if !h.isPluginAuthor(pluginID, user) {
		return actions
	}

	var allowed []command
	for _, action := range actions {
		if selfRestrictedLabels[action.label] {
			log.Warnf("Blocking self-moderation of %s by author @%s", action.label, user)
			continue
		}
		allowed = append(allowed, action)
	}
	return allowed
}

// filterAuthorLabels keeps only the status flags an unprivileged plugin author may set on
// their own plugin (e.g. /unmaintained, /deprecated).
func filterAuthorLabels(actions []command) []command {
	var allowed []command
	for _, action := range actions {
		if !authorLabels[action.label] {
			continue
		}
		allowed = append(allowed, action)
	}
	return allowed
}

func (h *HandlerGroup) applyCommand(ctx context.Context, issue int, action command) {
	if !action.add {
		if err := h.moderator.RemoveLabel(ctx, h.owner, h.repo, issue, action.label); err != nil {
			log.Error("Failed to remove label", "label", action.label, "err", err)
		}
		return
	}

	meta := statusLabels[action.label]
	if err := h.moderator.EnsureLabel(ctx, h.owner, h.repo, action.label, meta.color, meta.description); err != nil {
		log.Error("Failed to ensure label", "label", action.label, "err", err)
		return
	}
	if err := h.moderator.AddLabel(ctx, h.owner, h.repo, issue, action.label); err != nil {
		log.Error("Failed to add label", "label", action.label, "err", err)
	}
}

func (h *HandlerGroup) react(ctx context.Context, commentID int64, content string) {
	if err := h.moderator.CreateCommentReaction(ctx, h.owner, h.repo, commentID, content); err != nil {
		log.Error("Failed to react to comment", "err", err)
	}
}
