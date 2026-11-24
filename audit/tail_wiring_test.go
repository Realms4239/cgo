package audit

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
	"github.com/Realms4239/cgo/remediate"
)

// Oracle hostile : un percentile porté par l'échantillon mais soutenu par moins
// de 1000 mesures ne traverse pas la frontière du rapport. Aucun tail fabriqué.
func TestKernelTelemetryRefusesFabricatedTailBelowFloor(t *testing.T) {
	for _, n := range []uint64{0, 999} {
		tel := KernelTelemetryFrom(KernelSample{
			Source: "ebpf", LatencySamples: n,
			LatencyP50us: 64, LatencyP95us: 256, LatencyP999us: 1024,
			TailBinBounds: "bucket i = [2^i, 2^(i+1)) us",
		})
		if tel == nil {
			t.Fatalf("samples=%d: bloc absent", n)
		}
		if tel.LatencyP50us != 0 || tel.LatencyP95us != 0 || tel.LatencyP999us != 0 || tel.TailBinBounds != "" {
			t.Fatalf("samples=%d: percentile fabriqué: %+v", n, tel)
		}
		b, _ := json.Marshal(tel)
		for _, ghost := range []string{"latency_p50_us", "latency_p95_us", "latency_p999_us", "tail_bin_bounds"} {
			if strings.Contains(string(b), ghost) {
				t.Fatalf("samples=%d: %s sérialisé: %s", n, ghost, b)
			}
		}
	}
}

// Oracle positif : au plancher, les valeurs mesurées traversent jusqu'au JSON.
func TestKernelTelemetryCarriesTailAtFloor(t *testing.T) {
	tel := KernelTelemetryFrom(KernelSample{
		Source: "ebpf", LatencySamples: 1000,
		LatencyP50us: 64, LatencyP95us: 256, LatencyP999us: 1024,
		TailBinBounds:   "bucket i = [2^i, 2^(i+1)) us",
		ReserveFailures: 3, MapFull: 2,
		RawSRTTus: []uint32{12000, 15000},
	})
	if tel.LatencySamples != 1000 || tel.LatencyP50us != 64 || tel.LatencyP95us != 256 || tel.LatencyP999us != 1024 {
		t.Fatalf("queues perdues: %+v", tel)
	}
	if tel.TailBinBounds == "" || tel.ReserveFailures != 3 || tel.MapFull != 2 {
		t.Fatalf("provenance perdue: %+v", tel)
	}
	if len(tel.LatencySamplesSRTT) != 2 || tel.LatencySamplesSRTT[1] != 15000 {
		t.Fatalf("srtt brut perdu: %+v", tel.LatencySamplesSRTT)
	}
	if !strings.Contains(tel.Reason, "latency_samples_srtt") {
		t.Fatalf("basis des échantillons srtt non déclaré: %q", tel.Reason)
	}
	b, _ := json.Marshal(tel)
	for _, want := range []string{
		`"latency_p999_us":1024`, `"reserve_failures":3`, `"map_full":2`,
		`"latency_samples_srtt":[12000,15000]`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s absent du JSON: %s", want, b)
		}
	}
}

// Oracle de borne : le tableau brut est tronqué à 512 et absent s'il est vide.
func TestKernelTelemetryRawSRTTBoundedAndAbsentWhenEmpty(t *testing.T) {
	big := make([]uint32, 600)
	for i := range big {
		big[i] = uint32(i + 1)
	}
	tel := KernelTelemetryFrom(KernelSample{Source: "ebpf", RawSRTTus: big})
	if len(tel.LatencySamplesSRTT) != 512 || tel.LatencySamplesSRTT[511] != 512 {
		t.Fatalf("srtt non borné: %d valeurs, dernière %d", len(tel.LatencySamplesSRTT), tel.LatencySamplesSRTT[511])
	}
	empty := KernelTelemetryFrom(KernelSample{Source: "ebpf"})
	b, _ := json.Marshal(empty)
	if strings.Contains(string(b), "latency_samples_srtt") {
		t.Fatalf("tableau vide sérialisé: %s", b)
	}
}

func TestBuildReportCarriesTailFromKernelSample(t *testing.T) {
	res := &Result{AuditID: "a-tail", KernelSample: &KernelSample{
		Source: "ebpf", LatencySamples: 1200,
		LatencyP50us: 64, LatencyP95us: 256, LatencyP999us: 1024,
		TailBinBounds: "bucket i = [2^i, 2^(i+1)) us",
	}}
	rep, err := BuildReport(res, remediate.Link{}, "", false, false)
	if err != nil || rep.KernelTelemetry == nil {
		t.Fatalf("rapport sans télémétrie: %v %+v", err, rep)
	}
	if rep.KernelTelemetry.LatencyP999us != 1024 {
		t.Fatalf("p99.9 perdu dans le rapport: %+v", rep.KernelTelemetry)
	}
	b, _ := json.Marshal(rep)
	if !strings.Contains(string(b), `"latency_p999_us":1024`) || !strings.Contains(string(b), `"tail_bin_bounds"`) {
		t.Fatalf("queues absentes du JSON du rapport: %s", b)
	}
}

func TestBuildReportRefusesTailBelowFloor(t *testing.T) {
	res := &Result{AuditID: "a-short", KernelSample: &KernelSample{
		Source: "ebpf", LatencySamples: 999,
		LatencyP50us: 64, LatencyP999us: 1024,
		TailBinBounds: "bucket i = [2^i, 2^(i+1)) us",
	}}
	rep, err := BuildReport(res, remediate.Link{}, "", false, false)
	if err != nil || rep.KernelTelemetry == nil {
		t.Fatalf("rapport sans télémétrie: %v %+v", err, rep)
	}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "latency_p999_us") || strings.Contains(string(b), "tail_bin_bounds") {
		t.Fatalf("tail fabriqué sous le plancher: %s", b)
	}
}

// Oracle EMD : le nombre n'est calculé QUE si les deux jeux existent ; aucun
// verdict n'est fabriqué et le plafond est déclaré, jamais élargi.
func TestBuildReportEMDOnlyWithBothSampleSets(t *testing.T) {
	newResult := func(emulated []float64, live []uint32) *Result {
		s := &KernelSample{Source: "ebpf"}
		s.RawSRTTus = live
		return &Result{AuditID: "a-emd", EmulatedRTTSamples: emulated, KernelSample: s}
	}

	rep, _ := BuildReport(newResult([]float64{20, 21, 22}, []uint32{20000, 21000, 22000}), remediate.Link{}, "", false, false)
	if rep.Fidelity == nil {
		t.Fatal("deux jeux présents : bloc de fidélité absent")
	}
	if math.Abs(rep.Fidelity.EMD) > 1e-9 || rep.Fidelity.LiveSamples != 3 || rep.Fidelity.EmulatedSamples != 3 {
		t.Fatalf("EMD identique devrait être 0: %+v", rep.Fidelity)
	}
	if rep.Fidelity.EMDCeiling != metrics.FidelityEMDMax || !strings.Contains(rep.Fidelity.Basis, "srtt_us") {
		t.Fatalf("plafond/basis non déclarés: %+v", rep.Fidelity)
	}

	shift, _ := BuildReport(newResult([]float64{40, 41, 42}, []uint32{20000, 21000, 22000}), remediate.Link{}, "", false, false)
	if shift.Fidelity.EMD <= metrics.FidelityEMDMax {
		t.Fatalf("EMD décalé = %v, doit dépasser le plafond %v", shift.Fidelity.EMD, metrics.FidelityEMDMax)
	}

	for _, tc := range []struct {
		name     string
		emulated []float64
		live     []uint32
	}{
		{"sans émulé", nil, []uint32{20000}},
		{"sans vivant", []float64{20}, nil},
		{"aucun", nil, nil},
	} {
		r, _ := BuildReport(newResult(tc.emulated, tc.live), remediate.Link{}, "", false, false)
		if r.Fidelity != nil {
			t.Fatalf("%s: EMD fabriqué: %+v", tc.name, r.Fidelity)
		}
		b, _ := json.Marshal(r)
		if strings.Contains(string(b), `"fidelity"`) {
			t.Fatalf("%s: bloc vide sérialisé: %s", tc.name, b)
		}
	}
}
