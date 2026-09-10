package campagne

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Realms4239/cgo/pkg/probe"
)

// PlaneReady — le banc de mesure répond-il ? Sonde l'objet small ET le
// sink bulk avec des budgets courts (refus en < 7 s pire cas).
//
// Pourquoi : sur VM fraîche (pas de veth/netns/testbedsrv, pas de sudoers),
// chaque cellule échoue et la matrice — qui n'abandonne jamais — gèle un
// run à ZÉRO ligne : « Démarrer → idle → rien », sans un mot. Refus
// honnête plutôt que run fantôme (l'erreur remonte en 409 au dashboard).
func PlaneReady(ctx context.Context, d Deps) error {
	defaults(&d)
	ctx2, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	if _, err := probe.SmallObject(ctx2, probe.SmallClientTimeout(4*time.Second), d.SmallURL); err != nil {
		return fmt.Errorf("banc de mesure injoignable (%s : %v) — sur la VM : sudo bash kit/testbed.sh up — ou depuis le poste : cgo kit testbed up", d.SmallURL, err)
	}
	conn, err := (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx2, "tcp", d.BulkAddr)
	if err != nil {
		return fmt.Errorf("sink bulk injoignable (%s : %v) — sur la VM : sudo bash kit/testbed.sh up — ou depuis le poste : cgo kit testbed up", d.BulkAddr, err)
	}
	_ = conn.Close()
	return nil
}
