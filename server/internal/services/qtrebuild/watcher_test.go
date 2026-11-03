package qtrebuild

import (
	"context"
	"errors"
	"testing"
)

const stableBody = `{"agent":"bodhi","update":{"alias":"FEDORA-2026-d13f91ad0e","url":"https://bodhi.fedoraproject.org/updates/FEDORA-2026-d13f91ad0e","status":"stable","builds":[{"nvr":"LabPlot-2.12.1-28.fc44"},{"nvr":"mingw-qt6-qtbase-6.11.2-1.fc44"},{"nvr":"qt6-qtbase-6.11.2-2.fc44"},{"nvr":"qt6-qtdeclarative-6.11.2-1.fc44"},{"nvr":"qt6-qtwayland-6.11.2-1.fc44"}],"release":{"name":"F44","branch":"f44","dist_tag":"f44"}}}`

type fakeDispatcher struct {
	calls []Payload
	types []string
	err   error
}

func (f *fakeDispatcher) Dispatch(_ context.Context, _, _, eventType string, payload any) error {
	f.calls = append(f.calls, payload.(Payload))
	f.types = append(f.types, eventType)
	return f.err
}

func newTestWatcher(d Dispatcher) *Watcher {
	return NewWatcher("duxlabs-dev", "dux-repo", []string{"qt6-qtbase", "qt6-qtdeclarative", " qt6-qtwayland "}, d)
}

func TestHandleDispatchesMatchingBuilds(t *testing.T) {
	d := &fakeDispatcher{}
	newTestWatcher(d).Handle(context.Background(), Topic, []byte(stableBody))

	if len(d.calls) != 1 {
		t.Fatalf("expected 1 dispatch, got %d", len(d.calls))
	}
	if d.types[0] != EventType {
		t.Fatalf("expected event type %q, got %q", EventType, d.types[0])
	}
	p := d.calls[0]
	want := []string{"qt6-qtbase-6.11.2-2.fc44", "qt6-qtdeclarative-6.11.2-1.fc44", "qt6-qtwayland-6.11.2-1.fc44"}
	if len(p.Builds) != len(want) {
		t.Fatalf("expected builds %v, got %v", want, p.Builds)
	}
	for i := range want {
		if p.Builds[i] != want[i] {
			t.Fatalf("expected builds %v, got %v", want, p.Builds)
		}
	}
	if p.Release != "F44" || p.DistTag != "f44" || p.Update != "FEDORA-2026-d13f91ad0e" || p.Topic != Topic || p.Source != "fedora-messaging" {
		t.Fatalf("unexpected payload: %+v", p)
	}
}

func TestHandleIgnoresUnwatchedPackages(t *testing.T) {
	d := &fakeDispatcher{}
	body := `{"update":{"alias":"FEDORA-2026-x","builds":[{"nvr":"aqualung-flatpak-2.0-5"},{"nvr":"mingw-qt6-qtbase-6.11.2-1.fc44"}],"release":{"name":"F44","dist_tag":"f44"}}}`
	newTestWatcher(d).Handle(context.Background(), Topic, []byte(body))

	if len(d.calls) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(d.calls))
	}
}

func TestHandleIgnoresOtherTopics(t *testing.T) {
	d := &fakeDispatcher{}
	newTestWatcher(d).Handle(context.Background(), "org.fedoraproject.prod.bodhi.update.request.stable", []byte(stableBody))

	if len(d.calls) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(d.calls))
	}
}

func TestHandleTolerantOfBadBodies(t *testing.T) {
	d := &fakeDispatcher{}
	w := newTestWatcher(d)
	for _, body := range []string{"", "{", "[]", `{"update":null}`, `{"update":{"builds":"nope"}}`} {
		w.Handle(context.Background(), Topic, []byte(body))
	}

	if len(d.calls) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(d.calls))
	}
}

func TestHandleSurvivesDispatchError(t *testing.T) {
	d := &fakeDispatcher{err: errors.New("boom")}
	newTestWatcher(d).Handle(context.Background(), Topic, []byte(stableBody))

	if len(d.calls) != 1 {
		t.Fatalf("expected dispatch attempt, got %d", len(d.calls))
	}
}

func TestPackageName(t *testing.T) {
	cases := map[string]string{
		"qt6-qtbase-6.11.2-2.fc44":       "qt6-qtbase",
		"mingw-qt6-qtbase-6.11.2-1.fc44": "mingw-qt6-qtbase",
		"LabPlot-2.12.1-28.fc44":         "LabPlot",
		"a-b-c":                          "a",
		"bad":                            "",
		"bad-1":                          "",
		"-1-2":                           "",
		"":                               "",
	}
	for nvr, want := range cases {
		if got := packageName(nvr); got != want {
			t.Errorf("packageName(%q) = %q, want %q", nvr, got, want)
		}
	}
}
