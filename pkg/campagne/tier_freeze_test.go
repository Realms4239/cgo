package campagne

import (
	"context"
	"encoding/json"
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
