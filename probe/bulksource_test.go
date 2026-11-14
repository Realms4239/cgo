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

// TestBulkSourceWriteErrorPropagates — miroir de BulkSend : panne d'écriture
// (ctx vivant) = erreur, jamais une mesure silencieuse.
func TestBulkSourceWriteErrorPropagates(t *testing.T) {
	c, _ := net.Pipe()
	c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	total, err := BulkSource(ctx, c)
	if err == nil {
		t.Fatalf("write error masqué (total=%d)", total)
	}
}

// TestBulkSourceCancelledCtxNoError — fin de fenêtre = nil.
func TestBulkSourceCancelledCtxNoError(t *testing.T) {
	c, _ := net.Pipe()
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BulkSource(ctx, c); err != nil {
		t.Fatalf("ctx annulé rendu en erreur: %v", err)
	}
}

// TestParseDownloadHello — "D" historique (CC hôte), "D:<cc>" étiqueté,
// bruit rejeté.
func TestParseDownloadHello(t *testing.T) {
	if dl, cc := ParseDownloadHello([]byte("D")); !dl || cc != "" {
		t.Fatalf("D: dl=%v cc=%q", dl, cc)
	}
	if dl, cc := ParseDownloadHello([]byte("D:bbr\n")); !dl || cc != "bbr" {
		t.Fatalf("D:bbr: dl=%v cc=%q", dl, cc)
	}
	if dl, cc := ParseDownloadHello([]byte("d:cubic")); !dl || cc != "cubic" {
		t.Fatalf("d:cubic: dl=%v cc=%q", dl, cc)
	}
	if dl, _ := ParseDownloadHello([]byte("GET / HTTP/1.0")); dl {
		t.Fatal("bruit HTTP accepté comme download")
	}
	if dl, _ := ParseDownloadHello(nil); dl {
		t.Fatal("vide accepté comme download")
	}
}
