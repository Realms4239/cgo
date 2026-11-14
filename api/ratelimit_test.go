package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRateLimit25RapidPostsOneSource — 25 POST rapides depuis une même source
// : au moins un 429, avec Retry-After, et le 429 arrive après la rafale
// autorisée (20), pas avant.
func TestRateLimit25RapidPostsOneSource(t *testing.T) {
	h := New(Deps{})
	got429 := 0
	for i := 0; i < 25; i++ {
		req := httptest.NewRequest("POST", "/api/burst", nil)
		req.RemoteAddr = "127.0.0.1:9999"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			got429++
			if ra := rec.Header().Get("Retry-After"); ra != "1" {
				t.Fatalf("Retry-After=%q, want 1", ra)
			}
		}
	}
	if got429 == 0 {
		t.Fatal("aucun 429 sur 25 POST rapides depuis une même source")
	}

	// Les appels suivants sont toujours limités tant que la fenêtre n'a pas tourné.
	req := httptest.NewRequest("POST", "/api/burst", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("après rafale : %d, want 429", rec.Code)
	}
}

// TestRateLimitUnderLimitPasses — 20 POST dans la rafale autorisée : aucun
// 429 (le handler répondra 503 faute de moteur câblé, jamais 429).
func TestRateLimitUnderLimitPasses(t *testing.T) {
	h := New(Deps{})
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("POST", "/api/burst", nil)
		req.RemoteAddr = "127.0.0.1:9998"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("POST #%d : 429 dans la rafale autorisée (20)", i+1)
		}
	}
}

// TestRateLimitPerIP — la limitation est par IP : une autre adresse n'est pas
// pénalisée par la rafale de la première.
func TestRateLimitPerIP(t *testing.T) {
	h := New(Deps{})
	for i := 0; i < 25; i++ {
		req := httptest.NewRequest("POST", "/api/burst", nil)
		req.RemoteAddr = "127.0.0.1:9997"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
	req := httptest.NewRequest("POST", "/api/burst", nil)
	req.RemoteAddr = "10.0.0.5:9997"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusTooManyRequests {
		t.Fatal("IP distincte pénalisée par la rafale d'une autre")
	}
}

// TestRateLimitRefillsOverTime — le seau se recharge à 10/s : après épuisement,
// une seconde de recharge réautorise des appels.
func TestRateLimitRefillsOverTime(t *testing.T) {
	l := newRateLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	for i := 0; i < int(rateBurst); i++ {
		if !l.allow("192.0.2.9") {
			t.Fatalf("appel #%d refusé dans la rafale", i+1)
		}
	}
	if l.allow("192.0.2.9") {
		t.Fatal("appel au-delà de la rafale accepté")
	}
	now = now.Add(1 * time.Second) // exactement 10 jetons rechargés
	for i := 0; i < 10; i++ {
		if !l.allow("192.0.2.9") {
			t.Fatalf("appel #%d refusé après recharge", i+1)
		}
	}
	if l.allow("192.0.2.9") {
		t.Fatal("appel accepté au-delà des jetons rechargés")
	}
	// entrées inactives expirées purgées, la map ne fuit pas
	for i := 0; i < 3; i++ {
		l.allow(fmt.Sprintf("192.0.2.%d", i+100))
	}
	now = now.Add(10 * time.Minute)
	l.allow("203.0.113.7") // déclenche la purge paresseuse
	if len(l.buckets) > 4 {
		t.Fatalf("purge paresseuse : %d entrées, want ≤ 4", len(l.buckets))
	}
}
