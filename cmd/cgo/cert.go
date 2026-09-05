package main

// Certificat local auto-signé — meteolink.dev en HTTPS par défaut.
//
// .dev est préchargé HSTS : les navigateurs refusent le HTTP sur
// meteolink.dev. Au premier --serve, un certificat ECDSA P-256 auto-signé
// (10 ans, SANs meteolink.dev / meteolink.vm / localhost / 127.0.0.1 / ::1)
// est généré dans le dossier config utilisateur, puis réutilisé.
// Stdlib uniquement, zéro dépendance, zéro réseau.
//
// Avertissement navigateur attendu (CA non publique) : accepter une fois,
// ou installer cert.pem dans le magasin de confiance de l'OS.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// localCertDNS — noms couverts par le certificat local.
var localCertDNS = []string{"meteolink.dev", "meteolink.vm", "localhost"}

// localCertValidity — régénère si expiry dans moins de 30 jours.
const localCertValidity = 30 * 24 * time.Hour

func localCertDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "cgo"), nil
}

// ensureLocalCert génère le couple cert/clé local s'il est absent,
// illisible ou bientôt expiré. Idempotent.
func ensureLocalCert() (certFile, keyFile string, err error) {
	dir, err := localCertDir()
	if err != nil {
		return "", "", err
	}
	return ensureLocalCertIn(dir)
}

func ensureLocalCertIn(dir string) (certFile, keyFile string, err error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", err
	}
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	if validLocalCert(certFile, keyFile) {
		return certFile, keyFile, nil
	}
	if err := generateLocalCert(certFile, keyFile); err != nil {
		return "", "", err
	}
	return certFile, keyFile, nil
}

// validLocalCert — cert + clé lisibles, SAN meteolink.dev présent,
// expiration au-delà de 30 jours.
func validLocalCert(certFile, keyFile string) bool {
	craw, err := os.ReadFile(certFile)
	if err != nil {
		return false
	}
	kraw, err := os.ReadFile(keyFile)
	if err != nil {
		return false
	}
	cb, _ := pem.Decode(craw)
	kb, _ := pem.Decode(kraw)
	if cb == nil || kb == nil {
		return false
	}
	cert, err := x509.ParseCertificate(cb.Bytes)
	if err != nil {
		return false
	}
	if time.Until(cert.NotAfter) < localCertValidity {
		return false
	}
	for _, n := range cert.DNSNames {
		if n == "meteolink.dev" {
			return true
		}
	}
	return false
}

func generateLocalCert(certFile, keyFile string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}
	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "meteolink.dev", Organization: []string{"meteolink local"}},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     localCertDNS,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	kf, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if err := pem.Encode(kf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}); err != nil {
		_ = kf.Close()
		return err
	}
	if err := kf.Close(); err != nil {
		return err
	}
	cf, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		_ = cf.Close()
		return err
	}
	return cf.Close()
}

// localCertPool — racines système + certificat local (clients Go : TUI).
// Erreur -> nil, nil : l'appelant retombe sur les racines système.
func localCertPool() (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	dir, err := localCertDir()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "cert.pem"))
	if err != nil {
		return nil, err
	}
	if !pool.AppendCertsFromPEM(raw) {
		return nil, fmt.Errorf("certificat local illisible")
	}
	return pool, nil
}
