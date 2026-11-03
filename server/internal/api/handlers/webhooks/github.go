package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
)

const refreshDebounce = 2 * time.Second

const pluginLabel = "plugin"

type FeedbackRefresher interface {
	RefreshFeedback(ctx context.Context) error
	ApplyStatus(pluginID, status string, add bool)
	ApplySimilar(pluginID, similarID string, add bool)
}

type Config struct {
	Secret     string
	Owner      string
	Repo       string
	Org        string
	Team       string
	OwnersTeam string
	Cache      FeedbackRefresher
	Moderator  Moderator
	Authors    PluginLookup
}

type HandlerGroup struct {
	secret     string
	owner      string
	repo       string
	org        string
	team       string
	ownersTeam string
	cache      FeedbackRefresher
	moderator  Moderator
	authors    PluginLookup

	mu          sync.Mutex
	lastRefresh time.Time
}

type WebhookInput struct {
	Event     string `header:"X-GitHub-Event"`
	Signature string `header:"X-Hub-Signature-256"`
	RawBody   []byte
}

type WebhookOutput struct{}

type label struct {
	Name string `json:"name"`
}

type eventPayload struct {
	Action string `json:"action"`
	Issue  struct {
		Number int     `json:"number"`
		Body   string  `json:"body"`
		Labels []label `json:"labels"`
	} `json:"issue"`
	Comment struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
}

func RegisterHandlers(cfg Config, grp *huma.Group) {
	h := &HandlerGroup{
		secret:     cfg.Secret,
		owner:      cfg.Owner,
		repo:       cfg.Repo,
		org:        cfg.Org,
		team:       cfg.Team,
		ownersTeam: cfg.OwnersTeam,
		cache:      cfg.Cache,
		moderator:  cfg.Moderator,
		authors:    cfg.Authors,
	}

	huma.Register(grp, huma.Operation{
		OperationID: "github-webhook",
		Summary:     "GitHub Webhook",
		Path:        "",
		Method:      http.MethodPost,
	}, h.Handle)
}

func (h *HandlerGroup) Handle(_ context.Context, input *WebhookInput) (*WebhookOutput, error) {
	if !h.validSignature(input.RawBody, input.Signature) {
		return nil, huma.Error401Unauthorized("invalid signature")
	}

	var payload eventPayload
	if err := json.Unmarshal(input.RawBody, &payload); err != nil {
		return nil, huma.Error400BadRequest("invalid payload")
	}

	switch input.Event {
	case "issues":
		if relevantAction(payload.Action) {
			h.triggerRefresh()
		}
	case "issue_comment":
		if payload.Action == "created" {
			h.handleComment(payload)
		}
	}

	return &WebhookOutput{}, nil
}

func (h *HandlerGroup) validSignature(body []byte, signature string) bool {
	if h.secret == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (h *HandlerGroup) triggerRefresh() {
	h.mu.Lock()
	if time.Since(h.lastRefresh) < refreshDebounce {
		h.mu.Unlock()
		return
	}
	h.lastRefresh = time.Now()
	h.mu.Unlock()

	go func() {
		if err := h.cache.RefreshFeedback(context.Background()); err != nil {
			log.Error("Webhook feedback refresh failed", "err", err)
		}
	}()
}

func relevantAction(action string) bool {
	switch action {
	case "opened", "closed", "reopened", "labeled", "unlabeled", "deleted":
		return true
	default:
		return false
	}
}
