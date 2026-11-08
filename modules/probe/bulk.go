package probe

import (
	"context"
	"io"
	"net"
	"time"
)

// BulkSendTo connecte addr avec le cc et inonde de zéros jusqu'à la fin du ctx.
func BulkSendTo(ctx context.Context, addr, cc string) (uint64, error) {
	conn, err := DialWithCC(ctx, addr, cc)
	if err != nil {
		return 0, err
	}
	return BulkSend(ctx, conn)
}

// BulkSend inonde conn de zéros jusqu'à la fin du ctx; rend les octets envoyés.
func BulkSend(ctx context.Context, conn net.Conn) (uint64, error) {
	defer conn.Close()
	buf := make([]byte, 64*1024) // zeros
	var total uint64
	for {
		select {
		case <-ctx.Done():
			return total, nil
		default:
		}
		// échéance d'écriture : une connexion figée (récepteur bloqué, lien
		// mort) ne doit pas pendre la cellule — le puits du banc lit en
		// continu, 2 s par write de 64 Ko ≈ 20× la marge VSAT 5 Mb/s.
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Write(buf)
		total += uint64(n)
		if err != nil {
			return total, nil // receiver closed = end of charge
		}
	}
}

// BulkReceive lit jusqu'à EOF; signale le compteur d'octets à chaque lecture.
func BulkReceive(ctx context.Context, conn net.Conn, onBytes func(delta uint64)) (uint64, error) {
	defer conn.Close()
	buf := make([]byte, 256*1024)
	var total uint64
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.SetReadDeadline(pastTime())
			close(done)
		case <-done:
		}
	}()
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			total += uint64(n)
			if onBytes != nil {
				onBytes(uint64(n))
			}
		}
		if err == io.EOF || err == nil && n == 0 {
			break
		}
		if err != nil {
			break // deadline from cancel = charge window over
		}
		if ctx.Err() != nil {
			break
		}
	}
	return total, nil
}

func pastTime() time.Time { return time.Unix(1, 0) }
