package api

import (
	"testing"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
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
