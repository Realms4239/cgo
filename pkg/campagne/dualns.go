package campagne

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/results"
	"github.com/Realms4239/cgo/probe"
	"github.com/Realms4239/cgo/qdisc"
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
	switch d.DelayModel {
	case "", "scalar":
	case "dist":
		return applyPhysicsDist(d, prof, ge, ifaces)
	default:
		return fmt.Errorf("delay_model %q invalide (scalar|dist)", d.DelayModel)
	}
	for _, ifc := range ifaces {
		_, _ = d.TC.Run("qdisc", "del", "dev", ifc, "root")
		if err := qdisc.ApplyNetemPhysics(d.TC, ifc, prof.DelayMs, prof.JitterMs, prof.LossPct, ge); err != nil {
			return fmt.Errorf("netem %s: %w", ifc, err)
		}
	}
	return nil
}

// applyPhysicsDist — physique d'une trace : plancher RTT/2 sur chaque saut,
// table empirique sur le seul saut mesuré. Toute validation (calibration,
// profil, sens, banc) passe AVANT le premier reset/tc : un artefact absent
// ou périmé ne laisse jamais une racine à moitié configurée.
func applyPhysicsDist(d *Deps, prof model.Profile, ge *qdisc.GEModel, ifaces []string) error {
	if d.Direction == "both" {
		return fmt.Errorf("delay_model dist: direction both impossible — la trace mesure un seul sens")
	}
	if d.Bench != "dualns" {
		return fmt.Errorf("delay_model dist: banc dualns requis (banc %q)", d.Bench)
	}
	if d.CalibrationDir == "" {
		return fmt.Errorf("delay_model dist: répertoire de calibration vide (CGO_CALIBRATION_DIR)")
	}
	cal, err := qdisc.LoadCalibration(d.CalibrationDir, prof.ID)
	if err != nil {
		return err
	}
	if cal.FloorMs != prof.DelayMs {
		return fmt.Errorf("calibration %s périmée : plancher %.2f ms ≠ profil %.2f ms (relancer cgo calibrate)", prof.ID, cal.FloorMs, prof.DelayMs)
	}
	if len(d.PhysicsIfaces) != 1 {
		return fmt.Errorf("delay_model dist: une seule PhysicsIfaces attendue, %d fournie(s)", len(d.PhysicsIfaces))
	}
	measured := d.CliIf
	if d.Direction == "down" {
		measured = d.PhysicsIfaces[0]
	}
	for _, ifc := range ifaces {
		_, _ = d.TC.Run("qdisc", "del", "dev", ifc, "root")
		if ifc == measured {
			if err := qdisc.ApplyNetemDistribution(d.TC, ifc, cal.FloorMs/2, cal.ScaleMs, prof.LossPct, ge, prof.ID); err != nil {
				return fmt.Errorf("netem dist %s: %w", ifc, err)
			}
			continue
		}
		if err := qdisc.ApplyNetemPhysics(d.TC, ifc, cal.FloorMs/2, 0, prof.LossPct, ge); err != nil {
			return fmt.Errorf("netem dist %s: %w", ifc, err)
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
	delayModel := env("CGO_DELAY_MODEL", "scalar")
	calibDir := env("CGO_CALIBRATION_DIR", filepath.Join(results.DataRoot(), "calibrations"))
	wan := qdisc.NsRunner{Ns: env("CGO_WAN_NS", "ns_wan_emulator")}
	if delayModel == "dist" {
		if abs, err := filepath.Abs(calibDir); err == nil {
			calibDir = abs
		}
		wan.Env = []string{"TC_LIB_DIR=" + calibDir}
	}
	shapeMode := env("CGO_SHAPE_MODE", "egress")
	deps := Deps{
		TC:             wan, // netem : les deux sens
		TCShaper:       edge,
		CliIf:          env("CGO_CLI_IF", "w-edge"),
		PhysicsIfaces:  []string{env("CGO_DOWNLINK_PHYS_IF", "c-wan")},
		ShaperIf:       shaperIf,
		Target:         target,
		SmallURL:       small,
		BulkAddr:       bulk,
		Direction:      direction,
		Bench:          "dualns",
		ChannelModel:   env("CGO_CHANNEL_MODEL", "clean"),
		DelayModel:     delayModel,
		CalibrationDir: calibDir,
		ShapeMode:      shapeMode,
		BaselineSec:    envInt("CGO_BASELINE_SEC", 0),
		ChargeSec:      envInt("CGO_CHARGE_SEC", 0),
		RecupSec:       envInt("CGO_RECUP_SEC", 0),
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
	if shapeMode == "ingress" {
		clientNS := env("CGO_CLIENT_NS", "ns_client")
		ingressIf := env("CGO_INGRESS_IF", "c-cli")
		ingressDev := env("CGO_INGRESS_DEV", "ifb0")
		client := qdisc.NsRunner{Ns: clientNS}
		deps.TCShaper = client
		deps.ShaperIf = ingressDev
		deps.IngressPreflight = func() error {
			return qdisc.CheckIngress(qdisc.NsCommand{Ns: clientNS}, ingressIf, ingressDev)
		}
		deps.StatsFn = func() []qdisc.Stats {
			sts, err := qdisc.PollStats(client, ingressDev)
			if err != nil {
				return nil
			}
			return sts
		}
	}
	return deps
}
