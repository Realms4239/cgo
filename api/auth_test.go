package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRequireAuth_UnauthenticatedMutationIs401(t *testing.T) {
	os.Setenv("CGO_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CGO_API_TOKEN")
	h := requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/shape", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
