# Trace-Driven Netem `.dist` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn baseline audit RTTs into a sealed native netem `.dist` and let the `dualns` campaign replay it by profile, failing closed when the artifact is missing, stale, or tampered.

**Architecture:** A strict CSV reader feeds a pure quantile-table builder in `pkg/qdisc`; the CLI writes `<profile>.dist` plus a SHA-256 sidecar; the campaign loads the sidecar by profile before any qdisc mutation and applies the table to the measured hop only, with the profile delay split across both hops.

**Tech Stack:** Go 1.25, standard library only, `tc`/netem on Linux 6.8 (engine tests run on Windows).

## Global Constraints

- Spec: `docs/superpowers/specs/2026-10-27-cgo-trace-driven-dist-design.md`.
- Branch: `rebuild/plan1-economics-audit`; do not merge to `main` in this slice.
- No new Go dependency. Standard library only.
- `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv` headers and bytes are frozen; the frozen corpus guard is **159 runs / 208 valid lines** (never “208 runs”).
- Do not modify `epbfBBR.md`, `data/runs/`, or `data/audits/`; leave them untracked.
- Calibration artifacts live in `data/calibrations/`; never `data/dist/` (`dist/` is gitignored as a directory).
- Do not commit a `.dist` built from the two `SMOKE-TEST` traces. Tests use `t.TempDir()`.
- French comments and user-facing strings; English identifiers.
- Commit per task with explicit paths. Never `git add -A`.
- Each task runs its package tests; Task 7 runs the full Windows gates:
  `go test ./...`, `go vet ./...`, `GOOS=linux GOARCH=amd64 go build ./...`.
- Do not claim live netem acceptance until Task 8 runs on Linux 6.8.

## File structure

| File | Responsibility |
|---|---|
| `pkg/audit/telemetry.go` | Add strict `ReadTelemetry` + row parser. |
| `pkg/audit/telemetry_test.go` | Reader round-trip and rejection tests. |
| `pkg/qdisc/dist.go` | Calibration struct, quantile builder, `.dist` renderer, sealed write/load. |
| `pkg/qdisc/dist_test.go` | Table math, format, seal, fail-closed tests. |
| `pkg/qdisc/ge.go` | Shared netem argument builder + `ApplyNetemDistribution`. |
| `pkg/qdisc/qdisc.go` | `NsRunner.Env` + `nsTCArgs`. |
| `pkg/qdisc/dist_netem_test.go` | tc args and env-prefix tests. |
| `cmd/cgo/calibrate.go` | `cgo calibrate` command. |
| `cmd/cgo/calibrate_test.go` | End-to-end calibrate test in temp dirs. |
| `cmd/cgo/main.go` | Dispatch + usage line. |
| `pkg/campagne/campagne.go` | `Deps.DelayModel` / `Deps.CalibrationDir` + defaults. |
| `pkg/campagne/dualns.go` | `ProdDepsDualNS` env wiring + `applyPhysicsDist`. |
| `pkg/campagne/prod.go` | Legacy deps carry the env so dist+legacy fails closed. |
| `pkg/campagne/dualns_test.go` | Measured-hop, down, and fail-closed tests. |
| `docs/methodology.fr.md` | Document the calibration and the RTT convention. |

---

### Task 1: Strict telemetry reader

**Files:**
- Modify: `pkg/audit/telemetry.go` (append after `WriteTelemetry`)
- Test: `pkg/audit/telemetry_test.go` (append)

**Interfaces:**
- Consumes: `telemetryHeader`, `TelemetryRow`, `ArtifactID` from `pkg/audit`.
- Produces: `func ReadTelemetry(baseDir, id string) ([]TelemetryRow, error)`.

- [ ] **Step 1: Write the failing tests**

Append to `pkg/audit/telemetry_test.go`:

```go
func TestReadTelemetryStrict(t *testing.T) {
	dir := t.TempDir()
	res := &Result{AuditID: "audit-read", Telemetry: []TelemetryRow{
		{Phase: "baseline", OffsetMs: 0, RTTms: 20.5},
		{Phase: "baseline", OffsetMs: 400, RTTms: 21.25, SmallMs: 25},
		{Phase: "load", OffsetMs: 800, RTTms: 30, BulkBytes: 1234},
	}}
	if _, err := WriteTelemetry(dir, res); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadTelemetry(dir, "audit-read")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[1].RTTms != 21.25 || rows[1].SmallMs != 25 || rows[2].BulkBytes != 1234 {
		t.Fatalf("lignes relues : %+v", rows)
	}
}

func TestReadTelemetryRejectsUnknownHeader(t *testing.T) {
	dir := t.TempDir()
	auditDir := filepath.Join(dir, "audits", "audit-bad")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "phase,offset,rtt_ms,small_ms,bulk_bytes\nbaseline,0,20,0,\n"
	if err := os.WriteFile(filepath.Join(auditDir, "telemetry.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTelemetry(dir, "audit-bad"); err == nil {
		t.Fatal("en-tête inconnu accepté — les colonnes glisseraient en silence")
	}
}

func TestReadTelemetryRejectsEmptyProbe(t *testing.T) {
	dir := t.TempDir()
	auditDir := filepath.Join(dir, "audits", "audit-empty")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "phase,offset_ms,rtt_ms,small_ms,bulk_bytes\nbaseline,0,0,0,\n"
	if err := os.WriteFile(filepath.Join(auditDir, "telemetry.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTelemetry(dir, "audit-empty"); err == nil {
		t.Fatal("ligne sans mesure acceptée — un zéro deviendrait une mesure")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/audit -run TestReadTelemetry -v`

Expected: FAIL with `undefined: ReadTelemetry`.

- [ ] **Step 3: Implement the reader**

In `pkg/audit/telemetry.go`, extend the import block:

```go
import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)
```

Append after `WriteTelemetry`:

```go
// ReadTelemetry — lecture stricte de data/audits/<id>/telemetry.csv.
// L'en-tête doit être exactement celui du writer : une colonne déplacée
// changerait le sens des nombres. Une sonde absente reste absente (0) ;
// elle n'est jamais réinventée en mesure.
func ReadTelemetry(baseDir, id string) ([]TelemetryRow, error) {
	safe, err := ArtifactID(id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(baseDir, "audits", safe, "telemetry.csv")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("audit: lecture télémétrie %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = len(telemetryHeader)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("audit: en-tête télémétrie %s: %w", path, err)
	}
	if !slices.Equal(header, telemetryHeader) {
		return nil, fmt.Errorf("audit: en-tête télémétrie %s = %v, attendu %v", path, header, telemetryHeader)
	}

	var rows []TelemetryRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("audit: ligne télémétrie %s: %w", path, err)
		}
		row, err := parseTelemetryRow(rec)
		if err != nil {
			return nil, fmt.Errorf("audit: ligne télémétrie %s: %w", path, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseTelemetryRow(rec []string) (TelemetryRow, error) {
	var row TelemetryRow
	row.Phase = strings.TrimSpace(rec[0])
	if row.Phase == "" {
		return row, fmt.Errorf("phase vide")
	}
	offset, err := strconv.ParseInt(strings.TrimSpace(rec[1]), 10, 64)
	if err != nil || offset < 0 {
		return row, fmt.Errorf("offset_ms invalide %q", rec[1])
	}
	row.OffsetMs = offset

	for _, f := range []struct {
		name string
		dst  *float64
		raw  string
	}{
		{"rtt_ms", &row.RTTms, rec[2]},
		{"small_ms", &row.SmallMs, rec[3]},
	} {
		s := strings.TrimSpace(f.raw)
		if s == "" {
			continue
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return row, fmt.Errorf("%s invalide %q", f.name, f.raw)
		}
		*f.dst = v
	}

	if s := strings.TrimSpace(rec[4]); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return row, fmt.Errorf("bulk_bytes invalide %q", rec[4])
		}
		row.BulkBytes = v
	}
	if row.RTTms == 0 && row.SmallMs == 0 && row.BulkBytes == 0 {
		return row, fmt.Errorf("ligne sans mesure (rtt, small et bulk vides)")
	}
	return row, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/audit -run TestReadTelemetry -v`

Expected: PASS (3 tests).

- [ ] **Step 5: Run the package suite**

Run: `go test ./pkg/audit`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/audit/telemetry.go pkg/audit/telemetry_test.go
git commit -m "audit: strict telemetry reader"
```

---

### Task 2: Calibration math and `.dist` renderer

**Files:**
- Create: `pkg/qdisc/dist.go`
- Test: `pkg/qdisc/dist_test.go`

**Interfaces:**
- Consumes: `model.Profile` (`ID`, `DelayMs`).
- Produces:
  - `type Calibration struct { … Table []int16 }`
  - `func BuildCalibration(prof model.Profile, auditID string, rtts []float64) (Calibration, error)`
  - `func (c Calibration) Render() []byte`
  - constants `distEntries = 1024`, `distScale = 8192.0`, `distEntryMax = 32767`, `distMinSamples = 30`.

- [ ] **Step 1: Write the failing tests**

Create `pkg/qdisc/dist_test.go`:

```go
package qdisc

import (
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

func TestBuildCalibrationRTTTable(t *testing.T) {
	rtts := make([]float64, 200)
	for i := range rtts {
		rtts[i] = 100 + float64(i%10)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	if cal.FloorMs != 100 || cal.HopDelayMs != 50 {
		t.Fatalf("floor/hop = %v/%v", cal.FloorMs, cal.HopDelayMs)
	}
	if cal.TableSize != distEntries || len(cal.Table) != distEntries {
		t.Fatalf("table = %d entrées, sidecar %d", len(cal.Table), cal.TableSize)
	}
	if cal.Table[len(cal.Table)-1] != distEntryMax {
		t.Fatalf("dernière entrée = %d, want %d (p99.9 mappé au max)", cal.Table[len(cal.Table)-1], distEntryMax)
	}
	if cal.SourceP50RTTms < 104 || cal.SourceP50RTTms > 106 {
		t.Fatalf("p50 source = %v", cal.SourceP50RTTms)
	}
	if cal.RoundTripP50RTTms < 103 || cal.RoundTripP50RTTms > 107 {
		t.Fatalf("p50 reconstruit = %v", cal.RoundTripP50RTTms)
	}
	for i := 1; i < len(cal.Table); i++ {
		if cal.Table[i] < cal.Table[i-1] {
			t.Fatalf("table non monotone à %d : %d < %d", i, cal.Table[i], cal.Table[i-1])
		}
	}
	for _, v := range cal.Table {
		if v < 0 || v > distEntryMax {
			t.Fatalf("entrée hors bornes : %d", v)
		}
	}
	lines := strings.Split(strings.TrimSpace(string(cal.Render())), "\n")
	if len(lines) != distEntries {
		t.Fatalf("lignes .dist = %d", len(lines))
	}
	for _, ln := range lines {
		if len(ln) > 1 && ln[0] == '0' {
			t.Fatalf("zéro-padding interdit (strtol base 0) : %q", ln)
		}
	}
}

func TestBuildCalibrationRejectsBadTrace(t *testing.T) {
	short := make([]float64, distMinSamples-1)
	for i := range short {
		short[i] = 101
	}
	if _, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "a", short); err == nil {
		t.Fatal("trace courte acceptée")
	}

	below := make([]float64, distMinSamples)
	for i := range below {
		below[i] = 20
	}
	if _, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "a", below); err == nil {
		t.Fatal("trace sous le plancher acceptée")
	}
}

func TestQuantileNearestRank(t *testing.T) {
	s := []float64{1, 2, 3, 4}
	if got := quantile(s, 0.5); got != 3 {
		t.Fatalf("quantile p50 = %v, want 3", got)
	}
	if got := quantile(s, 0); got != 1 {
		t.Fatalf("quantile p0 = %v", got)
	}
	if got := quantile(s, 1); got != 4 {
		t.Fatalf("quantile p1 = %v", got)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/qdisc -run 'TestBuildCalibration|TestQuantile' -v`

Expected: FAIL with `undefined: BuildCalibration`.

- [ ] **Step 3: Implement the builder**

Create `pkg/qdisc/dist.go`:

```go
package qdisc

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/pkg/model"
)

// Constantes du format netem (tc/q_netem.c) : la table est lue comme des
// entiers base 0, et le noyau applique mu + (sigma/8192)*t. NETEM_DIST_MAX
// borne une entrée, pas la taille de la table (16384).
const (
	distEntries    = 1024
	distScale      = 8192.0
	distEntryMax   = 32767
	distMinSamples = 30
)

// Calibration — distribution empirique d'un profil et sa provenance. La
// table n'est pas dupliquée dans le JSON : le .dist la porte, le sidecar
// la scelle par DistSHA256.
type Calibration struct {
	Schema                int     `json:"schema"`
	Profile               string  `json:"profile"`
	Basis                 string  `json:"basis"`
	HopModel              string  `json:"hop_model"`
	SourceAudit           string  `json:"source_audit"`
	SourceReportSHA256    string  `json:"source_report_sha256,omitempty"`
	SourceTelemetrySHA256 string  `json:"source_telemetry_sha256"`
	SourceSite            string  `json:"source_site,omitempty"`
	SourceLinkType        string  `json:"source_link_type,omitempty"`
	Created               string  `json:"created"`

	ProfileDelayMs float64 `json:"profile_delay_ms"`
	FloorMs        float64 `json:"floor_ms"`
	HopDelayMs     float64 `json:"hop_delay_ms"`
	ScaleMs        float64 `json:"scale_ms"`
	TableSize      int     `json:"table_size"`
	DistFile       string  `json:"dist_file"`
	DistSHA256     string  `json:"dist_sha256"`

	Samples       int     `json:"samples"`
	BelowFloorPct float64 `json:"below_floor_pct"`
	ClippedPct    float64 `json:"clipped_pct"`

	SourceP50RTTms  float64 `json:"source_p50_rtt_ms"`
	SourceP95RTTms  float64 `json:"source_p95_rtt_ms"`
	SourceP999RTTms float64 `json:"source_p999_rtt_ms"`
	SourceMaxRTTms  float64 `json:"source_max_rtt_ms"`

	RoundTripP50RTTms  float64 `json:"roundtrip_p50_rtt_ms"`
	RoundTripP95RTTms  float64 `json:"roundtrip_p95_rtt_ms"`
	RoundTripP999RTTms float64 `json:"roundtrip_p999_rtt_ms"`

	Table []int16 `json:"-"`
}

// BuildCalibration — RTT baseline → table netem. Le délai du profil est un
// RTT (convention du corpus) : le résidu est RTT - plancher, et la table ne
// contient que ce résidu. Une trace dont la médiane ne dépasse pas le
// plancher ne décrit pas ce profil : refus, jamais de table dégénérée.
func BuildCalibration(prof model.Profile, auditID string, rtts []float64) (Calibration, error) {
	if strings.TrimSpace(prof.ID) == "" {
		return Calibration{}, errors.New("calibration: profil sans identifiant")
	}
	if prof.DelayMs <= 0 {
		return Calibration{}, fmt.Errorf("calibration %s: delay_ms %.2f ≤ 0", prof.ID, prof.DelayMs)
	}
	if len(rtts) < distMinSamples {
		return Calibration{}, fmt.Errorf("calibration %s: %d échantillons baseline < %d — trace trop courte", prof.ID, len(rtts), distMinSamples)
	}
	sorted := append([]float64(nil), rtts...)
	sort.Float64s(sorted)
	for _, v := range sorted {
		if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return Calibration{}, fmt.Errorf("calibration %s: RTT invalide %v", prof.ID, v)
		}
	}
	p50 := quantile(sorted, 0.50)
	if p50 <= prof.DelayMs {
		return Calibration{}, fmt.Errorf("calibration %s: p50 RTT %.2f ms ≤ plancher %.2f ms — la trace ne correspond pas au profil", prof.ID, p50, prof.DelayMs)
	}

	residuals := make([]float64, len(sorted))
	below := 0
	for i, v := range sorted {
		r := v - prof.DelayMs
		if r < 0 {
			below++
			r = 0
		}
		residuals[i] = r
	}
	p999 := quantile(residuals, 0.999)
	if p999 <= 0 {
		return Calibration{}, fmt.Errorf("calibration %s: queue résiduelle nulle", prof.ID)
	}

	scale := p999 * distScale / distEntryMax
	table := make([]int16, distEntries)
	clipped := 0
	for i := range table {
		q := quantile(residuals, (float64(i)+0.5)/float64(distEntries))
		v := int(math.Round(q / scale * distScale))
		if v < 0 {
			v = 0
		}
		if v > distEntryMax {
			v = distEntryMax
			clipped++
		}
		table[i] = int16(v)
	}

	cal := Calibration{
		Schema:          1,
		Profile:         prof.ID,
		Basis:           "measured",
		HopModel:        "rtt_floor_split",
		SourceAudit:     auditID,
		ProfileDelayMs:  prof.DelayMs,
		FloorMs:         prof.DelayMs,
		HopDelayMs:      prof.DelayMs / 2,
		ScaleMs:         scale,
		TableSize:       distEntries,
		DistFile:        prof.ID + ".dist",
		Samples:         len(sorted),
		BelowFloorPct:   100 * float64(below) / float64(len(sorted)),
		ClippedPct:      100 * float64(clipped) / float64(distEntries),
		SourceP50RTTms:  p50,
		SourceP95RTTms:  quantile(sorted, 0.95),
		SourceP999RTTms: quantile(sorted, 0.999),
		SourceMaxRTTms:  sorted[len(sorted)-1],
		Table:           table,
	}
	cal.recomputeRoundTrip()
	return cal, nil
}

func quantile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	i := int(math.Round(p * float64(len(sorted)-1)))
	if i < 0 {
		i = 0
	}
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}

func (c *Calibration) recomputeRoundTrip() {
	if len(c.Table) == 0 || c.ScaleMs <= 0 {
		return
	}
	vals := make([]float64, len(c.Table))
	for i, v := range c.Table {
		vals[i] = c.FloorMs + c.ScaleMs*float64(v)/distScale
	}
	sort.Float64s(vals)
	c.RoundTripP50RTTms = quantile(vals, 0.50)
	c.RoundTripP95RTTms = quantile(vals, 0.95)
	c.RoundTripP999RTTms = quantile(vals, 0.999)
}

// Render — format netem natif : un entier décimal par ligne, sans padding
// (strtol base 0 lirait « 010 » comme 8) ni commentaire.
func (c Calibration) Render() []byte {
	var b strings.Builder
	for _, v := range c.Table {
		b.WriteString(strconv.Itoa(int(v)))
		b.WriteByte('\n')
	}
	return []byte(b.String())
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/qdisc -run 'TestBuildCalibration|TestQuantile' -v`

Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add pkg/qdisc/dist.go pkg/qdisc/dist_test.go
git commit -m "qdisc: build and render empirical netem dist"
```

---

### Task 3: Sealed calibration write and load

**Files:**
- Modify: `pkg/qdisc/dist.go`
- Test: `pkg/qdisc/dist_test.go` (append)

**Interfaces:**
- Consumes: `Calibration`, `Render`, `distEntries` from Task 2.
- Produces:
  - `func WriteCalibration(dir string, cal Calibration) error`
  - `func LoadCalibration(dir, profile string) (Calibration, error)`
  - `func writeFileAtomic(path string, data []byte) error`
  - `func safeCalibrationName(id string) error`

- [ ] **Step 1: Write the failing tests**

Append to `pkg/qdisc/dist_test.go` (extend imports with `os`, `path/filepath`):

```go
func TestWriteLoadCalibrationSealsDist(t *testing.T) {
	dir := t.TempDir()
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = 100 + float64(i%8)
	}
	cal, err := BuildCalibration(model.Profile{ID: "P2", DelayMs: 100}, "audit-t", rtts)
	if err != nil {
		t.Fatal(err)
	}
	cal.SourceTelemetrySHA256 = "deadbeef"
	cal.Created = "2026-10-27T00:00:00Z"
	if err := WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCalibration(dir, "P2")
	if err != nil {
		t.Fatal(err)
	}
	if got.DistSHA256 == "" || got.SourceAudit != "audit-t" || got.TableSize != distEntries {
		t.Fatalf("sidecar relu : %+v", got)
	}
	if got.DistFile != "P2.dist" {
		t.Fatalf("dist_file = %q", got.DistFile)
	}
	if err := os.WriteFile(filepath.Join(dir, "P2.dist"), []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCalibration(dir, "P2"); err == nil {
		t.Fatal("hash divergent accepté")
	}
}

func TestLoadCalibrationMissingFailsClosed(t *testing.T) {
	if _, err := LoadCalibration(t.TempDir(), "P2"); err == nil {
		t.Fatal("calibration absente acceptée")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/qdisc -run 'TestWriteLoadCalibration|TestLoadCalibrationMissing' -v`

Expected: FAIL with `undefined: WriteCalibration`.

- [ ] **Step 3: Implement write and load**

Extend the import block of `pkg/qdisc/dist.go`:

```go
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/pkg/model"
)
```

Append:

```go
// WriteCalibration — écrit <profil>.dist puis <profil>.json, chacun remplacé
// atomiquement. Le sidecar scelle les octets du .dist ; un .dist orphelin
// (sidecar absent) est refusé à la lecture, jamais deviné.
func WriteCalibration(dir string, cal Calibration) error {
	if err := safeCalibrationName(cal.Profile); err != nil {
		return err
	}
	if cal.Schema != 1 {
		return fmt.Errorf("calibration %s: schema %d, attendu 1", cal.Profile, cal.Schema)
	}
	if cal.Basis != "measured" {
		return fmt.Errorf("calibration %s: basis %q, attendu measured", cal.Profile, cal.Basis)
	}
	if cal.HopModel != "rtt_floor_split" {
		return fmt.Errorf("calibration %s: hop_model %q, attendu rtt_floor_split", cal.Profile, cal.HopModel)
	}
	if cal.TableSize != distEntries || len(cal.Table) != distEntries {
		return fmt.Errorf("calibration %s: table %d entrées (sidecar %d), attendu %d", cal.Profile, len(cal.Table), cal.TableSize, distEntries)
	}
	if cal.ScaleMs <= 0 || cal.FloorMs <= 0 {
		return fmt.Errorf("calibration %s: échelle/plancher invalide", cal.Profile)
	}
	if cal.DistFile == "" {
		cal.DistFile = cal.Profile + ".dist"
	}
	if cal.DistFile != cal.Profile+".dist" {
		return fmt.Errorf("calibration %s: fichier %q incohérent", cal.Profile, cal.DistFile)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("calibration: répertoire %s: %w", dir, err)
	}

	raw := cal.Render()
	sum := sha256.Sum256(raw)
	cal.DistSHA256 = hex.EncodeToString(sum[:])
	if err := writeFileAtomic(filepath.Join(dir, cal.DistFile), raw); err != nil {
		return err
	}
	sidecar, err := json.MarshalIndent(cal, "", "  ")
	if err != nil {
		return fmt.Errorf("calibration %s: sidecar: %w", cal.Profile, err)
	}
	sidecar = append(sidecar, '\n')
	return writeFileAtomic(filepath.Join(dir, cal.Profile+".json"), sidecar)
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("calibration: écriture %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("calibration: renommage %s: %w", path, err)
	}
	return nil
}

func safeCalibrationName(id string) error {
	if id == "" || id == "." || id == ".." || id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("calibration: identifiant invalide %q", id)
	}
	return nil
}

// LoadCalibration — lecture fail-closed : sidecar présent, profil exact,
// .dist présent et de hash identique. Toute divergence arrête la campagne
// AVANT la première mutation qdisc ; jamais de repli synthétique.
func LoadCalibration(dir, profile string) (Calibration, error) {
	if err := safeCalibrationName(profile); err != nil {
		return Calibration{}, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, profile+".json"))
	if err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar absent (%w) — lancer cgo calibrate", profile, err)
	}
	var cal Calibration
	if err := json.Unmarshal(raw, &cal); err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar illisible: %w", profile, err)
	}
	if cal.Schema != 1 || cal.Profile != profile {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar incohérent (schema %d, profil %q)", profile, cal.Schema, cal.Profile)
	}
	if cal.Basis != "measured" {
		return Calibration{}, fmt.Errorf("calibration %s: basis %q, attendu measured", profile, cal.Basis)
	}
	if cal.HopModel != "rtt_floor_split" {
		return Calibration{}, fmt.Errorf("calibration %s: hop_model %q, attendu rtt_floor_split", profile, cal.HopModel)
	}
	if cal.FloorMs <= 0 || cal.ScaleMs <= 0 || cal.TableSize != distEntries {
		return Calibration{}, fmt.Errorf("calibration %s: plancher/échelle/table invalide", profile)
	}
	if cal.DistFile != profile+".dist" {
		return Calibration{}, fmt.Errorf("calibration %s: fichier %q incohérent", profile, cal.DistFile)
	}
	dist, err := os.ReadFile(filepath.Join(dir, cal.DistFile))
	if err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: .dist absent: %w", profile, err)
	}
	sum := sha256.Sum256(dist)
	if hex.EncodeToString(sum[:]) != cal.DistSHA256 {
		return Calibration{}, fmt.Errorf("calibration %s: hash .dist divergent — artefact modifié ou incomplet", profile)
	}
	if got := bytes.Count(dist, []byte{'\n'}); got != cal.TableSize {
		return Calibration{}, fmt.Errorf("calibration %s: .dist %d lignes, sidecar %d", profile, got, cal.TableSize)
	}
	return cal, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/qdisc -run 'TestWriteLoadCalibration|TestLoadCalibrationMissing' -v`

Expected: PASS (2 tests).

- [ ] **Step 5: Run the package suite**

Run: `go test ./pkg/qdisc`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/qdisc/dist.go pkg/qdisc/dist_test.go
git commit -m "qdisc: seal and load dist calibrations"
```

---

### Task 4: Netem distribution arguments and `TC_LIB_DIR`

**Files:**
- Modify: `pkg/qdisc/ge.go`
- Modify: `pkg/qdisc/qdisc.go`
- Test: `pkg/qdisc/dist_netem_test.go` (create)

**Interfaces:**
- Consumes: `TCRunner`, `GEModel`, existing `FakeRunner`.
- Produces:
  - `func ApplyNetemDistribution(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel, dist string) error`
  - `func applyNetemArgs(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel, dist string) error`
  - `NsRunner.Env []string`
  - `func nsTCArgs(ns string, env, args []string) []string`

- [ ] **Step 1: Write the failing tests**

Create `pkg/qdisc/dist_netem_test.go`:

```go
package qdisc

import (
	"strings"
	"testing"
)

func TestApplyNetemDistributionArgs(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 0.25, 0.5, nil, "P2"); err != nil {
		t.Fatal(err)
	}
	want := []string{"qdisc", "replace", "dev", "w-edge", "root", "handle", "1:", "netem",
		"delay", "50ms", "0.25ms", "loss", "0.5%", "distribution", "P2"}
	if strings.Join(f.Calls[0], " ") != strings.Join(want, " ") {
		t.Fatalf("args = %v, want %v", f.Calls[0], want)
	}
}

func TestApplyNetemDistributionRequiresNameAndJitter(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 1, 0, nil, ""); err == nil {
		t.Fatal("nom vide accepté")
	}
	if err := ApplyNetemDistribution(f, "w-edge", 50, 0, 0, nil, "P2"); err == nil {
		t.Fatal("jitter nul accepté")
	}
	if len(f.Calls) != 0 {
		t.Fatalf("tc appelé malgré un contrat invalide : %v", f.Calls)
	}
}

func TestApplyNetemDistributionWithGE(t *testing.T) {
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001}
	if err := ApplyNetemDistribution(f, "c-wan", 300, 2.5, 1, &ge, "P3"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Calls[0], " ")
	if !strings.Contains(joined, "loss gemodel 0.002 0.1 0.5 0.001") || !strings.Contains(joined, "distribution P3") {
		t.Fatalf("gemodel + distribution attendus : %v", f.Calls[0])
	}
}

func TestNsTCArgsCarriesEnv(t *testing.T) {
	got := nsTCArgs("ns_wan_emulator", []string{"TC_LIB_DIR=/tmp/cal"}, []string{"qdisc", "show"})
	want := []string{"env", "TC_LIB_DIR=/tmp/cal", "ip", "netns", "exec", "ns_wan_emulator", "tc", "qdisc", "show"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("args = %v, want %v", got, want)
	}
	if plain := nsTCArgs("ns", nil, []string{"qdisc"}); strings.Join(plain, " ") != "ip netns exec ns tc qdisc" {
		t.Fatalf("sans env = %v", plain)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/qdisc -run 'TestApplyNetemDistribution|TestNsTCArgs' -v`

Expected: FAIL with `undefined: ApplyNetemDistribution`.

- [ ] **Step 3: Refactor the netem argument builder**

In `pkg/qdisc/ge.go`, replace the bodies of `ApplyNetemGE` and `ApplyNetemPhysics`, and add the new functions:

```go
// ApplyNetemGE — netem racine avec perte Gilbert-Elliott (forme documentée
// à 4 arguments). Handle 1: conservé (un shaper peut s'empiler en 1:1).
func ApplyNetemGE(r TCRunner, iface string, delayMs, jitterMs float64, ge GEModel) error {
	if err := applyNetemArgs(r, iface, delayMs, jitterMs, 0, &ge, ""); err != nil {
		return fmt.Errorf("netem ge %s: %w", iface, err)
	}
	return nil
}

// ApplyNetemPhysics — le netem du saut de propagation : délai + gigue, et
// perte UNIFORME (ge nil, forme historique à l'octet près — le corpus gelé
// se reproduit par là) ou Gilbert-Elliott (ge fourni, 4 arguments). Une
// seule des deux formes de perte, jamais les deux.
func ApplyNetemPhysics(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel) error {
	if err := applyNetemArgs(r, iface, delayMs, jitterMs, lossPct, ge, ""); err != nil {
		return fmt.Errorf("netem %s: %w", iface, err)
	}
	return nil
}

// ApplyNetemDistribution — netem du saut mesuré : la table empirique
// remplace la gigue déclarée. tc exige delay ET jitter avec
// « distribution » ; un nom ou une gigue invalide est refusé AVANT tc.
func ApplyNetemDistribution(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel, dist string) error {
	if dist == "" {
		return fmt.Errorf("netem dist %s: nom de distribution vide", iface)
	}
	if jitterMs <= 0 {
		return fmt.Errorf("netem dist %s: jitter %.6g ≤ 0 (tc exige delay+jitter)", iface, jitterMs)
	}
	if err := applyNetemArgs(r, iface, delayMs, jitterMs, lossPct, ge, dist); err != nil {
		return fmt.Errorf("netem dist %s: %w", iface, err)
	}
	return nil
}

func applyNetemArgs(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel, dist string) error {
	args := []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs)}
	if ge != nil {
		args = append(args, "loss", "gemodel",
			fmt.Sprintf("%g", ge.GoodToBad),
			fmt.Sprintf("%g", ge.BadToGood),
			fmt.Sprintf("%g", ge.LossBad),
			fmt.Sprintf("%g", ge.LossGood))
	} else if lossPct > 0 {
		args = append(args, "loss", fmt.Sprintf("%g%%", lossPct))
	}
	if dist != "" {
		args = append(args, "distribution", dist)
	}
	_, err := r.Run(args...)
	return err
}
```

- [ ] **Step 4: Add `NsRunner.Env`**

In `pkg/qdisc/qdisc.go`, replace the `NsRunner` declaration and `Run` method:

```go
// NsRunner runs tc inside a network namespace (ip netns exec NS tc …).
// Env porte les variables à garder à travers sudo (TC_LIB_DIR pour .dist) :
// « sudo env VAR=… ip netns exec … tc » ne dépend pas d'env_keep.
type NsRunner struct {
	Ns  string
	Env []string // paires VAR=value
}

func (r NsRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("sudo", nsTCArgs(r.Ns, r.Env, args)...).CombinedOutput()
}

func nsTCArgs(ns string, env, args []string) []string {
	full := make([]string, 0, len(env)+len(args)+6)
	if len(env) > 0 {
		full = append(full, "env")
		full = append(full, env...)
	}
	full = append(full, "ip", "netns", "exec", ns, "tc")
	return append(full, args...)
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./pkg/qdisc -v`

Expected: PASS, including the existing `TestApplyNetemGEFourArguments`, `TestApplyNetemPhysicsUniformUnchanged`, and `TestApplyNetemPhysicsBursty`.

- [ ] **Step 6: Commit**

```bash
git add pkg/qdisc/ge.go pkg/qdisc/qdisc.go pkg/qdisc/dist_netem_test.go
git commit -m "qdisc: apply netem distribution with TC_LIB_DIR"
```

---

### Task 5: `cgo calibrate` command

**Files:**
- Create: `cmd/cgo/calibrate.go`
- Modify: `cmd/cgo/main.go` (switch + usage)
- Test: `cmd/cgo/calibrate_test.go` (create)

**Interfaces:**
- Consumes: `audit.ReadTelemetry`, `qdisc.BuildCalibration`, `qdisc.WriteCalibration`, `model.Profiles`, `profile.Load`.
- Produces: `func runCalibrate(args []string) int`, `func fileSHA256(path string) (string, error)`.

- [ ] **Step 1: Write the failing test**

Create `cmd/cgo/calibrate_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

func TestRunCalibrateWritesSealedArtifacts(t *testing.T) {
	dir := t.TempDir()
	res := &audit.Result{AuditID: "audit-cal"}
	for i := 0; i < 64; i++ {
		phase := "baseline"
		rtt := 100 + float64(i%8)
		if i >= 48 {
			phase = "load"
			rtt = 500
		}
		res.Telemetry = append(res.Telemetry, audit.TelemetryRow{Phase: phase, OffsetMs: int64(i) * 400, RTTms: rtt})
	}
	if _, err := audit.WriteTelemetry(dir, res); err != nil {
		t.Fatal(err)
	}
	report := `{"site":"TEST","link_type":"5g"}`
	if err := os.WriteFile(filepath.Join(dir, "audits", "audit-cal", "audit_report.json"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "calibrations")
	if code := runCalibrate([]string{"--audit", "audit-cal", "--profile", "P2", "--data", dir, "--out", out}); code != 0 {
		t.Fatalf("runCalibrate code %d", code)
	}
	cal, err := qdisc.LoadCalibration(out, "P2")
	if err != nil {
		t.Fatal(err)
	}
	if cal.SourceAudit != "audit-cal" || cal.SourceSite != "TEST" || cal.SourceLinkType != "5g" {
		t.Fatalf("provenance : %+v", cal)
	}
	if cal.Basis != "measured" || cal.HopModel != "rtt_floor_split" || cal.SourceTelemetrySHA256 == "" || cal.SourceReportSHA256 == "" {
		t.Fatalf("sceaux/provenance : %+v", cal)
	}
	if cal.Samples != 48 {
		t.Fatalf("échantillons baseline = %d, want 48 (les lignes load sont exclues)", cal.Samples)
	}
	if cal.SourceP999RTTms > 200 {
		t.Fatalf("ligne load comptée dans la queue : p99.9 %.1f", cal.SourceP999RTTms)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/cgo -run TestRunCalibrate -v`

Expected: FAIL with `undefined: runCalibrate`.

- [ ] **Step 3: Implement the command**

Create `cmd/cgo/calibrate.go`:

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/profile"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// runCalibrate — transforme une trace terrain (telemetry.csv, phase baseline
// seulement) en data/calibrations/<profil>.dist + sidecar scellé.
func runCalibrate(args []string) int {
	profile.Load()
	fs := flag.NewFlagSet("calibrate", flag.ExitOnError)
	auditID := fs.String("audit", "", "identifiant de l'audit source (data/audits/<id>)")
	profileID := fs.String("profile", "", "profil à calibrer (P2, P3, yas_4g…)")
	dataDir := fs.String("data", "data", "racine des artefacts d'audit")
	outDir := fs.String("out", "", "répertoire des .dist/.json (défaut <data>/calibrations)")
	_ = fs.Parse(args)
	if *auditID == "" || *profileID == "" {
		fmt.Fprintln(os.Stderr, "calibrate: --audit et --profile requis")
		return 2
	}
	if *outDir == "" {
		*outDir = filepath.Join(*dataDir, "calibrations")
	}

	model.ProfilesMu.RLock()
	prof, ok := model.Profiles[*profileID]
	model.ProfilesMu.RUnlock()
	if !ok {
		fmt.Fprintf(os.Stderr, "calibrate: profil %q inconnu\n", *profileID)
		return 2
	}

	rows, err := audit.ReadTelemetry(*dataDir, *auditID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	rtts := make([]float64, 0, len(rows))
	for _, r := range rows {
		if r.Phase == model.PhaseBaseline && r.RTTms > 0 {
			rtts = append(rtts, r.RTTms)
		}
	}
	cal, err := qdisc.BuildCalibration(prof, *auditID, rtts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}

	// provenance : la trace et son rapport d'audit, scellés s'ils existent.
	telPath := filepath.Join(*dataDir, "audits", *auditID, "telemetry.csv")
	sum, err := fileSHA256(telPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	cal.SourceTelemetrySHA256 = sum

	repPath := filepath.Join(*dataDir, "audits", *auditID, "audit_report.json")
	if raw, err := os.ReadFile(repPath); err == nil {
		reportSum := sha256.Sum256(raw)
		cal.SourceReportSHA256 = hex.EncodeToString(reportSum[:])
		var rep struct {
			Site     string `json:"site"`
			LinkType string `json:"link_type"`
		}
		if err := json.Unmarshal(raw, &rep); err != nil {
			fmt.Fprintln(os.Stderr, "calibrate: rapport illisible:", err)
			return 1
		}
		cal.SourceSite, cal.SourceLinkType = rep.Site, rep.LinkType
	} else if !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}

	cal.Created = time.Now().Format(time.RFC3339)
	if err := qdisc.WriteCalibration(*outDir, cal); err != nil {
		fmt.Fprintln(os.Stderr, "calibrate:", err)
		return 1
	}
	fmt.Printf("calibrate %s — %d échantillons baseline · plancher %.1f ms · échelle %.3f ms · p50 %.1f / p95 %.1f / p99.9 %.1f ms → %s\n",
		cal.Profile, cal.Samples, cal.FloorMs, cal.ScaleMs,
		cal.SourceP50RTTms, cal.SourceP95RTTms, cal.SourceP999RTTms,
		filepath.Join(*outDir, cal.DistFile))
	if cal.BelowFloorPct > 0 || cal.ClippedPct > 0 {
		fmt.Printf("attention: %.1f %% des RTT sous le plancher, %.1f %% de la table écrêtée — relire le sidecar\n", cal.BelowFloorPct, cal.ClippedPct)
	}
	return 0
}

func fileSHA256(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
```

- [ ] **Step 4: Wire the command**

In `cmd/cgo/main.go`, add the dispatch case after `case "run":`'s block (before `default:`):

```go
	case "calibrate":
		os.Exit(runCalibrate(os.Args[2:]))
```

Add to `usage()` before the `cgo verify` line:

```
       cgo calibrate --audit audit-XXXX --profile P2   # trace baseline → .dist netem
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./cmd/cgo -run TestRunCalibrate -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/cgo/calibrate.go cmd/cgo/calibrate_test.go cmd/cgo/main.go
git commit -m "cmd: add cgo calibrate"
```

---

### Task 6: Fail-closed campaign dist mode

**Files:**
- Modify: `pkg/campagne/campagne.go` (`Deps`, `defaults`)
- Modify: `pkg/campagne/dualns.go` (`ProdDepsDualNS`, `applyPhysics`)
- Modify: `pkg/campagne/prod.go` (carry the env on legacy deps)
- Test: `pkg/campagne/dualns_test.go` (append)

**Interfaces:**
- Consumes: `qdisc.LoadCalibration`, `qdisc.ApplyNetemDistribution`, `qdisc.ApplyNetemPhysics`, `qdisc.NsRunner.Env`, `qdisc.WriteCalibration`.
- Produces:
  - `Deps.DelayModel string`
  - `Deps.CalibrationDir string`
  - `func applyPhysicsDist(d *Deps, prof model.Profile, ge *qdisc.GEModel, ifaces []string) error`

- [ ] **Step 1: Write the failing tests**

Append to `pkg/campagne/dualns_test.go` (its imports already cover `strings`, `testing`, `model`, and `qdisc`):

```go
func writeDistCalibration(t *testing.T, dir, profile string, delay float64) {
	t.Helper()
	rtts := make([]float64, 64)
	for i := range rtts {
		rtts[i] = delay + float64(i%8)
	}
	cal, err := qdisc.BuildCalibration(model.Profile{ID: profile, DelayMs: delay}, "audit-test", rtts)
	if err != nil {
		t.Fatal(err)
	}
	cal.SourceTelemetrySHA256 = "test"
	cal.Created = "2026-10-27T00:00:00Z"
	if err := qdisc.WriteCalibration(dir, cal); err != nil {
		t.Fatal(err)
	}
}

func TestApplyPhysicsDistMeasuredHop(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	f := &qdisc.FakeRunner{}
	d := Deps{TC: f, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}
	if err := applyPhysics(&d, model.Profiles["P2"], nil); err != nil {
		t.Fatal(err)
	}
	var up, down string
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if !strings.Contains(joined, "netem") {
			continue
		}
		if strings.Contains(joined, "dev w-edge") {
			up = joined
		}
		if strings.Contains(joined, "dev c-wan") {
			down = joined
		}
	}
	if !strings.Contains(up, "delay 50ms") || !strings.Contains(up, "distribution P2") {
		t.Fatalf("w-edge doit porter la table : %q", up)
	}
	if !strings.Contains(down, "delay 50ms 0ms") || strings.Contains(down, "distribution") {
		t.Fatalf("c-wan doit rester au plancher sans table : %q", down)
	}
}

func TestApplyPhysicsDistDownUsesDownlinkHop(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	f := &qdisc.FakeRunner{}
	d := Deps{TC: f, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir, Direction: "down"}
	if err := applyPhysics(&d, model.Profiles["P2"], nil); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, "dev c-wan") && strings.Contains(joined, "netem") && !strings.Contains(joined, "distribution P2") {
			t.Fatalf("c-wan mesuré doit porter la table : %q", joined)
		}
		if strings.Contains(joined, "dev w-edge") && strings.Contains(joined, "distribution") {
			t.Fatalf("w-edge non mesuré ne doit pas porter la table : %q", joined)
		}
	}
}

func TestApplyPhysicsDistFailsClosed(t *testing.T) {
	dir := t.TempDir()
	writeDistCalibration(t, dir, "P2", 100)
	stale := model.Profiles["P2"]
	stale.DelayMs = 120
	cases := []struct {
		name string
		deps Deps
		prof model.Profile
	}{
		{"calibration absente", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: t.TempDir()}, model.Profiles["P2"]},
		{"direction both", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir, Direction: "both"}, model.Profiles["P2"]},
		{"banc legacy", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "legacy", DelayModel: "dist", CalibrationDir: dir}, model.Profiles["P2"]},
		{"deux ifaces", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan", "c-wan2"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}, model.Profiles["P2"]},
		{"plancher périmé", Deps{TC: &qdisc.FakeRunner{}, CliIf: "w-edge", PhysicsIfaces: []string{"c-wan"}, Bench: "dualns", DelayModel: "dist", CalibrationDir: dir}, stale},
	}
	for _, tc := range cases {
		f := tc.deps.TC.(*qdisc.FakeRunner)
		if err := applyPhysics(&tc.deps, tc.prof, nil); err == nil {
			t.Fatalf("%s : erreur attendue", tc.name)
		}
		if len(f.Calls) != 0 {
			t.Fatalf("%s : tc appelé avant l'échec : %v", tc.name, f.Calls)
		}
	}
}

func TestProdDepsDualNSDistEnv(t *testing.T) {
	t.Setenv("CGO_DELAY_MODEL", "dist")
	t.Setenv("CGO_CALIBRATION_DIR", t.TempDir())
	d := ProdDepsDualNS("up")
	if d.DelayModel != "dist" || d.CalibrationDir == "" {
		t.Fatalf("dist non câblé : %+v", d)
	}
	ns, ok := d.TC.(qdisc.NsRunner)
	if !ok || len(ns.Env) != 1 || ns.Env[0] != "TC_LIB_DIR="+d.CalibrationDir {
		t.Fatalf("NsRunner.Env = %+v (calib %q)", d.TC, d.CalibrationDir)
	}
}

func TestProdDepsLegacyDistWired(t *testing.T) {
	t.Setenv("CGO_DELAY_MODEL", "dist")
	if d := ProdDeps(); d.DelayModel != "dist" {
		t.Fatalf("CGO_DELAY_MODEL ignoré par le banc legacy : %+v", d)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/campagne -run 'TestApplyPhysicsDist|TestProdDepsDualNSDistEnv|TestProdDepsLegacyDistWired' -v`

Expected: FAIL with `unknown field DelayModel in struct literal`.

- [ ] **Step 3: Add the deps fields and defaults**

In `pkg/campagne/campagne.go`, add after `Bench string`:

```go
	// DelayModel — "scalar" (défaut : delay/jitter déclarés) | "dist"
	// (table empirique par profil, chargée fail-closed AVANT toute mutation).
	DelayModel string
	// CalibrationDir — répertoire des <profil>.dist/.json.
	CalibrationDir string
```

In `defaults`, after the `Bench` default, add:

```go
	if d.DelayModel == "" {
		d.DelayModel = "scalar"
	}
```

- [ ] **Step 4: Wire the environment in both benches**

In `pkg/campagne/dualns.go`, add `path/filepath` to the imports. Replace `ProdDepsDualNS` with:

```go
func ProdDepsDualNS(direction string) Deps {
	target := env("CGO_TARGET", "10.30.0.2")
	small := env("CGO_SMALL_URL", "http://10.30.0.2:8081/small")
	bulk := env("CGO_BULK_ADDR", "10.30.0.2:5201")
	shaperIf := env("CGO_SHAPER_IF", "e-srv") // émission edge → serveur : le upload
	if direction == "down" {
		shaperIf = env("CGO_SHAPER_IF", "e-wan") // émission edge → client : le download
	}
	edge := qdisc.NsRunner{Ns: env("CGO_EDGE_NS", "ns_edge_router")}
	delayModel := env("CGO_DELAY_MODEL", "scalar")
	calibDir := env("CGO_CALIBRATION_DIR", filepath.Join("data", "calibrations"))
	wan := qdisc.NsRunner{Ns: env("CGO_WAN_NS", "ns_wan_emulator")}
	if delayModel == "dist" {
		if abs, err := filepath.Abs(calibDir); err == nil {
			calibDir = abs
		}
		wan.Env = []string{"TC_LIB_DIR=" + calibDir}
	}
	return Deps{
		TC:             wan,
		TCShaper:       edge,
		CliIf:          env("CGO_CLI_IF", "w-edge"),
		PhysicsIfaces:  []string{env("CGO_DOWNLINK_PHYS_IF", "c-wan")},
		ShaperIf:       shaperIf,
		Target:         target,
		SmallURL:       small,
		BulkAddr:       bulk,
		Direction:      direction,
		Bench:          "dualns",
		ChannelModel:   env("CGO_CHANNEL_MODEL", "clean"),
		DelayModel:     delayModel,
		CalibrationDir: calibDir,
		BaselineSec:    envInt("CGO_BASELINE_SEC", 0),
		ChargeSec:      envInt("CGO_CHARGE_SEC", 0),
		RecupSec:       envInt("CGO_RECUP_SEC", 0),
		BulkN: func(ctx context.Context, addr, cc string, n int) ([]uint64, error) {
			return probe.BulkSendNTo(ctx, addr, cc, n)
		},
		StatsFn: func() []qdisc.Stats {
			// UN SEUL saut, celui du sens mesuré (l'émission façonnée) —
			// sommer les deux directions double-compterait chaque octet.
			sts, err := qdisc.PollStats(edge, shaperIf)
			if err != nil {
				return nil
			}
			return sts
		},
	}
}
```

In `pkg/campagne/prod.go`, add `path/filepath` to the imports and add to the returned `Deps` (after `Direction: direction`):

```go
		DelayModel:     env("CGO_DELAY_MODEL", "scalar"),
		CalibrationDir: env("CGO_CALIBRATION_DIR", filepath.Join("data", "calibrations")),
```

- [ ] **Step 5: Branch `applyPhysics` and add `applyPhysicsDist`**

In `pkg/campagne/dualns.go`, replace `applyPhysics` with:

```go
func applyPhysics(d *Deps, prof model.Profile, ge *qdisc.GEModel) error {
	ifaces := append([]string{d.CliIf}, d.PhysicsIfaces...)
	switch d.DelayModel {
	case "", "scalar":
	case "dist":
		return applyPhysicsDist(d, prof, ge, ifaces)
	default:
		return fmt.Errorf("delay_model %q invalide (scalar|dist)", d.DelayModel)
	}
	for _, ifc := range ifaces {
		_, _ = d.TC.Run("qdisc", "del", "dev", ifc, "root")
		if err := qdisc.ApplyNetemPhysics(d.TC, ifc, prof.DelayMs, prof.JitterMs, prof.LossPct, ge); err != nil {
			return fmt.Errorf("netem %s: %w", ifc, err)
		}
	}
	return nil
}

// applyPhysicsDist — physique d'une trace : plancher RTT/2 sur chaque saut,
// table empirique sur le seul saut mesuré. Toute validation (calibration,
// profil, sens, banc) passe AVANT le premier reset/tc : un artefact absent
// ou périmé ne laisse jamais une racine à moitié configurée.
func applyPhysicsDist(d *Deps, prof model.Profile, ge *qdisc.GEModel, ifaces []string) error {
	if d.Direction == "both" {
		return fmt.Errorf("delay_model dist: direction both impossible — la trace mesure un seul sens")
	}
	if d.Bench != "dualns" {
		return fmt.Errorf("delay_model dist: banc dualns requis (banc %q)", d.Bench)
	}
	if d.CalibrationDir == "" {
		return fmt.Errorf("delay_model dist: répertoire de calibration vide (CGO_CALIBRATION_DIR)")
	}
	cal, err := qdisc.LoadCalibration(d.CalibrationDir, prof.ID)
	if err != nil {
		return err
	}
	if cal.FloorMs != prof.DelayMs {
		return fmt.Errorf("calibration %s périmée : plancher %.2f ms ≠ profil %.2f ms (relancer cgo calibrate)", prof.ID, cal.FloorMs, prof.DelayMs)
	}
	if len(d.PhysicsIfaces) != 1 {
		return fmt.Errorf("delay_model dist: une seule PhysicsIfaces attendue, %d fournie(s)", len(d.PhysicsIfaces))
	}
	measured := d.CliIf
	if d.Direction == "down" {
		measured = d.PhysicsIfaces[0]
	}
	for _, ifc := range ifaces {
		_, _ = d.TC.Run("qdisc", "del", "dev", ifc, "root")
		if ifc == measured {
			if err := qdisc.ApplyNetemDistribution(d.TC, ifc, cal.FloorMs/2, cal.ScaleMs, prof.LossPct, ge, prof.ID); err != nil {
				return fmt.Errorf("netem dist %s: %w", ifc, err)
			}
			continue
		}
		if err := qdisc.ApplyNetemPhysics(d.TC, ifc, cal.FloorMs/2, 0, prof.LossPct, ge); err != nil {
			return fmt.Errorf("netem dist %s: %w", ifc, err)
		}
	}
	return nil
}
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./pkg/campagne -run 'TestApplyPhysicsDist|TestProdDepsDualNSDistEnv|TestProdDepsLegacyDistWired' -v`

Expected: PASS (5 tests).

- [ ] **Step 7: Run the package suites**

Run: `go test ./pkg/campagne ./pkg/qdisc ./pkg/audit ./cmd/cgo`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add pkg/campagne/campagne.go pkg/campagne/dualns.go pkg/campagne/prod.go pkg/campagne/dualns_test.go
git commit -m "campagne: opt-in fail-closed dist delay model"
```

---

### Task 7: Methodology docs and full gates

**Files:**
- Modify: `docs/methodology.fr.md`

**Interfaces:**
- Consumes: the command, mode, and conventions from Tasks 1–6.
- Produces: no code.

- [ ] **Step 1: Insert the methodology subsection**

In `docs/methodology.fr.md`, after the paragraph ending “(`pkg/audit` → `data/link_audit.csv`).” and before `## 3. Plan factoriel`, insert:

````markdown
### 2.1 Distribution empirique (`.dist`)

Le délai d'un profil peut être rejoué depuis une trace d'audit plutôt que
depuis le couple déclaré `delay/jitter`. La calibration ne retient que la
phase **baseline** de `data/audits/<id>/telemetry.csv` : les échantillons
sous charge décrivent l'AQM testé, pas le lien.

```
cgo calibrate --audit audit-XXXX --profile P2
```

Le profil doit exister (`pkg/model`, import compris) et sa médiane RTT doit
dépasser son `delay_ms`. La commande écrit `data/calibrations/P2.dist`
(table netem native, 1024 entrées, échelle 8192) et `P2.json` (provenance
scellée : audit source, SHA-256 de la télémétrie et du `.dist`, site, type
de lien, percentiles source et reconstruits).

Convention : `delay_ms` est un **RTT**. Le mode dist pose `plancher/2` sur
chaque saut et la table empirique sur le seul saut mesuré (`w-edge` en
montée, `c-wan` en descente), donc RTT ≈ plancher + résidu. Le mode est
opt-in :

```
CGO_DELAY_MODEL=dist CGO_CALIBRATION_DIR=$PWD/data/calibrations \
  cgo run --bench dualns --profiles P2 --direction up
```

Une calibration absente, périmée (profil modifié) ou dont le `.dist` ne
correspond plus au sceau arrête l'événement **avant** toute mutation qdisc :
jamais de repli synthétique. `both` est refusé (une trace = un sens). Le
scalaire reste le défaut ; son écart RTT 2× connu est traité dans la tranche
BBRv3. Les traces `SMOKE-TEST` locales ne sont pas une preuve terrain et
aucun `.dist` qui en dérive n'est cité.
````

- [ ] **Step 2: Run the Windows gates**

Run: `go test ./...`

Expected: PASS.

Run: `go vet ./...`

Expected: no output.

Run: `GOOS=linux GOARCH=amd64 go build ./...`

Expected: no output.

- [ ] **Step 3: Run the frozen guard**

Run: `python book/tools/audit-archive.py`

Expected: `159 runs`, `208 valid lines` (the exact wording may differ; the two numbers must match). If `book/` is absent, record the guard as “not runnable in this checkout” and do not claim it passed.

- [ ] **Step 4: Check the worktree**

Run: `git status --short`

Expected: only the doc/spec/plan files and the Go/test files from this slice
are new or modified. The pre-existing untracked eBPF spec/plan and
`epbfBBR.md` stay untracked. No `data/` artifact, no `data/runs/` change.

- [ ] **Step 5: Commit**

```bash
git add docs/methodology.fr.md docs/superpowers/specs/2026-10-27-cgo-trace-driven-dist-design.md docs/superpowers/plans/2026-10-27-cgo-trace-driven-dist.md
git commit -m "docs: methodologie distribution empirique"
```

---

### Task 8: Linux 6.8 netem acceptance (blocked, do not fake)

**Files:** none.

**Interfaces:**
- Consumes: the built `cgo-linux` binary and a calibration from a real trace.
- Produces: acceptance evidence, or a recorded blocker.

This task is **not** part of the Windows gates. Do not claim it passed
until every step runs on the stable Linux 6.8 bench.

- [ ] **Step 1: Build on Linux**

```bash
go build -o cgo-linux ./cmd/cgo
./cgo-linux version
```

Expected: a Linux binary; no clang, no BPF, no new runtime dependency.

- [ ] **Step 2: Calibrate a real trace**

```bash
./cgo-linux calibrate --audit <field-audit-id> --profile P2
ls -l data/calibrations/P2.dist data/calibrations/P2.json
```

Expected: both files exist; the sidecar's `samples` is at least 30 and its
`source_audit` names the field trace. Do not use a `SMOKE-TEST` trace as
the citable artifact.

- [ ] **Step 3: Apply the table manually**

```bash
sudo env TC_LIB_DIR="$PWD/data/calibrations" \
  ip netns exec ns_wan_emulator \
  tc qdisc replace dev w-edge root handle 1: netem delay 50ms 0.123ms distribution P2
sudo ip netns exec ns_wan_emulator tc qdisc show dev w-edge
```

Expected: tc exits 0; `show` prints the delay/jitter, not the table (the
distribution name is a userspace lookup key and is not echoed back). A
missing file makes tc fail loudly.

- [ ] **Step 4: Run one dist-mode cell**

```bash
CGO_DELAY_MODEL=dist CGO_CALIBRATION_DIR="$PWD/data/calibrations" \
  ./cgo-linux run --bench dualns --profiles P2 --direction up \
  --qdiscs pfifo_fast --cc cubic --reps 1 --run-id dist-accept \
  --data /tmp/cgo-dist-accept
```

Expected: the cell completes; `tc qdisc show` on `w-edge` shows the
delay/jitter (the table itself is not echoed); the RTT p50 is near
`floor + median residual`, not the scalar 100 ms on both hops. Record the
observed values in the slice notes.

- [ ] **Step 5: Prove fail-closed**

```bash
mv data/calibrations/P2.json /tmp/P2.json.bak
CGO_DELAY_MODEL=dist CGO_CALIBRATION_DIR="$PWD/data/calibrations" \
  ./cgo-linux run --bench dualns --profiles P2 --direction up \
  --qdiscs pfifo_fast --cc cubic --reps 1 --run-id dist-fail \
  --data /tmp/cgo-dist-accept
echo $?
sudo ip netns exec ns_wan_emulator tc qdisc show dev w-edge
mv /tmp/P2.json.bak data/calibrations/P2.json
```

Expected: non-zero exit naming the missing sidecar, and the qdisc on
`w-edge` unchanged (no half-applied table).

- [ ] **Step 6: Record the result**

Append a short note to the plan's execution log (not to `epbfBBR.md`):
kernel version, audit ID, sidecar hashes, observed RTTs, and the fail-closed
exit. If the bench is unstable, stop and record the blocker; do not mark the
step complete.
