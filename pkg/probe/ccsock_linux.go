//go:build linux

package probe

import (
	"net"
	"syscall"
)

// SetConnCC impose le contrôle de congestion d'une connexion ACCEPTÉE
// (côté serveur) : la cellule download porte sa CC comme la cellule upload
// (DialWithCC côté client) — sinon l'émetteur download tourne en CC
// par-défaut de l'hôte (cubic) quelle que soit la CC de la cellule, et la
// comparaison inter-CC du download est un mensonge d'étiquette.
func SetConnCC(conn net.Conn, cc string) error {
	if cc == "" {
		return nil
	}
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return nil
	}
	raw, err := tcp.SyscallConn()
	if err != nil {
		return err
	}
	var serr error
	if err := raw.Control(func(fd uintptr) {
		serr = syscall.SetsockoptString(int(fd), syscall.IPPROTO_TCP, tcpCongestion, cc)
	}); err != nil {
		return err
	}
	return serr
}
