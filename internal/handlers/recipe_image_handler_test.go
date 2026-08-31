package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

// fakeImageStore records what the handler asks of R2 and in what order, which
// is the part worth asserting: signed URL strings are the store's business.
type fakeImageStore struct {
	calls []string

	presignFunc func(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error)
	headFunc    func(ctx context.Context, key string) (models.StagedImage, error)
	promoteFunc func(ctx context.Context, stagedKey string, key string) error
	deleteFunc  func(ctx context.Context, key string) error

	promotedFrom string
	promotedTo   string
	deletedKeys  []string
}

func (f *fakeImageStore) PresignUpload(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error) {
	f.calls = append(f.calls, "presign")
	if f.presignFunc == nil {
		return models.PresignedUpload{}, errors.New("unexpected PresignUpload")
	}
	return f.presignFunc(ctx, contentType, contentLength)
}

func (f *fakeImageStore) Head(ctx context.Context, key string) (models.StagedImage, error) {
	f.calls = append(f.calls, "head")
	if f.headFunc == nil {
		return models.StagedImage{ContentType: "image/jpeg", ContentLength: 1024}, nil
	}
	return f.headFunc(ctx, key)
}

func (f *fakeImageStore) Promote(ctx context.Context, stagedKey string, key string) error {
	f.calls = append(f.calls, "promote")
	f.promotedFrom, f.promotedTo = stagedKey, key
	if f.promoteFunc == nil {
		return nil
	}
	return f.promoteFunc(ctx, stagedKey, key)
}

func (f *fakeImageStore) Delete(ctx context.Context, key string) error {
	f.calls = append(f.calls, "delete")
	f.deletedKeys = append(f.deletedKeys, key)
	if f.deleteFunc == nil {
		return nil
	}
	return f.deleteFunc(ctx, key)
}

const stagedTestKey = "incoming/9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60"

func TestPresignImageUpload(t *testing.T) {
	t.Run("requires auth", func(t *testing.T) {
		store := &fakeImageStore{}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)
		req := httptest.NewRequest(http.MethodPost, "/recipes/images", bytes.NewBufferString(`{"content_type":"image/jpeg","content_length":1024}`))
		rec := httptest.NewRecorder()

		h.PresignImageUpload(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("passes the content type through to the store", func(t *testing.T) {
		expiry := time.Now().Add(5 * time.Minute).UTC().Truncate(time.Second)
		store := &fakeImageStore{
			presignFunc: func(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error) {
				if contentType != "image/png" {
					t.Fatalf("expected the client's content type, got %q", contentType)
				}
				if contentLength != 2048 {
					t.Fatalf("expected the client's content length, got %d", contentLength)
				}
				return models.PresignedUpload{UploadURL: "https://account.r2.cloudflarestorage.com/signed", Key: stagedTestKey, ExpiresAt: expiry}, nil
			},
		}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)
		req := httptest.NewRequest(http.MethodPost, "/recipes/images", bytes.NewBufferString(`{"content_type":"image/png","content_length":2048}`))
		rec := authedRequest(t, req, h.PresignImageUpload)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body)
		}

		var upload models.PresignedUpload
		if err := json.Unmarshal(rec.Body.Bytes(), &upload); err != nil {
			t.Fatalf("failed to decode the upload: %v", err)
		}
		if upload.UploadURL == "" || upload.Key != stagedTestKey {
			t.Fatalf("unexpected upload %#v", upload)
		}
		if !upload.ExpiresAt.Equal(expiry) {
			t.Fatalf("expected the store's expiry, got %s", upload.ExpiresAt)
		}
	})

	t.Run("refuses a content type outside the allowlist before reaching the store", func(t *testing.T) {
		for _, contentType := range []string{"image/gif", "image/svg+xml", "image/heic", "image/jpg", ""} {
			store := &fakeImageStore{}
			h := NewRecipeHandler(fakeRecipeRepo{}, store)
			body := `{"content_type":"` + contentType + `","content_length":1024}`
			req := httptest.NewRequest(http.MethodPost, "/recipes/images", bytes.NewBufferString(body))
			rec := authedRequest(t, req, h.PresignImageUpload)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected %q to be refused with %d, got %d", contentType, http.StatusBadRequest, rec.Code)
			}
			if len(store.calls) != 0 {
				t.Fatalf("expected %q to be refused before the store, got calls %v", contentType, store.calls)
			}
		}
	})

	t.Run("refuses a length the share would reject anyway", func(t *testing.T) {
		for _, length := range []int64{0, -1, models.MaxImageBytes + 1} {
			store := &fakeImageStore{}
			h := NewRecipeHandler(fakeRecipeRepo{}, store)
			body := `{"content_type":"image/jpeg","content_length":` + strconv.FormatInt(length, 10) + `}`
			req := httptest.NewRequest(http.MethodPost, "/recipes/images", bytes.NewBufferString(body))
			rec := authedRequest(t, req, h.PresignImageUpload)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected length %d to be refused, got %d", length, rec.Code)
			}
			if len(store.calls) != 0 {
				t.Fatalf("expected length %d to be refused before the store, got calls %v", length, store.calls)
			}
		}
	})

	t.Run("reports a store failure as a server error", func(t *testing.T) {
		store := &fakeImageStore{
			presignFunc: func(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error) {
				return models.PresignedUpload{}, errors.New("r2 is down")
			},
		}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)
		req := httptest.NewRequest(http.MethodPost, "/recipes/images", bytes.NewBufferString(`{"content_type":"image/jpeg","content_length":1024}`))
		rec := authedRequest(t, req, h.PresignImageUpload)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestCreateRecipePromotesTheStagedImage(t *testing.T) {
	t.Run("heads, promotes to a server-derived key, then writes the row", func(t *testing.T) {
		store := &fakeImageStore{}
		var created models.CreateRecipeInput
		var createdID string
		repo := fakeRecipeRepo{
			createRecipeFunc: func(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
				store.calls = append(store.calls, "create")
				created, createdID = input, recipeID
				return &models.Recipe{ID: recipeID, UserID: userID, Name: input.Name, Category: input.Category}, nil
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body)
		}
		if want := []string{"head", "promote", "create", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if store.promotedFrom != stagedTestKey {
			t.Fatalf("expected the staged key to be promoted, got %q", store.promotedFrom)
		}
		if want := models.PromotedImageKey(createdID); store.promotedTo != want {
			t.Fatalf("expected promotion to %q, got %q", want, store.promotedTo)
		}
		if created.ImageKey != store.promotedTo {
			t.Fatalf("expected the row to hold the promoted key, got %q", created.ImageKey)
		}
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != stagedTestKey {
			t.Fatalf("expected the staged object to be cleaned up, got %v", store.deletedKeys)
		}
	})

	t.Run("refuses a key failing shape validation without touching the store", func(t *testing.T) {
		for _, key := range []string{"recipes/someone-elses-id/image.jpg", "incoming/not-a-uuid", "incoming/", "../incoming/x"} {
			store := &fakeImageStore{}
			repo := fakeRecipeRepo{
				createRecipeFunc: func(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
					t.Fatalf("expected no recipe to be created for key %q", key)
					return nil, nil
				},
			}
			h := NewRecipeHandler(repo, store)

			body := `{"name":"Shakshuka","category":"dinner","image_key":"` + key + `"}`
			req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
			rec := authedRequest(t, req, h.CreateRecipe)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected key %q to be refused with %d, got %d", key, http.StatusBadRequest, rec.Code)
			}
			if len(store.calls) != 0 {
				t.Fatalf("expected key %q to be refused before the store, got calls %v", key, store.calls)
			}
		}
	})

	t.Run("refuses an object the head reports as too large or the wrong type", func(t *testing.T) {
		staged := []models.StagedImage{
			{ContentType: "image/jpeg", ContentLength: 6 << 20},
			{ContentType: "image/gif", ContentLength: 1024},
		}

		for _, object := range staged {
			store := &fakeImageStore{
				headFunc: func(ctx context.Context, key string) (models.StagedImage, error) {
					return object, nil
				},
			}
			repo := fakeRecipeRepo{
				createRecipeFunc: func(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
					t.Fatalf("expected no recipe to be created for %#v", object)
					return nil, nil
				},
			}
			h := NewRecipeHandler(repo, store)

			body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
			req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
			rec := authedRequest(t, req, h.CreateRecipe)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected %#v to be refused with %d, got %d", object, http.StatusBadRequest, rec.Code)
			}
			if want := []string{"head"}; !reflect.DeepEqual(store.calls, want) {
				t.Fatalf("expected calls %v, got %v", want, store.calls)
			}
		}
	})

	t.Run("does not touch the store for a recipe with no image", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			createRecipeFunc: func(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
				return &models.Recipe{ID: recipeID, Name: input.Name, Category: input.Category}, nil
			},
		}
		h := NewRecipeHandler(repo, store)

		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(`{"name":"Soup","category":"dinner"}`))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("drops the promoted object when the row fails", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			createRecipeFunc: func(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
				return nil, errors.New("insert failed")
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
		// The promoted key sits outside the staging prefix, so no lifecycle
		// rule would ever reap it; the staged one is left to that rule.
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != store.promotedTo {
			t.Fatalf("expected the promoted object to be dropped, got %v", store.deletedKeys)
		}
	})

	t.Run("reports a bucket failure as a server error", func(t *testing.T) {
		store := &fakeImageStore{
			headFunc: func(ctx context.Context, key string) (models.StagedImage, error) {
				return models.StagedImage{}, errors.New("r2 is down")
			},
		}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("reports an upload that is not there as a client error", func(t *testing.T) {
		store := &fakeImageStore{
			headFunc: func(ctx context.Context, key string) (models.StagedImage, error) {
				return models.StagedImage{}, models.ErrImageNotFound
			},
		}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestUpdateRecipeImage(t *testing.T) {
	t.Run("promotes a newly staged image before the row commits", func(t *testing.T) {
		store := &fakeImageStore{}
		var updated models.Recipe
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "user-1", nil
			},
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				store.calls = append(store.calls, "update")
				updated = recipe
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(body))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body)
		}
		if want := []string{"head", "promote", "update", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if want := models.PromotedImageKey("recipe-1"); updated.ImageKey != want {
			t.Fatalf("expected the row to hold %q, got %q", want, updated.ImageKey)
		}
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != stagedTestKey {
			t.Fatalf("expected the staged object to be cleaned up, got %v", store.deletedKeys)
		}
	})

	t.Run("deletes the superseded object when the recipe keeps no image", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				store.calls = append(store.calls, "update")
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(`{"name":"Shakshuka","category":"dinner"}`))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
		if want := []string{"update", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if want := models.PromotedImageKey("recipe-1"); store.deletedKeys[0] != want {
			t.Fatalf("expected %q to be deleted, got %v", want, store.deletedKeys)
		}
	})

	t.Run("a store failure does not fail the update", func(t *testing.T) {
		store := &fakeImageStore{
			deleteFunc: func(ctx context.Context, key string) error {
				return errors.New("r2 is down")
			},
		}
		repo := fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(`{"name":"Shakshuka","category":"dinner"}`))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("promotes nothing into a recipe the caller does not author", func(t *testing.T) {
		// The promoted key is derived from the path id, so promoting before
		// ownership is known would let a stranger overwrite the Author's
		// object with a legally staged upload of their own.
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "user-2", nil
			},
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				t.Fatal("expected no update")
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(body))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("promotes nothing into a recipe that does not exist", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "", sql.ErrNoRows
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"` + stagedTestKey + `"}`
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(body))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("refuses a key failing shape validation without touching the store", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				t.Fatal("expected no update")
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		body := `{"name":"Shakshuka","category":"dinner","image_key":"recipes/someone-elses-id/image.jpg"}`
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(body))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})
}

func TestDeleteRecipeDeletesTheImage(t *testing.T) {
	t.Run("deletes the object once the row is gone", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) error {
				store.calls = append(store.calls, "delete-row")
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		req := httptest.NewRequest(http.MethodDelete, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.DeleteRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
		if want := []string{"delete-row", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if want := models.PromotedImageKey("recipe-1"); store.deletedKeys[0] != want {
			t.Fatalf("expected %q to be deleted, got %v", want, store.deletedKeys)
		}
	})

	t.Run("a store failure does not fail the delete", func(t *testing.T) {
		store := &fakeImageStore{
			deleteFunc: func(ctx context.Context, key string) error {
				return errors.New("r2 is down")
			},
		}
		repo := fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) error {
				return nil
			},
		}
		h := NewRecipeHandler(repo, store)

		req := httptest.NewRequest(http.MethodDelete, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.DeleteRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
