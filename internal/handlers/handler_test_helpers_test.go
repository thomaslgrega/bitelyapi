package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thomaslgrega/bitelyapi/internal/auth"
	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

var testJWTManager = auth.NewJWTManager("test-secret", "bitelyapi-test", time.Hour)

// testImageLocator stands in for the one the repository holds, so a fake
// repository answers Recipe Images the way the real one does.
var testImageLocator = models.NewImageLocator("https://images.test")

// newTestRecipeHandler builds a handler for the cases that are not about
// Recipe Images. Its image store answers every call with an error, so a case
// that reaches it fails rather than passing on a silent zero value.
func newTestRecipeHandler(repo recipeRepository) *RecipeHandler {
	return NewRecipeHandler(repo, &fakeImageStore{
		headFunc: func(ctx context.Context, key string) (models.StagedImage, error) {
			return models.StagedImage{}, errors.New("unexpected Head")
		},
		promoteFunc: func(ctx context.Context, stagedKey string, key string) error {
			return errors.New("unexpected Promote")
		},
		deleteFunc: func(ctx context.Context, key string) error {
			return errors.New("unexpected Delete")
		},
	})
}

func authedRequest(t *testing.T, req *http.Request, next http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	token, err := testJWTManager.CreateToken("user-1")
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	middleware.AuthMiddleware(testJWTManager)(next).ServeHTTP(rec, req)
	return rec
}
