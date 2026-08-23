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
