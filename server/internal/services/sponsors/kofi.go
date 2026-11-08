package sponsors

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// KofiPayload is the subset of a Ko-fi webhook this service reads; email,
// amount and message are deliberately never parsed.
type KofiPayload struct {
	VerificationToken     string `json:"verification_token"`
	MessageID             string `json:"message_id"`
	Timestamp             string `json:"timestamp"`
	Type                  string `json:"type"`
	IsPublic              bool   `json:"is_public"`
	FromName              string `json:"from_name"`
	IsSubscriptionPayment bool   `json:"is_subscription_payment"`
	KofiTransactionID     string `json:"kofi_transaction_id"`
}

type kofiSupporter struct {
	Name        string    `json:"name"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastPayment time.Time `json:"lastPayment"`
}

type kofiState struct {
	// Keyed by the lowercased display name.
	Supporters map[string]*kofiSupporter `json:"supporters"`
	// When each supporter last paid privately. A public payment older than this
	// (a delayed Ko-fi retry) must not put them back on the page.
	Private map[string]time.Time `json:"private,omitempty"`
	// Recent transaction ids, so Ko-fi's retries are not counted twice.
	Seen []string `json:"seen"`
}

const (
	maxSeenTransactions = 1000
	maxNameRunes        = 64
	// A month plus slack, so monthly members stay current between payments.
	kofiActiveWindow = 35 * 24 * time.Hour
)

func newKofiState() kofiState {
	return kofiState{Supporters: map[string]*kofiSupporter{}, Private: map[string]time.Time{}}
}

// apply folds one webhook into the state and reports whether anything changed.
func (s *kofiState) apply(p KofiPayload, now time.Time) bool {
	switch p.Type {
	case "Tip", "Subscription":
	default:
		// Shop orders and commissions are purchases, not sponsorships.
		return false
	}

	// Ko-fi retries with the same message_id.
	txID := p.MessageID
	if txID == "" {
		txID = p.KofiTransactionID
	}
	if txID != "" {
		if slices.Contains(s.Seen, txID) {
			return false
		}
		s.Seen = append(s.Seen, txID)
		if len(s.Seen) > maxSeenTransactions {
			s.Seen = s.Seen[len(s.Seen)-maxSeenTransactions:]
		}
	}

	name := cleanName(p.FromName)
	if name == "" {
		return txID != ""
	}
	key := strings.ToLower(name)

	at, err := time.Parse(time.RFC3339, p.Timestamp)
	if err != nil || at.After(now) {
		at = now
	}

	// The most recent payment's choice wins: a supporter who goes private is
	// taken off the page until they pay publicly again.
	if !p.IsPublic {
		if at.After(s.Private[key]) {
			s.Private[key] = at
		}
		delete(s.Supporters, key)
		return true
	}
	if private, ok := s.Private[key]; ok {
		if !at.After(private) {
			return txID != ""
		}
		delete(s.Private, key)
	}

	existing, ok := s.Supporters[key]
	if !ok {
		s.Supporters[key] = &kofiSupporter{Name: name, FirstSeen: at, LastPayment: at}
		return true
	}
	existing.Name = name
	if at.After(existing.LastPayment) {
		existing.LastPayment = at
	}
	if at.Before(existing.FirstSeen) {
		existing.FirstSeen = at
	}
	return true
}

func cleanName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	if utf8.RuneCountInString(name) > maxNameRunes {
		name = string([]rune(name)[:maxNameRunes])
	}
	return strings.TrimSpace(name)
}
