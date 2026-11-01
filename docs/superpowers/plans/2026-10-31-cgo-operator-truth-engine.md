# Operator truth engine (slice 20A) implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Put every number the operator console shows behind a Go-computed fact — verdict, remediation rates, economics basis, live timeline, duel pair, frozen tariff.

**Architecture:** One pure verdict function in `pkg/metrics` feeds the sealed audit report; `pkg/audit` records what it measured and never guesses; `pkg/remediate` names the interfaces and rates it used; `pkg/api` exposes two new read endpoints (`/api/duel`, an extended `/api/audit/live`) and one write guard. The frontend consumes these fields in slice 20B and computes nothing.

**Tech Stack:** Go 1.25 standard library, `encoding/json`, `net/http` mux patterns already in `pkg/api`, `os` + `path/filepath` fixtures in tests.

**Spec:** `docs/superpowers/specs/2026-10-31-cgo-operator-console-design.md` §3.2–§3.7, §6, §8.

## Global Constraints

- No new module dependency. Standard library only.
- French comments and French user-facing strings; English identifiers.
- Never modify `book/data/runs`, the frozen corpus, or the column order of `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`. New JSON fields are allowed; new CSV columns are not.
- Thresholds are fixed values, copied verbatim: CAP deadline `220 ms`, breach threshold `95 %` (`pkg/metrics/cost.go:81-82`); Waveform grade bands `<5 A+`, `<30 A`, `<60 B`, `<200 C`, `<400 D`, `≥400 F` (`pkg/metrics/bufferbloat.go:11`); shaper headroom `0.95` (`pkg/remediate/remediate.go:56`).
- No green verdict without a grade. No remediation command without measured capacity and idle RTT. An absent value is a word, never a zero.
- Never `git add -A`; stage explicit paths. Never commit `book/` changes, `*.docx`, `*.pdf`, `cgo.exe`, archives.
- No live bench run, no audit against a production host. Tests use `t.TempDir()` fixtures only.
- The builder does not certify: new `docs/VERIFY.md` rows land as builder-pass with an independent verifier pending.
- Every task ends with `go build ./... && go test ./pkg/<touched> -count=1` green, then a commit.

---

### Task 1: Verdict function

**Files:**
- Create: `pkg/metrics/verdict.go`
- Create: `pkg/metrics/verdict_test.go`

**Interfaces:**
- Consumes: `metrics.CAPSLA` (`pkg/metrics/cost.go:86-91`), `metrics.CAPDeadlineMs`, `metrics.CAPBreachThresholdPct`.
- Produces: `metrics.VerdictLevel`, `metrics.VerdictNeutral|VerdictRed|VerdictAmber|VerdictGreen`, `metrics.VerdictInput{LoadAttempted, LoadEstablished bool; Grade string; SLA CAPSLA}`, `metrics.AssessVerdict(VerdictInput) Verdict`, `metrics.Verdict{Level, Word, Sentence}`. Slice 20A Task 2 and slice 20B read these names exactly.

- [ ] **Step 1: Write the failing test**

Create `pkg/metrics/verdict_test.go`:

```go
package metrics

import "testing"

func sla(pct float64, measured bool) CAPSLA {
	return CAPSLA{Measured: measured, OKPct: pct, Breach: measured && pct < CAPBreachThresholdPct}
}

func TestAssessVerdictCascade(t *testing.T) {
	cases := []struct {
		name  string
		in    VerdictInput
		level VerdictLevel
		word  string
	}{
		{"jamais testé", VerdictInput{}, VerdictNeutral, "Pas encore mesuré"},
		{"charge non établie", VerdictInput{LoadAttempted: true}, VerdictNeutral, "Charge non établie"},
		{"alertes en retard", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "B", SLA: sla(60.2, true)}, VerdictRed, "Alertes en retard"},
		{"failure prime sur note absente", VerdictInput{LoadAttempted: true, LoadEstablished: true, SLA: sla(80, true)}, VerdictRed, "Alertes en retard"},
		{"note absente", VerdictInput{LoadAttempted: true, LoadEstablished: true}, VerdictNeutral, "Note indisponible"},
		{"saturation sans sonde alerte", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "F"}, VerdictRed, "Saturation critique"},
		{"fragile", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "C"}, VerdictAmber, "Lien fragile"},
		{"sain", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "A", SLA: sla(97.5, true)}, VerdictGreen, "Lien sain"},
		{"sain sans sonde alerte", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "B"}, VerdictGreen, "Lien sain"},
		{"note hors référentiel", VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "b"}, VerdictNeutral, "Note indisponible"},
		{"charge tentée prime sur manquement", VerdictInput{LoadAttempted: true, LoadEstablished: false, SLA: sla(60.2, true)}, VerdictNeutral, "Charge non établie"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := AssessVerdict(c.in)
			if got.Level != c.level || got.Word != c.word {
				t.Fatalf("verdict = %q/%q, voulu %q/%q", got.Level, got.Word, c.level, c.word)
			}
			if got.Sentence == "" {
				t.Fatal("phrase vide")
			}
		})
	}
}

// Le piège signalé en revue : une sonde d'alerte qui passe alors qu'aucune
// charge n'a été établie ne doit jamais produire de vert.
func TestVerdictNeverGreenWithoutGrade(t *testing.T) {
	got := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, SLA: sla(100, true)})
	if got.Level == VerdictGreen {
		t.Fatalf("vert sans note : %+v", got)
	}
	if got.Level != VerdictNeutral {
		t.Fatalf("niveau = %q, voulu neutral", got.Level)
	}
}

func TestVerdictSentencesCarryTheirNumbers(t *testing.T) {
	red := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "D", SLA: sla(60.2, true)})
	if !strings.Contains(red.Sentence, "60.2 %") {
		t.Fatalf("phrase rouge %q doit citer le pourcentage mesuré", red.Sentence)
	}
	crit := AssessVerdict(VerdictInput{LoadAttempted: true, LoadEstablished: true, Grade: "F"})
	if !strings.Contains(crit.Sentence, "note F") {
		t.Fatalf("phrase saturation %q doit nommer la note", crit.Sentence)
	}
}
```

Add `"strings"` to the test file's imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/metrics -run TestAssessVerdict -v`
Expected: FAIL — `undefined: AssessVerdict` (and `VerdictInput`).

- [ ] **Step 3: Write the implementation**

Create `pkg/metrics/verdict.go`:

```go
package metrics

import "fmt"

// VerdictLevel — les quatre niveaux du langage opérateur. Le rouge garde un
// seul niveau et deux mots, selon ce qu'il a mesuré : une mission manquée et
// une saturation sont deux urgences différentes, pas deux couleurs.
type VerdictLevel string

const (
	VerdictNeutral VerdictLevel = "neutral"
	VerdictRed     VerdictLevel = "red"
	VerdictAmber   VerdictLevel = "amber"
	VerdictGreen   VerdictLevel = "green"
)

// VerdictInput — les seuls faits que la cascade lit. Rien n'est recalculé ici :
// la note vient de BufferbloatGrade, la conformité de AssessCAPSLA.
type VerdictInput struct {
	LoadAttempted   bool
	LoadEstablished bool
	Grade           string
	SLA             CAPSLA
}

type Verdict struct {
	Level    VerdictLevel `json:"level"`
	Word     string       `json:"word"`
	Sentence string       `json:"sentence"`
}

// AssessVerdict — cascade à premier gagnant, bornée par les seuils déjà
// défendus par le moteur. Un vert sans note est impossible : la charge peut
// avoir fait passer des sondes sans jamais être établie.
func AssessVerdict(in VerdictInput) Verdict {
	switch {
	case !in.LoadAttempted:
		return Verdict{VerdictNeutral, "Pas encore mesuré",
			"Aucun test sous charge n'a été réalisé sur ce lien."}
	case !in.LoadEstablished:
		return Verdict{VerdictNeutral, "Charge non établie",
			"Le flux saturant n'a pas pu démarrer ; la tenue sous congestion n'est pas mesurée."}
	case in.SLA.Measured && in.SLA.Breach:
		return Verdict{VerdictRed, "Alertes en retard",
			fmt.Sprintf("%.1f %% des alertes passent sous %d ms (objectif %.0f %%).",
				in.SLA.OKPct, CAPDeadlineMs, CAPBreachThresholdPct)}
	case in.Grade == "":
		return Verdict{VerdictNeutral, "Note indisponible",
			"La charge a tourné mais l'écart de latence n'a pas pu être calculé."}
	case in.Grade == "D" || in.Grade == "F":
		return Verdict{VerdictRed, "Saturation critique",
			fmt.Sprintf("Le lien sature sous charge (note %s) ; les télétransmissions risquent d'échouer.", in.Grade)}
	case in.Grade == "C":
		return Verdict{VerdictAmber, "Lien fragile",
			"La latence monte nettement sous charge ; les alertes peuvent partir en retard."}
	case (in.Grade == "A+" || in.Grade == "A" || in.Grade == "B") && in.SLA.Measured:
		return Verdict{VerdictGreen, "Lien sain",
			fmt.Sprintf("Sous charge, la latence reste sous l'objectif et %.1f %% des alertes passent à l'heure.", in.SLA.OKPct)}
	case in.Grade == "A+" || in.Grade == "A" || in.Grade == "B":
		return Verdict{VerdictGreen, "Lien sain",
			"Sous charge, la latence reste sous l'objectif."}
	default:
		// Vocabulaire de note inconnu : rien à classer. Jamais un vert par
		// défaut — la porte 7 de la cascade n'accepte que A+, A et B.
		return Verdict{VerdictNeutral, "Note indisponible",
			"La charge a tourné mais l'écart de latence n'a pas pu être calculé."}
	}
}
```

- [ ] **Step 4: Run tests and make sure they pass**

Run: `go test ./pkg/metrics -count=1`
Expected: `ok github.com/Realms4239/cgo/pkg/metrics`

- [ ] **Step 5: Commit**

```bash
git add pkg/metrics/verdict.go pkg/metrics/verdict_test.go
git commit -m "metrics: le verdict opérateur vient d'une cascade à premier gagnant"
```

---

### Task 2: Verdict and `load_attempted` in the sealed report

**Files:**
- Modify: `pkg/audit/audit.go` (Result literal near `BloatGrade: grade`, line ~625)
- Modify: `pkg/audit/report.go` (`Measured` struct line 114, `Report` struct line 23, `BuildReport` line ~203, `RenderMarkdown` grade line ~509)
- Create: `pkg/audit/verdict_report_test.go`

**Interfaces:**
- Consumes: `metrics.AssessVerdict`, `metrics.Verdict` from Task 1.
- Produces: `audit.Result.LoadAttempted bool` (`json:"load_attempted"`), `audit.Measured.LoadAttempted` (`json:"load_attempted"`), `audit.Report.OperatorVerdict metrics.Verdict` (`json:"operator_verdict"`). Slice 20B reads `report.operator_verdict.{level,word,sentence}`.

- [ ] **Step 1: Write the failing test**

Create `pkg/audit/verdict_report_test.go`:

```go
package audit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/remediate"
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
```

The three tests then read:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit -run TestReportCarriesOperatorVerdict -v`
Expected: FAIL — `unknown field LoadAttempted` / `rep.OperatorVerdict undefined`.

- [ ] **Step 3: Add `LoadAttempted` to the result and freeze it in the report**

In `pkg/audit/audit.go`, add to the `Result` struct, next to `LoadMeasured`'s neighbours:

```go
	// LoadAttempted — une fenêtre de charge était programmée. Distinct de
	// LoadMeasured : un puits injoignable tente la charge et ne la mesure pas.
	LoadAttempted bool `json:"load_attempted"`
```

In the `Result` literal where `BloatGrade: grade` is set (line ~625), add:

```go
		LoadAttempted: withLoad,
```

In `pkg/audit/report.go`, add to `Measured`:

```go
	// LoadAttempted — la fenêtre de charge était-elle prévue ? Sans ce fait,
	// un audit sans puits se lirait « jamais testé ».
	LoadAttempted bool `json:"load_attempted"`
```

Add to `Report`, after `Verdict string`:

```go
	// OperatorVerdict — le mot que lit un opérateur qui ne connaît pas la
	// méthode. Dérivé des faits déjà gelés, jamais d'une moyenne inventée.
	OperatorVerdict metrics.Verdict `json:"operator_verdict"`
```

- [ ] **Step 4: Compute it in `BuildReport` and print it in Markdown**

In `BuildReport`, immediately after the block that sets `rep.CAPSLA` (line ~291):

```go
	rep.OperatorVerdict = metrics.AssessVerdict(metrics.VerdictInput{
		LoadAttempted:   res.LoadAttempted,
		LoadEstablished: rep.Measured.LoadMeasured,
		Grade:           rep.Grade,
		SLA:             rep.CAPSLA,
	})
```

Where `rep.Measured` is filled, add `LoadAttempted: res.LoadAttempted,`.

In `RenderMarkdown`, directly after the existing `**Bufferbloat grade: …**` lines (line ~509-511):

```go
	b.WriteString("**Verdict opérateur : " + r.OperatorVerdict.Word + "** — " + r.OperatorVerdict.Sentence + "\n\n")
```

- [ ] **Step 5: Run the package tests**

Run: `go test ./pkg/audit ./pkg/metrics -count=1`
Expected: `ok` for both. If an existing honesty test asserts the exact Markdown line order, update only the expectation string, not the renderer's order.

- [ ] **Step 6: Prove the sealed artifact still verifies**

Run: `python tools/verify.py --fast`
Expected: `frozen`, `provenance`, `hygiene`, `links` pass. The frozen corpus is not rewritten; the new field appears only in reports built from now on.

- [ ] **Step 7: Commit**

```bash
git add pkg/audit/audit.go pkg/audit/report.go pkg/audit/verdict_report_test.go
git commit -m "audit: le rapport scellé porte le verdict opérateur et la tentative de charge"
```

---

### Task 3: Remediation rates, interfaces, and no half-script

**Files:**
- Modify: `pkg/remediate/remediate.go` (`RouterOSIngress` line 310, `Validate` line 66, add `ShapedRate`)
- Modify: `pkg/remediate/remediate_test.go`
- Modify: `pkg/audit/report.go` (`Remediation` struct line 186, build block lines 337-400)
- Create: `pkg/audit/remediation_fields_test.go`

**Interfaces:**
- Consumes: `remediate.Link`, `remediate.shaperHeadroom`.
- Produces: `remediate.ShapedRate(mbps float64) float64`, `remediate.DefaultRouterOSLan = "bridge"`, `(*Link).RouterOSLan() (string, bool)`; `audit.Remediation.{UpMeasuredMbps,UpShapedMbps,DownMeasuredMbps,DownShapedMbps,LinuxIface,RouterOSWanIface,RouterOSLanIface,RouterOSLanAssumed}`. Slice 20B's card reads these names.

- [ ] **Step 1: Write the failing tests**

Append to `pkg/remediate/remediate_test.go`:

```go
func TestRouterOSIngressAssumesBridgeAndSaysSo(t *testing.T) {
	l := Link{RouterOSInterface: "ether1-wan", CapacityMbps: 18.5, RTTms: 100}
	script, err := RouterOSIngress(l, QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "parent=bridge") {
		t.Fatalf("script sans interface par défaut :\n%s", script)
	}
	if !strings.Contains(script, "ATTENTION") {
		t.Fatalf("l'avertissement d'interface supposée est absent :\n%s", script)
	}
}

func TestRouterOSIngressUsesDeclaredLanInterface(t *testing.T) {
	l := Link{RouterOSInterface: "ether1-wan", RouterOSLanInterface: "lan-bridge", CapacityMbps: 18.5, RTTms: 100}
	script, err := RouterOSIngress(l, QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "parent=lan-bridge") {
		t.Fatalf("interface déclarée ignorée :\n%s", script)
	}
	if strings.Contains(script, "ATTENTION") {
		t.Fatal("aucune hypothèse à signaler quand l'interface est déclarée")
	}
}

func TestShapedRateMatchesHeadroom(t *testing.T) {
	if got := ShapedRate(18.5); math.Abs(got-17.575) > 1e-9 {
		t.Fatalf("ShapedRate(18.5) = %v, voulu 17.575", got)
	}
}
```

Add `import ("math"; "strings")` to the test file if absent.

- [ ] **Step 2: Run to verify failure**

Run: `go test ./pkg/remediate -run "RouterOSIngress|ShapedRate" -v`
Expected: FAIL — `RouterOSIngress` returns `ErrUnmeasured` for the empty LAN interface, and `ShapedRate` is undefined.

- [ ] **Step 3: Implement the assumed interface and the exported rate**

In `pkg/remediate/remediate.go`, add after `IngressIFB`:

```go
// DefaultRouterOSLan — l'interface LAN que RouterOS crée sur la plupart des
// postes RB. Nommée explicitement comme hypothèse dans le script, jamais
// présentée comme une mesure : une remédiation absente est muette, une
// remédiation incomplète est dangereuse.
const DefaultRouterOSLan = "bridge"

// RouterOSLan rend l'interface LAN à cibler et dit si elle est supposée.
func (l Link) RouterOSLan() (string, bool) {
	if l.RouterOSLanInterface != "" {
		return l.RouterOSLanInterface, false
	}
	return DefaultRouterOSLan, true
}

// ShapedRate expose le calibrage du shaper (95 % du goulot mesuré) à
// l'UI et au rapport, pour qu'aucun client ne recopie la constante.
func ShapedRate(measuredMbps float64) float64 { return measuredMbps * shaperHeadroom }
```

Replace the head of `RouterOSIngress` (lines 310-318):

```go
func RouterOSIngress(l Link, qdisc string) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}
	iface, assumed := l.RouterOSLan()
	bits := int64(math.Round(l.shapedRate() * 1000000))

	var b strings.Builder
	b.WriteString("# CGO download remediation - RouterOS v7+\n")
	if assumed {
		b.WriteString("# ATTENTION : interface LAN supposée (« " + iface + " »).\n")
		b.WriteString("# Vérifiez avec /interface print brief et remplacez si nécessaire.\n")
		b.WriteString("# Sans cette étape, la réception n'est pas façonnée.\n")
	}
	b.WriteString("# RouterOS has no IFB. This is a queue tree on the LAN/private\n")
	b.WriteString("# interface, which is where download traffic egresses toward clients.\n")
	b.WriteString("# It shapes the same direction at a different point than the Linux IFB.\n")
	b.WriteString("# Requires RouterOS >= 7.10 for kind=cake. Verify with: /system resource print\n")
```

Keep the rest of the function unchanged; it already writes `parent=` + `iface`.

- [ ] **Step 4: Run the remediate tests**

Run: `go test ./pkg/remediate -count=1`
Expected: `ok`. A test that asserted the old `ErrUnmeasured` for an empty LAN interface must be replaced by `TestRouterOSIngressAssumesBridgeAndSaysSo`; delete only that assertion, not the file.

- [ ] **Step 5: Write the report-field test**

Create `pkg/audit/remediation_fields_test.go`:

```go
package audit

import (
	"math"
	"testing"

	"github.com/Realms4239/cgo/pkg/remediate"
)

func TestReportRemediationCarriesRatesAndInterfaces(t *testing.T) {
	res := &Result{
		AuditID: "audit-rem", Site: "Siège", LinkType: "4g",
		RTTIdleP50: 100, RTTIdleP95: 110,
		LoadAttempted: true, LoadMeasured: true, Direction: "both",
		BloatDeltaMs: 12, BloatGrade: "A",
		ThroughputMbps: 5.1, ThroughputDownMbps: 18.5, LoadSeconds: 40,
	}
	rep, err := BuildReport(res, remediate.Link{
		Interface: "eth0", RouterOSInterface: "ether1-wan",
	}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	r := rep.Remediation
	if r == nil {
		t.Fatal("remédiation absente alors que les deux sens sont mesurés")
	}
	if math.Abs(r.UpMeasuredMbps-5.1) > 1e-9 || math.Abs(r.DownMeasuredMbps-18.5) > 1e-9 {
		t.Fatalf("débits mesurés = %v / %v", r.UpMeasuredMbps, r.DownMeasuredMbps)
	}
	if math.Abs(r.UpShapedMbps-remediate.ShapedRate(5.1)) > 1e-9 {
		t.Fatalf("débit façonné montant = %v", r.UpShapedMbps)
	}
	if r.LinuxIface != "eth0" || r.RouterOSWanIface != "ether1-wan" {
		t.Fatalf("interfaces = %q / %q", r.LinuxIface, r.RouterOSWanIface)
	}
	if r.RouterOSLanIface != remediate.DefaultRouterOSLan || !r.RouterOSLanAssumed {
		t.Fatalf("interface LAN = %q assumée=%v", r.RouterOSLanIface, r.RouterOSLanAssumed)
	}
}

func TestReportRemediationOmitsUnmeasuredDirection(t *testing.T) {
	res := &Result{
		AuditID: "audit-up", Site: "Siège", LinkType: "4g",
		RTTIdleP50: 100, LoadAttempted: true, LoadMeasured: true, Direction: "up",
		BloatDeltaMs: 12, BloatGrade: "A", ThroughputMbps: 5.1, LoadSeconds: 40,
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0", RouterOSLanInterface: "lan"}, "cake", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Remediation.DownMeasuredMbps != 0 || len(rep.Remediation.LinuxTCIngress) != 0 {
		t.Fatalf("script descendant inventé : %+v", rep.Remediation)
	}
}
```

- [ ] **Step 6: Run to verify failure**

Run: `go test ./pkg/audit -run TestReportRemediation -v`
Expected: FAIL — unknown fields on `Remediation`.

- [ ] **Step 7: Add the fields and fill them**

In `pkg/audit/report.go`, extend `Remediation` (after `RouterOSIngress`):

```go
	// Rates and interfaces named by the measurement that produced them. The
	// UI reads these instead of parsing a command string — un taux recopié
	// depuis le texte d'un script dérive dès qu'on change la formule.
	UpMeasuredMbps   float64 `json:"up_measured_mbps,omitempty"`
	UpShapedMbps     float64 `json:"up_shaped_mbps,omitempty"`
	DownMeasuredMbps float64 `json:"down_measured_mbps,omitempty"`
	DownShapedMbps   float64 `json:"down_shaped_mbps,omitempty"`
	LinuxIface       string  `json:"linux_iface,omitempty"`
	RouterOSWanIface string  `json:"routeros_wan_iface,omitempty"`
	RouterOSLanIface string  `json:"routeros_lan_iface,omitempty"`
	RouterOSLanAssumed bool  `json:"routeros_lan_assumed,omitempty"`
```

In the remediation block of `BuildReport` (lines 337-400), after `rem := &Remediation{...}`:

```go
		rem.LinuxIface = link.Interface
		rem.RouterOSWanIface = link.RouterOSInterface
		lanIface, lanAssumed := link.RouterOSLan()
		rem.RouterOSLanIface, rem.RouterOSLanAssumed = lanIface, lanAssumed
		if upCap > 0 {
			rem.UpMeasuredMbps = upCap
			rem.UpShapedMbps = remediate.ShapedRate(upCap)
		}
		if downCap > 0 {
			rem.DownMeasuredMbps = downCap
			rem.DownShapedMbps = remediate.ShapedRate(downCap)
		}
```

The ingress script is now generated whenever `downCap > 0`, because `RouterOSIngress` no longer refuses an undeclared LAN interface. Keep the existing `else if res.Direction == "both"` unavailability records for a direction that never measured.

- [ ] **Step 8: Run tests and commit**

Run: `go test ./pkg/audit ./pkg/remediate -count=1`
Expected: `ok` for both.

```bash
git add pkg/remediate/remediate.go pkg/remediate/remediate_test.go pkg/audit/report.go pkg/audit/remediation_fields_test.go
git commit -m "remediate: script complet à interface supposée annoncée, taux et interfaces gelés dans le rapport"
```

---

### Task 4: Numeric economics basis

**Files:**
- Modify: `pkg/metrics/cost.go` (`EconomicsScenario` line 115, `BuildEconomics` line 157)
- Create: `pkg/metrics/economics_numbers_test.go`

**Interfaces:**
- Consumes: `EnterprisePlan{MonthlyOpExMGA, CIRMbps}`, `B2BTariff{ARPerGB}`, `RetailEmergencyARPerGB`.
- Produces: `metrics.EconomicsNumbers` and the field `metrics.EconomicsScenario.Numbers`.

- [ ] **Step 1: Write the failing test**

Create `pkg/metrics/economics_numbers_test.go`:

```go
package metrics

import "testing"

func TestBuildEconomicsNumbersPerModel(t *testing.T) {
	e := BuildEconomics("fiber", RetransRatio{Ratio: 0.0187, Basis: "measured"}, 1e9, 1.87e7, 40)
	if e.Primary == nil {
		t.Fatal("aucun scénario primaire")
	}
	var a, b, em *EconomicsScenario
	for i := range e.Scenarios {
		switch e.Scenarios[i].Model {
		case "A":
			if a == nil {
				a = &e.Scenarios[i]
			}
		case "B":
			if b == nil {
				b = &e.Scenarios[i]
			}
		case "emergency":
			em = &e.Scenarios[i]
		}
	}
	if a == nil || b == nil || em == nil {
		t.Fatalf("scénarios manquants : %+v", e.Scenarios)
	}
	if a.Numbers.MonthlyOpExMGA <= 0 || a.Numbers.CIRMbps <= 0 {
		t.Fatalf("modèle A sans charge mensuelle ni CIR : %+v", a.Numbers)
	}
	if a.Numbers.ARPerGB != 0 {
		t.Fatalf("modèle A ne doit pas passer par l'arithmétique du volume : %+v", a.Numbers)
	}
	if b.Numbers.ARPerGB <= 0 || b.Numbers.MonthlyOpExMGA != 0 {
		t.Fatalf("modèle B : %+v", b.Numbers)
	}
	if em.Numbers.ARPerGB != RetailEmergencyARPerGB {
		t.Fatalf("pénalité de secours : %+v", em.Numbers)
	}
	if a.Numbers.RetransRatio != 0.0187 || a.Numbers.WastedBytes != 1.87e7 {
		t.Fatalf("base de gaspillage absente : %+v", a.Numbers)
	}
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./pkg/metrics -run TestBuildEconomicsNumbers -v`
Expected: FAIL — `a.Numbers undefined`.

- [ ] **Step 3: Implement**

In `pkg/metrics/cost.go`, add after `EconomicsScenario`:

```go
// EconomicsNumbers — la base chiffrée d'un scénario, dans son propre modèle.
// Le texte `Basis` reste lisible par un humain ; l'interface ne parse plus une
// phrase française pour afficher un nombre.
type EconomicsNumbers struct {
	Model          string  `json:"model"`                      // "A" | "B" | "emergency"
	MonthlyOpExMGA float64 `json:"monthly_opex_mga,omitempty"` // A
	CIRMbps        float64 `json:"cir_mbps,omitempty"`         // A
	ARPerGB        float64 `json:"ar_per_gb,omitempty"`        // B et emergency
	RetransRatio   float64 `json:"retrans_ratio"`
	WastedBytes    float64 `json:"wasted_bytes"`
	TransportBytes float64 `json:"transport_bytes"`
	LoadSeconds    float64 `json:"load_seconds"`
}
```

and inside `EconomicsScenario`:

```go
	Numbers EconomicsNumbers `json:"numbers"`
```

Fill it in `BuildEconomics`: give the model-A loop `Numbers: numbers("A", p.MonthlyOpExMGA, p.CIRMbps, 0, ratio, transportedBytes, wastedBytes, loadSeconds)`, the model-B loop `numbers("B", 0, 0, tr.ARPerGB, ...)`, the emergency scenario `numbers("emergency", 0, 0, RetailEmergencyARPerGB, ...)`, with the helper:

```go
func numbers(model string, opex, cir, arPerGB, ratio float64, transported, wasted, loadSeconds float64) EconomicsNumbers {
	return EconomicsNumbers{
		Model: model, MonthlyOpExMGA: opex, CIRMbps: cir, ARPerGB: arPerGB,
		RetransRatio: ratio, WastedBytes: wasted, TransportBytes: transported,
		LoadSeconds: loadSeconds,
	}
}
```

Pass `ratio.Ratio` as the `ratio` argument. `e.Primary` is a copy of the chosen scenario, so it carries `Numbers` without extra work.

- [ ] **Step 4: Run tests and commit**

Run: `go test ./pkg/metrics ./pkg/audit -count=1`
Expected: `ok`.

```bash
git add pkg/metrics/cost.go pkg/metrics/economics_numbers_test.go
git commit -m "metrics: chaque scénario économique porte sa base chiffrée"
```

---

### Task 4b: Campaign-frozen tariff and the write guard

**Files:**
- Modify: `pkg/metrics/metrics.go` (add `TierByName` after `ActiveTier`, line 125)
- Modify: `pkg/campagne/config.go` (`MatrixConfig` line 14, `buildMatrixConfig` line 32)
- Modify: `pkg/campagne/campagne.go:429` and `:631`
- Modify: `pkg/api/server.go` (`POST /api/cost/tier`, line ~455-465)
- Create: `pkg/campagne/tier_freeze_test.go`; modify `pkg/api` tests

**Interfaces:**
- Consumes: `metrics.ActiveTier()`, `Deps.RunningFn`.
- Produces: `metrics.TierByName(name string) (PriceTier, bool)`; `MatrixConfig.Tier string` (`json:"tier"`); 409 from `POST /api/cost/tier` during a run.

- [ ] **Step 1: Write the failing tests**

Create `pkg/campagne/tier_freeze_test.go`:

```go
package campagne

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/metrics"
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
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./pkg/campagne -run "Tier|CostARPerH" -v`
Expected: FAIL — unknown `Tier` in config, `metrics.TierByName` and `costARPerH` undefined.

- [ ] **Step 3: Implement**

In `pkg/metrics/metrics.go`:

```go
// TierByName rend un palier à son nom exact, sans toucher au global.
func TierByName(name string) (PriceTier, bool) {
	tierMu.RLock()
	defer tierMu.RUnlock()
	for _, tr := range Tiers {
		if tr.Name == name {
			return tr, true
		}
	}
	return PriceTier{}, false
}
```

In `pkg/campagne/config.go`, add to the config struct:

```go
	Tier        string          `json:"tier"`
```

and in `buildMatrixConfig`, next to `Kernel: readKernelRelease()`:

```go
		Tier: ActiveTierName(),
```

`ActiveTierName` is already exported by `pkg/metrics`; import it as `metrics.ActiveTierName()` if this file is in package `campagne` and the struct is built with a `metrics.` prefix elsewhere.

In `pkg/campagne/campagne.go`, add the helper next to the other row-level helpers:

```go
// costARPerH — le coût d'une ligne se calcule au palier gelé avec son run.
// Lire le global à chaque événement donnerait deux bases dans un même run.
func costARPerH(tierName string, wastedBytes uint64) float64 {
	if tr, ok := metrics.TierByName(tierName); ok {
		return metrics.CostARPerHTier(wastedBytes, tr)
	}
	return metrics.CostARPerH(wastedBytes)
}
```

Replace both `metrics.CostARPerH(...)` call sites (lines 429 and 631) with `costARPerH(<frozen tier>, ...)`. At line 631 the tier is `m.Config.Tier` on the matrix receiver. At line 429 the live snapshot is built where the run config is already in scope; if it is not, thread the tier through the same struct that already carries `Bench` and `ChannelModel` rather than reading the global.

In `pkg/api/server.go`, inside the `POST /api/cost/tier` handler before `metrics.SetDefaultTier`:

```go
		if d.RunningFn != nil && d.RunningFn() {
			writeErr(w, r, "campagne en cours — le palier est gelé avec le run, changez-le après le gel", http.StatusConflict)
			return
		}
```

- [ ] **Step 4: Add the API test**

Append to `pkg/api` (in `capabilities_test.go` or a new `cost_tier_guard_test.go`):

```go
func TestCostTierRefusedWhileCampaignRuns(t *testing.T) {
	srv := httptest.NewServer(New(Deps{RunningFn: func() bool { return true }}))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/cost/tier", "application/json",
		strings.NewReader(`{"tier":"yas-day-1gb"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, voulu 409", resp.StatusCode)
	}
}
```

- [ ] **Step 5: Run tests and commit**

Run: `go test ./pkg/metrics ./pkg/campagne ./pkg/api -count=1`
Expected: `ok` for all three.

```bash
git add pkg/metrics/metrics.go pkg/campagne/config.go pkg/campagne/campagne.go pkg/campagne/tier_freeze_test.go pkg/api
git commit -m "campagne: le palier est gelé avec le run et protégé contre un changement en cours"
```

---

### Task 5: Full-run live timeline

**Files:**
- Create: `pkg/audit/timeline.go`
- Create: `pkg/audit/timeline_test.go`
- Modify: `pkg/audit/audit.go` (`Progress` line 138, probe append sites, `OnProgress` line 280)
- Modify: `pkg/api/server_audit.go` (live handler, line 257-259)
- Modify: `pkg/api` tests

**Interfaces:**
- Consumes: `metrics.Summarize` (`pkg/metrics/metrics.go:29-40`).
- Produces: `audit.TimelinePoint{TPtr int, Phase string, N int, RTTP50, RTTP95, SmallP95 *float64}` with JSON keys `t`, `phase`, `n`, `rtt_p50`, `rtt_p95`, `small_p95`; `audit.Progress.Timeline []TimelinePoint` (`json:"timeline"`).

- [ ] **Step 1: Write the failing test**

Create `pkg/audit/timeline_test.go`:

```go
package audit

import "testing"

func TestTimelineBucketsBySecondAndKeepsGaps(t *testing.T) {
	tl := newTimeline()
	// 3 s de repos, 2 s de charge sans réponse HTTP, puis une réponse.
	for i := 0; i < 15; i++ {
		tl.add(i/5, "baseline", 24+float64(i), 0, false)
	}
	tl.add(3, "charge", 300, 0, false)
	tl.add(4, "charge", 320, 900, true)

	got := tl.points()
	if len(got) != 5 {
		t.Fatalf("points = %d, voulu 5 (%+v)", len(got), got)
	}
	if got[0].N != 5 {
		t.Fatalf("n seconde 0 = %d, voulu 5", got[0].N)
	}
	if got[3].SmallP95 != nil {
		t.Fatal("une sonde muette ne doit pas devenir une mesure")
	}
	if got[4].SmallP95 == nil || *got[4].SmallP95 != 900 {
		t.Fatalf("small_p95 = %v, voulu 900", got[4].SmallP95)
	}
	if got[3].RTTP95 == nil {
		t.Fatal("le RTT de la seconde 3 a pourtant répondu")
	}
}

func TestTimelineCapIsTotalPlusFive(t *testing.T) {
	tl := newTimeline()
	for i := 0; i < 200; i++ {
		tl.add(i, "charge", 50, 0, false)
	}
	if len(tl.points()) > 65 {
		t.Fatalf("pas de garde de longueur : %d points", len(tl.points()))
	}
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./pkg/audit -run Timeline -v`
Expected: FAIL — `undefined: newTimeline`.

- [ ] **Step 3: Implement the accumulator**

Create `pkg/audit/timeline.go`:

```go
package audit

import "sort"

// TimelinePoint — une seconde de la fenêtre d'audit. Les valeurs sont pointeurs
// : nil veut dire « aucune sonde n'a répondu », ce qui n'est pas un 0 ms.
type TimelinePoint struct {
	T        int      `json:"t"`
	Phase    string   `json:"phase"`
	N        int      `json:"n"`
	RTTP50   *float64 `json:"rtt_p50"`
	RTTP95   *float64 `json:"rtt_p95"`
	SmallP95 *float64 `json:"small_p95"`
}

const timelineCapSec = 65

type secBucket struct {
	phase  string
	rtts   []float64
	smalls []float64
}

type timeline struct {
	buckets map[int]*secBucket
	max     int
}

func newTimeline() *timeline { return &timeline{buckets: map[int]*secBucket{}} }

// add enregistre une sonde. smallMs <= 0 ou smallOK faux = sonde muette.
func (t *timeline) add(sec int, phase string, rttMs, smallMs float64, smallOK bool) {
	if sec < 0 {
		return
	}
	if sec >= timelineCapSec {
		return
	}
	b := t.buckets[sec]
	if b == nil {
		b = &secBucket{}
		t.buckets[sec] = b
	}
	if b.phase == "" {
		b.phase = phase
	}
	if rttMs > 0 {
		b.rtts = append(b.rtts, rttMs)
	}
	if smallOK && smallMs > 0 {
		b.smalls = append(b.smalls, smallMs)
	}
	if sec > t.max {
		t.max = sec
	}
}

func (t *timeline) points() []TimelinePoint {
	secs := make([]int, 0, len(t.buckets))
	for s := range t.buckets {
		secs = append(secs, s)
	}
	sort.Ints(secs)
	out := make([]TimelinePoint, 0, len(secs))
	for _, s := range secs {
		b := t.buckets[s]
		p := TimelinePoint{T: s, Phase: b.phase, N: len(b.rtts)}
		if len(b.rtts) > 0 {
			sum := summarize(b.rtts)
			p.RTTP50, p.RTTP95 = &sum.P50, &sum.P95
		}
		if len(b.smalls) > 0 {
			sum := summarize(b.smalls)
			p.SmallP95 = &sum.P95
		}
		out = append(out, p)
	}
	return out
}
```

`summarize` is `metrics.Summarize`; alias it at the top of the file (`func summarize(xs []float64) metrics.Stats { return metrics.Summarize(xs) }`) using the real return type from `pkg/metrics/metrics.go:29`.

- [ ] **Step 4: Feed it and expose it**

In `pkg/audit/audit.go`, next to the `rtt` and `small` slices in the probe loop, create `tl := newTimeline()`. At every site that appends a ping sample or a small-object completion, also call:

```go
	tl.add(int(time.Since(started).Seconds()), phase, rttMs, smallMs, smallAnswered)
```

Locate the sites with `grep -n "rtt = append\|small = append" pkg/audit/audit.go` and update each. Add to `Progress`:

```go
	// Timeline — une seconde par point sur toute la fenêtre, pour que
	// l'observateur voie la marches du repos vers la charge au lieu des 32
	// dernières sondes.
	Timeline []TimelinePoint `json:"timeline"`
```

and in the `OnProgress(Progress{…})` literal (line ~280) `Timeline: tl.points(),`.

In `pkg/api/server_audit.go`, add to the live payload map:

```go
			"timeline":         pr.Timeline,
```

- [ ] **Step 5: Add the API test**

```go
func TestAuditLiveCarriesTimeline(t *testing.T) {
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/audit/live")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["running"]; !ok {
		t.Fatal("contrat live rompu")
	}
}
```

This asserts the handler still answers when no audit runs; the timeline itself is covered by the `pkg/audit` unit test, because a live audit needs a bench and is not run in tests.

- [ ] **Step 6: Run tests and commit**

Run: `go test ./pkg/audit ./pkg/api -count=1`
Expected: `ok`.

```bash
git add pkg/audit/timeline.go pkg/audit/timeline_test.go pkg/audit/audit.go pkg/api/server_audit.go pkg/api
git commit -m "audit: la fenêtre complète devient une série par seconde pour l'observateur"
```

---

### Task 6: Duel endpoint

**Files:**
- Modify: `pkg/results/results.go` (`Group` line 12, `Scan` bucket loop)
- Create: `pkg/api/duel.go`
- Create: `pkg/api/duel_test.go`

**Interfaces:**
- Consumes: `results.Scan(dataDir, runFilter) ([]Group, error)`, `Group.DeadlineValidN`, `Group.DeadlineValidMedian`, `Group.Smallp95ValidMedian`, `Group.Qdisc`, `Group.CC`, `Group.Direction`.
- Produces: `GET /api/duel?profile=P2&run=<id>` → `duelResponse{Profile, DeadlineMs, MinReps, Base *duelCell, Best *duelCell, DeltaDeadlinePts, DeltaSmallMs, Reason}` with `duelCell{Qdisc, CC, DeadlinePct, SmallP95Ms, N, Runs}`.

- [ ] **Step 1: Record the runs per group**

In `pkg/results/results.go`, add to `Group`:

```go
	// Runs — les identifiants qui ont alimenté la cellule. Une médiane sans
	// provenance n'est pas une preuve.
	Runs []string `json:"runs,omitempty"`
```

In `Scan`, extend the bucket struct with `runs map[string]bool`, initialise it per bucket, and add `b.runs[filepath.Base(filepath.Dir(f))] = true` where each file's rows are folded in. Emit a sorted `Runs` slice when building each `Group`.

- [ ] **Step 2: Write the failing test**

Create `pkg/api/duel_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const duelHeader = "run_id,event_id,profile,qdisc,cc,repetition,rtt_p50_ms,rtt_p95_ms,qdi_ms,small_p95_ms,deadline_ok_pct,bulk_goodput_mbps,drops,retransmissions,wasted_bytes,cost_ar_per_h,cpu_pct,gate_status"

func writeDuelRun(t *testing.T, dir, run, qdisc, cc string, reps int, small, deadline float64, gate string) {
	t.Helper()
	runDir := filepath.Join(dir, run)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := duelHeader + "\n"
	for i := 1; i <= reps; i++ {
		body += run + "," + itoa(i) + ",P2," + qdisc + "," + cc + "," + itoa(i) +
			",20,25,3," + ftoa(small) + "," + ftoa(deadline) + ",18,0,0,1000,10,30," + gate + "\n"
	}
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func itoa(i int) string     { return strconv.Itoa(i) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

func TestDuelPicksBestAndBaseWithGuardrail(t *testing.T) {
	dir := t.TempDir()
	writeDuelRun(t, dir, "run-a", "pfifo_fast", "bbr", 3, 328.2, 60.2, "valid")
	writeDuelRun(t, dir, "run-b", "fq_codel", "bbr", 3, 205.4, 96.2, "valid")
	writeDuelRun(t, dir, "run-c", "cake", "bbr", 1, 100, 100, "valid") // n=1 : exclu

	old := duelDir
	duelDir = dir
	t.Cleanup(func() { duelDir = old })

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/duel?profile=P2")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got duelResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Base == nil || got.Base.Qdisc != "pfifo_fast" {
		t.Fatalf("base = %+v", got.Base)
	}
	if got.Best == nil || got.Best.Qdisc != "fq_codel" {
		t.Fatalf("la cellule n=1 doit rester inéligible : %+v", got.Best)
	}
	if got.Best.N < 3 || got.MinReps != 3 {
		t.Fatalf("garde d'échantillonnage : %+v", got)
	}
	if len(got.Best.Runs) == 0 {
		t.Fatal("le duel doit citer ses runs")
	}
}

func TestDuelRefusesUnknownProfile(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/duel?profile=../../etc")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, voulu 400", resp.StatusCode)
	}
}
```

Add `"strconv"` to the test file's imports.

- [ ] **Step 3: Run to verify failure**

Run: `go test ./pkg/api -run Duel -v`
Expected: FAIL — `duelDir`, `duelResponse`, route unknown.

- [ ] **Step 4: Implement the handler**

Create `pkg/api/duel.go`:

```go
package api

import (
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/results"
)

// duelDir — même source que /api/results ; variable pour que le test pointe
// sur un corpus jetable sans écrire dans data/runs.
var duelDir = "data/runs"

var duelProfile = regexp.MustCompile(`^P[0-9]$`)

// duelMinReps — une cellule d'une répétition n'est pas un gagnant de
// référence : l'historique du banc connaît des cellules isolées favorisées par
// l'ordonnanceur. Trois répétitions est le défaut de la matrice.
const duelMinReps = 3

type duelCell struct {
	Qdisc       string   `json:"qdisc"`
	CC          string   `json:"cc"`
	DeadlinePct float64  `json:"deadline_pct"`
	SmallP95Ms  float64  `json:"small_p95_ms"`
	N           int      `json:"n"`
	Runs        []string `json:"runs,omitempty"`
}

type duelResponse struct {
	Profile          string     `json:"profile"`
	DeadlineMs       int        `json:"deadline_ms"`
	MinReps          int        `json:"min_reps"`
	Base             *duelCell  `json:"base"`
	Best             *duelCell  `json:"best"`
	DeltaDeadlinePts float64    `json:"delta_deadline_pts"`
	DeltaSmallMs     float64    `json:"delta_small_ms"`
	Reason           string     `json:"reason,omitempty"`
}

func qdiscRank(q string) int {
	switch q {
	case "pfifo_fast":
		return 0
	case "fq_codel":
		return 1
	case "cake":
		return 2
	}
	return 9
}

func eligible(g results.Group) bool {
	return g.Direction == "up" && g.DeadlineValidN >= duelMinReps
}

func handleDuel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile := strings.TrimSpace(r.URL.Query().Get("profile"))
		if profile == "" {
			profile = "P2"
		}
		if !duelProfile.MatchString(profile) {
			writeErr(w, r, "profil invalide", http.StatusBadRequest)
			return
		}
		run := r.URL.Query().Get("run")
		if run != "" && !validRunID(run) {
			writeErr(w, r, "id de run invalide", http.StatusBadRequest)
			return
		}
		groups, _ := results.Scan(duelDir, run)
		var cells []results.Group
		for _, g := range groups {
			if g.Profile == profile && eligible(g) {
				cells = append(cells, g)
			}
		}
		resp := duelResponse{Profile: profile, DeadlineMs: metrics.CAPDeadlineMs, MinReps: duelMinReps}
		if len(cells) == 0 {
			resp.Reason = "aucune cellule valide avec au moins " + strconv.Itoa(duelMinReps) + " répétitions — relancez une campagne"
			writeJSON(w, resp)
			return
		}
		sort.Slice(cells, func(i, j int) bool {
			a, b := cells[i], cells[j]
			if a.DeadlineValidMedian != b.DeadlineValidMedian {
				return a.DeadlineValidMedian > b.DeadlineValidMedian
			}
			if a.Smallp95ValidMedian != b.Smallp95ValidMedian {
				return a.Smallp95ValidMedian < b.Smallp95ValidMedian
			}
			if qdiscRank(a.Qdisc) != qdiscRank(b.Qdisc) {
				return qdiscRank(a.Qdisc) < qdiscRank(b.Qdisc)
			}
			return a.CC < b.CC
		})
		best := cellOf(cells[0])
		var base *duelCell
		for _, g := range cells {
			if g.Qdisc == "pfifo_fast" && (base == nil || g.DeadlineValidN > base.N) {
				base = ptr(cellOf(g))
			}
		}
		resp.Best = ptr(best)
		resp.Base = base
		if base != nil {
			resp.DeltaDeadlinePts = round1(best.DeadlinePct - base.DeadlinePct)
			resp.DeltaSmallMs = round1(best.SmallP95Ms - base.SmallP95Ms)
		} else {
			resp.Reason = "aucune ligne de base pfifo_fast exploitable"
		}
		writeJSON(w, resp)
	}
}

func cellOf(g results.Group) duelCell {
	return duelCell{
		Qdisc: g.Qdisc, CC: g.CC,
		DeadlinePct: g.DeadlineValidMedian,
		SmallP95Ms:  g.Smallp95ValidMedian,
		N:           g.DeadlineValidN,
		Runs:        g.Runs,
	}
}

func ptr[T any](v T) *T { return &v }

// round1 — une décimale, arrondi arithmétique. `pkg/api` n'a pas d'arrondi
// réutilisable (vérifié avant d'écrire cette tâche).
func round1(v float64) float64 { return math.Round(v*10) / 10 }
```

Register it in `pkg/api/server.go` beside the other results routes:

```go
	mux.HandleFunc("GET /api/duel", handleDuel())
```

- [ ] **Step 5: Run tests and commit**

Run: `go test ./pkg/results ./pkg/api -count=1`
Expected: `ok`. If an existing results test asserts the exact `Group` JSON, add `runs` to its expectation rather than dropping the field.

```bash
git add pkg/results/results.go pkg/api/duel.go pkg/api/duel_test.go pkg/api/server.go
git commit -m "api: le duel de référence est choisi côté moteur, avec garde d'échantillonnage"
```

---

### Task 7: Close the slice

**Files:**
- Modify: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`, `docs/ERRORS.md`, `docs/superpowers/plans/README.md`

- [ ] **Step 1: Full gates**

Run: `python tools/verify.py`
Expected: `6 passed, 0 failed, 0 not run`. If a gate cannot run, report exit 2 as not green and name the gate.

- [ ] **Step 2: Verify rows (builder pass, verifier pending)**

Append to `docs/VERIFY.md`, one row per deliverable, each with the exact command run, the date, the host from `hostname`, and one line naming what stays unproven:

- Operator verdict cascade — `go test ./pkg/metrics ./pkg/audit -count=1` → ok; no live audit run.
- Remediation rates and assumed LAN interface — `go test ./pkg/remediate ./pkg/audit -count=1` → ok; no CHR booted, so “validée par rejeu” is unchanged.
- Economics numbers per scenario — `go test ./pkg/metrics -count=1` → ok.
- Campaign-frozen tier and 409 guard — `go test ./pkg/campagne ./pkg/api -count=1` → ok; no campaign run.
- Live per-second timeline — `go test ./pkg/audit ./pkg/api -count=1` → ok; never observed against a real audit.
- Duel endpoint with `n ≥ 3` guard — `go test ./pkg/api -run Duel -v` → ok; the frozen corpus is not read by the test.

- [ ] **Step 3: Error ledger entry**

Append to `docs/ERRORS.md` following the file's format, numbered after the last entry:

```markdown
## E10. A passing alert probe could have produced a green verdict (observed 2026-10-31)

- Symptom: with the bulk sink unreachable, `capSamples` still collects load-window
  completions, so `cap_sla` can read 100 % while no load was ever established.
- Cause: a verdict derived from the alert percentage alone would treat that as a
  pass. `load_measured` is the fact that says whether congestion existed.
- Fix: `metrics.AssessVerdict` gates on `LoadAttempted` and `LoadEstablished`
  before any green, and the report freezes `load_attempted`.
- Proof: `TestVerdictNeverGreenWithoutGrade`, `TestReportVerdictRefusesGreenWithoutLoad`.
- Lesson: a metric measured in the absence of its own stress is not a pass.
```

- [ ] **Step 4: Checkpoint and plan index**

Append a `docs/CHECKPOINTS.md` entry with branch, HEAD, changed files, commands with results, and the unproven list. Add the slice 20A row to `docs/superpowers/plans/README.md` pointing at this plan and the spec, status `In progress` until the verifier pass.

- [ ] **Step 5: Commit**

```bash
git add docs/VERIFY.md docs/CHECKPOINTS.md docs/ERRORS.md docs/superpowers/plans/README.md
git commit -m "docs: 20A est tracé, le verdict et le duel attendent leur vérificateur"
```

## Self-review notes

- Spec coverage: §3.2 → Tasks 1-2; §3.3 → Task 3; §3.6 → Tasks 4 and 4b; §3.5 → Task 5; §3.7 → Task 6; §6 invariants 1-5 → Tasks 1-6 tests; §8 Go list → matched task by task. §3.1, §3.4, §3.8-§3.10, §4, §5 are slice 20B.
- Field names `operator_verdict`, `load_attempted`, `up_shaped_mbps`, `routeros_lan_assumed`, `numbers`, `timeline`, `/api/duel` are the contract 20B consumes; renaming one here breaks a task there.
- Two call sites in `pkg/campagne/campagne.go` need reading before editing; the plan says so explicitly rather than guessing their shape.
