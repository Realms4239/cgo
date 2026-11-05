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
