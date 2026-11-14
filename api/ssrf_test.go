package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/audit"
)

// startAudit POST /api/audit/start et rend le code HTTP.
func startAudit(t *testing.T, srv *httptest.Server, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

// auditStubDeps — l'audit tourne sans réseau : le petit objet rend 1 ms.
func auditStubDeps() audit.Deps {
	return audit.Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{42} },
		Small: func(context.Context) (float64, error) { return 1, nil },
	}
}

func ssrfDeps() Deps {
	return Deps{
		SinksFn:     func() audit.Sinks { return audit.Sinks{Mode: "external", Target: "8.8.8.8"} },
		AuditDepsFn: auditStubDeps,
	}
}

// TestAuditStartSSRFPrivateRefused — small_url/bulk_addr vers privé → 403
// avec erreur nommée, AVANT toute mesure (la session n'est pas armée).
func TestAuditStartSSRFPrivateRefused(t *testing.T) {
	h := New(ssrfDeps())
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	cases := []struct{ name, body, want string }{
		{"rfc1918 small_url", `{"site":"ssrf-rfc1918","link_type":"fiber","duration":10,"small_url":"http://10.30.0.2:8081/small"}`, "RFC1918"},
		{"link-local small_url", `{"site":"ssrf-ll","link_type":"fiber","duration":10,"small_url":"http://169.254.169.254/latest/meta-data"}`, "link-local"},
		{"ULA small_url", `{"site":"ssrf-ula","link_type":"fiber","duration":10,"small_url":"http://[fd00::1]:8081/"}`, "ULA"},
		{"loopback small_url", `{"site":"ssrf-lo","link_type":"fiber","duration":10,"small_url":"http://127.0.0.1:8081/small"}`, "loopback"},
		{"mapped-ipv4 small_url", `{"site":"ssrf-mapped","link_type":"fiber","duration":10,"small_url":"http://[::ffff:10.0.0.1]/"}`, "RFC1918"},
		{"rfc1918 bulk_addr", `{"site":"ssrf-bulk","link_type":"fiber","duration":10,"bulk_addr":"10.30.0.2:5201"}`, "RFC1918"},
		{"link-local bulk_addr", `{"site":"ssrf-bulk-ll","link_type":"fiber","duration":10,"bulk_addr":"169.254.0.1:5201"}`, "link-local"},
	}
	for _, tc := range cases {
		code, body := startAudit(t, srv, tc.body)
		if code != http.StatusForbidden {
			t.Fatalf("%s : code %d, want 403 (body %q)", tc.name, code, body)
		}
		if !strings.Contains(body, tc.want) {
			t.Fatalf("%s : erreur %q sans la plage nommée %q", tc.name, body, tc.want)
		}
	}
	waitAuditIdle(t, srv)
}

// TestAuditStartSSRFOptInFlips — l'opt-in banc CGO_ALLOW_PRIVATE_TARGETS=1
// rouvre le laboratoire (10.30.0.2).
func TestAuditStartSSRFOptInFlips(t *testing.T) {
	t.Setenv("CGO_ALLOW_PRIVATE_TARGETS", "1")
	h := New(ssrfDeps())
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	code, body := startAudit(t, srv, `{"site":"ssrf-optin","link_type":"fiber","duration":10,"small_url":"http://10.30.0.2:8081/small","bulk_addr":"10.30.0.2:5201"}`)
	if code != http.StatusOK {
		t.Fatalf("opt-in : code %d, want 200 (body %q)", code, body)
	}
	waitAuditIdle(t, srv)
}

// TestAuditStartSSRFPublicAllowed — une cible publique traverse la garde.
// (La sonde échoue après coup si l'hôte est injoignable — l'audit le
// consigne, ce n'est pas un 403 : le chemin d'erreur de l'audit est ignoré.)
func TestAuditStartSSRFPublicAllowed(t *testing.T) {
	h := New(ssrfDeps())
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	code, body := startAudit(t, srv, `{"site":"ssrf-public","link_type":"fiber","duration":10,"small_url":"http://1.1.1.1:80/small","bulk_addr":"1.1.1.1:5201"}`)
	if code != http.StatusOK {
		t.Fatalf("public : code %d, want 200 (body %q)", code, body)
	}
	waitAuditIdle(t, srv)
}

// TestAuditStartSSRFArgsMalformed — schémas et formes refusés → 400
// (erreur client), et un small_url vide garde le chemin existant : la
// garde ne s'applique pas à l'absence de charge.
func TestAuditStartSSRFArgsMalformed(t *testing.T) {
	h := New(ssrfDeps())
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	for _, tc := range []struct{ name, body string }{
		{"file scheme", `{"site":"ssrf-file","link_type":"fiber","duration":10,"small_url":"file:///etc/passwd"}`},
		{"bulk sans port", `{"site":"ssrf-bulknp","link_type":"fiber","duration":10,"bulk_addr":"10.30.0.2"}`},
	} {
		code, _ := startAudit(t, srv, tc.body)
		if code != http.StatusBadRequest {
			t.Fatalf("%s : code %d, want 400", tc.name, code)
		}
	}

	// small_url absent → pas de charge implicite, pas de 403 (chemin nominal).
	code, _ := startAudit(t, srv, `{"site":"ssrf-empty","link_type":"fiber","duration":10}`)
	if code != http.StatusOK {
		t.Fatalf("small_url absent : code %d, want 200", code)
	}
	waitAuditIdle(t, srv)
}

// TestAuditStartSSRFGuardRunsBeforeSideEffects — la porte est devant la
// session : un refus n'arme pas l'audit.
func TestAuditStartSSRFGuardRunsBeforeSideEffects(t *testing.T) {
	h := New(ssrfDeps())
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	code, _ := startAudit(t, srv, `{"site":"ssrf-sidefx","link_type":"fiber","duration":10,"small_url":"http://10.30.0.2:8081/"}`)
	if code != http.StatusForbidden {
		t.Fatalf("code %d, want 403", code)
	}
	st, err := http.Get(srv.URL + "/api/audit/status")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Body.Close()
	var status struct {
		Running bool `json:"running"`
	}
	_ = json.NewDecoder(st.Body).Decode(&status)
	if status.Running {
		t.Fatal("audit armé malgré le refus SSRF")
	}
}

// TestGuardTables — table de la garde (plages + opt-in) au niveau unité.
func TestGuardTables(t *testing.T) {
	cases := map[string]bool{
		"10.0.0.1": true, "172.16.0.1": true, "172.31.255.254": true,
		"192.168.1.1": true, "169.254.1.1": true, "127.0.0.1": true,
		"fd00::1": true, "fe80::dead:beef": true, "::ffff:10.1.2.3": true,
		"1.1.1.1": false, "2606:4700:4700::1111": false, "172.15.0.1": false,
		"192.167.0.1": false,
	}
	for host, wantBlocked := range cases {
		err := checkHost(context.Background(), host, false)
		var pe *privateTargetError
		got := errors.As(err, &pe)
		if got != wantBlocked {
			t.Fatalf("%s : bloquée=%v, want %v (err=%v)", host, got, wantBlocked, err)
		}
	}
}
