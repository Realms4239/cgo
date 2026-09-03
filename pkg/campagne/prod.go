package campagne

import (
	"os"

	"github.com/Realms4239/cgo/pkg/qdisc"
)

// ProdDeps retourne les dépendances de production du runner campagne,
// câblées à la topologie du banc de kit/testbed.sh :
// saut netem veth-c (ns principal) · saut façonnage+aqm veth-s (netns cgo-srv).
// Chaque adresse est surchargeable par env pour les setups exotiques.
//
// Placement du shaper : le bulk est un UPLOAD (BulkSendTo client→sink), donc le
// plafond de débit doit être sur l'émission client — veth-c, empilé en enfant de
// la racine netem à 1: (repli parent 1: d'ApplyShaper). Le façonnage sur l'émission
// veth-s ne couvre que le téléchargement et laisse l'upload non façonné
// (G4 échec → lignes invalid).
func ProdDeps() Deps {
	target := env("CGO_TARGET", "10.200.0.1")
	small := env("CGO_SMALL_URL", "http://10.200.0.1:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.200.0.1:5201")
	cliIf := env("CGO_CLI_IF", "veth-c")
	shaperIf := env("CGO_SHAPER_IF", cliIf) // même saut que netem — émission upload
	return Deps{
		TC:       qdisc.ExecRunner{}, // netem sur veth-c (ns principal)
		TCShaper: qdisc.ExecRunner{}, // shaper aussi empilé sur veth-c (émission upload)
		CliIf:    cliIf,
		ShaperIf: shaperIf,
		Target:   target,
		SmallURL: small,
		BulkAddr: bulk,
		StatsFn: func() []qdisc.Stats {
			// sonder UNIQUEMENT le saut client façonné (veth-c) : sommer les deux sauts
			// double-compte chaque octet (les mêmes paquets traversent veth-c et
			// veth-s), gonflant le goodput ~2× et cassant la cohérence G4.
			sts1, err1 := qdisc.PollStats(qdisc.ExecRunner{}, cliIf)
			if err1 != nil {
				return nil
			}
			return sts1
		},
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
