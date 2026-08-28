# Wall-kit 1 — Backend Robust (P3 Heterogenous + Audit Split + Leaf-only) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make backend meaningfully produce both lineages Lab P1/P2/P3×3×2×3 + Edge audit 30s 3-window split with receiver-side leaf-only goodput and single hardware.go source.

**Architecture:** Keep TCRunner handle 1: pluggable (parent 1: if netem exists else root 1:), StatsFn leaf-only, Probe BulkSendTo DialWithCC per socket, Metrics pure Go, Gates quarantined only invalid, P3 import-driven.

**Tech Stack:** Go 1.25, pkg/qdisc TCRunner FakeRunner NsRunner, pkg/probe Ping Small Bulk, pkg/metrics Percentile Summarize, pkg/campagne Matrix Writer, pkg/results Scan

## Global Constraints

- Build-then-embed TOTAL 600KB watchdog echarts 350KB
- Metrics Go pure, browser never derives, deadline 1000 fixed
- Gates only invalid quarantined, tiny mono rail only, hardwareRec per profile best single hardware.go
- qdisc stacked parent 1: tbf 80/20/5 Mbit leaf-only SumBytes last, SumDrops sum, instrumentation tag logs tc -s
- Bulk per-CC TCP_CONGESTION 0x0d Linux, receiver tc -s primary fallback sender, audit 3 windows + iperf3 optional
- P3 VSAT 5/600/30/1 + distribution normal, profile import JSON/CSV, decently portable not zero-dep
- Inline not subagents

---

### Task 1.1: P3 Heterogenous + Profile Import

**Files:**
- Create: `pkg/profile/profile.go` already, modify to support P3
- Modify: `pkg/model/types.go:14` add P3
- Modify: `pkg/api/server.go:264` handle CSV import
- Test: `pkg/profile/profile_test.go`

**Interfaces:**
- Consumes: `model.Profiles map`
- Produces: `POST /api/profile/import {id,capacity_mbps,delay_ms,jitter_ms,loss_pct} → model.Profiles[p.ID]=p`

- [ ] **Step 1: Write failing test**

```go
func TestImportP3(t *testing.T){
  if _, ok:=model.Profiles["P3"]; ok {t.Fatalf("P3 exists already")}
  p:=model.Profile{ID:"P3",CapacityMbps:5,DelayMs:600,JitterMs:30,LossPct:1}
  if err:=profile.Import(p); err!=nil {t.Fatal(err)}
  if _, ok:=model.Profiles["P3"];!ok {t.Fatal("P3 not imported")}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/profile -run TestImportP3 -v`
Expected: FAIL P3 not imported

- [ ] **Step 3: Write minimal implementation**

```go
// pkg/model/types.go
Profiles["P3"]={ID:"P3",CapacityMbps:5,DelayMs:600,JitterMs:30,LossPct:1}
// pkg/profile/profile.go func Import(p model.Profile) error { model.Profiles[p.ID]=p; return nil }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/profile -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/model/types.go pkg/profile/profile.go
git commit -m "feat(backend): P3 VSAT heterogenous + profile import"
```

### Task 1.2: Qdisc Stacked + Stats Leaf-only Robust

**Files:**
- Modify: `pkg/qdisc/qdisc.go:62`
- Modify: `pkg/qdisc/stats.go:99`
- Test: `pkg/qdisc/qdisc_test.go` golden, `pkg/qdisc/stats_test.go` leaf-only

- [ ] **Step 1: Write failing test**

```go
func TestStatsLeafOnly(t *testing.T){
  stats:=[]Stats{{Kind:"netem",Bytes:100},{Kind:"tbf",Bytes:80},{Kind:"fq_codel",Bytes:77}}
  if SumBytes(stats)!=77 {t.Fatalf("leaf not 77 got %d",SumBytes(stats))}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/qdisc -run TestStatsLeafOnly -v`
Expected: FAIL if sums

- [ ] **Step 3: Write minimal implementation**

```go
func SumBytes(s []Stats) uint64 { if len(s)==0 {return 0}; return s[len(s)-1].Bytes }
func ApplyShaper(r TCRunner, iface string, q model.Qdisc, cap, rtt float64) error {
  if _, err:=r.Run("qdisc","replace","dev",iface,"parent","1:","handle","10:","tbf","rate",fmt.Sprintf("%gmbit",cap)); err==nil {return nil}
  _, err:=r.Run("qdisc","replace","dev",iface,"root","handle","1:","tbf","rate",fmt.Sprintf("%gmbit",cap)); return err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/qdisc -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/qdisc/qdisc.go pkg/qdisc/stats.go
git commit -m "fix(qdisc): stacked parent 1: fallback + leaf-only goodput 77.7 valid G4"
```

### Task 1.3: Audit Split 3-Window + iperf3

**Files:**
- Modify: `pkg/audit/audit.go:51`
- Test: `pkg/audit/audit_test.go` split windows

- [ ] **Step 1: Write failing test**

```go
func TestAuditSplit(t *testing.T){
  p:=Params{Duration:30, Target:"1.1.1.1"}
  res,_:=Run(context.Background(),p,Deps{Ping:func(_ context.Context,_ string,_ int)[]float64{return []float64{20,21}}, Small:func(_ context.Context)(float64,error){return 50,nil}})
  if res.RTTIdleP50==res.RTTLoadedP50 {t.Fatal("idle==loaded not split")}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/audit -run TestAuditSplit -v`
Expected: FAIL idle==loaded

- [ ] **Step 3: Write minimal implementation**

```go
// split 0-12s idle, 12-22s bulk, 22-30s loaded
idleRTT, idleSmall := collect(12)
if deps.Bulk!=nil {go bulk}; else try iperf3
loadedRTT, loadedSmall := collect(8)
rSummaryIdle:=metrics.Summarize(idleRTT); rSummaryLoaded:=metrics.Summarize(loadedRTT)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/audit -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/audit.go
git commit -m "feat(audit): 3-window idle→bulk→loaded + iperf3 fallback"
```

### Task 1.4: Hardware Single Source + Gates Hidden

**Files:**
- Create: `pkg/results/hardware.go` already, ensure single
- Modify: `pkg/results/results.go:131` dedupe
- Modify: `pkg/campagne/campagne.go:233` gates
- Test: `pkg/results/hardware_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestHardwareSingle(t *testing.T){
  if HardwareRecommendation("fq_codel","P2")!="" && !strings.Contains(HardwareRecommendation("fq_codel","P2"),"MikroTik") {t.Fatal("not MikroTik")}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/results -run TestHardwareSingle -v`
Expected: FAIL if drift

- [ ] **Step 3: Write minimal implementation**

```go
func HardwareRecommendation(q,p string) string {
  if q=="fq_codel"||q=="cake" {return fmt.Sprintf("Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour %s — %s prouvé en lab",p,q)}
  return fmt.Sprintf("Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour %s — pilote isolé d'abord",p)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/results -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/results/hardware.go pkg/results/results.go pkg/campagne/campagne.go
git commit -m "fix(results): hardware single source + gates hidden C"
```
