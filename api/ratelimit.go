package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimiter — seau à jetons par IP cliente : 10 jetons/s, rafale 20.
// Seules les routes POST sont limitées (SSE et lectures restent libres).
// Purge paresseuse : les entrées inactives depuis > 2 minutes sont supprimées
// au plus une fois par minute — la map ne fuit pas, sans goroutine dédiée.
type rateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*tokenBucket
	now       func() time.Time
	lastPrune time.Time
}

type tokenBucket struct {
	tokens float64
	last   time.Time
}

const (
	ratePerSec = 10.0 // recharge, jetons par seconde
	rateBurst  = 20.0 // capacité du seau (rafale)
	pruneIdle  = 2 * time.Minute
	pruneEvery = time.Minute
)

func newRateLimiter() *rateLimiter {
	return &rateLimiter{buckets: make(map[string]*tokenBucket), now: time.Now}
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

func (l *rateLimiter) allow(ip string) bool {
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
	b, ok := l.buckets[ip]
	if !ok {
		b = &tokenBucket{tokens: rateBurst, last: now}
		l.buckets[ip] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * ratePerSec
	if b.tokens > rateBurst {
		b.tokens = rateBurst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
