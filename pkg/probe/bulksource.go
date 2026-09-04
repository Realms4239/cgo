package probe

import (
	"context"
	"net"
	"strings"
	"time"
)

// BulkSource — le SENS DESCENDANT : la source (serveur) inonde la connexion
// jusqu'au terme du ctx, le client reçoit. Miroir exact de BulkSend (même
// échéance d'écriture 2 s, même rendu d'octets) : le banc peut enfin
// façonner et mesurer le download — shaper sur l'émission du serveur.
func BulkSource(ctx context.Context, conn net.Conn) (uint64, error) {
	defer conn.Close()
	buf := make([]byte, 64*1024) // zeros
	var total uint64
	for {
		select {
		case <-ctx.Done():
			return total, nil
		default:
		}
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Write(buf)
		total += uint64(n)
		if err != nil {
			return total, nil // récepteur fermé = fin de charge
		}
	}
}

// BulkDownloadTo — protocole du banc : dial, "D" ou "D:<cc>\n" (mode source
// sur le sink, la CC voyage avec la poignée de main), puis le client
// RECEVOIT jusqu'au terme du ctx. Rend les octets reçus — la mesure
// download côté client, comme BulkSendTo est la mesure upload côté client.
func BulkDownloadTo(ctx context.Context, addr, cc string) (uint64, error) {
	conn, err := DialWithCC(ctx, addr, cc)
	if err != nil {
		return 0, err
	}
	hello := "D"
	if cc != "" {
		hello = "D:" + cc + "\n"
	}
	if _, err := conn.Write([]byte(hello)); err != nil {
		conn.Close()
		return 0, err
	}
	return BulkReceive(ctx, conn, nil)
}

// ParseDownloadHello — "D" (client historique, CC hôte) ou "D:<cc>" (cellule
// étiquetée). Tolérant : espaces, \n, casse du préfixe. "" si inconnu.
func ParseDownloadHello(b []byte) (isDownload bool, cc string) {
	s := strings.TrimSpace(string(b))
	if len(s) == 0 || (s[0] != 'D' && s[0] != 'd') {
		return false, ""
	}
	if len(s) > 2 && s[1] == ':' {
		return true, strings.TrimSpace(s[2:])
	}
	return true, ""
}
