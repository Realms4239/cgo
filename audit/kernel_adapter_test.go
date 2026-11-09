package audit

import "testing"

func TestKernelTelemetryFromMapsFlows(t *testing.T) {
	tel := KernelTelemetryFrom(KernelSample{
		Source: "ebpf", Attach: "fentry", ObjectSHA256: "abc",
		FlowsSeen: 2, FlowsTotal: 4, RetransPayload: 900, QdiscDrops: 3,
		Flows: []KernelFlow{{Local: "10.0.0.1:4000", Peer: "10.0.0.2:443", RetransPayloadBytes: 512}},
	})
	if tel == nil || !tel.Available || tel.Source != "ebpf" || len(tel.Flows) != 1 {
		t.Fatalf("mapping incomplet: %+v", tel)
	}
	if tel.ObjectSHA256 != "abc" || tel.RetransPayloadBytes != 900 || tel.FlowsTotal != 4 {
		t.Fatalf("champs perdus: %+v", tel)
	}
}

func TestKernelTelemetryFromEmptyIsAbsent(t *testing.T) {
	if KernelTelemetryFrom(KernelSample{}) != nil {
		t.Fatal("un échantillon vide ne doit pas produire de bloc")
	}
}
