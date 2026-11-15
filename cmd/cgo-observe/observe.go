// cgo-observe — périmètre : la logique pure, compilable partout.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// maxDurationSec — plafond dur du trace : 48 h. Aucun appelant ne peut
// prolonger le marquage.
const maxDurationSec = 172800

// clampDuration borne la demande à maxDurationSec ; <=0 → défaut 60 s.
func clampDuration(sec int) int {
	if sec <= 0 {
		return 60
	}
	if sec > maxDurationSec {
		return maxDurationSec
	}
	return sec
}

// pseudonym — HMAC-SHA256(ip, salt) tronqué à 8 octets hex : le flux reste
// traçable DANS une fenêtre, jamais réidentifiable hors d'elle. Sel tiré au
// lancement, jamais persisté.
func pseudonym(ip string, salt []byte) string {
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(ip))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum[:8])
}
