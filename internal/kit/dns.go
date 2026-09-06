package kit

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

// DNS — meteolink.dev pointe vers la VM dans le fichier hosts local.
// Le nom stable par défaut (dashboard_host) plutôt que l'IP DHCP qui change
// à chaque bail : l'opérateur ouvre http://meteolink.dev:9090, pas une IP.
// Écriture admin requise sous Windows : en cas de refus, la ligne exacte à
// ajouter est affichée (bloc-notes admin), échec honnête, pas de silence.
func (r *Runner) DNS(c *Config) int {
	ip := c.SSHHost
	if ip == "" || ip == "auto" {
		r.errf("[dns] hôte SSH inconnu — lancez d'abord : cgo kit ensure")
		return 5
	}
	name := c.DashHost
	if name == "" {
		name = "meteolink.dev"
	}
	// déjà bon ? ne rien toucher
	if addrs, err := net.LookupHost(name); err == nil {
		for _, a := range addrs {
			if a == ip {
			r.out("[dns] %s → %s déjà mappé", name, ip)
			r.out("[dns] tableau de bord : https://%s:%s", name, c.DashPort)
			return 0
			}
		}
	}
	path := "/etc/hosts"
	if runtime.GOOS == "windows" {
		sys := os.Getenv("SystemRoot")
		if sys == "" {
			sys = `C:\Windows`
		}
		path = sys + `\System32\drivers\etc\hosts`
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		r.errf("[dns] lecture %s impossible : %v", path, err)
		return 8
	}
	var kept []string
	for _, ln := range strings.Split(string(raw), "\n") {
		trim := strings.TrimSpace(strings.TrimSuffix(ln, "\r"))
		if trim == "" || strings.HasPrefix(trim, "#") {
			kept = append(kept, ln)
			continue
		}
		fields := strings.Fields(trim)
		hit := false
		for _, f := range fields[1:] {
			if strings.EqualFold(f, name) {
				hit = true
			}
		}
		if hit {
			continue // entrée périmée (vieille IP) : remplacée ci-dessous
		}
		kept = append(kept, ln)
	}
	kept = append(kept, fmt.Sprintf("%s %s  # cgo — tableau de bord (kit dns)", ip, name))
	if err := os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0644); err != nil {
		r.errf("[dns] écriture %s refusée : %v", path, err)
		if runtime.GOOS == "windows" {
			r.errf("[dns] relancez dans un terminal ADMINISTRATEUR, ou ajoutez à la main :")
		} else {
			r.errf("[dns] relancez avec sudo, ou ajoutez à la main :")
		}
		r.errf("[dns]   %s %s", ip, name)
		return 8
	}
	// vérification : le nom doit résoudre vers l'IP
	if addrs, err := net.LookupHost(name); err != nil {
		r.errf("[dns] écrit mais %s ne résout pas (%v) — cache DNS ? ipconfig /flushdns", name, err)
		return 8
	} else {
		ok := false
		for _, a := range addrs {
			if a == ip {
				ok = true
			}
		}
		if !ok {
			r.errf("[dns] %s résout vers %v, pas %s — entrée concurrente ?", name, addrs, ip)
			return 8
		}
	}
	r.out("[dns] %s → %s mappé et vérifié", name, ip)
	r.out("[dns] tableau de bord : https://%s:%s", name, c.DashPort)
	return 0
}
