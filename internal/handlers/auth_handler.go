package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

type AuthHandler struct {
	repo       *repository.AuthRepository
	jwtManager *auth.JWTManager
}

func NewAuthHandler(repo *repository.AuthRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (h *AuthHandler) SignInWithApple(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IdentityToken string  `json:"identity_token"`
		FirstName     *string `json:"first_name,omitempty"`
		LastName      *string `json:"last_name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyAppleIdentityToken(req.IdentityToken, "com.thomaslgrega.Bitely")
	if err != nil {
		http.Error(w, "invalid Apple token", http.StatusUnauthorized)
		return
	}

	appleSub := claims.Subject
	email := claims.Email
	firstName := req.FirstName
	lastName := req.LastName

	user, err := h.repo.FindOrCreateUser(r.Context(), appleSub, email, firstName, lastName)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.jwtManager.CreateToken(user.ID)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"user":         user,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	user, err := h.repo.CreateUserWithPassword(r.Context(), req.Email, hash)
	if err != nil {
		http.Error(w, "email already in use", http.StatusConflict)
		return
	}

	token, err := h.jwtManager.CreateToken(user.ID)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"user":         user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if user.PasswordHash == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(req.Password, *user.PasswordHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.CreateToken(user.ID)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"user":         user,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}