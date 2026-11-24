package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
)

const (
	fieldHeader = "phase,offset_ms,rtt_ms,small_ms,bulk_bytes\n"
	benchHeader = "kind,window_unix,local_label,peer_label,local_port,peer_port,retrans_payload_bytes,retrans_events,qdisc_drop_packets,qdisc_drop_bytes,srtt_us,snd_cwnd,snd_ssthresh,state,samples\n"
)

func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

func fieldBody(n int) string {
	var b strings.Builder
	b.WriteString(fieldHeader)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "baseline,%d,%.3f,,\n", i*100, 600+float64(i%37))
	}
	return b.String()
}

func benchBody(n int, base int64) string {
	var b strings.Builder
	b.WriteString(benchHeader)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "F,1763000000,local,peer,43001,443,0,0,0,0,%d,10,20,open,1\n", base+int64(i)*13)
	}
	return b.String()
}

func benchRow(kind, srtt string) string {
	return fmt.Sprintf("%s,1763000000,local,peer,43001,443,0,0,0,0,%s,10,20,open,1\n", kind, srtt)
}

// pairFixtures exprime la même valeur de délai des deux côtés : ms côté terrain,
// µs côté banc. L'EMD doit être exactement zéro, sans arrangement.
func pairFixtures(t *testing.T, values []float64) (string, string) {
	t.Helper()
	var f, b strings.Builder
	f.WriteString(fieldHeader)
	b.WriteString(benchHeader)
	for i, v := range values {
		fmt.Fprintf(&f, "baseline,%d,%.3f,,\n", i*100, v)
		fmt.Fprintf(&b, "F,1763000000,local,peer,43001,443,0,0,0,0,%d,10,20,open,1\n", int64(math.Round(v*1000)))
	}
	return writeTemp(t, "field.csv", f.String()), writeTemp(t, "bench.csv", b.String())
}

// pairFixturesShifted garde le même jeu mais retarde le banc de shiftMs :
// l'EMD doit dépasser le plafond déclaré et le verdict doit refuser.
func pairFixturesShifted(t *testing.T, values []float64, shiftMs float64) (string, string) {
	t.Helper()
	var f, b strings.Builder
	f.WriteString(fieldHeader)
	b.WriteString(benchHeader)
	for i, v := range values {
		fmt.Fprintf(&f, "baseline,%d,%.3f,,\n", i*100, v)
		fmt.Fprintf(&b, "F,1763000000,local,peer,43001,443,0,0,0,0,%d,10,20,open,1\n", int64(math.Round((v+shiftMs)*1000)))
	}
	return writeTemp(t, "field.csv", f.String()), writeTemp(t, "bench.csv", b.String())
}

func assertSHA(t *testing.T, path, got string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	want := hex.EncodeToString(sum[:])
	if got != want {
		t.Fatalf("sha256(%s) = %q, want %q", path, got, want)
	}
}

func TestParseOptionsRequired(t *testing.T) {
	if _, err := parseOptions([]string{"--field", "f.csv", "--bench", "b.csv"}); err == nil {
		t.Fatal("--out manquant : parseOptions doit refuser")
	}
	o, err := parseOptions([]string{"--field", "f.csv", "--bench", "b.csv", "--out", "r.json"})
	if err != nil {
		t.Fatalf("parseOptions: %v", err)
	}
	if o.minUs != 600000 {
		t.Fatalf("min-us par défaut = %d, want 600000", o.minUs)
	}
	o, err = parseOptions([]string{"--field", "f.csv", "--bench", "b.csv", "--out", "r.json", "--min-us", "610000"})
	if err != nil || o.minUs != 610000 {
		t.Fatalf("min-us explicite = %d, err %v", o.minUs, err)
	}
}

func TestReadFieldRealFormat(t *testing.T) {
	// 113 lignes terrain scellées, plus des lignes vides/non numériques à ignorer.
	body := fieldBody(113) + "\nbaseline,114,not-a-number,,\n\nbaseline,115,,,not-a-number\n"
	p := writeTemp(t, "telemetry.csv", body)
	got, err := readField(p)
	if err != nil {
		t.Fatalf("readField: %v", err)
	}
	if len(got) != 113 {
		t.Fatalf("lignes terrain = %d, want 113", len(got))
	}
	if got[0] != 600 || got[112] != 601 {
		t.Fatalf("échantillons terrain altérés : premier %v, dernier %v", got[0], got[112])
	}
}

func TestReadBenchRealFormat(t *testing.T) {
	// 41 lignes F dans la bande, plus une ligne S, une sous la bande et une non numérique.
	body := benchBody(41, 600000) +
		benchRow("S", "999999") +
		benchRow("F", "599999") +
		benchRow("F", "nope")
	p := writeTemp(t, "flows.csv", body)
	got, err := readBench(p, 600000)
	if err != nil {
		t.Fatalf("readBench: %v", err)
	}
	if len(got) != 41 {
		t.Fatalf("lignes banc = %d, want 41", len(got))
	}
	if got[0] != 600 || math.Abs(got[40]-600.52) > 1e-9 {
		t.Fatalf("conversion µs→ms altérée : premier %v, dernier %v", got[0], got[40])
	}
}

func values(n int, base float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = base + float64(i)/8
	}
	return out
}

func TestBuildReportIdenticalEMDZero(t *testing.T) {
	v := values(40, 600)
	fp, bp := pairFixtures(t, v)
	rep, err := buildReport(fp, bp, 600000)
	if err != nil {
		t.Fatalf("buildReport: %v", err)
	}
	if rep.FieldSamples != 40 || rep.BenchSamples != 40 {
		t.Fatalf("comptes = %d/%d, want 40/40", rep.FieldSamples, rep.BenchSamples)
	}
	if rep.EMD != 0 {
		t.Fatalf("EMD de jeux identiques = %v, want 0", rep.EMD)
	}
	if rep.State != metrics.FidelityTrust {
		t.Fatalf("state = %q, want %q", rep.State, metrics.FidelityTrust)
	}
	if rep.EMDCeiling != metrics.FidelityEMDMax {
		t.Fatalf("emd_ceiling = %v, want %v", rep.EMDCeiling, metrics.FidelityEMDMax)
	}
	assertSHA(t, fp, rep.FieldSHA256)
	assertSHA(t, bp, rep.BenchSHA256)
}

// Oracle hostile : jeux identiques mais décalés au-delà du plafond déclaré.
func TestBuildReportCeilingBreachDiscards(t *testing.T) {
	v := values(40, 600)
	fp, bp := pairFixturesShifted(t, v, 100)
	rep, err := buildReport(fp, bp, 600000)
	if err != nil {
		t.Fatalf("buildReport: %v", err)
	}
	field, _ := readField(fp)
	bench, _ := readBench(bp, 600000)
	want, err := metrics.DistributionEMD(field, bench)
	if err != nil {
		t.Fatalf("DistributionEMD oracle: %v", err)
	}
	if rep.EMD != want {
		t.Fatalf("EMD rendu = %v, oracle = %v", rep.EMD, want)
	}
	if rep.EMD <= metrics.FidelityEMDMax {
		t.Fatalf("EMD décalé = %v, doit dépasser le plafond %v", rep.EMD, metrics.FidelityEMDMax)
	}
	if rep.State != metrics.FidelityDiscard {
		t.Fatalf("state = %q, want %q (EMD au-dessus du plafond)", rep.State, metrics.FidelityDiscard)
	}
	if !strings.Contains(rep.Note, fmt.Sprintf("%.6f", rep.EMD)) ||
		!strings.Contains(rep.Note, fmt.Sprintf("%.6f", metrics.FidelityEMDMax)) {
		t.Fatalf("note ne nomme pas l'EMD mesuré et le plafond : %q", rep.Note)
	}
	// Preuve contraire : même avec trop peu d'échantillons, le dépassement refuse.
	shortF, shortB := pairFixturesShifted(t, values(12, 600), 100)
	short, err := buildReport(shortF, shortB, 600000)
	if err != nil {
		t.Fatalf("buildReport (12 samples): %v", err)
	}
	if short.State != metrics.FidelityDiscard {
		t.Fatalf("state (12 samples, EMD > plafond) = %q, want %q", short.State, metrics.FidelityDiscard)
	}
}

func TestBuildReportBoundedBelowFloor(t *testing.T) {
	for _, tc := range []struct {
		name      string
		fieldSize int
		benchSize int
	}{
		{"terrain court", 12, 40},
		{"banc court", 40, 12},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fp, _ := pairFixtures(t, values(tc.fieldSize, 600))
			_, bp := pairFixtures(t, values(tc.benchSize, 600.5))
			rep, err := buildReport(fp, bp, 600000)
			if err != nil {
				t.Fatalf("buildReport: %v", err)
			}
			if rep.EMD > metrics.FidelityEMDMax {
				t.Fatalf("fixture au-dessus du plafond (%v) : le cas bounded est invalide", rep.EMD)
			}
			if rep.State != metrics.FidelityBounded {
				t.Fatalf("state = %q, want %q", rep.State, metrics.FidelityBounded)
			}
			if rep.Note == "" || !strings.Contains(rep.Note, "not trusted") {
				t.Fatalf("note insuffisante : %q", rep.Note)
			}
			// Aucun nombre fabriqué : l'EMD rendu est exactement celui de la bibliothèque.
			field, _ := readField(fp)
			bench, _ := readBench(bp, 600000)
			want, err := metrics.DistributionEMD(field, bench)
			if err != nil {
				t.Fatalf("DistributionEMD oracle: %v", err)
			}
			if rep.EMD != want {
				t.Fatalf("EMD rendu = %v, oracle bibliothèque = %v", rep.EMD, want)
			}
			assertSHA(t, fp, rep.FieldSHA256)
			assertSHA(t, bp, rep.BenchSHA256)
		})
	}
}

func TestBuildReportRefusesEmptySet(t *testing.T) {
	fp := writeTemp(t, "field.csv", fieldHeader)
	bp := writeTemp(t, "bench.csv", benchHeader)
	rep, err := buildReport(fp, bp, 600000)
	if err != nil {
		t.Fatalf("buildReport: %v", err)
	}
	if rep.State != metrics.FidelityDiscard {
		t.Fatalf("state = %q, want %q", rep.State, metrics.FidelityDiscard)
	}
	if !strings.Contains(rep.Note, "DistributionEMD") {
		t.Fatalf("note ne nomme pas le refus de la bibliothèque : %q", rep.Note)
	}
}

func TestBuildReportNoFabricatedSamples(t *testing.T) {
	// Les lignes sans rtt numérique ne deviennent jamais des échantillons à 0.
	fp := writeTemp(t, "field.csv", fieldBody(40)+"baseline,40,,,\nbaseline,41,xxx,,\n")
	bp := writeTemp(t, "bench.csv", benchBody(40, 600000)+benchRow("F", "nope")+benchRow("S", "999999"))
	rep, err := buildReport(fp, bp, 600000)
	if err != nil {
		t.Fatalf("buildReport: %v", err)
	}
	if rep.FieldSamples != 40 || rep.BenchSamples != 40 {
		t.Fatalf("échantillons comptés = %d/%d, want 40/40 (junk exclu)", rep.FieldSamples, rep.BenchSamples)
	}
	field, _ := readField(fp)
	bench, _ := readBench(bp, 600000)
	want, err := metrics.DistributionEMD(field, bench)
	if err != nil {
		t.Fatalf("DistributionEMD oracle: %v", err)
	}
	if rep.EMD != want {
		t.Fatalf("EMD rendu = %v, oracle = %v", rep.EMD, want)
	}
}
