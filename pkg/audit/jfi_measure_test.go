package audit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/remediate"
)

// TestJFIIsSealedWithItsOwnMeasure — metrics.JFI rend 0 quand il n'y a pas de
// valeurs par flux, et 0 × 100 s'affiche « 0,0 % d'équité ». Un opérateur qui lit
// ça voit un réseau totalement injuste alors qu'aucune équité n'a été mesurée :
// c'est un zéro qui se fait mesure, la classe de défaut que ce projet refuse
// depuis le début. Le fait « mesuré ou non » est donc scellé à côté du nombre,
// et le Markdown rend une absence quand il n'y a pas de nombre.
func TestJFIIsSealedWithItsOwnMeasure(t *testing.T) {
	run := func(perFlow []uint64) *Result {
		d := Deps{
			Ping:  func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 30} },
			Small: func(context.Context) (float64, error) { return 15, nil },
			BulkLoad: func(context.Context, string, string, int) LoadOutcome {
				return LoadOutcome{Bytes: 25_000_000, PerFlow: perFlow, Elapsed: 10 * time.Second}
			},
		}
		res, err := Run(context.Background(), Params{AuditID: "audit-jfi", Site: "tana", LinkType: "4g", Duration: 2, Flows: 4, Target: "x", BulkAddr: "fake:5201"}, d)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	// Pas de contribution par flux : l'index n'est pas calculable.
	noPerFlow := run(nil)
	if noPerFlow.JFIMeasured {
		t.Errorf("JFI annoncé mesuré sans valeur par flux (jfi_pct=%v)", noPerFlow.JFIPct)
	}
	if noPerFlow.FlowCount <= 1 {
		t.Errorf("le garde actuel de la console est flow_count>1 ; avec %d flux et JFI non mesuré, elle affiche 0,0 %% comme une mesure", noPerFlow.FlowCount)
	}

	// Quatre contributions réelles : l'index est mesuré.
	withPerFlow := run([]uint64{6e6, 6e6, 6e6, 7e6})
	if !withPerFlow.JFIMeasured {
		t.Errorf("JFI annoncé non mesuré avec quatre contributions par flux (jfi=%v)", withPerFlow.JFIPct)
	}

	// Et le Markdown suit le fait scellé, pas le seul compteur de flux.
	md := func(res *Result) string {
		rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
		if err != nil {
			t.Fatal(err)
		}
		return rep.RenderMarkdown()
	}
	if got := md(noPerFlow); strings.Contains(got, "Équité inter-flux (JFI) | 0.0") {
		t.Errorf("le rapport énonce une équité de 0,0 %% là où rien n'est mesuré :\n%s", section(got, "Équité"))
	}
	if got := md(withPerFlow); !strings.Contains(got, "Équité inter-flux (JFI)") {
		t.Error("le rapport rend une équité mesurée comme si elle ne l'était pas")
	}
}

// TestJFIMeasuredSurvivesTheRoundTrip — la console ne peut pas deviner : le fait
// doit sortir du JSON.
func TestJFIMeasuredSurvivesTheRoundTrip(t *testing.T) {
	res := &Result{AuditID: "a", Site: "s", LinkType: "4g", FlowCount: 4, JFIPct: 0, JFIMeasured: false}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Measured.JFIMeasured {
		t.Error("jfi_measured doit être publié à false, pas omis")
	}
	b, err := rep.RenderJSON()
	if err != nil {
		t.Fatal(err)
	}
	// Le fait doit être présent et dire false — pas omis, ce qui laisserait la
	// console deviner. On relit le JSON plutôt que de chasser une octet-exacte.
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	measured, ok := raw["measured"].(map[string]any)
	if !ok {
		t.Fatal("pas d'objet measured dans le JSON")
	}
	v, present := measured["jfi_measured"]
	if !present {
		t.Fatalf("jfi_measured absent du contrat JSON : %s", section(string(b), "jfi"))
	}
	if v != false {
		t.Errorf("jfi_measured publié à %v, attendu false", v)
	}
}

// section — extrait de contexte pour le message d'échec, en runes : couper une
// chaîne à l'octet est exactement le défaut que la porte d'encodage refuse.
func section(s, needle string) string {
	i := strings.Index(s, needle)
	if i < 0 {
		return "(absent)"
	}
	r := []rune(s[i:])
	if len(r) > 160 {
		r = r[:160]
	}
	return string(r)
}
