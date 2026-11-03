package releases

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const maxHighlights = 8

var (
	// Credit written by DMS scripts/release-notes.py: " by @login in #N" or " by <login|name> (sha7)".
	creditRe = regexp.MustCompile(`^ by \S.*?(?: in #\d+| \([0-9a-f]{7,40}\))$`)
	prRefRe  = regexp.MustCompile(`\s*\(#\d+\)$`)
	// Mirrors TYPE_PREFIX_RE in DMS scripts/release-notes.py.
	typePrefixRe = regexp.MustCompile(`(?i)^(?:feat(?:ure)?|fix(?:es)?|hotfix|bugfix|docs?|refactor|chore|perf|style|test|i18n|build|ci)\b!?\s*(?:\([^)]*\))?\s*[:/\-]\s*`)
)

// parseNotes reads the "## What's Changed" section of a DMS release body; the
// install/assets boilerplate release.yml prepends is skipped.
func parseNotes(body string) (models.ReleaseCounts, []string) {
	var counts models.ReleaseCounts
	var features, fixes []string
	var bucket *int
	var titles *[]string

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "## "):
			bucket, titles = nil, nil
			// Unheaded items are release.yml's git-log fallback.
			if strings.TrimSpace(line[3:]) == "What's Changed" {
				bucket = &counts.Other
			}
		case strings.HasPrefix(line, "### "):
			if bucket == nil {
				continue
			}
			switch strings.TrimSpace(line[4:]) {
			case "Breaking Changes":
				bucket, titles = &counts.Breaking, nil
			case "Features":
				bucket, titles = &counts.Features, &features
			case "Fixes":
				bucket, titles = &counts.Fixes, &fixes
			default:
				bucket, titles = &counts.Other, nil
			}
		case strings.HasPrefix(line, "- ") && bucket != nil:
			*bucket++
			if titles != nil {
				*titles = append(*titles, cleanTitle(line[2:]))
			}
		}
	}

	// Features first, fixes fill the rest; patch releases are mostly fixes.
	highlights := make([]string, 0, maxHighlights)
	for _, src := range [][]string{features, fixes} {
		// release-notes.py lists oldest first; the head of a minor's list is mostly items already backported.
		for i := len(src) - 1; i >= 0 && len(highlights) < maxHighlights; i-- {
			highlights = append(highlights, src[i])
		}
	}
	return counts, highlights
}

func cleanTitle(item string) string {
	if i := strings.LastIndex(item, " by "); i >= 0 && creditRe.MatchString(item[i:]) {
		item = item[:i]
	}
	item = strings.TrimSpace(prRefRe.ReplaceAllString(item, ""))
	t := strings.TrimSpace(typePrefixRe.ReplaceAllString(item, ""))
	if t == "" {
		return item
	}
	r, size := utf8.DecodeRuneInString(t)
	return string(unicode.ToUpper(r)) + t[size:]
}
