// Package qdisc — tc configuration behind a TCRunner seam.
// Host tests use FakeRunner; the VM runs ExecRunner (real tc).
//
// Topology contract (documented ceiling, docs/PLAN.md M1.1): profile
// shaping and AQM live on separate layer-2 hops of the testbed veth pair
// so two classless roots never compete on one interface:
//
//	client veth ← netem(delay/jitter/loss) · shaping+aqm veth → server veth
//
// Shaping: cake carries its own bandwidth; tbf shapes for pfifo/fq_codel,
// with fq_codel attached as tbf's single child. pfifo_fast cell = tbf with
// its default fifo (pfifo_fast is not attachable as a child) — documented.
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
	return exec.Command("tc", args...).CombinedOutput()
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

// ApplyNetem puts delay/jitter/loss as the root qdisc of the latency hop.
func ApplyNetem(r TCRunner, iface string, delayMs, jitterMs, lossPct float64) error {
	args := []string{"qdisc", "replace", "dev", iface, "root", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs)}
	if lossPct > 0 {
		args = append(args, "loss", fmt.Sprintf("%g%%", lossPct))
	}
	if _, err := r.Run(args...); err != nil {
		return fmt.Errorf("netem %s: %w", iface, err)
	}
	return nil
}

// ApplyShaper configures capacity + AQM on the shaping hop.
func ApplyShaper(r TCRunner, iface string, q model.Qdisc, capMbps, rttMs float64) error {
	switch q {
	case model.Cake:
		_, err := r.Run("qdisc", "replace", "dev", iface, "root", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs))
		return err
	case model.FqCodel:
		if _, err := r.Run("qdisc", "replace", "dev", iface, "root", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms"); err != nil {
			return err
		}
		_, err := r.Run("qdisc", "add", "dev", iface, "parent", "1:1", "handle", "2:", "fq_codel")
		return err
	default: // pfifo_fast cell: tbf + default fifo
		_, err := r.Run("qdisc", "replace", "dev", iface, "root", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")
		return err
	}
}
