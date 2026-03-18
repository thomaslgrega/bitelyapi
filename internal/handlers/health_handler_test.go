package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerHealth(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", rec.Body.String())
	}
}
