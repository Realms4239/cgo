package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// serveLimiter — envoie une requête au travers de l'enveloppe avec une
// adresse d'origine stable (un seul principal).
func serveLimiter(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "192.0.2.10:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// countingHandler — espion : compte les appels du handler enveloppé.
func countingHandler(calls *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		(*calls)++
		w.WriteHeader(http.StatusNoContent)
	})
}

// TestRateLimitSignature — signature exacte attendue par le plan : les quatre
// limites sont des int, le handler d'entrée et de sortie sont http.Handler.
func TestRateLimitSignature(t *testing.T) {
	var fn func(http.Handler, int, int, int, int) http.Handler = RateLimit
	calls := 0
	h := fn(countingHandler(&calls), readPerMin, mutatePerMin, readBurstSize, mutateBurst)
	if h == nil {
		t.Fatal("RateLimit a renvoyé nil")
	}
	if rec := serveLimiter(h, http.MethodGet, "/api/cost/tiers"); rec.Code != http.StatusNoContent {
		t.Fatalf("requête dans la rafale : %d, want 204", rec.Code)
	}
	if calls != 1 {
		t.Fatalf("handler appelé %d fois, want 1", calls)
	}
}

// TestRateLimitMutateBurstRefillAndRetryAfter — oracle hostile : rafale 5,
// puis 429 avec Retry-After=6 s (10/min), sans appel au handler refusé.
func TestRateLimitMutateBurstRefillAndRetryAfter(t *testing.T) {
	calls := 0
	l := newPrincipalLimiter(countingHandler(&calls), readPerMin, mutatePerMin, readBurstSize, mutateBurst)
	now := time.Now()
	l.now = func() time.Time { return now }

	for i := 0; i < mutateBurst; i++ {
		if rec := serveLimiter(l, http.MethodPost, "/api/run/stop"); rec.Code != http.StatusNoContent {
			t.Fatalf("mutation #%d : %d, want 204", i+1, rec.Code)
		}
	}
	rec := serveLimiter(l, http.MethodPost, "/api/run/stop")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("mutation #%d : %d, want 429", mutateBurst+1, rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra != "6" {
		t.Fatalf("Retry-After=%q, want 6 (10/min)", ra)
	}
	body := rec.Body.String()
	if strings.Contains(body, "/api/") || strings.Contains(body, "192.0.2.10") {
		t.Fatalf("corps 429 avec fuite : %s", body)
	}
	if calls != mutateBurst {
		t.Fatalf("handler appelé %d fois, want %d", calls, mutateBurst)
	}

	// 7 s > 6 s : exactement un jeton rechargé, puis plus rien.
	now = now.Add(7 * time.Second)
	if rec := serveLimiter(l, http.MethodPost, "/api/run/stop"); rec.Code != http.StatusNoContent {
		t.Fatalf("après recharge : %d, want 204", rec.Code)
	}
	if rec := serveLimiter(l, http.MethodPost, "/api/run/stop"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("au-delà du jeton rechargé : %d, want 429", rec.Code)
	}
	if calls != mutateBurst+1 {
		t.Fatalf("handler appelé %d fois, want %d", calls, mutateBurst+1)
	}
}

// TestRateLimitReadBurstRefillAndRetryAfter — même oracle en lecture :
// rafale 20, puis 429 avec Retry-After=1 s (60/min).
func TestRateLimitReadBurstRefillAndRetryAfter(t *testing.T) {
	calls := 0
	l := newPrincipalLimiter(countingHandler(&calls), readPerMin, mutatePerMin, readBurstSize, mutateBurst)
	now := time.Now()
	l.now = func() time.Time { return now }

	for i := 0; i < readBurstSize; i++ {
		if rec := serveLimiter(l, http.MethodGet, "/api/cost/tiers"); rec.Code != http.StatusNoContent {
			t.Fatalf("lecture #%d : %d, want 204", i+1, rec.Code)
		}
	}
	rec := serveLimiter(l, http.MethodGet, "/api/cost/tiers")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("lecture #%d : %d, want 429", readBurstSize+1, rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra != "1" {
		t.Fatalf("Retry-After=%q, want 1 (60/min)", ra)
	}
	body := rec.Body.String()
	if strings.Contains(body, "/api/") || strings.Contains(body, "192.0.2.10") {
		t.Fatalf("corps 429 avec fuite : %s", body)
	}

	now = now.Add(time.Second + 100*time.Millisecond) // 1,1 jeton : un seul appel autorisé
	if rec := serveLimiter(l, http.MethodGet, "/api/cost/tiers"); rec.Code != http.StatusNoContent {
		t.Fatalf("après recharge : %d, want 204", rec.Code)
	}
	if rec := serveLimiter(l, http.MethodGet, "/api/cost/tiers"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("au-delà du jeton rechargé : %d, want 429", rec.Code)
	}
}

// TestRateLimitPerPrincipalBuckets — la limitation est par principal et par
// classe : épuiser le seau d'un principal n'affecte ni un autre principal ni
// l'autre classe du même principal.
func TestRateLimitPerPrincipalBuckets(t *testing.T) {
	l := newPrincipalLimiter(countingHandler(new(int)), readPerMin, mutatePerMin, readBurstSize, mutateBurst)
	now := time.Now()
	l.now = func() time.Time { return now }
	rate, cap := l.mutateRate, l.mutateCap

	for i := 0; i < mutateBurst; i++ {
		if ok, _ := l.allow("principal-A|w", rate, cap); !ok {
			t.Fatalf("principal A : jeton #%d refusé", i+1)
		}
	}
	if ok, _ := l.allow("principal-A|w", rate, cap); ok {
		t.Fatal("seau du principal A épuisé mais un jeton est encore accordé")
	}
	if ok, _ := l.allow("principal-B|w", rate, cap); !ok {
		t.Fatal("principal B pénalisé par l'épuisement de A")
	}
	if ok, _ := l.allow("principal-A|r", l.readRate, l.readCap); !ok {
		t.Fatal("seau de lecture de A pénalisé par son seau de mutation")
	}
}

// TestRateLimitLazyPrune — les seaux inactifs sont purgés : la map ne fuit pas.
func TestRateLimitLazyPrune(t *testing.T) {
	l := newPrincipalLimiter(countingHandler(new(int)), readPerMin, mutatePerMin, readBurstSize, mutateBurst)
	now := time.Now()
	l.now = func() time.Time { return now }
	rate, cap := l.mutateRate, l.mutateCap

	l.allow("a|w", rate, cap)
	l.allow("b|w", rate, cap)
	now = now.Add(10 * time.Minute)
	l.allow("c|w", rate, cap) // déclenche la purge paresseuse
	if len(l.buckets) != 1 {
		t.Fatalf("purge paresseuse : %d entrées, want 1", len(l.buckets))
	}
	if _, ok := l.buckets["c|w"]; !ok {
		t.Fatal("le seau actif a été purgé")
	}
}

// TestRateLimitInvalidConfigFailsClosed — des limites non positives ne
// deviennent jamais un accès illimité : 429 + Retry-After, handler non appelé.
func TestRateLimitInvalidConfigFailsClosed(t *testing.T) {
	calls := 0
	h := RateLimit(countingHandler(&calls), 0, 0, 0, 0)
	rec := serveLimiter(h, http.MethodGet, "/api/health")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("configuration invalide : %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After absent sur le refus fail-closed")
	}
	if calls != 0 {
		t.Fatal("handler appelé sous une configuration invalide")
	}
}

// TestRateLimitMethodClassification — GET/HEAD/OPTIONS/TRACE sont des lectures,
// POST/PUT/PATCH/DELETE des mutations.
func TestRateLimitMethodClassification(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace} {
		if !isReadMethod(m) {
			t.Fatalf("%s classé mutation, want lecture", m)
		}
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if isReadMethod(m) {
			t.Fatalf("%s classé lecture, want mutation", m)
		}
	}
}
