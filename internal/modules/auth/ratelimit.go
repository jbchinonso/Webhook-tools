package auth

import (
	"sync"
	"time"
)

// limiter is a tiny in-memory token bucket keyed by a string (IP, email, or
// both). It is a stopgap for single-instance deployments. Swap for a Redis
// or DB-backed limiter when you scale out — the interface is deliberately
// narrow.
type limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int           // tokens per window
	window  time.Duration
}

type bucket struct {
	tokens   int
	resetAt  time.Time
}

func newLimiter(rate int, window time.Duration) *limiter {
	return &limiter{
		buckets: map[string]*bucket{},
		rate:    rate,
		window:  window,
	}
}

// Allow reports whether the key is within its budget. It refills lazily:
// when a bucket's window has passed, it resets.
func (l *limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{tokens: l.rate - 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}