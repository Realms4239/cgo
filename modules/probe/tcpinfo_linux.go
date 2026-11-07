//go:build linux

package probe

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// MSS de référence du banc (veth, MTU 1500) : TCP_INFO rend des SEGMENTS
// réémis ; ×1448 les convertit en octets comme le fait le mémoire
// (wasted = drops × 1448). Proxy documenté, pas une mesure par paquet.
const retransMSS = 1448

// TCPRetransBytes — octets réémis cumulés d'une connexion TCP (TCP_INFO).
// Linux uniquement : ailleurs, l'appel rend une erreur et l'appelant déclare
// le ratio non mesuré — jamais estimé en silence.
func TCPRetransBytes(c net.Conn) (uint64, error) {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("tcpinfo: pas une connexion TCP")
	}
	f, err := tc.File()
	if err != nil {
		return 0, fmt.Errorf("tcpinfo: file: %w", err)
	}
	defer f.Close()
	// Le stdlib syscall n'expose pas de GetsockoptTCPInfo (c'est x/sys/unix,
	// interdit ici) : getsockopt brut, exactement ce que ferait la bibliothèque
	// — le noyau remplit min(taille struct, taille noyau) ; sur un noyau plus
	// vieux que la struct Go, les champs de queue (Total_retrans est le
	// dernier du noyau classique) restent à zéro, jamais faux.
	var info syscall.TCPInfo
	vlen := uint32(unsafe.Sizeof(info))
	_, _, errno := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		f.Fd(), syscall.IPPROTO_TCP, uintptr(syscall.TCP_INFO),
		uintptr(unsafe.Pointer(&info)), uintptr(unsafe.Pointer(&vlen)), 0)
	if errno != 0 {
		return 0, fmt.Errorf("tcpinfo: getsockopt: %w", errno)
	}
	return uint64(info.Total_retrans) * retransMSS, nil
}
