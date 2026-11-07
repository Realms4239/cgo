//go:build !linux

package probe

import (
	"errors"
	"net"
)

// ErrTCPInfoUnavailable — TCP_INFO n'existe que sur Linux ; ailleurs le ratio
// de réémission est NON MESURÉ (l'audit le déclare, il ne l'invente pas).
var ErrTCPInfoUnavailable = errors.New("tcpinfo: TCP_INFO indisponible hors Linux")

func TCPRetransBytes(net.Conn) (uint64, error) { return 0, ErrTCPInfoUnavailable }
