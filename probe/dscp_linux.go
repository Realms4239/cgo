//go:build linux

package probe

import (
	"syscall"
)

// dscpMark — socket du petit objet marqué DSCP EF (46) : la sonde vit la
// classe temps réel que l'AQM doit protéger ; sans marquage, la comparaison
// pfifo-vs-cake sous-teste exactement la différenciation qui justifie la
// recommandation. Linux seul (banc) ; autres OS : pas de marquage, sonde BE.
const dscpEF = 46 // Expedited Forwarding, RFC 3246

// dscpControl — marquage DSCP EF au moment du connect (via net.Dialer.Control).
var dscpControl = func(_, _ string, c syscall.RawConn) error {
	var serr error
	if err := c.Control(func(fd uintptr) {
		// DSCP dans les 6 bits hauts du TOS : 46 << 2 = 184
		serr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TOS, dscpEF<<2)
	}); err != nil {
		return err
	}
	return serr
}
