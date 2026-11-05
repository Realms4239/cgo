package campagne

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

func TestMatrixConfigFreezesActiveTier(t *testing.T) {
	cfg, err := buildMatrixConfig([]string{"P2"}, []string{"cake"}, []string{"bbr"}, 1, Deps{})
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Tier string `json:"tier"`
	}
	if err := json.Unmarshal(cfg, &got); err != nil {
		t.Fatal(err)
	}
	if got.Tier != metrics.ActiveTier().Name {
		t.Fatalf("tier gelé = %q, voulu %q", got.Tier, metrics.ActiveTier().Name)
	}
}

func TestTierByNameRejectsUnknown(t *testing.T) {
	if _, ok := metrics.TierByName("pas-un-forfait"); ok {
		t.Fatal("palier inconnu accepté")
	}
	tr, ok := metrics.TierByName(metrics.ActiveTier().Name)
	if !ok || tr.Name == "" {
		t.Fatal("palier valide refusé")
	}
}

// Le coût d'une ligne vient du palier gelé avec son run, pas du global courant.
func TestCostARPerHUsesFrozenTier(t *testing.T) {
	if got := costARPerH("yas-day-1gb", 1073741824); got != 20000 {
		t.Fatalf("palier gelé ignoré : %v, voulu 20000", got)
	}
	if got := costARPerH("pas-un-forfait", 1073741824); got != costARPerH(metrics.ActiveTier().Name, 1073741824) {
		t.Fatalf("palier inconnu doit retomber sur le global sans inventer : %v vs %v", got, costARPerH(metrics.ActiveTier().Name, 1073741824))
	}
}

// La ligne gelée porte le coût du palier du run : le global (5556 Ar/Go) et le
// palier gelé (1000 Ar/Go) donnent deux chiffres distincts, donc un câblage
// silencieusement oublié ne peut pas passer.
func TestRunEventCostUsesFrozenTier(t *testing.T) {
	d := fastDeps()
	d.Tier = "yas-day-1gb"
	calls := 0
	d.StatsFn = func() []qdisc.Stats {
		calls++
		// 1 = socle initial, 2 = re-socle de charge, 3 = comptage final
		// (même séquence que TestRunEventCounterResetClamped).
		if calls <= 2 {
			return []qdisc.Stats{{Kind: "cake", Bytes: 10_000_000, Drops: 0}}
		}
		return []qdisc.Stats{{Kind: "cake", Bytes: 10_000_000, Drops: 10}}
	}
	ev := model.Event{RunID: "r1", EventID: 10, Profile: "P2", Qdisc: model.Cake, CC: model.Cubic, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d)
	if err != nil {
		t.Fatal(err)
	}
	if got.Drops != 10 {
		t.Fatalf("drops = %d, voulu 10", got.Drops)
	}
	want := round1(float64(10*1448) / metrics.GB * 1000 * 20)
	if got.CostARPerH != want {
		t.Fatalf("coût = %v, voulu %v au palier gelé (1000 Ar/Go) — le global 5556 donnerait %v",
			got.CostARPerH, want, round1(float64(10*1448)/metrics.GB*5556*20))
	}
}

// Une reprise garde la base de coût du run, pas le palier global du jour :
// Freeze a écrit la config dans manifest.json, OpenRun ne relit que le CSV.
func TestResumeKeepsTheFrozenTier(t *testing.T) {
	dir := t.TempDir()
	run := "run-interrompu"
	cfg := []byte(`{"profiles":["P2"],"qdiscs":["cake"],"ccs":["bbr"],"reps":3,"tier":"yas-day-1gb"}`)
	sum := sha256.Sum256(cfg)
	doc, err := json.Marshal(map[string]any{
		"config_sha256": hex.EncodeToString(sum[:]),
		"config":        string(cfg),
		"files":         []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, run), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, run, "manifest.json"), doc, 0o644); err != nil {
		t.Fatal(err)
	}

	before := metrics.ActiveTierName()
	if err := metrics.SetDefaultTier("yas-ftth-100gb"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = metrics.SetDefaultTier(before) })

	if got := tierForRun(dir, run); got != "yas-day-1gb" {
		t.Fatalf("reprise = %q, voulu le palier gelé yas-day-1gb (global %q)", got, metrics.ActiveTierName())
	}
	if got := tierForRun(dir, "run-neuf"); got != "yas-ftth-100gb" {
		t.Fatalf("run neuf = %q, voulu le global yas-ftth-100gb", got)
	}
}
