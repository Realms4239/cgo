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
// Env porte les variables à garder à travers sudo (TC_LIB_DIR pour .dist) :
// « sudo env VAR=… ip netns exec … tc » ne dépend pas d'env_keep.
type NsRunner struct {
	Ns  string
	Env []string // paires VAR=value
}

func (r NsRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("sudo", nsTCArgs(r.Ns, r.Env, args)...).CombinedOutput()
}

func nsTCArgs(ns string, env, args []string) []string {
	full := make([]string, 0, len(env)+len(args)+6)
	if len(env) > 0 {
		full = append(full, "env")
		full = append(full, env...)
	}
	full = append(full, "ip", "netns", "exec", ns, "tc")
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
		if out, err := r.Run(append([]string{"qdisc", "replace", "dev", iface, "parent", "1:"}, "handle", "10:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")...); err != nil {
			if out2, err2 := r.Run(tbfRootArgs()...); err2 != nil {
				return fmt.Errorf("shaper tbf %s: parent: %v (%s) ; root: %v (%s)", iface, err, oneLine(out), err2, oneLine(out2))
			}
			_ = tbfArgs
			return nil
		}
		_, err := r.Run("qdisc", "replace", "dev", iface, "parent", "10:1", "handle", "20:", "fq_codel", "target", "5ms", "interval", "100ms", "limit", "10240", "quantum", "1514", "flows", "1024")
		if err == nil {
			return nil
		}
		// fallback: tbf was at root 1:, so fq_codel parent is 1:1
		_, err = r.Run("qdisc", "replace", "dev", iface, "parent", "1:1", "handle", "20:", "fq_codel", "target", "5ms", "interval", "100ms", "limit", "10240", "quantum", "1514", "flows", "1024")
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
