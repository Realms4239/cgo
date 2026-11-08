//go:build !linux

package probe

import (
	"context"
	"net"
	"time"
)

func DialWithCC(ctx context.Context, addr, _ string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 5 * time.Second}
	return d.DialContext(ctx, "tcp", addr)
}
