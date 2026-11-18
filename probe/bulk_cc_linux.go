//go:build linux

package probe

import (
	"context"
	"net"
	"syscall"
)

const tcpCongestion = 0x0d

func DialWithCC(ctx context.Context, addr, cc string) (net.Conn, error) {
	var control func(network, address string, c syscall.RawConn) error
	if cc != "" {
		control = func(_, _ string, c syscall.RawConn) error {
			var serr error
			if err := c.Control(func(fd uintptr) {
				serr = syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, tcpCongestion, socketToken(cc))
			}); err != nil {
				return err
			}
			return serr
		}
	}
	return guardedDialContext(ctx, "tcp", addr, AllowPrivateTargets(), control)
}
