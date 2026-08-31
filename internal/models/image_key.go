package models

import (
	"strings"

	"github.com/google/uuid"
)

const (
	// stagedImagePrefix is the one prefix a client can upload into. Everything
	// under it is disposable: a lifecycle rule deletes what is abandoned or
	// rejected, which is why there is no reaper (ADR-0006).
	stagedImagePrefix = "incoming/"

	// MaxImageBytes caps a Recipe Image. R2 has no signing-time size limit, so
	// this is checked with a HEAD on the staged object (ADR-0006).
	MaxImageBytes = 5 << 20
)

// allowedImageContentTypes is the allowlist a signed Content-Type is pinned
// to. SVG is absent deliberately: served from a public bucket it is a stored
// XSS vector.
var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

func AllowedImageContentType(contentType string) bool {
	return allowedImageContentTypes[contentType]
}

// NewStagedImageKey mints the key a presigned upload is signed against.
func NewStagedImageKey() (string, error) {
	staged, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return stagedImagePrefix + staged.String(), nil
}

// IsStagedImageKey reports whether a client-supplied key is one this server
// would have minted. Without it a client could hand back
// `recipes/<someone-elses-id>/image.jpg` and have the share promote it.
func IsStagedImageKey(key string) bool {
	staged, found := strings.CutPrefix(key, stagedImagePrefix)
	if !found {
		return false
	}

	// Length is checked first because uuid.Parse also accepts the braced and
	// urn: forms, neither of which this server mints.
	return len(staged) == 36 && uuid.Validate(staged) == nil
}

// PromotedImageKey is where a staged image lands once its Recipe exists. The
// server derives it from the Recipe id so a client can name nothing outside
// the staging prefix (ADR-0006).
func PromotedImageKey(recipeID string) string {
	return "recipes/" + recipeID + "/image.jpg"
}
