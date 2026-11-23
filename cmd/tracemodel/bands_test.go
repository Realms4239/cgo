package main

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/audit"
)

func testFloors() map[string]float64 {
	return map[string]float64{"P1": 20, "P2": 100, "P3": 600, "P4": 40}
}

func assertSamples(t *testing.T, got []float64, want ...float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d samples %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("sample %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestBucketFlowsProfileBands(t *testing.T) {
	floors := testFloors()

	// srtt_us=839657 → 839.657 ms : P3 seul, jamais P1/P2/P4.
	high := BucketFlows([]FlowRow{{SRTTus: 839657}}, floors)
	assertSamples(t, high["P3"], 839.657)
	assertSamples(t, high["P1"])
	assertSamples(t, high["P2"])
	assertSamples(t, high["P4"])

	// srtt_us=150000 → 150.000 ms : P2 seul.
	mid := BucketFlows([]FlowRow{{SRTTus: 150000}}, floors)
	assertSamples(t, mid["P2"], 150)
	assertSamples(t, mid["P1"])
	assertSamples(t, mid["P3"])
	assertSamples(t, mid["P4"])
}

func TestBucketFlowsSkipsInvalidSRTT(t *testing.T) {
	// Oracle hostile : un plancher synthétique à 0 ms piégerait toute valeur
	// invalide coercée en zéro — la bande doit rester vide, pas mesurée.
	floors := map[string]float64{"P1": 20, "Z": 0}
	rows := []FlowRow{{SRTTus: math.NaN()}, {SRTTus: -1}, {SRTTus: 0}}
	got := BucketFlows(rows, floors)
	for _, p := range []string{"P1", "Z"} {
		assertSamples(t, got[p])
	}

	// Une mesure valide au milieu de lignes invalides survit intacte.
	got = BucketFlows([]FlowRow{{SRTTus: math.NaN()}, {SRTTus: 150000}, {SRTTus: 0}}, testFloors())
	assertSamples(t, got["P2"], 150)
}

func TestValidateBandThreshold(t *testing.T) {
	few := make([]float64, 29)
	err := ValidateBand(few, 30, "P3")
	if err == nil {
		t.Fatal("29 samples acceptés : la garde -30 est contournable")
	}
	for _, want := range []string{"P3", "29", "30"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("erreur %q ne nomme pas %q", err, want)
		}
	}
	if err := ValidateBand(make([]float64, 30), 30, "P3"); err != nil {
		t.Fatalf("30 samples refusés : %v", err)
	}
	// La signature à deux arguments reste appelable.
	if err := ValidateBand(few, 30); err == nil {
		t.Fatal("appel à deux arguments a accepté 29 samples")
	}
}

func TestWriteTelemetryFormat(t *testing.T) {
	dir := t.TempDir()
	samples := []float64{839.657, 839.657, 150}
	sum, err := WriteTelemetry(dir, "trace-x", "P3", samples)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "audits", "trace-x-P3", "telemetry.csv")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "phase,offset_ms,rtt_ms,small_ms,bulk_bytes\n" +
		"baseline,0,839.657,,\n" +
		"baseline,5000,839.657,,\n" +
		"baseline,10000,150.000,,\n"
	if string(raw) != want {
		t.Fatalf("telemetry.csv:\n%s\nattendu:\n%s", raw, want)
	}
	digest := sha256.Sum256(raw)
	if sum != hex.EncodeToString(digest[:]) {
		t.Fatalf("sha256 rendu %s, attendu %s", sum, hex.EncodeToString(digest[:]))
	}

	// Oracle indépendant : le vrai lecteur de `cgo calibrate` doit lire ce fichier.
	rows, err := audit.ReadTelemetry(dir, "trace-x-P3")
	if err != nil {
		t.Fatalf("consommateur calibrate incapable de lire la télémétrie : %v", err)
	}
	assertSamples(t, []float64{rows[0].RTTms, rows[1].RTTms, rows[2].RTTms}, 839.657, 839.657, 150)
	for i, r := range rows {
		if r.Phase != "baseline" || r.OffsetMs != int64(i*5000) || r.SmallMs != 0 || r.BulkBytes != 0 {
			t.Fatalf("ligne %d inattendue : %+v", i, r)
		}
	}
}

func TestWriteTelemetryRefusesPathEscape(t *testing.T) {
	if _, err := WriteTelemetry(t.TempDir(), "../hors-zone", "P3", []float64{839.657}); err == nil {
		t.Fatal("identifiant traversant accepté : écriture hors zone d'audits")
	}
}
