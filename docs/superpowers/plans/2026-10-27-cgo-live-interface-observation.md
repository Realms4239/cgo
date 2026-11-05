# Live Interface Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a real 10 Hz interface-throughput observation path to the live wall: a `/proc/net/dev` sampler driven by the existing hub clock, additive `iface_*` frame fields, and a raw RX/TX chart with real timestamps and explicit gaps.

**Architecture:** A new leaf package `pkg/iface` parses cumulative kernel counters and turns two reads into one windowed rate with measured `dt`. `cmd/cgo` composes the campaign snapshot and the interface sample into one JSON frame; the hub calls `Read()` and `/api/state` calls `Latest()`, so state polls cannot shorten an SSE window. The frontend keeps a separate 600-point interface ring, pushes it before the campaign `measured` gate, and paints it at 10 Hz in the existing rAF loop while every other chart keeps its 250 ms gate.

**Tech Stack:** Go 1.25+ standard library only; React 19 + TypeScript + ECharts 5 + Vitest; Bash for the dual-namespace bench.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md`.
- Branch: `rebuild/plan1-economics-audit`; do not merge to `main` in this slice.
- No new Go or npm dependency. Standard library only.
- `pkg/iface` is a leaf package: it imports only the Go standard library and no project package.
- French comments and user-facing strings; English identifiers and JSON keys.
- `campagne.Snapshot`, the campaign path, `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv`, and `data/runs/` are untouched. New fields are additive at the API frame only.
- The frozen guard and its status live in `docs/VERIFY.md`, created by slice 9 Phase A. Cite that row; never restate a count from memory. Do not run or rewrite the frozen corpus.
- Never fabricate a measurement. `iface_valid=false` gates every numeric interface field; the UI must never render an invalid zero as a rate.
- No interpolation, no replay buffer, no playout buffer, no zero-fill, no LTTB on the interface series.
- No production write, control action, packet capture, or root access. Production SNMP/RouterOS is Phase 2 and out of this plan.
- Commit per logical group with explicit paths. Never `git add -A`.
- Leave `epbfBBR.md`, `data/`, `backup/`, and user artifacts untracked. The spec, this plan, and the roadmap are the only docs committed as the slice contract.
- Do not start Graylog. The remote PC `100.114.101.13` is off-limits.
- Twin bench: CGO-DEV, Ubuntu 24.04, Linux 6.8, at `/home/lab/cgo-current`. The old `/home/lab/cgo` is not a source. Local bench traffic only; the VM's NAT link is very slow.
- The browser performance trace runs on a host that can reach `http://10.99.0.2:9090`. If no such browser is available, record "not runnable here" and do not claim the paint gate.
- Statements about live behaviour are not "verified" until Task 9 passes on Linux 6.8.

## Pre-flight (before Task 1)

Execution starts only after slice 9 Phase A has run on this checkout: `docs/VERIFY.md` exists with the frozen guard row, and `docs/CHECKPOINTS.md` / `docs/ERRORS.md` exist. If any is missing, stop and run slice 9 Phase A first. Do not invent the files here.

Commit the slice contract first, explicitly:

```bash
git add docs/superpowers/specs/2026-10-27-cgo-live-interface-observation-design.md \
        docs/superpowers/plans/2026-10-27-cgo-live-interface-observation.md \
        docs/superpowers/plans/README.md
git commit -m "docs: live interface observation spec, plan, and roadmap"
```

Do not commit `epbfBBR.md`, `data/`, `backup/`, `web/frontend/dist`, or any file under `.perch/`.

## File Structure

| File | Responsibility |
|---|---|
| `pkg/iface/iface.go` | `Sample`, `Counters`, `Reader`, `NormalizeSource`, `ReadCounters`, `Sampler`, `New`. |
| `pkg/iface/iface_test.go` | Parser, missing-name, missing-file, source normalisation. |
| `pkg/iface/sampler_test.go` | Warmup, rate math, reset, slow, clock, absent, read, `Latest`, concurrency. |
| `pkg/api/server.go` | `Deps.NextSnap`; the hub reads `NextSnap`, `/api/state` reads `GetSnap`. |
| `pkg/api/next_snap_test.go` | The hub uses `NextSnap`; a state poll does not. |
| `cmd/cgo/server.go` | `liveFrame`, `envSampler`, `GetSnap`/`NextSnap` composition. |
| `cmd/cgo/server_iface_test.go` | Env resolution and disabled paths. |
| `cmd/cgo/main.go` | One usage line naming `CGO_IFACE` / `CGO_IFACE_SOURCE`. |
| `kit/run-dualns.sh` | `serve` passes `CGO_IFACE=c-cli`, `CGO_IFACE_SOURCE=twin` by default. |
| `docs/api.md` | Additive `iface_*` field contract. |
| `web/frontend/src/lib/types.ts` | Optional `iface_*` fields on `LiveFrame`. |
| `web/frontend/src/lib/live.ts` | `live.iface` ring, `pushIface`, dedupe, gap handling. |
| `web/frontend/src/lib/live.test.ts` | Ring behaviour and `clearLive` preservation. |
| `web/frontend/src/lib/sse.ts` | Push interface frames before the campaign gate. |
| `web/frontend/src/lib/chartGrammar.ts` | `rawLineSeries` — no smoothing, no LTTB, no null connection. |
| `web/frontend/src/lib/chartGrammar.test.ts` | Raw series flags. |
| `web/frontend/src/views/LiveView.tsx` | Interface card, 10 Hz paint gate, relabel the app goodput card. |
| `web/frontend/src/views/LiveView.test.ts` | Source-level assertions for the new labels and raw series. |
| `docs/superpowers/plans/README.md` | Status after execution. |

---

### Task 1: Parse `/proc/net/dev`

**Files:**
- Create: `pkg/iface/iface.go`
- Test: `pkg/iface/iface_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Counters struct { RXBytes, RXDrops, TXBytes, TXDrops uint64 }`
  - `type Sample struct { Name string; Source string; RXMbps, TXMbps float64; RXDrops, TXDrops uint64; WindowMs float64; SampleTS int64; Valid bool; Gap string }` with the JSON tags from the spec
  - `func ReadCounters(path string) (map[string]Counters, error)`
  - `func NormalizeSource(s string) string`
  - `type Reader interface { Read() Sample; Latest() Sample }`

- [ ] **Step 1: Write the failing tests**

Create `pkg/iface/iface_test.go`:

```go
package iface

import (
	"os"
	"path/filepath"
	"testing"
)

const fixture = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1000 10 0 2 0 0 0 0 2000 20 0 3 0 0 0 0
  eth0: 123456789012 100 0 4 0 0 0 0 987654321098 200 0 5 0 0 0 0
 eth:0: 18446744073709551615 7 0 8 0 0 0 0 9 1 0 2 0 0 0 0
`

func writeFile(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "dev")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadCountersParses(t *testing.T) {
	got, err := ReadCounters(writeFile(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	if c := got["lo"]; c.RXBytes != 1000 || c.RXDrops != 2 || c.TXBytes != 2000 || c.TXDrops != 3 {
		t.Fatalf("lo = %+v", c)
	}
	if c := got["eth0"]; c.RXBytes != 123456789012 || c.TXDrops != 5 {
		t.Fatalf("eth0 = %+v", c)
	}
	// le nom peut contenir ':' : on coupe au DERNIER séparateur
	if c := got["eth:0"]; c.RXBytes != 18446744073709551615 {
		t.Fatalf("eth:0 = %+v", c)
	}
}

func TestReadCountersRejectsShortLine(t *testing.T) {
	body := "a\nb\n eth0: 1 2 3\n"
	if _, err := ReadCounters(writeFile(t, body)); err == nil {
		t.Fatal("ligne courte acceptée : une ligne tronquée doit faire échouer la lecture")
	}
}

func TestReadCountersMissingFile(t *testing.T) {
	if _, err := ReadCounters(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("fichier absent accepté")
	}
}

func TestNormalizeSource(t *testing.T) {
	for in, want := range map[string]string{
		"": "unknown", "twin": "twin", "prod": "prod", "production": "unknown",
	} {
		if got := NormalizeSource(in); got != want {
			t.Fatalf("%q -> %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/iface`

Expected: FAIL — `undefined: ReadCounters`, `undefined: NormalizeSource`.

- [ ] **Step 3: Write the minimal implementation**

Create `pkg/iface/iface.go`:

```go
// Package iface — débit d'interface observé par les compteurs noyau.
// stdlib seul : /proc/net/dev sur Linux, un fichier de même format ailleurs
// (les tests Windows passent par ReadCounters avec un fichier temporaire).
package iface

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const procNetDev = "/proc/net/dev"

// Counters — compteurs cumulés d'une interface. Octets et pertes seulement ;
// les paquets ne sont pas exposés, personne ne les affiche.
type Counters struct {
	RXBytes uint64
	RXDrops uint64
	TXBytes uint64
	TXDrops uint64
}

// Sample — un échantillon de débit, champs additifs du frame live.
// Valid=false invalide tous les champs numériques : un zéro n'est jamais
// une mesure dans ce contrat.
type Sample struct {
	Name     string  `json:"iface_name"`
	Source   string  `json:"iface_source"`
	RXMbps   float64 `json:"iface_rx_mbps"`
	TXMbps   float64 `json:"iface_tx_mbps"`
	RXDrops  uint64  `json:"iface_rx_drops"`
	TXDrops  uint64  `json:"iface_tx_drops"`
	WindowMs float64 `json:"iface_window_ms"`
	SampleTS int64   `json:"iface_sample_ts"`
	Valid    bool    `json:"iface_valid"`
	Gap      string  `json:"iface_gap"`
}

// Reader — la couture que l'adaptateur SNMP/RouterOS (phase 2) implémentera.
type Reader interface {
	Read() Sample
	Latest() Sample
}

// NormalizeSource — provenance déclarée, jamais devinée. Toute valeur inconnue
// retombe sur "unknown" ; l'appelant journalise la valeur refusée.
func NormalizeSource(s string) string {
	switch s {
	case "twin", "prod", "unknown":
		return s
	}
	return "unknown"
}

// ReadCounters — parse /proc/net/dev (ou un fichier de même format).
// Deux lignes d'en-tête, puis une ligne par interface. Une ligne tronquée est
// une erreur, pas un saut : lire un fichier partiel produirait une mesure fausse.
func ReadCounters(path string) (map[string]Counters, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Counters)
	for i, line := range strings.Split(string(raw), "\n") {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}
		colon := strings.LastIndexByte(line, ':')
		if colon < 0 {
			return nil, fmt.Errorf("ligne %d : séparateur ':' absent", i+1)
		}
		name := strings.TrimSpace(line[:colon])
		f := strings.Fields(line[colon+1:])
		if len(f) < 16 {
			return nil, fmt.Errorf("ligne %d (%s) : %d champs, 16 attendus", i+1, name, len(f))
		}
		rx, err := strconv.ParseUint(f[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : rx_bytes : %w", i+1, name, err)
		}
		rxd, err := strconv.ParseUint(f[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : rx_drops : %w", i+1, name, err)
		}
		tx, err := strconv.ParseUint(f[8], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : tx_bytes : %w", i+1, name, err)
		}
		txd, err := strconv.ParseUint(f[11], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : tx_drops : %w", i+1, name, err)
		}
		out[name] = Counters{RXBytes: rx, RXDrops: rxd, TXBytes: tx, TXDrops: txd}
	}
	return out, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/iface`

Expected: PASS (`ok github.com/Realms4239/cgo/pkg/iface`).

- [ ] **Step 5: Commit**

```bash
git add pkg/iface/iface.go pkg/iface/iface_test.go
git commit -m "feat(iface): parse /proc/net/dev counters"
```

---

### Task 2: Windowed sampler with explicit gaps

**Files:**
- Modify: `pkg/iface/iface.go`
- Test: `pkg/iface/sampler_test.go`

**Interfaces:**
- Consumes: `Counters`, `ReadCounters`, `NormalizeSource`, `Sample` from Task 1.
- Produces:
  - `func New(name, source string) *Sampler`
  - `func newSampler(path, name, source string, now func() time.Time) *Sampler` (unexported test seam)
  - `func (s *Sampler) Read() Sample`
  - `func (s *Sampler) Latest() Sample`

- [ ] **Step 1: Write the failing tests**

Create `pkg/iface/sampler_test.go`:

```go
package iface

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func devLine(rx, tx, rxd, txd uint64) string {
	return "h1\nh2\n eth0: " + strconv.FormatUint(rx, 10) + " 0 0 " +
		strconv.FormatUint(rxd, 10) + " 0 0 0 0 " +
		strconv.FormatUint(tx, 10) + " 0 0 " +
		strconv.FormatUint(txd, 10) + " 0 0 0 0\n"
}

func writeDev(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func newTestSampler(t *testing.T, now *time.Time) (*Sampler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dev")
	return newSampler(path, "eth0", "twin", func() time.Time { return *now }), path
}

func TestSamplerWarmupThenRate(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(1000, 2000, 1, 2))

	first := s.Read()
	if first.Valid || first.Gap != "warmup" {
		t.Fatalf("premier échantillon = %+v, want warmup", first)
	}
	if first.SampleTS != now.UnixMilli() || first.Name != "eth0" || first.Source != "twin" {
		t.Fatalf("warmup = %+v", first)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(1000+125000, 2000+25000, 3, 4))
	got := s.Read()
	if !got.Valid || got.Gap != "" {
		t.Fatalf("échantillon = %+v, want valid", got)
	}
	if !near(got.RXMbps, 10) || !near(got.TXMbps, 2) {
		t.Fatalf("débits = %.3f/%.3f, want 10/2", got.RXMbps, got.TXMbps)
	}
	if got.WindowMs != 100 {
		t.Fatalf("fenêtre = %v, want 100", got.WindowMs)
	}
	if got.SampleTS != now.UnixMilli() {
		t.Fatalf("ts = %d, want %d", got.SampleTS, now.UnixMilli())
	}
	if got.RXDrops != 3 || got.TXDrops != 4 {
		t.Fatalf("pertes = %d/%d", got.RXDrops, got.TXDrops)
	}
}

func TestSamplerResetRebaselines(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(500000, 500000, 0, 0))
	_ = s.Read()

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(10, 10, 0, 0))
	if got := s.Read(); got.Valid || got.Gap != "reset" || got.WindowMs != 100 {
		t.Fatalf("reset = %+v", got)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(10+125000, 10+25000, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 10) {
		t.Fatalf("reprise = %+v", got)
	}
}

func TestSamplerSlowAndClock(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	_ = s.Read()

	now = now.Add(600 * time.Millisecond)
	writeDev(t, path, devLine(1250000, 0, 0, 0))
	if got := s.Read(); got.Valid || got.Gap != "slow" || got.WindowMs != 600 {
		t.Fatalf("slow = %+v", got)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(2500000, 0, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 100) {
		t.Fatalf("après slow = %+v", got)
	}

	now = now.Add(-50 * time.Millisecond)
	if got := s.Read(); got.Valid || got.Gap != "clock" || got.WindowMs != 0 {
		t.Fatalf("clock = %+v", got)
	}
}

func TestSamplerAbsentThenWarmup(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, "h1\nh2\n lo: 1 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0\n")
	if got := s.Read(); got.Valid || got.Gap != "absent" {
		t.Fatalf("absent = %+v", got)
	}

	writeDev(t, path, devLine(100, 100, 0, 0))
	now = now.Add(100 * time.Millisecond)
	if got := s.Read(); got.Gap != "warmup" {
		t.Fatalf("retour = %+v, want warmup", got)
	}
	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(100+12500, 100, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 1) {
		t.Fatalf("reprise = %+v", got)
	}
}

func TestSamplerReadError(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	_ = s.Read()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := s.Read(); got.Valid || got.Gap != "read" {
		t.Fatalf("read = %+v", got)
	}
}

func TestSamplerLatest(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	if got := s.Latest(); got.Gap != "warmup" || got.Name != "eth0" || got.Source != "twin" {
		t.Fatalf("latest initial = %+v", got)
	}
	writeDev(t, path, devLine(0, 0, 0, 0))
	first := s.Read()
	if got := s.Latest(); got != first {
		t.Fatalf("latest = %+v, want %+v", got, first)
	}
}

func TestSamplerConcurrentReadLatest(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.Read()
				s.Latest()
			}
		}()
	}
	wg.Wait()
}

func TestNewNormalizesSource(t *testing.T) {
	s := New("eth0", "production")
	if got := s.Latest().Source; got != "unknown" {
		t.Fatalf("source = %q, want unknown", got)
	}
	if got := s.Latest().Name; got != "eth0" {
		t.Fatalf("name = %q", got)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/iface`

Expected: FAIL — `undefined: New`, `undefined: newSampler`, `s.Read undefined`.

- [ ] **Step 3: Write the minimal implementation**

Add to `pkg/iface/iface.go` after `ReadCounters`, and add `math`, `sync`, `time` to the imports:

```go
type prev struct {
	c Counters
	t time.Time
}

// Sampler — un seul échantillonneur par interface, lu par le hub 10 Hz.
// Read avance la base ; Latest ne la touche jamais.
type Sampler struct {
	mu     sync.Mutex
	path   string
	name   string
	source string
	now    func() time.Time
	prev   *prev
	last   Sample
}

// New — échantillonneur de production sur /proc/net/dev.
func New(name, source string) *Sampler {
	return newSampler(procNetDev, name, source, time.Now)
}

// newSampler — le chemin et l'horloge sont injectables pour les tests Windows.
func newSampler(path, name, source string, now func() time.Time) *Sampler {
	s := &Sampler{path: path, name: name, source: NormalizeSource(source), now: now}
	s.last = Sample{Name: name, Source: s.source, Gap: "warmup"}
	return s
}

// Read — une fenêtre réelle entre deux lectures de compteurs.
// Aucun plancher d'octets : les compteurs sont exacts, un petit delta est un
// petit débit réel. Aucun clamp : un compteur qui n'a pas reculé est réel.
func (s *Sampler) Read() Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()

	all, err := ReadCounters(s.path)
	if err != nil {
		s.prev = nil
		return s.gap(now, 0, "read")
	}
	c, ok := all[s.name]
	if !ok {
		s.prev = nil
		return s.gap(now, 0, "absent")
	}
	if s.prev == nil {
		s.prev = &prev{c: c, t: now}
		return s.gap(now, 0, "warmup")
	}

	dt := now.Sub(s.prev.t)
	ms := float64(dt) / float64(time.Millisecond)
	if dt <= 0 {
		s.prev = &prev{c: c, t: now}
		return s.gap(now, 0, "clock")
	}
	if dt > 500*time.Millisecond {
		s.prev = &prev{c: c, t: now}
		return s.gap(now, round1(ms), "slow")
	}
	if c.RXBytes < s.prev.c.RXBytes || c.TXBytes < s.prev.c.TXBytes {
		s.prev = &prev{c: c, t: now}
		return s.gap(now, round1(ms), "reset")
	}

	rx := float64(c.RXBytes-s.prev.c.RXBytes) * 8 / 1e6 / dt.Seconds()
	tx := float64(c.TXBytes-s.prev.c.TXBytes) * 8 / 1e6 / dt.Seconds()
	s.prev = &prev{c: c, t: now}
	return s.emit(Sample{
		Name: s.name, Source: s.source,
		RXMbps: rx, TXMbps: tx,
		RXDrops: c.RXDrops, TXDrops: c.TXDrops,
		WindowMs: round1(ms), SampleTS: now.UnixMilli(),
		Valid: true,
	})
}

// Latest — le dernier échantillon, sans avancer la base. Utilisé par /api/state.
func (s *Sampler) Latest() Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

func (s *Sampler) gap(now time.Time, windowMs float64, reason string) Sample {
	return s.emit(Sample{
		Name: s.name, Source: s.source,
		SampleTS: now.UnixMilli(), WindowMs: windowMs, Gap: reason,
	})
}

func (s *Sampler) emit(x Sample) Sample {
	s.last = x
	return x
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/iface`

Expected: PASS. Then run the race gate: `go test -race ./pkg/iface`. If the Windows toolchain has no race support, record "not runnable here" and run the same command in Task 9 on the VM.

- [ ] **Step 5: Commit**

```bash
git add pkg/iface/iface.go pkg/iface/sampler_test.go
git commit -m "feat(iface): 10 Hz windowed sampler with explicit gaps"
```

---

### Task 3: Hub reads a fresh sample, state returns the last one

**Files:**
- Modify: `pkg/api/server.go`
- Test: `pkg/api/next_snap_test.go`

**Interfaces:**
- Consumes: existing `Deps`, `New`, `NewHub`, `Hub.Serve`.
- Produces:
  - `Deps.NextSnap func() any` — optional; nil falls back to `GetSnap`.
  - `GET /api/state` keeps using `GetSnap`.
  - the 10 Hz hub calls `NextSnap`.

- [ ] **Step 1: Write the failing test**

Create `pkg/api/next_snap_test.go`:

```go
package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Le hub 10 Hz doit lire NextSnap (échantillon frais) ; /api/state doit lire
// GetSnap (dernier échantillon). Un poll d'état ne raccourcit jamais une
// fenêtre SSE.
func TestNextSnapDrivesHubAndNotState(t *testing.T) {
	var hubCalls, stateCalls atomic.Int64
	h := New(Deps{
		GetSnap: func() any {
			stateCalls.Add(1)
			return map[string]any{"running": false, "src": "state"}
		},
		NextSnap: func() any {
			hubCalls.Add(1)
			return map[string]any{"running": false, "src": "hub"}
		},
	})
	defer h.CloseHub()
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `"src":"state"`) {
		t.Fatalf("/api/state n'a pas utilisé GetSnap : %s", body)
	}

	time.Sleep(250 * time.Millisecond)
	if got := hubCalls.Load(); got < 2 {
		t.Fatalf("hub NextSnap calls = %d, want >= 2", got)
	}
	if got := stateCalls.Load(); got != 1 {
		t.Fatalf("GetSnap calls = %d, want 1 (le hub ne doit pas l'appeler)", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./pkg/api -run TestNextSnapDrivesHubAndNotState -v`

Expected: FAIL — `unknown field NextSnap in struct literal`.

- [ ] **Step 3: Write the minimal implementation**

In `pkg/api/server.go`, add the field to `Deps`:

```go
	// NextSnap — fournisseur du hub 10 Hz. Nil ⇒ GetSnap. Le hub lit un
	// échantillon frais ; /api/state garde GetSnap (dernier échantillon).
	NextSnap func() any
```

In `New`, after the existing `snap` default block:

```go
	next := d.NextSnap
	if next == nil {
		next = snap
	}
```

Replace the `hub.Serve` call at the end of `New`:

```go
	hub.Serve(func() any { return next() })
```

Do not change the `/api/state` handler: it must keep calling `snap()`.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./pkg/api -run TestNextSnapDrivesHubAndNotState -v`

Expected: PASS. Then `go test ./pkg/api` — PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/api/server.go pkg/api/next_snap_test.go
git commit -m "feat(api): NextSnap seam for the 10 Hz hub"
```

---

### Task 4: Compose the interface sample into the live frame

**Files:**
- Modify: `cmd/cgo/server.go`
- Test: `cmd/cgo/server_iface_test.go`

**Interfaces:**
- Consumes: `iface.New`, `iface.Sampler.Read`, `iface.Sampler.Latest`, `iface.NormalizeSource`, `api.Deps.NextSnap`.
- Produces:
  - `type liveFrame struct { campagne.Snapshot; iface.Sample }`
  - `func envSampler(getenv func(string) string, goos string) *iface.Sampler`
  - `GetSnap` returns `liveFrame{Snapshot: live.Get(), Sample: sampler.Latest()}`
  - `NextSnap` returns `liveFrame{Snapshot: live.Get(), Sample: sampler.Read()}`

- [ ] **Step 1: Write the failing test**

Create `cmd/cgo/server_iface_test.go`:

```go
package main

import "testing"

func TestEnvSampler(t *testing.T) {
	env := map[string]string{}
	get := func(k string) string { return env[k] }

	if s := envSampler(get, "windows"); s != nil {
		t.Fatal("Windows sans CGO_IFACE doit désactiver l'échantillonneur")
	}
	env["CGO_IFACE"] = "c-cli"
	if s := envSampler(get, "windows"); s != nil {
		t.Fatal("Windows avec CGO_IFACE doit journaliser et désactiver")
	}

	s := envSampler(get, "linux")
	if s == nil {
		t.Fatal("Linux avec CGO_IFACE doit créer l'échantillonneur")
	}
	got := s.Latest()
	if got.Name != "c-cli" || got.Source != "unknown" || got.Gap != "warmup" {
		t.Fatalf("latest = %+v", got)
	}

	env["CGO_IFACE_SOURCE"] = "prod"
	if got := envSampler(get, "linux").Latest(); got.Source != "prod" {
		t.Fatalf("source = %q, want prod", got.Source)
	}
	env["CGO_IFACE_SOURCE"] = "production"
	if got := envSampler(get, "linux").Latest(); got.Source != "unknown" {
		t.Fatalf("source inconnue = %q, want unknown", got.Source)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/cgo -run TestEnvSampler -v`

Expected: FAIL — `undefined: envSampler`.

- [ ] **Step 3: Write the minimal implementation**

In `cmd/cgo/server.go`, add imports `runtime` and `github.com/Realms4239/cgo/pkg/iface`.

Add at file scope, near `runServer`:

```go
// liveFrame — snapshot de campagne + échantillon interface, aplatis en JSON.
// Les champs iface_* sont additifs ; absents quand CGO_IFACE est vide.
type liveFrame struct {
	campagne.Snapshot
	iface.Sample
}

// envSampler — CGO_IFACE / CGO_IFACE_SOURCE, lus une fois au démarrage.
// Pas d'auto-détection : une mauvaise interface serait une mesure mal étiquetée.
func envSampler(getenv func(string) string, goos string) *iface.Sampler {
	name := getenv("CGO_IFACE")
	if name == "" {
		return nil
	}
	if goos != "linux" {
		log.Printf("CGO_IFACE=%s ignoré — /proc/net/dev absent sur %s", name, goos)
		return nil
	}
	src := getenv("CGO_IFACE_SOURCE")
	if src != "" && iface.NormalizeSource(src) != src {
		log.Printf("CGO_IFACE_SOURCE=%q inconnu — source non déclarée", src)
	}
	return iface.New(name, src)
}
```

In `runServer`, right after `live := campagne.NewLive()`:

```go
	sampler := envSampler(os.Getenv, runtime.GOOS)
	getSnap := func() any { return live.Get() }
	nextSnap := getSnap
	if sampler != nil {
		getSnap = func() any {
			return liveFrame{Snapshot: live.Get(), Sample: sampler.Latest()}
		}
		nextSnap = func() any {
			return liveFrame{Snapshot: live.Get(), Sample: sampler.Read()}
		}
	}
```

Replace the handler's provider line:

```go
		GetSnap:   func() any { return live.Get() },
```

with:

```go
		GetSnap:   getSnap,
		NextSnap:  nextSnap,
```

Do not touch `startFn`, `stopFn`, `pumpSnapshots`, or any campaign code.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./cmd/cgo -run TestEnvSampler -v`

Expected: PASS. Then `go test ./cmd/cgo` — PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/cgo/server.go cmd/cgo/server_iface_test.go
git commit -m "feat(server): compose iface sample into the live frame"
```

---

### Task 5: Twin defaults, usage text, and API contract

**Files:**
- Modify: `kit/run-dualns.sh`
- Modify: `cmd/cgo/main.go`
- Modify: `docs/api.md`

**Interfaces:**
- Consumes: `CGO_IFACE`, `CGO_IFACE_SOURCE` from Task 4.
- Produces: no Go/TS symbol. The twin `serve` path sets the variables; the docs describe the additive fields.

- [ ] **Step 1: Wire the twin defaults**

In `kit/run-dualns.sh`, replace the `exec` block in the `up|serve` arm:

```bash
    exec $SUDO ip netns exec ns_client env CGO_DASHBOARD__ADDR=10.99.0.2:9090 \
      CGO_IFACE="${CGO_IFACE:-c-cli}" \
      CGO_IFACE_SOURCE="${CGO_IFACE_SOURCE:-twin}" \
      "$BIN" --serve --tls=false --http-addr=
```

- [ ] **Step 2: Add the usage line**

In `cmd/cgo/main.go`, inside the raw `usage()` string, after the `--serve` line:

```
  env: CGO_IFACE=<iface> CGO_IFACE_SOURCE=twin|prod|unknown   # débit interface 10 Hz
```

Keep the existing backtick raw string intact.

- [ ] **Step 3: Document the fields**

In `docs/api.md`, after the "Santé & état" table, add:

```markdown
**Champs d'interface (additifs).** Quand `CGO_IFACE` est défini côté serveur,
chaque frame `/api/state` et `/api/stream` porte en plus :
`iface_name`, `iface_source` (`twin` | `prod` | `unknown`), `iface_rx_mbps`,
`iface_tx_mbps`, `iface_rx_drops`, `iface_tx_drops`, `iface_window_ms`,
`iface_sample_ts`, `iface_valid`, `iface_gap`. `iface_valid:false` invalide
tous les champs numériques ; `iface_gap` vaut `warmup`, `absent`, `reset`,
`slow`, `read` ou `clock`. Aucun champ `iface_*` n'est présent quand
`CGO_IFACE` est vide. Le hub 10 Hz lit un échantillon frais ; `/api/state`
renvoie le dernier échantillon sans avancer la fenêtre.
```

- [ ] **Step 4: Verify the shell syntax**

Run: `bash -n kit/run-dualns.sh`

Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add kit/run-dualns.sh cmd/cgo/main.go docs/api.md
git commit -m "docs(bench): interface observation env and API fields"
```

---

### Task 6: Frontend ring and SSE push

**Files:**
- Modify: `web/frontend/src/lib/types.ts`
- Modify: `web/frontend/src/lib/live.ts`
- Modify: `web/frontend/src/lib/sse.ts`
- Test: `web/frontend/src/lib/live.test.ts`

**Interfaces:**
- Consumes: `LiveFrame` from `types.ts`.
- Produces:
  - `export type IfacePt = [number, number | null]`
  - `live.iface` with `rx`, `tx`, `name`, `source`, `valid`, `gap`, `windowMs`, `rxDrops`, `txDrops`, `ts`, `seq`, `max`
  - `export function pushIface(f: Partial<LiveFrame>): void`
  - `sse.ts` calls `pushIface(data)` for every non-replay frame.

- [ ] **Step 1: Write the failing tests**

Create `web/frontend/src/lib/live.test.ts`:

```ts
import { describe, it, expect, beforeEach } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { readFileSync } from 'node:fs'
import { live, pushIface, clearLive } from './live'
import type { LiveFrame } from './types'

function frame(ts: number, over: Partial<LiveFrame> = {}): Partial<LiveFrame> {
  return {
    iface_name: 'c-cli',
    iface_source: 'twin',
    iface_sample_ts: ts,
    iface_valid: true,
    iface_rx_mbps: 10,
    iface_tx_mbps: 2,
    iface_rx_drops: 1,
    iface_tx_drops: 2,
    iface_window_ms: 100,
    iface_gap: '',
    ...over,
  }
}

beforeEach(() => {
  const r = live.iface
  r.rx.length = 0
  r.tx.length = 0
  r.name = ''
  r.source = ''
  r.valid = false
  r.gap = ''
  r.windowMs = 0
  r.rxDrops = 0
  r.txDrops = 0
  r.ts = 0
  r.seq = 0
})

describe('pushIface', () => {
  it('ignore un timestamp identique (frame répétée)', () => {
    pushIface(frame(1000))
    pushIface(frame(1000))
    expect(live.iface.rx.length).toBe(1)
    expect(live.iface.seq).toBe(1)
  })

  it('un échantillon invalide casse la série avec null', () => {
    pushIface(frame(1000))
    pushIface(frame(1100, { iface_valid: false, iface_gap: 'absent' }))
    expect(live.iface.rx.at(-1)).toEqual([1100, null])
    expect(live.iface.tx.at(-1)).toEqual([1100, null])
    expect(live.iface.gap).toBe('absent')
  })

  it('un changement de nom vide l’anneau', () => {
    pushIface(frame(1000))
    pushIface(frame(1100, { iface_name: 'e-srv' }))
    expect(live.iface.rx.length).toBe(1)
    expect(live.iface.name).toBe('e-srv')
  })

  it('un saut d’horloge arrière > 1 s vide l’anneau', () => {
    pushIface(frame(5000))
    pushIface(frame(3000))
    expect(live.iface.rx.length).toBe(1)
  })

  it('un échantillon en retard ≤ 1 s est ignoré', () => {
    pushIface(frame(5000))
    pushIface(frame(4500))
    expect(live.iface.rx.length).toBe(1)
  })

  it('clearLive ne vide pas l’interface', () => {
    pushIface(frame(1000))
    clearLive()
    expect(live.iface.rx.length).toBe(1)
  })

  it('plafonne à 600 points', () => {
    for (let i = 0; i < 601; i++) pushIface(frame(1000 + i * 100))
    expect(live.iface.rx.length).toBe(600)
  })

  it('pousse un frame idle sans running ni phase', () => {
    pushIface(frame(1000))
    expect(live.iface.rx.length).toBe(1)
  })
})

describe('sse', () => {
  it('pousse l’interface avant la porte measured', () => {
    const cands = ['src/lib/sse.ts', 'web/frontend/src/lib/sse.ts']
    let s = ''
    for (const p of cands) {
      try { s = readFileSync(p, 'utf8'); break } catch {}
    }
    expect(s).toContain('pushIface(data)')
    expect(s.indexOf('pushIface(data)')).toBeLessThan(s.indexOf('const measured'))
  })
})
```

- [ ] **Step 2: Run the tests to verify they fail**

Run from `web/frontend`: `npm test -- src/lib/live.test.ts`

Expected: FAIL — `pushIface` is not exported, `live.iface` is undefined.

- [ ] **Step 3: Add the `LiveFrame` fields**

In `web/frontend/src/lib/types.ts`, add after `bulk_goodput_mbps`:

```ts
  // Interface observée (additif, absent quand CGO_IFACE est vide).
  // iface_valid=false invalide tous les champs numériques.
  iface_name?: string
  iface_source?: 'twin' | 'prod' | 'unknown'
  iface_rx_mbps?: number
  iface_tx_mbps?: number
  iface_rx_drops?: number
  iface_tx_drops?: number
  iface_window_ms?: number
  iface_sample_ts?: number
  iface_valid?: boolean
  iface_gap?: string
```

- [ ] **Step 4: Add the ring and `pushIface`**

In `web/frontend/src/lib/live.ts`, add the type import at the top (the module
currently has no imports; `types.ts` does not import `live.ts`, so there is no
cycle):

```ts
import type { LiveFrame } from './types'
```

Add the type near `Ring`:

```ts
export type IfacePt = [number, number | null]
```

Add `iface` inside the `live` object, after `drops` and before `max`:

```ts
  // anneau interface séparé — 60 s à 10 Hz. clearLive ne le touche pas :
  // l'observation continue survive aux transitions de campagne.
  iface: {
    rx: [] as IfacePt[],
    tx: [] as IfacePt[],
    name: '',
    source: '',
    valid: false,
    gap: '',
    windowMs: 0,
    rxDrops: 0,
    txDrops: 0,
    ts: 0,
    seq: 0,
    max: 600,
  },
```

Add the push function after `pushSparse` (do NOT reuse `push` — it coerces null to 0):

```ts
export function pushIface(f: Partial<LiveFrame>) {
  const ts = f.iface_sample_ts
  if (!ts) return
  const r = live.iface
  if ((f.iface_name ?? '') !== r.name) {
    r.rx.length = 0
    r.tx.length = 0
    r.ts = 0
    r.name = f.iface_name ?? ''
  }
  if (r.ts) {
    if (ts === r.ts) return
    if (ts < r.ts) {
      if (r.ts - ts <= 1000) return
      // saut d'horloge arrière > 1 s : la série repart, sans pont
      r.rx.length = 0
      r.tx.length = 0
    }
  }
  const valid = !!f.iface_valid
  const mbps = (v?: number) => (valid && v != null && Number.isFinite(v) ? v : null)
  r.rx.push([ts, mbps(f.iface_rx_mbps)])
  r.tx.push([ts, mbps(f.iface_tx_mbps)])
  if (r.rx.length > r.max) {
    r.rx.shift()
    r.tx.shift()
  }
  r.ts = ts
  r.valid = valid
  r.gap = f.iface_gap ?? ''
  r.windowMs = f.iface_window_ms ?? 0
  r.source = f.iface_source ?? r.source
  if (valid) {
    if (typeof f.iface_rx_drops === 'number') r.rxDrops = f.iface_rx_drops
    if (typeof f.iface_tx_drops === 'number') r.txDrops = f.iface_tx_drops
  }
  r.seq++
}
```

Add a comment to `clearLive` explaining the exclusion, and do not clear `live.iface` there.

- [ ] **Step 5: Push every interface frame**

In `web/frontend/src/lib/sse.ts`, import `pushIface`:

```ts
import { pushFrame, pushIface, live as liveRing } from './live'
```

After `const data = JSON.parse(ev.data) as Partial<LiveFrame>`, add:

```ts
      // l'interface se mesure hors campagne : pousser AVANT la porte "measured"
      if (data.ts) pushIface(data)
```

- [ ] **Step 6: Run the tests to verify they pass**

Run from `web/frontend`: `npm test -- src/lib/live.test.ts`

Expected: PASS (9 tests).

- [ ] **Step 7: Commit**

```bash
git add web/frontend/src/lib/types.ts web/frontend/src/lib/live.ts web/frontend/src/lib/sse.ts web/frontend/src/lib/live.test.ts
git commit -m "feat(ui): iface ring and SSE push"
```

---

### Task 7: Raw interface chart and relabel

**Files:**
- Modify: `web/frontend/src/lib/chartGrammar.ts`
- Modify: `web/frontend/src/lib/chartGrammar.test.ts`
- Modify: `web/frontend/src/views/LiveView.tsx`
- Modify: `web/frontend/src/views/LiveView.test.ts`

**Interfaces:**
- Consumes: `live.iface`, `IfacePt`, `pushIface` from Task 6; `baseOption`, `chargeMarkArea`, `CRAFT`.
- Produces:
  - `export function rawLineSeries(name: string, data: [number, number | null][], color: string)`
  - a `Débit interface` card with `chart-iface`
  - the existing app card relabeled `Débit utile applicatif`

- [ ] **Step 1: Write the failing tests**

In `web/frontend/src/lib/chartGrammar.test.ts`, import `rawLineSeries` and add:

```ts
  it('rawLineSeries ne lisse pas, n’échantillonne pas et ne comble pas les trous', () => {
    const s: any = rawLineSeries('RX', [[0, 1], [1, null]], '#5ad3e3')
    expect(s.smooth).toBe(false)
    expect(s.connectNulls).toBe(false)
    expect(s.sampling).toBeUndefined()
  })
```

In `web/frontend/src/views/LiveView.test.ts`, add:

```ts
  it('expose le débit interface brut et distingue les deux plans', () => {
    const s = readLive()
    expect(s).toContain('Débit interface')
    expect(s).toContain('chart-iface')
    expect(s).toContain('rawLineSeries')
    expect(s).toContain('Débit utile applicatif')
  })
```

- [ ] **Step 2: Run the tests to verify they fail**

Run from `web/frontend`: `npm test -- src/lib/chartGrammar.test.ts src/views/LiveView.test.ts`

Expected: FAIL — `rawLineSeries` is not exported; the new strings are absent.

- [ ] **Step 3: Add `rawLineSeries`**

In `web/frontend/src/lib/chartGrammar.ts`, after `lineSeries`:

```ts
// rawLineSeries — série brute : aucun lissage, aucun LTTB, les trous coupent.
// Le débit d'interface est une mesure échantillonnée ; une courbe lissée
// inventerait des points entre deux fenêtres réelles.
export function rawLineSeries(name: string, data: [number, number | null][], color: string) {
  return {
    name,
    type: 'line' as const,
    showSymbol: false,
    smooth: false,
    connectNulls: false,
    lineStyle: { width: 2, cap: 'round' as const, join: 'round' as const, color },
    emphasis: { focus: 'series' as const, lineStyle: { width: 3 } },
    blur: { lineStyle: { opacity: 0.2 } },
    data,
  }
}
```

- [ ] **Step 4: Wire the chart in `LiveView.tsx`**

Add the import:

```ts
import { baseOption, lineSeries, barSeries, chargeMarkArea, rawLineSeries, CRAFT } from '../lib/chartGrammar'
```

Add the chart handle after the goodput one:

```ts
  const iface = useChart('Interface', 'Mbit/s')
```

Widen the shared `setData` signature so a raw series is accepted (the body
already reads `.name` and casts `data`, so nothing else changes):

```ts
  const setData = (series: { name?: string; data?: unknown[] }[], extra?: Record<string, unknown>) => {
```

Add the refs next to `lastRef`:

```ts
  const ifaceSeqRef = useRef(-1)
  const ifaceLastRef = useRef(0)
```

Add the interface data to the rAF loop, before the existing `live.seq` early
return, so an idle observation still paints:

```ts
  useRafLoop((ts) => {
    // interface — porte 10 Hz propre, série brute, jamais passé par LTTB
    if (live.iface.seq !== ifaceSeqRef.current && ts - ifaceLastRef.current >= 100) {
      const charging = live.phase === 'charge'
      const cs = charging ? live.phaseSince['charge'] ?? live.iface.rx[0]?.[0] ?? Date.now() - 1000 : 0
      const ce = charging ? live.iface.rx.at(-1)?.[0] ?? Date.now() : 0
      const ma = chargeMarkArea(cs, ce, charging)
      iface.setData([
        { ...rawLineSeries('RX', live.iface.rx, CRAFT.live), markArea: ma } as any,
        { ...rawLineSeries('TX', live.iface.tx, CRAFT.ok) } as any,
      ])
      ifaceSeqRef.current = live.iface.seq
      ifaceLastRef.current = ts
    }
    // dirty-check — les anneaux n'ont pas bougé : rien à re-render, frame gratuite
    if (live.seq === seqRef.current) return
```

Extend `rafActive`:

```ts
  const rafActive = live.rtt95.length > 0 || live.small.length > 0 || live.goodput.length > 0 || lockedRingRef.current.length > 0 || live.iface.rx.length > 0
```

Add the header values after `goodputEmpty`:

```ts
  const ifaceRx = live.iface.rx.at(-1)?.[1] ?? null
  const ifaceTx = live.iface.tx.at(-1)?.[1] ?? null
  const ifaceFmt = (v: number | null) => (v == null ? '—' : v.toFixed(1))
  const ifaceSourceLabel = live.iface.source === 'twin' ? 'twin'
    : live.iface.source === 'prod' ? 'production' : 'source non déclarée'
  const ifaceDrops = live.iface.valid
    ? `pertes IF RX ${live.iface.rxDrops} · TX ${live.iface.txDrops}`
    : 'pertes IF RX — · TX —'
  // raisons connues en français ; une raison inconnue s'affiche telle quelle
  const GAP_FR: Record<string, string> = {
    warmup: 'démarrage', absent: 'interface absente', reset: 'compteur réinitialisé',
    slow: 'fenêtre trop longue', read: 'lecture impossible', clock: 'horloge',
  }
  const ifaceGap = live.iface.gap ? (GAP_FR[live.iface.gap] ?? live.iface.gap) : ''
  const ifaceRight = `${live.iface.name || '—'} · RX ${ifaceFmt(ifaceRx)} · TX ${ifaceFmt(ifaceTx)} Mbit/s · fenêtre ${live.iface.windowMs > 0 ? live.iface.windowMs.toFixed(0) : '—'} ms · ${ifaceDrops} · ${ifaceSourceLabel}${ifaceGap ? ' · ' + ifaceGap : ''}`
  const ifaceEmpty = live.iface.rx.length === 0
  const ifaceHint = !live.iface.name
    ? 'interface non observée — déclarer CGO_IFACE côté serveur'
    : live.iface.gap === 'absent' ? 'interface absente' : 'en attente d\'échantillons'
  const ifaceSub = !live.iface.name ? 'compteurs noyau'
    : live.iface.windowMs > 0 && live.iface.windowMs <= 150 ? 'compteurs noyau · échantillon 10 Hz · fenêtre 60 s'
    : live.iface.windowMs > 0 ? `compteurs noyau · échantillon ${live.iface.windowMs.toFixed(0)} ms`
    : 'compteurs noyau'
```

Insert the card between the live-empty block and `<LiveGuide />`:

```tsx
      {!liveSnap && !showLiveSrc && <div className="card live-empty"><EmptyState kind="idle" hint={IDLE_HINT} /></div>}
      {showLiveSrc && (
        <Card
          head="Débit interface"
          sub={ifaceSub}
          className="live-hero-card"
          right={<span className="mono" style={{ fontSize: 'var(--fs-micro)', color: 'var(--text-faint)' }}>{ifaceRight}</span>}
        >
          <div className="chart-flex">
            <ChartSurface title="Interface" unit="Mbit/s" domId="chart-iface" height="100%" empty={ifaceEmpty} hint={ifaceHint} onReady={iface.onReady} />
          </div>
        </Card>
      )}
      {/* guide de lecture (U6a) — que disent ces chiffres, repliable, au-dessus du bento */}
```

Relabel the existing app chart:

- `const goodput = useChart('Bulk goodput', 'Mbit/s')` becomes `useChart('Débit utile applicatif', 'Mbit/s')`.
- the card `head="Bulk goodput"` becomes `head="Débit utile applicatif"`.
- the `ChartSurface title="Bulk goodput"` becomes `title="Débit utile applicatif"`.
- keep the card `sub="mesuré au récepteur (compteur noyau)"` and the metric card label `Débit utile`.

Add the chart to the instrumentation seam:

```ts
  if (typeof window !== 'undefined') (window as any).__CGO_CHARTS = { rtt: rtt.chart.current, small: small.chart.current, goodput: goodput.chart.current, iface: iface.chart.current }
```

- [ ] **Step 5: Run the tests to verify they pass**

Run from `web/frontend`:

```bash
npm test -- src/lib/chartGrammar.test.ts src/views/LiveView.test.ts
npm run typecheck
```

Expected: PASS, and `tsc --noEmit` exits 0 with no error.

- [ ] **Step 6: Commit**

```bash
git add web/frontend/src/lib/chartGrammar.ts web/frontend/src/lib/chartGrammar.test.ts web/frontend/src/views/LiveView.tsx web/frontend/src/views/LiveView.test.ts
git commit -m "feat(ui): raw interface throughput chart"
```

---

### Task 8: Windows gates

**Files:** none (fix only what fails).

- [ ] **Step 1: Run the Go gates**

```bash
go test ./...
go vet ./...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...
```

Expected: all exit 0. `go vet` needs `web/frontend/dist` present; it is in the checkout. If a gate is not runnable here, record "not runnable here" in the task report; do not claim it passed.

- [ ] **Step 2: Run the frontend gates**

```bash
cd web/frontend
npm test
npm run build
```

Expected: all tests pass, `tsc --noEmit` clean, Vite build clean, bundle check clean.

- [ ] **Step 3: Run the archive audit only if `book/` is present**

```bash
test -d book && go run ./cmd/cgo verify || echo "book/ absent — audit non applicable"
```

Expected: the guard output or the explicit not-applicable line. Never restate the frozen counts from memory; cite `docs/VERIFY.md`.

- [ ] **Step 4: Commit only if a gate forced a fix**

If a gate required a code fix, run `git status --short`, add those files by
explicit path (never `git add -A`), and commit with a message naming the gate
that caught it. If every gate passed unchanged, skip this step.

---

### Task 9: Twin live acceptance

**Files:** none (measurement only). Do not edit code during acceptance; a failure stops the task and is reported.

**Preconditions:** Task 8 green; CGO-DEV reachable; no other namespace/qdisc experiment running on the same host; `docs/VERIFY.md` exists (slice 9 Phase A).

- [ ] **Step 1: Cross-build and install on CGO-DEV**

From the checkout:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/cgo-linux ./cmd/cgo
scp -P 2222 -i ~/.ssh/nemo_lab /tmp/cgo-linux lab@100.72.251.32:/home/lab/cgo-current/cgo-linux
```

Expected: transfer completes. Do not copy the old `/home/lab/cgo` tree or build on the VM (its NAT link is very slow).

- [ ] **Step 2: Bring up the twin and serve from inside `ns_client`**

On CGO-DEV:

```bash
cd /home/lab/cgo-current
bash kit/run-dualns.sh down
nohup bash kit/run-dualns.sh serve --bin /home/lab/cgo-current/cgo-linux >/tmp/cgo-serve.log 2>&1 &
sleep 5
sudo ip netns exec ns_client curl -s http://10.99.0.2:9090/api/health
```

Expected: bench up, `testbedsrv` listening, `/api/health` returns `{"ok":true,...}`.

- [ ] **Step 3: Capture 32 s of SSE during a 25 s local upload**

```bash
sudo ip netns exec ns_client cat /proc/net/dev | grep 'c-cli'
sudo ip netns exec ns_edge_router tc -s qdisc show dev e-srv | grep -o 'Sent [0-9]* bytes'

sudo ip netns exec ns_client sh -c 'timeout 32 curl -sN http://10.99.0.2:9090/api/stream > /tmp/iface-sse.txt' &
sleep 2
sudo ip netns exec ns_client sh -c 'dd if=/dev/zero bs=1M count=60 2>/dev/null | nc 10.30.0.2 5201'
wait

sudo ip netns exec ns_client cat /proc/net/dev | grep 'c-cli'
sudo ip netns exec ns_edge_router tc -s qdisc show dev e-srv | grep -o 'Sent [0-9]* bytes'
```

Expected: the upload runs about 25 s at the 20 Mbit/s CAKE ceiling; the capture holds at least 300 frames.

- [ ] **Step 4: Assert the cadence, window, and rate**

```bash
grep -o '"iface_sample_ts":[0-9]*' /tmp/iface-sse.txt | cut -d: -f2 | sort -u | wc -l
grep -o '"iface_window_ms":[0-9.]*' /tmp/iface-sse.txt | cut -d: -f2 | sort -n | awk '{a[NR]=$1} END{printf "n=%d p95=%.1f max=%.1f\n", NR, a[int(NR*0.95)], a[NR]}'
grep -o '"iface_tx_mbps":[0-9.]*' /tmp/iface-sse.txt | cut -d: -f2 | sort -n | awk '{a[NR]=$1} END{printf "p50=%.1f p95=%.1f max=%.1f\n", a[int(NR*0.5)], a[int(NR*0.95)], a[NR]}'
grep -c '"iface_valid":true' /tmp/iface-sse.txt
```

Accept when:

- unique timestamps ≥ 8 per second of the run (about 230 for 32 s, replay duplicates excluded by `sort -u`);
- active-window p95 `iface_window_ms` ≤ 150 ms;
- every frame carries `iface_sample_ts`; no fabricated or repeated-as-new sample;
- active `iface_tx_mbps` p95 is in the order of the 20 Mbit/s shaped stream.

Copy the command output verbatim into the task report. Do not summarize it into a nicer number.

- [ ] **Step 5: Assert byte accounting**

Compute `(tx_after - tx_before)` for `c-cli` and `(sent_after - sent_before)` for `e-srv`. Accept when the difference is within 1 % (the pre-slice probe measured 0.2 %). They are different layers: report the delta, never call them equal.

- [ ] **Step 6: Gap test**

```bash
sudo ip netns exec ns_client sh -c 'timeout 8 curl -sN http://10.99.0.2:9090/api/stream > /tmp/iface-gap.txt' &
sleep 1
sudo ip netns exec ns_client ip link set c-cli down
sudo ip netns exec ns_client ip link set c-cli name c-cli-gap
sleep 2
sudo ip netns exec ns_client ip link set c-cli-gap name c-cli
sudo ip netns exec ns_client ip link set c-cli up
wait
grep -o '"iface_gap":"[a-z]*"' /tmp/iface-gap.txt | sort | uniq -c
sudo ip netns exec ns_client ping -c1 -W2 10.30.0.2
```

Accept when the capture contains `absent` during the rename and `warmup` after the return, the next window is valid, and the bench still reaches the sink. If the rename breaks the bench, restore it with `bash kit/run-dualns.sh down` then `up`; record the failure, do not hand-wave it.

- [ ] **Step 7: Browser trace (host with access to `http://10.99.0.2:9090`)**

Open the live wall, start the same upload, and record a performance trace or the chart's paint count for 25 s. Accept when the interface chart advances at least 8 times per second and no long task exceeds 50 ms. If no reachable browser is available, write "not runnable here" and do not claim the paint gate; the SSE gate above still stands.

- [ ] **Step 8: Race gate**

Run from the checkout: `go test -race ./pkg/iface`. Expected: PASS.

If the Windows toolchain has no race support, copy the three package files to
CGO-DEV and run them as a standalone stdlib module:

```bash
scp -P 2222 -i ~/.ssh/nemo_lab pkg/iface/iface.go pkg/iface/iface_test.go pkg/iface/sampler_test.go lab@100.72.251.32:/tmp/
ssh -p 2222 -i ~/.ssh/nemo_lab lab@100.72.251.32 \
  'mkdir -p /tmp/ifacerace && mv /tmp/iface.go /tmp/iface_test.go /tmp/sampler_test.go /tmp/ifacerace/ && cd /tmp/ifacerace && go mod init ifacerace >/dev/null && go test -race ./...'
```

Expected: PASS. If neither path works, record "not runnable here"; do not claim it.

- [ ] **Step 9: Teardown**

```bash
bash kit/run-dualns.sh down
ip netns list
```

Expected: no `ns_client`, `ns_wan_emulator`, `ns_edge_router`, `ns_server` left. Leave the VM running.

- [ ] **Step 10: Record the acceptance**

Append to `docs/VERIFY.md` in the live-acceptance section created by slice 9 Phase A, using its table columns. Include: date, host (CGO-DEV, kernel), the exact commands, the cadence count, window p95/max, TX p95/max, byte delta, gap reasons observed, the browser result (or "not runnable here"), and `PASS`/`FAIL`. Do not restate the frozen corpus counts; cite the existing guard row.

- [ ] **Step 11: Commit the record**

```bash
git add docs/VERIFY.md
git commit -m "docs: record interface observation twin acceptance"
```

---

### Task 10: Close the slice

**Files:**
- Modify: `docs/superpowers/plans/README.md`

**Interfaces:** none.

- [ ] **Step 1: Update the slice map and section**

In `docs/superpowers/plans/README.md`:

- Slice 10 status becomes **Done** only if Task 8 passed, the SSE gate in Task 9
  passed, and the browser paint gate passed. If the browser gate is "not
  runnable here", keep **In progress** and name that gate as the blocker; do
  not claim 10 Hz rendering. Also keep **In progress** if any other gate failed.
- Next action: `—` when Done; otherwise the exact blocker.
- Section `### Slice 10 — live interface observation`: add the measured acceptance result in one or two lines, citing `docs/VERIFY.md`.
- Leave slices 5 and 7 **Deferred**.

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/plans/README.md
git commit -m "docs: close live interface observation slice"
```

---

## Acceptance summary

- Unit: parser, rate math, every gap reason, `Latest`, race, the hub seam, env resolution, ring dedupe, raw series.
- Windows: `go test ./...`, `go vet ./...`, Linux cross-build, frontend tests and build.
- Twin: 10 Hz unique timestamps, measured window p95 ≤ 150 ms, TX in the shaped range, byte delta ≤ 1 %, `absent`/`warmup` gap observed, teardown clean.
- Browser: ≥ 8 paints/s and no long task > 50 ms, or "not runnable here".
- No production claim. Production SNMP/RouterOS is Phase 2 behind the same `Reader` seam.
