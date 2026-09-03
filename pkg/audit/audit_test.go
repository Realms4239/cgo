package audit

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAuditRunHappyPath(t *testing.T) {
	p := Params{AuditID: "a1", Site: "Dept X", LinkType: "5g", Provider: "Yas", Duration: 1, Target: "127.0.0.1"}
	deps := Deps{
		Ping:  func(ctx context.Context, _ string, n int) []float64 { return []float64{20, 21, 22} },
		Small: func(ctx context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), p, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.RTTIdleP50 == 0 || res.HTTPSmallP95 == 0 {
		t.Fatalf("empty result %+v", res)
	}
}

func TestAppendLinkAudit(t *testing.T) {
	dir := t.TempDir()
	r := &Result{AuditID: "a1", Site: "Dept X", LinkType: "5g", Provider: "Yas", RTTIdleP50: 20, RTTIdleP95: 25, HTTPSmallP95: 30}
	if err := AppendLinkAudit(dir, r); err != nil {
		t.Fatal(err)
	}
	if err := AppendLinkAudit(dir, r); err != nil {
		t.Fatal(err)
	}
	// le fichier doit avoir l'en-tête + 2 lignes
}

// TestAuditLossCountsAttempts — ping lent (1,2 s/appel) mais complet : la
// perte doit être ~0. L'ancien dénominateur (durée × 12,5/s) ignorait la durée
// réelle d'un appel et annonçait ~20 %+ sur un lien parfait.
func TestAuditLossCountsAttempts(t *testing.T) {
	p := Params{AuditID: "a-loss", Site: "X", LinkType: "4g", Duration: 2, Target: "127.0.0.1"}
	deps := Deps{
		Ping: func(ctx context.Context, _ string, _ int) []float64 {
			select {
			case <-time.After(1200 * time.Millisecond):
			case <-ctx.Done():
				return nil
			}
			return []float64{20, 21, 22, 23, 24}
		},
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), p, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.LossPct > 5 {
		t.Fatalf("loss = %v %% on a perfect slow link, want ~0", res.LossPct)
	}
}

// TestAuditLossyLink — 3 réponses sur 5 demandées = 40 % de perte, quel que
// soit le rythme des appels.
func TestAuditLossyLink(t *testing.T) {
	p := Params{AuditID: "a-lossy", Site: "X", LinkType: "4g", Duration: 1, Target: "127.0.0.1"}
	deps := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{20, 21, 22} },
		Small: func(context.Context) (float64, error) { return 25, nil },
	}
	res, err := Run(context.Background(), p, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.LossPct < 39.5 || res.LossPct > 40.5 {
		t.Fatalf("loss = %v %%, want 40 (3/5 replies per call)", res.LossPct)
	}
}

// TestAuditNoSmallURLHonest — sans SmallURL : petit objet NON mesuré (0 +
// note), jamais la constante 25 ms synthétique dans le CSV gelé.
func TestAuditNoSmallURLHonest(t *testing.T) {
	p := Params{AuditID: "a-small", Site: "X", LinkType: "4g", Duration: 1, Target: "127.0.0.1"}
	deps := Deps{
		Ping: func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 24} },
		// Small nil → défaut, SmallURL vide
	}
	res, err := Run(context.Background(), p, deps)
	if err != nil {
		t.Fatal(err)
	}
	if res.HTTPSmallP95 != 0 {
		t.Fatalf("HTTPSmallP95 = %v without SmallURL, want 0 (no synthetic 25 ms)", res.HTTPSmallP95)
	}
	if !strings.Contains(res.Notes, "petit objet non mesuré") {
		t.Fatalf("notes must flag unmeasured small object: %q", res.Notes)
	}
}
