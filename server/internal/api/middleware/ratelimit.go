package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/duxlabs-dev/dux-docs/server/internal/utils"
)

type rateLimitEntry struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiter struct {
	mu         sync.RWMutex
	entries    map[string]*rateLimitEntry
	rate       float64
	burst      float64
	cleanupInt time.Duration
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		entries:    make(map[string]*rateLimitEntry),
		rate:       requestsPerSecond,
		burst:      float64(burst),
		cleanupInt: 5 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInt)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanupInt)
		for ip, entry := range rl.entries {
			if entry.lastCheck.Before(cutoff) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]

	if !exists {
		rl.entries[ip] = &rateLimitEntry{
			tokens:    rl.burst - 1,
			lastCheck: now,
		}
		return true
	}

	elapsed := now.Sub(entry.lastCheck).Seconds()
	entry.tokens += elapsed * rl.rate
	if entry.tokens > rl.burst {
		entry.tokens = rl.burst
	}
	entry.lastCheck = now

	if entry.tokens < 1 {
		return false
	}

	entry.tokens--
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetIPAddress(r)

		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","retry_after":1}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) HumaMiddleware(ctx huma.Context, next func(huma.Context)) {
	ip := utils.GetIPAddressFromHeaders(ctx)

	if !rl.Allow(ip) {
		ctx.SetHeader("Content-Type", "application/json")
		ctx.SetHeader("Retry-After", "1")
		ctx.SetStatus(http.StatusTooManyRequests)
		ctx.BodyWriter().Write([]byte(`{"error":"rate limit exceeded","retry_after":1}`))
		return
	}

	next(ctx)
}
