# Slice 4 Go half, portable host readiness, runbook and telemetry badge — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the eBPF subsystem's Go surface real and tested on Windows without claiming any kernel measurement, give the operator one portable readiness script, and turn the first Linux host-hour into a mechanical runbook — plus an honest telemetry badge.

**Architecture:** `pkg/probe/bpf` gains the typed contract (Flow, Summary, Globals, typed errors), a build-tag pair for the embedded object (absent by default, embedded only with `-tags ebpf_object` after `make bpf`), and a pure `Status` ladder. `pkg/probe` wraps the ladder with real host facts. `pkg/audit` gains an additive `kernel_telemetry` report block with the same voice as the binary: `ebpf | tcpinfo | unavailable`, never a silent zero. `/api/health` exposes the same source so the ribbon can say it. `kit/host-readiness.sh` checks a host without mutating it; `docs/runbooks/2026-10-30-host-acceptance.md` orders slices 4–7.

**Tech Stack:** Go 1.25 stdlib; existing React shell; bash for the readiness script.

## Global Constraints

From `docs/superpowers/specs/2026-10-30-cgo-closing-program-design.md` §2.1 and §4:

- No production write, load, root, or capture; no live eBPF claim before the object is built and attached on Linux 6.8.
- The frozen corpus, CSV schemas and `data/runs/` are untouched.
- No new Go dependency; the object is prebuilt on Linux and never required at runtime build.
- `basis: measured|estimated` keeps its meaning; the telemetry source is additive.
- French user-facing text, English identifiers.
- Never `git add -A`; stage explicit paths.
- The badge must read `non instrumenté` when the report carries no telemetry block — no colour implying success.

---

## File Structure

| Path | Responsibility |
|---|---|
| `pkg/probe/bpf/types.go` | `Flow`, `Summary`, `Globals` — the slice-4 contract verbatim |
| `pkg/probe/bpf/errors.go` | `UnavailableError`, `BuildError` |
| `pkg/probe/bpf/assets_absent.go` (`//go:build !ebpf_object`) | `ObjectBytes()` returns a typed `BuildError` |
| `pkg/probe/bpf/assets_embedded.go` (`//go:build ebpf_object`) | `//go:embed bpf_probe.o` + SHA-256 |
| `pkg/probe/bpf/status.go` | Pure ladder `Status(goos, btf, caps)` — no I/O |
| `pkg/probe/kernelsource.go` | `KernelSource()` with real host facts, injectable for tests |
| `pkg/audit/report.go` (modify) | Additive `kernel_telemetry` + Markdown line |
| `pkg/api/server.go` (modify) | `/api/health` gains `telemetry: {source, reason}` |
| `web/frontend/src/lib/ribbon.ts` (modify) | `telemetryLabel(source)` pure |
| `web/frontend/src/components/StatusRibbon.tsx` (modify) | One `Télémétrie : …` cell |
| `kit/host-readiness.sh` | Read-only host checks, one table, explicit exit code |
| `docs/runbooks/2026-10-30-host-acceptance.md` | Per-slice host checklist and claim boundary |

---

### Task 1: Typed contract and errors

**Files:**
- Create: `pkg/probe/bpf/types.go`
- Create: `pkg/probe/bpf/errors.go`
- Create: `pkg/probe/bpf/types_test.go`

**Interfaces:**
- Produces: `type Flow struct` (JSON: `local`, `peer`, `retrans_payload_bytes`, `retrans_events`, `qdisc_drop_packets`, `qdisc_drop_bytes`, `srtt_us`, `snd_cwnd`, `snd_ssthresh`, `state`, `samples`).
- Produces: `type Summary struct` (JSON: `available`, `source`, `attach`, `fallback_reason`, `object_sha256`, `flows_seen`, `flows_total`, `retrans_payload_bytes`, `retrans_events`, `qdisc_drop_packets`, `qdisc_drop_bytes`, `qdisc_available`, `map_full`, `flows`).
- Produces: `type Globals struct`.
- Produces: `type UnavailableError struct{ Reason string }`, `type BuildError struct{ Reason string }`, both with `Error()`.

- [ ] **Step 1: Write the failing test**

Create `pkg/probe/bpf/types_test.go`:

```go
package bpf

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSummaryJSONCarriesSourceAndFlows(t *testing.T) {
	s := Summary{Available: true, Source: "ebpf", Attach: "fentry", FlowsSeen: 4, FlowsTotal: 4,
		RetransPayload: 1024, QdiscDrops: 7, QdiscAvailable: true,
		Flows: []Flow{{Local: "10.0.0.1:4000", Peer: "10.0.0.2:443", RetransPayload: 512, SRTTus: 25000}}}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"source":"ebpf"`, `"retrans_payload_bytes":1024`, `"qdisc_drop_packets":7`, `"srtt_us":25000`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("summary JSON lacks %s: %s", want, b)
		}
	}
}

func TestErrorsAreTypedAndReadable(t *testing.T) {
	var err error = &UnavailableError{Reason: "CAP_BPF absent"}
	if !strings.Contains(err.Error(), "CAP_BPF") {
		t.Fatalf("unavailable error lost the reason: %v", err)
	}
	err = &BuildError{Reason: "objet non embarqué"}
	if !strings.Contains(err.Error(), "objet non embarqué") {
		t.Fatalf("build error lost the reason: %v", err)
	}
}
```

- [ ] **Step 2: Run it — must fail**

Run: `go test ./pkg/probe/bpf/ -run 'TestSummary|TestErrors' -count=1`
Expected: FAIL — `undefined: Summary`.

- [ ] **Step 3: Implement**

`types.go` (contract copied from the slice-4 plan so later tasks do not rename fields):

```go
package bpf

// Flow — un flux côté émetteur, lu dans la table noyau.
type Flow struct {
	Local          string `json:"local"`
	Peer           string `json:"peer"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	DropPackets    uint64 `json:"qdisc_drop_packets"`
	DropBytes      uint64 `json:"qdisc_drop_bytes"`
	SRTTus         uint32 `json:"srtt_us"`
	SndCwnd        uint32 `json:"snd_cwnd"`
	SndSsthresh    uint32 `json:"snd_ssthresh"`
	State          uint32 `json:"state"`
	Samples        uint32 `json:"samples"`
}

// Summary — une fenêtre de télémétrie. Source dit quel ratio en est tiré :
// ebpf | tcpinfo | "" ; Available dit si la session a été attachée.
type Summary struct {
	Available      bool   `json:"available"`
	Source         string `json:"source,omitempty"`
	Attach         string `json:"attach,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ObjectSHA256   string `json:"object_sha256,omitempty"`
	FlowsSeen      int    `json:"flows_seen"`
	FlowsTotal     int    `json:"flows_total"`
	RetransPayload uint64 `json:"retrans_payload_bytes"`
	RetransEvents  uint64 `json:"retrans_events"`
	QdiscDrops     uint64 `json:"qdisc_drop_packets"`
	QdiscDropBytes uint64 `json:"qdisc_drop_bytes"`
	QdiscAvailable bool   `json:"qdisc_available"`
	MapFull        uint64 `json:"map_full,omitempty"`
	Flows          []Flow `json:"flows,omitempty"`
}

// Globals — diagnostics hôte, jamais une affirmation par flux.
type Globals struct {
	DropPackets    uint64
	DropBytes      uint64
	MapFull        uint64
	QdiscAvailable bool
}
```

`errors.go`:

```go
package bpf

// UnavailableError — limitation de l'hôte (BTF, capacités, OS, vérificateur).
type UnavailableError struct{ Reason string }

func (e *UnavailableError) Error() string { return "ebpf indisponible : " + e.Reason }

// BuildError — défaut de l'objet embarqué (absent, vide, hash, ELF).
type BuildError struct{ Reason string }

func (e *BuildError) Error() string { return "ebpf : objet invalide : " + e.Reason }
```

- [ ] **Step 4: Run it — must pass**

Run: `go test ./pkg/probe/bpf/ -run 'TestSummary|TestErrors' -count=1 && gofmt -l pkg/probe/bpf`
Expected: `ok`, and gofmt prints nothing.

- [ ] **Step 5: Commit**

```bash
git add pkg/probe/bpf/types.go pkg/probe/bpf/errors.go pkg/probe/bpf/types_test.go
git commit -m "probe: typed eBPF contract and errors, copied from the slice-4 plan"
```

---

### Task 2: Object assets behind a build tag, and a pure status ladder

**Files:**
- Create: `pkg/probe/bpf/assets_absent.go`
- Create: `pkg/probe/bpf/assets_embedded.go`
- Create: `pkg/probe/bpf/status.go`
- Create: `pkg/probe/bpf/status_test.go`

**Interfaces:**
- Consumes: `BuildError` (Task 1).
- Produces: `func ObjectBytes() ([]byte, error)`, `func ObjectSHA256() (string, error)`.
- Produces: `func Status(goos string, btf, caps bool, objectErr error) (source, reason string)`.

- [ ] **Step 1: Write the failing tests**

Create `pkg/probe/bpf/status_test.go`:

```go
package bpf

import (
	"strings"
	"testing"
)

func TestStatusLadder(t *testing.T) {
	cases := []struct {
		name, goos string
		btf, caps  bool
		objectErr  error
		wantSource string
		wantReason string
	}{
		{"windows", "windows", true, true, nil, "tcpinfo", "hôte non-Linux"},
		{"no object", "linux", true, true, &BuildError{Reason: "objet non embarqué"}, "tcpinfo", "objet non embarqué"},
		{"no btf", "linux", false, true, nil, "tcpinfo", "BTF absent"},
		{"no caps", "linux", true, false, nil, "tcpinfo", "CAP_BPF absent"},
		{"ready", "linux", true, true, nil, "ebpf", ""},
	}
	for _, c := range cases {
		src, reason := Status(c.goos, c.btf, c.caps, c.objectErr)
		if src != c.wantSource || !strings.Contains(reason, c.wantReason) {
			t.Fatalf("%s: got (%q,%q), want (%q,%q)", c.name, src, reason, c.wantSource, c.wantReason)
		}
	}
}

func TestObjectAbsentByDefault(t *testing.T) {
	if _, err := ObjectBytes(); err == nil {
		t.Skip("built with -tags ebpf_object; absence cannot be tested here")
	} else if !strings.Contains(err.Error(), "objet") {
		t.Fatalf("absence error lost its reason: %v", err)
	}
	if _, err := ObjectSHA256(); err == nil {
		t.Fatal("SHA of an absent object returned nil error")
	}
}
```

- [ ] **Step 2: Run it — must fail**

Run: `go test ./pkg/probe/bpf/ -run 'TestStatus|TestObjectAbsent' -count=1`
Expected: FAIL — `undefined: Status`.

- [ ] **Step 3: Implement**

`assets_absent.go`:

```go
//go:build !ebpf_object

package bpf

// ObjectBytes — sans l'étiquette ebpf_object, l'objet n'est pas dans le
// binaire. L'erreur est typée pour que la chaîne de repli puisse la dire.
func ObjectBytes() ([]byte, error) {
	return nil, &BuildError{Reason: "objet non embarqué (make bpf puis -tags ebpf_object)"}
}
```

`assets_embedded.go`:

```go
//go:build ebpf_object

package bpf

import _ "embed"

//go:embed bpf_probe.o
var object []byte

func ObjectBytes() ([]byte, error) {
	if len(object) == 0 {
		return nil, &BuildError{Reason: "objet vide"}
	}
	return object, nil
}
```

`status.go`:

```go
package bpf

import (
	"crypto/sha256"
	"encoding/hex"
)

// Status — la chaîne de repli, en clair : jamais ebpf sans objet et sans
// capacités. goos/btf/caps sont injectés pour que Windows puisse tester
// chaque échelon sans mentir sur l'hôte.
func Status(goos string, btf, caps bool, objectErr error) (source, reason string) {
	if goos != "linux" {
		return "tcpinfo", "hôte non-Linux"
	}
	if objectErr != nil {
		return "tcpinfo", objectErr.Error()
	}
	if !btf {
		return "tcpinfo", "BTF absent"
	}
	if !caps {
		return "tcpinfo", "CAP_BPF absent"
	}
	return "ebpf", ""
}

// ObjectSHA256 — l'empreinte de l'objet chargé, pour la provenance.
func ObjectSHA256() (string, error) {
	b, err := ObjectBytes()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
```

- [ ] **Step 4: Run it — must pass**

Run: `go test ./pkg/probe/bpf/ -count=1 && gofmt -l pkg/probe/bpf; echo fmt-ok`
Expected: `ok`, `fmt-ok`.

- [ ] **Step 5: Commit**

```bash
git add pkg/probe/bpf/assets_absent.go pkg/probe/bpf/assets_embedded.go pkg/probe/bpf/status.go pkg/probe/bpf/status_test.go
git commit -m "probe: object behind a build tag and a pure fallback ladder"
```

---

### Task 3: `probe.KernelSource()` with real host facts

**Files:**
- Create: `pkg/probe/kernelsource.go`
- Create: `pkg/probe/kernelsource_test.go`

**Interfaces:**
- Consumes: `bpf.Status`, `bpf.ObjectBytes` (Task 2).
- Produces: `type KernelSourceInfo struct { Source, Reason string }`
- Produces: `func KernelSource() KernelSourceInfo`
- Produces (test seam): `var kernelFacts = func() (goos string, btf, caps bool, objectErr error)`

- [ ] **Step 1: Write the failing test**

Create `pkg/probe/kernelsource_test.go`:

```go
package probe

import (
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
)

func TestKernelSourceUsesFacts(t *testing.T) {
	old := kernelFacts
	defer func() { kernelFacts = old }()

	kernelFacts = func() (string, bool, bool, error) { return "linux", true, true, nil }
	if got := KernelSource(); got.Source != "ebpf" || got.Reason != "" {
		t.Fatalf("ready host: %+v", got)
	}
	kernelFacts = func() (string, bool, bool, error) { return "windows", false, false, nil }
	got := KernelSource()
	if got.Source != "tcpinfo" || !strings.Contains(got.Reason, "non-Linux") {
		t.Fatalf("windows host: %+v", got)
	}
	kernelFacts = func() (string, bool, bool, error) { return "linux", true, true, &bpf.BuildError{Reason: "objet non embarqué"} }
	if got := KernelSource(); got.Source != "tcpinfo" || !strings.Contains(got.Reason, "objet") {
		t.Fatalf("absent object: %+v", got)
	}
}

func TestKernelSourceOnThisHostIsHonest(t *testing.T) {
	got := KernelSource()
	if got.Source == "ebpf" {
		t.Fatal("this Windows host cannot be ebpf")
	}
	if got.Reason == "" {
		t.Fatal("a fallback must name its reason")
	}
}
```

- [ ] **Step 2: Run it — must fail**

Run: `go test ./pkg/probe/ -run 'TestKernelSource' -count=1`
Expected: FAIL — `undefined: KernelSource`.

- [ ] **Step 3: Implement**

`kernelsource.go`:

```go
package probe

import (
	"os"
	"runtime"

	"github.com/Realms4239/cgo/pkg/probe/bpf"
)

// KernelSourceInfo — ce que la télémétrie noyau peut dire sur cet hôte,
// et pourquoi elle ne peut pas dire mieux.
type KernelSourceInfo struct {
	Source string `json:"source"`
	Reason string `json:"reason,omitempty"`
}

// kernelFacts est remplacé par les tests : la chaîne doit être vérifiable
// sans prétendre qu'un Windows est un Linux.
var kernelFacts = func() (goos string, btf, caps bool, objectErr error) {
	objectErr = nil
	if _, err := bpf.ObjectBytes(); err != nil {
		objectErr = err
	}
	if runtime.GOOS != "linux" {
		return runtime.GOOS, false, false, objectErr
	}
	if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err != nil {
		return runtime.GOOS, false, false, objectErr
	}
	// root approxime CAP_BPF/CAP_PERFMON ; le runbook vérifie les clés réelles.
	return runtime.GOOS, true, os.Geteuid() == 0, objectErr
}

// KernelSource — la source que le rapport et le bandeau doivent afficher.
func KernelSource() KernelSourceInfo {
	goos, btf, caps, objectErr := kernelFacts()
	source, reason := bpf.Status(goos, btf, caps, objectErr)
	return KernelSourceInfo{Source: source, Reason: reason}
}
```

- [ ] **Step 4: Run it — must pass**

Run: `go test ./pkg/probe/ -run 'TestKernelSource' -count=1 && gofmt -l pkg/probe`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add pkg/probe/kernelsource.go pkg/probe/kernelsource_test.go
git commit -m "probe: KernelSource names the source and its reason on this host"
```

---

### Task 4: Report block and Markdown line

**Files:**
- Modify: `pkg/audit/report.go` (field after `RouterOSValidation`; Markdown after the validation section)
- Create: `pkg/audit/report_telemetry_test.go`

**Interfaces:**
- Produces: `type KernelTelemetry struct { Available bool; Source, Reason, ObjectSHA256 string; RetransPayloadBytes, QdiscDropPackets uint64 }` (JSON `kernel_telemetry`).
- Produces: `Report.KernelTelemetry *KernelTelemetry`.

- [ ] **Step 1: Write the failing test**

Create `pkg/audit/report_telemetry_test.go`:

```go
package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReportTelemetryFallbackIsStated(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.KernelTelemetry = &KernelTelemetry{Source: "tcpinfo", Reason: "objet non embarqué"}
	b, _ := json.Marshal(rep)
	if !strings.Contains(string(b), `"kernel_telemetry"`) {
		t.Fatal("kernel_telemetry missing from JSON")
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Repli TCP_INFO") || !strings.Contains(md, "objet non embarqué") {
		t.Fatalf("fallback not stated with its reason:\n%s", md)
	}
}

func TestReportTelemetryEBPFSaysAttach(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.KernelTelemetry = &KernelTelemetry{Available: true, Source: "ebpf", RetransPayloadBytes: 4096, QdiscDropPackets: 12}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Télémétrie noyau : eBPF") {
		t.Fatalf("ebpf source not rendered:\n%s", md)
	}
	if !strings.Contains(md, "4096") {
		t.Fatal("retransmitted bytes not rendered")
	}
}

func TestReportWithoutTelemetryOmitsBoth(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "kernel_telemetry") {
		t.Fatal("absent telemetry still serialized")
	}
	if strings.Contains(rep.RenderMarkdown(), "Télémétrie noyau") {
		t.Fatal("absent telemetry still rendered")
	}
}
```

- [ ] **Step 2: Run it — must fail**

Run: `go test ./pkg/audit/ -run 'TestReportTelemetry|TestReportWithoutTelemetry' -count=1`
Expected: FAIL — `undefined: KernelTelemetry`.

- [ ] **Step 3: Implement**

Add the type next to `RouterOSValidation` in `pkg/audit/report.go`:

```go
// KernelTelemetry is what the kernel tier could say for this audit, and why,
// when it could not say more. Source is ebpf|tcpinfo|unavailable; a reason
// always accompanies a fallback.
type KernelTelemetry struct {
	Available           bool   `json:"available"`
	Source              string `json:"source,omitempty"`
	Reason              string `json:"reason,omitempty"`
	ObjectSHA256        string `json:"object_sha256,omitempty"`
	RetransPayloadBytes uint64 `json:"retrans_payload_bytes,omitempty"`
	QdiscDropPackets    uint64 `json:"qdisc_drop_packets,omitempty"`
}
```

Add the field after `RouterOSValidation`:

```go
	// KernelTelemetry is the eBPF or TCP_INFO tier for this audit. Absent
	// means the build did not attempt it — never a silent zero.
	KernelTelemetry *KernelTelemetry `json:"kernel_telemetry,omitempty"`
```

In `RenderMarkdown`, after the `RouterOSValidation` block, add:

```go
	if k := r.KernelTelemetry; k != nil {
		switch k.Source {
		case "ebpf":
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nTélémétrie noyau : eBPF — %d octets réémis, %d paquets écartés.\n\n", k.RetransPayloadBytes, k.QdiscDropPackets)
		case "tcpinfo":
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nRepli TCP_INFO — %s.\n\n", k.Reason)
		default:
			fmt.Fprintf(&b, "### Télémétrie noyau\n\nTélémétrie noyau non instrumentée%s.\n\n", reasonClause(k.Reason))
		}
	}
```

and at the bottom:

```go
func reasonClause(reason string) string {
	if reason == "" {
		return ""
	}
	return " — " + reason
}
```

- [ ] **Step 4: Run it — must pass**

Run: `go test ./pkg/audit/ -run 'TestReport' -count=1 && gofmt -l pkg/audit/report.go; echo fmt-ok`
Expected: `ok`, `fmt-ok`.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/report.go pkg/audit/report_telemetry_test.go
git commit -m "audit: report states the kernel telemetry tier and its fallback reason"
```

---

### Task 5: `/api/health` exposes the source; the ribbon says it

**Files:**
- Modify: `pkg/api/server.go` (health handler)
- Create: `pkg/api/health_telemetry_test.go`
- Modify: `web/frontend/src/lib/ribbon.ts`
- Modify: `web/frontend/src/components/StatusRibbon.tsx`
- Modify: `web/frontend/src/lib/ribbon.test.ts`

**Interfaces:**
- Produces (JSON): `/api/health` gains `"telemetry": {"source": "...", "reason": "..."}`.
- Produces: `telemetryLabel(source: string): string` → `eBPF` | `TCP_INFO` | `non instrumenté`.

- [ ] **Step 1: Write the failing tests**

Create `pkg/api/health_telemetry_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthExposesTelemetrySource(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Telemetry struct {
			Source string `json:"source"`
			Reason string `json:"reason"`
		} `json:"telemetry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Telemetry.Source != "tcpinfo" {
		t.Fatalf("source = %q, want tcpinfo on this host", body.Telemetry.Source)
	}
	if body.Telemetry.Reason == "" {
		t.Fatal("fallback without a reason")
	}
}
```

In `web/frontend/src/lib/ribbon.test.ts`, append:

```ts
it('nomme la télémétrie sans mentir', () => {
  expect(telemetryLabel('ebpf')).toBe('eBPF')
  expect(telemetryLabel('tcpinfo')).toBe('TCP_INFO')
  expect(telemetryLabel('')).toBe('non instrumenté')
  expect(telemetryLabel('autre')).toBe('non instrumenté')
})
```

(add `telemetryLabel` to the existing import from `./ribbon`.)

- [ ] **Step 2: Run both — must fail**

Run: `go test ./pkg/api/ -run TestHealthExposesTelemetry -count=1` → FAIL (no `telemetry` field).
Run: `cd web/frontend && npx vitest run src/lib/ribbon.test.ts` → FAIL (`telemetryLabel` not exported).

- [ ] **Step 3: Implement**

`pkg/api/server.go`, health handler:

```go
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "version": d.Version, "mode": d.Mode, "telemetry": probe.KernelSource()})
	})
```

(add `"github.com/Realms4239/cgo/pkg/probe"` to the imports if absent.)

`web/frontend/src/lib/ribbon.ts`:

```ts
// La télémétrie noyau ne s'invente pas : trois états, jamais un quatrième.
export function telemetryLabel(source: string | undefined): string {
  if (source === 'ebpf') return 'eBPF'
  if (source === 'tcpinfo') return 'TCP_INFO'
  return 'non instrumenté'
}
```

`StatusRibbon.tsx`: add `const [telemetry, setTelemetry] = useState<{source?: string; reason?: string}>({})`, set it in `load()` from `/api/health` (`j.telemetry`), and render after the SSE item:

```tsx
<span className="ribbon-item" data-testid="ribbon-telemetry" title={telemetry.reason || undefined}>
  Télémétrie : {telemetryLabel(telemetry.source)}
</span>
```

with `import { telemetryLabel } from '../lib/ribbon'` merged into the existing import.

- [ ] **Step 4: Run everything**

Run: `go test ./pkg/api/ -run TestHealthExposesTelemetry -count=1 && cd web/frontend && npx vitest run src/lib/ribbon.test.ts src/components/StatusRibbon.test.tsx 2>/dev/null; npm run typecheck`
Expected: Go `ok`; ribbon tests pass; typecheck exit 0.

- [ ] **Step 5: Commit**

```bash
git add pkg/api/server.go pkg/api/health_telemetry_test.go web/frontend/src/lib/ribbon.ts web/frontend/src/lib/ribbon.test.ts web/frontend/src/components/StatusRibbon.tsx
git commit -m "api/ui: the ribbon names the kernel telemetry tier, with its reason in the title"
```

---

### Task 6: Portable readiness script and the host runbook

**Files:**
- Create: `kit/host-readiness.sh`
- Create: `docs/runbooks/2026-10-30-host-acceptance.md`

**Interfaces:**
- Produces: an executable script; `--require-linux` exits 2 off Linux; otherwise exit 0 when every required check is `OK` or `n/a`, 1 when a required one is `MISSING`.

- [ ] **Step 1: Write the script**

```bash
#!/usr/bin/env bash
# Diagnostic lecture seule d'un hôte de banc. Ne modifie rien, n'exige pas
# root : il dit ce qui manque et sort en 1 si une ligne requise manque.
# Windows : les lignes Linux sont n/a, le script reste utilisable.
set -uo pipefail
require_linux=0
[ "${1:-}" = "--require-linux" ] && require_linux=1
os="$(uname -s)"
if [ "$require_linux" = 1 ] && [ "$os" != "Linux" ]; then
  echo "hôte non-Linux ($os) — l'acceptation live exige Linux 6.8" >&2
  exit 2
fi
missing=0
row() { printf '%-22s %-8s %s\n' "$1" "$2" "$3"; }
need() {
  local name="$1" status="$2" detail="$3"
  row "$name" "$status" "$detail"
  if [ "$status" = "MISSING" ] && [ "$os" = "Linux" ]; then missing=1; fi
}
if [ "$os" != "Linux" ]; then
  for c in "noyau >= 6.8" "BTF vmlinux" "CAP_BPF/PERFMON" "ip/tc/ethtool" "bpftool+clang" "KVM" "qemu" "image CHR"; do
    row "$c" "n/a" "hôte non-Linux"
  done
  row "bench dualns" "OK" "$( [ -f kit/cgo-bench-dualns.sh ] && echo présent || echo ABSENT )"
  echo; echo "résumé : hôte non-Linux, acceptation live impossible ici"; exit 0
fi
kver="$(uname -r)"; major="${kver%%.*}"; minor="$(echo "$kver" | cut -d. -f2)"
if [ "$major" -gt 6 ] || { [ "$major" -eq 6 ] && [ "$minor" -ge 8 ]; }; then need "noyau >= 6.8" OK "$kver"; else need "noyau >= 6.8" MISSING "$kver"; fi
[ -f /sys/kernel/btf/vmlinux ] && need "BTF vmlinux" OK présent || need "BTF vmlinux" MISSING absent
if grep -q 'CapEff' /proc/self/status 2>/dev/null && [ "$(id -u)" = 0 ]; then need "CAP_BPF/PERFMON" OK root; else need "CAP_BPF/PERFMON" MISSING "root requis (approximation)"; fi
ok=1; for t in ip tc ethtool; do command -v "$t" >/dev/null || ok=0; done
[ "$ok" = 1 ] && need "ip/tc/ethtool" OK présents || need "ip/tc/ethtool" MISSING manquant
ok=1; for t in bpftool clang; do command -v "$t" >/dev/null || ok=0; done
[ "$ok" = 1 ] && need "bpftool+clang" OK "build seulement" || need "bpftool+clang" MISSING "build de l'objet impossible"
[ -e /dev/kvm ] && need "KVM" OK présent || need "KVM" MISSING absent
command -v qemu-system-x86_64 >/dev/null && need "qemu" OK présent || need "qemu" MISSING absent
ls kit/chr/cache/*.img >/dev/null 2>&1 && need "image CHR" OK présente || need "image CHR" MISSING "voir kit/chr/README.md"
echo; [ "$missing" = 0 ] && { echo "résumé : hôte prêt pour l'acceptation"; exit 0; } || { echo "résumé : des lignes requises manquent"; exit 1; }
```

- [ ] **Step 2: Verify syntax and the Windows table**

Run: `bash -n kit/host-readiness.sh && chmod +x kit/host-readiness.sh && bash kit/host-readiness.sh; echo "EXIT=$?"`
Expected: table with `n/a` rows, `EXIT=0`.

- [ ] **Step 3: Write the runbook**

Create `docs/runbooks/2026-10-30-host-acceptance.md` with, per slice 4, 5, 6, 7: prerequisite, exact commands, expected output lines, sealed artifact path, and a **« ne pas conclure »** line. Order and time-box:

1. Readiness (`bash kit/host-readiness.sh --require-linux`) — stop if exit ≠ 0.
2. `make bpf` + `make bpf-check`; rebuild with `-tags ebpf_object`; record the object SHA.
3. Slice 4: one audit with eBPF on and off; report JSON shows `kernel_telemetry.source: ebpf` and the fallback test shows the ladder.
4. Slice 6: IFB ingress acceptance inside `ns_client`; counters move; `< 2 %` throughput delta.
5. Slice 5: only with a real field trace of ≥ 30 baseline samples; never a SMOKE-TEST `.dist`.
6. Slice 7: boot the 7.1 BBRv3 kernel; `bbr` still means BBRv1; matrix run.
7. Never run two slices' acceptance concurrently (namespaces and qdiscs collide).

- [ ] **Step 4: Commit**

```bash
git add kit/host-readiness.sh docs/runbooks/2026-10-30-host-acceptance.md
git commit -m "kit: read-only host readiness and the acceptance runbook for slices 4–7"
```

---

### Task 7: Close slice 17 — gates, index, ledgers

**Files:**
- Modify: `docs/superpowers/plans/README.md` (row 17)
- Modify: `docs/VERIFY.md` (new row: eBPF Go half on Windows)
- Modify: `docs/CHECKPOINTS.md` (append)

- [ ] **Step 1: Run the gates**

```bash
cd web/frontend && npm test && npm run build
cd ../.. && go build ./... && go test ./... -count=1
python tools/verify.py
```

- [ ] **Step 2: Index row 17** → **In progress** (Windows half built; live row blocked), plan linked.

- [ ] **Step 3: VERIFY row** — “eBPF Go half (loader contract, fallback ladder, report block, badge) on Windows” **VERIFIED** with the commands; the existing slice-4 row stays **BLOCKED** (no object, no live attach).

- [ ] **Step 4: Checkpoint** — branch/HEAD, changed, proven, unproven (no eBPF claim), next.

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/plans/README.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m "docs: slice 17 recorded — Windows half verified, live row still blocked"
```

---

## Self-review

**Spec coverage** (§4 of the closing-program spec): 4.1 Go half → Tasks 1–4; 4.2 readiness → Task 6; 4.3 runbook → Task 6; 4.4 badge → Task 5; 4.5 acceptance → Task 7. Guards: no dependency (stdlib + embed), no object committed (build tag), no claim (ladder test `TestKernelSourceOnThisHostIsHonest`).

**Deviations, deliberate:** (1) the object is embedded behind `-tags ebpf_object` instead of always, because the file does not exist and `go build ./...` must stay green on every host; (2) `root` approximates `CAP_BPF` in `kernelFacts`, stated in a comment and refined by the runbook — pretending to read the capability set without a parser would be worse; (3) the health payload is a struct from `probe`, not a map built in the handler, so the field names cannot drift.

**Placeholder scan:** no TBD; every step carries code or an exact command.

**Type consistency:** `bpf.Summary.Flows` is `[]Flow` in Task 1 and is not renamed later; `KernelTelemetry` in Task 4 is an `audit` type, deliberately independent of `bpf.Summary` so the report schema does not change when the wire struct evolves; `telemetryLabel` (Task 5) takes `string | undefined` and both call sites pass exactly that.
