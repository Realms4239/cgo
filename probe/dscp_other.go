//go:build !linux

package probe

import (
	"context"
	"net"
	"time"
)

// dscpMark — hors Linux (banc) : pas de marquage DSCP disponible, la sonde
// small part best-effort. Documenté : la mesure EF n'est interprétable que
// sur le banc Linux.
func dialSmallMarked(ctx context.Context, addr string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 5 * time.Second}
	return d.DialContext(ctx, "tcp", addr)
}
