package campagne

import (
	"context"
	"os"

	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/probe"
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
// ProdDeps — charge MONTANTE (upload) : shaper sur l'émission client (veth-c).
func ProdDeps() Deps {
	return prodDeps("")
}

// ProdDepsDown — charge DESCENDANTE (download) : shaper sur l'émission
// serveur, veth-s dans le netns cgo-srv (NsRunner). Le netem reste sur
// veth-c (le délai s'applique aux deux sens au saut client).
func ProdDepsDown() Deps {
	return prodDeps("down")
}

func prodDeps(direction string) Deps {
	target := env("CGO_TARGET", "10.200.0.1")
	small := env("CGO_SMALL_URL", "http://10.200.0.1:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.200.0.1:5201")
	cliIf := env("CGO_CLI_IF", "veth-c")
	shaperIf := env("CGO_SHAPER_IF", cliIf) // même saut que netem — émission upload
	var shaper qdisc.TCRunner = qdisc.ExecRunner{}
	statsIf := cliIf
	statsRunner := qdisc.TCRunner(qdisc.ExecRunner{})
	if direction == "down" {
		srvIf := env("CGO_SRV_IF", "veth-s")
		shaper = qdisc.NsRunner{Ns: env("CGO_SRV_NS", "cgo-srv")}
		shaperIf = srvIf
		// compteurs du SENS MESURÉ : en download, les octets et pertes
		// vivent sur l'émission serveur (veth-s) — sonder veth-c ne
		// compterait que les ACK (le gel affichait 0,0 malgré des Mo
		// reçus : prouvé 2026-09-04, serveur 24 Mo → gel 0,0).
		statsIf = srvIf
		statsRunner = qdisc.NsRunner{Ns: env("CGO_SRV_NS", "cgo-srv")}
	}
	return Deps{
		TC:        qdisc.ExecRunner{}, // netem sur veth-c (ns principal)
		TCShaper:  shaper,             // shaper : veth-c (up) ou veth-s via netns (down)
		CliIf:     cliIf,
		ShaperIf:  shaperIf,
		Target:    target,
		SmallURL:  small,
		BulkAddr:  bulk,
		Direction: direction,
		// fenêtre multi-flux : N connexions vers le sink, per-flow (JFI)
		BulkN: func(ctx context.Context, addr string, n int) ([]uint64, error) {
			return probe.BulkSendNTo(ctx, addr, string(model.Cubic), n)
		},
		StatsFn: func() []qdisc.Stats {
			// UN SEUL saut, celui du sens mesuré : sommer les deux sauts
			// double-compte chaque octet (les mêmes paquets traversent veth-c
			// et veth-s), gonflant le goodput ~2× et cassant la cohérence G4.
			// up = émission cliente (veth-c), down = émission serveur (veth-s).
			sts1, err1 := qdisc.PollStats(statsRunner, statsIf)
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
