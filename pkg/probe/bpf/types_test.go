package bpf

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSummaryJSONCarriesSourceAndFlows(t *testing.T) {
	s := Summary{Available: true, Source: "ebpf", Attach: "fentry", FlowsSeen: 4, FlowsTotal: 4,
		RetransPayload: 1024, QdiscDrops: 7, QdiscAvailable: true,
		Flows: []Flow{{Local: "10.0.0.1:4000", Peer: "10.0.0.2:443", RetransPayload: 512, SRTTus: 25000}}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"source":"ebpf"`, `"retrans_payload_bytes":1024`, `"qdisc_drop_packets":7`, `"srtt_us":25000`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("summary JSON lacks %s: %s", want, b)
		}
	}
}

func TestErrorsAreTypedAndReadable(t *testing.T) {
	var err error = &UnavailableError{Reason: "CAP_BPF absent"}
	if !strings.Contains(err.Error(), "CAP_BPF") {
		t.Fatalf("unavailable error lost the reason: %v", err)
	}
	err = &BuildError{Reason: "objet non embarqué"}
	if !strings.Contains(err.Error(), "objet non embarqué") {
		t.Fatalf("build error lost the reason: %v", err)
	}
}
