package main

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func parseLocalTestCert(t *testing.T, certFile string) *x509.Certificate {
	t.Helper()
	raw, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := pem.Decode(raw)
	if b == nil {
		t.Fatal("cert.pem non PEM")
	}
	c, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEnsureLocalCert(t *testing.T) {
	dir := t.TempDir()
	cf, kf, err := ensureLocalCertIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cf != filepath.Join(dir, "cert.pem") || kf != filepath.Join(dir, "key.pem") {
		t.Fatalf("chemins inattendus %s %s", cf, kf)
	}
	if fi, err := os.Stat(kf); err != nil {
		t.Fatalf("clé absente : %v", err)
	} else if runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600 {
		t.Fatalf("clé trop permissive : %v", fi.Mode())
	}
	c := parseLocalTestCert(t, cf)
	found := false
	for _, n := range c.DNSNames {
		if n == "meteolink.dev" {
			found = true
		}
	}
	if !found {
		t.Fatalf("SAN meteolink.dev absent : %v", c.DNSNames)
	}
	if !c.IPAddresses[0].Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("IP SAN inattendues : %v", c.IPAddresses)
	}
	if time.Until(c.NotAfter) < 9*365*24*time.Hour {
		t.Fatalf("validité trop courte : %v", c.NotAfter)
	}
	// idempotent : second appel, mêmes octets
	before, _ := os.ReadFile(cf)
	cf2, _, err := ensureLocalCertIn(dir)
	if err != nil || cf2 != cf {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(cf)
	if string(before) != string(after) {
		t.Fatal("second appel a régénéré le certificat")
	}
	// corrompu -> régénéré
	if err := os.WriteFile(cf, []byte("poubelle"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ensureLocalCertIn(dir); err != nil {
		t.Fatal(err)
	}
	if c2 := parseLocalTestCert(t, cf); c2.SerialNumber.Cmp(c.SerialNumber) == 0 {
		t.Fatal("certificat corrompu non régénéré")
	}
}

// TestLocalCertPoolTrust — le pool client (TUI) valide le cert serveur.
func TestLocalCertPoolTrust(t *testing.T) {
	if _, _, err := ensureLocalCert(); err != nil {
		t.Fatal(err)
	}
	pool, err := localCertPool()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := localCertDir()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "cert.pem"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := pem.Decode(raw)
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cert.Verify(x509.VerifyOptions{Roots: pool, DNSName: "meteolink.dev", CurrentTime: time.Now()}); err != nil {
		t.Fatalf("le pool local ne valide pas meteolink.dev : %v", err)
	}
}

func TestRedirectHTTPS(t *testing.T) {
	h := redirectHTTPS("meteolink.dev:9090")
	req := httptest.NewRequest("GET", "http://127.0.0.1:9080/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("code %d, attendu 307", rec.Code)
	}
	want := "https://meteolink.dev:9090/api/health"
	if loc := rec.Header().Get("Location"); loc != want {
		t.Fatalf("Location %q, attendu %q", loc, want)
	}
}
