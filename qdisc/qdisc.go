// Package qdisc — tc configuration behind a TCRunner seam.
// Tests hôte: FakeRunner; la VM exécute ExecRunner (tc réel).
//
// Topology contract (documented ceiling, docs/PLAN.md M1.1): profile
// façonnage et AQM sur deux sauts distincts de la paire veth du banc
// pour que deux racines sans classe ne se disputent jamais une interface:
//
//	veth client ← netem(délai/gigue/perte) · veth façonnage+aqm → veth serveur
//
// Façonnage: cake porte sa propre bande passante; tbf façonne pour pfifo/fq_codel,
// fq_codel attaché comme unique enfant du tbf. Cellule pfifo_fast = tbf avec
// son fifo par défaut (pfifo_fast n'est pas attachable en enfant) — documenté.
package qdisc

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Realms4239/cgo/model"
)

type TCRunner interface {
	Run(args ...string) ([]byte, error)
}

// Chemin absolu épinglé : /etc/sudoers.d/cgo-network n'autorise QUE
// /usr/local/bin/cgo-netctl (jamais tc/ip/ethtool nus — l'ancienne ligne
// joker rendait root via `sudo ip netns exec x bash`). Tout appel Go passe
// par ce wrapper à liste blanche ; sans lui, les cellules de campagne
// héritent d'un « a password is required » muet.
const cgoNetctl = "/usr/local/bin/cgo-netctl"

type ExecRunner struct{}

// Run — `sudo -n cgo-netctl tc …` (jamais tc nu : sudoers ne le couvre pas).
func (ExecRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("sudo", append([]string{"-n", cgoNetctl, "tc"}, args...)...).CombinedOutput()
}

// NsRunner runs tc inside a network namespace via cgo-netctl's ré-entrant
// `ip netns exec --` branch (arguments re-validated inside the namespace).
// Env porte les variables à garder à travers sudo (TC_LIB_DIR pour .dist) :
// « sudo -n /usr/local/bin/cgo-netctl --env TC_LIB_DIR=… ip netns exec … ».
type NsRunner struct {
	Ns  string
	Env []string // paires VAR=value (TC_LIB_DIR uniquement, validé par le wrapper)
}

func (r NsRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("sudo", nsTCArgs(r.Ns, r.Env, args)...).CombinedOutput()
}

// nsTCArgs — argv APRÈS `sudo` : toujours via cgo-netctl. tc nus n'ayant
// plus droit d'accès (sudoers = wrapper seulement), la commande embarquée
// est ré-validée dans le netns.
func nsTCArgs(ns string, env, args []string) []string {
	full := make([]string, 0, len(env)+len(args)+9)
	full = append(full, "-n", cgoNetctl)
	for _, kv := range env {
		full = append(full, "--env", kv)
	}
	full = append(full, "ip", "netns", "exec", ns, "--", "tc")
	return append(full, args...)
}

// FakeRunner records invocations; Run always succeeds.
type FakeRunner struct{ Calls [][]string }

func (f *FakeRunner) Run(args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, args)
	return nil, nil
}

func mbps(v float64) string { return fmt.Sprintf("%gmbit", v) }

// Reset removes any root qdisc on iface.
func Reset(r TCRunner, iface string) error {
	_, err := r.Run("qdisc", "del", "dev", iface, "root")
	if err != nil && !strings.Contains(err.Error(), "Cannot find specified qdisc") &&
		!strings.Contains(err.Error(), "No such file") {
		return fmt.Errorf("reset %s: %w", iface, err)
	}
	return nil
}

// ApplyNetem pose délai/gigue/perte uniforme en qdisc racine du saut de
// latence. It uses handle 1: so that a shaper can be stacked as child 1:1.
// (La perte Gilbert-Elliott vit dans ApplyNetemPhysics/ApplyNetemGE — ge.go.)
func ApplyNetem(r TCRunner, iface string, delayMs, jitterMs, lossPct float64) error {
	return ApplyNetemPhysics(r, iface, delayMs, jitterMs, lossPct, nil)
}

// ApplyShaper configure capacité + AQM sur le saut de façonnage, empilé enfant
// du netem 1: pour que délai et débit touchent la même sortie.
// Si le parent 1: n'existe pas (veth-s sans netem), repli sur la racine.
// Les sorties tc des tentatives sont conservées dans l'erreur : un "exit
// status 2" nu a déjà coûté une demi-journée de diagnostic aveugle.
func ApplyShaper(r TCRunner, iface string, q model.Qdisc, capMbps, rttMs float64) error {
	cakeArgs := func(parent ...string) []string {
		return append(append([]string{"qdisc", "replace", "dev", iface}, parent...),
			"handle", "10:", "cake", "bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs))
	}
	cakeRootArgs := func() []string {
		return []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs)}
	}
	tbfArgs := func(parent ...string) []string {
		return append(append([]string{"qdisc", "replace", "dev", iface}, parent...),
			"handle", "10:", "tbf", "rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")
	}
	tbfRootArgs := func() []string {
		return []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms"}
	}
	switch q {
	case model.Cake:
		if out, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:"}, "handle", "10:", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs))...); err == nil {
			return nil
		} else {
			_ = cakeArgs
			if out2, err2 := r.Run(cakeRootArgs()...); err2 != nil {
				return fmt.Errorf("shaper cake %s: parent: %v (%s) ; root: %v (%s)", iface, err, oneLine(out), err2, oneLine(out2))
			}
			return nil
		}
	case model.FqCodel:
		fq := strings.Fields(fqCoDelString(iface)) // source unique des params RFC 8290
		if out, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:"}, "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")...); err != nil {
			if out2, err2 := r.Run(tbfRootArgs()...); err2 != nil {
				return fmt.Errorf("shaper tbf %s: parent: %v (%s) ; root: %v (%s)", iface, err, oneLine(out), err2, oneLine(out2))
			}
			_ = tbfArgs
			return nil
		}
		_, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "10:1", "handle", "20:"}, fq...)...)
		if err == nil {
			return nil
		}
		// fallback: tbf was at root 1:, so fq_codel parent is 1:1
		_, err = r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:1", "handle", "20:"}, fq...)...)
		return err
	default: // pfifo_fast cell: tbf as child of netem
		if _, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:"}, "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")...); err == nil {
			return nil
		} else {
			_ = tbfArgs
		}
		_, err := r.Run(tbfRootArgs()...)
		return err
	}
}

func fqCoDelString(iface string) string {
	return "fq_codel target 5ms interval 100ms limit 10240 quantum 1514 flows 1024"
}

func oneLine(b []byte) string {
	s := strings.TrimSpace(string(b))
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 160 {
		s = s[:160]
	}
	return s
}
