package campagne

import (
	"context"
	"fmt"

	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/probe"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// BurstGE — le canal en salves (Gilbert-Elliott) par profil : P2/P3
// uniquement (4G tropical / VSAT affaibli par la pluie). Les autres profils
// n'ont pas de mode rafales : le lien fibre n'est pas un canal radio, et le
// LEO P4 porte sa gigue DÉCLARÉE (G6 adaptatif) sans salves de perte.
// Cibles vérifiées par qdisc.GEMeanLoss : P2 ≈ 0,544 %, P3 ≈ 1,08 %.
func BurstGE(profileID string) (qdisc.GEModel, bool) {
	switch profileID {
	case "P2":
		return qdisc.GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}, true
	case "P3":
		return qdisc.GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001}, true
	}
	return qdisc.GEModel{}, false
}

// burstGEOrNull — le modèle en salves si la campagne l'exige ET que le
// profil en offre un ; nil sinon (perte uniforme). L'appelant a déjà
// validé le couple bursty×profil AVANT toute mutation.
func burstGEOrNull(profileID, channelModel string) *qdisc.GEModel {
	if channelModel != "bursty" {
		return nil
	}
	if ge, ok := BurstGE(profileID); ok {
		return &ge
	}
	return nil
}

// applyPhysics — pose la physique du profil (délai/gigue/perte) sur le saut
// de latence ET sur chaque iface supplémentaire : dans le banc durci les
// DEUX sens traversent ns_wan (w-edge montant, c-wan descendant) — poser un
// seul sens ferait d'un run « symétrique » un lien asymétrique en secret.
// Reset d'abord : une cellule périmée peut avoir laissé une racine étrangère.
func applyPhysics(d *Deps, prof model.Profile, ge *qdisc.GEModel) error {
	ifaces := append([]string{d.CliIf}, d.PhysicsIfaces...)
	for _, ifc := range ifaces {
		_, _ = d.TC.Run("qdisc", "del", "dev", ifc, "root")
		if err := qdisc.ApplyNetemPhysics(d.TC, ifc, prof.DelayMs, prof.JitterMs, prof.LossPct, ge); err != nil {
			return fmt.Errorf("netem %s: %w", ifc, err)
		}
	}
	return nil
}

// ProdDepsDualNS — les dépendances de production sur le BANC DURCI
// (kit/run-dualns.sh) : physique dans ns_wan_emulator (w-edge + c-wan),
// façonnage dans ns_edge_router (e-srv montant / e-wan descendant), puits
// dans ns_server (10.30.0.2 — les QUATRE sauts sur le chemin mesuré).
// CGO_* surcharge chaque adresse pour les setups exotiques, comme prodDeps.
func ProdDepsDualNS(direction string) Deps {
	target := env("CGO_TARGET", "10.30.0.2")
	small := env("CGO_SMALL_URL", "http://10.30.0.2:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.30.0.2:5201")
	shaperIf := env("CGO_SHAPER_IF", "e-srv") // émission edge → serveur : le upload
	if direction == "down" {
		shaperIf = env("CGO_SHAPER_IF", "e-wan") // émission edge → client : le download
	}
	edge := qdisc.NsRunner{Ns: env("CGO_EDGE_NS", "ns_edge_router")}
	return Deps{
		TC:            qdisc.NsRunner{Ns: env("CGO_WAN_NS", "ns_wan_emulator")}, // netem : les deux sens
		TCShaper:      edge,                                                     // shaper : là où naît le congestionnement
		CliIf:         env("CGO_CLI_IF", "w-edge"),
		PhysicsIfaces: []string{env("CGO_DOWNLINK_PHYS_IF", "c-wan")},
		ShaperIf:      shaperIf,
		Target:        target,
		SmallURL:      small,
		BulkAddr:      bulk,
		Direction:     direction,
		Bench:         "dualns",
		ChannelModel:  env("CGO_CHANNEL_MODEL", "clean"),
		BaselineSec:   envInt("CGO_BASELINE_SEC", 0),
		ChargeSec:     envInt("CGO_CHARGE_SEC", 0),
		RecupSec:      envInt("CGO_RECUP_SEC", 0),
		BulkN: func(ctx context.Context, addr, cc string, n int) ([]uint64, error) {
			return probe.BulkSendNTo(ctx, addr, cc, n)
		},
		StatsFn: func() []qdisc.Stats {
			// UN SEUL saut, celui du sens mesuré (l'émission façonnée) —
			// sommer les deux directions double-compterait chaque octet.
			sts, err := qdisc.PollStats(edge, shaperIf)
			if err != nil {
				return nil
			}
			return sts
		},
	}
}
