package qdisc

import (
	"reflect"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
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
		{"qdisc", "replace", "dev", "veth-s", "root", "cake", "bandwidth", "20mbit", "rtt", "100ms"},
		{"qdisc", "replace", "dev", "veth-s", "root", "tbf", "rate", "20mbit", "burst", "256kbit", "latency", "400ms"},
		{"qdisc", "add", "dev", "veth-s", "parent", "1:1", "handle", "2:", "fq_codel"},
		{"qdisc", "replace", "dev", "veth-s", "root", "tbf", "rate", "80mbit", "burst", "256kbit", "latency", "400ms"},
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
	if len(got) != 2 || got[0][9] != "loss" || len(got[1]) != 9 {
		t.Fatalf("unexpected vectors: %v", got)
	}
}
