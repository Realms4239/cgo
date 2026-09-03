package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePing(t *testing.T) {
	out := "64 bytes from x: icmp_seq=1 ttl=57 time=21.3 ms\n" +
		"64 bytes from x: icmp_seq=2 ttl=57 time=20.9 ms\n" +
		"request timed out\n" +
		"64 bytes from x: icmp_seq=3 ttl=57 time=88.1 ms\n"
	s := parsePing(out)
	if len(s) != 3 || s[0].RTTms != 20.9 || s[2].RTTms != 88.1 {
		t.Fatalf("parse = %+v", s)
	}
}

// TestParsePingWindowsLocales — sorties Windows EN + FR : `time=12ms`
// (collé), `temps=12ms`, `temps<1ms` (→ 0,5, moitié d'intervalle).
func TestParsePingWindowsLocales(t *testing.T) {
	out := "Reply from 8.8.8.8: bytes=32 time=12ms TTL=117\n" +
		"Réponse de 8.8.8.8 : octets=32 temps=13ms TTL=117\n" +
		"Réponse de 8.8.8.8 : octets=32 temps<1ms TTL=128\n" +
		"Paquets : envoyés = 4, reçus = 3, perdus = 1 (perte 25%),\n"
	s := parsePing(out)
	if len(s) != 3 || s[0].RTTms != 0.5 || s[1].RTTms != 12 || s[2].RTTms != 13 {
		t.Fatalf("parse windows = %+v", s)
	}
}

// TestPingLiveLoopback — le binaire ping réel répond sur 127.0.0.1 avec les
// args de la plateforme (`-n` Windows, `-c` Unix). ~2 s.
func TestPingLiveLoopback(t *testing.T) {
	s, err := Ping(context.Background(), "127.0.0.1", 2, 200)
	if err != nil {
		t.Fatalf("live ping failed: %v", err)
	}
	if len(s) == 0 {
		t.Fatal("live ping returned no samples")
	}
}

func TestSmallObjectTimesCompletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(make([]byte, 16*1024)) // small object payload
	}))
	defer srv.Close()
	ms, err := SmallObject(context.Background(), srv.Client(), srv.URL)
	if err != nil || ms <= 0 {
		t.Fatalf("ms=%v err=%v", ms, err)
	}
}
