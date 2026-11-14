package kit

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

// certTestPEM — certificat auto-signé de test (durées paramétrables).
func certTestPEM(t *testing.T, notBefore, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "meteolink.dev"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// TestInspectCertPEM — SPKI sha256 (hex), empreinte SHA-256 colon-hex et
// fenêtre de validité, issus d'un certificat réel généré en mémoire.
func TestInspectCertPEM(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second) // DER encode la seconde, en UTC
	nb := now.Add(-24 * time.Hour)
	na := now.Add(300 * 24 * time.Hour)
	ci, err := inspectCertPEM(certTestPEM(t, nb, na))
	if err != nil {
		t.Fatal(err)
	}
	if len(ci.SPKISha256Hex) != 64 || strings.ToLower(ci.SPKISha256Hex) != ci.SPKISha256Hex {
		t.Fatalf("SPKI sha256 hex inattendu : %q", ci.SPKISha256Hex)
	}
	parts := strings.Split(ci.Fingerprint, ":")
	if len(parts) != 32 {
		t.Fatalf("empreinte SHA-256 : %d octets, want 32 (%q)", len(parts), ci.Fingerprint)
	}
	for _, p := range parts {
		if len(p) != 2 {
			t.Fatalf("empreinte mal formée : %q", p)
		}
	}
	if !ci.NotBefore.Equal(nb) || !ci.NotAfter.Equal(na) {
		t.Fatalf("fenêtre : %s → %s, want %s → %s", ci.NotBefore, ci.NotAfter, nb, na)
	}
	if d := int(ci.Lifetime.Hours() / 24); d < 299 || d > 302 {
		t.Fatalf("durée de vie : %d j, want ~301", d)
	}
}

// TestInspectCertSPKIMatchesDER — la SPKI est dérivée de la clé publique,
// pas de la signature du certificat : deux certificats de MÊME clé
// (renouvelés) partagent la SPKI ; l'empreinte DER, elle, change.
func TestInspectCertSPKIMatchesDER(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	mk := func(notAfter time.Time) []byte {
		serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
		if err != nil {
			t.Fatal(err)
		}
		tmpl := &x509.Certificate{
			SerialNumber: serial,
			Subject:      pkix.Name{CommonName: "meteolink.dev"},
			NotBefore:    now.Add(-time.Hour),
			NotAfter:     notAfter,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	}
	ciA, errA := inspectCertPEM(mk(now.Add(365 * 24 * time.Hour)))
	ciB, errB := inspectCertPEM(mk(now.Add(730 * 24 * time.Hour)))
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}
	if ciA.SPKISha256Hex != ciB.SPKISha256Hex {
		t.Fatal("même clé, SPKI différentes — la garde n'épingle pas la clé")
	}
	if ciA.Fingerprint == ciB.Fingerprint {
		t.Fatal("renouvellement non visible (empreinte DER identique)")
	}
}

// TestCertUsable — expiré → refus, pas-encore-valable → refus, en cours → OK.
func TestCertUsable(t *testing.T) {
	now := time.Now()
	pemOK := certTestPEM(t, now.Add(-time.Hour), now.Add(24*time.Hour))
	ci, err := inspectCertPEM(pemOK)
	if err != nil {
		t.Fatal(err)
	}
	if err := certUsable(ci, now); err != nil {
		t.Fatalf("cert valide refusé : %v", err)
	}
	pemExpired := certTestPEM(t, now.Add(-48*time.Hour), now.Add(-time.Hour))
	ciExp, err := inspectCertPEM(pemExpired)
	if err != nil {
		t.Fatal(err)
	}
	if err := certUsable(ciExp, now); err == nil {
		t.Fatal("cert expiré accepté")
	}
	pemFuture := certTestPEM(t, now.Add(time.Hour), now.Add(48*time.Hour))
	ciFut, err := inspectCertPEM(pemFuture)
	if err != nil {
		t.Fatal(err)
	}
	if err := certUsable(ciFut, now); err == nil {
		t.Fatal("cert pas encore valable accepté")
	}
}

// TestInspectCertRefusals — PEM absent / mauvais type / DER illisible.
func TestInspectCertRefusals(t *testing.T) {
	if _, err := inspectCertPEM([]byte("pas un pem")); err == nil {
		t.Fatal("PEM absent accepté")
	}
	block := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("x")})
	if _, err := inspectCertPEM(block); err == nil {
		t.Fatal("bloc non-CERTIFICATE accepté")
	}
}
