package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReportTelemetryFallbackIsStated(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.KernelTelemetry = &KernelTelemetry{Source: "tcpinfo", Reason: "objet non embarqué"}
	b, _ := json.Marshal(rep)
	if !strings.Contains(string(b), `"kernel_telemetry"`) {
		t.Fatal("kernel_telemetry missing from JSON")
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Repli TCP_INFO") || !strings.Contains(md, "objet non embarqué") {
		t.Fatalf("fallback not stated with its reason:\n%s", md)
	}
}

func TestReportTelemetryEBPFSaysAttach(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.KernelTelemetry = &KernelTelemetry{Available: true, Source: "ebpf", RetransPayloadBytes: 4096, QdiscDropPackets: 12}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Télémétrie noyau : eBPF") {
		t.Fatalf("ebpf source not rendered:\n%s", md)
	}
	if !strings.Contains(md, "4096") {
		t.Fatal("retransmitted bytes not rendered")
	}
}

func TestReportWithoutTelemetryOmitsBoth(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "kernel_telemetry") {
		t.Fatal("absent telemetry still serialized")
	}
	if strings.Contains(rep.RenderMarkdown(), "Télémétrie noyau") {
		t.Fatal("absent telemetry still rendered")
	}
}
