package campagne

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/Realms4239/cgo/model"
)

// Le clamp NaN reste local au pourcentage de deadline : int(NaN) vaut
// min-int64 et fabriquerait un grand nombre fini en CSV et en JSON.
func TestDeadlinePctNaNNeverFabricatesFiniteValue(t *testing.T) {
	got := deadlinePct(nil, 220)
	if math.IsNaN(got) || math.IsInf(got, 0) || got != 0 {
		t.Fatalf("deadlinePct(nil, 220) = %v, attendu 0 (absence marquée par le booléen)", got)
	}
}

// Cellule sans aucune complétion small : le pourcentage ne doit pas être un
// nombre fabriqué. Le drapeau d'absence est vérifié dans le test de provenance.
func TestRunEventDeadlineNeverFabricatesAbsentValue(t *testing.T) {
	d := fastDeps()
	d.Small = func(context.Context) (float64, error) { return 0, errors.New("sonde indisponible") }
	ev := model.Event{RunID: "r1", EventID: 1, Profile: "P3", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P3"], d)
	if err != nil {
		t.Fatal(err)
	}
	if math.IsNaN(got.DeadlineOKPct) || math.IsInf(got.DeadlineOKPct, 0) || got.DeadlineOKPct != 0 {
		t.Fatalf("DeadlineOKPct sans complétion = %v, attendu 0 (absence, jamais un nombre fabriqué)", got.DeadlineOKPct)
	}
}
