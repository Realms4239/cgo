package kit

import (
	"crypto/tls"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ShipCheck — porte d'embarquement avant release : chaque maillon de la
// chaîne navigateur est vérifié dans l'ordre (DNS → TCP → TLS → certificat
// → health → confiance OS) avec le remède exact en cas d'échec. Sortie 0
// ssi tout est vert : un go ici signifie que https://meteolink.dev:9090
// s'ouvre SANS avertissement. Le navigateur ne pardonne ni le HSTS ni
// l'auto-signé non installé — cette porte non plus.
func (r *Runner) ShipCheck(c *Config) int {
	name := c.DashHost
	if name == "" {
		name = "meteolink.dev"
	}
	host := c.SSHHost
	port := c.DashPort
	fail := 0
	gate := func(ok bool, what, detail, remedy string) {
		if ok {
			r.out("  [ok] %-10s %s", what, detail)
			return
		}
		fail++
		r.errf("  [KO] %-10s %s", what, detail)
		if remedy != "" {
			r.errf("       → %s", remedy)
		}
	}

	// 1. DNS : le nom stable résout vers la VM
	addrs, err := net.LookupHost(name)
	dnsOK := err == nil && len(addrs) > 0
	dnsDetail := strings.Join(addrs, ",")
	if !dnsOK {
		dnsDetail = fmt.Sprintf("ne résout pas (%v)", err)
	}
	gate(dnsOK, "dns", name+" → "+dnsDetail, "cgo kit dns (terminal admin)")

	// 2. TCP : le port répond
	tcpOK := portOpen(host, port)
	gate(tcpOK, "tcp", host+":"+port+(map[bool]string{true: " ouvert", false: " fermé"})[tcpOK],
		"cgo kit ensure (VM éteinte ?) puis cgo kit svc start")

	// 3-4. TLS + certificat : poignée de main + SAN + validité
	var leafOK, sanOK, dateOK bool
	var cn, sans, dates string
	if tcpOK {
		d := &net.Dialer{Timeout: 5 * time.Second}
		conn, err := tls.DialWithDialer(d, "tcp", host+":"+port, &tls.Config{InsecureSkipVerify: true, ServerName: name})
		if err == nil {
			defer conn.Close()
			st := conn.ConnectionState()
			if len(st.PeerCertificates) > 0 {
				leafOK = true
				leaf := st.PeerCertificates[0]
				cn = leaf.Subject.CommonName
				var parts []string
				parts = append(parts, leaf.DNSNames...)
				for _, ip := range leaf.IPAddresses {
					parts = append(parts, ip.String())
				}
				sans = strings.Join(parts, ", ")
				for _, n := range leaf.DNSNames {
					if strings.EqualFold(n, name) {
						sanOK = true
					}
				}
				dateOK = time.Until(leaf.NotAfter) > 30*24*time.Hour
				dates = leaf.NotBefore.Format("2006-01-02") + " → " + leaf.NotAfter.Format("2006-01-02")
			}
		}
	}
	gate(leafOK, "tls", map[bool]string{true: "poignée de main OK", false: "échec handshake"}[leafOK],
		"cgo kit svc restart (dashboard non-TLS ? vieux binaire ?)")
	gate(sanOK, "cert-san", map[bool]string{true: "couvre " + name + " [" + sans + "]", false: "SANs [" + sans + "] ≠ " + name}[sanOK],
		"certificat régénéré nécessaire (supprimer ~/.config/cgo/cert.pem sur la VM puis kit svc restart)")
	gate(dateOK, "cert-dates", map[bool]string{true: dates + " (CN=" + cn + ")", false: "expiré/bientôt : " + dates}[dateOK],
		"même remède que cert-san")

	// 5. health applicatif
	hOK := c.Health()
	gate(hOK, "health", map[bool]string{true: "/api/health ok:true", false: "injoignable ou ok:false"}[hOK],
		"cgo kit logs, puis cgo kit svc restart")

	// 6. confiance OS : SANS elle, HSTS + auto-signé = page d'erreur
	// infranchissable (pas de contournement au clic). C'est LA porte qui
	// a bloqué la 1.2.3.
	trustOK, trustDetail := checkOSTrust(name)
	gate(trustOK, "confiance", trustDetail, "cgo kit tls DANS UN TERMINAL ADMINISTRATEUR, puis rouvrir l'onglet")

	if fail > 0 {
		r.errf("[shipcheck] %d porte(s) fermée(s) — release BLOQUÉE, corrigez ci-dessus", fail)
		return 8
	}
	r.out("[shipcheck] tout est vert — https://%s:%s s'ouvre sans avertissement, on peut shipper", name, port)
	return 0
}

// checkOSTrust — le certificat meteolink.dev est-il dans un magasin racine ?
// Magasin machine d'abord (tous utilisateurs), puis magasin utilisateur
// (compte courant, posé par `kit tls` sans admin) — les deux sont honorés
// par les navigateurs Chromium.
func checkOSTrust(name string) (bool, string) {
	if runtime.GOOS != "windows" {
		return false, "vérification manuelle hors Windows (magasin système)"
	}
	if out, err := exec.Command("certutil", "-store", "root").CombinedOutput(); err == nil &&
		strings.Contains(strings.ToLower(string(out)), strings.ToLower(name)) {
		return true, name + " présent dans le magasin racine Windows (machine)"
	}
	if out, err := exec.Command("certutil", "-user", "-store", "root").CombinedOutput(); err == nil &&
		strings.Contains(strings.ToLower(string(out)), strings.ToLower(name)) {
		return true, name + " présent dans le magasin racine Windows (utilisateur courant)"
	}
	return false, name + " ABSENT des magasins racines — le navigateur refusera (HSTS, pas de contournement)"
}
