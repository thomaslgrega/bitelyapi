package r2

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

func testStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStore(Config{
		AccountID:       "account-1",
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
		Bucket:          "bitely-images",
	})
	if err != nil {
		t.Fatalf("failed to build the store: %v", err)
	}

	return store
}

// Presigning is a local signing operation, so this test needs no network and
// no bucket.
func TestPresignUploadAddressesTheAccountEndpoint(t *testing.T) {
	store := testStore(t)

	upload, err := store.PresignUpload(context.Background(), "image/jpeg", 1024)
	if err != nil {
		t.Fatalf("failed to presign an upload: %v", err)
	}

	if upload.UploadURL == "" {
		t.Fatal("expected a signed upload URL")
	}

	signed, err := url.Parse(upload.UploadURL)
	if err != nil {
		t.Fatalf("failed to parse the upload URL: %v", err)
	}

	// A presigned URL works on neither the Public Development URL nor a custom
	// domain, so the write host is the account endpoint by construction
	// (ADR-0006).
	if signed.Host != "account-1.r2.cloudflarestorage.com" {
		t.Fatalf("expected the account endpoint, got %q", signed.Host)
	}
	if signed.Scheme != "https" {
		t.Fatalf("expected an https URL, got %q", signed.Scheme)
	}

	if !models.IsStagedImageKey(upload.Key) {
		t.Fatalf("expected a staged key, got %q", upload.Key)
	}
	if signed.Path != "/bitely-images/"+upload.Key {
		t.Fatalf("expected the URL to address the staged key, got %q", signed.Path)
	}
}

func TestPresignUploadSignsTheDeclaredUpload(t *testing.T) {
	store := testStore(t)

	upload, err := store.PresignUpload(context.Background(), "image/png", 2048)
	if err != nil {
		t.Fatalf("failed to presign an upload: %v", err)
	}

	signed, err := url.Parse(upload.UploadURL)
	if err != nil {
		t.Fatalf("failed to parse the upload URL: %v", err)
	}
	query := signed.Query()

	headers := query.Get("X-Amz-SignedHeaders")
	for _, header := range []string{"content-length", "content-type"} {
		if !strings.Contains(headers, header) {
			t.Fatalf("expected %q to be signed, got %q", header, headers)
		}
	}

	// Signing a checksum header covers an empty body, and every real upload
	// through the URL then fails (ADR-0006).
	if strings.Contains(headers, "x-amz-checksum") {
		t.Fatalf("expected no checksum header to be signed, got %q", headers)
	}

	if query.Get("X-Amz-Expires") != "300" {
		t.Fatalf("expected a five minute URL, got %q", query.Get("X-Amz-Expires"))
	}

	remaining := time.Until(upload.ExpiresAt)
	if remaining <= 4*time.Minute || remaining > 5*time.Minute {
		t.Fatalf("expected the expiry to be about five minutes out, got %s", remaining)
	}
}
