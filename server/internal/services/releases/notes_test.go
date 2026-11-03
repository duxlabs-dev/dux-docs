package releases

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const boilerplate = "# Package Maintainers - PLEASE READ\n\n" +
	"- Add `-tags withshell` to the normal build tags\n\n" +
	"## Installation\n\n```bash\ncurl -fsSL https://install.danklinux.com | sh\n```\n\n" +
	"## Assets\n\n### Complete Packages\n" +
	"- **`dms-full-amd64.tar.gz`** - Complete package for x86_64 systems\n\n" +
	"### Checksums\n- **`*.sha256`** - SHA256 checksums\n\n---\n\n"

func TestParseNotes(t *testing.T) {
	var manyFeatures strings.Builder
	for i := range 10 {
		fmt.Fprintf(&manyFeatures, "- feat: thing %d by @a in #%d\n", i, i)
	}

	tests := []struct {
		name       string
		body       string
		counts     models.ReleaseCounts
		highlights []string
	}{
		{
			name: "all headings, boilerplate skipped",
			body: boilerplate + "## What's Changed\n\n" +
				"### Breaking Changes\n- drop legacy config by @a in #1\n\n" +
				"### Features\n" +
				"- feat(screenshot): add flag to control HUD scale by @hthienloc in #3398\n" +
				"- Feature/split move size hyprland windowrules by @Mikilio in #2824\n\n" +
				"### Fixes\n" +
				"- fix(clipboard): preserve pinned state by @hthienloc in #3412\n" +
				"- niri: fix workspace updates never applying by @bbedward (a7db035)\n\n" +
				"### Packaging\n- distro: bump spec by @a (1234567)\n\n" +
				"### Internationalization\n- i18n: sync terms by @a (1234567)\n\n" +
				"### Documentation\n- docs: fix typo by @a in #5\n\n" +
				"### Other Changes\n- core: bump dankgop by @bbedward (243d716)\n\n" +
				"**Full Changelog**: https://github.com/duxlabs-dev/dux-shell/compare/v1.6.1...v1.6.2\n",
			counts:     models.ReleaseCounts{Breaking: 1, Features: 2, Fixes: 2, Other: 4},
			highlights: []string{"Split move size hyprland windowrules", "Add flag to control HUD scale", "Niri: fix workspace updates never applying", "Preserve pinned state"},
		},
		{
			name:       "no features falls back to fixes",
			body:       boilerplate + "## What's Changed\n\n### Fixes\n- fix: stop leaking timers by @a in #9\n",
			counts:     models.ReleaseCounts{Fixes: 1},
			highlights: []string{"Stop leaking timers"},
		},
		{
			name:       "highlights are the newest capped",
			body:       "## What's Changed\n### Features\n" + manyFeatures.String(),
			counts:     models.ReleaseCounts{Features: 10},
			highlights: []string{"Thing 9", "Thing 8", "Thing 7", "Thing 6", "Thing 5", "Thing 4", "Thing 3", "Thing 2"},
		},
		{
			name:       "git log fallback counts as other",
			body:       boilerplate + "## What's Changed\n\n- fix words (bdfd565b)\n- bump VERSION 1.5.0 (f3b69859)\n",
			counts:     models.ReleaseCounts{Other: 2},
			highlights: []string{},
		},
		{
			name:       "crlf body",
			body:       "## What's Changed\r\n### Features\r\n- feat: crlf by @a in #1\r\n",
			counts:     models.ReleaseCounts{Features: 1},
			highlights: []string{"Crlf"},
		},
		{
			name:       "empty body",
			body:       "",
			highlights: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts, highlights := parseNotes(tt.body)
			if counts != tt.counts {
				t.Errorf("counts %+v, want %+v", counts, tt.counts)
			}
			if highlights == nil || !reflect.DeepEqual(highlights, tt.highlights) {
				t.Errorf("highlights %#v, want %#v", highlights, tt.highlights)
			}
		})
	}
}

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"feat(media): add toggle for album art accent colors (#2831) by @hthienloc in #2832", "Add toggle for album art accent colors"},
		{"fix: sort by name by @x in #1", "Sort by name"},
		{"feat!: drop legacy config by Jane Doe (abc1234)", "Drop legacy config"},
		{"dock: fix visibility in niri overview by @korbash in #3460", "Dock: fix visibility in niri overview"},
		{"Sort by name", "Sort by name"},
	}
	for _, tt := range tests {
		if got := cleanTitle(tt.in); got != tt.want {
			t.Errorf("cleanTitle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
