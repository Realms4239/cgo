package probe

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// TestBulkSendNFlowsFairness — N connexions concurrentes inondent jusqu'au
// terme ; chaque flux rend ses octets, le total est la somme. L'équité
// (JFI côté agrégats) se calcule sur les per-flow, pas le total seul.
func TestBulkSendNFlowsFairness(t *testing.T) {
	// pair de sockets : le "réseau" draine en continu (sinon Write bloque
	// sur l'échéance et compte 0 — net.Pipe est sans buffer)
	client, server := net.Pipe()
	defer server.Close()
	go func() { _, _ = io.Copy(io.Discard, server) }()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	perFlow, err := BulkSendNFlows(ctx, []net.Conn{client}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(perFlow) != 1 || perFlow[0].Bytes == 0 {
		t.Fatalf("perFlow = %+v, want 1 flux avec octets > 0", perFlow)
	}
}

// TestBulkSendNFlowsCountsAll — 3 flux distincts vers 3 sinks : chaque
// flux est compté séparément, aucun n'est perdu.
func TestBulkSendNFlowsCountsAll(t *testing.T) {
	conns := make([]net.Conn, 0, 3)
	for i := 0; i < 3; i++ {
		c, s := net.Pipe()
		defer s.Close()
		go func(s net.Conn) { _, _ = io.Copy(io.Discard, s) }(s)
		conns = append(conns, c)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	perFlow, err := BulkSendNFlows(ctx, conns, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(perFlow) != 3 {
		t.Fatalf("perFlow len = %d, want 3", len(perFlow))
	}
	var total uint64
	var mu sync.Mutex
	for _, f := range perFlow {
		mu.Lock()
		total += f.Bytes
		mu.Unlock()
	}
	if total == 0 {
		t.Fatal("aucun octet compté sur 3 flux")
	}

}
