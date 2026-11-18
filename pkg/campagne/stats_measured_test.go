package campagne

import (
	"context"
	"testing"

	"github.com/Realms4239/cgo/model"
	"github.com/Realms4239/cgo/qdisc"
)

func TestStatsUnmeasuredForcesInvalidGate(t *testing.T) {
	_, dropsOK := qdisc.SumDropsMeasured(nil)
	_, bytesOK := qdisc.SumBytesMeasured(nil)
	if !statsUnmeasured(dropsOK, bytesOK) {
		t.Fatal("nil snapshot must count as unmeasured")
	}
	_, dropsOK = qdisc.SumDropsMeasured([]qdisc.Stats{{Drops: 1}})
	_, bytesOK = qdisc.SumBytesMeasured([]qdisc.Stats{{Bytes: 1}})
	if statsUnmeasured(dropsOK, bytesOK) {
		t.Fatal("populated snapshot must count as measured")
	}
}

func TestAppendStatsFailureGateIsJSONOnly(t *testing.T) {
	var failed []string
	failed = appendStatsFailureGate(failed)
	if len(failed) != 1 || failed[0] != "G8StatsMeasured" {
		t.Fatalf("failed gates = %v, want [G8StatsMeasured]", failed)
	}
}

// TestRunEventUnmeasuredScrapeForcesInvalidGate prouve le câblage, pas seulement
// l'existence des helpers : une scrape tc vide à une extrémité de la charge ne
// doit jamais devenir un zéro mesuré (failed_gates JSON + gate_status invalid).
func TestRunEventUnmeasuredScrapeForcesInvalidGate(t *testing.T) {
	run := func(t *testing.T, statsFn func() []qdisc.Stats) model.Event {
		t.Helper()
		d := fastDeps()
		d.StatsFn = statsFn
		ev := model.Event{RunID: "r1", EventID: 1, Profile: "P2", Qdisc: model.FqCodel, CC: model.Cubic, Repetition: 1}
		got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}

	t.Run("measured scrape stays valid", func(t *testing.T) {
		got := run(t, func() []qdisc.Stats { return []qdisc.Stats{{Bytes: 1000, Drops: 0}} })
		if got.GateStatus != model.GatePass {
			t.Fatalf("status=%s want valid, failed=%v", got.GateStatus, got.FailedGates)
		}
		for _, f := range got.FailedGates {
			if f == "G8StatsMeasured" {
				t.Fatal("measured scrape must not raise G8StatsMeasured")
			}
		}
	})

	t.Run("empty scrape is invalid", func(t *testing.T) {
		got := run(t, func() []qdisc.Stats { return nil })
		if got.GateStatus != model.GateInvalid {
			t.Fatalf("status=%s want invalid", got.GateStatus)
		}
		var found bool
		for _, f := range got.FailedGates {
			if f == "G8StatsMeasured" {
				found = true
			}
		}
		if !found {
			t.Fatalf("G8StatsMeasured absent from failed_gates: %v", got.FailedGates)
		}
	})
}
