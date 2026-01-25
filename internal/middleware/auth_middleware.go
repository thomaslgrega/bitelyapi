package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
)

type contextKey string

const userIDKey contextKey = "user_id"

func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := jwtManager.VerifyToken(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			userId := claims.Subject
			if userId == "" {
				http.Error(w, "invalid token subject", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(userIDKey)
	userID, ok := v.(string)
	if !ok || userID == "" {
		return "", errors.New("missing user id in context")
	}

	return userID, nil
}