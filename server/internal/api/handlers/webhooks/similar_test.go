package webhooks

import "testing"

const emptyBody = "Body text\n\n" + similarStart + "\n" + similarEnd + "\n\n<!-- dms-plugin-id: worldClockMulti -->"

func testRender(entries []similarEntry) string {
	return renderSimilarBlock(entries, func(id string) string { return id }, "duxlabs-dev", "dux-plugin-registry")
}

func TestParseSimilarCommands(t *testing.T) {
	cmds := parseSimilarCommands("/similar WorldClock #530\n/unsimilar #42\nnoise /similarity #99")

	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d: %+v", len(cmds), cmds)
	}
	if cmds[0].remove || cmds[0].number != 530 {
		t.Fatalf("unexpected first command: %+v", cmds[0])
	}
	if !cmds[1].remove || cmds[1].number != 42 {
		t.Fatalf("unexpected second command: %+v", cmds[1])
	}
}

func TestParseSimilarCommandsMultipleRefs(t *testing.T) {
	cmds := parseSimilarCommands("/similar #403 #411 #403")

	if len(cmds) != 2 {
		t.Fatalf("expected 2 deduped commands, got %d: %+v", len(cmds), cmds)
	}
	for i, want := range []int{403, 411} {
		if cmds[i].remove || cmds[i].number != want {
			t.Fatalf("command %d: expected add #%d, got %+v", i, want, cmds[i])
		}
	}
}

func TestParseSimilarCommandsMixedVerbsOneLine(t *testing.T) {
	cmds := parseSimilarCommands("/similar #1 #2 /unsimilar #3")

	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d: %+v", len(cmds), cmds)
	}
	if cmds[0].remove || cmds[1].remove {
		t.Fatalf("first two refs should be additions: %+v", cmds)
	}
	if !cmds[2].remove || cmds[2].number != 3 {
		t.Fatalf("trailing ref should be an unlink of #3: %+v", cmds[2])
	}
}

func TestFilterAuthorLabels(t *testing.T) {
	actions := parseCommands("/deprecated /unmaintained /broken /review")

	allowed := filterAuthorLabels(actions)
	if len(allowed) != 2 {
		t.Fatalf("expected only deprecated + unmaintained, got %+v", allowed)
	}
	for _, action := range allowed {
		if !authorLabels[action.label] {
			t.Fatalf("author should not be allowed to set %s", action.label)
		}
	}
}

func TestEditSimilarBlockAddRemove(t *testing.T) {
	added, changed := editSimilarBlock(emptyBody, similarEntry{id: "worldClock", number: 530}, true, testRender)
	if !changed {
		t.Fatal("expected add to change body")
	}
	if got := parseSimilarEntries(added); len(got) != 1 || got[0].id != "worldClock" || got[0].number != 530 {
		t.Fatalf("unexpected entries after add: %+v", got)
	}

	if _, changed := editSimilarBlock(added, similarEntry{id: "worldClock", number: 530}, true, testRender); changed {
		t.Fatal("re-adding the same entry should be a no-op")
	}

	second, changed := editSimilarBlock(added, similarEntry{id: "foo", number: 12}, true, testRender)
	if !changed || len(parseSimilarEntries(second)) != 2 {
		t.Fatalf("expected two entries after second add, got %+v", parseSimilarEntries(second))
	}

	removed, changed := editSimilarBlock(second, similarEntry{id: "worldClock", number: 530}, false, testRender)
	if !changed {
		t.Fatal("expected remove to change body")
	}
	got := parseSimilarEntries(removed)
	if len(got) != 1 || got[0].id != "foo" {
		t.Fatalf("unexpected entries after remove: %+v", got)
	}

	if _, changed := editSimilarBlock(emptyBody, similarEntry{id: "foo", number: 12}, false, testRender); changed {
		t.Fatal("removing an absent entry should be a no-op")
	}
}

func TestEditSimilarBlockInsertsWhenMissing(t *testing.T) {
	body := "Plain body\n\n<!-- dms-plugin-id: foo -->"

	out, changed := editSimilarBlock(body, similarEntry{id: "bar", number: 7}, true, testRender)
	if !changed {
		t.Fatal("expected insert to change body")
	}
	if got := parseSimilarEntries(out); len(got) != 1 || got[0].id != "bar" {
		t.Fatalf("unexpected entries after insert: %+v", got)
	}
}
