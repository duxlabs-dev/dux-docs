package poeditor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	discord "github.com/latte-soft/discord-webhooks-go"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

const notifyCooldown = 1 * time.Hour

const translatorRoleID = "1474112366412300309"

type HandlerGroup struct {
	callbackSecret string
	webhookURL     string

	mu             sync.Mutex
	lastNotifiedAt time.Time
}

type callbackPayload struct {
	Event struct {
		Name string `json:"name"`
	} `json:"event"`
	Project struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Public  int    `json:"public"`
		Open    int    `json:"open"`
		Created string `json:"created"`
	} `json:"project"`
}

type CallbackInput struct {
	Secret  string `header:"X-Callback-Secret" required:"false"`
	RawBody []byte
}

type CallbackOutput struct{}

func RegisterHandlers(callbackSecret, webhookURL string, grp *huma.Group) {
	h := &HandlerGroup{
		callbackSecret: callbackSecret,
		webhookURL:     webhookURL,
	}

	huma.Register(grp, huma.Operation{
		OperationID: "poeditor-callback",
		Summary:     "POEditor Callback",
		Path:        "",
		Method:      http.MethodPost,
	}, h.HandleCallback)
}

func (h *HandlerGroup) HandleCallback(_ context.Context, input *CallbackInput) (*CallbackOutput, error) {
	if h.callbackSecret != "" && input.Secret != h.callbackSecret {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	form, err := url.ParseQuery(string(input.RawBody))
	if err != nil {
		return nil, huma.Error400BadRequest("invalid form body")
	}

	raw := form.Get("payload")
	if raw == "" {
		return nil, huma.Error400BadRequest("missing payload field")
	}

	var payload callbackPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, huma.Error400BadRequest("invalid JSON in payload field")
	}

	switch payload.Event.Name {
	case "test":
		log.Info("POEditor test callback received")
	case "new_terms.added":
		go h.notifyNewTerms(payload)
	}

	return &CallbackOutput{}, nil
}

func (h *HandlerGroup) notifyNewTerms(payload callbackPayload) {
	if h.webhookURL == "" {
		log.Warn("DISCORD_WEBHOOK_URL not configured, skipping notification")
		return
	}

	h.mu.Lock()
	if time.Since(h.lastNotifiedAt) < notifyCooldown {
		h.mu.Unlock()
		log.Info("Skipping Discord notification (cooldown)", "remaining", notifyCooldown-time.Since(h.lastNotifiedAt))
		return
	}
	h.lastNotifiedAt = time.Now()
	h.mu.Unlock()

	msg := &discord.Message{
		Content: fmt.Sprintf(
			"<@&%s> New terms have been added to **%s** on POEditor and need translation.",
			translatorRoleID,
			payload.Project.Name,
		),
		AllowedMentions: &discord.AllowedMentions{
			Roles: &[]string{translatorRoleID},
		},
	}

	if _, err := discord.PostMessage(h.webhookURL, msg); err != nil {
		log.Error("Failed to send Discord webhook", "err", err)
	}
}
