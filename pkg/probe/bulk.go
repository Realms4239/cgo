package probe

import (
	"context"
	"io"
	"net"
	"time"
)

// BulkSend floods conn with zeros until ctx is done; returns bytes sent.
func BulkSend(ctx context.Context, conn net.Conn) (uint64, error) {
	buf := make([]byte, 64*1024) // zeros
	var total uint64
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return total, nil
		default:
		}
		n, err := conn.Write(buf)
		total += uint64(n)
		if err != nil {
			return total, nil // receiver closed = end of charge
		}
	}
}

// BulkReceive reads until EOF; reports byte count via callback every read.
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
