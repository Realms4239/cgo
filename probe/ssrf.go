// Garde SSRF — le petit objet et le bulk partent du processus d'audit vers
// une URL/adresse décidées par l'appelant (dashboard, corpus). Sans cette
// garde, un appel d'API peut viser le loopback ou un RFC1918 (métadonnées
// cloud, services internes) depuis la VM du banc.
//
// Portée : loopback, RFC1918, link-local (IPv4/IPv6), ULA IPv6, non spécifié.
// L'opt-in banc CGO_ALLOW_PRIVATE_TARGETS=1 autorise les cibles privées
// (10.30.0.2 du banc dualns) ; sans lui, tout privé est refusé.
//
// TOCTOU : les IP résolues sont vérifiées AVANT la connexion et la connexion
// part sur les IP vérifiées (jamais de re-résolution entre le contrôle et le
// dial). Échec de résolution = refus (fail-closed).
package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"syscall"
	"time"
)

// AllowPrivateTargets lit l'opt-in banc à l'instant présent (les tests
// posent CGO_ALLOW_PRIVATE_TARGETS via t.Setenv).
func AllowPrivateTargets() bool { return os.Getenv("CGO_ALLOW_PRIVATE_TARGETS") == "1" }

// PrivateTargetError nomme l'IP et la plage refusées — jamais un repli muet.
type PrivateTargetError struct {
	Host  string
	IP    string
	Range string
}

func (e *PrivateTargetError) Error() string {
	return fmt.Sprintf("cible privée refusée (%s) : %s — hors portée sans CGO_ALLOW_PRIVATE_TARGETS=1", e.Range, e.IP)
}

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("255.255.255.255/32"),
	netip.MustParsePrefix("::/128"), netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("fec0::/10"), netip.MustParsePrefix("ff00::/8"),
}

// IsBlockedAddr — garde unique : Unmap() d'abord, donc ::ffff:127.0.0.1 est
// bien le loopback v4. Aucune adresse non publique ne sort de ce predicate.
func IsBlockedAddr(addr netip.Addr) bool {
	a := addr.Unmap()
	for _, p := range blockedPrefixes {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// dialControl — re-valide l'IP exacte au moment du connect (TOCTOU) avant de
// déléguer au contrôle appelant (marquage DSCP / congruence TCP).
func dialControl(allowPrivate bool, control func(network, address string, c syscall.RawConn) error) func(string, string, syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		ap, err := netip.ParseAddrPort(address)
		if err != nil {
			return fmt.Errorf("adresse de connexion %q illisible — refusée : %w", address, err)
		}
		if !allowPrivate && IsBlockedAddr(ap.Addr()) {
			return &PrivateTargetError{Host: ap.Addr().String(), IP: ap.Addr().String(), Range: "plage réservée/non publique"}
		}
		if control != nil {
			return control(network, address, c)
		}
		return nil
	}
}

// blockedIP — adapte la garde netip aux appelants net.IP existants (nom de
// plage générique). La décision de blocage vient uniquement de IsBlockedAddr.
func blockedIP(ip net.IP) (string, bool) {
	if ip == nil {
		return "adresse absente", true
	}
	a, ok := netip.AddrFromSlice(ip)
	if !ok {
		return "forme non reconnue", true
	}
	if IsBlockedAddr(a) {
		return "plage réservée ou non publique", true
	}
	return "", false
}

// GuardHost vérifie que host (IP littérale ou nom à résoudre) ne vise pas une
// adresse privée/spéciale. host vide = erreur. Résolution impossible = refus.
func GuardHost(ctx context.Context, host string, allowPrivate bool) error {
	if host == "" {
		return fmt.Errorf("hôte vide — refusé")
	}
	if ip := net.ParseIP(host); ip != nil {
		if r, blocked := blockedIP(ip); blocked && !allowPrivate {
			return &PrivateTargetError{Host: host, IP: ip.String(), Range: r}
		}
		return nil
	}
	rctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(rctx, "ip", host)
	if err != nil {
		return fmt.Errorf("résolution de %q impossible — refusé : %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("résolution de %q : aucune adresse — refusé", host)
	}
	for _, ip := range ips {
		if r, blocked := blockedIP(ip); blocked && !allowPrivate {
			return &PrivateTargetError{Host: host, IP: ip.String(), Range: r}
		}
	}
	return nil
}

// GuardURL vérifie le schéma (http|https uniquement) et l'hôte d'une petite-
// objet URL. Toute autre forme → refus (file://, gopher://, …).
func GuardURL(ctx context.Context, raw string, allowPrivate bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL illisible — refusée : %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("schéma %q refusé — http|https uniquement", u.Scheme)
	}
	return GuardHost(ctx, u.Hostname(), allowPrivate)
}

// guardedDialContext résout, vérifie puis compose la connexion sur les IP
// vérifiées (contrôle optionnel via net.Dialer.Control — marquage DSCP).
func guardedDialContext(ctx context.Context, network, addr string, allowPrivate bool, control func(network, address string, c syscall.RawConn) error) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("adresse %q : host:port attendu — refusée", addr)
	}
	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else {
		rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		ips, err = net.DefaultResolver.LookupIP(rctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("résolution de %q impossible — refusé : %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("résolution de %q : aucune adresse — refusé", host)
		}
	}
	for _, ip := range ips {
		if r, blocked := blockedIP(ip); blocked && !allowPrivate {
			return nil, &PrivateTargetError{Host: host, IP: ip.String(), Range: r}
		}
	}
	d := &net.Dialer{Timeout: 5 * time.Second, Control: dialControl(allowPrivate, control)}
	var lastErr error
	for _, ip := range ips {
		conn, err := d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// NewSafeTransport — transport HTTP dont CHAQUE connexion passe par la garde.
// allowPrivate=true : opt-in banc (10.30.0.2), tout le reste inchangé.
func NewSafeTransport(allowPrivate bool) *http.Transport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return guardedDialContext(ctx, network, addr, allowPrivate, nil)
	}
	return base
}
