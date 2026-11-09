package probe

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestBulkStatsMultiFlowLocal(t *testing.T) {
	// Puits local : 3 flux vers un listener discard — le total par flux doit
	// remplir Bytes, et RetransMeasured suit la plateforme (false sur Windows,
	// indifférent ici : on teste la structure, pas le noyau).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("loopback indisponible")
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = BulkReceive(ctx, c, nil)
			}(c)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	st, err := BulkSendNStats(ctx, ln.Addr().String(), "cubic", 3)
	if err != nil {
		t.Fatalf("BulkSendNStats: %v", err)
	}
	if len(st.Bytes) != 3 {
		t.Fatalf("3 flux attendus, %d rendus", len(st.Bytes))
	}
	for _, b := range st.Bytes {
		if b == 0 {
			t.Fatal("un flux n'a rien transporté — la charge doit être réelle")
		}
	}
}

func TestTCPRetransBytesErrorOnNonTCP(t *testing.T) {
	// Un non-TCPConn doit rendre une erreur propre, pas un panic.
	if _, err := TCPRetransBytes(udpConn(t)); err == nil {
		t.Fatal("non-TCP conn : erreur attendue")
	}
}

func udpConn(t *testing.T) net.Conn {
	t.Helper()
	c, err := net.Dial("udp", "127.0.0.1:9")
	if err != nil {
		t.Skip("udp indisponible")
	}
	t.Cleanup(func() { c.Close() })
	return c
}
