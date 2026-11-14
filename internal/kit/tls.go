package kit

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// certInfo — ce qu'un opérateur doit voir AVANT d'accepter la confiance :
// SPKI sha256 (hex), empreinte SHA-256 du DER, fenêtre de validité.
type certInfo struct {
	SPKISha256Hex string
	Fingerprint   string // SHA-256 du DER, octets séparés par « : »
	SHA1          string // pour certutil (-delstore cible cette empreinte sous Windows)
	NotBefore     time.Time
	NotAfter      time.Time
	Lifetime      time.Duration
}

// inspectCertPEM — décode le PREMIER bloc CERTIFICATE et calcule les
// empreintes. Fonction pure : testée sans VM (tls_spki_test.go).
func inspectCertPEM(pemBytes []byte) (*certInfo, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("aucun bloc PEM trouvé")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("bloc PEM %q — CERTIFICATE attendu", block.Type)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("certificat illisible : %w", err)
	}
	spkiDER, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("SPKI indisponible : %w", err)
	}
	spkiSum := sha256.Sum256(spkiDER)
	derSum := sha256.Sum256(cert.Raw)
	sha1Sum := sha1.Sum(cert.Raw)
	return &certInfo{
		SPKISha256Hex: hex.EncodeToString(spkiSum[:]),
		Fingerprint:   colonHex(derSum[:]),
		SHA1:          strings.ToUpper(hex.EncodeToString(sha1Sum[:])),
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		Lifetime:      cert.NotAfter.Sub(cert.NotBefore),
	}, nil
}

func colonHex(b []byte) string {
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return strings.Join(parts, ":")
}

// fetchVMCert — la SEULE source acceptée : ~/.config/cgo/cert.pem sur la VM.
// (L'ancien repli `find / -maxdepth 4 -name cert.pem | xargs cat` installait
// le premier PEM venu trouvé sur le disque — refusé par construction.)
func fetchVMCert(c *Config) ([]byte, error) {
	out, err := c.SSH("cat ~/.config/cgo/cert.pem 2>/dev/null")
	if err != nil || len(out) == 0 {
		return nil, fmt.Errorf("cert.pem introuvable sur la VM — dashboard TLS démarré ? (kit svc restart après deploy)")
	}
	if !strings.Contains(out, "BEGIN CERTIFICATE") {
		return nil, fmt.Errorf("cert.pem illisible (pas un PEM CERTIFICATE)")
	}
	return []byte(out), nil
}

// certUsable — refus AVANT toute écriture : expiré ou pas encore valable.
// (La vie > 365 j n'est qu'un avertissement, pas un refus.)
func certUsable(ci *certInfo, now time.Time) error {
	if now.Before(ci.NotBefore) {
		return fmt.Errorf("certificat pas encore valable (notBefore %s)", ci.NotBefore.Format("2006-01-02 15:04:05 MST"))
	}
	if now.After(ci.NotAfter) {
		return fmt.Errorf("certificat expiré (notAfter %s)", ci.NotAfter.Format("2006-01-02 15:04:05 MST"))
	}
	return nil
}

// printCert — ce que l'opérateur doit comparer avec la VM AVANT le « y ».
func printCert(r *Runner, ci *certInfo) {
	r.out("[tls] SPKI sha256 : %s", ci.SPKISha256Hex)
	r.out("[tls] empreinte SHA-256 (DER) : %s", ci.Fingerprint)
	r.out("[tls] empreinte SHA-1 (certutil) : %s", ci.SHA1)
	r.out("[tls] validité : %s → %s (%d j)",
		ci.NotBefore.Format("2006-01-02 15:04:05 MST"),
		ci.NotAfter.Format("2006-01-02 15:04:05 MST"),
		(int)(ci.Lifetime.Hours()/24))
	if ci.Lifetime > 365*24*time.Hour {
		r.out("[tls] ATTENTION : durée de vie totale %d j > 365 j — un certificat aussi long fuit silencieusement en cas de vol de clé", (int)(ci.Lifetime.Hours()/24))
	}
}

// confirmYes — confirmation EXPLICITE « y » sur stdin ; toute autre
// entrée (Entrée, n, EOF) → faux. Jamais de défaut.
func confirmYes(r *Runner, action string) bool {
	fmt.Fprintf(r.Stdout, "[tls] %s CE certificat (empreintes ci-dessus) ? [y/N] : ", action)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

// TLS — confiance du certificat auto-signé du tableau de bord.
// meteolink.dev est préchargé HSTS : les navigateurs exigent HTTPS, et un
// certificat auto-signé déclenche l'avertissement d'autorité inconnue.
// Cette action rapatrie cert.pem DEPUIS la VM (jamais un PEM ramassé sur le
// disque local), imprime les empreintes (SPKI, DER, fenêtre), REFUSE un
// certificat expiré ou pas encore valable, exige « y » explicite sur stdin,
// puis l'installe dans le magasin local — et vérifie le HTTPS de bout en
// bout avec la confiance système (zéro -k, zéro contournement).
func (r *Runner) TLS(c *Config) int {
	// 1. rapatrier UNIQUEMENT le certificat de la VM
	pemBytes, err := fetchVMCert(c)
	if err != nil {
		r.errf("[tls] %v", err)
		return 8
	}
	ci, err := inspectCertPEM(pemBytes)
	if err != nil {
		r.errf("[tls] %v", err)
		return 8
	}
	now := time.Now()
	if err := certUsable(ci, now); err != nil {
		r.errf("[tls] REFUSÉ : %v", err)
		return 8
	}
	printCert(r, ci)
	if !confirmYes(r, "installer") {
		r.errf("[tls] refusé — confiance NON modifiée")
		return 8
	}
	out := string(pemBytes)
	tmp := filepath.Join(os.TempDir(), "meteolink-dev-cert.pem")
	if err := os.WriteFile(tmp, []byte(out), 0600); err != nil {
		r.errf("[tls] écriture locale impossible : %v", err)
		return 8
	}
	r.out("[tls] certificat rapatrié (%d octets)", len(out))
	// 2. installer dans le magasin de confiance
	if runtime.GOOS == "windows" {
		// magasin machine d'abord (tous utilisateurs, admin requis), repli
		// magasin utilisateur (utilisateur courant seul, sans admin —
		// Windows peut demander une confirmation unique à l'écran).
		if bs, err := bgCmd("certutil", "-addstore", "root", tmp).CombinedOutput(); err != nil {
			r.out("[tls] magasin machine refusé (pas admin ?) : %s", firstLine(string(bs)))
			r.out("[tls] repli : magasin de l'utilisateur courant…")
			if bs2, err2 := bgCmd("certutil", "-user", "-addstore", "root", tmp).CombinedOutput(); err2 != nil {
				r.errf("[tls] magasin utilisateur refusé : %s", firstLine(string(bs2)))
				r.errf("[tls] relancez dans un terminal ADMINISTRATEUR : cgo kit tls")
				r.errf("[tls] en attendant : accepter l'avertissement une fois dans le navigateur")
				return 8
			}
			r.out("[tls] certificat installé dans le magasin de l'utilisateur (%s) — valable pour ce compte Windows", os.Getenv("USERNAME"))
		} else {
			r.out("[tls] certificat installé dans le magasin racine Windows")
		}
	} else if os.Geteuid() == 0 {
		// root : installation directe dans le magasin système — `sudo ./cgo
		// kit tls` donne la confiance totale sans autre commande.
		dst := "/usr/local/share/ca-certificates/meteolink-dev.crt"
		data, err := os.ReadFile(tmp)
		if err != nil {
			r.errf("[tls] relecture certificat : %v", err)
			return 8
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			r.errf("[tls] écriture %s : %v", dst, err)
			return 8
		}
		if out, err := bgCmd("update-ca-certificates").CombinedOutput(); err != nil {
			r.errf("[tls] update-ca-certificates : %s", firstLine(string(out)))
			return 8
		}
		r.out("[tls] certificat installé dans le magasin système (update-ca-certificates OK)")
	} else {
		r.out("[tls] certificat : %s", tmp)
		r.out("[tls] confiance système : relancez avec sudo pour l'installer, ou à la main :")
		r.out("[tls]   sudo cp %s /usr/local/share/ca-certificates/meteolink-dev.crt && sudo update-ca-certificates", tmp)
	}
	// 3. vérification HTTPS avec confiance système (pas de -k)
	name := c.DashHost
	if name == "" {
		name = "meteolink.dev"
	}
	cl := &http.Client{Timeout: 6 * time.Second}
	resp, err := cl.Get("https://" + name + ":" + c.DashPort + "/api/health")
	if err != nil {
		r.errf("[tls] https://%s:%s injoignable avec confiance système : %v", name, c.DashPort, err)
		return 8
	}
	defer resp.Body.Close()
	var doc struct {
		OK bool `json:"ok"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&doc)
	if !doc.OK {
		r.errf("[tls] health inattendu via HTTPS")
		return 8
	}
	r.out("[tls] vérifié → https://%s:%s (chaîne de confiance OK)", name, c.DashPort)
	return 0
}

// TLSUninstall — retire l'entrée meteolink-dev du magasin local :
// Windows : certutil -delstore (magasin machine d'abord, repli utilisateur) ;
// Linux root : supprime /usr/local/share/ca-certificates/meteolink-dev.crt +
// update-ca-certificates --fresh ; sinon imprime les commandes exactes.
// L'empreinte SHA-1 sert de cible certutil ; « y » explicite exigé.
func (r *Runner) TLSUninstall(c *Config) int {
	// l'empreinte vient de la même source que l'installation (la VM) ;
	// sans réponse, on refuse plutôt que de deviner.
	pemBytes, err := fetchVMCert(c)
	if err != nil {
		r.errf("[tls-uninstall] %v", err)
		return 8
	}
	ci, err := inspectCertPEM(pemBytes)
	if err != nil {
		r.errf("[tls-uninstall] %v", err)
		return 8
	}
	printCert(r, ci)
	if !confirmYes(r, "retirer") {
		r.errf("[tls-uninstall] refusé — magasin NON modifié")
		return 8
	}
	if runtime.GOOS == "windows" {
		// magasin machine d'abord (comme l'installation), repli utilisateur.
		if bs, err := bgCmd("certutil", "-delstore", "root", ci.SHA1).CombinedOutput(); err != nil {
			r.out("[tls-uninstall] magasin machine : %s", firstLine(string(bs)))
			if bs2, err2 := bgCmd("certutil", "-user", "-delstore", "root", ci.SHA1).CombinedOutput(); err2 != nil {
				r.errf("[tls-uninstall] magasin utilisateur : %s", firstLine(string(bs2)))
				r.errf("[tls-uninstall] relancez dans un terminal ADMINISTRATEUR : cgo kit tls-uninstall")
				return 8
			}
			r.out("[tls-uninstall] certificat retiré du magasin de l'utilisateur")
			return 0
		}
		r.out("[tls-uninstall] certificat retiré du magasin racine Windows")
		return 0
	}
	const dst = "/usr/local/share/ca-certificates/meteolink-dev.crt"
	if os.Geteuid() == 0 {
		if _, err := os.Stat(dst); err != nil {
			r.out("[tls-uninstall] %s absent — magasin déjà propre", dst)
			return 0
		}
		if err := os.Remove(dst); err != nil {
			r.errf("[tls-uninstall] suppression %s : %v", dst, err)
			return 8
		}
		if out, err := bgCmd("update-ca-certificates", "--fresh").CombinedOutput(); err != nil {
			r.errf("[tls-uninstall] update-ca-certificates --fresh : %s", firstLine(string(out)))
			return 8
		}
		r.out("[tls-uninstall] %s retiré (update-ca-certificates --fresh OK)", dst)
		return 0
	}
	r.out("[tls-uninstall] Linux : relancez avec sudo, ou à la main :")
	r.out("[tls-uninstall]   sudo rm -f %s && sudo update-ca-certificates --fresh", dst)
	return 8
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' || c == '\r' {
			return s[:i]
		}
	}
	return s
}
