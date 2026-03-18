package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type fakeAuthRepo struct {
	findOrCreateUserFunc       func(ctx context.Context, appleSub string, email *string, firstName *string, lastName *string) (models.User, error)
	createUserWithPasswordFunc func(ctx context.Context, email string, passwordHash string) (models.User, error)
	getUserByEmailFunc         func(ctx context.Context, email string) (models.User, error)
	getUserByIDFunc            func(ctx context.Context, id string) (models.User, error)
}

func (f fakeAuthRepo) FindOrCreateUser(ctx context.Context, appleSub string, email *string, firstName *string, lastName *string) (models.User, error) {
	return f.findOrCreateUserFunc(ctx, appleSub, email, firstName, lastName)
}

func (f fakeAuthRepo) CreateUserWithPassword(ctx context.Context, email string, passwordHash string) (models.User, error) {
	return f.createUserWithPasswordFunc(ctx, email, passwordHash)
}

func (f fakeAuthRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return f.getUserByEmailFunc(ctx, email)
}

func (f fakeAuthRepo) GetUserByID(ctx context.Context, id string) (models.User, error) {
	return f.getUserByIDFunc(ctx, id)
}

func newTestAuthHandler(repo fakeAuthRepo) *AuthHandler {
	return NewAuthHandler(repo, auth.NewJWTManager("test-secret", "bitelyapi-test", time.Hour))
}

func TestAuthHandlerRegister(t *testing.T) {
	t.Run("rejects invalid json", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("{"))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("rejects blank email after trim", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"   ","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("maps repository error to conflict", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{
			createUserWithPasswordFunc: func(ctx context.Context, email string, passwordHash string) (models.User, error) {
				if email != "user@example.com" {
					t.Fatalf("expected trimmed email to be used, got %q", email)
				}
				if passwordHash == "" || passwordHash == "secret" {
					t.Fatalf("expected password hash to be generated")
				}
				return models.User{}, errors.New("duplicate")
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":" user@example.com ","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}
	})

	t.Run("returns token and user", func(t *testing.T) {
		email := "user@example.com"
		h := newTestAuthHandler(fakeAuthRepo{
			createUserWithPasswordFunc: func(ctx context.Context, emailArg string, passwordHash string) (models.User, error) {
				return models.User{ID: "user-1", Email: &email}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected application/json content type, got %q", got)
		}

		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body["access_token"] == "" {
			t.Fatalf("expected access token in response")
		}
		userMap, ok := body["user"].(map[string]any)
		if !ok {
			t.Fatalf("expected user object in response")
		}
		if userMap["id"] != "user-1" {
			t.Fatalf("expected user id user-1, got %#v", userMap["id"])
		}
	})
}

func TestAuthHandlerLogin(t *testing.T) {
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	t.Run("rejects invalid json", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{"))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":""}`))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("rejects when repository misses user", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (models.User, error) {
				return models.User{}, sql.ErrNoRows
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("rejects users without password auth", func(t *testing.T) {
		email := "apple@example.com"
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByEmailFunc: func(ctx context.Context, emailArg string) (models.User, error) {
				return models.User{ID: "user-1", Email: &email}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"apple@example.com","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		email := "user@example.com"
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByEmailFunc: func(ctx context.Context, emailArg string) (models.User, error) {
				return models.User{ID: "user-1", Email: &email, PasswordHash: &hash}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong"}`))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns token and user on success", func(t *testing.T) {
		email := "user@example.com"
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByEmailFunc: func(ctx context.Context, emailArg string) (models.User, error) {
				if emailArg != email {
					t.Fatalf("expected trimmed email %q, got %q", email, emailArg)
				}
				return models.User{ID: "user-1", Email: &email, PasswordHash: &hash}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":" user@example.com ","password":"secret"}`))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body["access_token"] == "" {
			t.Fatalf("expected access token in response")
		}
	})
}

func TestAuthHandlerMe(t *testing.T) {
	t.Run("requires auth context", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns not found for missing user", func(t *testing.T) {
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByIDFunc: func(ctx context.Context, id string) (models.User, error) {
				if id != "user-1" {
					t.Fatalf("expected user-1, got %q", id)
				}
				return models.User{}, sql.ErrNoRows
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := authedRequest(t, req, h.Me)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns user payload", func(t *testing.T) {
		email := "user@example.com"
		h := newTestAuthHandler(fakeAuthRepo{
			getUserByIDFunc: func(ctx context.Context, id string) (models.User, error) {
				if id != "user-1" {
					t.Fatalf("expected user id user-1, got %q", id)
				}
				return models.User{ID: "user-1", Email: &email}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := authedRequest(t, req, h.Me)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
