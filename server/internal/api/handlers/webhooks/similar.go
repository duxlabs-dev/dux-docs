package webhooks

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

const (
	similarStart = "<!-- dms-similar-start -->"
	similarEnd   = "<!-- dms-similar-end -->"
)

// similarCmdRe matches a `/similar` or `/unsimilar` command keyword. Each command may be
// followed by one or more `#<issue>` references on the same line (e.g. `/similar #403 #411`),
// optionally with free-text plugin names in between (e.g. `/similar WorldClock #530`).
var similarCmdRe = regexp.MustCompile(`/(un)?similar\b`)

// issueRefRe extracts each `#<issue>` reference that follows a similar command.
var issueRefRe = regexp.MustCompile(`#(\d+)`)

// similarDataRe reads the machine-readable payload of the managed similar block: a
// comma-separated list of `id=issueNumber` pairs.
var similarDataRe = regexp.MustCompile(`<!--\s*dms-similar:\s*([^>]*?)\s*-->`)

type similarCommand struct {
	remove bool
	number int
}

type similarEntry struct {
	id     string
	number int
}

func parseSimilarCommands(body string) []similarCommand {
	locs := similarCmdRe.FindAllStringSubmatchIndex(body, -1)

	var cmds []similarCommand
	seen := map[int]bool{}
	for i, loc := range locs {
		remove := loc[2] != -1 // capture group 1 ("un") participated

		// A command owns every issue ref up to the next command, bounded by the line it
		// lives on, so `/similar #1 /unsimilar #2` keeps each ref under the right verb.
		segment := body[loc[1]:]
		if i+1 < len(locs) {
			segment = body[loc[1]:locs[i+1][0]]
		}
		if nl := strings.IndexByte(segment, '\n'); nl != -1 {
			segment = segment[:nl]
		}

		for _, ref := range issueRefRe.FindAllStringSubmatch(segment, -1) {
			number, err := strconv.Atoi(ref[1])
			if err != nil || seen[number] {
				continue
			}
			seen[number] = true
			cmds = append(cmds, similarCommand{remove: remove, number: number})
		}
	}
	return cmds
}

// applySimilar links (or unlinks) the commented-on plugin and the referenced plugin on
// both of their issues, keeping the relationship bidirectional. Returns an audit line, or
// an empty string when nothing changed.
func (h *HandlerGroup) applySimilar(ctx context.Context, pluginID, user string, cmd similarCommand, timestamp string) string {
	if h.authors == nil || pluginID == "" {
		return ""
	}

	self, ok := h.authors.PluginByID(pluginID)
	if !ok {
		return ""
	}

	target, ok := h.authors.PluginByIssue(cmd.number)
	if !ok {
		log.Warnf("similar: no plugin found for issue #%d", cmd.number)
		return ""
	}

	if target.ID == self.ID {
		return ""
	}

	add := !cmd.remove
	changedSelf := h.editIssueSimilar(ctx, self.IssueNumber, similarEntry{id: target.ID, number: target.IssueNumber}, add)
	changedTarget := h.editIssueSimilar(ctx, target.IssueNumber, similarEntry{id: self.ID, number: self.IssueNumber}, add)
	if !changedSelf && !changedTarget {
		return ""
	}

	h.cache.ApplySimilar(self.ID, target.ID, add)
	h.cache.ApplySimilar(target.ID, self.ID, add)

	verb := "linked"
	if cmd.remove {
		verb = "unlinked"
	}
	return fmt.Sprintf("- %s · @%s %s `%s` ↔ `%s`", timestamp, user, verb, self.ID, target.ID)
}

func (h *HandlerGroup) editIssueSimilar(ctx context.Context, issue int, other similarEntry, add bool) bool {
	body, err := h.moderator.GetIssueBody(ctx, h.owner, h.repo, issue)
	if err != nil {
		log.Error("similar: failed to read issue body", "issue", issue, "err", err)
		return false
	}

	newBody, changed := editSimilarBlock(body, other, add, h.renderSimilarBlock)
	if !changed {
		return false
	}

	if err := h.moderator.UpdateIssueBody(ctx, h.owner, h.repo, issue, newBody); err != nil {
		log.Error("similar: failed to update issue body", "issue", issue, "err", err)
		return false
	}
	return true
}

func editSimilarBlock(body string, other similarEntry, add bool, render func([]similarEntry) string) (string, bool) {
	entries := parseSimilarEntries(body)

	idx := -1
	for i, entry := range entries {
		if entry.id == other.id {
			idx = i
			break
		}
	}

	switch {
	case add && idx == -1:
		entries = append(entries, other)
	case !add && idx != -1:
		entries = append(entries[:idx], entries[idx+1:]...)
	default:
		return body, false
	}

	block := render(entries)
	if replaced, ok := replaceSimilarRegion(body, block); ok {
		return replaced, true
	}
	return insertSimilarBlock(body, block), true
}

func parseSimilarEntries(body string) []similarEntry {
	match := similarDataRe.FindStringSubmatch(body)
	if match == nil {
		return nil
	}

	var entries []similarEntry
	for _, part := range strings.Split(match[1], ",") {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		number, err := strconv.Atoi(strings.TrimSpace(pair[1]))
		if err != nil {
			continue
		}
		entries = append(entries, similarEntry{id: strings.TrimSpace(pair[0]), number: number})
	}
	return entries
}

func (h *HandlerGroup) renderSimilarBlock(entries []similarEntry) string {
	return renderSimilarBlock(entries, h.pluginName, h.owner, h.repo)
}

func (h *HandlerGroup) pluginName(id string) string {
	if h.authors == nil {
		return id
	}
	if plugin, ok := h.authors.PluginByID(id); ok && plugin.Name != "" {
		return plugin.Name
	}
	return id
}

// renderSimilarBlock builds the managed block as a markdown list of related plugins, each
// linking to its tracking issue by full name. The hidden `dms-similar` marker carries the
// machine-readable `id=issueNumber` pairs the feedback parser and later edits rely on.
func renderSimilarBlock(entries []similarEntry, nameOf func(string) string, owner, repo string) string {
	if len(entries) == 0 {
		return similarStart + "\n" + similarEnd
	}

	items := make([]string, len(entries))
	data := make([]string, len(entries))
	for i, entry := range entries {
		url := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, entry.number)
		items[i] = fmt.Sprintf("- [%s](%s)", nameOf(entry.id), url)
		data[i] = fmt.Sprintf("%s=%d", entry.id, entry.number)
	}

	return fmt.Sprintf("%s\n**Related plugins:**\n%s\n<!-- dms-similar: %s -->\n%s",
		similarStart, strings.Join(items, "\n"), strings.Join(data, ","), similarEnd)
}

func replaceSimilarRegion(body, block string) (string, bool) {
	start := strings.Index(body, similarStart)
	if start == -1 {
		return "", false
	}
	end := strings.Index(body, similarEnd)
	if end == -1 || end < start {
		return "", false
	}
	end += len(similarEnd)
	return body[:start] + block + body[end:], true
}

func insertSimilarBlock(body, block string) string {
	if loc := pluginIDMarker.FindStringIndex(body); loc != nil {
		return body[:loc[0]] + block + "\n\n" + body[loc[0]:]
	}
	return strings.TrimRight(body, "\n") + "\n\n" + block
}
