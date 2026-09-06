package kit

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TLS — confiance du certificat auto-signé du tableau de bord.
// meteolink.dev est préchargé HSTS : les navigateurs exigent HTTPS, et un
// certificat auto-signé déclenche l'avertissement d'autorité inconnue.
// Cette action rapatrie cert.pem depuis la VM et l'installe dans le magasin
// de confiance local (admin requis sous Windows) puis vérifie le HTTPS
// de bout en bout avec la confiance système — zéro -k, zéro contournement.
func (r *Runner) TLS(c *Config) int {
	// 1. rapatrier le certificat depuis la VM
	out, err := c.SSH("cat ~/.config/cgo/cert.pem 2>/dev/null || find / -maxdepth 4 -name cert.pem -path '*cgo*' 2>/dev/null | head -n 1 | xargs cat 2>/dev/null")
	if err != nil || len(out) == 0 || !strings.Contains(out, "BEGIN CERTIFICATE") {
		r.errf("[tls] cert.pem introuvable sur la VM — dashboard TLS démarré ? (kit svc restart après deploy)")
		return 8
	}
	tmp := filepath.Join(os.TempDir(), "meteolink-dev-cert.pem")
	if err := os.WriteFile(tmp, []byte(out), 0600); err != nil {
		r.errf("[tls] écriture locale impossible : %v", err)
		return 8
	}
	r.out("[tls] certificat rapatrié (%d octets)", len(out))
	// 2. installer dans le magasin de confiance
	if runtime.GOOS == "windows" {
		cmd := exec.Command("certutil", "-addstore", "root", tmp)
		if bs, err := cmd.CombinedOutput(); err != nil {
			r.errf("[tls] magasin de confiance refusé : %s", firstLine(string(bs)))
			r.errf("[tls] relancez dans un terminal ADMINISTRATEUR : cgo kit tls")
			r.errf("[tls] en attendant : accepter l'avertissement une fois dans le navigateur")
			return 8
		}
		r.out("[tls] certificat installé dans le magasin racine Windows")
	} else {
		r.out("[tls] certificat : %s", tmp)
		r.out("[tls] confiance système : copiez-le dans votre magasin local")
		r.out("[tls]   Debian/Ubuntu : sudo cp %s /usr/local/share/ca-certificates/meteolink-dev.crt && sudo update-ca-certificates", tmp)
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

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' || c == '\r' {
			return s[:i]
		}
	}
	return s
}
