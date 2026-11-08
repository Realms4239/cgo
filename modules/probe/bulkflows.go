package probe

import (
	"context"
	"net"
	"sync"

	"github.com/Realms4239/cgo/model"
)

// FlowResult — un flux de charge : octets envoyés. Per-flow, pas total seul :
// l'équité inter-flux (JFI, Jain) se calcule sur les contributions
// individuelles — un total identique peut cacher la famine d'un flux.
type FlowResult struct {
	Bytes uint64
}

// BulkSendNFlows — N connexions inondent en parallèle jusqu'au terme du
// ctx (fenêtre de charge). Chaque flux = un BulkSend indépendant avec sa
// propre échéance d'écriture ; les résultats arrivent per-flow.
// flux réels concurrents (lien partagé) : l'isolation que fq_codel/cake
// promettent devient mesurable au lieu d'augurée.
func BulkSendNFlows(ctx context.Context, conns []net.Conn, _ int) ([]FlowResult, error) {
	var wg sync.WaitGroup
	results := make([]FlowResult, len(conns))
	for i, c := range conns {
		wg.Add(1)
		go func(i int, c net.Conn) {
			defer wg.Done()
			n, _ := BulkSend(ctx, c) // BulkSend ferme c et rend les octets
			results[i] = FlowResult{Bytes: n}
		}(i, c)
	}
	wg.Wait()
	return results, nil
}

// BulkSendNTo — N dials indépendants vers addr (même CC), inondation
// parallèle, per-flow rendu. Utilisé par la campagne multi-flux. `cc` est le
// contrôle de congestion de la cellule, pas un défaut d'hôte : un multi-flux
// BBR doit utiliser BBR, sinon la comparaison AQM/CC est faussée.
func BulkSendNTo(ctx context.Context, addr, cc string, n int) ([]uint64, error) {
	if cc == "" {
		cc = string(model.Cubic)
	}
	if n <= 1 {
		b, err := BulkSendTo(ctx, addr, cc)
		return []uint64{b}, err
	}
	conns := make([]net.Conn, 0, n)
	for i := 0; i < n; i++ {
		c, err := DialWithCC(ctx, addr, cc)
		if err != nil {
			continue // lien en défaut : les flux vivants comptent, pas l'échec total
		}
		conns = append(conns, c)
	}
	res, err := BulkSendNFlows(ctx, conns, len(conns))
	out := make([]uint64, len(res))
	for i, r := range res {
		out[i] = r.Bytes
	}
	return out, err
}
