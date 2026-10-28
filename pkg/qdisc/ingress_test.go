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
