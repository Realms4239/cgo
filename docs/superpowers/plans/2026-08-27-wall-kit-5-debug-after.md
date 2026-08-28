# Wall-kit 5 — Debug-After Whole Execution (Both Lineages Triple-Provenance) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Heavy debug-after whole execution — both lineages Lab 36 P1/P2/P3 + audit 30s frozen triple-provenance (Wall drawer bar + Archives SVG + Report md same hash 8-char) proven via playwright clip + 12 shots + embed analysis.

**Architecture:** After Layers 1–4 build green, prove `Tableau 9/10 Figure 5/6` triple lineage via `kit/logs` + `playwright trace on` + `screenshots 12` + `go test -tags=real + embed analysis` double surgery.

**Tech Stack:** Playwright 1.62 trace clip, screenshots 1920/1366/390, go test -tags=real, figures Generate bar+scatter RDF, report export, verify manifest

## Global Constraints

- Both lineages Lab 36 P1/P2/P3×3×2×3→12 groups + audit 30s split 3-window real 20→375 → data/frozen-wave3 manifest sha chain quarantine 4
- Triple-provenance hash 8-char same in Wall drawer + Archives SVG + Report md + Footer ticker
- Playwright wall/sheet/archives clip 8/8 + screenshots 12 + embed analysis dist sha==buildSha==integrity sha8
- Full ledger C: smoke P2×2 reps → full 36 overnight → audit → hardware translate → figures regen → report → verify
- Inline not subagents, kit/logs gitignored

---

### Task 5.1: Figures Triple-Provenance (Bar+Scatter RDF)

**Files:**
- Modify: `pkg/figures/figures.go:17`
- Modify: `pkg/api/server.go:94` report export md/csv hash
- Test: `pkg/figures/figures_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestFiguresRDF(t *testing.T){
  groups:=[]results.Group{{Profile:"P1",Qdisc:"pfifo_fast",Smallp95Median:738.4,Best:true}}
  if err:=figures.Generate("testdata/runs","testdata/figures"); err==nil {t.Fatal("should need data")}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/figures -run TestFiguresRDF -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
func Generate(dataDir,outDir string) error {
  groups,_:=results.Scan(dataDir,"")
  if len(groups)==0 {return fmt.Errorf("no data")}
  meta:=provenanceMeta(dataDir) // sha256 data/runs/*/aqm_eval.csv + date + creator
  writeBar(filepath.Join(outDir,"small_p95.svg"),groups,meta)
  writeScatter(filepath.Join(outDir,"scatter.svg"),groups,meta)
  return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/figures -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/figures/figures.go pkg/api/server.go
git commit -m "feat(figures): bar+scatter RDF triple-provenance hash 8-char"
```

### Task 5.2: Playwright Trace + 12 Shots + Embed Analysis

**Files:**
- Create: `web/frontend/e2e/wall.spec.ts`
- Create: `web/frontend/e2e/sheet.spec.ts`
- Create: `web/frontend/e2e/archives.spec.ts`
- Modify: `web/frontend/playwright.config.ts` trace on, video retain-on-failure
- Test: `npx playwright test --project=chromium`

- [ ] **Step 1: Write failing test**

```ts
import {test,expect} from '@playwright/test'
test('wall clip', async({page})=>{
  await page.goto('/')
  await expect(page.locator('[data-testid="live-wall-overlay"]')).toBeVisible()
  await expect(page.locator('[data-metric="small_p95"]')).toBeVisible()
  await page.screenshot({path:'shots/wall-hero-1920.png', clip:{x:0,y:0,width:1920,height:400}})
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx playwright test wall -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```ts
// ensure LiveView renders data-testid live-wall-overlay with baseline grey dashed vs cyan solid + hash pill
// MetricCard data-metric attr + clipPath, Rail data-pinned, Sheet data-testid
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx playwright test`
Expected: PASS 8/8 + SHOT_DIR=shots npx playwright test screenshots 12 shots

- [ ] **Step 5: Commit**

```bash
git add web/frontend/e2e/ web/frontend/playwright.config.ts
git commit -m "test(e2e): wall/sheet/archives clip + trace on + 12 shots embed"
```

### Task 5.3: Both Lineages Frozen + Report + Verify

**Files:**
- Modify: `pkg/campagne/writer.go:141` quarantine only invalid
- Test: `go test -tags=real ./pkg/campagne -v`

- [ ] **Step 1: Write failing test**

```go
func TestBothLineages(t *testing.T){
  // after POST /api/run/start {profiles:["P2"],reps:2} smoke + audit 30s, GET /api/results groups len 12
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash kit/engine.sh --action deploy --config kit/cgo-vm.yaml && curl http://192.168.174.128:9090/api/results`
Expected: FAIL available:false

- [ ] **Step 3: Write minimal implementation**

```bash
curl -X POST http://192.168.174.128:9090/api/run/start -d '{"profiles":["P1","P2","P3"],"reps":3}' # 36 overnight
./cgo audit --site "Département X" --duration 30 && curl http://192.168.174.128:9090/api/audit/list
curl -X POST http://192.168.174.128:9090/api/figures/regen && ls data/figures/*.svg
curl http://192.168.174.128:9090/api/report/export?format=md > report.md
./cgo verify && echo "hash8 $(sha256sum data/runs/*/aqm_eval.csv | cut -c1-8)"
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -tags=real ./... -v`
Expected: PASS groups 12 hash8 same across Wall/Drawer/Archives/Report

- [ ] **Step 5: Commit**

```bash
git add pkg/campagne/writer.go data/frozen-wave3/
git commit -m "feat(provenance): both lineages Lab 36 + audit 30s frozen triple hash"
```
