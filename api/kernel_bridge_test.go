package api

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/probe/bpf"
)

func TestKernelTelemetryFromSummary(t *testing.T) {
	s := bpf.Summary{
		Available: true, Source: "ebpf", Attach: "fentry", ObjectSHA256: "h",
		FlowsSeen: 1, FlowsTotal: 4, RetransPayload: 1448, QdiscDrops: 2,
		Flows: []bpf.Flow{{Local: "10.0.0.1:1", Peer: "10.0.0.2:2", RetransPayload: 1448, SRTTus: 25000}},
	}
	tel := kernelTelemetryFromSummary(s)
	if tel == nil || tel.FlowsSeen != 1 || tel.FlowsTotal != 4 {
		t.Fatalf("compteurs: %+v", tel)
	}
	if len(tel.Flows) != 1 || tel.Flows[0].RetransPayloadBytes != 1448 || tel.Flows[0].SRTTus != 25000 {
		t.Fatalf("flux mal mappé: %+v", tel.Flows)
	}
	if kernelTelemetryFromSummary(bpf.Summary{}) != nil {
		t.Fatal("résumé vide ne doit pas produire de bloc")
	}
}

// Oracle positif/hostile — le compteur hors-ring remonte et force la note de
// borne inférieure ; zéro reste zéro et disparaît du JSON (octets identiques).
func TestKernelTelemetryFromSummaryCarriesReserveFailures(t *testing.T) {
	tel := kernelTelemetryFromSummary(bpf.Summary{Source: "ebpf", ReserveFailures: 3})
	if tel == nil {
		t.Fatal("résumé ebpf doit produire un bloc")
	}
	if tel.ReserveFailures != 3 || !tel.LowerBound {
		t.Fatalf("perte non propagée: %+v", tel)
	}

	zero := kernelTelemetryFromSummary(bpf.Summary{Source: "ebpf"})
	if zero == nil || zero.ReserveFailures != 0 || zero.LowerBound {
		t.Fatalf("zéro doit rester zéro sans note: %+v", zero)
	}
	b, _ := json.Marshal(zero)
	if strings.Contains(string(b), "reserve_failures") || strings.Contains(string(b), "lower_bound") {
		t.Fatalf("champs zéro sérialisés: %s", b)
	}
}

// Oracle du budget d'observateur : le dénominateur vient des compteurs réels
// de la sonde (hors ring saturé), jamais d'un nombre de circumference inventé.
func TestKernelTelemetryFromSummaryCarriesObserverLoss(t *testing.T) {
	tel := kernelTelemetryFromSummary(bpf.Summary{
		Source: "ebpf", ReserveFailures: 3, RetransEvents: 900, QdiscDrops: 100,
	})
	if tel == nil {
		t.Fatal("résumé ebpf doit produire un bloc")
	}
	if tel.ObserverEvents != 1000 || math.Abs(tel.ObserverLossPct-0.3) > 1e-9 {
		t.Fatalf("budget observateur = %d / %v %%, voulu 1000 / 0.3 %% : %+v",
			tel.ObserverEvents, tel.ObserverLossPct, tel)
	}
	if !strings.Contains(tel.ObserverLossDenominator, "retrans_events") {
		t.Fatalf("dénominateur non nommé : %q", tel.ObserverLossDenominator)
	}

	// Zéro réservation manquée : 0 % mais dénominateur toujours déclaré.
	zero := kernelTelemetryFromSummary(bpf.Summary{Source: "ebpf", RetransEvents: 10})
	if zero == nil || zero.ObserverEvents != 10 || zero.ObserverLossPct != 0 || zero.ObserverLossDenominator == "" {
		t.Fatalf("zéro perte doit garder le dénominateur : %+v", zero)
	}

	// Aucun événement : on n'invente pas de compteur.
	none := kernelTelemetryFromSummary(bpf.Summary{Source: "ebpf"})
	if none == nil || none.ObserverEvents != 0 || none.ObserverLossPct != 0 || none.ObserverLossDenominator == "" {
		t.Fatalf("aucun événement : %+v", none)
	}
	b, _ := json.Marshal(none)
	if strings.Contains(string(b), "observer_events") || strings.Contains(string(b), "observer_loss_pct") {
		t.Fatalf("champs zéro sérialisés : %s", b)
	}
}
