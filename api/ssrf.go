// Garde SSRF côté API — le petit objet (small_url) et le puits bulk
// (bulk_addr) sont décidés par l'appelant : sans garde, un POST /api/audit
// vise le loopback ou un RFC1918 depuis la VM du banc (métadonnées cloud,
// services internes). Réponse : 403 avec erreur nommée ; un argument malformé
// reste un 400 (erreur client), jamais un 200.
//
// Opt-in banc CGO_ALLOW_PRIVATE_TARGETS=1 : autorise les cibles privées
// (10.30.0.2) ; sans lui, tout privé est refusé. Résolution impossible =
// refus (fail-closed) : une cible qu'on ne peut pas prouver publique est
// refusée. La décision de plage est déléguée à probe.GuardHost : une seule
// table de plages, dans probe/ssrf.go, pour le banc et l'API.
package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/Realms4239/cgo/probe"
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

// checkHost — délègue la décision à probe.GuardHost (table de plages unique).
// L'erreur probe est traduite dans les types d'erreur API pour que le handler
// garde son mapping 403/400 ; allowPrivate ouvre l'opt-in banc.
func checkHost(ctx context.Context, host string, allowPrivate bool) error {
	if host == "" {
		return &ssrfRefusedError{Why: "hôte vide — refusé"}
	}
	if err := probe.GuardHost(ctx, host, allowPrivate); err != nil {
		var blocked *probe.PrivateTargetError
		if errors.As(err, &blocked) {
			return &privateTargetError{IP: blocked.IP, Range: blocked.Range}
		}
		return &ssrfRefusedError{Why: err.Error()}
	}
	return nil
}

// guardSmallURL — schéma http|https uniquement, puis l'hôte (garde host).
func guardSmallURL(ctx context.Context, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return &ssrfArgError{Why: fmt.Sprintf("small_url illisible — refusée : %v", err)}
	}
	// Refuser l'userinfo : des identifiants embarqués déguisent l'hôte réel
	// et peuvent induire le transport en erreur (erreur client → 400).
	if u.User != nil {
		return &ssrfArgError{Why: "small_url avec userinfo refusée"}
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
