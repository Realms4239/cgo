package audit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
	"github.com/Realms4239/cgo/remediate"
)

// BuildReport dérive Measured.LoadMeasured de
// « res.ThroughputMbps > 0 && res.BloatGrade != "" » (pkg/audit/report.go:225).
// Les cas ci-dessous jouent donc sur ces deux champs, jamais sur un booléen
// LoadMeasured qui n'existe pas sur Result.
func resultForVerdict(mut func(*Result)) *Result {
	res := &Result{
		AuditID: "audit-verdict", Site: "Siège", LinkType: "4g",
		RTTIdleP50: 24, RTTIdleP95: 31,
		LoadAttempted: true, ThroughputMbps: 18.2, LoadSeconds: 40,
		BloatDeltaMs: 12, BloatGrade: "A", BloatVerdict: "ok",
	}
	if mut != nil {
		mut(res)
	}
	return res
}

func reportOf(t *testing.T, res *Result) Report {
	t.Helper()
	rep, err := BuildReport(res, remediate.Link{}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	return *rep
}

func TestReportCarriesOperatorVerdict(t *testing.T) {
	rep := reportOf(t, resultForVerdict(nil))
	if rep.OperatorVerdict.Level != metrics.VerdictGreen || rep.OperatorVerdict.Word != "Lien sain" {
		t.Fatalf("verdict = %+v", rep.OperatorVerdict)
	}
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"operator_verdict"`) {
		t.Fatal("verdict absent du JSON scellé")
	}
	if !strings.Contains(rep.RenderMarkdown(), "Verdict opérateur") {
		t.Fatal("verdict absent du Markdown")
	}
}

// Le cas de la revue : sonde d'alerte réussie, charge jamais établie.
func TestReportVerdictRefusesGreenWithoutLoad(t *testing.T) {
	rep := reportOf(t, resultForVerdict(func(r *Result) {
		r.ThroughputMbps = 0
		r.BloatGrade = ""
		r.BloatDeltaMs = 0
		r.CAPSLAMeasured = true
		r.CAPDeadlineOKPct = 100
	}))
	if rep.OperatorVerdict.Level == metrics.VerdictGreen {
		t.Fatalf("vert sans charge établie : %+v", rep.OperatorVerdict)
	}
	if rep.OperatorVerdict.Word != "Charge non établie" {
		t.Fatalf("mot = %q", rep.OperatorVerdict.Word)
	}
}

func TestReportVerdictRedOnSlaBreach(t *testing.T) {
	rep := reportOf(t, resultForVerdict(func(r *Result) {
		r.CAPSLAMeasured = true
		r.CAPDeadlineOKPct = 60.2
		r.CAPSLABreach = true
	}))
	if rep.OperatorVerdict.Level != metrics.VerdictRed || rep.OperatorVerdict.Word != "Alertes en retard" {
		t.Fatalf("verdict = %+v", rep.OperatorVerdict)
	}
}

// load_attempted est scellé à deux niveaux : la ligne d'audit et le rapport.
// Sans le fait « la fenêtre était prévue », un puits injoignable se relit
// « jamais testé » — deux absences différentes, jamais interchangeables.
func TestReportSealsLoadAttempted(t *testing.T) {
	rep := reportOf(t, resultForVerdict(nil))
	if !rep.Measured.LoadAttempted {
		t.Fatal("measured.load_attempted perdu alors que Result.LoadAttempted est vrai")
	}
	if rep.Measured.LoadMeasured != true {
		t.Fatal("la charge établie doit rester le fait dérivé, pas un copié-collé de load_attempted")
	}
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"load_attempted":true`) {
		t.Fatalf("load_attempted absent du JSON scellé :\n%s", b)
	}
}

// Un rapport antérieur au moteur de verdict (artifact relu depuis le disque,
// Report construit à la main) ne porte pas d'operator_verdict : le Markdown
// doit nommer l'absence, jamais imprimer un verdict vide qui se lirait
// « rien à signaler ».
func TestReportMarkdownWithoutVerdictNamesTheAbsence(t *testing.T) {
	md := (&Report{AuditID: "audit-legacy", Grade: "C", Verdict: "queue sous charge"}).RenderMarkdown()
	if !strings.Contains(md, "**Verdict opérateur : non calculé**") {
		t.Fatalf("verdict absent rendu sans dire l'absence :\n%s", md)
	}
	if strings.Contains(md, "Verdict opérateur : **") {
		t.Fatalf("ligne de verdict à mot vide :\n%s", md)
	}
}
