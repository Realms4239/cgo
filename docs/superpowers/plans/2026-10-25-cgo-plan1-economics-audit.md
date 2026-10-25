# CGO Enterprise Rebuild — Plan 1: Economics + Audit Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the enterprise valuation package (`pkg/metrics/cost.go`), the TCP_INFO retransmit sampler, and the audit rebuild (direction, multi-flow + JFI, CAP SLA, telemetry CSV with dual seals, simplified parameter surface, exit codes) — the report-schema foundation Plans 2–5 consume.

**Architecture:** Pure-Go packages, no new dependencies. `pkg/metrics` gains the financial models; `pkg/probe` gains the Linux TCP_INFO socket sampler and stats-returning bulk helpers; `pkg/audit` gains direction/flows/CAP-SLA/telemetry and the report v2 economics section; `cmd/cgo` collapses the audit flag surface with sink auto-detection. Frozen `link_audit.csv` schema stays immutable except two appended columns via the existing migration path.

**Tech Stack:** Go 1.25 stdlib only (`syscall` for TCP_INFO under a `linux` build tag), existing test style (stdlib `testing`, table tests, fake deps).

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-25-cgo-enterprise-rebuild-design.md` (§2 honesty invariants bind every task).
- Go module `github.com/Realms4239/cgo`, Go 1.25. **No new external dependencies.**
- Comments in French (repo convention), identifiers in English.
- Every command runs from repo root `C:\cgo` on Windows (Git Bash); `go build ./...` and `go test ./pkg/metrics/... ./pkg/probe/... ./pkg/audit/... ./cmd/cgo/...` must be green on Windows. Linux-only code lives behind `//go:build linux` with an `other` stub so Windows builds.
- Frozen-data guard: `book/data/runs`, `pkg/results/runs-frozen.tar.gz` untouched; `link_audit.csv` schema changes are append-only.
- Never fabricate: unmeasured values are `Unavailable` entries / zero-valued struct with `Measured=false`, never invented numbers.
- Commit after every task with the message given in the task's last step.

---

### Task 1: Model A — amortized CIR depreciation (`pkg/metrics/cost.go`)

**Files:**
- Create: `pkg/metrics/cost.go`
- Create: `pkg/metrics/cost_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `EnterprisePlan` struct; vars `PlanGOLD`, `PlanPLATINE`; `AmortizedARPerSecond(monthlyOpEx float64) float64`; `CapacityWasteARPerHour(monthlyOpEx, wastedBytes, transportedBytes float64) float64`.

- [ ] **Step 1: Write the failing test**

Create `pkg/metrics/cost_test.go`:

```go
package metrics

import "testing"

func TestAmortizedARPerSecond(t *testing.T) {
	// 1 000 000 Ar / (720 h × 3600 s) = 0.3858 Ar/s (3 déciales utiles)
	got := AmortizedARPerSecond(1_000_000)
	want := 1_000_000.0 / (720 * 3600)
	if diff := got - want; diff > 1e-12 || diff < -1e-12 {
		t.Fatalf("AmortizedARPerSecond = %v, want %v", got, want)
	}
}

func TestCapacityWasteARPerHour(t *testing.T) {
	// 10 % de gaspillage sur un GOLD : 0.1 × (1 000 000/720) = 138,89 Ar/h
	got := CapacityWasteARPerHour(1_000_000, 100, 1000)
	want := 0.1 * (1_000_000.0 / 720)
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("CapacityWasteARPerHour = %v, want %v", got, want)
	}
	if CapacityWasteARPerHour(1_000_000, 100, 0) != 0 {
		t.Fatal("transported=0 doit rendre 0, pas une division par zéro")
	}
}

func TestEnterprisePlans(t *testing.T) {
	if PlanGOLD.MonthlyOpExMGA != 1_000_000 || PlanGOLD.CIRMbps != 200 {
		t.Fatalf("GOLD = %+v", PlanGOLD)
	}
	if PlanPLATINE.MonthlyOpExMGA != 2_500_000 || PlanPLATINE.FUPBytes != 30*1024*1024*1024*1024 {
		t.Fatalf("PLATINE = %+v", PlanPLATINE)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/metrics/ -run 'TestAmortized|TestCapacityWaste|TestEnterprisePlans' -v`
Expected: FAIL — `undefined: AmortizedARPerSecond`.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/metrics/cost.go`:

```go
package metrics

// Modèle économique entreprise — spéc. 2026-10-25 §4. Les tarifs sont des
// scénarios contractuels sourcés (docs/data-prices.md), pas des mesures.

// EnterprisePlan — liaison dédiée CIR (modèle A : dépréciation de capacité).
type EnterprisePlan struct {
	Name           string  `json:"name"`
	MonthlyOpExMGA float64 `json:"monthly_opex_mga"`
	CIRMbps        float64 `json:"cir_mbps"`
	FUPBytes       float64 `json:"fup_bytes"`
}

// Plans entreprise Yas Grande Entreprise (yas.mg, relevé 31/08/2026).
var (
	PlanGOLD = EnterprisePlan{Name: "enterprise-gold-cir", MonthlyOpExMGA: 1_000_000, CIRMbps: 200}
	// 30 To = 30 × 2^40 octets : 2 500 000 / 30 720 GiB = 81,38 Ar/Go.
	PlanPLATINE = EnterprisePlan{Name: "enterprise-platine1", MonthlyOpExMGA: 2_500_000, FUPBytes: 30 * 1024 * 1024 * 1024 * 1024}
)

// AmortizedARPerSecond — coût amorti de la liaison à la seconde.
// Le mois contractuel est compté à 720 h (30 j × 24 h), convention du mémoire.
func AmortizedARPerSecond(monthlyOpEx float64) float64 {
	return monthlyOpEx / (720 * 3600)
}

// CapacityWasteARPerHour — dépréciation horaire de capacité contractuelle :
// la part des octets transportés gaspillée en réémissions, rapportée au
// douzième de l'OpEx mensuel. Un lien dédié ne facture pas l'octet ; il
// facture la capacité que le bufferbloat rend inutilisable.
func CapacityWasteARPerHour(monthlyOpEx, wastedBytes, transportedBytes float64) float64 {
	if transportedBytes <= 0 {
		return 0
	}
	return (wastedBytes / transportedBytes) * (monthlyOpEx / 720)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/metrics/ -run 'TestAmortized|TestCapacityWaste|TestEnterprisePlans' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add pkg/metrics/cost.go pkg/metrics/cost_test.go
git commit -m "metrics: modèle A — dépréciation CIR amortie (GOLD/PLATINE)"
```

---

### Task 2: Model B — metered B2B waste with window generalization

**Files:**
- Modify: `pkg/metrics/cost.go`
- Modify: `pkg/metrics/cost_test.go`

**Interfaces:**
- Produces: `B2BTariff` struct; vars `TariffWholesale`, `TariffCorporateAPN`; const `RetailEmergencyARPerGB = 5556`; `MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) float64`.

- [ ] **Step 1: Write the failing test**

Append to `pkg/metrics/cost_test.go`:

```go
func TestMeteredWasteARPerHour(t *testing.T) {
	// 1 GiB gaspillé à 81,38 Ar/Go sur une fenêtre de 180 s :
	// 81,38 × 20 = 1627,6 Ar/h — le ×20 de la directive est LE cas 180 s.
	got := MeteredWasteARPerHour(1024*1024*1024, 81.38, 180)
	want := 81.38 * 20
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("×20 : got %v want %v", got, want)
	}
	// Fenêtre mesurée 40 s : facteur 90, pas 20.
	got40 := MeteredWasteARPerHour(1024*1024*1024, 150, 40)
	if diff := got40 - 150*90; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("fenêtre 40 s : got %v want %v", got40, 150*90)
	}
	if MeteredWasteARPerHour(1e9, 81.38, 0) != 0 {
		t.Fatal("windowSeconds=0 doit rendre 0")
	}
}

func TestB2BTariffs(t *testing.T) {
	if TariffWholesale.ARPerGB != 81.38 || TariffCorporateAPN.ARPerGB != 150 {
		t.Fatalf("tarifs B2B = %+v %+v", TariffWholesale, TariffCorporateAPN)
	}
	if RetailEmergencyARPerGB != 5556 {
		t.Fatalf("secours détail = %v", RetailEmergencyARPerGB)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/metrics/ -run 'TestMeteredWaste|TestB2BTariffs' -v`
Expected: FAIL — `undefined: MeteredWasteARPerHour`.

- [ ] **Step 3: Write minimal implementation**

Append to `pkg/metrics/cost.go`:

```go
// B2BTariff — tarif grossiste/apn facturé au gigaoctet (modèle B).
type B2BTariff struct {
	Name     string  `json:"name"`
	ARPerGB  float64 `json:"ar_per_gb"`
	// Quoted marque les tarifs relevés sur devis opérateur (pas de grille
	// publique) : le rapport doit le dire, c'est un scénario contractuel.
	Quoted bool `json:"quoted,omitempty"`
}

var (
	// Dérivé de PLATINE 1 : 2 500 000 Ar / 30 720 GiB.
	TariffWholesale = B2BTariff{Name: "b2b-platine-bulk", ARPerGB: 81.38}
	// APN corporate M2M/IoT — devis opérateur (oct. 2026), pas une grille publique.
	TariffCorporateAPN = B2BTariff{Name: "b2b-m2m-apn", ARPerGB: 150, Quoted: true}
)

// RetailEmergencyARPerGB — le prépayé grand public (Yas Net Month 4,5 Go)
// ne survit que comme pénalité de hors-forfait d'urgence, jamais comme
// modèle par défaut d'une infrastructure opérationnelle.
const RetailEmergencyARPerGB = 5556

// MeteredWasteARPerHour — coût horaire du gaspillage sur un lien facturé au
// volume : le ×20 de la directive est le cas particulier fenêtre = 180 s ;
// l'audit passe sa fenêtre de charge MESURÉE, la campagne son 180 s historique.
func MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) float64 {
	if windowSeconds <= 0 {
		return 0
	}
	gb := wastedBytes / (1024 * 1024 * 1024)
	return gb * tariffARPerGB * (3600 / windowSeconds)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/metrics/ -run 'TestMeteredWaste|TestB2BTariffs' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/metrics/cost.go pkg/metrics/cost_test.go
git commit -m "metrics: modèle B — gaspillage facturé au volume, fenêtre généralisée"
```

---

### Task 3: Model C — CAP SLA at the 220 ms operator objective

**Files:**
- Modify: `pkg/metrics/cost.go`
- Modify: `pkg/metrics/cost_test.go`

**Interfaces:**
- Consumes: `DeadlineOKPct(completionsMs []float64, dueMs float64) float64` (`metrics.go:47`).
- Produces: consts `CAPDeadlineMs = 220`, `CAPBreachThresholdPct = 95`; struct `CAPSLA{Measured bool, OKPct float64, Breach bool, Label string}`; `AssessCAPSLA(loadedSmallCompletionsMs []float64) CAPSLA`.

- [ ] **Step 1: Write the failing test**

Append to `pkg/metrics/cost_test.go`:

```go
func TestAssessCAPSLA(t *testing.T) {
	// Aucun échantillon chargé : non mesuré, jamais un succès de complaisance.
	if s := AssessCAPSLA(nil); s.Measured || s.Breach {
		t.Fatalf("sans échantillons : %+v", s)
	}
	// 3/5 sous 220 ms = 60 % < 95 % : manquement (100,150,210 passent ; 700,300 échouent).
	s := AssessCAPSLA([]float64{100, 150, 210, 700, 300})
	if !s.Measured || !s.Breach || s.OKPct != 60 {
		t.Fatalf("60%% attendu en échec : %+v", s)
	}
	if s.Label == "" {
		t.Fatal("un manquement doit porter son étiquette opérationnelle")
	}
	// Tous sous 220 ms : conforme, pas de label d'alerte.
	ok := AssessCAPSLA([]float64{100, 150, 210, 219})
	if !ok.Measured || ok.Breach || ok.Label != "" {
		t.Fatalf("conforme : %+v", ok)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/metrics/ -run TestAssessCAPSLA -v`
Expected: FAIL — `undefined: AssessCAPSLA`.

- [ ] **Step 3: Write minimal implementation**

Append to `pkg/metrics/cost.go`:

```go
// Objectif opérationnel DGM pour une transaction CAP (sonde 16 KiB). Ce n'est
// PAS un seuil imposé par l'OMM : le Guide WIS 2.0 (WMO-No 1061) définit le
// « temps réel » comme « de quelques secondes à quelques minutes », sans
// seuil numérique par bulletin — le mémoire cite ce cadre comme contexte,
// jamais comme source du 220 ms.
const (
	CAPDeadlineMs         = 220
	CAPBreachThresholdPct = 95.0
)

// CAPSLA — conformité des transactions CAP sous charge (modèle C).
type CAPSLA struct {
	Measured bool    `json:"measured"`
	OKPct    float64 `json:"ok_pct"`
	Breach   bool    `json:"breach"`
	Label    string  `json:"label,omitempty"`
}

// AssessCAPSLA évalue les complétions 16 KiB CHARGÉES contre l'objectif DGM.
// Sans échantillon : non mesuré (Measured=false), jamais un passage par défaut.
func AssessCAPSLA(loadedSmallCompletionsMs []float64) CAPSLA {
	if len(loadedSmallCompletionsMs) == 0 {
		return CAPSLA{}
	}
	pct := DeadlineOKPct(loadedSmallCompletionsMs, CAPDeadlineMs)
	s := CAPSLA{Measured: true, OKPct: pct, Breach: pct < CAPBreachThresholdPct}
	if s.Breach {
		s.Label = "ALERTE OPÉRATIONNELLE — MANQUEMENT SLA (objectif DGM 220 ms, cadre WIS 2.0)"
	}
	return s
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/metrics/ -run TestAssessCAPSLA -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/metrics/cost.go pkg/metrics/cost_test.go
git commit -m "metrics: modèle C — SLA CAP à l'objectif opérateur 220 ms"
```

---

### Task 4: Scenario assembly — `Economics` with primary by link type

**Files:**
- Modify: `pkg/metrics/cost.go`
- Modify: `pkg/metrics/cost_test.go`

**Interfaces:**
- Consumes: Tasks 1–3 (`CapacityWasteARPerHour`, `MeteredWasteARPerHour`, plans, tariffs, `RetailEmergencyARPerGB`).
- Produces: `RetransRatio{Ratio float64, Basis string}` (Basis ∈ `"measured"`, `"estimated"`, `""`); `EconomicsScenario{Model, Name string, ARPerHour float64, Basis string}`; `Economics{PrimaryLinkType string, Primary *EconomicsScenario, Scenarios []EconomicsScenario, RetransRatio RetransRatio, WastedBytes float64, TransportedBytes float64, LoadSeconds float64, Disclaimer string}`; `BuildEconomics(linkType string, ratio RetransRatio, transportedBytes, wastedBytes, loadSeconds float64) *Economics`.

- [ ] **Step 1: Write the failing test**

Append to `pkg/metrics/cost_test.go`:

```go
func TestBuildEconomicsScenarios(t *testing.T) {
	e := BuildEconomics("4g", RetransRatio{Ratio: 0.05, Basis: "measured"}, 100_000_000, 5_000_000, 40)
	if len(e.Scenarios) != 5 {
		t.Fatalf("5 scénarios (GOLD, PLATINE, 81,38, APN 150, détail) : %d", len(e.Scenarios))
	}
	if e.Primary == nil || e.Primary.Model != "B" || e.Primary.Name != TariffCorporateAPN.Name {
		t.Fatalf("primaire 4g = APN corporate : %+v", e.Primary)
	}
	for _, s := range e.Scenarios {
		if s.ARPerHour <= 0 || s.Basis == "" {
			t.Fatalf("scénario incohérent : %+v", s)
		}
	}
	if e.Disclaimer == "" {
		t.Fatal("le rapport doit dire que ce sont des scénarios contractuels, pas des mesures")
	}
}

func TestPrimaryScenarioMapping(t *testing.T) {
	cases := map[string]string{
		"fiber": PlanGOLD.Name, "other": PlanGOLD.Name,
		"4g": TariffCorporateAPN.Name, "5g": TariffCorporateAPN.Name,
		"vsat": TariffWholesale.Name,
	}
	for lt, want := range cases {
		e := BuildEconomics(lt, RetransRatio{Ratio: 0.01, Basis: "estimated"}, 1e9, 1e7, 40)
		if e.Primary == nil || e.Primary.Name != want {
			t.Fatalf("link %q : primaire %+v, want %s", lt, e.Primary, want)
		}
	}
}

func TestBuildEconomicsNoRatio(t *testing.T) {
	// Sans ratio mesuré ni estimé : aucun scénario chiffré — jamais un zéro
	// qui se lirait « aucun gaspillage ».
	e := BuildEconomics("fiber", RetransRatio{}, 1e9, 0, 40)
	if len(e.Scenarios) != 0 {
		t.Fatalf("sans ratio : zéro scénario, pas des zéros : %+v", e.Scenarios)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/metrics/ -run 'TestBuildEconomics|TestPrimaryScenario' -v`
Expected: FAIL — `undefined: RetransRatio`.

- [ ] **Step 3: Write minimal implementation**

Append to `pkg/metrics/cost.go`:

```go
// RetransRatio — part des octets réémis. Basis dit d'où elle vient :
// "measured" (TCP_INFO sur nos sockets), "estimated" (proxy perte sonde ×
// volume), "" (non mesuré — alors aucun scénario n'est chiffré).
type RetransRatio struct {
	Ratio  float64 `json:"ratio"`
	Basis  string  `json:"basis"`
}

// EconomicsScenario — une valorisation horaire du gaspillage mesuré.
type EconomicsScenario struct {
	Model     string  `json:"model"` // "A" (CIR) | "B" (au volume) | "emergency"
	Name      string  `json:"name"`
	ARPerHour float64 `json:"ar_per_hour"`
	Basis     string  `json:"basis"`
}

// Economics — valorisation complète d'un audit : tous les scénarios, le
// primaire MARQUÉ selon le type de lien déclaré (une correspondance de
// scénario, pas une affirmation sur le contrat réel du site).
type Economics struct {
	PrimaryLinkType  string              `json:"primary_link_type"`
	Primary          *EconomicsScenario  `json:"primary,omitempty"`
	Scenarios        []EconomicsScenario `json:"scenarios"`
	RetransRatio     RetransRatio        `json:"retrans_ratio"`
	WastedBytes      float64             `json:"wasted_bytes"`
	TransportedBytes float64             `json:"transported_bytes"`
	LoadSeconds      float64             `json:"load_seconds"`
	Disclaimer       string              `json:"disclaimer"`
}

func primaryName(linkType string) string {
	switch linkType {
	case "4g", "5g":
		return TariffCorporateAPN.Name
	case "vsat":
		return TariffWholesale.Name
	default: // fiber, other, vide
		return PlanGOLD.Name
	}
}

// BuildEconomics assemble tous les scénarios à partir du gaspillage mesuré.
// transportedBytes/wastedBytes en octets ; loadSeconds = fenêtre de charge
// mesurée (généralisation du ×20 de la directive).
func BuildEconomics(linkType string, ratio RetransRatio, transportedBytes, wastedBytes, loadSeconds float64) *Economics {
	e := &Economics{
		PrimaryLinkType:  linkType,
		RetransRatio:     ratio,
		WastedBytes:      wastedBytes,
		TransportedBytes: transportedBytes,
		LoadSeconds:      loadSeconds,
		Disclaimer:       "Scénarios contractuels (yas.mg 31/08/2026 ; APN 150 Ar/Go sur devis), pas des mesures : le chiffre retenu dépend du contrat du site, pas de la physique du lien.",
	}
	if ratio.Basis == "" || transportedBytes <= 0 {
		return e // non mesuré : aucun zéro fabriqué
	}
	// Modèle A : même ratio, deux plans dédiés.
	for _, p := range []EnterprisePlan{PlanGOLD, PlanPLATINE} {
		e.Scenarios = append(e.Scenarios, EconomicsScenario{
			Model: "A", Name: p.Name,
			ARPerHour: CapacityWasteARPerHour(p.MonthlyOpExMGA, wastedBytes, transportedBytes),
			Basis:     fmt.Sprintf("(%.4f réémis / %.4f transportés) × %s Ar/720 h", ratio.Ratio, ratio.Ratio, humanMGA(p.MonthlyOpExMGA)),
		})
	}
	// Modèle B : gaspillage facturé au volume, deux tarifs B2B.
	for _, tr := range []B2BTariff{TariffWholesale, TariffCorporateAPN} {
		e.Scenarios = append(e.Scenarios, EconomicsScenario{
			Model: "B", Name: tr.Name,
			ARPerHour: MeteredWasteARPerHour(wastedBytes, tr.ARPerGB, loadSeconds),
			Basis:     fmt.Sprintf("%.2f Mo gaspillés / fenêtre %.0f s × %.2f Ar/Go", wastedBytes/1e6, loadSeconds, tr.ARPerGB),
		})
	}
	// Pénalité de secours détail — visible, jamais primaire.
	e.Scenarios = append(e.Scenarios, EconomicsScenario{
		Model: "emergency", Name: "retail-outofbundle",
		ARPerHour: MeteredWasteARPerHour(wastedBytes, RetailEmergencyARPerGB, loadSeconds),
		Basis:     "hors-forfait d'urgence (Net Month 4,5 Go) — pénalité, pas un modèle d'infrastructure",
	})
	want := primaryName(linkType)
	for i := range e.Scenarios {
		if e.Scenarios[i].Name == want {
			prim := e.Scenarios[i]
			e.Primary = &prim
			break
		}
	}
	return e
}

// humanMGA — 1 000 000 → "1 000 000" (lisibilité des basis strings).
func humanMGA(v float64) string {
	s := strconv.FormatFloat(v, 'f', 0, 64)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	return string(out)
}
```

Add `"fmt"` and `"strconv"` to cost.go imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/metrics/ -run 'TestBuildEconomics|TestPrimaryScenario' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add pkg/metrics/cost.go pkg/metrics/cost_test.go
git commit -m "metrics: assemblage Economics — tous scénarios, primaire par type de lien"
```

---

### Task 5: TCP_INFO retransmit sampler + stats-returning bulk helpers

**Files:**
- Create: `pkg/probe/tcpinfo_linux.go`
- Create: `pkg/probe/tcpinfo_other.go`
- Create: `pkg/probe/bulkstats.go`
- Create: `pkg/probe/bulkstats_test.go`

**Interfaces:**
- Consumes: existing `DialWithCC` (`bulk_cc_linux.go:14` / `bulk_cc_other.go:11`), `BulkSendNTo` (`bulkflows.go:52`), `BulkDownloadNTo` (`bulksource.go:58`).
- Produces: `TCPRetransBytes(c net.Conn) (uint64, error)` (Linux real, other: error); `BulkStats{Bytes []uint64, RetransBytes uint64, RetransMeasured bool}`; `BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)`; `BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error)`.

- [ ] **Step 1: Write the failing test**

Create `pkg/probe/bulkstats_test.go`:

```go
package probe

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestBulkStatsMultiFlowLocal(t *testing.T) {
	// Puits local : 3 flux vers un listener discard — le total par flux doit
	// remplir Bytes, et RetransMeasured suit la plateforme (false sur Windows,
	// indifférent ici : on teste la structure, pas le noyau).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("loopback indisponible")
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = BulkReceive(ctx, c, nil)
			}(c)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	st, err := BulkSendNStats(ctx, ln.Addr().String(), "cubic", 3)
	if err != nil {
		t.Fatalf("BulkSendNStats: %v", err)
	}
	if len(st.Bytes) != 3 {
		t.Fatalf("3 flux attendus, %d rendus", len(st.Bytes))
	}
	for _, b := range st.Bytes {
		if b == 0 {
			t.Fatal("un flux n'a rien transporté — la charge doit être réelle")
		}
	}
}

func TestTCPRetransBytesErrorOnNonTCP(t *testing.T) {
	// Un non-TCPConn doit rendre une erreur propre, pas un panic.
	if _, err := TCPRetransBytes(udpConn(t)); err == nil {
		t.Fatal("non-TCP conn : erreur attendue")
	}
}

func udpConn(t *testing.T) net.Conn {
	t.Helper()
	c, err := net.Dial("udp", "127.0.0.1:9")
	if err != nil {
		t.Skip("udp indisponible")
	}
	t.Cleanup(func() { c.Close() })
	return c
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/probe/ -run 'TestBulkStats|TestTCPRetrans' -v`
Expected: FAIL — `undefined: BulkSendNStats`.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/probe/tcpinfo_linux.go`:

```go
//go:build linux

package probe

import (
	"fmt"
	"net"
	"syscall"
)

// MSS de référence du banc (veth, MTU 1500) : TCP_INFO rend des SEGMENTS
// réémis ; ×1448 les convertit en octets comme le fait le mémoire
// (wasted = drops × 1448). Proxy documenté, pas une mesure par paquet.
const retransMSS = 1448

// TCPRetransBytes — octets réémis cumulés d'une connexion TCP (TCP_INFO).
// Linux uniquement : ailleurs, l'appel rend une erreur et l'appelant déclare
// le ratio non mesuré — jamais estimé en silence.
func TCPRetransBytes(c net.Conn) (uint64, error) {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("tcpinfo: pas une connexion TCP")
	}
	f, err := tc.File()
	if err != nil {
		return 0, fmt.Errorf("tcpinfo: file: %w", err)
	}
	defer f.Close()
	info, err := syscall.GetsockoptTCPInfo(int(f.Fd()), syscall.IPPROTO_TCP, syscall.TCP_INFO)
	if err != nil {
		return 0, fmt.Errorf("tcpinfo: getsockopt: %w", err)
	}
	return uint64(info.TotalRetrans) * retransMSS, nil
}
```

Create `pkg/probe/tcpinfo_other.go`:

```go
//go:build !linux

package probe

import (
	"errors"
	"net"
)

// ErrTCPInfoUnavailable — TCP_INFO n'existe que sur Linux ; ailleurs le ratio
// de réémission est NON MESURÉ (l'audit le déclare, il ne l'invente pas).
var ErrTCPInfoUnavailable = errors.New("tcpinfo: TCP_INFO indisponible hors Linux")

func TCPRetransBytes(net.Conn) (uint64, error) { return 0, ErrTCPInfoUnavailable }
```

Create `pkg/probe/bulkstats.go`:

```go
package probe

import (
	"context"
	"net"
	"sync"
)

// BulkStats — charge multi-flux avec la vérité par flux ET le ratio de
// réémission : JFI se calcule sur Bytes, le modèle A sur RetransBytes.
type BulkStats struct {
	Bytes           []uint64 `json:"bytes"`
	RetransBytes    uint64   `json:"retrans_bytes"`
	RetransMeasured bool     `json:"retrans_measured"`
}

// bulkNStats — moteur commun : N connexions, échantillonnage TCP_INFO à la
// fermeture de chaque flux (le compteur est cumulatif par socket).
func bulkNStats(ctx context.Context, dial func(ctx context.Context) (net.Conn, error), send func(ctx context.Context, c net.Conn) (uint64, error), n int) (BulkStats, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	st := BulkStats{Bytes: make([]uint64, 0, n)}
	oneTCP := false // au moins un socket échantillonné => measured
	for i := 0; i < n; i++ {
		c, err := dial(ctx)
		if err != nil {
			continue // lien en défaut : les flux vivants comptent
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			b, _ := send(ctx, c)
			rb, rerr := TCPRetransBytes(c)
			c.Close()
			mu.Lock()
			st.Bytes = append(st.Bytes, b)
			st.RetransBytes += rb
			if rerr == nil {
				oneTCP = true
			}
			mu.Unlock()
		}(c)
	}
	wg.Wait()
	st.RetransMeasured = oneTCP
	return st, nil
}

// BulkSendNStats — charge MONTANTE multi-flux avec statistiques.
func BulkSendNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkSend(ctx, c) },
		n)
}

// BulkDownloadNStats — charge DESCENDANTE multi-flux avec statistiques.
func BulkDownloadNStats(ctx context.Context, addr, cc string, n int) (BulkStats, error) {
	return bulkNStats(ctx,
		func(ctx context.Context) (net.Conn, error) { return DialWithCC(ctx, addr, cc) },
		func(ctx context.Context, c net.Conn) (uint64, error) { return BulkRecvSource(ctx, c) },
		n)
}
```

Note: check whether `BulkDownloadTo` (`bulksource.go:37`) reads from a dialed conn via a helper; if a `BulkRecvSource(ctx, conn)` reader does not exist, extract the read-loop from `BulkDownloadTo` into `BulkRecvSource` in `bulksource.go` so both share it — do not duplicate logic. The dial for download must send the `"D:"+cc` hello exactly as `BulkDownloadTo` does; reuse its hello-writing code path the same way.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/probe/ -run 'TestBulkStats|TestTCPRetrans' -v` then `go build ./...`
Expected: PASS on Windows (stub path; `RetransMeasured=false`); build clean on both platforms. On the CGO-DEV VM later (Plan 5), the Linux path is exercised for real.

- [ ] **Step 5: Commit**

```bash
git add pkg/probe/tcpinfo_linux.go pkg/probe/tcpinfo_other.go pkg/probe/bulkstats.go pkg/probe/bulkstats_test.go pkg/probe/bulksource.go
git commit -m "probe: échantillonneur TCP_INFO + charge multi-flux avec statistiques"
```

---

### Task 6: Audit direction, multi-flow load, JFI

**Files:**
- Modify: `pkg/audit/audit.go`
- Modify: `pkg/audit/audit_test.go` (and `honesty_test.go` / `phase_test.go` / `audit_split_test.go` call sites that construct `Deps.Bulk`)

**Interfaces:**
- Consumes: Task 5 (`probe.BulkSendNStats`, `probe.BulkDownloadNStats`, `probe.TCPRetransBytes`), `metrics.JFI(xs []uint64) float64` (`metrics.go:152`).
- Produces: `Params.Direction string` (`""`/`up`/`down`/`both`, default `up`), `Params.Flows int` (default 4); `Deps.BulkLoad func(ctx context.Context, addr, direction string, flows int) LoadOutcome` (replaces `Deps.Bulk`); `LoadOutcome{Bytes uint64, PerFlow []uint64, RetransBytes uint64, RetransMeasured bool, Err error, Elapsed time.Duration}`; `Result.Direction string`, `Result.FlowCount int`, `Result.JFIPct float64`, `Result.RetransRatio metrics.RetransRatio`, download-mirror fields `RTTLoadedDownP50/P95 float64`, `ThroughputDownMbps float64`, `BloatDeltaDownMs float64`, `GradeDown string`, `DrainDownMs float64`, `DrainDownMeasured bool`, `JFIDownPct float64`.

- [ ] **Step 1: Write the failing test**

Append to `pkg/audit/audit_test.go` (adapt existing fake-bulk tests in the same commit — they move from `Deps.Bulk` to `Deps.BulkLoad`):

```go
func TestRunDirectionDown(t *testing.T) {
	// down : la charge doit passer par BulkLoad(direction="down") et remplir
	// les champs miroirs (ThroughputDownMbps), pas les champs up.
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
	if res.ThroughputDownMbps == 0 || res.ThroughputMbps != 0 {
		t.Fatalf("down doit remplir les champs miroirs : up=%v down=%v", res.ThroughputMbps, res.ThroughputDownMbps)
	}
	if res.JFIDownPct <= 0 || res.JFIPct != 0 {
		t.Fatalf("JFI down = %v, JFI up = %v", res.JFIDownPct, res.JFIPct)
	}
	if res.Direction != "down" {
		t.Fatalf("Direction gelé = %q", res.Direction)
	}
}

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit/ -run 'TestRunDirection|TestRunRetrans' -v`
Expected: FAIL — `unknown field Direction` / `unknown field BulkLoad`.

- [ ] **Step 3: Write minimal implementation**

In `pkg/audit/audit.go`:

1. Add to `Params` (after `BulkAddr`):

```go
	// Direction — sens de la charge : "" ou "up" (upload), "down"
	// (download — le scénario DGM : l'image EUMETSAT qui déborde la file),
	// "both" (deux fenêtres de charge successives, un seul rapport).
	Direction string
	// Flows — flux de charge concurrents (défaut 4, régime du banc durci).
	Flows int
```

2. Replace `Deps.Bulk` with:

```go
type Deps struct {
	Ping      func(ctx context.Context, target string, n int) []float64
	Small     func(ctx context.Context) (float64, error)
	BulkLoad  func(ctx context.Context, addr, direction string, flows int) LoadOutcome
}
```

3. Add the outcome type near `bulkResult` (delete `bulkResult`, superseded):

```go
// LoadOutcome — le résultat d'une fenêtre de charge : total, par flux (JFI),
// réémissions (modèle A) et durée réelle. Err non-nil = charge échouée.
type LoadOutcome struct {
	Bytes           uint64
	PerFlow         []uint64
	RetransBytes    uint64
	RetransMeasured bool
	Err             error
	Elapsed         time.Duration
}
```

4. Default wiring where `d.Bulk` was built (replace the `probe.BulkSendTo` block):

```go
	if d.BulkLoad == nil {
		d.BulkLoad = func(ctx context.Context, addr, direction string, flows int) LoadOutcome {
			t0 := time.Now()
			var st probe.BulkStats
			var err error
			if direction == "down" {
				st, err = probe.BulkDownloadNStats(ctx, addr, "bbr", flows)
			} else {
				st, err = probe.BulkSendNStats(ctx, addr, "bbr", flows)
			}
			out := LoadOutcome{Bytes: sumU64(st.Bytes), PerFlow: st.Bytes,
				RetransBytes: st.RetransBytes, RetransMeasured: st.RetransMeasured,
				Elapsed: time.Since(t0)}
			if len(st.Bytes) == 0 && err == nil {
				err = fmt.Errorf("aucun flux établi vers %s", addr)
			}
			out.Err = err
			return out
		}
	}
	if p.Flows <= 0 {
		p.Flows = 4
	}
	switch p.Direction {
	case "", "up":
		p.Direction = "up"
	case "down", "both":
	default:
		return nil, fmt.Errorf("audit %s : direction %q invalide (up|down|both)", p.AuditID, p.Direction)
	}
```

Add helper `func sumU64(xs []uint64) uint64` in audit.go.

5. Restructure the load block: extract the current single-window sequence into a closure `runLoadWindow(dir string, idleRTTs, idleSmalls *[]float64) (loadedRTTs, loadedSmalls []float64, oc LoadOutcome)` that starts the bulk, collects `loadedSecs`, cancels, and (for `dir == "down"`) collects into the down-labelled slices. For `Direction == "both"`, call it twice (`up` then `down`) with a 3 s settle `time.Sleep` between, sharing ONE baseline window collected before the first load; the drain window runs after the second load and fills `DrainDownMs` for `both`/`down` and `DrainMs` for `up`/`both` (for `both`, drain is measured after the down window into `DrainDownMs`, and `DrainMs` after the up window using a 5 s mini-drain before the settle). For single directions, behavior matches today: up → existing fields; down → mirror fields (`RTTLoadedDownP50/P95` from that window's summary, `ThroughputDownMbps`, `BloatDeltaDownMs = loadedDownP95 - idleP95`, `GradeDown = metrics.BufferbloatGrade(BloatDeltaDownMs)` when load measured, `JFIDownPct = metrics.JFI(oc.PerFlow)` when `len(oc.PerFlow) > 1`).

6. Ratio computation after the window(s):

```go
	ratio := metrics.RetranRatio{}
	if loadMeasured {
		total := float64(bulkBytes)
		switch {
		case oc.RetransMeasured && total > 0:
			ratio = metrics.RetranRatio{Ratio: float64(oc.RetransBytes) / total, Basis: "measured"}
		case res.LossPct > 0 && total > 0: // proxy documenté : perte sonde × volume
			ratio = metrics.RetransRatio{Ratio: res.LossPct / 100, Basis: "estimated"}
		}
	}
```

(stored in `Result.RetransRatio`; for `both`, the up-window ratio wins and the down-window ratio is dropped — one headline ratio per report, the conservative choice is the larger: `if downRatio > upRatio { keep downRatio }`).

7. Adapt existing tests: replace every `Deps{Bulk: func(...)(uint64,error){...}}` with the equivalent `BulkLoad` returning `LoadOutcome{Bytes: n, Elapsed: d}`. Files: `audit_test.go`, `honesty_test.go`, `phase_test.go`, `audit_split_test.go`, `report_fixture_test.go`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/audit/ -v`
Expected: PASS — all existing tests green after the `BulkLoad` migration, plus the 3 new ones.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/ pkg/probe/
git commit -m "audit: sens up|down|both, charge multi-flux (4), JFI et ratio de réémission"
```

---

### Task 7: CAP SLA + retrans ratio in `Result`, CSV appended columns

**Files:**
- Modify: `pkg/audit/audit.go` (compute in `Run`)
- Modify: `pkg/audit/writer.go`
- Modify: `pkg/audit/writer_test.go` (create if absent)

**Interfaces:**
- Consumes: Task 3 (`metrics.AssessCAPSLA`), Task 6 (`loadedSmalls` slices).
- Produces: `Result.CAPDeadlineOKPct float64` (−1 = non mesuré), `Result.CAPSLABreach bool`, `Result.CAPSLAMeasured bool`, `Result.CAPSLALabel string`; `link_audit.csv` columns `cap_deadline_ok_pct`, `cap_sla_breach` appended via the existing migration.

- [ ] **Step 1: Write the failing test**

Create/extend `pkg/audit/writer_test.go`:

```go
package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendLinkAuditCAPColumns(t *testing.T) {
	dir := t.TempDir()
	res := &Result{
		AuditID: "audit-x", Timestamp: "2026-10-25T10:00:00Z",
		Site: "S", LinkType: "4g", Provider: "yas",
		CAPSLAMeasured: true, CAPDeadlineOKPct: 88.5, CAPSLABreach: true,
		CAPSLALabel: "ALERTE OPÉRATIONNELLE",
	}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "link_audit.csv"))
	hdr := strings.Split(strings.Split(string(raw), "\n")[0], ",")
	found := 0
	for _, h := range hdr {
		if h == "cap_deadline_ok_pct" || h == "cap_sla_breach" {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("colonnes CAP absentes : %v", hdr)
	}
	if !strings.Contains(string(raw), "88.5") || !strings.Contains(string(raw), "true") {
		t.Fatal("valeurs CAP non écrites")
	}
}

func TestAppendLinkAuditMigratesOldHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "link_audit.csv")
	// historique court : 3 colonnes — la migration doit rembourrer, pas casser.
	os.WriteFile(path, []byte("audit_id,timestamp,site\nold,2026,t\n"), 0644)
	res := &Result{AuditID: "new", Timestamp: "now", Site: "S", LinkType: "4g"}
	if err := AppendLinkAudit(dir, res); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	cols := strings.Split(lines[0], ",")
	for _, row := range lines[1:] {
		if n := len(strings.Split(row, ",")); n != len(cols) {
			t.Fatalf("ligne %q a %d champs, en-tête en a %d", row, n, len(cols))
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit/ -run TestAppendLinkAudit -v`
Expected: FAIL — no CAP columns.

- [ ] **Step 3: Write minimal implementation**

`audit.go` — add fields to `Result` (after `RecoverySeconds`):

```go
	// SLA CAP (modèle C) : complétions 16 KiB CHARGÉES contre l'objectif
	// DGM 220 ms. CAPSLAMeasured=false ⇒ non mesuré, jamais un succès.
	CAPSLAMeasured   bool    `json:"cap_sla_measured"`
	CAPDeadlineOKPct float64 `json:"cap_deadline_ok_pct"`
	CAPSLABreach     bool    `json:"cap_sla_breach"`
	CAPSLALabel      string  `json:"cap_sla_label,omitempty"`
```

In `Run`, after the load windows (before building `res`), using the loaded small samples of the LAST exercised window (`loadedSmalls` for up/both-final, the down-labelled slice for down/both — for `both`, pool both windows' loaded samples):

```go
	capSamples := loadedSmalls
	if p.Direction == "down" {
		capSamples = loadedSmallsDown
	} else if p.Direction == "both" {
		capSamples = append(append([]float64(nil), loadedSmalls...), loadedSmallsDown...)
	}
	sla := metrics.AssessCAPSLA(capSamples)
```

and set the four fields on `res`.

`writer.go` — extend:

```go
	newCols := []string{"bloat_delta_ms", "bloat_grade", "bloat_verdict", "profile_match", "match_delta", "cap_deadline_ok_pct", "cap_sla_breach"}
```

and append to the row (after `r.MatchDelta`):

```go
		capPct, capBreach := "", ""
		if r.CAPSLAMeasured {
			capPct = fmt.Sprintf("%.1f", r.CAPDeadlineOKPct)
			capBreach = fmt.Sprintf("%t", r.CAPSLABreach)
		}
		// … puis dans row :
		capPct, capBreach,
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/audit/ -v`
Expected: PASS — migration test proves old files are padded, not broken.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/
git commit -m "audit: SLA CAP gelée dans le résultat et le CSV (colonnes append-only)"
```

---

### Task 8: Telemetry timeseries + dual seals

**Files:**
- Create: `pkg/audit/telemetry.go`
- Create: `pkg/audit/telemetry_test.go`
- Modify: `pkg/audit/audit.go` (record rows), `pkg/audit/report.go` (`Provenance`, `SealSource`, `WriteArtifacts`)

**Interfaces:**
- Produces: `TelemetryRow{Phase string, OffsetMs int64, RTTms float64, SmallMs float64, BulkBytes uint64}` (RTTms=0 && SmallMs=0 marker rows not written); `Result.Telemetry []TelemetryRow`; `WriteTelemetry(dir string, res *Result) (string, error)`; `Provenance` gains `TelemetryFile string`, `TelemetrySHA256 string`, `RowFileSHA256 string`; `(*Report).SealSource(path string)` behavior extended; `WriteArtifacts` writes `telemetry.csv` into the audit dir.

- [ ] **Step 1: Write the failing test**

Create `pkg/audit/telemetry_test.go`:

```go
package audit

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTelemetry(t *testing.T) {
	dir := t.TempDir()
	res := &Result{AuditID: "audit-t", Telemetry: []TelemetryRow{
		{Phase: "baseline", OffsetMs: 0, RTTms: 20.5},
		{Phase: "load", OffsetMs: 10000, RTTms: 140.2, SmallMs: 180.0},
		{Phase: "drain", OffsetMs: 50000, RTTms: 30.1},
	}}
	path, err := WriteTelemetry(dir, res)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	r := csv.NewReader(strings.NewReader(string(raw)))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][0] != "phase" || len(rows) != 4 {
		t.Fatalf("en-tête/lignes : %v", rows)
	}
	if rows[2][1] != "10000" || rows[2][4] != "" { // offset + bulk vide pour la sonde
		t.Fatalf("ligne load : %v", rows[2])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit/ -run TestWriteTelemetry -v`
Expected: FAIL — `undefined: TelemetryRow`.

- [ ] **Step 3: Write minimal implementation**

Create `pkg/audit/telemetry.go`:

```go
package audit

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TelemetryRow — une sonde horodatée : phase, offset depuis le début de
// l'audit, RTT ping, complétion petit objet, cumul d'octets de charge.
// C'est la matière PREMIÈRE du rapport ; le sceau SHA-256 la scelle.
type TelemetryRow struct {
	Phase     string  `json:"phase"`
	OffsetMs  int64   `json:"offset_ms"`
	RTTms     float64 `json:"rtt_ms"`
	SmallMs   float64 `json:"small_ms"`
	BulkBytes uint64  `json:"bulk_bytes"`
}

var telemetryHeader = []string{"phase", "offset_ms", "rtt_ms", "small_ms", "bulk_bytes"}

// WriteTelemetry — data/audits/<id>/telemetry.csv. Retourne le chemin pour
// que l'appelant puisse sceller ce qu'il vient d'écrire.
func WriteTelemetry(baseDir string, res *Result) (string, error) {
	dir := filepath.Join(baseDir, "audits", filepath.Base(strings.TrimSpace(res.AuditID)))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "telemetry.csv")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write(telemetryHeader)
	for _, r := range res.Telemetry {
		_ = w.Write([]string{
			r.Phase, fmt.Sprintf("%d", r.OffsetMs),
			fmt.Sprintf("%.2f", r.RTTms), fmt.Sprintf("%.2f", r.SmallMs),
			fmt.Sprintf("%d", r.BulkBytes),
		})
	}
	w.Flush()
	return path, w.Error()
}
```

In `audit.go`: add `Telemetry []TelemetryRow` to `Result`; in `Run`, record `start := time.Now()` (exists) and append a row in `collectWindow` per probe batch (phase label passed in: `"baseline"`, `"load"`, `"load-down"`, `"drain"`) plus a row per successful small fetch; skip rows where both RTT and Small are zero (lost probes are represented by absence, consistent with `lossPercent` counting).

In `report.go`: extend `Provenance` with the three new fields; add:

```go
// SealTelemetry scelle la série brute (telemetry.csv) : c'est la preuve
// première, avant toute agrégation.
func (r *Report) SealTelemetry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audit: seal telemetry %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	r.Provenance.TelemetryFile = path
	r.Provenance.TelemetrySHA256 = hex.EncodeToString(sum[:])
	return nil
}
```

and in `SealSource`, also store `RowFileSHA256 = hex.EncodeToString(sum[:])` (keeping `SourceSHA256` for compatibility). `WriteArtifacts` signature unchanged — it is called after `WriteTelemetry`/sealing by the CLI (Task 10 wires the order: `WriteTelemetry` → `SealTelemetry` → `AppendLinkAudit` → `SealSource` → `WriteArtifacts`).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/audit/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/
git commit -m "audit: télémétrie brute par sonde + double sceau SHA-256 (série + ligne CSV)"
```

---

### Task 9: Report v2 — Economics section, JFI, CAP badge, direction

**Files:**
- Modify: `pkg/audit/report.go`
- Modify: `pkg/audit/report_test.go`

**Interfaces:**
- Consumes: Task 4 (`metrics.BuildEconomics`), Tasks 6–7 (`Result` fields).
- Produces: `Report.Economics *metrics.Economics` (replaces `Waste`, which is removed); `Report.CAPSLA metrics.CAPSLA`; `Report.Measured.JFIPct float64`, `Report.Measured.Direction string`, mirrored down fields in `Measured`; Markdown renders the enterprise scenario table with primary marked, CAP badge, JFI, direction.

- [ ] **Step 1: Write the failing test**

Append to `pkg/audit/report_test.go` (keep the existing no-load honesty tests — update their `Waste` references to `Economics`):

```go
func TestBuildReportEconomicsAllScenarios(t *testing.T) {
	res := &Result{
		AuditID: "a", Timestamp: "t", LinkType: "4g",
		ThroughputMbps: 10, BloatGrade: "A", LoadSeconds: 40,
		DataUsedMB: 50, LossPct: 4,
		RTTIdleP50: 20, RTTIdleP95: 25, RTTLoadedP50: 25, RTTLoadedP95: 30,
		RetransRatio: metrics.RetransRatio{Ratio: 0.04, Basis: "measured"},
		Direction: "up", FlowCount: 4, JFIPct: 97.5,
		CAPSLAMeasured: true, CAPDeadlineOKPct: 97, CAPSLABreach: false,
	}
	rep, err := BuildReport(res, remediate.Link{}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Economics == nil || len(rep.Economics.Scenarios) != 5 {
		t.Fatalf("Economics : %+v", rep.Economics)
	}
	if rep.Economics.Primary == nil || rep.Economics.Primary.Name != metrics.TariffCorporateAPN.Name {
		t.Fatalf("primaire 4g : %+v", rep.Economics.Primary)
	}
	if rep.CAPSLA.Measured != true || rep.CAPSLA.Breach {
		t.Fatalf("CAP : %+v", rep.CAPSLA)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "97,5") && !strings.Contains(md, "97.5") {
		t.Fatal("JFI absent du rendu")
	}
	if !strings.Contains(md, metrics.TariffWholesale.Name) || !strings.Contains(md, "retail-outofbundle") {
		t.Fatal("tous les scénarios doivent être visibles, y compris secours")
	}
	if !strings.Contains(md, "objectif DGM 220 ms") && rep.CAPSLA.Measured {
		t.Fatal("le cadre du 220 ms doit être dit (objectif opérateur, pas un seuil OMM)")
	}
}

func TestBuildReportNoRatioNoEconomics(t *testing.T) {
	res := &Result{
		AuditID: "a", Timestamp: "t", LinkType: "fiber",
		ThroughputMbps: 10, BloatGrade: "A", LoadSeconds: 40, DataUsedMB: 50, LossPct: 0,
		RTTIdleP50: 20, RTTIdleP95: 25, RTTLoadedP50: 25, RTTLoadedP95: 30,
	}
	rep, _ := BuildReport(res, remediate.Link{}, "", true, false)
	if rep.Economics != nil {
		t.Fatal("sans ratio : pas de valorisation chiffrée (jamais de zéro fabriqué)")
	}
	found := false
	for _, u := range rep.Unavailable {
		if u.Metric == "economics" {
			found = true
		}
	}
	if !found {
		t.Fatal("l'absence de valorisation doit être déclarée indisponible")
	}
}
```

Add `"github.com/Realms4239/cgo/pkg/metrics"` to report_test.go imports if absent.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit/ -run TestBuildReport -v`
Expected: FAIL — `rep.Economics` undefined.

- [ ] **Step 3: Write minimal implementation**

In `report.go`:

1. Replace `Waste *Waste` with `Economics *metrics.Economics \`json:"economics,omitempty"\`` and `CAPSLA metrics.CAPSLA \`json:"cap_sla"\``; delete the `Waste` struct and its block in `BuildReport`.
2. In `BuildReport`, after the VoIP block:

```go
	// Valorisation entreprise : tous les scénarios, primaire MARQUÉ pour le
	// type de lien déclaré. Scénarios contractuels, pas des mesures.
	if loadMeasured && res.RetransRatio.Basis != "" && res.DataUsedMB > 0 && res.LoadSeconds > 0 {
		transported := res.DataUsedMB * 1e6
		var wasted float64
		if res.RetransRatio.Basis == "measured" {
			wasted = res.RetransRatio.Ratio * transported
		} else { // proxy documenté : la même part appliquée au volume transporté
			wasted = res.DataUsedMB * res.LossPct / 100 * 1e6
		}
		rep.Economics = metrics.BuildEconomics(res.LinkType, res.RetransRatio, transported, wasted, res.LoadSeconds)
	} else if loadMeasured {
		rep.Unavailable = append(rep.Unavailable, Unavailable{
			Metric: "economics",
			Reason: "ratio de réémission non mesuré (TCP_INFO indisponible et perte sonde nulle) ; aucun scénario n'est chiffré plutôt qu'un zéro fabriqué",
		})
	}
	rep.CAPSLA = metrics.CAPSLA{Measured: res.CAPSLAMeasured, OKPct: res.CAPDeadlineOKPct, Breach: res.CAPSLABreach, Label: res.CAPSLALabel}
```

3. `Measured` gains `JFIPct float64`, `FlowCount int`, `Direction string`, and the down mirrors (`RTTLoadedDownP95Ms`, `ThroughputDownMbps`, `BloatDeltaDownMs`, `JFIDownPct`) populated from `res`.
4. In `RenderMarkdown`: header line `- Sens: up|down|both (flux: N)`; Measured table rows for JFI (« Équité inter-flux (JFI) ») and down-mirrors when `Direction` is `down`/`both`; after the grade line, a CAP block:

```go
	if r.CAPSLA.Measured {
		state := "CONFORME"
		if r.CAPSLA.Breach {
			state = r.CAPSLA.Label
		}
		fmt.Fprintf(&b, "**SLA CAP (objectif DGM 220 ms, cadre WIS 2.0) : %.1f %% sous échéance — %s**\n\n", r.CAPSLA.OKPct, state)
	} else {
		fmt.Fprintf(&b, "**SLA CAP : non mesuré** — aucune complétion 16 KiB chargée ; jamais un succès par défaut.\n\n")
	}
```

and replace the `## Financial waste` section with `## Impact économique (scénarios contractuels)` rendering every scenario as a table row `| scénario | Ar/h | base |`, bolding the primary, keeping the emergency row last, then the disclaimer line `rep.Economics.Disclaimer`.
5. Update the old `deadline_ok_pct` Unavailable entry text to: `link audit exposes cap_deadline_ok_pct at the fixed 220 ms operator objective; the campaign's operator-defined deadline_ok_pct stays a separate, campaign-only metric`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/audit/ -v && go build ./...`
Expected: PASS (update the two legacy `Waste` assertions in `report_test.go` to `Economics` in the same edit; `cmd/cgo` still compiles because Task 10 updates its call site — if `printAuditSummary` referenced `Waste`, fix it minimally here: guard `if rep.Economics != nil`).

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/ cmd/cgo/
git commit -m "audit: rapport v2 — économie entreprise complète, JFI, SLA CAP, sens"
```

---

### Task 10: CLI surface simplification, sink auto-detection, exit codes, 80×24 summary

**Files:**
- Modify: `cmd/cgo/main.go:148-227` (audit case)
- Create: `cmd/cgo/auditsinks.go`
- Create: `cmd/cgo/auditsinks_test.go`

**Interfaces:**
- Consumes: Tasks 6–9 (`Params.Direction/Flows`, `WriteTelemetry`, `SealTelemetry`, `Report.Economics/CAPSLA`).
- Produces: `ResolveSinks() Sinks` where `Sinks{Mode string /* "dualns" | "legacy" | "external" */, Target, SmallURL, BulkAddr string}`; audit exit code 2 when the report carries load-dependent `Unavailable` entries; `printAuditSummary` renders the 80×24 table.

- [ ] **Step 1: Write the failing test**

Create `cmd/cgo/auditsinks_test.go`:

```go
package main

import (
	"net"
	"testing"
)

func TestResolveSinksPrefersLiveListener(t *testing.T) {
	// Un puits qui ÉCOUTE est le banc : le mode doit le refléter, pas se
	// fier à une adresse en dur.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("loopback indisponible")
	}
	defer ln.Close()
	s := ResolveSinks()
	if s.Mode == "" {
		t.Fatal("mode vide — la détection doit toujours trancher (au pire external)")
	}
	if s.Mode == "external" && s.BulkAddr != "" {
		t.Fatal("mode externe : pas de charge implicite (jamais de A+ sans charge)")
	}
}
```

(The probe order in `ResolveSinks` dials `10.30.0.1:5201` then `10.200.0.1:5201` with a 300 ms timeout each; in unit tests on the dev host neither answers, so `external` is the expected outcome — the test above asserts the invariants, not the environment.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/cgo/ -run TestResolveSinks -v`
Expected: FAIL — `undefined: ResolveSinks`.

- [ ] **Step 3: Write minimal implementation**

Create `cmd/cgo/auditsinks.go`:

```go
package main

import (
	"net"
	"time"
)

// Sinks — la détection du banc décide des défauts : dualns (banc durci,
// puits 10.30.0.1), legacy (testbed.sh, 10.200.0.1), external (lien réel —
// pas de charge implicite : jamais de note sans charge).
type Sinks struct {
	Mode     string
	Target   string
	SmallURL string
	BulkAddr string
}

func probeTCP(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func ResolveSinks() Sinks {
	if probeTCP("10.30.0.1:5201") {
		return Sinks{Mode: "dualns", Target: "10.30.0.1", SmallURL: "http://10.30.0.1:8081/small", BulkAddr: "10.30.0.1:5201"}
	}
	if probeTCP("10.200.0.1:5201") {
		return Sinks{Mode: "legacy", Target: "10.200.0.1", SmallURL: "http://10.200.0.1:8081/small", BulkAddr: "10.200.0.1:5201"}
	}
	return Sinks{Mode: "external", Target: "8.8.8.8"}
}

// envOrSink — l'opérateur garde la main : CGO_BULK_ADDR/CGO_SMALL_URL
// écrasent la détection (liaison de terrain explicite).
func (s Sinks) envOrBulk() string {
	if v := envDefault("CGO_BULK_ADDR", ""); v != "" {
		return v
	}
	return s.BulkAddr
}
```

Rewrite the audit case in `main.go`:

```go
	case "audit":
		fs := flag.NewFlagSet("audit", flag.ExitOnError)
		// Surface opérateur : quatre réglages. Le reste est avancé/override.
		linkType := fs.String("link-type", "5g", "link_type: fiber, 5g, 4g, vsat, other")
		site := fs.String("site", "Site X", "site / département")
		duration := fs.Int("duration", 60, "duration seconds (60 = 10/40/10)")
		direction := fs.String("direction", "up", "sens de charge : up | down | both")
		provider := fs.String("provider", "", "provider")
		flows := fs.Int("flows", 4, "flux de charge concurrents (1–8 ; 1 = régime historique)")
		// Avancé : overrides explicites de terrain.
		target := fs.String("target", "", "cible ping (défaut : détection banc)")
		small := fs.String("small", "", "URL petit objet (défaut : détection banc)")
		bulk := fs.String("bulk", "", "puits bulk host:port (défaut : détection banc ; vide en externe = pas de charge)")
		iface := fs.String("iface", "", "Linux egress device to shape in the remediation (e.g. eth0)")
		rosIface := fs.String("routeros-iface", "", "RouterOS interface for the remediation script (e.g. ether1-wan)")
		reportPath := fs.String("report", "", "additional report path (.md or .json); the standard data/audits/<id>/ artifacts are always written")
		qdisc := fs.String("qdisc", "", "qdisc for the remediation: cake, fq_codel, pfifo_fast (empty = recommended)")
		asymmetric := fs.Bool("asymmetric", true, "link is asymmetric (uplink << downlink); drives the qdisc recommendation")
		cpuBound := fs.Bool("cpu-constrained", false, "gateway is CPU-constrained; drives the qdisc recommendation")
		_ = fs.Parse(os.Args[2:])
		sinks := ResolveSinks()
		if *target == "" {
			*target = sinks.Target
		}
		if *small == "" {
			*small = sinks.SmallURL
		}
		if *bulk == "" {
			*bulk = sinks.envOrBulk()
		}
		p := audit.Params{
			AuditID: fmt.Sprintf("audit-%d", time.Now().Unix()),
			Site:    *site, LinkType: *linkType, Provider: *provider,
			Duration: *duration, Target: *target, Direction: *direction, Flows: *flows,
			SmallURL: *small, BulkAddr: *bulk,
		}
		fmt.Printf("audit %s — %s %s %ds sens %s flux %d (banc %s) → %s (bulk %s)\n",
			p.AuditID, p.LinkType, p.Site, p.Duration, p.Direction, p.Flows, sinks.Mode, p.Target, p.BulkAddr)
		res, err := audit.Run(context.Background(), p, audit.Deps{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit:", err)
			os.Exit(1)
		}
		telPath, err := audit.WriteTelemetry("data", res)
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit telemetry:", err)
			os.Exit(1)
		}
		if err := audit.AppendLinkAudit("data", res); err != nil {
			fmt.Fprintln(os.Stderr, "audit write:", err)
			os.Exit(1)
		}
		link := remediate.Link{Interface: *iface, RouterOSInterface: *rosIface}
		rep, err := audit.BuildReport(res, link, *qdisc, *asymmetric, *cpuBound)
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit report:", err)
			os.Exit(1)
		}
		_ = rep.SealTelemetry(telPath)
		if err := rep.SealSource(filepath.Join("data", "link_audit.csv")); err != nil {
			fmt.Fprintln(os.Stderr, "audit report seal:", err)
			os.Exit(1)
		}
		jsonPath, mdPath, err := audit.WriteArtifacts(rep, "data")
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit artifacts:", err)
			os.Exit(1)
		}
		fmt.Printf("rapport → %s · %s\n", jsonPath, mdPath)
		printAuditSummary(rep)
		// Code de sortie 2 : audit PARTIEL (entrées indisponibles liées à la
		// charge) — le rapport existe, la citation doit dire ce qui manque.
		partial := false
		for _, u := range rep.Unavailable {
			if u.Metric == "bufferbloat grade" || u.Metric == "economics" || u.Metric == "throughput" {
				partial = true
			}
		}
		if partial {
			os.Exit(2)
		}
```

Extend `printAuditSummary` (same file) to the 80×24 contract:

```go
// printAuditSummary — table 80 colonnes, lisible en 24 lignes : note, Δp95,
// QDI, drain, VoIP R, JFI, SLA CAP, sens, Ar/h primaire + B2B, sceau, scripts.
func printAuditSummary(rep *audit.Report) {
	line := strings.Repeat("─", 78)
	fmt.Println(line)
	fmt.Printf(" AUDIT %-30s %-8s flux:%d sens:%s\n", rep.AuditID, rep.Grade, rep.Measured.FlowCount, rep.Measured.Direction)
	fmt.Println(line)
	fmt.Printf(" Δp95 %6.1f ms   QDI %6.1f ms   drain %6.1f ms\n", rep.Measured.BloatDeltaMs, rep.Measured.QDIMs, rep.Measured.DrainMs)
	fmt.Printf(" VoIP R %4.1f        JFI %5.1f %%      débit %6.2f Mbit/s\n", rep.Measured.VoIPR, rep.Measured.JFIPct, rep.Measured.ThroughputMbps)
	if rep.CAPSLA.Measured {
		fmt.Printf(" SLA CAP (220 ms) %5.1f %%  %s\n", rep.CAPSLA.OKPct, map[bool]string{true: "MANQUEMENT", false: "conforme"}[rep.CAPSLA.Breach])
	}
	if rep.Economics != nil && rep.Economics.Primary != nil {
		fmt.Printf(" Coût primaire (%s) %10.0f Ar/h\n", rep.Economics.Primary.Name, rep.Economics.Primary.ARPerHour)
		for _, s := range rep.Economics.Scenarios {
			if s.Model == "B" {
				fmt.Printf("   B2B %-22s %10.0f Ar/h\n", s.Name, s.ARPerHour)
			}
		}
	}
	if rep.Provenance.TelemetrySHA256 != "" {
		fmt.Printf(" sceau %s…\n", rep.Provenance.TelemetrySHA256[:16])
	}
	if rep.Remediation != nil {
		fmt.Println(" scripts : COPIER LINUX TC / COPIER ROUTEROS → rapport " + rep.Provenance.SourceFile)
	}
	fmt.Println(line)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/cgo/ -v && go build ./...`
Expected: PASS, build clean. Manually smoke on Windows: `./cgo.exe audit --site TEST --link-type other --duration 20` (external mode, no bulk) → exits 2 with « attention » note and a report whose Not-measured section names the absences.

- [ ] **Step 5: Commit**

```bash
git add cmd/cgo/
git commit -m "cli: surface d'audit simplifiée, détection de puits, sortie 2 partiel, résumé 80×24"
```

---

### Task 11: API — direction field and report passthrough

**Files:**
- Modify: `pkg/api/server.go:1004` (`POST /api/audit/start` handler)
- Modify: `pkg/api/server_test.go` (audit handler test if present; add one)

**Interfaces:**
- Consumes: Task 6 (`audit.Params.Direction/Flows`), Task 10 concepts (sinks are resolved server-side on Linux; on Windows observe-mode the handler already degrades).
- Produces: request field `direction` (default `up`) and `flows` (default 4); the response/`/api/audit/latest` JSON now carries `economics`, `cap_sla`, `jfi`, `direction` because `Report` is serialized as-is.

- [ ] **Step 1: Write the failing test**

Add to `pkg/api/server_test.go` (follow the existing handler test pattern — fake `audit.Deps` via whatever injection the server uses; if the handler calls `audit.Run` directly with `audit.Deps{}`, add the test at the `Params` level instead and assert the handler copies `direction` through):

```go
func TestAuditStartDirectionParam(t *testing.T) {
	// La requête doit accepter direction/flows et les passer à l'audit.
	// (Test au niveau Params : le handler ne peut pas lancer d'audit réel ici.)
	p := auditParamsFromRequest(map[string]any{
		"site": "S", "link_type": "4g", "duration": 30,
		"direction": "down", "flows": 8,
	})
	if p.Direction != "down" || p.Flows != 8 {
		t.Fatalf("params = %+v", p)
	}
	q := auditParamsFromRequest(map[string]any{"site": "S"})
	if q.Direction != "up" || q.Flows != 4 {
		t.Fatalf("défauts up/4 : %+v", q)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/api/ -run TestAuditStartDirection -v`
Expected: FAIL — `undefined: auditParamsFromRequest`.

- [ ] **Step 3: Write minimal implementation**

In `pkg/api/server.go`, extract the handler's inline `audit.Params{...}` construction into:

```go
// auditParamsFromRequest — le mapping requête → audit.Params, défauts up/4.
func auditParamsFromRequest(body map[string]any) audit.Params {
	get := func(k string) string {
		if v, ok := body[k].(string); ok {
			return v
		}
		return ""
	}
	num := func(k string, def int) int {
		if v, ok := body[k].(float64); ok && v > 0 {
			return int(v)
		}
		return def
	}
	dir := get("direction")
	if dir == "" {
		dir = "up"
	}
	return audit.Params{
		AuditID:  fmt.Sprintf("audit-%d", time.Now().Unix()),
		Site:     get("site"), LinkType: get("link_type"), Provider: get("provider"),
		Duration: num("duration", 60), Target: get("target"),
		Direction: dir, Flows: num("flows", 4),
		SmallURL: get("small_url"), BulkAddr: get("bulk_addr"),
	}
}
```

and have the handler call it (keeping its existing sinks/degradation logic — on the Linux host it fills `BulkAddr` from the detected bench exactly like Task 10's `ResolveSinks`, via a shared call; on Windows observe-mode it stays empty and the audit reports honestly).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/api/ -v && go build ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/api/
git commit -m "api: audit accepte direction/flows, défauts up/4, rapport v2 passé au client"
```

---

## Plan 1 self-review (filled by the plan author — keep in the document)

**Spec coverage (Plan 1 slice):** §4 Models A/B/C → Tasks 1–4; decision 9 TCP_INFO → Task 5; §5 direction/flows/JFI → Task 6; CAP SLA + CSV columns → Task 7; telemetry + dual seals → Task 8; report v2 economics → Task 9; parameter simplification + exit codes + 80×24 → Task 10; API exposure → Task 11. Not in Plan 1 (by design): bench/campaign/doctor (Plan 2), UI/`/api/cost/enterprise` (Plan 3), book (Plan 4), VM proofs (Plan 5).

**Type consistency:** `metrics.RetranRatio` (Task 4) vs Task 6's `Result.RetransRatio metrics.RetranRatio` — one type, correct. `LoadOutcome` (Task 6) consumed only inside audit. `CAPSLA` defined once in metrics (Task 3), re-used verbatim in `Report.CAPSLA` (Task 9). `TelemetryRow` (Task 8) defined in telemetry.go, referenced in `Result` (audit.go) — single definition.

**Known integration notes for the executor:** Task 6 renames `Deps.Bulk` → `Deps.BulkLoad`; every constructor in `pkg/audit/*_test.go` and any external caller (`grep -rn "audit.Deps{" --include="*.go" .`) must be updated in the same commit. Task 9 removes `Report.Waste`; `grep -rn "\.Waste" --include="*.go" .` must come back empty before committing.
