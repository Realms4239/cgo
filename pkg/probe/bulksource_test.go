package probe

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

// TestBulkSourceServesDownload — le mode source : la source inonde le
// client (download). Le client reçoit via BulkReceive et compte les octets ;
// la source rend ses octets envoyés à la fin du contexte. Les deux comptes
// doivent converger (ce qui part est reçu).
func TestBulkSourceServesDownload(t *testing.T) {
	src, dst := net.Pipe()
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	received := make(chan uint64, 1)
	go func() {
		n, _ := BulkReceive(ctx, dst, nil)
		received <- n
	}()

	sent, err := BulkSource(ctx, src)
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	if sent == 0 {
		t.Fatal("source n'a rien envoyé")
	}
	select {
	case n := <-received:
		if n < sent/2 {
			t.Fatalf("reçu %d << envoyé %d — le download ne passe pas", n, sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client n'a jamais fini de recevoir")
	}
	_ = io.Discard
}
