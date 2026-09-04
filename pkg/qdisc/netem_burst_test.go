package qdisc

import (
	"strings"
	"testing"
)

// TestApplyNetemBurstLoss — Gilbert-Elliott (perte en rafales, la vraie vie
// 4G/VSAT) : netem "loss gemodel p r h 1-r k" quand le profil porte
// loss_burst. Sans champs : perte uniforme, comportement historique.
func TestApplyNetemBurstLoss(t *testing.T) {
	f := &FakeRunner{}
	// (iface, delay, jitter, loss%, p, r, h, k) — gemodel p>0 active le mode rafales
	if err := ApplyNetemBurst(f, "veth-c", 100, 15, 1, 0.01, 0.001, 200, 0.05); err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, c := range f.Calls {
		joined += strings.Join(c, " ") + "\n"
	}
	if !strings.Contains(joined, "loss gemodel 0.01 0.001 200 0.999 0.05") {
		t.Fatalf("gemodel args manquants: %s", joined)
	}

	f2 := &FakeRunner{}
	if err := ApplyNetemBurst(f2, "veth-c", 100, 15, 0.5, 0, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	joined2 := ""
	for _, c := range f2.Calls {
		joined2 += strings.Join(c, " ") + "\n"
	}
	if !strings.Contains(joined2, "loss 0.5%") || strings.Contains(joined2, "gemodel") {
		t.Fatalf("sans burst : perte uniforme attendue: %s", joined2)
	}
}
