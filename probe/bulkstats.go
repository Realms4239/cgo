package probe

import (
	"context"
	"net"
	"sync"
)

// BulkStats — charge multi-flux avec la vérité par flux ET le ratio de
// réémission : JFI se calcule sur Bytes, le modèle A sur RetransBytes.
type BulkStats struct {
	Bytes           []uint64 `json:"bytes"`
	RetransBytes    uint64   `json:"retrans_bytes"`
	RetransMeasured bool     `json:"retrans_measured"`
}

// keepConn — enveloppe qui retient Close : les primitives de charge
// (BulkSend, BulkReceive au fond de BulkRecvSource) ferment la connexion en
// rendant la main, mais l'échantillon TCP_INFO exige un fd VIVANT (File()
// sur un socket fermé rend « use of closed network connection » — vérifié :
// poll.FD.Dup incref échoue après Close). bulkNStats neutralise donc la
// fermeture pendant la charge, échantillonne, puis ferme la vraie connexion.
type keepConn struct{ net.Conn }

func (keepConn) Close() error { return nil }

// bulkNStats — moteur commun : N connexions, échantillonnage TCP_INFO à la
// fermeture de chaque flux (le compteur est cumulatif par socket).
func bulkNStats(ctx context.Context, dial func(ctx context.Context) (net.Conn, error), send func(ctx context.Context, c net.Conn) (uint64, error), n int) (BulkStats, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	st := BulkStats{Bytes: make([]uint64, 0, n)}
	oneTCP := false // au moins un socket échantillonné => measured
	var firstErr error
	for i := 0; i < n; i++ {
		c, err := dial(ctx)
		if err != nil {
			continue // lien en défaut : les flux vivants comptent
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			b, err := send(ctx, keepConn{c}) // fermeture retenue, pas perdue
			rb, rerr := TCPRetransBytes(c)   // socket encore vivant : l'échantillon est vrai
			c.Close()
			mu.Lock()
			st.Bytes = append(st.Bytes, b)
			st.RetransBytes += rb
			if rerr == nil {
				oneTCP = true
			}
			// vraie panne d'écriture : première erreur remonte (une fin de
			// fenêtre normale est déjà nil côté send)
			if err != nil && firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}(c)
	}
	wg.Wait()
	st.RetransMeasured = oneTCP
	return st, firstErr
}

// BulkSendNStats — charge MONTANTE multi-flux avec statistiques.
func BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkSend(ctx, c) },
		n)
}

// BulkDownloadNStats — charge DESCENDANTE multi-flux avec statistiques.
func BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkRecvSource(ctx, c, cc) },
		n)
}
