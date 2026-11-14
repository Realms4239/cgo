//go:build !linux

package probe

import (
	"syscall"
)

// dscpMark — hors Linux (banc) : pas de marquage DSCP disponible, la sonde
// small part best-effort. Documenté : la mesure EF n'est interprétable que
// sur le banc Linux.
var dscpControl func(_, _ string, c syscall.RawConn) error // nil : pas de contrôle
