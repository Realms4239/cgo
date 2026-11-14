package qdisc

import (
	"strings"
	"testing"
)

func TestApplyNetemDistributionArgs(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 0.25, 0.5, nil, "P2"); err != nil {
		t.Fatal(err)
	}
	want := []string{"qdisc", "replace", "dev", "w-edge", "root", "handle", "1:", "netem",
		"delay", "50ms", "0.25ms", "loss", "0.5%", "distribution", "P2"}
	if strings.Join(f.Calls[0], " ") != strings.Join(want, " ") {
		t.Fatalf("args = %v, want %v", f.Calls[0], want)
	}
}

func TestApplyNetemDistributionRequiresNameAndJitter(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 1, 0, nil, ""); err == nil {
		t.Fatal("nom vide accepté")
	}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 0, 0, nil, "P2"); err == nil {
		t.Fatal("jitter nul accepté")
	}
	if len(f.Calls) != 0 {
		t.Fatalf("tc appelé malgré un contrat invalide : %v", f.Calls)
	}
}

func TestApplyNetemDistributionWithGE(t *testing.T) {
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001}
	if err := ApplyNetemDistribution(f, "c-wan", 300, 2.5, 1, &ge, "P3"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Calls[0], " ")
	if !strings.Contains(joined, "loss gemodel 0.2% 10% 50% 0.1%") || !strings.Contains(joined, "distribution P3") {
		t.Fatalf("gemodel + distribution attendus : %v", f.Calls[0])
	}
}

func TestApplyNetemDistributionRejectsNonBareName(t *testing.T) {
	// Défense en profondeur : tc charge la .dist DANS TC_LIB_DIR — un nom
	// de chemin ou de traversée serait chargé en dehors, ou refusé en silence.
	for _, dist := range []string{"../P2", "sub/P2", `sub\P2`, ".."} {
		f := &FakeRunner{}
		if err := ApplyNetemDistribution(f, "w-edge", 50, 1, 0, nil, dist); err == nil {
			t.Fatalf("nom %q accepté", dist)
		}
		if len(f.Calls) != 0 {
			t.Fatalf("tc appelé pour un nom non nu %q : %v", dist, f.Calls)
		}
	}
}

func TestNsTCArgsCarriesEnv(t *testing.T) {
	got := nsTCArgs("ns_wan_emulator", []string{"TC_LIB_DIR=/tmp/cal"}, []string{"qdisc", "show"})
	want := []string{"-n", "/usr/local/bin/cgo-netctl", "--env", "TC_LIB_DIR=/tmp/cal", "ip", "netns", "exec", "ns_wan_emulator", "--", "tc", "qdisc", "show"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("args = %v, want %v", got, want)
	}
	if plain := nsTCArgs("ns", nil, []string{"qdisc"}); strings.Join(plain, " ") != "-n /usr/local/bin/cgo-netctl ip netns exec ns -- tc qdisc" {
		t.Fatalf("sans env = %v", plain)
	}
}
