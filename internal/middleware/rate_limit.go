package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimitPerUser caps how often one authenticated user may reach the handler
// it wraps, in fixed windows. It reads the user id the auth middleware put in
// the context, so it only composes inside that one.
func RateLimitPerUser(limit int, window time.Duration) func(http.Handler) http.Handler {
	return rateLimitPerUser(limit, window, time.Now)
}

func rateLimitPerUser(limit int, window time.Duration, now func() time.Time) func(http.Handler) http.Handler {
	limiter := &userRateLimiter{
		limit:   limit,
		window:  window,
		now:     now,
		windows: make(map[string]*userWindow),
		sweptAt: now(),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := UserIDFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !limiter.allow(userID) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type userWindow struct {
	startedAt time.Time
	count     int
}

type userRateLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu      sync.Mutex
	windows map[string]*userWindow
	sweptAt time.Time
}

func (l *userRateLimiter) allow(userID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	current, seen := l.windows[userID]
	if !seen || now.Sub(current.startedAt) >= l.window {
		current = &userWindow{startedAt: now}
		l.windows[userID] = current
	}

	current.count++

	return current.count <= l.limit
}

// sweep drops the windows that have expired, so the map tracks the users
// currently spending their budget rather than every user ever seen. Once per
// window is enough: a stale entry costs nothing but memory, and sweeping per
// request would walk every user on every call.
func (l *userRateLimiter) sweep(now time.Time) {
	if now.Sub(l.sweptAt) < l.window {
		return
	}

	for userID, tracked := range l.windows {
		if now.Sub(tracked.startedAt) >= l.window {
			delete(l.windows, userID)
		}
	}
	l.sweptAt = now
}
