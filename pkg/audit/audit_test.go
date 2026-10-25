package audit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestRunDirectionDown — down : la charge doit passer par
// BulkLoad(direction="down") et remplir les champs miroirs
// (ThroughputDownMbps, JFIDownPct). Le headline suit le sens exercé
// (résolution d'ambiguïté du brief : MatchProfile/BuildReport lisent les
// champs headline) : pour un audit down, headline ET miroir portent la
// fenêtre descendante, Direction dit d'où ils viennent.
func TestRunDirectionDown(t *testing.T) {
	var gotDir string
	d := Deps{
		Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{20, 21, 22, 23, 30} },
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _ string, dir string, _ int) LoadOutcome {
			gotDir = dir
			return LoadOutcome{Bytes: 25 * 1000 * 1000, PerFlow: []uint64{6e6, 6e6, 6e6, 7e6}, Elapsed: 10 * time.Second}
		},
	}
	res, err := Run(context.Background(), Params{Duration: 30, Direction: "down", Target: "x"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if gotDir != "down" {
		t.Fatalf("direction passée au chargeur : %q", gotDir)
	}
	if res.ThroughputDownMbps == 0 || res.ThroughputMbps == 0 {
		t.Fatalf("down : headline et miroir portent le débit mesuré : up=%v down=%v", res.ThroughputMbps, res.ThroughputDownMbps)
	}
	if res.JFIDownPct <= 0 || res.JFIPct <= 0 {
		t.Fatalf("JFI down = %v, JFI headline = %v", res.JFIDownPct, res.JFIPct)
	}
	if res.Direction != "down" {
		t.Fatalf("Direction gelé = %q", res.Direction)
	}
	if res.FlowCount != 4 {
		t.Fatalf("FlowCount = %d, want 4 (défaut)", res.FlowCount)
	}
	if res.RTTLoadedDownP95 == 0 || res.RTTLoadedP95 == 0 {
		t.Fatalf("RTT chargé down : miroir=%v headline=%v, les deux doivent être mesurés", res.RTTLoadedDownP95, res.RTTLoadedP95)
	}
	if res.GradeDown == "" || res.BloatGrade == "" {
		t.Fatalf("note down mesurée : grade=%q grade_down=%q", res.BloatGrade, res.GradeDown)
	}
}

// TestRunDirectionBothRunsTwoLoadWindows — both : une seule baseline partagée,
// fenêtre up PUIS fenêtre down (avec répit entre les deux), mini-drain up →
// DrainMs, drain final down → DrainDownMs.
func TestRunDirectionBothRunsTwoLoadWindows(t *testing.T) {
	var dirs []string
	d := Deps{
		Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{20, 21, 22, 23, 30} },
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _ string, dir string, _ int) LoadOutcome {
			dirs = append(dirs, dir)
			return LoadOutcome{Bytes: 10 * 1000 * 1000, PerFlow: []uint64{25e5, 25e5, 25e5, 25e5}, Elapsed: 8 * time.Second}
		},
	}
	res, err := Run(context.Background(), Params{Duration: 75, Direction: "both", Target: "x"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 2 || dirs[0] != "up" || dirs[1] != "down" {
		t.Fatalf("both = fenêtre up puis fenêtre down : %v", dirs)
	}
	if res.ThroughputMbps == 0 || res.ThroughputDownMbps == 0 {
		t.Fatalf("les deux sens doivent être mesurés : up=%v down=%v", res.ThroughputMbps, res.ThroughputDownMbps)
	}
	if !res.DrainMeasured || !res.DrainDownMeasured {
		t.Fatalf("both : mini-drain (up) et drain final (down) mesurés : up=%v down=%v", res.DrainMeasured, res.DrainDownMeasured)
	}
	if res.JFIDownPct != 100 || res.JFIPct != 100 {
		t.Fatalf("4 flux égaux = JFI 100 %% : up=%v down=%v", res.JFIPct, res.JFIDownPct)
	}
}

func TestRunRetransRatioBasis(t *testing.T) {
	// TCP_INFO rendu (RetransMeasured) => basis "measured" ; sinon le proxy
	// perte×volume => "estimated" ; sans perte => "" (rien à chiffrer).
	d := Deps{
		Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{20, 21, 22} },
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _ string, _ string, _ int) LoadOutcome {
			return LoadOutcome{Bytes: 10e6, PerFlow: []uint64{10e6}, RetransBytes: 100e3, RetransMeasured: true, Elapsed: 5 * time.Second}
		},
	}
	res, err := Run(context.Background(), Params{Duration: 20, Direction: "up", Target: "x"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if res.RetransRatio.Basis != "measured" || res.RetransRatio.Ratio <= 0 {
		t.Fatalf("ratio mesuré attendu : %+v", res.RetransRatio)
	}
}

// TestRunRetransRatioEstimatedBasis — sans TCP_INFO mais avec de la perte
// sonde : le proxy documenté perte×volume, base "estimated" (jamais
// présentée comme une mesure).
func TestRunRetransRatioEstimatedBasis(t *testing.T) {
	d := Deps{
		Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{20, 21, 22} }, // 3/5 = 40 % de perte
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _ string, _ string, _ int) LoadOutcome {
			return LoadOutcome{Bytes: 10e6, PerFlow: []uint64{10e6}, Elapsed: 5 * time.Second}
		},
	}
	res, err := Run(context.Background(), Params{Duration: 20, Direction: "up", Target: "x"}, d)
	if err != nil {
		t.Fatal(err)
	}
	if res.RetransRatio.Basis != "estimated" {
		t.Fatalf("proxy attendu sans TCP_INFO : %+v", res.RetransRatio)
	}
	if res.RetransRatio.Ratio <= 0.39 || res.RetransRatio.Ratio >= 0.41 {
		t.Fatalf("ratio proxy = perte/100 = 0.40, got %v", res.RetransRatio.Ratio)
	}
}

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
