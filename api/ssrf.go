// Garde SSRF côté API — le petit objet (small_url) et le puits bulk
// (bulk_addr) sont décidés par l'appelant : sans garde, un POST /api/audit
// vise le loopback ou un RFC1918 depuis la VM du banc (métadonnées cloud,
// services internes). Réponse : 403 avec erreur nommée ; un argument malformé
// reste un 400 (erreur client), jamais un 200.
//
// Opt-in banc CGO_ALLOW_PRIVATE_TARGETS=1 : autorise les cibles privées
// (10.30.0.2) ; sans lui, tout privé est refusé. Résolution impossible =
// refus (fail-closed) : une cible qu'on ne peut pas prouver publique est
// refusée. La même logique vit dans le module probe (probe/ssrf.go) —
// dupliquée ici car le module api lie le probe PUBLIÉ, pas l'arbre local.
package api

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

// privateTargetError nomme la plage refusée — jamais un repli muet (→ 403).
type privateTargetError struct {
	IP    string
	Range string
}

func (e *privateTargetError) Error() string {
	return fmt.Sprintf("cible privée refusée (%s) : %s — hors portée sans CGO_ALLOW_PRIVATE_TARGETS=1", e.Range, e.IP)
}

// ssrfRefusedError — résolution impossible ou hôte vide : fail-closed (→ 403).
type ssrfRefusedError struct{ Why string }

func (e *ssrfRefusedError) Error() string { return e.Why }

// ssrfArgError — URL illisible, schéma refusé, forme host:port : erreur
// client (→ 400), jamais un déguisement en 403.
type ssrfArgError struct{ Why string }

func (e *ssrfArgError) Error() string { return e.Why }

func allowPrivateTargets() bool { return os.Getenv("CGO_ALLOW_PRIVATE_TARGETS") == "1" }

// blockedIP — (plage nommée, true) si ip est privée/spéciale.
func blockedIP(ip net.IP) (string, bool) {
	if ip.IsLoopback() {
		return "loopback", true
	}
	if ip.IsUnspecified() {
		return "non spécifié", true
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return "RFC1918 (10.0.0.0/8)", true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return "RFC1918 (172.16.0.0/12)", true
		case ip4[0] == 192 && ip4[1] == 168:
			return "RFC1918 (192.168.0.0/16)", true
		case ip4[0] == 169 && ip4[1] == 254:
			return "link-local (169.254.0.0/16)", true
		}
		return "", false
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return "", false
	}
	if ip6[0] == 0xfe && ip6[1]&0xc0 == 0x80 {
		return "link-local (fe80::/10)", true
	}
	if ip6[0]&0xfe == 0xfc { // fc00::/7
		return "ULA (fc00::/7)", true
	}
	return "", false
}

// checkHost — IP littérale vérifiée directement ; nom résolu sous délai court,
// fail-closed si la résolution échoue. allowPrivate ouvre l'opt-in banc.
func checkHost(ctx context.Context, host string, allowPrivate bool) error {
	if host == "" {
		return &ssrfRefusedError{Why: "hôte vide — refusé"}
	}
	if ip := net.ParseIP(host); ip != nil {
		if r, blocked := blockedIP(ip); blocked && !allowPrivate {
			return &privateTargetError{IP: ip.String(), Range: r}
		}
		return nil
	}
	rctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(rctx, "ip", host)
	if err != nil {
		return &ssrfRefusedError{Why: fmt.Sprintf("résolution de %q impossible — refusé : %v", host, err)}
	}
	if len(ips) == 0 {
		return &ssrfRefusedError{Why: fmt.Sprintf("résolution de %q : aucune adresse — refusé", host)}
	}
	for _, ip := range ips {
		if r, blocked := blockedIP(ip); blocked && !allowPrivate {
			return &privateTargetError{IP: ip.String(), Range: r}
		}
	}
	return nil
}

// guardSmallURL — schéma http|https uniquement, puis l'hôte (garde host).
func guardSmallURL(ctx context.Context, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return &ssrfArgError{Why: fmt.Sprintf("small_url illisible — refusée : %v", err)}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return &ssrfArgError{Why: fmt.Sprintf("schéma %q refusé — http|https uniquement", u.Scheme)}
	}
	return checkHost(ctx, u.Hostname(), allowPrivateTargets())
}

// guardBulkAddr — forme host:port stricte, puis l'hôte (garde host).
func guardBulkAddr(ctx context.Context, raw string) error {
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		return &ssrfArgError{Why: fmt.Sprintf("bulk_addr %q : host:port attendu — refusé", raw)}
	}
	return checkHost(ctx, host, allowPrivateTargets())
}
