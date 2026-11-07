//go:build linux

package probe

import (
	"context"
	"net"
	"syscall"
	"time"
)

const tcpCongestion = 0x0d

func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 5 * time.Second}
	if cc != "" {
		d.Control = func(_, _ string, c syscall.RawConn) error {
			var serr error
			if err := c.Control(func(fd uintptr) {
				serr = syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, tcpCongestion, cc)
			}); err != nil {
				return err
			}
			return serr
		}
	}
	return d.DialContext(ctx, "tcp", addr)
}
