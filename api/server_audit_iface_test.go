package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRemediateRejectsInvalidIface(t *testing.T) {
	for _, q := range []string{
		"?iface=eth0%3Bid", "?iface=eth0%20ls", "?iface=eth0%7Cid",
		"?routeros_iface=ether1%3Bid", "?routeros_lan=..%2Fetc%2Fpasswd",
		"?iface=abcdefghijklmnop",
	} {
		r := httptest.NewRequest(http.MethodGet, "/api/remediate"+q, nil)
		w := httptest.NewRecorder()
		handleRemediate(nil)(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET /api/remediate%s = %d, want 400", q, w.Code)
		}
	}
}

func TestValidIfaceTokens(t *testing.T) {
	for _, name := range []string{"eth0", "ifb0", "ether1-wan", "bridge"} {
		if !validIface(name) {
			t.Errorf("validIface(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"eth0;id", "eth0 id", "../etc/passwd", "abcdefghijklmnop", ""} {
		if validIface(name) {
			t.Errorf("validIface(%q) = true, want false", name)
		}
	}
}
