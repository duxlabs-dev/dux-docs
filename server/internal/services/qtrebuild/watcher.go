package qtrebuild

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

const (
	Topic     = "org.fedoraproject.prod.bodhi.update.complete.stable"
	EventType = "fedora-qt-update"
)

type Dispatcher interface {
	Dispatch(ctx context.Context, owner, repo, eventType string, payload any) error
}

type Watcher struct {
	owner    string
	repo     string
	packages map[string]struct{}
	dispatch Dispatcher
}

type Payload struct {
	Source  string   `json:"source"`
	Topic   string   `json:"topic"`
	Update  string   `json:"update"`
	URL     string   `json:"url"`
	Release string   `json:"release"`
	DistTag string   `json:"dist_tag"`
	Builds  []string `json:"builds"`
}

type bodhiMessage struct {
	Update struct {
		Alias  string `json:"alias"`
		URL    string `json:"url"`
		Builds []struct {
			NVR string `json:"nvr"`
		} `json:"builds"`
		Release struct {
			Name    string `json:"name"`
			DistTag string `json:"dist_tag"`
		} `json:"release"`
	} `json:"update"`
}

func NewWatcher(owner, repo string, packages []string, dispatch Dispatcher) *Watcher {
	watched := make(map[string]struct{}, len(packages))
	for _, p := range packages {
		if p = strings.TrimSpace(p); p != "" {
			watched[p] = struct{}{}
		}
	}
	return &Watcher{owner: owner, repo: repo, packages: watched, dispatch: dispatch}
}

func (w *Watcher) Handle(ctx context.Context, topic string, body []byte) {
	if topic != Topic {
		return
	}

	payload, ok := w.match(topic, body)
	if !ok {
		return
	}

	log.Info("Fedora qt6 update reached stable, dispatching rebuild", "release", payload.Release, "builds", payload.Builds)
	if err := w.dispatch.Dispatch(ctx, w.owner, w.repo, EventType, payload); err != nil {
		log.Error("Failed to dispatch qt rebuild", "err", err, "update", payload.Update)
	}
}

func (w *Watcher) match(topic string, body []byte) (Payload, bool) {
	var msg bodhiMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Warn("Unparseable bodhi message", "err", err)
		return Payload{}, false
	}

	var builds []string
	for _, b := range msg.Update.Builds {
		if _, watched := w.packages[packageName(b.NVR)]; watched {
			builds = append(builds, b.NVR)
		}
	}
	if len(builds) == 0 {
		return Payload{}, false
	}

	return Payload{
		Source:  "fedora-messaging",
		Topic:   topic,
		Update:  msg.Update.Alias,
		URL:     msg.Update.URL,
		Release: msg.Update.Release.Name,
		DistTag: msg.Update.Release.DistTag,
		Builds:  builds,
	}, true
}

func packageName(nvr string) string {
	i := strings.LastIndexByte(nvr, '-')
	if i <= 0 {
		return ""
	}
	j := strings.LastIndexByte(nvr[:i], '-')
	if j <= 0 {
		return ""
	}
	return nvr[:j]
}
