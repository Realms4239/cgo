package audit

import (
	"context"
	"testing"
	"time"
)

// Découpe 3 fenêtres : 0–12 s idle, 12–22 s bulk, 22–30 s chargé — le RTT chargé doit différer de l'idle
// quand le lien est réellement chargé (pings factices simulent la charge apparaissant avec la fenêtre bulk).
// Audit réel de 30 s par conception (découpe B Q26) — pas de couture d'horloge factice, durée honnête.
func TestAuditSplit(t *testing.T) {
	if testing.Short() {
		t.Skip("30s real audit — skip in -short")
	}
	start := time.Now()
	p := Params{Duration: 30, Target: "1.1.1.1"}
	res, err := Run(context.Background(), p, Deps{
		Ping: func(_ context.Context, _ string, _ int) []float64 {
			// la charge s'active à la marque 22 s (le bulk flood 12–22 s sature le lien)
			if time.Since(start) > 22*time.Second {
				return []float64{120, 121}
			}
			return []float64{20, 21}
		},
		Small: func(_ context.Context) (float64, error) { return 50, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RTTIdleP50 == res.RTTLoadedP50 {
		t.Fatalf("idle==loaded not split: %v == %v", res.RTTIdleP50, res.RTTLoadedP50)
	}
	if res.RTTLoadedP50 <= res.RTTIdleP50 {
		t.Fatalf("loaded should exceed idle under bulk: idle %v loaded %v", res.RTTIdleP50, res.RTTLoadedP50)
	}
}