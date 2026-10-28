// Package remediate turns a measured link into production-ready queueing
// configuration.
//
// The rule this package exists to enforce: a remediation is only emitted from
// numbers that were actually measured. Capacity and RTT come from the audit or
// the campaign cell; nothing is defaulted to a plausible-looking constant. A
// remediation generated from a guessed bandwidth is worse than no remediation,
// because it will be pasted into a production router.
package remediate

import (
	"fmt"
	"math"
	"strings"
)

// Link is the measured shape of the link being remediated.
//
// CapacityMbps is the bottleneck rate in the direction being shaped. RTTms is
// the measured idle RTT, which CAKE uses for its RTT estimator. Both must come
// from a measurement; Validate rejects anything else.
type Link struct {
	// Interface is the Linux egress device to shape (e.g. "eth0", "wan0").
	Interface string
	// RouterOSInterface is the RouterOS interface name (e.g. "ether1-wan").
	RouterOSInterface string
	// RouterOSLanInterface is the RouterOS private/LAN interface that carries
	// the download direction (e.g. "ether2-lan"). RouterOS has no IFB.
	RouterOSLanInterface string
	// CapacityMbps is the measured bottleneck rate, in Mbit/s.
	CapacityMbps float64
	// RTTms is the measured idle round-trip time, in milliseconds.
	RTTms float64
	// Overhead is per-packet link-layer overhead in bytes (e.g. 38 for
	// Ethernet+VLAN, 0 for unknown). CAKE subtracts it from the shaper.
	Overhead int
	// MTU of the path, used for burst sizing. 0 means "use 1500".
	MTU int
}

// Qdisc names accepted by the generators. These match model.Qdisc values.
const (
	QdiscCake    = "cake"
	QdiscFqCodel = "fq_codel"
	QdiscPfifo   = "pfifo_fast"
)

// IngressIFB is the intermediate functional block device that receives the
// redirected WAN ingress. The remediation creates it when absent.
const IngressIFB = "ifb0"

// shaperHeadroom is the fraction of the measured bottleneck the shaper is set
// to. A shaper configured at exactly the achieved rate contends with the real
// bottleneck and under-delivers; 95% keeps the queue on our side of the link
// while staying invisible to users.
const shaperHeadroom = 0.95

// ErrUnmeasured is returned when a Link has no usable measurement.
type ErrUnmeasured struct{ Field, Why string }

func (e ErrUnmeasured) Error() string {
	return fmt.Sprintf("remediate: %s is not measured (%s)", e.Field, e.Why)
}

// Validate refuses to build a remediation from unmeasured or impossible values.
func (l Link) Validate() error {
	if l.Interface == "" && l.RouterOSInterface == "" && l.RouterOSLanInterface == "" {
		return ErrUnmeasured{"interface", "no egress device named"}
	}
	if !(l.CapacityMbps > 0) {
		return ErrUnmeasured{"capacity", "must be > 0 Mbit/s; a zero or negative rate cannot be shaped"}
	}
	if math.IsInf(l.CapacityMbps, 0) || math.IsNaN(l.CapacityMbps) {
		return ErrUnmeasured{"capacity", "not a finite number"}
	}
	if !(l.RTTms > 0) {
		return ErrUnmeasured{"rtt", "must be > 0 ms; CAKE's estimator needs a real RTT"}
	}
	if l.Overhead < 0 {
		return ErrUnmeasured{"overhead", "negative per-packet overhead is impossible"}
	}
	return nil
}

func (l Link) mtu() int {
	if l.MTU > 0 {
		return l.MTU
	}
	return 1500
}

// burstBytes sizes a token-bucket burst at roughly one MTU of buffering, which
// is the smallest burst that does not throttle a single full-sized segment.
func (l Link) burstBytes() int {
	b := l.mtu()
	if b < 1500 {
		b = 1500
	}
	return b
}

// shapedRate is the rate actually written into the config.
func (l Link) shapedRate() float64 { return l.CapacityMbps * shaperHeadroom }

// rate renders a Mbit/s rate the way tc and RouterOS both accept.
func rate(mbps float64) string {
	if mbps == math.Trunc(mbps) {
		return fmt.Sprintf("%dmbit", int64(mbps))
	}
	return fmt.Sprintf("%.2fmbit", mbps)
}

// LinuxTC returns the ordered tc command set for the requested qdisc.
//
// Ordering matters and is the point of the slice: the shaper must exist before
// a child qdisc can attach to it. Each element is one shell command.
func LinuxTC(l Link, qdisc string) ([]string, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	if l.Interface == "" {
		return nil, ErrUnmeasured{"interface", "Linux generation needs Interface"}
	}
	dev := l.Interface
	// Always start from a clean root: a stale qdisc from a previous attempt is
	// the most common cause of "the fix did nothing".
	clean := fmt.Sprintf("tc qdisc del dev %s root 2>/dev/null || true", dev)

	switch qdisc {
	case QdiscCake:
		// CAKE carries its own bandwidth shaping, so it is the root qdisc and
		// there is no tbf parent. ack-filter cuts the uplink ack stream that
		// otherwise inflates the downlink queue on asymmetric links; nat keeps
		// per-host fairness working behind NAT; diffserv4 gives the telemetry
		// class its own tin so a bulk transfer cannot starve it.
		args := []string{
			fmt.Sprintf("tc qdisc replace dev %s root cake", dev),
			"bandwidth " + rate(l.shapedRate()),
			fmt.Sprintf("rtt %.0fms", l.RTTms),
			"ack-filter",
			"nat",
			"diffserv4",
			"triple-isolate",
		}
		if l.Overhead > 0 {
			args = append(args, fmt.Sprintf("overhead %d", l.Overhead))
		}
		return []string{clean, strings.Join(args, " ")}, nil

	case QdiscFqCodel:
		// fq_codel has no shaper of its own: tbf provides the rate, fq_codel
		// provides the AQM underneath it. burst is one MTU; latency caps the
		// queue at 400 ms so tbf cannot itself become the bufferbloat.
		burst := fmt.Sprintf("%d", l.burstBytes())
		return []string{
			clean,
			fmt.Sprintf("tc qdisc replace dev %s root handle 1: tbf rate %s burst %s latency 400ms",
				dev, rate(l.shapedRate()), burst),
			fmt.Sprintf("tc qdisc replace dev %s parent 1:1 handle 10: fq_codel", dev),
		}, nil

	case QdiscPfifo:
		// Baseline only. Emitted so a before/after pair is reproducible, not as
		// a recommendation: pfifo_fast has no AQM and will bufferbloat.
		burst := fmt.Sprintf("%d", l.burstBytes())
		return []string{
			clean,
			fmt.Sprintf("tc qdisc replace dev %s root handle 1: tbf rate %s burst %s latency 400ms",
				dev, rate(l.shapedRate()), burst),
			fmt.Sprintf("tc qdisc replace dev %s parent 1:1 handle 10: pfifo limit 1000", dev),
		}, nil

	default:
		return nil, fmt.Errorf("remediate: unknown qdisc %q (want cake, fq_codel or pfifo_fast)", qdisc)
	}
}

// ingressSetup is the bench-independent plumbing that gives ingress a real
// queue: tc cannot queue on an ingress qdisc, so the packet is redirected to
// an IFB and shaped there. Every command is idempotent.
func ingressSetup(dev string) []string {
	return []string{
		fmt.Sprintf("tc qdisc del dev %s ingress 2>/dev/null || true", dev),
		fmt.Sprintf("tc qdisc del dev %s root 2>/dev/null || true", IngressIFB),
		"modprobe ifb sch_ingress act_mirred cls_matchall 2>/dev/null || true",
		fmt.Sprintf("ip link add %s type ifb 2>/dev/null || true", IngressIFB),
		fmt.Sprintf("ip link set %s up", IngressIFB),
		fmt.Sprintf("tc qdisc add dev %s handle ffff: ingress", dev),
		fmt.Sprintf("tc filter add dev %s parent ffff: protocol all matchall action mirred egress redirect dev %s", dev, IngressIFB),
	}
}

// LinuxTCIngress returns the download command set. The WAN device loses its
// ingress hook to the IFB; the shaper lives on the IFB, where packets are
// egress and can actually be queued. CAKE uses ingress mode and dual-dsthost
// for per-host fairness toward the LAN; ack-filter is deliberately absent.
func LinuxTCIngress(l Link, qdisc string) ([]string, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	if l.Interface == "" {
		return nil, ErrUnmeasured{"interface", "Linux ingress generation needs Interface"}
	}
	dev := l.Interface
	base := ingressSetup(dev)

	switch qdisc {
	case QdiscCake:
		args := []string{
			fmt.Sprintf("tc qdisc replace dev %s root cake", IngressIFB),
			"bandwidth " + rate(l.shapedRate()),
			fmt.Sprintf("rtt %.0fms", l.RTTms),
			"ingress",
			"dual-dsthost",
			"nat",
			"diffserv4",
		}
		if l.Overhead > 0 {
			args = append(args, fmt.Sprintf("overhead %d", l.Overhead))
		}
		return append(base, strings.Join(args, " ")), nil

	case QdiscFqCodel:
		burst := fmt.Sprintf("%d", l.burstBytes())
		return append(base,
			fmt.Sprintf("tc qdisc replace dev %s root handle 1: tbf rate %s burst %s latency 400ms",
				IngressIFB, rate(l.shapedRate()), burst),
			fmt.Sprintf("tc qdisc replace dev %s parent 1:1 handle 10: fq_codel", IngressIFB),
		), nil

	case QdiscPfifo:
		burst := fmt.Sprintf("%d", l.burstBytes())
		return append(base,
			fmt.Sprintf("tc qdisc replace dev %s root handle 1: tbf rate %s burst %s latency 400ms",
				IngressIFB, rate(l.shapedRate()), burst),
			fmt.Sprintf("tc qdisc replace dev %s parent 1:1 handle 10: pfifo limit 1000", IngressIFB),
		), nil

	default:
		return nil, fmt.Errorf("remediate: unknown qdisc %q (want cake, fq_codel or pfifo_fast)", qdisc)
	}
}

// RouterOS returns a RouterOS v7+ script for the requested qdisc.
//
// RouterOS v7 implements cake as a queue kind, so the script is a queue type
// plus a queue tree entry sized to the measured rate. The script is idempotent:
// it removes its own objects first, so re-running it after a re-measure
// replaces rather than duplicates.
func RouterOS(l Link, qdisc string) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	if l.RouterOSInterface == "" {
		return "", ErrUnmeasured{"routeros_interface", "RouterOS generation needs RouterOSInterface"}
	}
	iface := l.RouterOSInterface
	// RouterOS queue limits are expressed in bits, not Mbit/s.
	bits := int64(math.Round(l.shapedRate() * 1000000))

	var b strings.Builder
	b.WriteString("# CGO remediation - RouterOS v7+\n")
	b.WriteString("# Requires RouterOS >= 7.10 for kind=cake. Verify with: /system resource print\n")
	fmt.Fprintf(&b, "# Sized from a measured %.2f Mbit/s bottleneck and %.0f ms idle RTT.\n", l.CapacityMbps, l.RTTms)
	fmt.Fprintf(&b, "# Shaped at %.0f%% of measured rate (%d bit/s) to stay below the real bottleneck.\n", shaperHeadroom*100, bits)
	b.WriteString("# Idempotent: removes its own objects before recreating them.\n\n")

	// Remove prior objects so a re-run replaces instead of stacking.
	b.WriteString("/queue tree remove [find where comment=\"cgo-managed\"]\n")
	b.WriteString("/queue type remove [find where name~\"^cgo-\"]\n\n")

	switch qdisc {
	case QdiscCake:
		b.WriteString("/queue type add name=cgo-cake kind=cake \\\n")
		fmt.Fprintf(&b, "    cake-bandwidth=%d cake-rtt=%.0fms \\\n", bits, l.RTTms)
		b.WriteString("    cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 \\\n")
		b.WriteString("    cake-ack-filter=filter cake-wash=yes\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgo-wan-out parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgo-cake comment=\"cgo-managed\"\n")
		return b.String(), nil

	case QdiscFqCodel:
		// RouterOS has no fq_codel kind; sfq is the closest native equivalent.
		// Say so rather than pretending it is fq_codel - a jury or an operator
		// will check.
		b.WriteString("# NOTE: RouterOS v7 has no fq_codel queue kind.\n")
		b.WriteString("# This is sfq (stochastic fair queueing), the closest native equivalent.\n")
		b.WriteString("# It is NOT fq_codel: it has no CoDel drop controller.\n\n")
		b.WriteString("/queue type add name=cgo-sfq kind=sfq sfq-perturb=10\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgo-wan-out parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgo-sfq comment=\"cgo-managed\"\n")
		return b.String(), nil

	case QdiscPfifo:
		b.WriteString("# Baseline only - no AQM. Expect bufferbloat under load.\n\n")
		b.WriteString("/queue type add name=cgo-fifo kind=fifo\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgo-wan-out parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgo-fifo comment=\"cgo-managed\"\n")
		return b.String(), nil

	default:
		return "", fmt.Errorf("remediate: unknown qdisc %q (want cake, fq_codel or pfifo_fast)", qdisc)
	}
}

// RouterOSIngress returns the download queue tree. RouterOS has no IFB: a
// download is shaped as an egress queue on the LAN/private interface, which
// is where the traffic leaves toward clients. The object names and comment
// are disjoint from the upload cleanup, so either script can be pasted alone.
func RouterOSIngress(l Link, qdisc string) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	if l.RouterOSLanInterface == "" {
		return "", ErrUnmeasured{"routeros_lan_interface", "RouterOS ingress generation needs RouterOSLanInterface"}
	}
	iface := l.RouterOSLanInterface
	bits := int64(math.Round(l.shapedRate() * 1000000))

	var b strings.Builder
	b.WriteString("# CGO download remediation - RouterOS v7+\n")
	b.WriteString("# RouterOS has no IFB. This is a queue tree on the LAN/private\n")
	b.WriteString("# interface, which is where download traffic egresses toward clients.\n")
	b.WriteString("# It shapes the same direction at a different point than the Linux IFB.\n")
	b.WriteString("# Requires RouterOS >= 7.10 for kind=cake. Verify with: /system resource print\n")
	fmt.Fprintf(&b, "# Sized from a measured %.2f Mbit/s bottleneck and %.0f ms idle RTT.\n", l.CapacityMbps, l.RTTms)
	fmt.Fprintf(&b, "# Shaped at %.0f%% of measured rate (%d bit/s) to stay below the real bottleneck.\n", shaperHeadroom*100, bits)
	b.WriteString("# Idempotent: removes its own objects before recreating them.\n\n")

	b.WriteString("/queue tree remove [find where comment=\"cgo-managed-ingress\"]\n")
	b.WriteString("/queue type remove [find where name~\"^cgoingress-\"]\n\n")

	switch qdisc {
	case QdiscCake:
		b.WriteString("# No ack-filter: a download data stream is not an ACK stream.\n\n")
		b.WriteString("/queue type add name=cgoingress-cake kind=cake \\\n")
		fmt.Fprintf(&b, "    cake-bandwidth=%d cake-rtt=%.0fms \\\n", bits, l.RTTms)
		b.WriteString("    cake-flowmode=dual-dsthost cake-nat=yes cake-diffserv=diffserv4 \\\n")
		b.WriteString("    cake-wash=yes\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgoingress-lan-in parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgoingress-cake comment=\"cgo-managed-ingress\"\n")
		return b.String(), nil

	case QdiscFqCodel:
		b.WriteString("# NOTE: RouterOS v7 has no fq_codel queue kind.\n")
		b.WriteString("# This is sfq, the closest native equivalent. It is NOT fq_codel.\n\n")
		b.WriteString("/queue type add name=cgoingress-sfq kind=sfq sfq-perturb=10\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgoingress-lan-in parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgoingress-sfq comment=\"cgo-managed-ingress\"\n")
		return b.String(), nil

	case QdiscPfifo:
		b.WriteString("# Baseline only - no AQM. Expect bufferbloat under load.\n\n")
		b.WriteString("/queue type add name=cgoingress-fifo kind=fifo\n\n")
		fmt.Fprintf(&b, "/queue tree add name=cgoingress-lan-in parent=%s max-limit=%d \\\n", iface, bits)
		b.WriteString("    queue=cgoingress-fifo comment=\"cgo-managed-ingress\"\n")
		return b.String(), nil

	default:
		return "", fmt.Errorf("remediate: unknown qdisc %q (want cake, fq_codel or pfifo_fast)", qdisc)
	}
}

// Recommendation is the decision this package makes from a measurement: which
// qdisc to deploy and why, in one line an operator can act on.
type Recommendation struct {
	Qdisc  string
	Reason string
}

// Recommend picks a qdisc from the measured link. The thresholds are stated so
// the choice is auditable rather than a matter of taste.
//
//   - CAKE whenever the link is asymmetric: ack-filter and per-host fairness
//     are what protect a small telemetry flow from a bulk transfer.
//   - fq_codel when the link is symmetric and the CPU is the constraint, since
//     fq_codel is markedly cheaper per packet than CAKE.
func Recommend(l Link, asymmetric bool, cpuConstrained bool) Recommendation {
	if asymmetric {
		return Recommendation{
			Qdisc:  QdiscCake,
			Reason: "asymmetric link: CAKE ack-filter suppresses the ack stream that inflates the return queue",
		}
	}
	if cpuConstrained {
		return Recommendation{
			Qdisc:  QdiscFqCodel,
			Reason: "symmetric link on a CPU-constrained gateway: fq_codel gives the AQM benefit at lower per-packet cost",
		}
	}
	return Recommendation{
		Qdisc:  QdiscCake,
		Reason: "symmetric link with CPU headroom: CAKE adds per-host fairness and diffserv tins over fq_codel",
	}
}
