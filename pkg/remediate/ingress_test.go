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
