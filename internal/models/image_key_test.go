package models

import (
	"strings"
	"testing"
)

func TestNewStagedImageKeyStagesUnderIncoming(t *testing.T) {
	key, err := NewStagedImageKey()
	if err != nil {
		t.Fatalf("failed to mint a staged key: %v", err)
	}

	if !strings.HasPrefix(key, "incoming/") {
		t.Fatalf("expected a key staged under incoming/, got %q", key)
	}
	if !IsStagedImageKey(key) {
		t.Fatalf("expected a minted key to pass shape validation, got %q", key)
	}

	other, err := NewStagedImageKey()
	if err != nil {
		t.Fatalf("failed to mint a second staged key: %v", err)
	}
	if other == key {
		t.Fatalf("expected each staged key to be distinct, got %q twice", key)
	}
}

func TestIsStagedImageKeyRefusesAnythingButIncomingUUID(t *testing.T) {
	refused := []string{
		"",
		"incoming/",
		"incoming/not-a-uuid",
		"incoming/9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60/../../etc",
		"incoming/9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60.jpg",
		"recipes/recipe-1/image.jpg",
		"9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60",
		"Incoming/9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60",
		"incoming/{9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60}",
	}

	for _, key := range refused {
		if IsStagedImageKey(key) {
			t.Fatalf("expected %q to be refused", key)
		}
	}

	if !IsStagedImageKey("incoming/9f1b9a4e-6e2a-4a3f-9d4a-1c2b3d4e5f60") {
		t.Fatal("expected a well-formed staged key to be accepted")
	}
}

func TestPromotedImageKeyIsDerivedFromTheRecipe(t *testing.T) {
	if got := PromotedImageKey("recipe-1"); got != "recipes/recipe-1/image.jpg" {
		t.Fatalf("expected a key derived from the recipe id, got %q", got)
	}
}

func TestAllowedImageContentType(t *testing.T) {
	for _, allowed := range []string{"image/jpeg", "image/png"} {
		if !AllowedImageContentType(allowed) {
			t.Fatalf("expected %q to be allowed", allowed)
		}
	}

	// image/jpg is the client mistake worth naming: a signed content type pins
	// one exact string, so it uploads to a 403 (ADR-0006).
	for _, refused := range []string{"image/jpg", "image/heic", "image/gif", "image/svg+xml", "IMAGE/PNG", ""} {
		if AllowedImageContentType(refused) {
			t.Fatalf("expected %q to be refused", refused)
		}
	}
}
