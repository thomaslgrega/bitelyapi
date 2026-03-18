package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
)

func TestAuthMiddleware(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", "bitelyapi-test", time.Hour)
	validToken, err := jwtManager.CreateToken("user-1")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	t.Run("missing header", func(t *testing.T) {
		handler := AuthMiddleware(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("invalid header format", func(t *testing.T) {
		handler := AuthMiddleware(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token "+validToken)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		handler := AuthMiddleware(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("passes user id to next handler", func(t *testing.T) {
		handler := AuthMiddleware(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := UserIDFromContext(r.Context())
			if err != nil {
				t.Fatalf("expected user id in context: %v", err)
			}
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}
			w.WriteHeader(http.StatusNoContent)
		}))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
