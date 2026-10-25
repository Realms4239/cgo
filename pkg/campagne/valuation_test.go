package campagne

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

func TestWriteValuationSidecar(t *testing.T) {
	dir := t.TempDir()
	ev := model.Event{
		Profile: "P2", Qdisc: "cake", CC: "bbr",
		Bytes: 150 * 1000 * 1000, Drops: 1200, // ratio ≈ 1200×1448/150e6 ≈ 1.16 %
	}
	labels := ValuationLabels{ChargeSec: 120, Flows: 4, ChannelModel: "bursty", Bench: "dualns"}
	if err := WriteValuation(dir, ev, labels); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "valuation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v ValuationSidecar
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Entries) != 1 {
		t.Fatalf("une entrée par événement gelé : %+v", v.Entries)
	}
	e := v.Entries[0]
	// Tous les scénarios, aucun primaire marqué : une campagne ne déclare
	// pas de type de lien — la note du sidecar le dit.
	if len(e.Scenarios) != 5 {
		t.Fatalf("5 scénarios attendus : %+v", e.Scenarios)
	}
	if e.Primary != nil {
		t.Fatalf("sans type de lien, aucun primaire : %+v", e.Primary)
	}
	if !strings.Contains(v.Note, "aucun primaire marqué") {
		t.Fatalf("note de non-primaire absente : %q", v.Note)
	}
	if v.Labels.Flows != 4 || v.Labels.ChannelModel != "bursty" || v.Labels.Bench != "dualns" {
		t.Fatalf("étiquettes : %+v", v.Labels)
	}
	if e.Ratio.Basis != "estimated" || e.Ratio.Ratio <= 0 {
		t.Fatalf("ratio drops×1448 attendu (estimated) : %+v", e.Ratio)
	}
	if !strings.Contains(v.Disclaimer, "Scénarios contractuels") {
		t.Fatalf("disclaimer absent : %q", v.Disclaimer)
	}
}

func TestWriteValuationNoBytesNoScenarios(t *testing.T) {
	// Sans octets transportés : zéro scénario chiffré — jamais des zéros
	// qui se liraient « aucun gaspillage ».
	dir := t.TempDir()
	ev := model.Event{Profile: "P2", Qdisc: "cake", CC: "bbr", Bytes: 0, Drops: 0}
	if err := WriteValuation(dir, ev, ValuationLabels{ChargeSec: 120, Flows: 4}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "valuation.json"))
	var v ValuationSidecar
	json.Unmarshal(raw, &v)
	if len(v.Entries) != 1 || len(v.Entries[0].Scenarios) != 0 {
		t.Fatalf("sans octets : zéro scénario chiffré : %+v", v.Entries)
	}
}

func TestValuationIdempotentByCell(t *testing.T) {
	// La propriété de reprise : une cellule rejouée REMPLACE son entrée,
	// elle ne la duplique pas — le sidecar reste lisible après un crash.
	dir := t.TempDir()
	labels := ValuationLabels{ChargeSec: 120, Flows: 4, ChannelModel: "clean", Bench: "dualns"}
	ev := model.Event{Profile: "P2", Qdisc: "cake", CC: "bbr", Bytes: 80_000_000, Drops: 300}
	if err := WriteValuation(dir, ev, labels); err != nil {
		t.Fatal(err)
	}
	rejoue := ev
	rejoue.Drops = 900 // mêmes profile/qdisc/cc : remplacement, pas doublon
	if err := WriteValuation(dir, rejoue, labels); err != nil {
		t.Fatal(err)
	}
	autre := model.Event{Profile: "P3", Qdisc: "cake", CC: "bbr", Bytes: 80_000_000, Drops: 10}
	if err := WriteValuation(dir, autre, labels); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "valuation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v ValuationSidecar
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Entries) != 2 {
		t.Fatalf("reprise sans doublon : 2 entrées attendues, %+v", v.Entries)
	}
	for _, e := range v.Entries {
		if e.Cell == "P2/cake/bbr" && e.Drops != 900 {
			t.Fatalf("l'entrée rejouée doit être remplacée : %+v", e)
		}
	}
}

func TestBuildMatrixConfigRecordsLabels(t *testing.T) {
	// Le manifeste doit porter flows (défaut 4), channel_model et bench :
	// la provenance d'un run se lit dans son gel, pas dans la mémoire.
	d := Deps{Flows: 0, ChannelModel: "bursty", Bench: "dualns",
		Direction: "up", Target: "10.30.0.2", BulkAddr: "10.30.0.2:5201"}
	raw, err := buildMatrixConfig([]string{"P2"}, []string{"cake"}, []string{"bbr"}, 1, d)
	if err != nil {
		t.Fatal(err)
	}
	var c matrixConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	if c.Flows != 4 {
		t.Fatalf("défaut flows = 4 : %d", c.Flows)
	}
	if c.ChannelModel != "bursty" || c.Bench != "dualns" {
		t.Fatalf("étiquettes : %+v %+v", c.ChannelModel, c.Bench)
	}
	// Legacy par défaut : rien de déclaré ⇒ legacy/clean, jamais vide.
	raw2, _ := buildMatrixConfig([]string{"P2"}, nil, nil, 1, Deps{})
	var c2 matrixConfig
	json.Unmarshal(raw2, &c2)
	if c2.Bench != "legacy" || c2.ChannelModel != "clean" || c2.Flows != 4 {
		t.Fatalf("défauts legacy/clean/4 : %+v", c2)
	}
}

func TestSnapshotCostFields(t *testing.T) {
	// Le snapshot live porte le coût entreprise et le JFI aux côtés des
	// champs gelés historiques (sémantique de cost_ar_per_h inchangée).
	s := Snapshot{CostPrimaryARPerH: 138.9, CostB2BARPerH: 62.9, JFIPct: 97.5}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"cost_primary_ar_per_h":138.9`, `"cost_b2b_ar_per_h":62.9`, `"jfi_pct":97.5`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("clé absente du snapshot : %s (%s)", key, raw)
		}
	}
}
