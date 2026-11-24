package api

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	pruneIdle  = 2 * time.Minute
	pruneEvery = time.Minute
)

// tokenBucket — seau à jetons d'un principal pour une classe (lecture/mutation).
type tokenBucket struct {
	tokens float64
	last   time.Time
}

// principalLimiter — limiteur par principal posé en enveloppe du mux. Chaque
// principal a deux seaux (lecture et mutation) ; les limites sont déclarées à
// la construction et jamais élargies à l'exécution. Purge paresseuse : les
// entrées inactives depuis > 2 minutes partent au plus une fois par minute,
// sans goroutine dédiée.
type principalLimiter struct {
	next       http.Handler
	readRate   float64 // jetons/seconde
	mutateRate float64
	readCap    float64 // rafale
	mutateCap  float64
	mu         sync.Mutex
	buckets    map[string]*tokenBucket
	now        func() time.Time
	lastPrune  time.Time
}

func newPrincipalLimiter(next http.Handler, readsPerMin, mutatesPerMin, readBurst, mutateBurst int) *principalLimiter {
	return &principalLimiter{
		next:       next,
		readRate:   float64(readsPerMin) / 60,
		mutateRate: float64(mutatesPerMin) / 60,
		readCap:    float64(readBurst),
		mutateCap:  float64(mutateBurst),
		buckets:    make(map[string]*tokenBucket),
		now:        time.Now,
	}
}

// RateLimit — enveloppe du mux : 429 + Retry-After quand le seau du principal
// est vide. Aucune limite n'est lue dans l'environnement ni modifiable après
// la construction ; l'authentification reste portée par ServeHTTP, en amont.
func RateLimit(next http.Handler, readsPerMin, mutatesPerMin, readBurst, mutateBurst int) http.Handler {
	return newPrincipalLimiter(next, readsPerMin, mutatesPerMin, readBurst, mutateBurst)
}

func isReadMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

// principalKey — identité du titulaire du seau : empreinte du jeton pour un
// principal authentifié (Principal.OperatorID), adresse d'origine sinon. Les
// anonymes sont refusés par ServeHTTP avant d'atteindre cette enveloppe.
func principalKey(r *http.Request) string {
	if id := principalFrom(r).OperatorID; id != "" {
		return id
	}
	return "anon:" + clientIP(r)
}

func (l *principalLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rate, capacity, class := l.mutateRate, l.mutateCap, "w"
	if isReadMethod(r.Method) {
		rate, capacity, class = l.readRate, l.readCap, "r"
	}
	key := principalKey(r) + "|" + class
	if ok, retry := l.allow(key, rate, capacity); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(retry))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		// Corps fixe : ni chemin demandé, ni secret, ni identité du principal.
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "trop de requêtes — réessayez plus tard"})
		return
	}
	l.next.ServeHTTP(w, r)
}

// allow — consomme un jeton ; renvoie le délai avant le prochain jeton sinon.
func (l *principalLimiter) allow(key string, rate, capacity float64) (bool, int) {
	if rate <= 0 || capacity <= 0 {
		return false, 1 // configuration invalide : fail closed, jamais une limite élargie
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.lastPrune) > pruneEvery {
		for k, b := range l.buckets {
			if now.Sub(b.last) > pruneIdle {
				delete(l.buckets, k)
			}
		}
		l.lastPrune = now
	}
	b := l.buckets[key]
	if b == nil {
		b = &tokenBucket{tokens: capacity, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * rate
	if b.tokens > capacity {
		b.tokens = capacity
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	retry := int(math.Ceil((1 - b.tokens) / rate))
	if retry < 1 {
		retry = 1
	}
	return false, retry
}

// clientIP — hôte d'origine (RemoteAddr sans le port) ; l'adresse complète
// si le port est absent.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
