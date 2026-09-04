//go:build !linux

package probe

import "net"

// SetConnCC — hors Linux : pas de TCP_CONGESTION disponible, l'émetteur
// download tourne en CC par-défaut (documenté : les cellules download
// inter-CC ne sont interprétables que sur le banc Linux).
func SetConnCC(_ net.Conn, _ string) error { return nil }
