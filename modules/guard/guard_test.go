package guard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func ok(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

// TestRequireAuthToken — jeton correct passe, absence = 401.
func TestRequireAuthToken(t *testing.T) {
	h := RequireAuthToken("secret", ok)
	req := httptest.NewRequest("POST", "/api/run/start", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("jeton correct : want 200, got %d", rec.Code)
	}
	req2 := httptest.NewRequest("POST", "/api/run/start", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("sans jeton : want 401, got %d", rec2.Code)
	}
}

// TestRequireAuthLoopback — jeton vide : loopback passe, distant 401.
func TestRequireAuthLoopback(t *testing.T) {
	h := RequireAuthToken("", ok)
	req := httptest.NewRequest("GET", "/api/health", nil)
	req.RemoteAddr = "127.0.0.1:9090"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dev loopback : want 200, got %d", rec.Code)
	}
	req2 := httptest.NewRequest("GET", "/api/health", nil)
	req2.RemoteAddr = "203.0.113.7:9090"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("distant sans jeton : want 401, got %d", rec2.Code)
	}
}
