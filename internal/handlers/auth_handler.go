package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

type AuthHandler struct {
	repo *repository.AuthRepository
	jwtManager *auth.JWTManager
}

func NewAuthHandler(repo *repository.AuthRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		jwtManager: jwtManager,
	}
}

func (h *AuthHandler) SignInWithApple(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IdentityToken string `json:"identity_token"`
		FirstName *string `json:"first_name,omitempty"`
		LastName *string `json:"last_name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyAppleIdentityToken(req.IdentityToken, "com.thomaslgrega.WhatWeEating")
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
		"user": user,
	})
}