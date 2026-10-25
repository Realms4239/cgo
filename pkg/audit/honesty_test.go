package audit

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestAuditNoLoadDoesNotGrade — sans charge bulk, l'écart idle→chargé n'est
// pas mesuré : la note doit rester vide et les notes le dire. L'ancien
// comportement notait A+ un lien jamais chargé (jitter déguisé en preuve).
func TestAuditNoLoadDoesNotGrade(t *testing.T) {
	d := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 24} },
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), Params{Duration: 1, Target: "t"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if res.BloatGrade != "" {
		t.Fatalf("BloatGrade = %q sans charge, want vide (non mesuré)", res.BloatGrade)
	}
	if res.BloatDeltaMs != 0 {
		t.Fatalf("BloatDeltaMs = %v sans charge, want 0", res.BloatDeltaMs)
	}
	if !strings.Contains(strings.ToLower(res.Notes), "non mesuré") {
		t.Fatalf("notes sans charge : %q — doit dire non mesuré", res.Notes)
	}
}

// TestAuditMeasuresLoadAndGrades — quand un bulk est câblé, la fenêtre
// chargée doit coïncider avec la charge, le débit doit être calculé sur la
// durée réelle, et la note doit venir du p95 chargé (pas d'un zéro).
func TestAuditMeasuresLoadAndGrades(t *testing.T) {
	var loading int32
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	d := Deps{
		Ping: func(context.Context, string, int) []float64 {
			if atomic.LoadInt32(&loading) == 1 {
				return []float64{120, 121, 122, 123, 124}
			}
			return []float64{20, 21, 22, 23, 24}
		},
		Small: func(context.Context) (float64, error) { return 25, nil },
		Bulk: func(ctx context.Context, _ string) (uint64, error) {
			atomic.StoreInt32(&loading, 1)
			<-ctx.Done()
			return 1_000_000, nil
		},
	}
	res, err := Run(ctx, Params{Duration: 2, Target: "t", BulkAddr: "fake:1"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if res.RTTLoadedP50 <= res.RTTIdleP50 {
		t.Fatalf("idle %v >= loaded %v : la fenêtre chargée ne couvre pas la charge", res.RTTIdleP50, res.RTTLoadedP50)
	}
	if res.ThroughputMbps <= 0 {
		t.Fatalf("ThroughputMbps = %v avec bulk mesuré, want > 0", res.ThroughputMbps)
	}
	if res.DataUsedMB <= 0 {
		t.Fatalf("DataUsedMB = %v avec bulk mesuré, want > 0", res.DataUsedMB)
	}
	if res.BloatDeltaMs < 50 {
		t.Fatalf("BloatDeltaMs = %v, want p95 chargé − p95 idle ≈ 100", res.BloatDeltaMs)
	}
	if res.BloatGrade == "" || res.BloatGrade == "A+" {
		t.Fatalf("BloatGrade = %q sur un écart de ~100 ms, want une note réelle", res.BloatGrade)
	}
	if res.BloatVerdict == "" {
		t.Fatal("verdict vide sur une note mesurée")
	}
}

// TestMatchProfileRequiresThroughput — sans débit mesuré, aucun profil P1–P4
// ne doit être revendiqué : un délai seul ne distingue pas 4G et Starlink.
func TestMatchProfileRequiresThroughput(t *testing.T) {
	id, delta := MatchProfile(&Result{RTTIdleP95: 46})
	if id != "" {
		t.Fatalf("MatchProfile sans débit = %q, want vide (non déterminé)", id)
	}
	if !strings.Contains(delta, "débit non mesuré") {
		t.Fatalf("delta = %q, doit dire que le débit n'est pas mesuré", delta)
	}
}
