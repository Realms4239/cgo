package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// authTokenFile — jeton d'authentification du banc, 0600 root:root, posé par
// `cgo service install` et lu via EnvironmentFile par le unit (jamais en ligne).
const authTokenFile = "/etc/cgo/auth.token"

// checkExposure — porte de démarrage : un bind non-loopback exige --expose
// ET un token (CGO_API_TOKEN ou le fichier jeton). Loopback toujours permis ;
// adresse illisible ou nom non résolu : refusé (fail closed).
func checkExposure(addr string, expose bool, getenv func(string) string, tokenFile string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("serve: adresse %q illisible (hôte:port attendu) : %w", addr, err)
	}
	if isLoopbackHost(host) {
		return nil
	}
	if !expose {
		return fmt.Errorf("serve: écoute sur %q refusée — hors loopback sans --expose.\nUtilisez 127.0.0.1 (dev) ou --expose avec un token : CGO_API_TOKEN, ou %s (0600)", addr, tokenFile)
	}
	if getenv("CGO_API_TOKEN") == "" && !tokenFilePresent(tokenFile) {
		return fmt.Errorf("serve: --expose sur %q exige un token — CGO_API_TOKEN vide et aucun %s.\nTLS seul n'authentifie personne : générez le jeton (cgo service install) ou exportez CGO_API_TOKEN", addr, tokenFile)
	}
	return nil
}

// isLoopbackHost — 127.0.0.0/8, ::1, localhost ; un nom est tranché par
// résolution : toutes les IP résolues doivent être loopback, sinon refusé.
func isLoopbackHost(host string) bool {
	h := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if h == "localhost" {
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}
	ips, err := net.LookupIP(h)
	if err != nil || len(ips) == 0 {
		return false // nom non résolu → fail closed
	}
	for _, ip := range ips {
		if !ip.IsLoopback() {
			return false
		}
	}
	return true
}

func tokenFilePresent(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir() && st.Size() > 0
}

// ensureAuthTokenFile — génère 32 octets aléatoires hex dans un fichier 0600,
// root:root, si absent ; n'écrase jamais un jeton existant. Le contenu est
// une ligne CGO_API_TOKEN=<hex> ; le jeton n'est jamais affiché.
func ensureAuthTokenFile(path string) error {
	if tokenFilePresent(path) {
		return nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("génération du jeton : %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("création de %s : %w", filepath.Dir(path), err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("création de %s : %w", path, err)
	}
	if _, err := f.WriteString("CGO_API_TOKEN=" + hex.EncodeToString(b) + "\n"); err != nil {
		_ = f.Close()
		return fmt.Errorf("écriture de %s : %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("fermeture de %s : %w", path, err)
	}
	// root:root — best-effort : l'installation tourne en root, un chown
	// manqué (déjà root, FS exotique) n'ouvre pas le fichier.
	_ = os.Chown(path, 0, 0)
	return nil
}
