package handlers

import (
	"context"
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

// PresignImageUpload mints one direct-to-R2 upload. It is the only endpoint
// that hands out write capability into the bucket, which is why it sits behind
// auth and a per-user rate limit.
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

	// The signed length is not the enforcement point — the HEAD at share time
	// is — but signing a length the share would refuse only buys the client a
	// wasted upload.
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

// promoteImage turns a client's claim ticket into the key a Recipe row holds.
// The staged object is measured here because R2 has no signing-time size cap,
// making this the only point anything about the bytes can be enforced
// (ADR-0006).
func (h *RecipeHandler) promoteImage(ctx context.Context, stagedKey string, recipeID string) (string, error) {
	if !models.IsStagedImageKey(stagedKey) {
		return "", fmt.Errorf("%w: image_key is not a staged upload", errImageRejected)
	}

	staged, err := h.images.Head(ctx, stagedKey)
	if err != nil {
		return "", fmt.Errorf("%w: image_key names no upload", errImageRejected)
	}
	if staged.ContentLength > models.MaxImageBytes {
		return "", fmt.Errorf("%w: image exceeds %d bytes", errImageRejected, models.MaxImageBytes)
	}
	if !models.AllowedImageContentType(staged.ContentType) {
		return "", fmt.Errorf("%w: image must be image/jpeg or image/png", errImageRejected)
	}

	promoted := models.PromotedImageKey(recipeID)
	if err := h.images.Promote(ctx, stagedKey, promoted); err != nil {
		return "", err
	}

	return promoted, nil
}

// writeImageError answers a client's bad claim ticket with 400 and a bucket
// failure with 500.
func writeImageError(w http.ResponseWriter, err error) {
	if errors.Is(err, errImageRejected) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Error(w, "failed to store image", http.StatusInternalServerError)
}

// discardImage drops an object the corpus no longer refers to. Failure is
// logged and swallowed: the row has already changed, and the request describes
// the Recipe rather than the bucket (ADR-0006).
func (h *RecipeHandler) discardImage(ctx context.Context, key string) {
	if err := h.images.Delete(ctx, key); err != nil {
		log.Printf("failed to delete image %q: %v", key, err)
	}
}
