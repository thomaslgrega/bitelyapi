package models

import "time"

// PresignedUpload is write capability into the bucket, handed to one client
// for one object: where to PUT the bytes, the key to name when the Recipe is
// shared, and when the URL stops working. The API never sees the bytes
// (ADR-0006).
type PresignedUpload struct {
	UploadURL string    `json:"upload_url"`
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StagedImage is what an uploaded object turned out to be, as opposed to what
// the client declared when the upload was signed. R2 caps neither size nor
// type at signing time, so this is what the share checks (ADR-0006).
type StagedImage struct {
	ContentType   string
	ContentLength int64
}
