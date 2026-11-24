package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRequireConfirm — la garde refuse un POST non confirmé et laisse passer
// un POST portant X-Confirm-Action: true.
func TestRequireConfirm(t *testing.T) {
	called := false
	h := requireConfirm(func(w http.ResponseWriter, r *http.Request) { called = true })
	r := httptest.NewRequest(http.MethodPost, "/api/shape", nil)
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusBadRequest || called {
		t.Fatalf("missing X-Confirm-Action: code=%d called=%v, want 400 and no call", w.Code, called)
	}
	r2 := httptest.NewRequest(http.MethodPost, "/api/shape", nil)
	r2.Header.Set("X-Confirm-Action", "true")
	w2 := httptest.NewRecorder()
	h(w2, r2)
	if !called {
		t.Fatal("confirmed request must reach the handler")
	}
}

// TestDestructiveRoutesRequireConfirm — chaque POST destructif porte la garde :
// sans en-tête, la réponse porte le message de confirmation (400) avant même
// d'atteindre le handler ; avec l'en-tête, ce message disparaît. Un corps JSON
// invalide est volontairement envoyé pour éviter tout effet de bord métier :
// l'oracle est le message d'erreur, pas le code final.
func TestDestructiveRoutesRequireConfirm(t *testing.T) {
	SetDevNoAuth(false)
	defer SetDevNoAuth(false)
	t.Setenv("CGO_API_TOKEN", "confirm-secret")

	for _, path := range []string{
		"/api/shape", "/api/burst", "/api/run/start", "/api/profile/delete", "/api/cost/tier",
	} {
		// Un handler neuf par route : la limite de mutation déclarée n'est
		// pas le sujet de ce test de confirmation.
		h := New(Deps{})
		call := func(confirm bool) string {
			req := httptest.NewRequest(http.MethodPost, "http://example.com"+path, strings.NewReader("{"))
			req.RemoteAddr = "10.9.8.7:1234"
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer confirm-secret")
			if confirm {
				req.Header.Set("X-Confirm-Action", "true")
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			body, _ := io.ReadAll(rec.Result().Body)
			return string(body)
		}

		without := call(false)
		if !strings.Contains(without, "confirmation requise") {
			t.Errorf("POST %s sans X-Confirm-Action : corps %q, want message de confirmation", path, without)
		}
		with := call(true)
		if strings.Contains(with, "confirmation requise") {
			t.Errorf("POST %s avec X-Confirm-Action : corps %q, confirmation encore requise", path, with)
		}
	}
}
