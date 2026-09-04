package campagne

import (
	"context"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// TestG4DownDeadPipeOnly — en download, le plancher G4 ne détecte que le
// tuyau mort : un goodput faible mais non nul (horlogé ACK, prouvé sur le
// banc : ~0,5-2,5 Mb/s mono-flux tous qdiscs) reste VALIDE — le verdict
// download se joue sur la latence. Zéro = invalide.
func TestG4DownDeadPipeOnly(t *testing.T) {
	mkdeps := func() Deps {
		d := fastDeps()
		d.Direction = "down"
		return d
	}
	// 100 ko en 1 s de charge instantanée ≈ 0,8 Mb/s : tuyau vivant
	d := mkdeps()
	d.Bulk = func(context.Context, string) (uint64, error) { return 100_000, nil }
	ev := model.Event{RunID: "g4d", EventID: 1, Profile: "P2", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
	got, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d)
	if err != nil {
		t.Fatal(err)
	}
	if got.GateStatus == model.GateInvalid {
		t.Fatalf("download 0,8 Mb/s invalidé — le plancher down doit tolérer l'horloge ACK (goodput=%.1f)", got.BulkGoodputMbps)
	}
	// zéro octet : tuyau mort → invalidé
	d2 := mkdeps()
	d2.Bulk = func(context.Context, string) (uint64, error) { return 0, nil }
	got2, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d2)
	if err != nil {
		t.Fatal(err)
	}
	if got2.GateStatus != model.GateInvalid {
		t.Fatalf("download 0 octet validé — le tuyau mort doit être invalidé: %+v", got2)
	}
	// montant inchangé : 0,8 Mb/s sur P2 (plancher 10) reste invalidé
	d3 := fastDeps()
	d3.Bulk = func(context.Context, string) (uint64, error) { return 100_000, nil }
	got3, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d3)
	if err != nil {
		t.Fatal(err)
	}
	if got3.GateStatus != model.GateInvalid {
		t.Fatal("upload 0,8 Mb/s validé — le plancher montant 50 % doit tenir")
	}
}
