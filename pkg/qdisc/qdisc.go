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

	"github.com/Realms4239/cgo/pkg/model"
)

type TCRunner interface {
	Run(args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(args ...string) ([]byte, error) {
	// la VM exige sudo pour tc (NOPASSWD dans /etc/sudoers.d/cgo-network)
	return exec.Command("sudo", append([]string{"tc"}, args...)...).CombinedOutput()
}

// NsRunner runs tc inside a network namespace (ip netns exec NS tc …).
type NsRunner struct{ Ns string }

func (r NsRunner) Run(args ...string) ([]byte, error) {
	full := append([]string{"netns", "exec", r.Ns, "tc"}, args...)
	return exec.Command("sudo", append([]string{"ip"}, full...)...).CombinedOutput()
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

// ApplyNetem pose délai/gigue/perte en qdisc racine du saut de latence.
// It uses handle 1: so that a shaper can be stacked as child 1:1.
func ApplyNetem(r TCRunner, iface string, delayMs, jitterMs, lossPct float64) error {
	return ApplyNetemBurst(r, iface, delayMs, jitterMs, lossPct, 0, 0, 0, 0)
}

// ApplyNetemBurst — perte uniforme (défaut historique) ou Gilbert-Elliott
// quand la proba de perte gemodel est fournie : netem
// "loss gemodel p r h 1-r k", la vraie vie des liens mobiles/satellite où
// la perte arrive en rafales, pas en gouttes uniformes. Chaîne à 2 états :
// p = proba de perte, r = bon→mauvais, h = persistance mauvais,
// k = mauvais→bon (1-k = mauvais→bon ; netem paramètre 1-r/k à sa guise
// selon la version — on passe p r h 1-r k, ordre du man netem).
// Kernel banc 6.8 — support vérifié.
func ApplyNetemBurst(r TCRunner, iface string, delayMs, jitterMs, lossPct, gemodelP, gemodelR, gemodelH, gemodelK float64) error {
	args := []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs)}
	if gemodelP > 0 {
		args = append(args, "loss", "gemodel",
			fmt.Sprintf("%g", gemodelP),   // p : proba de perte
			fmt.Sprintf("%g", gemodelR),   // r : bon → mauvais
			fmt.Sprintf("%g", gemodelH),   // h : persistance mauvais
			fmt.Sprintf("%g", 1-gemodelR), // 1-r
			fmt.Sprintf("%g", gemodelK))   // k : mauvais → bon
	} else if lossPct > 0 {
		args = append(args, "loss", fmt.Sprintf("%g%%", lossPct))
	}
	if _, err := r.Run(args...); err != nil {
		return fmt.Errorf("netem %s: %w", iface, err)
	}
	return nil
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
		if out, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:"}, "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")...); err != nil {
			if out2, err2 := r.Run(tbfRootArgs()...); err2 != nil {
				return fmt.Errorf("shaper tbf %s: parent: %v (%s) ; root: %v (%s)", iface, err, oneLine(out), err2, oneLine(out2))
			}
			_ = tbfArgs
			return nil
		}
		_, err := r.Run("qdisc", "replace", "dev", iface, "parent", "10:1", "handle", "20:", "fq_codel")
		if err == nil {
			return nil
		}
		// fallback: tbf was at root 1:, so fq_codel parent is 1:1
		_, err = r.Run("qdisc", "replace", "dev", iface, "parent", "1:1", "handle", "20:", "fq_codel")
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
