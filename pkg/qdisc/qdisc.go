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
	args := []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs)}
	if lossPct > 0 {
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
func ApplyShaper(r TCRunner, iface string, q model.Qdisc, capMbps, rttMs float64) error {
	switch q {
	case model.Cake:
		if _, err := r.Run("qdisc", "replace", "dev", iface, "parent", "1:", "handle", "10:", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs)); err == nil {
			return nil
		}
		_, err := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs))
		return err
	case model.FqCodel:
		if _, err := r.Run("qdisc", "replace", "dev", iface, "parent", "1:", "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms"); err != nil {
			if _, err2 := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "tbf",
				"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms"); err2 != nil {
				return err
			}
		}
		_, err := r.Run("qdisc", "replace", "dev", iface, "parent", "10:1", "handle", "20:", "fq_codel")
		if err == nil {
			return nil
		}
		// fallback: tbf was at root 1:, so fq_codel parent is 1:1
		_, err = r.Run("qdisc", "replace", "dev", iface, "parent", "1:1", "handle", "20:", "fq_codel")
		return err
	default: // pfifo_fast cell: tbf as child of netem
		if _, err := r.Run("qdisc", "replace", "dev", iface, "parent", "1:", "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms"); err == nil {
			return nil
		}
		_, err := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")
		return err
	}
}
