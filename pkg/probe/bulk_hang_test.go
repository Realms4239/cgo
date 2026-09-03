package probe

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

// deadlineConn enregistre les échéances posées (net.Pipe les honore).
type deadlineConn struct {
	net.Conn
	mu        sync.Mutex
	deadlines []time.Time
}

func (c *deadlineConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	c.deadlines = append(c.deadlines, t)
	c.mu.Unlock()
	return c.Conn.SetWriteDeadline(t)
}

// TestBulkSendBlockedConnReturns — récepteur figé, aucun lecteur : BulkSend
// doit revenir à l'échéance d'écriture (~2 s), pas pendre la cellule.
// (Sans échéance, ce test pend — le prouver avec -timeout.)
func TestBulkSendBlockedConnReturns(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close() // jamais lu : chaque Write bloquerait sans échéance
	wrapped := &deadlineConn{Conn: client}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	done := make(chan uint64, 1)
	go func() {
		n, _ := BulkSend(ctx, wrapped)
		done <- n
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("BulkSend hung on blocked conn — no write deadline")
	}
	if d := time.Since(start); d > 8*time.Second {
		t.Fatalf("BulkSend took %v on blocked conn, want ~2s deadline", d)
	}
	wrapped.mu.Lock()
	nd := len(wrapped.deadlines)
	wrapped.mu.Unlock()
	if nd == 0 {
		t.Fatal("no write deadline set before blocking Write")
	}
}
