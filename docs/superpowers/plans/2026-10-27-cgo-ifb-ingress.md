# Ingress IFB Shaping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the missing download side of the remediation: Linux ingress shaping through an IFB, the RouterOS LAN queue-tree equivalent, a direction-aware sealed report, and an opt-in `dualns` campaign path that fails closed before it mutates a qdisc.

**Architecture:** `pkg/remediate` generates the production command sets. `pkg/qdisc` applies an IFB root shaper and reads the bench-owned ingress plumbing without mutating it. `pkg/campagne` gains one opt-in mode that shapes `ifb0` only for `bench=dualns` + `direction=down`, after a read-only preflight. `pkg/audit` selects upload, download, or both from `Result.Direction` and the measured capacities. The bench owns `ifb0` and the `mirred` redirect; the campaign owns only the qdisc on `ifb0`.

**Tech Stack:** Go 1.25, standard library only, `tc`/`ip` on Linux 6.8 (engine tests run on Windows), Bash for the `dualns` bench.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-27-cgo-ifb-ingress-design.md`.
- Branch: `rebuild/plan1-economics-audit`; do not merge to `main` in this slice.
- No new Go dependency. Standard library only.
- `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, and `data/runs/` are frozen; the frozen corpus guard is **159 runs / 208 valid lines** (never “208 runs”).
- No change to the frozen CSV schemas. Provenance stays `basis: measured|estimated`; this slice adds no eBPF `source` field.
- French comments and user-facing strings; English identifiers. The existing report headings are English because operators copy them; keep that convention.
- Commit per logical group with explicit paths. Never `git add -A`.
- Leave `epbfBBR.md` and user artifacts untracked. Only the IFB spec, plan, and roadmap are committed as the slice contract.
- Do not start Graylog. The remote PC `100.114.101.13` is off-limits.
- Never fabricate a measurement. Do not call the ingress path “verified” until Task 9 runs on Linux 6.8.
- Ingress is `direction=down` + `bench=dualns` only. `up` keeps the egress path; `both` is refused, not approximated.
- `e-wan` must be unshaped in ingress mode. A download egress CAKE there puts the bottleneck upstream of `ifb0` and the IFB queue never fills.
- CAKE `ingress` counts dropped packets as transferred. Never compare raw drop counts across egress and ingress without stating it.
- Do not claim queue dwell time at the remote link. RTT includes the netem link and the IFB queue; no kernel timestamp separates them.

## Pre-flight (before Task 1)

The spec, this plan, and the roadmap are the slice contract. Commit them first, explicitly, so execution has a stable reference:

```bash
git add docs/superpowers/specs/2026-10-27-cgo-ifb-ingress-design.md \
        docs/superpowers/plans/2026-10-27-cgo-ifb-ingress.md \
        docs/superpowers/plans/README.md
git commit -m "docs: IFB ingress spec, plan, and roadmap"
```

Do not commit `epbfBBR.md` or any file under `data/`, `backup/`, or `web/frontend/dist`.

## File Structure

| File | Responsibility |
|---|---|
| `pkg/remediate/remediate.go` | `Link.RouterOSLanInterface`; `LinuxTCIngress`; `RouterOSIngress`; `IngressIFB`. |
| `pkg/remediate/ingress_test.go` | Exact command strings, ordering, cleanup-name disjointness, refusals. |
| `pkg/qdisc/ingress.go` | `ApplyIngressShaper`; `IngressRunner`; `NsCommand`; `CheckIngress`; `linkUp`. |
| `pkg/qdisc/ingress_test.go` | Fake tc/ip runner; shaper argument vectors; every preflight failure. |
| `pkg/campagne/campagne.go` | `ShapeMode`, `IngressPreflight`, validation, preflight order, ingress placement. |
| `pkg/campagne/dualns.go` | `CGO_SHAPE_MODE` wiring, client namespace runner, `ifb0` stats. |
| `pkg/campagne/ingress_test.go` | Refusals, preflight-before-mutation, `ifb0` shaping and cleanup. |
| `pkg/campagne/dualns_test.go` | Ingress env wiring; default egress stays unchanged. |
| `pkg/audit/report.go` | Direction-aware capacities; additive ingress fields; direction-specific Markdown. |
| `pkg/audit/report_ingress_test.go` | Down-only, both, missing capacity, empty `linux_tc`, headings. |
| `cmd/cgo/main.go` | `--ros-lan`. |
| `pkg/api/server.go` | `routeros_lan` on audit start and `/api/remediate`. |
| `docs/api.md` | Additive `routeros_lan` field. |
| `kit/cgo-bench-dualns.sh` | `--ingress on|off`; `ifb0` setup, check, show; no `e-wan` CAKE. |
| `kit/run-dualns.sh` | `--ingress` pass-through. |
| `docs/methodology.fr.md` | IFB/RouterOS methodology and accounting limits. |
| `docs/superpowers/plans/README.md` | Status after execution. |

---

### Task 1: Production ingress generators

**Files:**
- Modify: `pkg/remediate/remediate.go`
- Create: `pkg/remediate/ingress_test.go`

**Interfaces:**
- Consumes: `Link`, `rate`, `l.shapedRate`, `l.burstBytes`, `l.Validate`, `ErrUnmeasured` from `remediate.go`.
- Produces:
  - `const IngressIFB = "ifb0"`
  - `Link.RouterOSLanInterface string`
  - `func LinuxTCIngress(l Link, qdisc string) ([]string, error)`
  - `func RouterOSIngress(l Link, qdisc string) (string, error)`

- [ ] **Step 1: Write the failing tests**

Create `pkg/remediate/ingress_test.go`:

```go
package remediate

import (
	"strings"
	"testing"
)

func ingressLink() Link {
	l := goodLink()
	l.RouterOSLanInterface = "ether2-lan"
	return l
}

func TestLinuxTCIngressCake(t *testing.T) {
	cmds, err := LinuxTCIngress(ingressLink(), QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(cmds, "\n")
	for _, want := range []string{
		"tc qdisc del dev wan0 ingress",
		"tc qdisc del dev ifb0 root",
		"modprobe ifb",
		"ip link add ifb0 type ifb",
		"ip link set ifb0 up",
		"tc qdisc add dev wan0 handle ffff: ingress",
		"tc filter add dev wan0 parent ffff: protocol all matchall action mirred egress redirect dev ifb0",
		"tc qdisc replace dev ifb0 root cake bandwidth 19mbit rtt 100ms ingress dual-dsthost nat diffserv4",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "ack-filter") {
		t.Fatalf("ack-filter is meaningless on a download data stream:\n%s", joined)
	}
	if strings.Contains(joined, "triple-isolate") {
		t.Fatalf("production ingress must use dual-dsthost, not the campaign default:\n%s", joined)
	}
	if strings.Index(joined, "mirred") > strings.Index(joined, "replace dev ifb0 root cake") {
		t.Fatalf("cake is attached before the redirect exists:\n%s", joined)
	}
}

func TestLinuxTCIngressFqCodelHasNoIngressKeyword(t *testing.T) {
	cmds, err := LinuxTCIngress(ingressLink(), QdiscFqCodel)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(cmds, "\n")
	if !strings.Contains(joined, "tbf rate 19mbit") {
		t.Fatalf("fq_codel is not sized from the measured rate:\n%s", joined)
	}
	if !strings.Contains(joined, "parent 1:1 handle 10: fq_codel") {
		t.Fatalf("fq_codel is not attached under the tbf:\n%s", joined)
	}
	for _, c := range cmds {
		if strings.Contains(c, "replace dev ifb0 root cake") {
			t.Fatalf("fq_codel emitted a CAKE line: %s", c)
		}
	}
}

func TestLinuxTCIngressNeedsWANInterface(t *testing.T) {
	l := ingressLink()
	l.Interface = ""
	if _, err := LinuxTCIngress(l, QdiscCake); err == nil {
		t.Fatal("emitted ingress config with no WAN interface")
	}
}

func TestRouterOSIngressCake(t *testing.T) {
	script, err := RouterOSIngress(ingressLink(), QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`comment="cgo-managed-ingress"`,
		`name~"^cgoingress-"`,
		"name=cgoingress-cake kind=cake",
		"cake-bandwidth=19000000 cake-rtt=100ms",
		"cake-flowmode=dual-dsthost",
		"cake-nat=yes",
		"cake-diffserv=diffserv4",
		"name=cgoingress-lan-in parent=ether2-lan",
		"max-limit=19000000",
		"queue=cgoingress-cake",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("missing %q:\n%s", want, script)
		}
	}
	if strings.Contains(script, "ack-filter=yes") {
		t.Fatalf("ack-filter does not belong in a download queue:\n%s", script)
	}
	if strings.Contains(script, "parent=ether1-wan") {
		t.Fatalf("download queue tree is attached to the WAN interface:\n%s", script)
	}
}

func TestRouterOSIngressNeedsLANInterface(t *testing.T) {
	l := ingressLink()
	l.RouterOSLanInterface = ""
	if _, err := RouterOSIngress(l, QdiscCake); err == nil {
		t.Fatal("emitted RouterOS ingress config with no LAN interface")
	}
}

// The upload cleanup removes types matching ^cgo- and queues commented
// cgo-managed. The ingress objects must be disjoint from both.
func TestIngressCleanupCannotDeleteUploadObjects(t *testing.T) {
	up, _ := RouterOS(goodLink(), QdiscCake)
	in, _ := RouterOSIngress(ingressLink(), QdiscCake)
	if !strings.Contains(up, `name~"^cgo-"`) || !strings.Contains(up, `comment="cgo-managed"`) {
		t.Fatalf("upload cleanup contract changed:\n%s", up)
	}
	if strings.Contains(in, `name~"^cgo-"`) || strings.Contains(in, `comment="cgo-managed"`) {
		t.Fatalf("ingress cleanup can delete upload objects:\n%s", in)
	}
	if strings.Contains(up, "cgoingress-") {
		t.Fatalf("upload script removes ingress objects:\n%s", up)
	}
}

func TestIngressRefusesUnmeasuredCapacity(t *testing.T) {
	l := ingressLink()
	l.CapacityMbps = 0
	if _, err := LinuxTCIngress(l, QdiscCake); err == nil {
		t.Fatal("emitted Linux ingress config for an unmeasured link")
	}
	if _, err := RouterOSIngress(l, QdiscCake); err == nil {
		t.Fatal("emitted RouterOS ingress config for an unmeasured link")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/remediate -run 'TestLinuxTCIngress|TestRouterOSIngress|TestIngress' -v`

Expected: FAIL with `undefined: LinuxTCIngress`, `undefined: RouterOSIngress`, and `ingressLink` field errors.

- [ ] **Step 3: Add the Link field and validate at least one device**

In `pkg/remediate/remediate.go`, add to `Link` after `RouterOSInterface`:

```go
	// RouterOSLanInterface is the RouterOS private/LAN interface that carries
	// the download direction (e.g. "ether2-lan"). RouterOS has no IFB.
	RouterOSLanInterface string
```

Change `Validate`'s first condition to accept any named device:

```go
	if l.Interface == "" && l.RouterOSInterface == "" && l.RouterOSLanInterface == "" {
		return ErrUnmeasured{"interface", "no egress device named"}
	}
```

- [ ] **Step 4: Implement `LinuxTCIngress`**

Add the constant near the qdisc constants:

```go
// IngressIFB is the intermediate functional block device that receives the
// redirected WAN ingress. The remediation creates it when absent.
const IngressIFB = "ifb0"
```

Add the helper and generator after `LinuxTC`:

```go
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
```

- [ ] **Step 5: Implement `RouterOSIngress`**

Add after `RouterOS`:

```go
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
```

- [ ] **Step 6: Run the package tests**

Run: `go test ./pkg/remediate -v`

Expected: PASS. Existing upload tests must be unchanged; the new ingress tests pass.

- [ ] **Step 7: Commit**

```bash
git add pkg/remediate/remediate.go pkg/remediate/ingress_test.go
git commit -m "remediate: ingress IFB and RouterOS LAN queue generators"
```

---

### Task 2: IFB shaper and read-only preflight

**Files:**
- Create: `pkg/qdisc/ingress.go`
- Create: `pkg/qdisc/ingress_test.go`

**Interfaces:**
- Consumes: `TCRunner`, `FakeRunner`, `oneLine`, `mbps` from `pkg/qdisc/qdisc.go`; `model.Qdisc`, `model.Cake`, `model.FqCodel`.
- Produces:
  - `type IngressRunner interface { Run(prog string, args ...string) ([]byte, error) }`
  - `type NsCommand struct{ Ns string }`
  - `func ApplyIngressShaper(r TCRunner, iface string, q model.Qdisc, capMbps, rttMs float64) error`
  - `func CheckIngress(r IngressRunner, ingressIf, ifb string) error`

- [ ] **Step 1: Write the failing tests**

Create `pkg/qdisc/ingress_test.go`:

```go
package qdisc

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

type ingressReply struct {
	out string
	err error
}

type fakeIngress struct {
	replies map[string]ingressReply
	calls   []string
}

func newFakeIngress(replies map[string]ingressReply) *fakeIngress {
	return &fakeIngress{replies: replies}
}

func (f *fakeIngress) Run(prog string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{prog}, args...), " ")
	f.calls = append(f.calls, key)
	r, ok := f.replies[key]
	if !ok {
		return nil, fmt.Errorf("unexpected command: %s", key)
	}
	return []byte(r.out), r.err
}

func healthyIngress() map[string]ingressReply {
	return map[string]ingressReply{
		"ip -o link show dev ifb0": {
			out: "2: ifb0: <BROADCAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc cake state UNKNOWN mode DEFAULT group default qlen 1000",
		},
		"tc qdisc show dev c-cli": {
			out: "qdisc ingress ffff: parent ffff:fff1 ----------------",
		},
		"tc filter show dev c-cli parent ffff:": {
			out: "filter protocol all pref 49152 matchall chain 0\n\taction order 1: mirred (Egress Redirect to device ifb0) stolen",
		},
	}
}

func TestApplyIngressShaperCake(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyIngressShaper(f, "ifb0", model.Cake, 20, 100); err != nil {
		t.Fatal(err)
	}
	got := strings.Join(f.Calls[len(f.Calls)-1], " ")
	want := "qdisc replace dev ifb0 root handle 1: cake bandwidth 20mbit rtt 100ms ingress"
	if got != want {
		t.Fatalf("cake args:\n got %q\nwant %q", got, want)
	}
}

func TestApplyIngressShaperFqCodelHasNoCakeIngress(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyIngressShaper(f, "ifb0", model.FqCodel, 20, 100); err != nil {
		t.Fatal(err)
	}
	if len(f.Calls) != 2 {
		t.Fatalf("want tbf then fq_codel, got %v", f.Calls)
	}
	first := strings.Join(f.Calls[0], " ")
	second := strings.Join(f.Calls[1], " ")
	if !strings.Contains(first, "root handle 1: tbf rate 20mbit") {
		t.Fatalf("tbf is not the root shaper: %s", first)
	}
	if second != "qdisc replace dev ifb0 parent 1:1 handle 10: fq_codel" {
		t.Fatalf("fq_codel args: %s", second)
	}
	for _, c := range f.Calls {
		if strings.Contains(strings.Join(c, " "), "ingress") {
			t.Fatalf("ingress is CAKE-only, emitted for fq_codel: %v", c)
		}
	}
}

func TestApplyIngressShaperPfifoIsTbfRoot(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyIngressShaper(f, "ifb0", model.PfifoFast, 20, 100); err != nil {
		t.Fatal(err)
	}
	if len(f.Calls) != 1 {
		t.Fatalf("pfifo baseline is one tbf, got %v", f.Calls)
	}
	if got := strings.Join(f.Calls[0], " "); !strings.Contains(got, "tbf rate 20mbit") {
		t.Fatalf("pfifo has no tbf: %s", got)
	}
}

func TestCheckIngressHealthy(t *testing.T) {
	f := newFakeIngress(healthyIngress())
	if err := CheckIngress(f, "c-cli", "ifb0"); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 3 {
		t.Fatalf("preflight must be read-only: %v", f.calls)
	}
}

func TestCheckIngressMissingDevice(t *testing.T) {
	m := healthyIngress()
	m["ip -o link show dev ifb0"] = ingressReply{err: errors.New("exit status 1")}
	if err := CheckIngress(newFakeIngress(m), "c-cli", "ifb0"); err == nil {
		t.Fatal("accepted a missing IFB")
	}
}

func TestCheckIngressDownDevice(t *testing.T) {
	m := healthyIngress()
	m["ip -o link show dev ifb0"] = ingressReply{
		out: "2: ifb0: <BROADCAST,NOARP> mtu 1500 qdisc noop state DOWN mode DEFAULT group default qlen 1000",
	}
	if err := CheckIngress(newFakeIngress(m), "c-cli", "ifb0"); err == nil {
		t.Fatal("accepted a down IFB")
	}
}

func TestCheckIngressMissingQdisc(t *testing.T) {
	m := healthyIngress()
	m["tc qdisc show dev c-cli"] = ingressReply{out: "qdisc noqueue 0: root refcnt 2"}
	if err := CheckIngress(newFakeIngress(m), "c-cli", "ifb0"); err == nil {
		t.Fatal("accepted a missing ingress qdisc")
	}
}

func TestCheckIngressMissingMirred(t *testing.T) {
	m := healthyIngress()
	m["tc filter show dev c-cli parent ffff:"] = ingressReply{
		out: "filter protocol all pref 49152 matchall chain 0\n\taction order 1: gact action drop",
	}
	if err := CheckIngress(newFakeIngress(m), "c-cli", "ifb0"); err == nil {
		t.Fatal("accepted a filter without mirred")
	}
}

func TestCheckIngressWrongTarget(t *testing.T) {
	m := healthyIngress()
	m["tc filter show dev c-cli parent ffff:"] = ingressReply{
		out: "filter protocol all pref 49152 matchall chain 0\n\taction order 1: mirred (Egress Redirect to device ifb1) stolen",
	}
	if err := CheckIngress(newFakeIngress(m), "c-cli", "ifb0"); err == nil {
		t.Fatal("accepted a redirect to the wrong IFB")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/qdisc -run 'TestApplyIngress|TestCheckIngress' -v`

Expected: FAIL with `undefined: ApplyIngressShaper`, `undefined: CheckIngress`, `undefined: IngressRunner`.

- [ ] **Step 3: Implement the runner, shaper, and preflight**

Create `pkg/qdisc/ingress.go`:

```go
package qdisc

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Realms4239/cgo/pkg/model"
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
```

- [ ] **Step 4: Run the package tests**

Run: `go test ./pkg/qdisc -v`

Expected: PASS. Existing netem/shaper/stats tests must stay unchanged.

- [ ] **Step 5: Commit**

```bash
git add pkg/qdisc/ingress.go pkg/qdisc/ingress_test.go
git commit -m "qdisc: IFB ingress shaper and read-only preflight"
```

---

### Task 3: Campaign shape mode and fail-closed ordering

**Files:**
- Modify: `pkg/campagne/campagne.go`
- Create: `pkg/campagne/ingress_test.go`

**Interfaces:**
- Consumes: `Deps`, `RunEvent`, `defaults`, `applyPhysics` from `campagne.go`; `qdisc.ApplyIngressShaper`, `qdisc.FakeRunner`.
- Produces:
  - `Deps.ShapeMode string`
  - `Deps.IngressPreflight func() error`
  - unexported `validateShapeMode(d *Deps) error`

- [ ] **Step 1: Write the failing tests**

Create `pkg/campagne/ingress_test.go`:

```go
package campagne

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

func ingressDeps(preflight func() error, shaper qdiscRunner) Deps {
	d := fastDeps()
	d.Bench = "dualns"
	d.Direction = "down"
	d.ShapeMode = "ingress"
	d.TCShaper = shaper
	d.ShaperIf = "ifb0"
	d.CliIf = "w-edge"
	d.IngressPreflight = preflight
	return d
}

func TestRunEventIngressRefusesWrongTopology(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Deps)
		want   string
	}{
		{"up", func(d *Deps) { d.Direction = "up" }, "down"},
		{"both", func(d *Deps) { d.Direction = "both" }, "down"},
		{"legacy", func(d *Deps) { d.Bench = "legacy" }, "dualns"},
		{"unknown", func(d *Deps) { d.ShapeMode = "sideways" }, "egress|ingress"},
		{"no preflight", func(d *Deps) { d.IngressPreflight = nil }, "preflight"},
		{"no shaper", func(d *Deps) { d.TCShaper = nil }, "shaper"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &qdisc.FakeRunner{}
			d := ingressDeps(func() error { return nil }, f)
			d.TC = f
			c.mutate(&d)
			ev := model.Event{Profile: "P2", Qdisc: model.Cake, CC: model.BBR}
			if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err == nil {
				t.Fatal("accepted an invalid ingress cell")
			} else if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error does not state the rule: %v", err)
			}
			if len(f.Calls) != 0 {
				t.Fatalf("qdisc mutation before validation: %v", f.Calls)
			}
		})
	}
}

func TestRunEventIngressPreflightFailsBeforeMutation(t *testing.T) {
	f := &qdisc.FakeRunner{}
	d := ingressDeps(func() error { return errors.New("ifb0 missing") }, f)
	d.TC = f
	ev := model.Event{Profile: "P2", Qdisc: model.Cake, CC: model.BBR}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err == nil {
		t.Fatal("accepted a cell whose IFB preflight failed")
	} else if !strings.Contains(err.Error(), "ifb0 missing") {
		t.Fatalf("preflight error hidden: %v", err)
	}
	if len(f.Calls) != 0 {
		t.Fatalf("qdisc mutation after a failed preflight: %v", f.Calls)
	}
}

func TestRunEventIngressShapesIFBAndCleans(t *testing.T) {
	f := &qdisc.FakeRunner{}
	d := ingressDeps(func() error { return nil }, f)
	d.TC = f
	ev := model.Event{Profile: "P2", Qdisc: model.Cake, CC: model.BBR}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, c := range f.Calls {
		all = append(all, strings.Join(c, " "))
	}
	joined := strings.Join(all, "\n")
	if !strings.Contains(joined, "qdisc replace dev ifb0 root handle 1: cake bandwidth 20mbit rtt 100ms ingress") {
		t.Fatalf("ifb0 was not shaped in ingress mode:\n%s", joined)
	}
	if !strings.Contains(joined, "qdisc del dev ifb0 root") {
		t.Fatalf("ifb0 was not reset/cleaned:\n%s", joined)
	}
	if strings.Contains(joined, "dev e-wan") {
		t.Fatalf("ingress mode shaped e-wan as well:\n%s", joined)
	}
}

func TestRunEventEgressDoesNotUseIngress(t *testing.T) {
	f := &qdisc.FakeRunner{}
	d := fastDeps()
	d.TC = f
	d.TCShaper = f
	d.Bench = "dualns"
	ev := model.Event{Profile: "P2", Qdisc: model.Cake, CC: model.BBR}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, "ifb0") || strings.Contains(joined, "ingress") {
			t.Fatalf("default egress path used ingress plumbing: %s", joined)
		}
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/campagne -run 'TestRunEventIngress|TestRunEventEgressDoesNotUseIngress' -v`

Expected: FAIL with `unknown field ShapeMode` and `unknown field IngressPreflight`.

- [ ] **Step 3: Add the Deps fields and default**

In `pkg/campagne/campagne.go`, add after `Direction`:

```go
	// ShapeMode — "egress" (défaut) ou "ingress" (IFB sur l'ingress WAN).
	// ingress exige bench=dualns + direction=down + un preflight non nil.
	ShapeMode string
	// IngressPreflight — lecture seule de ifb0/ingress/mirred, AVANT toute
	// mutation qdisc. nil en mode ingress ⇒ la cellule est refusée.
	IngressPreflight func() error
```

In `defaults`, after the `Bench` default:

```go
	if d.ShapeMode == "" {
		d.ShapeMode = "egress"
	}
```

- [ ] **Step 4: Add the validator and the preflight order**

Add after `defaults`:

```go
// validateShapeMode refuses an ingress cell whose topology or ownership is
// wrong, before any qdisc mutation. The bench owns ifb0 and the redirect; the
// campaign owns only the qdisc on the IFB.
func validateShapeMode(d *Deps) error {
	switch d.ShapeMode {
	case "egress":
		return nil
	case "ingress":
		if d.Bench != "dualns" {
			return fmt.Errorf("shape_mode ingress invalide sur bench %q — dualns uniquement", d.Bench)
		}
		if d.Direction != "down" {
			return fmt.Errorf("shape_mode ingress invalide en direction %q — down uniquement (le shaper ingress ne corrige pas l'upload)", d.Direction)
		}
		if d.TCShaper == nil || d.ShaperIf == "" {
			return fmt.Errorf("shape_mode ingress sans shaper/iface — ifb0 et son runner sont requis")
		}
		if d.IngressPreflight == nil {
			return fmt.Errorf("shape_mode ingress sans preflight — ifb0/ingress/mirred non vérifiés, cellule refusée")
		}
		return nil
	default:
		return fmt.Errorf("shape_mode %q invalide (egress|ingress)", d.ShapeMode)
	}
}
```

In `RunEvent`, after `ge := burstGEOrNull(...)` and before `applyPhysics`:

```go
	if err := validateShapeMode(&d); err != nil {
		return ev, err
	}
	if d.ShapeMode == "ingress" {
		if err := d.IngressPreflight(); err != nil {
			return ev, fmt.Errorf("ingress preflight: %w", err)
		}
	}
```

- [ ] **Step 5: Add the ingress placement branch**

In `RunEvent`, replace the existing `capDown`/`if d.Direction == "down"` block with:

```go
	capDown := prof.CapDown()
	if d.ShapeMode == "ingress" {
		if err := qdisc.ApplyIngressShaper(shaper, shaperIf, ev.Qdisc, capDown, prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper ingress: %w", err)
		}
	} else if d.Direction == "down" {
		if err := qdisc.ApplyShaper(shaper, shaperIf, ev.Qdisc, capDown, prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper down: %w", err)
		}
	} else {
		if err := qdisc.ApplyShaper(shaper, shaperIf, ev.Qdisc, prof.CapUp(), prof.DelayMs); err != nil {
			return ev, fmt.Errorf("shaper: %w", err)
		}
	}
```

The existing reset/defer block stays exactly as it is: in ingress mode `shaperIf` is `ifb0`, so it deletes the IFB root qdisc before the cell and again on exit. It does not touch the bench-owned `c-cli` ingress qdisc or `mirred` filter.

- [ ] **Step 6: Run the tests**

Run: `go test ./pkg/campagne -run 'TestRunEventIngress|TestRunEventEgressDoesNotUseIngress' -v`

Expected: PASS (six refusal subtests, the preflight failure, the IFB success, and the unchanged-egress test).

- [ ] **Step 7: Run the package suite**

Run: `go test ./pkg/campagne -v`

Expected: PASS. Existing download/up/bursty tests unchanged.

- [ ] **Step 8: Commit**

```bash
git add pkg/campagne/campagne.go pkg/campagne/ingress_test.go
git commit -m "campagne: opt-in ingress IFB shape mode with preflight"
```

---

### Task 4: `dualns` ingress environment wiring

**Files:**
- Modify: `pkg/campagne/dualns.go`
- Modify: `pkg/campagne/dualns_test.go`

**Interfaces:**
- Consumes: `qdisc.NsRunner`, `qdisc.NsCommand`, `qdisc.CheckIngress`, `qdisc.PollStats`, `Deps.ShapeMode`, `Deps.IngressPreflight`.
- Produces: `ProdDepsDualNS` returns ingress wiring when `CGO_SHAPE_MODE=ingress`.

- [ ] **Step 1: Write the failing test**

Append to `pkg/campagne/dualns_test.go`:

```go
func TestProdDepsDualNSIngressWiring(t *testing.T) {
	t.Setenv("CGO_SHAPE_MODE", "ingress")
	t.Setenv("CGO_CLIENT_NS", "ns_test_client")
	t.Setenv("CGO_INGRESS_IF", "c-test")
	t.Setenv("CGO_INGRESS_DEV", "ifb9")

	d := ProdDepsDualNS("down")
	if d.ShapeMode != "ingress" {
		t.Fatalf("shape mode lost: %q", d.ShapeMode)
	}
	if d.ShaperIf != "ifb9" {
		t.Fatalf("shaper is not on the IFB: %q", d.ShaperIf)
	}
	if d.IngressPreflight == nil {
		t.Fatal("ingress preflight not wired")
	}
	if d.StatsFn == nil {
		t.Fatal("ifb0 stats not wired")
	}
	r, ok := d.TCShaper.(qdisc.NsRunner)
	if !ok || r.Ns != "ns_test_client" {
		t.Fatalf("shaper runner is not the client namespace: %#v", d.TCShaper)
	}
}
```

Update `TestProdDepsDualNSWiring` so an ambient environment cannot change the default path. At the top of that test add:

```go
	t.Setenv("CGO_SHAPE_MODE", "egress")
```

And after the `up` assertions add:

```go
	if up.ShapeMode != "egress" || up.IngressPreflight != nil {
		t.Fatalf("default path must stay egress: %+v", up)
	}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/campagne -run 'TestProdDepsDualNS' -v`

Expected: FAIL. `ShapeMode` is empty, `ShaperIf` is `e-wan`, and `IngressPreflight` is nil.

- [ ] **Step 3: Wire the environment**

Replace the body of `ProdDepsDualNS` after `edge := ...` with a `deps` value and an ingress branch:

```go
	edge := qdisc.NsRunner{Ns: env("CGO_EDGE_NS", "ns_edge_router")}
	shapeMode := env("CGO_SHAPE_MODE", "egress")
	deps := Deps{
		TC:            qdisc.NsRunner{Ns: env("CGO_WAN_NS", "ns_wan_emulator")},
		TCShaper:      edge,
		CliIf:         env("CGO_CLI_IF", "w-edge"),
		PhysicsIfaces: []string{env("CGO_DOWNLINK_PHYS_IF", "c-wan")},
		ShaperIf:      shaperIf,
		Target:        target,
		SmallURL:      small,
		BulkAddr:      bulk,
		Direction:     direction,
		Bench:         "dualns",
		ChannelModel:  env("CGO_CHANNEL_MODEL", "clean"),
		ShapeMode:     shapeMode,
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
```

- [ ] **Step 4: Run the package tests**

Run: `go test ./pkg/campagne -v`

Expected: PASS. `ProdDepsDualNS("up")` and `("down")` without the env keep `e-srv`/`e-wan` and `ShapeMode=="egress"`.

- [ ] **Step 5: Commit**

```bash
git add pkg/campagne/dualns.go pkg/campagne/dualns_test.go
git commit -m "campagne: wire dualns ingress IFB environment"
```

---

### Task 5: Direction-aware audit report

**Files:**
- Modify: `pkg/audit/report.go`
- Create: `pkg/audit/report_ingress_test.go`

**Interfaces:**
- Consumes: `remediate.LinuxTCIngress`, `remediate.RouterOSIngress`, `remediate.Link.RouterOSLanInterface`.
- Produces:
  - `Remediation.LinuxTCIngress []string` (`json:"linux_tc_ingress,omitempty"`)
  - `Remediation.RouterOSIngress string` (`json:"routeros_ingress,omitempty"`)

- [ ] **Step 1: Write the failing tests**

Create `pkg/audit/report_ingress_test.go`:

```go
package audit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
)

func testIngressLink() remediate.Link {
	l := testLink()
	l.RouterOSLanInterface = "ether2-lan"
	return l
}

func TestReportDownEmitsOnlyIngress(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	// The down window is the headline, so the measured capacity is
	// ThroughputMbps; ThroughputDownMbps is the `both` mirror and stays 0.
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("no remediation for a measured down link")
	}
	if len(rep.Remediation.LinuxTC) != 0 {
		t.Fatalf("down report emitted an upload script: %v", rep.Remediation.LinuxTC)
	}
	if len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("down report emitted no ingress script")
	}
	joined := strings.Join(rep.Remediation.LinuxTCIngress, "\n")
	if !strings.Contains(joined, "ifb0") || !strings.Contains(joined, " ingress") {
		t.Fatalf("ingress script is not an IFB ingress config:\n%s", joined)
	}
	if !strings.Contains(joined, "4.66mbit") {
		t.Fatalf("ingress script is not sized from the measured down rate:\n%s", joined)
	}
	if rep.Remediation.RouterOS != "" {
		t.Fatalf("down report emitted an upload RouterOS script:\n%s", rep.Remediation.RouterOS)
	}
	if !strings.Contains(rep.Remediation.RouterOSIngress, "parent=ether2-lan") {
		t.Fatalf("download queue tree is not on the LAN interface:\n%s", rep.Remediation.RouterOSIngress)
	}
	if !strings.Contains(rep.Remediation.Reason, "download path") {
		t.Fatalf("down reason still justifies upload ack-filter: %q", rep.Remediation.Reason)
	}
}

func TestReportDownWithoutLANStillEmitsLinuxIngress(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	rep, err := BuildReport(res, testLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("down report lost Linux ingress when no LAN interface was named")
	}
	if rep.Remediation.RouterOSIngress != "" {
		t.Fatal("emitted RouterOS ingress without a LAN interface")
	}
}

func TestReportDownHasEmptyLinuxTCArray(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	js, err := json.Marshal(rep.Remediation)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(js), `"linux_tc":[]`) {
		t.Fatalf("down remediation must carry an empty linux_tc array, not null:\n%s", js)
	}
	if strings.Contains(string(js), `"linux_tc":null`) {
		t.Fatalf("linux_tc is null:\n%s", js)
	}
}

func TestReportBothEmitsBothDirections(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputDownMbps = 10
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("no remediation for a measured both link")
	}
	up := strings.Join(rep.Remediation.LinuxTC, "\n")
	in := strings.Join(rep.Remediation.LinuxTCIngress, "\n")
	if !strings.Contains(up, "4.66mbit") {
		t.Fatalf("upload script is not sized from ThroughputMbps:\n%s", up)
	}
	if !strings.Contains(in, "9.50mbit") {
		t.Fatalf("download script is not sized from ThroughputDownMbps:\n%s", in)
	}
	if !strings.Contains(rep.Remediation.RouterOS, "cgo-") {
		t.Fatal("both report has no upload RouterOS script")
	}
	if !strings.Contains(rep.Remediation.RouterOSIngress, "cgoingress-") {
		t.Fatal("both report has no download RouterOS script")
	}
	md := rep.RenderMarkdown()
	for _, want := range []string{"upload (egress)", "download (ingress IFB)", "download (LAN queue tree)"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown heading %q missing:\n%s", want, md)
		}
	}
}

func TestReportBothMissingDownCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputDownMbps = 0
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTC) == 0 {
		t.Fatal("the measured upload remediation was lost")
	}
	if len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatal("a download config was sized from the upload number")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation (download)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing down capacity is silent: %+v", rep.Unavailable)
	}
}

func TestReportBothMissingUpCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "both"
	res.ThroughputMbps = 0
	res.ThroughputDownMbps = 10
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil || len(rep.Remediation.LinuxTCIngress) == 0 {
		t.Fatal("the measured download remediation was lost")
	}
	if len(rep.Remediation.LinuxTC) != 0 {
		t.Fatal("an upload config was sized from the download number")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation (upload)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing up capacity is silent: %+v", rep.Unavailable)
	}
}

func TestReportDownMissingCapacityIsNamed(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	res.ThroughputMbps = 0
	rep, err := BuildReport(res, testIngressLink(), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation != nil {
		t.Fatal("emitted a remediation for an unmeasured down link")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "remediation" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unmeasured down remediation is silent: %+v", rep.Unavailable)
	}
}

func TestReportDownWithoutLinuxInterfaceStillNamesRouterOS(t *testing.T) {
	res := loadedResult()
	res.Direction = "down"
	l := testIngressLink()
	l.Interface = ""
	rep, err := BuildReport(res, l, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation == nil {
		t.Fatal("RouterOS ingress was lost with no Linux interface")
	}
	if len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatal("invented a Linux interface")
	}
	if rep.Remediation.RouterOSIngress == "" {
		t.Fatal("RouterOS ingress not generated from its own interface")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/audit -run 'TestReportDown|TestReportBoth' -v`

Expected: FAIL: the down report still emits upload commands, and `LinuxTCIngress`/`RouterOSIngress` are unknown fields.

- [ ] **Step 3: Add the additive fields**

In `pkg/audit/report.go`, extend `Remediation`:

```go
	// LinuxTCIngress is the ordered download command set for a down (or both)
	// measurement. It stays an empty array for an upload-only report.
	LinuxTCIngress []string `json:"linux_tc_ingress,omitempty"`
	// RouterOSIngress is the LAN queue-tree script for the download direction.
	RouterOSIngress string `json:"routeros_ingress,omitempty"`
```

- [ ] **Step 4: Replace the remediation block**

Replace the whole block from `// Remediation only from a real measurement.` through the closing brace before `return rep, nil` with:

```go
	// Remediation only from a real measurement. This is the guard that keeps a
	// guessed bandwidth out of a production router config. A `down` headline
	// already lives in ThroughputMbps; a `both` result has separate mirrors.
	upCap, downCap := 0.0, 0.0
	switch res.Direction {
	case "down":
		downCap = res.ThroughputMbps
	case "both":
		upCap = res.ThroughputMbps
		downCap = res.ThroughputDownMbps
	default: // "", "up"
		upCap = res.ThroughputMbps
	}
	if res.RTTIdleP50 <= 0 || (upCap <= 0 && downCap <= 0) {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "remediation",
			Reason: "capacity or idle RTT unmeasured; a queue config sized from a guess would be deployed blindly",
		})
	} else {
		link.RTTms = res.RTTIdleP50
		rec := remediate.Recommend(link, asymmetric, cpuConstrained)
		useQdisc := qdisc
		if useQdisc == "" {
			useQdisc = rec.Qdisc
		}
		rem := &Remediation{Qdisc: useQdisc, Reason: rec.Reason, LinuxTC: []string{}}
		if upCap > 0 {
			link.CapacityMbps = upCap
			if cmds, err := remediate.LinuxTC(link, useQdisc); err == nil {
				rem.LinuxTC = cmds
			}
			if link.RouterOSInterface != "" {
				if script, err := remediate.RouterOS(link, useQdisc); err == nil {
					rem.RouterOS = script
				}
			}
		} else if res.Direction == "both" {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation (upload)",
				Reason: "upload throughput unmeasured; the download remediation is not promoted to an upload config",
			})
		}
		if downCap > 0 {
			link.CapacityMbps = downCap
			if cmds, err := remediate.LinuxTCIngress(link, useQdisc); err == nil {
				rem.LinuxTCIngress = cmds
			}
			if link.RouterOSLanInterface != "" {
				if script, err := remediate.RouterOSIngress(link, useQdisc); err == nil {
					rem.RouterOSIngress = script
				}
			}
		} else if res.Direction == "both" {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation (download)",
				Reason: "download throughput unmeasured; the upload remediation is not promoted to a download config",
			})
		}
		if res.Direction == "down" {
			rem.Reason = "download path: the shaper queues packets after they cross the link (Linux IFB ingress or RouterOS LAN queue tree); the upload ack-filter does not apply to a download data stream"
		}
		if len(rem.LinuxTC) > 0 || len(rem.LinuxTCIngress) > 0 || rem.RouterOS != "" || rem.RouterOSIngress != "" {
			rep.Remediation = rem
		} else {
			rep.Unavailable = append(rep.Unavailable, Unavailable{
				Metric: "remediation",
				Reason: "no interface was named, so no command set could be generated",
			})
		}
	}
```

- [ ] **Step 5: Render direction-specific Markdown**

Replace the `if r.Remediation != nil { ... }` block at the end of `RenderMarkdown` with:

```go
	if r.Remediation != nil {
		b.WriteString("## Remediation\n\n")
		fmt.Fprintf(&b, "Recommended qdisc: **%s** — %s\n\n", r.Remediation.Qdisc, r.Remediation.Reason)
		if len(r.Remediation.LinuxTC) > 0 {
			b.WriteString("### Linux — upload (egress)\n\n```sh\n")
			for _, c := range r.Remediation.LinuxTC {
				b.WriteString(c + "\n")
			}
			b.WriteString("```\n\n")
		}
		if len(r.Remediation.LinuxTCIngress) > 0 {
			b.WriteString("### Linux — download (ingress IFB)\n\n```sh\n")
			for _, c := range r.Remediation.LinuxTCIngress {
				b.WriteString(c + "\n")
			}
			b.WriteString("```\n\n")
			if r.Remediation.Qdisc == remediate.QdiscCake {
				b.WriteString("CAKE runs in `ingress` mode on `ifb0`: the packet has already crossed the link when it is dropped, and `ack-filter` does not apply to a download data stream.\n\n")
			} else {
				b.WriteString("The download is shaped on `ifb0` after the packet has crossed the link; `ingress` is CAKE-only and is not used with this qdisc, and `ack-filter` does not apply to a download data stream.\n\n")
			}
		}
		if r.Remediation.RouterOS != "" {
			b.WriteString("### MikroTik RouterOS v7+ — upload (WAN)\n\n```\n")
			b.WriteString(r.Remediation.RouterOS)
			b.WriteString("```\n\n")
		}
		if r.Remediation.RouterOSIngress != "" {
			b.WriteString("### MikroTik RouterOS v7+ — download (LAN queue tree)\n\n```\n")
			b.WriteString(r.Remediation.RouterOSIngress)
			b.WriteString("```\n\n")
			b.WriteString("RouterOS has no IFB: this is a LAN-interface egress queue tree, not byte-equivalent to the Linux IFB.\n\n")
		}
	}
```

- [ ] **Step 6: Run the audit tests**

Run: `go test ./pkg/audit -v`

Expected: PASS. Existing upload/report/provenance tests stay unchanged; the fixture file `data/audit-report-fixture.md` may be rewritten with the new upload heading, which is untracked and not a contract.

- [ ] **Step 7: Commit**

```bash
git add pkg/audit/report.go pkg/audit/report_ingress_test.go
git commit -m "audit: direction-aware ingress remediation report"
```

---

### Task 6: CLI and API interface

**Files:**
- Modify: `cmd/cgo/main.go`
- Modify: `pkg/api/server.go`
- Modify: `docs/api.md`

**Interfaces:**
- Consumes: `remediate.Link.RouterOSLanInterface`, `Remediation` fields.
- Produces: `--ros-lan`, `routeros_lan` on `POST /api/audit/start` and `GET /api/remediate`.

- [ ] **Step 1: Add the CLI flag**

In `cmd/cgo/main.go`, after the `routeros-iface` flag:

```go
		rosLan := fs.String("ros-lan", "", "RouterOS LAN/private interface for the download remediation script (e.g. ether2-lan)")
```

Replace the link construction with:

```go
		link := remediate.Link{
			Interface:            *iface,
			RouterOSInterface:    *rosIface,
			RouterOSLanInterface: *rosLan,
		}
```

- [ ] **Step 2: Add the API body field and plumbing**

In `pkg/api/server.go`, add to the audit start body after `RouterOSIface`:

```go
			RouterOSLan   string `json:"routeros_lan"`
```

Replace the link construction after `audit.Run` with:

```go
			link := remediate.Link{
				Interface:            body.Iface,
				RouterOSInterface:    body.RouterOSIface,
				RouterOSLanInterface: body.RouterOSLan,
			}
```

In the `/api/remediate` handler, replace the rebuilt link with:

```go
				link := remediate.Link{
					Interface:            iface,
					RouterOSInterface:    strings.TrimSpace(q.Get("routeros_iface")),
					RouterOSLanInterface: strings.TrimSpace(q.Get("routeros_lan")),
				}
```

- [ ] **Step 3: Update the API contract**

In `docs/api.md`, add `routeros_lan?` to the `/api/audit/start` body list, immediately after `routeros_iface?`:

```
`{site, link_type, provider, duration?, target?, direction?, flows?, small_url?, bulk_addr?, iface?, routeros_iface?, routeros_lan?, qdisc?, asymmetric?, cpu_constrained?}`
```

Do not change the response shapes; the ingress fields are additive inside the existing `<Report>`.

- [ ] **Step 4: Build and run the affected tests**

Run: `go test ./cmd/cgo ./pkg/api`

Expected: PASS. No new test is added: the flag and body fields are plumbing into `BuildReport`, whose direction logic is already covered by Task 5.

- [ ] **Step 5: Commit**

```bash
git add cmd/cgo/main.go pkg/api/server.go docs/api.md
git commit -m "cli/api: expose routeros_lan for ingress remediation"
```

---

### Task 7: Opt-in `dualns` bench plumbing

**Files:**
- Modify: `kit/cgo-bench-dualns.sh`
- Modify: `kit/run-dualns.sh`

**Interfaces:**
- Consumes: existing `NS_CLIENT`, `NS_EDGE`, `RATE_MBIT`, `RTT_MS`, `SUDO`, `die`, `log`, `check_offloads`.
- Produces: `cgo-bench-dualns.sh up|check|show --ingress on|off`; `run-dualns.sh up|serve --ingress on|off`.

- [ ] **Step 1: Add the ingress flag and module loading**

In `kit/cgo-bench-dualns.sh`, add near `OFFLOADS=off`:

```bash
# Download ingress is opt-in. The campaign's mode selects the bench mode too,
# so `cgo run` preflight (which has no flag of its own) checks the same
# topology through the inherited environment.
INGRESS=off
if [ "${CGO_SHAPE_MODE:-}" = "ingress" ]; then INGRESS=on; fi
```

Extend the usage comment:

```bash
#    cgo-bench-dualns.sh up    [--delay MS] [--loss PCT] [--rate MBIT] [--rtt MS] [--ingress on|off]
```

In `need_modules`, after the existing `modprobe` line:

```bash
  if [ "$INGRESS" = "on" ]; then
    $SUDO modprobe ifb sch_ingress act_mirred cls_matchall 2>/dev/null || true
  fi
```

`check` and `show` must be told the mode too: pass `--ingress on` or export
`CGO_SHAPE_MODE=ingress`. `up --ingress on` does not persist a mode for later
invocations.

- [ ] **Step 2: Create the IFB plumbing and make `e-wan` conditional**

In `up`, after the two netem commands and before the queue-policy block, insert:

```bash
  if [ "$INGRESS" = "on" ]; then
    if ! $SUDO ip netns exec "$NS_CLIENT" ip link show ifb0 >/dev/null 2>&1; then
      $SUDO ip netns exec "$NS_CLIENT" ip link add ifb0 type ifb \
        || die "cannot create ifb0 (modprobe ifb failed)"
    fi
    $SUDO ip netns exec "$NS_CLIENT" ip link set ifb0 up \
      || die "cannot bring ifb0 up"
    $SUDO ip netns exec "$NS_CLIENT" tc qdisc del dev c-cli ingress 2>/dev/null || true
    $SUDO ip netns exec "$NS_CLIENT" tc qdisc add dev c-cli handle ffff: ingress \
      || die "cannot add ingress qdisc on c-cli"
    $SUDO ip netns exec "$NS_CLIENT" tc filter add dev c-cli parent ffff: protocol all \
      matchall action mirred egress redirect dev ifb0 \
      || die "cannot redirect c-cli ingress to ifb0"
    $SUDO ip netns exec "$NS_CLIENT" tc qdisc replace dev ifb0 root cake \
      bandwidth "${RATE_MBIT}mbit" rtt "${RTT_MS}ms" ingress \
      || die "cannot put cake on ifb0"
  fi
```

Replace the unconditional `e-wan` CAKE with:

```bash
  if [ "$INGRESS" = "on" ]; then
    # The download is shaped on ifb0. An egress CAKE here would put the
    # bottleneck upstream of the IFB and its queue would never fill.
    $SUDO ip netns exec "$NS_EDGE" tc qdisc del dev e-wan root 2>/dev/null || true
  else
    $SUDO ip netns exec "$NS_EDGE" tc qdisc replace dev e-wan root cake \
      bandwidth "${RATE_MBIT}mbit" rtt "${RTT_MS}ms" \
      ack-filter nat diffserv4 triple-isolate
  fi
```

Replace the final `log "up — ..."` line with:

```bash
  if [ "$INGRESS" = "on" ]; then
    log "up — netem ${DELAY_MS}ms/${LOSS_PCT}% on w-edge+c-wan ($NS_WAN); cake ${RATE_MBIT}mbit on e-srv + ifb0 ingress ($NS_CLIENT); e-wan unshaped"
  else
    log "up — netem ${DELAY_MS}ms/${LOSS_PCT}% on w-edge ($NS_WAN); cake ${RATE_MBIT}mbit on e-srv+e-wan ($NS_EDGE)"
  fi
```

- [ ] **Step 3: Make `check` mode-aware**

In `check`, keep the namespace, netem, reachability, route, and offload checks. Replace the `down_q` cake assertion and add the ingress assertions:

```bash
  if [ "$INGRESS" = "on" ]; then
    local ifb_link ifb_q ingress_q filter_q
    ifb_link=$($SUDO ip netns exec "$NS_CLIENT" ip -o link show dev ifb0 2>/dev/null)
    echo "$ifb_link" | grep -Eq '<([^>]*,)?UP(,|>)' \
      || { echo "[!] ifb0 missing or down: $ifb_link"; ok=0; }
    ifb_q=$($SUDO ip netns exec "$NS_CLIENT" tc qdisc show dev ifb0 2>/dev/null)
    ingress_q=$($SUDO ip netns exec "$NS_CLIENT" tc qdisc show dev c-cli 2>/dev/null)
    filter_q=$($SUDO ip netns exec "$NS_CLIENT" tc filter show dev c-cli parent ffff: 2>/dev/null)
    echo "$ifb_q" | grep -q "cake" \
      || { echo "[!] no cake on ifb0: $ifb_q"; ok=0; }
    echo "$ifb_q" | grep -q "ingress" \
      || { echo "[!] cake on ifb0 is not in ingress mode: $ifb_q"; ok=0; }
    echo "$ingress_q" | grep -q "ingress" \
      || { echo "[!] no ingress qdisc on c-cli: $ingress_q"; ok=0; }
    if ! echo "$filter_q" | grep -q "mirred" || ! echo "$filter_q" | grep -q "ifb0"; then
      echo "[!] no mirred redirect to ifb0: $filter_q"; ok=0
    fi
    if echo "$down_q" | grep -q "cake"; then
      echo "[!] cake on e-wan in ingress mode — bottleneck upstream of ifb0: $down_q"; ok=0
    fi
  else
    echo "$down_q" | grep -q cake \
      || { echo "[!] no cake on e-wan (download path unshaped): $down_q"; ok=0; }
  fi
```

The existing `up_q` cake check stays unconditional. The existing netem/cake stacking checks stay; in ingress mode `down_q` is empty or noqueue and cannot trip them.

- [ ] **Step 4: Make `show` mode-aware**

Replace the `e-wan` block in `show` with:

```bash
  if [ "$INGRESS" = "on" ]; then
    echo "=== $NS_CLIENT (ifb0) — download queue policy (ingress IFB) ==="
    $SUDO ip netns exec "$NS_CLIENT" tc qdisc show dev ifb0 2>/dev/null
    echo
    echo "=== $NS_CLIENT (c-cli) — ingress redirect ==="
    $SUDO ip netns exec "$NS_CLIENT" tc qdisc show dev c-cli 2>/dev/null
    $SUDO ip netns exec "$NS_CLIENT" tc filter show dev c-cli parent ffff: 2>/dev/null
  else
    echo "=== $NS_EDGE (e-wan) — download queue policy ==="
    $SUDO ip netns exec "$NS_EDGE" tc qdisc show dev e-wan 2>/dev/null
  fi
```

- [ ] **Step 5: Parse and validate the flag**

In the option parser, add:

```bash
    --ingress) INGRESS="$2"; shift 2;;
```

After the parser and before the `case "$CMD"` dispatch, add:

```bash
case "$INGRESS" in
  on|off) ;;
  *) echo "ingress must be on|off" >&2; exit 2;;
esac
```

Update the final usage line:

```bash
  *) echo "usage: $0 up|down|check|show [--delay MS] [--loss PCT] [--rate MBIT] [--rtt MS] [--offloads on|off] [--ingress on|off]"; exit 2;;
```

- [ ] **Step 6: Pass the flag through `run-dualns.sh`**

In `kit/run-dualns.sh`, add to the header options:

```bash
#  Options: --bin PATH (défaut $CGO_BIN ou ~/cgo/cgo-linux)
#           --ingress on|off (façonnage descendant IFB, opt-in)
```

Add `INGRESS=""` before the parser. In the parser case, add:

```bash
    --ingress) INGRESS="$2"; shift 2;;
```

After the parser, validate:

```bash
case "$INGRESS" in
  ""|on|off) ;;
  *) echo "ingress must be on|off" >&2; exit 2;;
esac
```

Replace the `up` call in the `up|serve` branch with:

```bash
    if [ -n "$INGRESS" ]; then
      "$HERE/cgo-bench-dualns.sh" up --ingress "$INGRESS" || exit 1
    else
      "$HERE/cgo-bench-dualns.sh" up || exit 1
    fi
```

Do not set `CGO_SHAPE_MODE` in `run-dualns.sh`; the campaign env stays the operator's explicit choice.

Update the final usage line to include the new option:

```bash
  *) echo "usage: $0 up|serve|down [--bin PATH] [--ingress on|off]" >&2; exit 2;;
```

- [ ] **Step 7: Syntax-check and inspect the default path**

Run: `bash -n kit/cgo-bench-dualns.sh kit/run-dualns.sh`

Expected: no output.

Run: `grep -n "INGRESS\|ifb0\|e-wan" kit/cgo-bench-dualns.sh kit/run-dualns.sh`

Expected: every `ifb0` line is inside an `INGRESS = on` branch; the `else` branch still applies CAKE to `e-wan`.

- [ ] **Step 8: Commit**

```bash
git add kit/cgo-bench-dualns.sh kit/run-dualns.sh
git commit -m "bench: opt-in ifb0 ingress plumbing for dualns"
```

---

### Task 8: Methodology and Windows gates

**Files:**
- Modify: `docs/methodology.fr.md`
- Modify: `docs/superpowers/plans/README.md`

**Interfaces:**
- Consumes: the behaviour implemented in Tasks 1–7.
- Produces: the methodology section and the gate record.

- [ ] **Step 1: Add the methodology section**

Append to `docs/methodology.fr.md`:

```markdown
## 9. Façonnage descendant : IFB (Linux) et file LAN (RouterOS)

Le façonnage egress ne corrige pas un téléchargement : quand le paquet arrive,
il a déjà traversé le lien. Linux le résout par un **IFB** : l'ingress de
l'interface WAN est redirigé par `matchall`/`mirred` vers `ifb0`, puis un qdisc
racine ordinaire met le paquet en file sur `ifb0` — côté réception.

### 9.1 Banc `dualns`

- `ifb0` vit dans `ns_client` ; le redirect et le qdisc ingress vivent sur
  `c-cli`, l'interface d'accès mesurée.
- Le CAKE descendant de `e-wan` est **absent** en mode ingress. S'il restait,
  le goulot serait en amont de l'IFB et la file de l'IFB ne se formerait
  jamais.
- Le CAKE montant de `e-srv` reste en place : il n'est pas sur le chemin
  descendant.
- Le CAKE par défaut sur `ifb0` est un confort de banc (`up --ingress on`) ;
  une cellule de campagne remplace la racine `ifb0` puis la supprime en
  sortie. Après une campagne, `check` peut donc signaler l'absence de CAKE
  sur `ifb0` jusqu'au prochain `up --ingress on`.
- Le mode du banc et celui de la campagne doivent coïncider :
  `CGO_SHAPE_MODE=ingress` sert aussi de défaut à `up`/`check`/`show` du banc
  lorsque `--ingress` est absent, pour que le pré-contrôle de `cgo run` voie
  la même topologie.
- Le banc ne modélise pas le LAN derrière le client : le téléchargement
  aboutit dans le même namespace que les sondes. C'est un raccourci de banc
  documenté, pas une topologie CPE complète.

### 9.2 Comptabilité

- Les pertes à l'IFB sont des pertes **hôtes**, après traversée du lien
  émulé ; elles ne prouvent pas une perte au lien distant et ne libèrent pas
  de capacité.
- En mode `ingress`, CAKE compte les paquets jetés comme transférés et
  maintient au moins deux paquets par flux. Un compteur de pertes ingress
  n'est donc pas comparable tel quel à un compteur egress ; toute table qui
  les compare doit le dire.
- Le RTT mesuré inclut le lien netem et la file IFB. Aucun horodatage noyau
  ne sépare les deux : on ne présente pas `RTT − repos` comme un temps de
  séjour dans la file du lien distant.
- Le mode `ingress` est opt-in (`CGO_SHAPE_MODE=ingress`) et limité à
  `bench=dualns` + `direction=down`. `up` garde le chemin egress ; `both` est
  refusé plutôt qu'approché.

### 9.3 RouterOS

RouterOS n'a pas d'IFB. Le script descendant est une **file sur l'interface
LAN/privée**, là où le trafic descendant sort vers les clients. Ce n'est pas
un équivalent octet pour octet de l'IFB Linux : le point de passage diffère.
Les objets portent `cgoingress-*` et `comment="cgo-managed-ingress"`, disjoints
du nettoyage montant (`^cgo-`, `comment="cgo-managed"`).

### 9.4 Preuve

Les tests Windows prouvent la génération des commandes, les champs du rapport
et l'ordre fermé-échec du préflight. Ils ne prouvent pas qu'un noyau 6.8
accepte les commandes ni que l'IFB jette réellement sous charge. Cette preuve
est la tâche d'acceptation Linux, distincte et bloquée.
```

- [ ] **Step 2: Run the focused package gates**

Run:

```bash
go test ./pkg/remediate/... ./pkg/qdisc/... ./pkg/campagne/... ./pkg/audit/... ./cmd/cgo/... ./pkg/api/...
```

Expected: PASS.

- [ ] **Step 3: Run the full Windows gates**

Run: `go test ./...`

Expected: PASS.

Run: `go vet ./...`

Expected: no output. `web/frontend/dist` must exist; if it does not, build the frontend first or record the gate as “not runnable here” and do not claim it passed.

Run: `GOOS=linux GOARCH=amd64 go build ./...`

Expected: no output.

Run: `bash -n kit/cgo-bench-dualns.sh kit/run-dualns.sh`

Expected: no output.

- [ ] **Step 4: Run the frozen guard**

Run: `python book/tools/audit-archive.py`

Expected: the archive reports `159 runs`, `528` rows, `208` valid lines, `74` degraded, `246` invalid. The two numbers that matter for the guard are **159 runs / 208 valid lines**. Never say “208 runs”.

- [ ] **Step 5: Record the status**

In `docs/superpowers/plans/README.md`, change slice 6's status from **Proposed** to **Windows-green** and its next action to:

```
Live acceptance blocked on a stable Linux 6.8 bench; Task 9 is not executed.
```

Do not mark it **Done** or **Verified**.

- [ ] **Step 6: Inspect the diff and commit the docs**

Run: `git status --short`

Expected: only `docs/methodology.fr.md` and `docs/superpowers/plans/README.md` are modified for this task; code changes were committed in Tasks 1–7.

```bash
git add docs/methodology.fr.md docs/superpowers/plans/README.md
git commit -m "docs: IFB ingress methodology and Windows-green status"
```

---

### Task 9: Linux 6.8 live acceptance (blocked)

**Do not execute this task until a stable Linux 6.8 bench is available.** It is the only task that can prove the kernel accepts the IFB plumbing and that the AQM drops under load. Windows cannot run it. Leave it unchecked and record the blocker; do not mark the slice verified.

**Files:**
- Modify (only after a run): `docs/superpowers/plans/2026-10-27-cgo-ifb-ingress.md` (evidence block)
- Modify (only after a run): `docs/superpowers/plans/README.md` (status)

- [ ] **Step 1: Bring up the ingress bench**

On the Linux 6.8 host:

```bash
cd ~/cgo
./kit/run-dualns.sh down
./kit/run-dualns.sh up --ingress on
./kit/cgo-bench-dualns.sh check --ingress on
./kit/cgo-bench-dualns.sh show --ingress on
```

Expected: `check` prints `[dualns] ok`; `ifb0` exists and is UP in `ns_client`; `c-cli` has an `ingress` qdisc and a `mirred` redirect to `ifb0`; `ifb0` has `cake … ingress`; `e-wan` has no CAKE; `e-srv` still has upload CAKE.

- [ ] **Step 2: Run one ingress campaign cell**

```bash
cd ~/cgo
sudo ip netns exec ns_client env CGO_SHAPE_MODE=ingress ~/cgo/cgo-linux run \
  --bench dualns --direction down --profiles P2 --qdiscs cake \
  --cc bbr --reps 1 --flows 4 --target 10.30.0.2
```

Expected: the cell completes; the evidence row is appended under `data/runs/`; `FailedGates` does not contain a shape/preflight failure. This is a new run, not a frozen corpus row. The campaign runs inside `ns_client` because its ping/small/bulk probes are raw sockets: from the host, `10.30.0.2` is not routable.

- [ ] **Step 3: Inspect the kernel state while the cell is under load**

Run these in a second shell during Step 2's charge window. The campaign's
deferred cleanup removes the `ifb0` root qdisc when the cell returns, so a
post-run sample shows no CAKE by design, not a failure:

```bash
for i in 1 2 3; do
  sudo ip netns exec ns_client tc -s qdisc show dev ifb0
  sleep 5
done
sudo ip netns exec ns_client tc filter show dev c-cli parent ffff:
sudo ip netns exec ns_edge_router tc qdisc show dev e-wan
```

Expected: at least one `ifb0` sample shows `cake … ingress` and non-zero drops/bytes; the filter shows `mirred` to `ifb0`; `e-wan` has no CAKE. Record the exact output, not a paraphrase. After the cell, `ifb0` may show `noqueue`; that is the campaign's cleanup, not a failure.

- [ ] **Step 4: Check the report**

The campaign left `ifb0` unshaped. Restore the bench before measuring, or the
audit sizes the remediation from an unshaped veth rate:

```bash
cd ~/cgo
./kit/run-dualns.sh up --ingress on
sudo ip netns exec ns_client ~/cgo/cgo-linux audit --direction down --iface c-cli \
  --ros-lan ether2-lan --duration 30
```

Expected: a down report whose JSON has `linux_tc_ingress` sized from the measured down rate and whose Markdown has `### Linux — download (ingress IFB)` and `### MikroTik RouterOS v7+ — download (LAN queue tree)`. `linux_tc` is `[]`, not `null`. The audit runs inside `ns_client` so it reaches the bench sink at `10.30.0.2`; `--iface c-cli` is the measured WAN device.

- [ ] **Step 5: Clean up**

```bash
./kit/run-dualns.sh down
ip netns list
```

Expected: no `ns_client`, `ns_wan_emulator`, `ns_edge_router`, or `ns_server`; no leaked `ifb0`. `ifb0` disappears with its namespace.

- [ ] **Step 6: Record the evidence or the failure**

Paste the exact commands and outputs into this task only after they run. If any step fails, record the failure and leave the slice **Windows-green**, never **Verified**. Then update the roadmap status only if the evidence justifies it.

---

## Self-Review

- **Spec coverage:** D1 scope (Tasks 1–8, no UI/BBRv3/eBPF/.dist); D2 location (Tasks 3–4, 7); D3 direction (Task 3); D4 bench (Task 3); D5 opt-in (Tasks 3–4, 7); D6 ownership/order (Tasks 2–3); D7 no double shaping (Task 7); D8 CAKE parameters (Tasks 1, 3); D9 report direction (Task 5); D10 RouterOS LAN (Tasks 1, 5); D11 managed names (Task 1); D12 accounting (Task 8); D13 honesty (Task 9); D14 frozen data (Task 8); D15 no UI (Task 6).
- **Type consistency:** `ShapeMode`/`IngressPreflight` are defined in Task 3 and consumed in Task 4. `LinuxTCIngress`/`RouterOSIngress` are defined in Task 1 and consumed in Task 5. `CheckIngress` is defined in Task 2 and consumed in Task 4. `RouterOSLanInterface` is defined in Task 1, consumed in Tasks 5–6. `IngressIFB` is used only in Task 1.
- **No placeholders:** every code step contains the code; every test step contains the test; every run step has the exact command and expected result.
- **Known shortcut:** the bench terminates the download in the same namespace as the probes and does not model a LAN behind the client. The methodology section states it; do not present the bench as a full CPE.
