package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

// imageStore is the bucket a Recipe Image lives in. It is declared here rather
// than in the package that implements it so a handler test can stand in for
// R2: the suite runs with neither database nor network.
type imageStore interface {
	PresignUpload(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error)
	Head(ctx context.Context, key string) (models.StagedImage, error)
	Promote(ctx context.Context, stagedKey string, key string) error
	Delete(ctx context.Context, key string) error
}

// errImageRejected marks the client's half of the image errors, so one
// promotion path can answer 400 for a bad claim ticket and 500 for a bucket
// that is down.
var errImageRejected = errors.New("image rejected")

type presignUploadRequest struct {
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
}

// PresignImageUpload mints one direct-to-R2 upload (ADR-0006).
func (h *RecipeHandler) PresignImageUpload(w http.ResponseWriter, r *http.Request) {
	if _, err := middleware.UserIDFromContext(r.Context()); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request presignUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// The type is pinned into the signature, so a client sending image/jpg
	// would upload to a 403 from R2 rather than a message it can act on.
	if !models.AllowedImageContentType(request.ContentType) {
		http.Error(w, "content_type must be image/jpeg or image/png", http.StatusBadRequest)
		return
	}

	// R2 refuses an upload that does not match its signed length, so signing a
	// length the share would refuse only buys the client a 403 (ADR-0006).
	if request.ContentLength < 1 || request.ContentLength > models.MaxImageBytes {
		http.Error(w, fmt.Sprintf("content_length must be between 1 and %d bytes", models.MaxImageBytes), http.StatusBadRequest)
		return
	}

	upload, err := h.images.PresignUpload(r.Context(), request.ContentType, request.ContentLength)
	if err != nil {
		http.Error(w, "failed to presign upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(upload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

type updateImageRequest struct {
	ImageKey string `json:"image_key"`
}

type imageResponse struct {
	ImageURL string `json:"image_url"`
}

// UpdateRecipeImage sets a Recipe Image and answers the URL the promotion
// produced (ADR-0006).
func (h *RecipeHandler) UpdateRecipeImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	var request updateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Shape first, so a malformed key costs no query.
	if !models.IsStagedImageKey(request.ImageKey) {
		writeImageError(w, errNotStaged())
		return
	}

	// Authorship is established before the copy rather than by the row write
	// that follows it, so a stranger's legally staged upload cannot reach the
	// Author's Recipe at all.
	if err := h.authorizeRecipe(r.Context(), id, userID); err != nil {
		writeRecipeLookupError(w, err)
		return
	}

	promoted, err := h.promoteImage(r.Context(), request.ImageKey, id)
	if err != nil {
		writeImageError(w, err)
		return
	}

	imageURL, superseded, err := h.repo.SetRecipeImage(r.Context(), id, userID, promoted)
	if err != nil {
		// The promotion landed on its own key, so the Recipe Image is still
		// whatever the row says — and if the row is gone, the copy belongs to
		// nothing. Either way it sits outside the staging prefix, where no
		// lifecycle rule would reach it.
		h.discardImage(r.Context(), promoted)
		writeImageRowError(w, err)
		return
	}

	h.discardImage(r.Context(), request.ImageKey)

	// The write reports the object the row stopped naming — read inside its own
	// transaction, so a write that landed since the authorship check is the one
	// this cleans up after. It is never the key just promoted, and the guard is
	// what keeps that true of a mint that stops being unique per upload.
	if superseded != "" && superseded != promoted {
		h.discardImage(r.Context(), superseded)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(imageResponse{ImageURL: imageURL}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteRecipeImage removes a Recipe Image, and answers 204 whether or not
// there was one (ADR-0006).
func (h *RecipeHandler) DeleteRecipeImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.authorizeRecipe(r.Context(), id, userID); err != nil {
		writeRecipeLookupError(w, err)
		return
	}

	// The clear names the object the row held, rather than a read before it
	// naming one a concurrent write may already have replaced.
	superseded, err := h.repo.ClearRecipeImage(r.Context(), id, userID)
	if err != nil {
		writeImageRowError(w, err)
		return
	}

	if superseded != "" {
		h.discardImage(r.Context(), superseded)
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeImageRowError keeps a database that could not answer distinct from a
// Recipe that is missing or belongs to someone else, which on these routes are
// the same 404.
func writeImageRowError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipe not found", http.StatusNotFound)
		return
	}

	http.Error(w, "failed to write the recipe image", http.StatusInternalServerError)
}

// promoteImage turns a client's claim ticket into the key a Recipe row holds.
// The HEAD is the only point anything about the bytes can be enforced
// (ADR-0006).
func (h *RecipeHandler) promoteImage(ctx context.Context, stagedKey string, recipeID string) (string, error) {
	if !models.IsStagedImageKey(stagedKey) {
		return "", errNotStaged()
	}

	staged, err := h.images.Head(ctx, stagedKey)
	if errors.Is(err, models.ErrImageNotFound) {
		return "", fmt.Errorf("%w: image_key names no upload", errImageRejected)
	}
	if err != nil {
		return "", err
	}
	if staged.ContentLength > models.MaxImageBytes {
		return "", fmt.Errorf("%w: image exceeds %d bytes", errImageRejected, models.MaxImageBytes)
	}
	if !models.AllowedImageContentType(staged.ContentType) {
		return "", fmt.Errorf("%w: image must be image/jpeg or image/png", errImageRejected)
	}

	// The destination is fresh per upload, so the object a Recipe currently
	// serves is untouched until its row names the new one.
	promoted, err := models.NewImageKey(recipeID)
	if err != nil {
		return "", err
	}
	if err := h.images.Promote(ctx, stagedKey, promoted); err != nil {
		return "", err
	}

	return promoted, nil
}

// errNotStaged is the refusal both write paths give a key the server would not
// have minted itself.
func errNotStaged() error {
	return fmt.Errorf("%w: image_key is not a staged upload", errImageRejected)
}

func writeImageError(w http.ResponseWriter, err error) {
	if errors.Is(err, errImageRejected) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Error(w, "failed to store image", http.StatusInternalServerError)
}

// authorizeRecipe refuses anyone but a Recipe's Author. A Recipe written by
// someone else answers the same "not found" a missing one does, so it tells a
// stranger nothing about what exists.
func (h *RecipeHandler) authorizeRecipe(ctx context.Context, recipeID string, userID string) error {
	author, err := h.repo.GetRecipeAuthor(ctx, recipeID)
	if err != nil {
		return err
	}
	if author != userID {
		return sql.ErrNoRows
	}

	return nil
}

// writeRecipeLookupError keeps a database that could not answer distinct from
// a Recipe the caller may not change.
func writeRecipeLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipe not found", http.StatusNotFound)
		return
	}

	http.Error(w, "failed to fetch recipe", http.StatusInternalServerError)
}

// discardImage drops an object the corpus no longer refers to. Failure is
// logged and swallowed, because the Recipe genuinely did change (ADR-0006).
func (h *RecipeHandler) discardImage(ctx context.Context, key string) {
	if err := h.images.Delete(ctx, key); err != nil {
		log.Printf("failed to delete image %q: %v", key, err)
	}
}
