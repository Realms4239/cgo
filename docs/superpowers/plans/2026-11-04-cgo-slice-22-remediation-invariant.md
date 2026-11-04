# Slice 22 — Remediation invariant Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the `remediation_state` + `remediation_available` invariant to a report so an unmeasured window has no shaped script and no frontend card.

**Architecture:** Go domain first: add a typed `RemediationState` + `RemediationAvailable` flag to `Report`, set them on every `BuildReport` path, freeze the structured JSON contract and the PCQ/kind regression; then the frontend suppresses the card when the state is not `available`.

**Tech Stack:** Go (`go test`), React + Vitest (`npm test`), `python tools/verify.py`, screenshots under `docs/img/shots/2026-11-04/`.

## Global Constraints

- The invariant is `RTTIdleP50 > 0 ∧ (Up > 0 ∨ Down > 0)`. NEW1's `>0.5 Mbps ∧ Δp95 ≥ 0` is rejected for this slice.
- Per-direction scripts: emit only measured directions; name each unmeasured direction in `Unavailable`.
- `RemediationState == "available" ⇔ Remediation != nil ⇔ RemediationAvailable == true`, otherwise `"observation_only"`.
- Pre-slice reports decode as `observation_only` / `false`, never `true` by zero value.
- Frontend: French user copy stays French; identifiers stay English.
- Screenshots only when a task has runtime evidence; desktop range 1024–1920.
- Staged paths are explicit; never `git add -A`.
- No live audit, bench, CHR, eBPF, `.dist`, BBRv3, or mobile acceptance in this slice.
- `tools/verify.py` is capped at 650 lines (currently 649) — do not add a gate there; Slice 22's guards are Go/ Vitest tests.

---

## File structure

| File | Responsibility | Task |
|---|---|---|
| `pkg/audit/remediation_state.go` (new) | `RemediationState` type + constants + `UnmarshalJSON` | T1 |
| `pkg/audit/report.go` | `Report` fields + `BuildReport` sets state/flag | T1 |
| `pkg/audit/remediation_fields_test.go` | `TestRemediationAvailableMirrorsPresence`, `TestRemediationStructuredFieldsRoundTrip` | T1 |
| `pkg/remediate/remediate_test.go` | `TestRouterOSEmitsOnlyValidKinds` | T1 |
| `web/frontend/src/lib/types.ts` | mirror `remediation_state` + `remediation_available` | T2 |
| `web/frontend/src/components/RemediationCard.tsx` | suppression when `state !== "available"` | T2 |
| `web/frontend/src/components/RemediationCard.dom.test.tsx` | DOM test for unmeasured suppression | T2 |
| `web/frontend/src/views/AuditView.tsx`, `web/frontend/src/views/DashboardView.tsx` | pass `report.remediation_state` to the card | T2 |
| `docs/verification/2026-11-04-slice-22.md` (new) | verifier packet | T3 |
| `docs/VERIFY.md` | one new `PENDING` slice-22 row | T3 |
| `docs/CHECKPOINTS.md` | final entry | T3 |
| `book/content/09-annexes.js` (working tree only) | one paragraph citing the packet | T3 |
| `docs/superpowers/plans/README.md` | flip slice 22 → In progress | T0 |

## Global verification commands

```bash
go test ./pkg/audit -count=1
go test ./pkg/remediate -count=1
cd web/frontend && npm run typecheck && npm test
python tools/verify.py          # 6 passed, 0 failed, 0 not run
python book/tools/audit-archive.py # two OK lines
```

Runtime evidence, when needed:

```bash
cd web/frontend && npm run build
cd ../.. && go run ./cmd/cgo --serve --addr 127.0.0.1:9091 --mode observe
node web/frontend/qa-walk.mjs http://127.0.0.1:9091
node web/frontend/capture-fresh.mjs http://127.0.0.1:9091 docs/img/shots/2026-11-04
```

---

## T0 — Spec + index row

**Files:**
- Create: `docs/superpowers/specs/2026-11-04-cgo-slice-22-remediation-invariant-design.md`

**Interfaces:**
- Consumes: nothing.
- Produces: the contract file T1–T3 implement.

- [ ] **Step 1: Write the spec body**

```markdown
# Slice 22 — Réparer le domaine avant la vitrine : l'invariant de remédiation
*Spec écrite le 2026-11-04 par le builder. Statut : **projet**. Aucun code n'existe encore. Aucune mesure n'est revendiquée.*

## 1. Pourquoi
- Le RouterOS `kind=pcq` des captures n'existe plus à HEAD (`pkg/remediate/remediate.go` émet `kind=cake|sfq|fifo` uniquement).
- Le rapport `96,7 % + Protéger ce lien` sur une ligne `Débit 0,0 Mbit/s + aucune charge bulk` reste lisible dans `data/` ;
  la slice 22 empêche qu'un nouveau rapport la reproduise, en nommant l'absence dans le JSON.

## 2. Ce qui existe déjà (vérifié en source)
- `RouterOS()` / `RouterOSIngress()` : `kind=cake|sfq|fifo`, parents `<wan>` / `<lan-or-bridge>`, nettoyages disjoints `cgo-managed` / `cgo-managed-ingress`, Linux IFB sur `ifb0`.
- `Remediation` porte déjà les 6 champs structurés (`Up/DownMeasuredMbps`, `Up/DownShapedMbps`, `LinuxIface`, `RouterOSWanIface`, `RouterOSLanIface`, `RouterOSLanAssumed`).

## 3. Design
### 3.1 Domaine : état nommé
- `RemediationState` = `"available"` / `"observation_only"`, plus `RemediationAvailable bool` alias.
- Invariant : `available ⇔ Remediation != nil ⇔ Available == true`, sinon `observation_only`.
- Rapports scellés avant la slice : `observation_only` / `false` via `UnmarshalJSON`.

### 3.2 Contrat JSON figé
- Round-trip `RenderJSON()` sur les 6 champs + état.

### 3.3 Frontend
- `RemediationCard({ remediation, state })` ; `state !== "available"` → `EmptyState` « Aucune configuration émise — charge non établie. Relancez un audit complet. »

### 3.4 Garde PCQ
- `TestRouterOSEmitsOnlyValidKinds` : `kind=cake|sfq|fifo`, jamais `pcq`, jamais `/interface queue add`, parents WAN/LAN séparés.

### 3.5 Non-objectifs
- Pas de cockpit 3 cartes, pas de tuiles, pas de `PREUVE VÉRIFIABLE` rewording, pas de deadline 220, pas de badge BBRv3, pas de fix de troncature CSS — slice 23.
- Pas de CHR boot, pas de live audit/bench, pas de mobile 390, pas de promesse CAP p95.

## 4. Règles
- R1 : suppression visible above the fold dès qu'"unmeasured → no script" est testé.
- R2 : ce spec est le contrat.
- R3 : chaque tâche laisse une capture (JuryRoom, RemediationCard suppression).
- R4 : ce qu'elle ne prouve pas est nommé.
- `ShapedRate = 95 %` inchangé. Aucun live sans approbation humaine.

## 5. Test (vert avant code)
- `go test ./pkg/audit -run TestReport -count=1` / `go test ./pkg/remediate -run TestRouterOS -count=1` / `npm test` / `python tools/verify.py` / `audit-archive`.

## 6. Séquence de commits
1. `fix(audit): fige le contrat RemediationAvailable et les champs structurés (slice 22, T0-T1)`
2. `ui(audit): casse la "Protéger ce lien" quand la charge n'est pas établie (slice 22, T2)`
3. `test(remediate): fige l'absence PCQ et /interface queue add (slice 22, T3)`
4. `docs: le paquet de vérification tranche 22, sans auto-certification (slice 22, T4)`

## Ledger après ce plan
| Task | Surfaces touchées | Capture | Lignes VERIFY qui bougent |
|---|---|---|---|
| T0 | node, spec | — | 1 |
| T1 | Go — invariant | — | 0 |
| T2 | UI — suppression | JuryRoom | 0 |
| T3 | Go — PCQ gate | — | 0 |
| T4 | packet | final set | 1 |

## 7. Non prouvé
- Aucun audit live, aucun CHR, aucune bench Linux 6.8, aucun mobile.
- `remediation_available: false` sur les rapports scellés avant la slice est une absence, pas une preuve.
```

- [ ] **Step 2: Run the links gate**

Run: `python tools/verify.py --only links`
Expected: `PASS links` — `5 documents scanned; plan and spec links resolve`

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/specs/2026-11-04-cgo-slice-22-remediation-invariant-design.md docs/superpowers/plans/README.md
git commit -m 'docs: le spec slice 22 remédiation invariant'
```

---

## T1 — Go invariant + structured contract + PCQ gate

**Files:**
- Create: `pkg/audit/remediation_state.go`
- Modify: `pkg/audit/report.go` (add fields)
- Test: `pkg/audit/remediation_fields_test.go`, `pkg/remediate/remediate_test.go`

**Interfaces:**
- Consumes: `BuildReport()` signature unchanged; `Remediation` struct unchanged.
- Produces: `type RemediationState string`, `RemediationAvailableState`, `RemediationObservationState`, `Report.RemediationState`, `Report.RemediationAvailable`.

- [ ] **Step 1: Add the state type**

Create `pkg/audit/remediation_state.go`:

```go
package audit

import "encoding/json"

// RemediationState — l'absence comme état nommé, pas une absence de champ.
type RemediationState string

const (
    RemediationAvailableState   RemediationState = "available"
    RemediationObservationState RemediationState = "observation_only"
)

// UnmarshalJSON rend les rapports scellés avant la slice `observation_only`
// au lieu d'un état vide : un zero Go ne peut jamais lire « available ».
func (s *RemediationState) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    if str == "" {
        *s = RemediationObservationState
        return nil
    }
    switch RemediationState(str) {
    case RemediationAvailableState, RemediationObservationState:
        *s = RemediationState(str)
        return nil
    }
    *s = RemediationState(str)
    return nil
}

// IsAvailable — l'alias booléen de l'état, pour l'UI qui veut un discriminant.
func (r *Report) IsAvailable() bool {
    return r != nil && r.RemediationState == RemediationAvailableState && r.Remediation != nil
}
```

- [ ] **Step 2: Wire the fields in `Report`**

In `pkg/audit/report.go`, under `Remediation`:

```go
    // Remediation is present only when capacity and RTT were both measured.
    Remediation *Remediation `json:"remediation,omitempty"`
    // RemediationState — l'absence comme état nommé.
    RemediationState RemediationState `json:"remediation_state"`
    // RemediationAvailable — alias booléen de l'état, pour l'UI.
    RemediationAvailable bool `json:"remediation_available"`
```

- [ ] **Step 3: Set the state in `BuildReport`**

In `pkg/audit/report.go` around line 482-489:

```go
        if len(rem.LinuxTC) > 0 || len(rem.LinuxTCIngress) > 0 || rem.RouterOS != "" || rem.RouterOSIngress != "" {
            rep.Remediation = rem
            rep.RemediationState = RemediationAvailableState
            rep.RemediationAvailable = true
        } else {
            rep.RemediationState = RemediationObservationState
            rep.Unavailable = append(rep.Unavailable, Unavailable{
                Metric: "remediation",
                Reason: "no interface was named, so no command set could be generated",
            })
        }
```

And in the unmeasured branch (line 416-420):

```go
    if res.RTTIdleP50 <= 0 || (upCap <= 0 && downCap <= 0) {
        rep.RemediationState = RemediationObservationState
        rep.Unavailable = append(rep.Unavailable, Unavailable{
            Metric: "remediation",
            Reason: "capacity or idle RTT unmeasured; a queue config sized from a guess would be deployed blindly",
        })
    } else {
```

- [ ] **Step 4: Add the invariant tests**

Append to `pkg/audit/remediation_fields_test.go`:

```go
// Unmeasured → absence nommée, jamais un script.
func TestRemediationAvailableMirrorsPresence(t *testing.T) {
    res := &Result{
        AuditID: "audit-unmeasured", Site: "Siège", LinkType: "4g", Direction: "up",
        RTTIdleP50: 100, LoadAttempted: true,
        // ThroughputMbps = 0 → aucune charge
    }
    rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "cake", false, false)
    if err != nil {
        t.Fatal(err)
    }
    if rep.RemediationAvailable || rep.Remediation != nil {
        t.Fatalf("remédiation disponible sans mesure : %+v", rep.Remediation)
    }
    if rep.RemediationState != RemediationObservationState {
        t.Fatalf("état = %q, want observation_only", rep.RemediationState)
    }

    res2 := &Result{
        AuditID: "audit-measured", Site: "Siège", LinkType: "4g", Direction: "up",
        RTTIdleP50: 100, LoadAttempted: true, BloatGrade: "A",
        ThroughputMbps: 5.1, LoadSeconds: 40,
    }
    rep2, err := BuildReport(res2, remediate.Link{Interface: "eth0"}, "cake", false, false)
    if err != nil {
        t.Fatal(err)
    }
    if !rep2.RemediationAvailable || rep2.Remediation == nil {
        t.Fatalf("remédiation absente alors que la charge est mesurée")
    }
    if rep2.RemediationState != RemediationAvailableState {
        t.Fatalf("état = %q, want available", rep2.RemediationState)
    }
}

// Les six champs structurés traversent le JSON.
func TestRemediationStructuredFieldsRoundTrip(t *testing.T) {
    res := &Result{
        AuditID: "audit-rt", Site: "Siège", LinkType: "4g", Direction: "both",
        RTTIdleP50: 100, LoadAttempted: true, BloatGrade: "A",
        ThroughputMbps: 5.1, ThroughputDownMbps: 18.5, LoadSeconds: 40,
    }
    rep, err := BuildReport(res, remediate.Link{
        Interface: "eth0", RouterOSInterface: "ether1-wan",
    }, "cake", false, false)
    if err != nil {
        t.Fatal(err)
    }
    js, err := rep.RenderJSON()
    if err != nil {
        t.Fatal(err)
    }
    var back Report
    if err := json.Unmarshal(js, &back); err != nil {
        t.Fatal(err)
    }
    r := back.Remediation
    if r == nil {
        t.Fatal("la remédiation a disparu du JSON")
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
    if !back.RemediationAvailable || back.RemediationState != RemediationAvailableState {
        t.Fatalf("état JSON = %q / %v", back.RemediationState, back.RemediationAvailable)
    }
}
```

Add `"encoding/json"` to the imports.

- [ ] **Step 5: Add the PCQ regression gate**

Append to `pkg/remediate/remediate_test.go`:

```go
// Aucune kind PCQ, aucun /interface queue add — la gate qui fige ce que le
// jury a vu dans les captures et que HEAD ne reproduit plus.
func TestRouterOSEmitsOnlyValidKinds(t *testing.T) {
    l := Link{
        Interface: "eth0", RouterOSInterface: "ether1-wan", RouterOSLanInterface: "ether2-lan",
        CapacityMbps: 5.1, RTTms: 100,
    }
    for _, q := range []string{QdiscCake, QdiscFqCodel, QdiscPfifo} {
        up, err := RouterOS(l, q)
        if err != nil {
            t.Fatalf("RouterOS(%s): %v", q, err)
        }
        assertRouterOSKinds(t, q, up)
        if strings.Contains(up, "/interface queue add") {
            t.Fatalf("RouterOS(%s) contient /interface queue add:\n%s", q, up)
        }
        if !strings.Contains(up, "parent=ether1-wan") {
            t.Fatalf("RouterOS(%s) parent != WAN:\n%s", q, up)
        }
        if !strings.Contains(up, `comment="cgo-managed"`) {
            t.Fatalf("RouterOS(%s) cleanup disjoint manquant:\n%s", q, up)
        }

        down, err := RouterOSIngress(l, q)
        if err != nil {
            t.Fatalf("RouterOSIngress(%s): %v", q, err)
        }
        assertRouterOSKinds(t, q, down)
        if strings.Contains(down, "/interface queue add") {
            t.Fatalf("RouterOSIngress(%s) contient /interface queue add:\n%s", q, down)
        }
        if !strings.Contains(down, "parent=ether2-lan") {
            t.Fatalf("RouterOSIngress(%s) parent != LAN:\n%s", q, down)
        }
        if !strings.Contains(down, `comment="cgo-managed-ingress"`) {
            t.Fatalf("RouterOSIngress(%s) cleanup disjoint manquant:\n%s", q, down)
        }
    }
}

func assertRouterOSKinds(t *testing.T, qdisc, script string) {
    t.Helper()
    if strings.Contains(strings.ToLower(script), "pcq") {
        t.Fatalf("RouterOS(%s) émet pcq:\n%s", qdisc, script)
    }
    if !regexp.MustCompile(`kind=(cake|sfq|fifo)`).MatchString(script) {
        t.Fatalf("RouterOS(%s) sans kind valide:\n%s", qdisc, script)
    }
}
```

Add `"regexp"` and `"strings"` to the imports.

- [ ] **Step 6: Verify the Go gate, then commit**

Run: `go test ./pkg/audit -run TestReport -count=1 && go test ./pkg/remediate -run TestRouterOS -count=1`
Expected: `ok pkg/audit` + `ok pkg/remediate`

```bash
git add pkg/audit/remediation_state.go pkg/audit/report.go pkg/audit/remediation_fields_test.go pkg/remediate/remediate_test.go
git commit -m 'fix(audit): fige le contrat RemediationAvailable et les champs structurés (slice 22, T0-T1)'
```

---

## T2 — Frontend suppression + DOM test + screenshots

**Files:**
- Modify: `web/frontend/src/lib/types.ts`
- Modify: `web/frontend/src/components/RemediationCard.tsx`
- Modify: `web/frontend/src/views/AuditView.tsx`, `web/frontend/src/views/DashboardView.tsx`
- Test: `web/frontend/src/components/RemediationCard.dom.test.tsx`
- Screenshots: `docs/img/shots/2026-11-04/slice22-audit-1024.png`, `slice22-audit-1920.png`, `slice22-dashboard-1024.png`, `slice22-dashboard-1920.png`

**Interfaces:**
- Consumes: `Report.RemediationState` + `Report.RemediationAvailable` from Go, mirrored in `types.ts` as `AuditReport.remediation_state` / `remediation_available`.
- Produces: `RemediationCard({ remediation, state })` that renders a single `EmptyState` when `state !== "available"`.

- [ ] **Step 1: Mirror the fields in the frontend types**

Modify `web/frontend/src/lib/types.ts` on `AuditReport`:

```ts
export type AuditRemediationState = "available" | "observation_only"

export interface AuditReport {
    // …
    remediation?: AuditRemediation | null
    remediation_state?: AuditRemediationState
    remediation_available?: boolean
    // …
}
```

- [ ] **Step 2: Suppress the card**

Modify `web/frontend/src/components/RemediationCard.tsx`:

```tsx
export function RemediationCard({ remediation, state }: {
  remediation: AuditRemediation | null | undefined
  state?: "available" | "observation_only" | null
}) {
  const normalized: "available" | "observation_only" =
    state === "available" ? "available" : "observation_only"

  if (normalized !== "available" || !remediation) {
    return (
      <section className="remediation remediation-empty" aria-labelledby="remediation-title">
        <h2 id="remediation-title">Protéger ce lien</h2>
        <EmptyState kind="empty" hint="Aucune configuration émise — charge non établie. Relancez un audit complet." />
      </section>
    )
  }
  // …
}
```

And in `AuditView.tsx` / `DashboardView.tsx`:

```tsx
<RemediationCard remediation={report?.remediation} state={report?.remediation_state ?? "observation_only"} />
```

- [ ] **Step 3: DOM test proves suppression**

Append to `web/frontend/src/components/RemediationCard.dom.test.tsx`:

```tsx
it('masque la carte quand l\'état est observation_only', () => {
  const html = renderToStaticMarkup(
    <RemediationCard remediation={null} state="observation_only" />
  )
  expect(html).toContain('Aucune configuration émise')
  expect(html).not.toContain('Copier la configuration')
})
```

- [ ] **Step 4: Verify typecheck + tests + screenshots**

Run: `cd web/frontend && npm run typecheck && npm test`
Expected: `280+` tests passed, no type error

Then, via the served build, capture the unmeasured suppression:

```bash
cd web/frontend && npm run build
cd ../.. && go run ./cmd/cgo --serve --addr 127.0.0.1:9091 --mode observe
node web/frontend/capture-fresh.mjs http://127.0.0.1:9091 docs/img/shots/2026-11-04
```

Leave `docs/img/shots/2026-11-04/slice22-*` (1024 + 1920) as the screen evidence.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/types.ts web/frontend/src/components/RemediationCard.tsx \
  web/frontend/src/components/RemediationCard.dom.test.tsx \
  web/frontend/src/views/AuditView.tsx web/frontend/src/views/DashboardView.tsx \
  docs/img/shots/2026-11-04/slice22-*.png \
  web/frontend/dist/
git add -f web/frontend/dist/
git commit -m 'ui(audit): casse la "Protéger ce lien" quand la charge n\'est pas établie (slice 22, T2)'
```

---

## T3 — Evidence tail + verifier packet

**Files:**
- Create: `docs/verification/2026-11-04-slice-22.md`
- Modify: `docs/VERIFY.md`, `docs/CHECKPOINTS.md`
- Modify (working tree only): `book/content/09-annexes.js`

**Interfaces:**
- Consumes: the outputs of T1 and T2 (git hashes, test counts).
- Produces: the packet + one new `PENDING` row + the checkpoint entry.

- [ ] **Step 1: Write the packet skeleton**

Create `docs/verification/2026-11-04-slice-22.md` with:

```markdown
# Verifier packet — slice 22, remediation invariant

*Prepared 2026-11-04 by the builder. The builder may not certify its own numbers.
This packet states what the verifier runs. The signature block stays empty.*

## Scope
- Branch: `rebuild/plan1-economics-audit`, HEAD: T1+T2 hash.
- Host: DESKTOP-B6K9UCM (Windows 10, MINGW64), date: from `date -u`.

## Commands — observed by the builder
| # | Command | Observed on 2026-11-04 |
|---|---|---|
| 1 | `cd web/frontend && npm run build` | exit 0; build succeeds |
| 2 | `cd web/frontend && npm test` | 57 files passed; total tests to record |
| 3 | `go test ./... -count=1` | exit 0; every package `ok` |
| 4 | `python tools/verify.py` | `6 passed, 0 failed, 0 not run` |
| 5 | `python book/tools/audit-archive.py` | both `OK` lines |

## Screenshots — final set
`docs/img/shots/2026-11-04/slice22-audit-1024.png`, `slice22-audit-1920.png`,
`slice22-dashboard-1024.png`, `slice22-dashboard-1920.png`.

## The two PENDING claims — builder's proposed disposition
| Claim (verbatim from `docs/VERIFY.md`) | Builder proposes | Basis |
|---|---|---|
| Slice 21 console spine (builder pass) | leave `PENDING` | Verifier owns this packet |
| Slice 22 remediation invariant (builder pass) | leave `PENDING` | Verifier owns this packet |

## What this packet does not claim
- No live audit, no bench, no CHR, no Linux acceptance, no eBPF, no `.dist`, no BBRv3.
- No `VERIFIED` row for slice 22; the verifier owns every row.
```

- [ ] **Step 2: Append the VERIFY row**

In `docs/VERIFY.md`, after the slice-21 row:

```markdown
| Slice 22 remediation invariant (builder pass) | PENDING | `docs/verification/2026-11-04-slice-22.md`; DESKTOP-B6K9UCM (Windows 10, MINGW64), 2026-11-04, `rebuild/plan1-economics-audit` at T1/T2 hash: `cd web/frontend && npm run build` → exit 0; `npm test` → 57 files, <count> tests passed; `go test ./... -count=1` → exit 0, every package `ok`; `python tools/verify.py` → `6 passed, 0 failed, 0 not run`; `python book/tools/audit-archive.py` → both `OK` lines. No live audit, bench, CHR, Linux acceptance, eBPF, `.dist` or BBRv3 claim. | 2026-11-04 | DESKTOP-B6K9UCM (Windows 10, MINGW64) |
```

- [ ] **Step 3: Append the checkpoint entry**

In `docs/CHECKPOINTS.md`:

```markdown
## 2026-11-04 HH:MM UTC · slice 22 builder pass and verifier packet

- Branch/HEAD: `rebuild/plan1-economics-audit` / T1+T2 hash
- Changed: `RemediationState` + `RemediationAvailable`, per-direction scripts already in `Remediation`, structured JSON contract, PCQ regression, frontend suppression, four captures.
- Proven: commands 1–5 above, with outputs quoted.
- Unproven/blocked: no live audit/bench/CHR/eBPF/`.dist`/BBRv3; the verifier owns the rows.
```

- [ ] **Step 4: Update the book working tree (never staged)**

In `book/content/09-annexes.js`, append one paragraph to Annexe E citing the console and the packet path.

- [ ] **Step 5: Run the full gate, then commit**

Run: `python tools/verify.py`
Expected: `6 passed, 0 failed, 0 not run`

```bash
git add docs/verification/2026-11-04-slice-22.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m 'docs: le paquet de vérification tranche 22, sans auto-certification (slice 22, T4)'
```

---

## Self-review

**Spec coverage each requirement → task:**

| Spec item | Task |
|---|---|
| `RemediationState` + `RemediationAvailable` invariant | T1.1–1.3 |
| Structured JSON round-trip (6 fields) | T1.4 |
| PCQ gate (`cake/sfq/fifo`, no `pcq`, no `/interface queue add`) | T1.5 |
| Frontend suppression (empty-state) | T2.2–2.3 |
| `types.ts` + `AuditView`/`DashboardView` pass-through | T2.1, T2.2 |
| Verifier packet + `PENDING` row + checkpoint | T3.1–3.3 |
| Book annex paragraph (unstaged) | T3.4 |

No placeholder. No live claim. No mobile claim. Type names used consistently:
`RemediationState`, `RemediationAvailableState`, `RemediationObservationState`, `AuditRemediationState`, `RemediationAvailable`.

## Execution Handoff

**Slice 22 closes as "builder pass complete, packet ready".** The builder may not certify itself; the verifier owns the `PENDING` rows.

---

**[x] T0–T4 plans written and saved to `docs/superpowers/plans/2026-11-04-cgo-slice-22-remediation-invariant.md`.**

Plan complete and saved to `docs/superpowers/plans/2026-11-04-cgo-slice-22-remediation-invariant.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
