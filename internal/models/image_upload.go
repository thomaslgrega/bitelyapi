package models

import (
	"errors"
	"time"
)

// ErrImageNotFound is what a store reports when a key names no object. It
// separates a client's stale claim ticket from a bucket that is unreachable,
// which are a 400 and a 500.
var ErrImageNotFound = errors.New("image not found")

// PresignedUpload is write capability into the bucket, handed to one client
// for one object (ADR-0006).
type PresignedUpload struct {
	UploadURL string    `json:"upload_url"`
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StagedImage is what an uploaded object turned out to be, as opposed to what
// the client declared when the upload was signed (ADR-0006).
type StagedImage struct {
	ContentType   string
	ContentLength int64
}
