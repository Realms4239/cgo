package qdisc

import (
	"reflect"
	"testing"

	"github.com/Realms4239/cgo/model"
)

func TestApplyShaperArgVectors(t *testing.T) {
	r := &FakeRunner{}
	if err := ApplyShaper(r, "veth-s", model.Cake, 20, 100); err != nil {
		t.Fatal(err)
	}
	if err := ApplyShaper(r, "veth-s", model.FqCodel, 20, 0); err != nil {
		t.Fatal(err)
	}
	if err := ApplyShaper(r, "veth-s", model.PfifoFast, 80, 0); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "cake", "bandwidth", "20mbit", "rtt", "100ms"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "tbf", "rate", "20mbit", "burst", "256kbit", "latency", "400ms"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "10:1", "handle", "20:", "fq_codel", "target", "5ms", "interval", "100ms", "limit", "10240", "quantum", "1514", "flows", "1024"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "tbf", "rate", "80mbit", "burst", "256kbit", "latency", "400ms"},
	}
	if !reflect.DeepEqual(r.Calls, want) {
		t.Fatalf("got %v\nwant %v", r.Calls, want)
	}
}

func TestApplyNetemLossOmittedWhenZero(t *testing.T) {
	r := &FakeRunner{}
	if err := ApplyNetem(r, "veth-c", 100, 15, 0.5); err != nil {
		t.Fatal(err)
	}
	if err := ApplyNetem(r, "veth-c", 20, 2, 0); err != nil {
		t.Fatal(err)
	}
	got := r.Calls
	if len(got) != 2 || got[0][11] != "loss" || len(got[1]) != 11 {
		t.Fatalf("unexpected vectors: %v", got)
	}
}

func TestFQCoDel_PinsRFC8290Params(t *testing.T) {
	s := fqCoDelString("eth0")
	for _, want := range []string{"target 5ms", "interval 100ms", "limit 10240", "quantum 1514", "flows 1024"} {
		if !contains(s, want) {
			t.Fatalf("missing %q in %q", want, s)
		}
	}
}
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
