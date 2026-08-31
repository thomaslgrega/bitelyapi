package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func servedAs(t *testing.T, handler http.Handler, userID string) int {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/recipes/images", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec.Code
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimitPerUser(t *testing.T) {
	t.Run("refuses a user past the limit", func(t *testing.T) {
		clock := time.Now()
		handler := rateLimitPerUser(2, time.Minute, func() time.Time { return clock })(okHandler())

		if code := servedAs(t, handler, "user-1"); code != http.StatusOK {
			t.Fatalf("expected the first request to pass, got %d", code)
		}
		if code := servedAs(t, handler, "user-1"); code != http.StatusOK {
			t.Fatalf("expected the second request to pass, got %d", code)
		}
		if code := servedAs(t, handler, "user-1"); code != http.StatusTooManyRequests {
			t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, code)
		}
	})

	t.Run("budgets each user separately", func(t *testing.T) {
		clock := time.Now()
		handler := rateLimitPerUser(1, time.Minute, func() time.Time { return clock })(okHandler())

		servedAs(t, handler, "user-1")
		if code := servedAs(t, handler, "user-2"); code != http.StatusOK {
			t.Fatalf("expected another user to have its own budget, got %d", code)
		}
	})

	t.Run("refills once the window has passed", func(t *testing.T) {
		clock := time.Now()
		handler := rateLimitPerUser(1, time.Minute, func() time.Time { return clock })(okHandler())

		servedAs(t, handler, "user-1")
		if code := servedAs(t, handler, "user-1"); code != http.StatusTooManyRequests {
			t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, code)
		}

		clock = clock.Add(time.Minute)
		if code := servedAs(t, handler, "user-1"); code != http.StatusOK {
			t.Fatalf("expected a fresh window to pass, got %d", code)
		}
	})

	t.Run("refuses a request carrying no user", func(t *testing.T) {
		clock := time.Now()
		handler := rateLimitPerUser(1, time.Minute, func() time.Time { return clock })(okHandler())

		req := httptest.NewRequest(http.MethodPost, "/recipes/images", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
