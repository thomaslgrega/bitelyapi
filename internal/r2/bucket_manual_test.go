//go:build manual

// This file answers the two questions ADR-0006 could not settle from
// Cloudflare's documentation. It talks to a real bucket, so it is behind a
// build tag and never runs in `go test ./...`.
//
//	R2_ENV_FILE=/path/to/.env go test -tags manual -v ./internal/r2/
package r2

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

func liveStore(t *testing.T) *Store {
	t.Helper()

	if envFile := os.Getenv("R2_ENV_FILE"); envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			t.Fatalf("failed to read %s: %v", envFile, err)
		}
	}

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("%v (point R2_ENV_FILE at the .env the wizard wrote)", err)
	}

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("failed to build the store: %v", err)
	}

	return store
}

// upload PUTs body through a presigned URL, declaring declaredLength as the
// signed content length. It answers the status R2 gave.
func upload(t *testing.T, upload models.PresignedUpload, contentType string, declaredLength int64, body []byte) int {
	t.Helper()

	request, err := http.NewRequest(http.MethodPut, upload.UploadURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build the upload request: %v", err)
	}
	request.Header.Set("Content-Type", contentType)
	request.ContentLength = declaredLength
	request.Header.Set("Content-Length", strconv.FormatInt(declaredLength, 10))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("the upload never reached R2: %v", err)
	}
	defer response.Body.Close()

	return response.StatusCode
}

// TestLiveSignedContentLengthIsEnforced asks question one: does R2 reject an
// upload whose size differs from the signed Content-Length? Cloudflare
// documents the content-type case and only that case.
func TestLiveSignedContentLengthIsEnforced(t *testing.T) {
	store := liveStore(t)
	ctx := context.Background()

	honest := bytes.Repeat([]byte{0x41}, 1024)
	signed, err := store.PresignUpload(ctx, "image/jpeg", int64(len(honest)))
	if err != nil {
		t.Fatalf("failed to presign: %v", err)
	}
	t.Cleanup(func() { store.Delete(ctx, signed.Key) })

	if status := upload(t, signed, "image/jpeg", int64(len(honest)), honest); status != http.StatusOK {
		t.Fatalf("an upload matching its signature was refused with %d", status)
	}
	t.Log("an upload matching its signed length is accepted")

	// The real question. A client cannot send more bytes than it declares —
	// the declared length *is* the body over HTTP/1.1 — so an oversized upload
	// means declaring the true, larger length against a signature minted for a
	// small one. content-length is in X-Amz-SignedHeaders, so the question is
	// whether R2 verifies it.
	oversized := bytes.Repeat([]byte{0x41}, 4096)
	lying, err := store.PresignUpload(ctx, "image/jpeg", 1024)
	if err != nil {
		t.Fatalf("failed to presign: %v", err)
	}
	t.Cleanup(func() { store.Delete(ctx, lying.Key) })

	status := upload(t, lying, "image/jpeg", int64(len(oversized)), oversized)
	t.Logf("ANSWER: %d bytes declared honestly against a signature for 1024 answered %d", len(oversized), status)

	staged, err := store.Head(ctx, lying.Key)
	switch {
	case errors.Is(err, models.ErrImageNotFound):
		t.Log("ANSWER: R2 stored nothing, so the signed Content-Length is enforced")
	case err != nil:
		t.Fatalf("the HEAD failed: %v", err)
	default:
		t.Logf("ANSWER: R2 stored %d bytes, so the signed Content-Length is NOT a cap", staged.ContentLength)
	}
}

// TestLiveHeadAndCopyPath asks question two: does the direct HEAD/CopyObject
// path hit the checksum 501 that affected non-presigned PutObject? It appears
// fixed server-side, and the copy path is where it would show.
func TestLiveHeadAndCopyPath(t *testing.T) {
	store := liveStore(t)
	ctx := context.Background()

	body := bytes.Repeat([]byte{0x42}, 2048)
	signed, err := store.PresignUpload(ctx, "image/png", int64(len(body)))
	if err != nil {
		t.Fatalf("failed to presign: %v", err)
	}
	t.Cleanup(func() { store.Delete(ctx, signed.Key) })

	if status := upload(t, signed, "image/png", int64(len(body)), body); status != http.StatusOK {
		t.Fatalf("the upload was refused with %d", status)
	}

	staged, err := store.Head(ctx, signed.Key)
	if err != nil {
		t.Fatalf("ANSWER: the HEAD failed: %v", err)
	}
	t.Logf("ANSWER: HEAD reports %d bytes of %q", staged.ContentLength, staged.ContentType)

	promoted := models.PromotedImageKey("manual-verification")
	t.Cleanup(func() { store.Delete(ctx, promoted) })

	if err := store.Promote(ctx, signed.Key, promoted); err != nil {
		t.Fatalf("ANSWER: the copy failed: %v", err)
	}
	t.Logf("ANSWER: the copy to %s succeeded, so the checksum 501 does not bite", promoted)

	// The promoted object is what the public URL serves, so read it back the
	// way a client would.
	copied, err := store.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(promoted),
	})
	if err != nil {
		t.Fatalf("the promoted object could not be read back: %v", err)
	}
	if *copied.ContentLength != int64(len(body)) {
		t.Fatalf("the copy is %d bytes, expected %d", *copied.ContentLength, len(body))
	}
	if copied.ContentType == nil || *copied.ContentType != "image/png" {
		t.Errorf("ANSWER: the copy lost its content type: %v", copied.ContentType)
	} else {
		t.Log("ANSWER: the copy carries the content type across")
	}
}
