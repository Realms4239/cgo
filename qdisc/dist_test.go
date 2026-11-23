package qdisc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

func TestBuildCalibrationRTTTable(t *testing.T) {
	rtts := make([]float64, 200)
	for i := range rtts {
		rtts[i] = 100 + float64(i%10)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	if cal.FloorMs != 100 || cal.HopDelayMs != 50 {
		t.Fatalf("floor/hop = %v/%v", cal.FloorMs, cal.HopDelayMs)
	}
	if cal.TableSize != distEntries || len(cal.Table) != distEntries {
		t.Fatalf("table = %d entrées, sidecar %d", len(cal.Table), cal.TableSize)
	}
	if cal.Table[len(cal.Table)-1] != distEntryMax {
		t.Fatalf("dernière entrée = %d, want %d (p99.9 mappé au max)", cal.Table[len(cal.Table)-1], distEntryMax)
	}
	if cal.SourceP50RTTms < 104 || cal.SourceP50RTTms > 106 {
		t.Fatalf("p50 source = %v", cal.SourceP50RTTms)
	}
	if cal.RoundTripP50RTTms < 103 || cal.RoundTripP50RTTms > 107 {
		t.Fatalf("p50 reconstruit = %v", cal.RoundTripP50RTTms)
	}
	for i := 1; i < len(cal.Table); i++ {
		if cal.Table[i] < cal.Table[i-1] {
			t.Fatalf("table non monotone à %d : %d < %d", i, cal.Table[i], cal.Table[i-1])
		}
	}
	for _, v := range cal.Table {
		if v < 0 || v > distEntryMax {
			t.Fatalf("entrée hors bornes : %d", v)
		}
	}
	lines := strings.Split(strings.TrimSpace(string(cal.Render())), "\n")
	if len(lines) != distEntries {
		t.Fatalf("lignes .dist = %d", len(lines))
	}
	for _, ln := range lines {
		if len(ln) > 1 && ln[0] == '0' {
			t.Fatalf("zéro-padding interdit (strtol base 0) : %q", ln)
		}
	}
}

func TestBuildCalibrationRejectsBadTrace(t *testing.T) {
	short := make([]float64, distMinSamples-1)
	for i := range short {
		short[i] = 101
	}
	if _, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "a", short); err == nil {
		t.Fatal("trace courte acceptée")
	}

	below := make([]float64, distMinSamples)
	for i := range below {
		below[i] = 20
	}
	if _, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "a", below); err == nil {
		t.Fatal("trace sous le plancher acceptée")
	}
}

func TestQuantileNearestRank(t *testing.T) {
	s := []float64{1, 2, 3, 4}
	if got := quantile(s, 0.5); got != 3 {
		t.Fatalf("quantile p50 = %v, want 3", got)
	}
	if got := quantile(s, 0); got != 1 {
		t.Fatalf("quantile p0 = %v", got)
	}
	if got := quantile(s, 1); got != 4 {
		t.Fatalf("quantile p1 = %v", got)
	}
}

func TestWriteLoadCalibrationSealsDist(t *testing.T) {
	dir := t.TempDir()
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = 100 + float64(i%8)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	cal.SourceTelemetrySHA256 = "deadbeef"
	cal.Created = "2026-10-27T00:00:00Z"
	if err := WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCalibration(dir, "P2")
	if err != nil {
		t.Fatal(err)
	}
	if got.DistSHA256 == "" || got.SourceAudit != "audit-t" || got.TableSize != distEntries {
		t.Fatalf("sidecar relu : %+v", got)
	}
	if got.DistFile != "P2.dist" {
		t.Fatalf("dist_file = %q", got.DistFile)
	}
	if err := os.WriteFile(filepath.Join(dir, "P2.dist"), []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCalibration(dir, "P2"); err == nil {
		t.Fatal("hash divergent accepté")
	}
}

func TestLoadCalibrationMissingFailsClosed(t *testing.T) {
	if _, err := LoadCalibration(t.TempDir(), "P2"); err == nil {
		t.Fatal("calibration absente acceptée")
	}
}

// Oracle — vocabulaire fermé : seuls les trois plans déclarés sont valides.
// Le libellé legacy est un affichage, jamais une valeur stockable.
func TestValidatePhysicsPlaneClosedVocabulary(t *testing.T) {
	for _, p := range []string{PhysicsPlaneEBPFObservation, PhysicsPlaneRTTBaseline, PhysicsPlaneLiterature} {
		if err := ValidatePhysicsPlane(p); err != nil {
			t.Fatalf("plan %q refusé à tort: %v", p, err)
		}
	}
	for _, p := range []string{"", "bogus", "ebpf_observation_forced", PhysicsPlaneLegacy} {
		if err := ValidatePhysicsPlane(p); err == nil {
			t.Fatalf("plan %q accepté à tort", p)
		}
	}
}

// Oracle — le constructeur déclare honnêtement sa provenance RTT, jamais eBPF.
func TestBuildCalibrationDeclaresRTTBaselinePlane(t *testing.T) {
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = 100 + float64(i%8)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	if cal.PhysicsPlane != PhysicsPlaneRTTBaseline {
		t.Fatalf("plan = %q, want %q", cal.PhysicsPlane, PhysicsPlaneRTTBaseline)
	}
	if cal.PhysicsPlane == PhysicsPlaneEBPFObservation {
		t.Fatal("le constructeur RTT ne peut pas revendiquer eBPF")
	}
}

// Oracle — un sidecar archivé sans physics_plane lit le libellé legacy
// explicite, sans jamais être promu en observation eBPF.
func TestLoadCalibrationLegacyPlaneLabel(t *testing.T) {
	dir := t.TempDir()
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = 100 + float64(i%8)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
	// Sidecar antérieur au champ : seule la clé disparaît, les octets .dist
	// scellés restent intacts.
	sidecar := filepath.Join(dir, "P2.json")
	raw, err := os.ReadFile(sidecar)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	delete(doc, "physics_plane")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sidecar, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCalibration(dir, "P2")
	if err != nil {
		t.Fatalf("sidecar legacy refusé: %v", err)
	}
	if got.PhysicsPlane != PhysicsPlaneLegacy {
		t.Fatalf("plan legacy = %q, want %q", got.PhysicsPlane, PhysicsPlaneLegacy)
	}
	if got.PhysicsPlane == PhysicsPlaneEBPFObservation {
		t.Fatal("sidecar legacy promu en eBPF")
	}
}

// Oracle hostile — un plan inconnu (ou le libellé legacy) est refusé avant
// toute écriture de sidecar : aucun artefact partiel n'est publié.
func TestWriteCalibrationRejectsUnknownPlane(t *testing.T) {
	for _, plane := range []string{"bogus", PhysicsPlaneLegacy} {
		t.Run(plane, func(t *testing.T) {
			dir := t.TempDir()
			rtts := make([]float64, 64)
			for i := range rtts {
				rtts[i] = 100 + float64(i%8)
			}
			cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
			if err != nil {
				t.Fatal(err)
			}
			cal.PhysicsPlane = plane
			if err := WriteCalibration(dir, cal); err == nil {
				t.Fatalf("plan %q accepté à l'écriture", plane)
			}
			if _, err := os.Stat(filepath.Join(dir, "P2.json")); !os.IsNotExist(err) {
				t.Fatalf("sidecar écrit malgré le plan %q: %v", plane, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "P2.dist")); !os.IsNotExist(err) {
				t.Fatalf(".dist écrit malgré le plan %q: %v", plane, err)
			}
		})
	}
}

// Oracle hostile — un plan stocké inconnu (ou le libellé legacy) fait échouer
// la lecture fermée, avant le lancement de toute campagne.
func TestLoadCalibrationRejectsUnknownStoredPlane(t *testing.T) {
	for _, plane := range []string{"bogus", PhysicsPlaneLegacy} {
		t.Run(plane, func(t *testing.T) {
			dir := t.TempDir()
			rtts := make([]float64, 64)
			for i := range rtts {
				rtts[i] = 100 + float64(i%8)
			}
			cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
			if err != nil {
				t.Fatal(err)
			}
			if err := WriteCalibration(dir, cal); err != nil {
				t.Fatal(err)
			}
			sidecar := filepath.Join(dir, "P2.json")
			raw, err := os.ReadFile(sidecar)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			doc["physics_plane"] = plane
			raw, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sidecar, raw, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCalibration(dir, "P2"); err == nil {
				t.Fatalf("plan stocké %q accepté", plane)
			}
		})
	}
}

// Oracle F1 : source_ref porte l'id de trace, distinct de l'id d'audit, et
// reste sérialisé dans le sidecar scellé avec schema=1.
func TestBuildCalibrationSourceRefProvenance(t *testing.T) {
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = 100 + float64(i%8)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "organic-20261123-P2", rtts, "organic-20261123")
	if err != nil {
		t.Fatal(err)
	}
	if cal.Schema != 1 || cal.SourceAudit != "organic-20261123-P2" || cal.SourceRef != "organic-20261123" {
		t.Fatalf("provenance : schema=%d source_audit=%q source_ref=%q", cal.Schema, cal.SourceAudit, cal.SourceRef)
	}

	dir := t.TempDir()
	cal.SourceTelemetrySHA256 = "deadbeef"
	cal.Created = "2026-11-23T00:00:00Z"
	if err := WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "P2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["source_ref"] != "organic-20261123" || doc["source_audit"] != "organic-20261123-P2" || doc["schema"] != float64(1) {
		t.Fatalf("sidecar JSON : %s", raw)
	}
}
