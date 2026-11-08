package qdisc

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Realms4239/cgo/model"
)

// IngressRunner runs ip and tc commands, optionally inside a netns. The
// program is explicit because the preflight needs both ip and tc.
type IngressRunner interface {
	Run(prog string, args ...string) ([]byte, error)
}

// NsCommand runs a command inside a network namespace.
type NsCommand struct{ Ns string }

func (c NsCommand) Run(prog string, args ...string) ([]byte, error) {
	full := append([]string{"netns", "exec", c.Ns, prog}, args...)
	return exec.Command("sudo", append([]string{"ip"}, full...)...).CombinedOutput()
}

// ApplyIngressShaper shapes an IFB at its root. The WAN ingress has already
// been redirected into this device by the bench; an ingress qdisc itself can
// filter and redirect but cannot build the queue an AQM needs.
func ApplyIngressShaper(r TCRunner, iface string, q model.Qdisc, capMbps, rttMs float64) error {
	switch q {
	case model.Cake:
		out, err := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "cake",
			"bandwidth", mbps(capMbps), "rtt", fmt.Sprintf("%gms", rttMs), "ingress")
		if err != nil {
			return fmt.Errorf("shaper ingress cake %s: %v (%s)", iface, err, oneLine(out))
		}
		return nil
	case model.FqCodel:
		out, err := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")
		if err != nil {
			return fmt.Errorf("shaper ingress tbf %s: %v (%s)", iface, err, oneLine(out))
		}
		out, err = r.Run("qdisc", "replace", "dev", iface, "parent", "1:1", "handle", "10:", "fq_codel")
		if err != nil {
			return fmt.Errorf("shaper ingress fq_codel %s: %v (%s)", iface, err, oneLine(out))
		}
		return nil
	default:
		out, err := r.Run("qdisc", "replace", "dev", iface, "root", "handle", "1:", "tbf",
			"rate", mbps(capMbps), "burst", "256kbit", "latency", "400ms")
		if err != nil {
			return fmt.Errorf("shaper ingress tbf %s: %v (%s)", iface, err, oneLine(out))
		}
		return nil
	}
}

// CheckIngress verifies, read-only, that the bench-owned ingress plumbing is
// present. It runs before any qdisc mutation so a missing IFB fails closed
// instead of shaping the wrong device. It never creates or changes anything.
func CheckIngress(r IngressRunner, ingressIf, ifb string) error {
	if ingressIf == "" || ifb == "" {
		return fmt.Errorf("ingress preflight: empty interface name")
	}
	out, err := r.Run("ip", "-o", "link", "show", "dev", ifb)
	if err != nil {
		return fmt.Errorf("ingress preflight: %s missing: %v (%s)", ifb, err, oneLine(out))
	}
	if !linkUp(out) {
		return fmt.Errorf("ingress preflight: %s is not UP (%s)", ifb, oneLine(out))
	}
	out, err = r.Run("tc", "qdisc", "show", "dev", ingressIf)
	if err != nil {
		return fmt.Errorf("ingress preflight: cannot read %s qdiscs: %v (%s)", ingressIf, err, oneLine(out))
	}
	if !strings.Contains(string(out), "ingress") {
		return fmt.Errorf("ingress preflight: no ingress qdisc on %s", ingressIf)
	}
	out, err = r.Run("tc", "filter", "show", "dev", ingressIf, "parent", "ffff:")
	if err != nil {
		return fmt.Errorf("ingress preflight: cannot read %s filters: %v (%s)", ingressIf, err, oneLine(out))
	}
	if !strings.Contains(string(out), "mirred") ||
		(!strings.Contains(string(out), "device "+ifb) && !strings.Contains(string(out), "dev "+ifb)) {
		return fmt.Errorf("ingress preflight: no mirred redirect to %s on %s (%s)", ifb, ingressIf, oneLine(out))
	}
	return nil
}

// linkUp reports whether an `ip -o link show` line has the UP flag. The state
// field of an IFB is UNKNOWN even when it is up, so flags are authoritative.
func linkUp(out []byte) bool {
	for _, f := range strings.Fields(string(out)) {
		if strings.HasPrefix(f, "<") && strings.HasSuffix(f, ">") {
			for _, flag := range strings.Split(strings.Trim(f, "<>"), ",") {
				if flag == "UP" {
					return true
				}
			}
		}
	}
	return false
}
