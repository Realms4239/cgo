package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/probe/bpf"
)

func TestKernelTelemetrySummaryCarriesTailAndRawSRTT(t *testing.T) {
	var h bpf.Log2Histogram
	h[6] = 1200 // bucket 64 us, 1200 échantillons -> plancher atteint
	s := bpf.Summary{
		Source: "ebpf", LatencySamples: 1200, LatencyHistogram: h,
		ReserveFailures: 3, MapFull: 2,
		Flows: []bpf.Flow{{SRTTus: 12000}, {SRTTus: 15000}},
	}
	tel := kernelTelemetryFromSummary(s)
	if tel == nil {
		t.Fatal("résumé ebpf doit produire un bloc")
	}
	if tel.LatencySamples != 1200 || tel.LatencyP50us != 64 || tel.LatencyP95us != 64 || tel.LatencyP999us != 64 {
		t.Fatalf("queues non calculées: %+v", tel)
	}
	if tel.TailBinBounds == "" || tel.ReserveFailures != 3 || tel.MapFull != 2 {
		t.Fatalf("provenance perdue: %+v", tel)
	}
	if len(tel.LatencySamplesSRTT) != 2 || tel.LatencySamplesSRTT[1] != 15000 {
		t.Fatalf("srtt brut perdu: %+v", tel.LatencySamplesSRTT)
	}
	if !strings.Contains(tel.Reason, "latency_samples_srtt") {
		t.Fatalf("basis des srtt non marqué: %q", tel.Reason)
	}
	b, _ := json.Marshal(tel)
	if !strings.Contains(string(b), `"latency_samples_srtt":[12000,15000]`) {
		t.Fatalf("srtt absent du JSON: %s", b)
	}
}

func TestKernelTelemetrySummaryBelowFloorNoTail(t *testing.T) {
	var h bpf.Log2Histogram
	h[6] = 999
	tel := kernelTelemetryFromSummary(bpf.Summary{Source: "ebpf", LatencySamples: 999, LatencyHistogram: h})
	if tel == nil {
		t.Fatal("résumé ebpf doit produire un bloc")
	}
	if tel.LatencyP50us != 0 || tel.LatencyP95us != 0 || tel.LatencyP999us != 0 || tel.TailBinBounds != "" {
		t.Fatalf("tail fabriqué sous le plancher: %+v", tel)
	}
	b, _ := json.Marshal(tel)
	for _, ghost := range []string{"latency_p50_us", "latency_p999_us", "tail_bin_bounds"} {
		if strings.Contains(string(b), ghost) {
			t.Fatalf("%s sérialisé sans plancher: %s", ghost, b)
		}
	}
}
