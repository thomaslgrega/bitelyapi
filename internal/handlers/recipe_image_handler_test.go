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
	"strings"
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
		if !strings.HasPrefix(store.promotedTo, "recipes/"+createdID+"/") {
			t.Fatalf("expected promotion under the new recipe, got %q", store.promotedTo)
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
	stagedBody := func() *bytes.Buffer {
		return bytes.NewBufferString(`{"image_key":"` + stagedTestKey + `"}`)
	}
	imageRequest := func(body *bytes.Buffer) *http.Request {
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1/image", body)
		req.SetPathValue("id", "recipe-1")
		return req
	}

	t.Run("promotes the staged image and answers where it landed", func(t *testing.T) {
		store := &fakeImageStore{}
		var written string
		repo := fakeRecipeRepo{
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				store.calls = append(store.calls, "set-image")
				if recipeID != "recipe-1" || userID != "user-1" {
					t.Fatalf("expected the path recipe and the caller, got %q and %q", recipeID, userID)
				}
				written = key
				return testImageLocator.URLFor(key), "recipes/recipe-1/live.jpg", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body)
		}
		if want := []string{"head", "promote", "set-image", "delete", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if !strings.HasPrefix(written, "recipes/recipe-1/") {
			t.Fatalf("expected the row to hold a key derived from the recipe, got %q", written)
		}
		if written == "recipes/recipe-1/live.jpg" {
			t.Fatal("expected the replacement to land beside the live object, not on it")
		}
		// The superseded key is whatever the write reports it replaced, so a
		// write that landed in between is the one this cleans up after.
		if want := []string{stagedTestKey, "recipes/recipe-1/live.jpg"}; !reflect.DeepEqual(store.deletedKeys, want) {
			t.Fatalf("expected the staged and superseded objects cleaned up, got %v", store.deletedKeys)
		}

		var body struct {
			ImageURL string `json:"image_url"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode the response: %v", err)
		}
		if body.ImageURL != testImageLocator.URLFor(written) {
			t.Fatalf("expected the URL the promotion produced, got %q", body.ImageURL)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		store := &fakeImageStore{}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)
		rec := httptest.NewRecorder()

		h.UpdateRecipeImage(rec, imageRequest(stagedBody()))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		store := &fakeImageStore{}
		h := NewRecipeHandler(fakeRecipeRepo{}, store)

		rec := authedRequest(t, imageRequest(bytes.NewBufferString("{")), h.UpdateRecipeImage)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("refuses a key failing shape validation without touching the store", func(t *testing.T) {
		for _, key := range []string{"", "recipes/someone-elses-id/image.jpg", "incoming/not-a-uuid", "incoming/", "../incoming/x"} {
			store := &fakeImageStore{}
			repo := fakeRecipeRepo{
				setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, promoted string) (string, string, error) {
					t.Fatalf("expected no write for key %q", key)
					return "", "", nil
				},
			}
			h := NewRecipeHandler(repo, store)

			body := bytes.NewBufferString(`{"image_key":"` + key + `"}`)
			rec := authedRequest(t, imageRequest(body), h.UpdateRecipeImage)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected key %q to be refused with %d, got %d", key, http.StatusBadRequest, rec.Code)
			}
			if len(store.calls) != 0 {
				t.Fatalf("expected key %q to be refused before the store, got calls %v", key, store.calls)
			}
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
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				t.Fatal("expected no write")
				return "", "", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

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

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("reports a lookup failure as a server error", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "", errors.New("the database is down")
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("discards the promoted object when the row is gone", func(t *testing.T) {
		// The Recipe can be deleted between the authorship check and the
		// write, and the promoted key sits outside the staging prefix where no
		// lifecycle rule would reach it.
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				return "", "", sql.ErrNoRows
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != store.promotedTo {
			t.Fatalf("expected the promoted orphan to be dropped, got %v", store.deletedKeys)
		}
	})

	t.Run("leaves the live image alone when the row fails", func(t *testing.T) {
		// The copy lands on its own key, so a Recipe Image changes only when
		// the row does.
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				return "", "", errors.New("commit failed")
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
		if len(store.deletedKeys) != 1 || store.deletedKeys[0] != store.promotedTo {
			t.Fatalf("expected only the promoted orphan to be dropped, got %v", store.deletedKeys)
		}
	})

	t.Run("a store failure does not fail the write", func(t *testing.T) {
		store := &fakeImageStore{
			deleteFunc: func(ctx context.Context, key string) error {
				return errors.New("r2 is down")
			},
		}
		repo := fakeRecipeRepo{
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				return testImageLocator.URLFor(key), "recipes/recipe-1/live.jpg", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("never discards the object the row now names", func(t *testing.T) {
		// Nothing at the delete enforces that a promoted key is fresh per
		// upload, so the write's own answer is what the cleanup is checked
		// against.
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			setRecipeImageFunc: func(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
				return testImageLocator.URLFor(key), key, nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(stagedBody()), h.UpdateRecipeImage)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if want := []string{stagedTestKey}; !reflect.DeepEqual(store.deletedKeys, want) {
			t.Fatalf("expected only the staged object cleaned up, got %v", store.deletedKeys)
		}
	})
}

func TestDeleteRecipeImage(t *testing.T) {
	imageRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodDelete, "/recipes/recipe-1/image", nil)
		req.SetPathValue("id", "recipe-1")
		return req
	}

	t.Run("clears the row, then drops the object it named", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				store.calls = append(store.calls, "clear-image")
				if recipeID != "recipe-1" || userID != "user-1" {
					t.Fatalf("expected the path recipe and the caller, got %q and %q", recipeID, userID)
				}
				return "recipes/recipe-1/live.jpg", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body)
		}
		if want := []string{"clear-image", "delete"}; !reflect.DeepEqual(store.calls, want) {
			t.Fatalf("expected calls %v, got %v", want, store.calls)
		}
		if store.deletedKeys[0] != "recipes/recipe-1/live.jpg" {
			t.Fatalf("expected the stored object to be deleted, got %v", store.deletedKeys)
		}
	})

	t.Run("answers no content for a recipe that had no image", func(t *testing.T) {
		// A retried save must not fail on its second attempt (ADR-0006).
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				return "", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				t.Fatal("expected no write")
				return "", nil
			},
		}
		h := NewRecipeHandler(repo, store)
		rec := httptest.NewRecorder()

		h.DeleteRecipeImage(rec, imageRequest())

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("clears nothing on a recipe the caller does not author", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "user-2", nil
			},
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				t.Fatal("expected no write")
				return "", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("reports a lookup failure as a server error", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			getRecipeAuthorFunc: func(ctx context.Context, id string) (string, error) {
				return "", errors.New("the database is down")
			},
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				t.Fatal("expected no write")
				return "", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("answers not found for a recipe that is missing or someone else's", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				return "", sql.ErrNoRows
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
		if len(store.calls) != 0 {
			t.Fatalf("expected no store calls, got %v", store.calls)
		}
	})

	t.Run("reports a row failure as a server error", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				return "", errors.New("the database is down")
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("a store failure does not fail the delete", func(t *testing.T) {
		store := &fakeImageStore{
			deleteFunc: func(ctx context.Context, key string) error {
				return errors.New("r2 is down")
			},
		}
		repo := fakeRecipeRepo{
			clearRecipeImageFunc: func(ctx context.Context, recipeID string, userID string) (string, error) {
				return "recipes/recipe-1/live.jpg", nil
			},
		}
		h := NewRecipeHandler(repo, store)

		rec := authedRequest(t, imageRequest(), h.DeleteRecipeImage)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}

func TestDeleteRecipeDeletesTheImage(t *testing.T) {
	t.Run("deletes the object once the row is gone", func(t *testing.T) {
		store := &fakeImageStore{}
		repo := fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) (string, error) {
				store.calls = append(store.calls, "delete-row")
				return "recipes/recipe-1/live.jpg", nil
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
		if store.deletedKeys[0] != "recipes/recipe-1/live.jpg" {
			t.Fatalf("expected the stored object to be deleted, got %v", store.deletedKeys)
		}
	})

	t.Run("a store failure does not fail the delete", func(t *testing.T) {
		store := &fakeImageStore{
			deleteFunc: func(ctx context.Context, key string) error {
				return errors.New("r2 is down")
			},
		}
		repo := fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) (string, error) {
				return "recipes/recipe-1/live.jpg", nil
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
