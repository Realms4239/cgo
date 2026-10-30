# RouterOS v7 compliance via CHR — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove that a RouterOS 7.x instance accepts the sealed report's remediation script and that its `export` matches the intended queue structure, with a sealed artifact, on Windows (TCG) and Linux (KVM + namespaces) from one code path.

**Architecture:** A pure Go package (`pkg/audit`) parses both dialects — our generated script and a RouterOS `export` — into one normalized structure, diffs them into per-field checks, and seals a proof artifact. A small auxiliary command (`cmd/cgo-chr`, not shipped in the product binary) builds the QEMU command line, drives the serial console through an injectable transport, and applies the failure ladder. `kit/chr/setup-mikrotik-chr.sh` is a thin portable wrapper; all judgements live in Go so they are testable without a VM.

**Tech Stack:** Go 1.25 stdlib only; QEMU `qemu-system-x86_64` (human-installed); MikroTik CHR disk image (human-downloaded); React (existing) for the one report line.

## Global Constraints

Copied from `docs/superpowers/specs/2026-10-30-cgo-closing-program-design.md` §2.1:

- No production write, load, root, or capture; production stays read-only and unnamed (pseudonym) until the operator grants otherwise.
- The frozen corpus, `telemetry.csv`, `link_audit.csv`, `aqm_eval.csv` and `data/runs/` are untouched. New provenance is additive only.
- No VM image, ISO, CHR disk, or archive is committed; the CHR cache lives under a gitignored path and its SHA-256 is recorded in the proof artifact.
- No new Go runtime dependency. QEMU, CHR and host checks live in `kit/` and `docs/`; the single static binary stays single.
- No eBPF, BBRv3 or IFB claim before its live row runs; a stub is a stub.
- UI changes additive only: one validation line and one telemetry badge (the badge belongs to slice 17, not this plan).
- French for user-facing text and comments; English for identifiers, this plan, and gate messages.
- Never `git add -A`; stage explicit paths. Never call a result "validated" when RouterOS rejected the script.

**Host prerequisite (measured 2026-10-30):** `qemu-system-x86_64` is absent on DESKTOP-B6K9UCM; no CHR image is cached. Tasks 1–6 and 8 are fully testable without either. Task 7's real boot and Task 9's live acceptance are **human-gated** and must be recorded as such.

---

## File Structure

| Path | Responsibility |
|---|---|
| `pkg/audit/routeros_config.go` | Tokenizer + parser for both dialects; normalized `RouterOSConfig` |
| `pkg/audit/routeros_compare.go` | Expected-vs-found diff into `[]RouterOSCheck` |
| `pkg/audit/routeros_proof.go` | `CHRProof` type + atomic writer (proof, export, transcript, hashes) |
| `pkg/audit/report.go` (modify) | Additive `RouterOSValidation` field + Markdown line |
| `cmd/cgo-chr/main.go` | Flags, report loading, QEMU args, transport, ladder, exit codes |
| `cmd/cgo-chr/serial.go` | Prompt-matching serial driver over an injectable transport |
| `kit/chr/setup-mikrotik-chr.sh` | Portable wrapper: locates the repo, forwards args to `go run ./cmd/cgo-chr` |
| `kit/chr/chr-image.sha256` | Pinned image name + hash; the script refuses anything else |
| `kit/chr/README.md` | Human steps: download CHR, accept licence, place in gitignored cache |
| `kit/chr/export-fixture.txt` | Sanitized export used by tests (no addresses, no secrets) |
| `web/frontend/src/lib/types.ts` (modify) | `RouterOSValidation` type on `AuditReport` |
| `web/frontend/src/components/AuditReportCard.tsx` (modify) | The one validation line |

---

### Task 1: Parse both dialects into one normalized structure

**Files:**
- Create: `pkg/audit/routeros_config.go`
- Create: `pkg/audit/routeros_config_test.go`
- Create: `kit/chr/export-fixture.txt`

**Interfaces:**
- Produces: `type RouterOSObject struct { Section, Name string; Props map[string]string }`
- Produces: `type RouterOSConfig struct { Objects []RouterOSObject }`
- Produces: `func ParseRouterOSConfig(text string) RouterOSConfig`
- Produces: `func (c RouterOSConfig) Get(section, name string) []RouterOSObject`

- [ ] **Step 1: Write the fixture**

Create `kit/chr/export-fixture.txt` with exactly this content (sanitized: no addresses, no secrets):

```text
# 2026-10-30 12:00:00 by RouterOS 7.15.3
/queue type
add cake-ack-filter=filter cake-bandwidth=19000000 cake-diffserv=diffserv4 \
    cake-flowmode=triple-isolate cake-nat=yes cake-rtt=25ms cake-wash=yes \
    kind=cake name=cgo-cake

/queue tree
add comment=cgo-managed max-limit=19000000 name=cgo-wan-out parent=ether1-wan \
    queue=cgo-cake
```

- [ ] **Step 2: Write the failing test**

Create `pkg/audit/routeros_config_test.go`:

```go
package audit

import (
	"os"
	"testing"
)

const scriptCake = `# CGO remediation - RouterOS v7+
/queue tree remove [find where comment="cgo-managed"]
/queue type remove [find where name~"^cgo-"]

/queue type add name=cgo-cake kind=cake \
    cake-bandwidth=19000000 cake-rtt=25ms \
    cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 \
    cake-ack-filter=filter cake-wash=yes

/queue tree add name=cgo-wan-out parent=ether1-wan max-limit=19000000 \
    queue=cgo-cake comment="cgo-managed"
`

func TestParseScriptDialect(t *testing.T) {
	c := ParseRouterOSConfig(scriptCake)
	types := c.Get("queue type", "cgo-cake")
	if len(types) != 1 {
		t.Fatalf("queue type cgo-cake: %d objects, want 1", len(types))
	}
	if got := types[0].Props["kind"]; got != "cake" {
		t.Fatalf("kind = %q, want cake", got)
	}
	if got := types[0].Props["cake-diffserv"]; got != "diffserv4" {
		t.Fatalf("cake-diffserv = %q, want diffserv4", got)
	}
	trees := c.Get("queue tree", "cgo-wan-out")
	if len(trees) != 1 {
		t.Fatalf("queue tree cgo-wan-out: %d objects, want 1", len(trees))
	}
	if got := trees[0].Props["parent"]; got != "ether1-wan" {
		t.Fatalf("parent = %q, want ether1-wan", got)
	}
	if got := trees[0].Props["comment"]; got != "cgo-managed" {
		t.Fatalf("comment = %q, want cgo-managed (quotes stripped)", got)
	}
}

func TestParseExportDialect(t *testing.T) {
	b, err := os.ReadFile("../../kit/chr/export-fixture.txt")
	if err != nil {
		t.Fatal(err)
	}
	c := ParseRouterOSConfig(string(b))
	if len(c.Get("queue type", "cgo-cake")) != 1 {
		t.Fatal("export: queue type cgo-cake not found")
	}
	trees := c.Get("queue tree", "cgo-wan-out")
	if len(trees) != 1 {
		t.Fatalf("export: queue tree cgo-wan-out: %d, want 1", len(trees))
	}
	if got := trees[0].Props["max-limit"]; got != "19000000" {
		t.Fatalf("max-limit = %q, want 19000000", got)
	}
}

func TestParseIgnoresRemovalsAndComments(t *testing.T) {
	c := ParseRouterOSConfig("/queue tree remove [find where comment=\"cgo-managed\"]\n# note\n")
	if len(c.Objects) != 0 {
		t.Fatalf("%d objects, want 0 (removal and comment lines are not objects)", len(c.Objects))
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./pkg/audit/ -run 'TestParse' -count=1`
Expected: FAIL — `undefined: ParseRouterOSConfig`.

- [ ] **Step 4: Implement the parser**

Create `pkg/audit/routeros_config.go`:

```go
package audit

import "strings"

// RouterOSObject is one /queue type or /queue tree entry, normalized so the
// generated script and a RouterOS export compare in the same shape. The
// script names the path on every command; an export prints a /queue type
// header once and then bare `add` lines. Both parse to this type.
type RouterOSObject struct {
	Section string
	Name    string
	Props   map[string]string
}

type RouterOSConfig struct {
	Objects []RouterOSObject
}

// Get returns every object with this section and name; more than one means
// the device stacked duplicates, which is a failure, not a detail.
func (c RouterOSConfig) Get(section, name string) []RouterOSObject {
	var out []RouterOSObject
	for _, o := range c.Objects {
		if o.Section == section && o.Name == name {
			out = append(out, o)
		}
	}
	return out
}

// ParseRouterOSConfig reads both dialects and keeps only queue type and
// queue tree objects; removals, prints and comments are ignored.
func ParseRouterOSConfig(text string) RouterOSConfig {
	var out RouterOSConfig
	section := ""
	for _, line := range logicalLines(text) {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var path []string
		var args []string
		if strings.HasPrefix(line, "/") {
			fields := splitProps(line)
			verb := -1
			for i, f := range fields {
				if isCommandVerb(f) {
					verb = i
					break
				}
			}
			if verb < 0 {
				section = strings.TrimPrefix(strings.Join(fields, " "), "/")
				continue
			}
			path = fields[:verb]
			section = strings.TrimPrefix(strings.Join(path, " "), "/")
			args = fields[verb+1:]
			if fields[verb] != "add" {
				continue
			}
		} else {
			fields := splitProps(line)
			if len(fields) == 0 || fields[0] != "add" {
				continue
			}
			args = fields[1:]
		}
		if section != "queue type" && section != "queue tree" {
			continue
		}
		props := map[string]string{}
		for _, a := range args {
			k, v, ok := strings.Cut(a, "=")
			if !ok {
				continue
			}
			props[k] = strings.Trim(v, `"`)
		}
		name := props["name"]
		if name == "" {
			continue
		}
		delete(props, "name")
		out.Objects = append(out.Objects, RouterOSObject{Section: section, Name: name, Props: props})
	}
	return out
}

// logicalLines joins continuation lines (trailing backslash), as RouterOS
// prints and accepts them, then trims each line.
func logicalLines(text string) []string {
	var out []string
	var cur strings.Builder
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasSuffix(line, `\`) {
			cur.WriteString(strings.TrimSuffix(line, `\`))
			cur.WriteByte(' ')
			continue
		}
		cur.WriteString(line)
		out = append(out, strings.TrimSpace(cur.String()))
		cur.Reset()
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		out = append(out, s)
	}
	return out
}

func isCommandVerb(s string) bool {
	switch s {
	case "add", "set", "remove", "print", "export", "comment", "unset", "enable", "disable":
		return true
	}
	return false
}

// splitProps splits on spaces but keeps quoted values intact and keeps the
// key=value pair together when the value contains spaces.
func splitProps(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./pkg/audit/ -run 'TestParse' -count=1`
Expected: `ok github.com/Realms4239/cgo/pkg/audit`.

- [ ] **Step 6: Commit**

```bash
git add pkg/audit/routeros_config.go pkg/audit/routeros_config_test.go kit/chr/export-fixture.txt
git commit -m "audit: parse RouterOS script and export into one normalized structure"
```

---

### Task 2: Diff expected against found, field by field

**Files:**
- Create: `pkg/audit/routeros_compare.go`
- Create: `pkg/audit/routeros_compare_test.go`

**Interfaces:**
- Consumes: `RouterOSConfig`, `ParseRouterOSConfig`, `Get` (Task 1).
- Produces: `type RouterOSCheck struct { Name, Expected, Found string; Pass bool }` (JSON tags `name`, `expected`, `found`, `pass`).
- Produces: `func CompareRouterOS(expected, found RouterOSConfig) []RouterOSCheck`

- [ ] **Step 1: Write the failing tests**

Create `pkg/audit/routeros_compare_test.go`:

```go
package audit

import (
	"strings"
	"testing"
)

const exportCake = `/queue type
add cake-ack-filter=filter cake-bandwidth=19000000 cake-diffserv=diffserv4 \
    cake-flowmode=triple-isolate cake-nat=yes cake-rtt=25ms cake-wash=yes \
    kind=cake name=cgo-cake
/queue tree
add comment=cgo-managed max-limit=19000000 name=cgo-wan-out parent=ether1-wan \
    queue=cgo-cake
`

func TestCompareMatchPasses(t *testing.T) {
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(exportCake))
	if len(checks) == 0 {
		t.Fatal("no checks produced")
	}
	for _, c := range checks {
		if !c.Pass {
			t.Fatalf("%s failed: expected %q found %q", c.Name, c.Expected, c.Found)
		}
	}
}

func TestCompareMutatedExportFails(t *testing.T) {
	mutated := strings.Replace(exportCake, "cake-ack-filter=filter", "cake-ack-filter=none", 1)
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(mutated))
	var failed []RouterOSCheck
	for _, c := range checks {
		if !c.Pass {
			failed = append(failed, c)
		}
	}
	if len(failed) != 1 {
		t.Fatalf("%d failures, want 1: %+v", len(failed), failed)
	}
	if got := failed[0].Found; got != "none" {
		t.Fatalf("found = %q, want none", got)
	}
}

func TestCompareMissingObjectFails(t *testing.T) {
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig("/queue type\n"))
	var sawMissing bool
	for _, c := range checks {
		if !c.Pass && c.Found == "absent" {
			sawMissing = true
		}
	}
	if !sawMissing {
		t.Fatal("missing queue type reported as present")
	}
}

func TestCompareDuplicateObjectFails(t *testing.T) {
	dup := exportCake + "/queue type\nadd kind=cake name=cgo-cake\n"
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(dup))
	var sawDup bool
	for _, c := range checks {
		if !c.Pass && strings.Contains(c.Name, "duplicate") {
			sawDup = true
		}
	}
	if !sawDup {
		t.Fatal("duplicate object not reported")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/audit/ -run 'TestCompare' -count=1`
Expected: FAIL — `undefined: CompareRouterOS`.

- [ ] **Step 3: Implement the diff**

Create `pkg/audit/routeros_compare.go`:

```go
package audit

import (
	"fmt"
	"sort"
	"strings"
)

// RouterOSCheck is one field-level comparison shown in the proof artifact and
// the report. Expected/Found are always strings so a UI never has to guess.
type RouterOSCheck struct {
	Name     string `json:"name"`
	Expected string `json:"expected"`
	Found    string `json:"found"`
	Pass     bool   `json:"pass"`
}

// CompareRouterOS compares every object the report intended against what the
// device exported. Extra objects the device carries are ignored; duplicate
// copies of an intended name are a failure because the script is meant to be
// idempotent.
func CompareRouterOS(expected, found RouterOSConfig) []RouterOSCheck {
	var checks []RouterOSCheck
	for _, want := range expected.Objects {
		got := found.Get(want.Section, want.Name)
		if len(got) > 1 {
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s duplicate", want.Section, want.Name),
				Expected: "1 object",
				Found:    fmt.Sprintf("%d objects", len(got)),
				Pass:     false,
			})
		}
		if len(got) == 0 {
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s", want.Section, want.Name),
				Expected: fmt.Sprintf("%d properties", len(want.Props)),
				Found:    "absent",
				Pass:     false,
			})
			continue
		}
		keys := make([]string, 0, len(want.Props))
		for k := range want.Props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			exp := want.Props[k]
			f, ok := got[0].Props[k]
			if !ok {
				f = "absent"
			}
			checks = append(checks, RouterOSCheck{
				Name:     fmt.Sprintf("%s %s.%s", want.Section, want.Name, k),
				Expected: exp,
				Found:    f,
				Pass:     ok && normalizeROS(exp) == normalizeROS(f),
			})
		}
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	return checks
}

// normalizeROS absorbs RouterOS formatting noise only: quotes, boolean case
// and a trailing ".0" on integers. It never equates two different values.
func normalizeROS(v string) string {
	v = strings.Trim(strings.TrimSpace(v), `"`)
	v = strings.ToLower(v)
	v = strings.TrimSuffix(v, ".0")
	return v
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/audit/ -run 'TestCompare' -count=1`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/routeros_compare.go pkg/audit/routeros_compare_test.go
git commit -m "audit: diff intended RouterOS structure against the device export"
```

---

### Task 3: Seal the proof artifact

**Files:**
- Create: `pkg/audit/routeros_proof.go`
- Create: `pkg/audit/routeros_proof_test.go`

**Interfaces:**
- Consumes: `RouterOSCheck` (Task 2).
- Produces: `type CHRProof struct` with `Schema, AuditID, ReportSHA256 string; CHR CHRInfo; Level, Accelerator string; Import ImportResult; Checks []RouterOSCheck; ExportSHA256, TranscriptSHA256 string`
- Produces: `func WriteCHRProof(dir string, p CHRProof, export, transcript []byte) (string, error)`
- Produces: `func FileSHA256(b []byte) string`

- [ ] **Step 1: Write the failing tests**

Create `pkg/audit/routeros_proof_test.go`:

```go
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testProof() CHRProof {
	return CHRProof{
		AuditID:      "audit-1792926281",
		ReportSHA256: "aa",
		CHR:          CHRInfo{Version: "7.15.3", Arch: "x86_64", ImageSHA256: "bb"},
		Level:        "windows-tcg",
		Accelerator:  "tcg",
		Import:       ImportResult{Accepted: true},
	}
}

func TestWriteCHRProofSealsHashes(t *testing.T) {
	dir := t.TempDir()
	export := []byte("/queue type\n")
	transcript := []byte("login: admin\n")
	p := testProof()
	path, err := WriteCHRProof(dir, p, export, transcript)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got CHRProof
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Schema != "cgo.chr-proof/1" {
		t.Fatalf("schema = %q", got.Schema)
	}
	if got.ExportSHA256 != FileSHA256(export) {
		t.Fatal("export hash not sealed")
	}
	if got.TranscriptSHA256 != FileSHA256(transcript) {
		t.Fatal("transcript hash not sealed")
	}
	for _, name := range []string{"export.txt", "transcript.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s not written: %v", name, err)
		}
	}
}

func TestWriteCHRProofRejectsUnknownLevel(t *testing.T) {
	if _, err := WriteCHRProof(t.TempDir(), CHRProof{AuditID: "a", Level: "kvm"}, nil, nil); err == nil {
		t.Fatal("unknown level accepted")
	}
}

func TestWriteCHRProofRejectsEmptyAuditID(t *testing.T) {
	if _, err := WriteCHRProof(t.TempDir(), CHRProof{Level: "windows-tcg"}, nil, nil); err == nil {
		t.Fatal("empty audit_id accepted")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./pkg/audit/ -run 'TestWriteCHRProof' -count=1`
Expected: FAIL — `undefined: CHRProof`.

- [ ] **Step 3: Implement the artifact writer**

Create `pkg/audit/routeros_proof.go`:

```go
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CHRInfo struct {
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	ImageSHA256 string `json:"image_sha256"`
}

type ImportResult struct {
	Accepted  bool   `json:"accepted"`
	ErrorLine string `json:"error_line,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// CHRProof is the sealed evidence that a RouterOS instance accepted the
// generated configuration. It is written under data/ (never committed) and
// cited from docs/VERIFY.md by path and hash.
type CHRProof struct {
	Schema           string          `json:"schema"`
	AuditID          string          `json:"audit_id"`
	ReportSHA256     string          `json:"report_sha256"`
	CHR              CHRInfo         `json:"chr"`
	Level            string          `json:"level"`
	Accelerator      string          `json:"accelerator"`
	Import           ImportResult    `json:"import"`
	Checks           []RouterOSCheck `json:"checks"`
	ExportSHA256     string          `json:"export_sha256"`
	TranscriptSHA256 string          `json:"transcript_sha256"`
}

func FileSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// WriteCHRProof writes proof.json, export.txt and transcript.txt atomically
// (temp file + rename) so a killed run never leaves a half proof that reads
// as complete.
func WriteCHRProof(dir string, p CHRProof, export, transcript []byte) (string, error) {
	if p.AuditID == "" {
		return "", fmt.Errorf("chr: audit_id absent")
	}
	if p.Level != "windows-tcg" && p.Level != "linux-namespace" {
		return "", fmt.Errorf("chr: niveau %q inconnu (windows-tcg ou linux-namespace)", p.Level)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p.Schema = "cgo.chr-proof/1"
	p.ExportSHA256 = FileSHA256(export)
	p.TranscriptSHA256 = FileSHA256(transcript)
	if err := writeAtomic(filepath.Join(dir, "export.txt"), export); err != nil {
		return "", err
	}
	if err := writeAtomic(filepath.Join(dir, "transcript.txt"), transcript); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "proof.json")
	if err := writeAtomic(path, b); err != nil {
		return "", err
	}
	return path, nil
}

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/audit/ -run 'TestWriteCHRProof' -count=1`
Expected: PASS (3 tests), then `gofmt -l pkg/audit/routeros_proof.go` prints nothing.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/routeros_proof.go pkg/audit/routeros_proof_test.go
git commit -m "audit: seal CHR proof (schema, hashes, atomic write)"
```

---

### Task 4: Report field and Markdown line

**Files:**
- Modify: `pkg/audit/report.go` (add field to `Report` at line 56; add line in `RenderMarkdown` near the RouterOS block at line ~539)
- Create: `pkg/audit/report_routeros_test.go`

**Interfaces:**
- Consumes: `RouterOSCheck` (Task 2).
- Produces: `type RouterOSValidation struct { Version, Level, ExportSHA256, ProofPath string; Checks []RouterOSCheck }`
- Produces: `Report.RouterOSValidation *RouterOSValidation` (JSON `routeros_validation,omitempty`).

- [ ] **Step 1: Write the failing test**

Create `pkg/audit/report_routeros_test.go`:

```go
package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReportCarriesRouterOSValidation(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.RouterOSValidation = &RouterOSValidation{
		Version: "7.15.3", Level: "windows-tcg", ExportSHA256: "abcdef123456",
		Checks: []RouterOSCheck{{Name: "queue type cgo-cake.kind", Expected: "cake", Found: "cake", Pass: true}},
	}
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"routeros_validation"`) {
		t.Fatal("routeros_validation missing from JSON")
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Configuration RouterOS validée (7.15.3, CHR)") {
		t.Fatalf("markdown lacks the validation line:\n%s", md)
	}
	if strings.Contains(md, "appliquée") {
		t.Fatal("markdown claims the config was applied")
	}
}

func TestReportWithoutValidationOmitsBoth(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	b, _ := json.Marshal(rep)
	if strings.Contains(string(b), "routeros_validation") {
		t.Fatal("absent validation still serialized")
	}
	if strings.Contains(rep.RenderMarkdown(), "Configuration RouterOS") {
		t.Fatal("absent validation still rendered")
	}
}

func TestReportFailedValidationSaysNonValidée(t *testing.T) {
	rep := Report{AuditID: "audit-1"}
	rep.RouterOSValidation = &RouterOSValidation{
		Version: "7.15.3", Level: "windows-tcg", ExportSHA256: "abcdef123456",
		Checks: []RouterOSCheck{{Name: "queue type cgo-cake.kind", Expected: "cake", Found: "none", Pass: false}},
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Configuration RouterOS non validée") {
		t.Fatal("failed validation not stated")
	}
	if strings.Contains(md, "validée (7.15.3") {
		t.Fatal("failed validation rendered as a pass")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./pkg/audit/ -run 'TestReport.*RouterOS' -count=1`
Expected: FAIL — `undefined: RouterOSValidation`.

- [ ] **Step 3: Implement the field and the Markdown line**

In `pkg/audit/report.go`, add after the `CAPSLA` field (line 56):

```go
	// RouterOSValidation is the CHR proof for this report's remediation
	// script. Absent means "not validated", never "validated by default".
	RouterOSValidation *RouterOSValidation `json:"routeros_validation,omitempty"`
```

And the type, near `Remediation`:

```go
type RouterOSValidation struct {
	Version      string          `json:"version"`
	Level        string          `json:"level"`
	ExportSHA256 string          `json:"export_sha256"`
	ProofPath    string          `json:"proof_path,omitempty"`
	Checks       []RouterOSCheck `json:"checks"`
}
```

In `RenderMarkdown`, immediately after the RouterOS ingress block (`if r.Remediation.RouterOSIngress != "" { … }`), add:

```go
		if v := r.RouterOSValidation; v != nil {
			ok := true
			for _, c := range v.Checks {
				if !c.Pass {
					ok = false
				}
			}
			level := "CHR"
			if v.Level == "linux-namespace" {
				level = "CHR + banc"
			}
			if ok {
				fmt.Fprintf(&b, "### Validation RouterOS\n\nConfiguration RouterOS validée (%s, %s) · export %s\n\n", v.Version, level, short8(v.ExportSHA256))
			} else {
				fmt.Fprintf(&b, "### Validation RouterOS\n\nConfiguration RouterOS non validée (%s, %s) — écarts dans le rapport.\n\n", v.Version, level)
			}
		}
```

Add the helper at the bottom of the file:

```go
func short8(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./pkg/audit/ -run 'TestReport' -count=1`
Expected: PASS, including the pre-existing report tests.

- [ ] **Step 5: Commit**

```bash
git add pkg/audit/report.go pkg/audit/report_routeros_test.go
git commit -m "audit: report carries the CHR validation and states it without claiming application"
```

---

### Task 5: QEMU command line and `--dry-run`

**Files:**
- Create: `cmd/cgo-chr/main.go`
- Create: `cmd/cgo-chr/main_test.go`

**Interfaces:**
- Produces: `type RunConfig struct { QEMU, Image, Level string; Timeout time.Duration }`
- Produces: `func BuildQEMUArgs(c RunConfig) []string`
- Produces: `func main()` with flags `--report`, `--image`, `--level`, `--timeout`, `--dry-run`, `--out`, `--qemu`, `--export-file`.

- [ ] **Step 1: Write the failing test**

Create `cmd/cgo-chr/main_test.go`:

```go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestBuildQEMUArgsTCG(t *testing.T) {
	args := strings.Join(BuildQEMUArgs(RunConfig{QEMU: "qemu-system-x86_64", Image: "chr.img", Level: "windows-tcg", Timeout: time.Minute}), " ")
	for _, want := range []string{"-accel tcg", "-serial stdio", "-display none", "file=chr.img,format=raw,if=virtio"} {
		if !strings.Contains(args, want) {
			t.Fatalf("args lack %q: %s", want, args)
		}
	}
	if strings.Contains(args, "-accel kvm") {
		t.Fatalf("TCG level asked for KVM: %s", args)
	}
}

func TestBuildQEMUArgsKVM(t *testing.T) {
	args := strings.Join(BuildQEMUArgs(RunConfig{QEMU: "qemu-system-x86_64", Image: "chr.img", Level: "linux-namespace", Timeout: time.Minute}), " ")
	if !strings.Contains(args, "-accel kvm") {
		t.Fatalf("linux-namespace did not ask for KVM: %s", args)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/cgo-chr/ -run 'TestBuildQEMUArgs' -count=1`
Expected: FAIL — `undefined: BuildQEMUArgs`.

- [ ] **Step 3: Implement the flags and the builder**

Create `cmd/cgo-chr/main.go`:

```go
// cgo-chr is an auxiliary tool, not part of the shipped binary: it boots the
// MikroTik CHR image in QEMU, feeds it the sealed report's RouterOS script,
// and seals proof that RouterOS 7 accepted it. Every judgement lives in
// pkg/audit; this file only orchestrates.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type RunConfig struct {
	QEMU    string
	Image   string
	Level   string
	Timeout time.Duration
}

// BuildQEMUArgs returns the exact command line for both backends. The
// accelerator is part of the evidence: the proof records which one ran.
func BuildQEMUArgs(c RunConfig) []string {
	accel := "tcg"
	if c.Level == "linux-namespace" {
		accel = "kvm"
	}
	return []string{
		c.QEMU,
		"-m", "256",
		"-smp", "1",
		"-drive", "file=" + c.Image + ",format=raw,if=virtio",
		"-netdev", "user,id=n0",
		"-device", "virtio-net-pci,netdev=n0",
		"-serial", "stdio",
		"-display", "none",
		"-accel", accel,
	}
}

func main() {
	report := flag.String("report", "", "chemin du rapport scellé (data/audits/<id>/audit_report.json)")
	image := flag.String("image", "", "image disque CHR (voir kit/chr/README.md)")
	level := flag.String("level", "windows-tcg", "windows-tcg | linux-namespace")
	timeout := flag.Duration("timeout", 180*time.Second, "délai de démarrage de CHR")
	out := flag.String("out", "", "dossier du sceau (défaut data/chr/<audit_id>)")
	qemu := flag.String("qemu", "qemu-system-x86_64", "binaire QEMU")
	exportFile := flag.String("export-file", "", "mode rejeu: valider un export existant sans QEMU")
	dry := flag.Bool("dry-run", false, "afficher la ligne de commande QEMU et sortir")
	flag.Parse()

	cfg := RunConfig{QEMU: *qemu, Image: *image, Level: *level, Timeout: *timeout}
	if *dry {
		fmt.Println(joinCommand(BuildQEMUArgs(cfg)))
		return
	}
	if *report == "" {
		fatal(2, "chr: --report requis")
	}
	if *exportFile == "" && *image == "" {
		fatal(2, "chr: --image requis (téléchargez l'image CHR, voir kit/chr/README.md)")
	}
	code := run(cfg, *report, *exportFile, *out)
	os.Exit(code)
}

func joinCommand(args []string) string {
	s := ""
	for i, a := range args {
		if i > 0 {
			s += " "
		}
		s += a
	}
	return s
}

func fatal(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
```

Create `cmd/cgo-chr/run.go` (stub returning "not implemented" until Task 6):

```go
package main

import "fmt"

func run(cfg RunConfig, reportPath, exportFile, outDir string) int {
	fmt.Fprintln(os.Stderr, "chr: orchestrateur non câblé (Task 6)")
	return 1
}
```

(Add `import "os"` in `run.go`.)

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./cmd/cgo-chr/ -run 'TestBuildQEMUArgs' -count=1`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add cmd/cgo-chr/main.go cmd/cgo-chr/main_test.go cmd/cgo-chr/run.go
git commit -m "chr: QEMU command line for TCG and KVM, with a dry run"
```

---

### Task 6: Serial orchestration and the failure ladder

**Files:**
- Create: `cmd/cgo-chr/serial.go`
- Create: `cmd/cgo-chr/serial_test.go`
- Modify: `cmd/cgo-chr/run.go`

**Interfaces:**
- Produces: `type Transport interface { io.ReadWriteCloser }`
- Produces: `type Deps struct { Start func(RunConfig) (Transport, error) }`
- Produces: `func drive(t Transport, script string, timeout time.Duration) (transcript, export string, importErr string, err error)`
- Produces: `func runWith(deps Deps, cfg RunConfig, reportPath, exportFile, outDir string) int`

- [ ] **Step 1: Write the failing tests**

Create `cmd/cgo-chr/serial_test.go`:

```go
package main

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeCHR replays a RouterOS console: banner, login, password, prompt, then
// answers each line. It is the seam that keeps CI VM-free.
// Read blocks until a Write has produced output: returning (0, nil) would trip
// bufio's empty-read limit and fail the test for the wrong reason.
type fakeCHR struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	closed bool
}

func newFakeCHR() *fakeCHR {
	f := &fakeCHR{}
	f.buf.WriteString("MikroTik RouterOS 7.15.3 (c) 1999-2025\r\n\r\nlogin: ")
	return f
}

func (f *fakeCHR) Read(p []byte) (int, error) {
	for {
		f.mu.Lock()
		if f.buf.Len() > 0 {
			n, _ := f.buf.Read(p)
			f.mu.Unlock()
			return n, nil
		}
		if f.closed {
			f.mu.Unlock()
			return 0, io.EOF
		}
		f.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func (f *fakeCHR) emit(s string) {
	f.mu.Lock()
	f.buf.WriteString(s)
	f.mu.Unlock()
}

func (f *fakeCHR) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	switch {
	case line == "admin":
		f.emit("\r\npassword: ")
	case line == "":
		f.emit("\r\n[admin@MikroTik] > ")
	case strings.HasPrefix(line, "/export"):
		f.emit("\r\n/queue type\r\nadd cake-ack-filter=filter cake-bandwidth=19000000 cake-diffserv=diffserv4 cake-flowmode=triple-isolate cake-nat=yes cake-rtt=25ms cake-wash=yes kind=cake name=cgo-cake\r\n/queue tree\r\nadd comment=cgo-managed max-limit=19000000 name=cgo-wan-out parent=ether1-wan queue=cgo-cake\r\n[admin@MikroTik] > ")
	case line == "REJECT":
		f.emit("\r\nfailure: bad command name\r\n[admin@MikroTik] > ")
	default:
		f.emit("\r\n[admin@MikroTik] > ")
	}
	return len(p), nil
}

func (f *fakeCHR) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func TestDriveHappyPath(t *testing.T) {
	tr := newFakeCHR()
	transcript, export, importErr, err := drive(tr, "/queue type add name=cgo-cake kind=cake\n", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if importErr != "" {
		t.Fatalf("import rejected: %s", importErr)
	}
	if !strings.Contains(transcript, "7.15.3") {
		t.Fatal("transcript lacks the version banner")
	}
	if !strings.Contains(export, "name=cgo-cake") {
		t.Fatalf("export not captured: %q", export)
	}
}

func TestDriveRejection(t *testing.T) {
	tr := newFakeCHR()
	_, _, importErr, err := drive(tr, "REJECT\n", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(importErr, "failure: bad command name") {
		t.Fatalf("RouterOS error not stored verbatim: %q", importErr)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./cmd/cgo-chr/ -run 'TestDrive' -count=1`
Expected: FAIL — `undefined: drive`.

- [ ] **Step 3: Implement the serial driver**

Create `cmd/cgo-chr/serial.go`:

```go
package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	loginRe  = regexp.MustCompile(`(?i)login:\s*$`)
	passRe   = regexp.MustCompile(`(?i)password:\s*$`)
	promptRe = regexp.MustCompile(`\[admin@[^\]]*\]\s*>\s*$`)
	errorRe  = regexp.MustCompile(`(?i)(failure:|syntax error|bad command name|expected end of command|input does not match)`)
)

var errTimeout = errors.New("chr: délai dépassé")

// drive logs in, sends the script line by line, then captures the export.
// The transcript keeps every byte the device sent, including refusals.
func drive(tr io.ReadWriteCloser, script string, timeout time.Duration) (transcript, export, importErr string, err error) {
	r := bufio.NewReader(tr)
	var all strings.Builder
	read := func(re *regexp.Regexp) (string, error) {
		chunk, err := expect(r, re, timeout)
		all.WriteString(chunk)
		return chunk, err
	}
	if _, err := read(loginRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("admin\n")); err != nil {
		return all.String(), "", "", err
	}
	if _, err := read(passRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("\n")); err != nil {
		return all.String(), "", "", err
	}
	if _, err := read(promptRe); err != nil {
		return all.String(), "", "", err
	}
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, err := tr.Write([]byte(line + "\n")); err != nil {
			return all.String(), "", "", err
		}
		chunk, err := read(promptRe)
		if err != nil {
			return all.String(), "", "", err
		}
		if m := errorRe.FindString(chunk); m != "" {
			return all.String(), "", strings.TrimSpace(lastErrorLine(chunk)), nil
		}
	}
	if _, err := tr.Write([]byte("/export terse\n")); err != nil {
		return all.String(), "", "", err
	}
	chunk, err := read(promptRe)
	if err != nil {
		return all.String(), "", "", err
	}
	return all.String(), cleanupExport(chunk), "", nil
}

func expect(r *bufio.Reader, re *regexp.Regexp, d time.Duration) (string, error) {
	type res struct {
		s   string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		var b strings.Builder
		for {
			c, err := r.ReadByte()
			if err != nil {
				ch <- res{b.String(), err}
				return
			}
			b.WriteByte(c)
			if re.MatchString(b.String()) {
				ch <- res{b.String(), nil}
				return
			}
		}
	}()
	select {
	case v := <-ch:
		return v.s, v.err
	case <-time.After(d):
		return "", errTimeout
	}
}

// cleanupExport drops the echoed command and the trailing prompt, keeping
// only what RouterOS printed for /export.
func cleanupExport(chunk string) string {
	if i := strings.Index(chunk, "\n"); i >= 0 {
		chunk = chunk[i+1:]
	}
	if i := strings.LastIndex(chunk, "[admin@"); i >= 0 {
		chunk = chunk[:i]
	}
	return strings.TrimSpace(strings.ReplaceAll(chunk, "\r", ""))
}

func lastErrorLine(chunk string) string {
	lines := strings.Split(strings.ReplaceAll(chunk, "\r", ""), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if errorRe.MatchString(lines[i]) {
			return lines[i]
		}
	}
	return ""
}
```

- [ ] **Step 4: Wire `run` and its ladder**

Replace `cmd/cgo-chr/run.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/remediate"
)

type reportFile struct {
	AuditID     string                `json:"audit_id"`
	Remediation *remediate.Remediation `json:"remediation"`
}

func run(cfg RunConfig, reportPath, exportFile, outDir string) int {
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chr: rapport illisible: %v\n", err)
		return 2
	}
	var rep reportFile
	if err := json.Unmarshal(raw, &rep); err != nil || rep.AuditID == "" {
		fmt.Fprintln(os.Stderr, "chr: rapport sans audit_id")
		return 2
	}
	if rep.Remediation == nil || (rep.Remediation.RouterOS == "" && rep.Remediation.RouterOSIngress == "") {
		fmt.Fprintln(os.Stderr, "chr: le rapport ne porte aucun script RouterOS")
		return 2
	}
	if outDir == "" {
		outDir = filepath.Join("data", "chr", rep.AuditID)
	}
	script := rep.Remediation.RouterOS + "\n" + rep.Remediation.RouterOSIngress
	expected := audit.ParseRouterOSConfig(script)

	var export, transcript string
	importErr := ""
	if exportFile != "" {
		b, err := os.ReadFile(exportFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chr: export illisible: %v\n", err)
			return 2
		}
		export = string(b)
		transcript = "mode rejeu — aucun transcript de console\n"
	} else {
		if _, err := os.Stat(cfg.Image); err != nil {
			fmt.Fprintf(os.Stderr, "chr: image CHR absente (%s) — voir kit/chr/README.md\n", cfg.Image)
			return 2
		}
		tr, err := startQEMU(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chr: %v\n", err)
			return 2
		}
		defer tr.Close()
		var exp string
		transcript, exp, importErr, err = drive(tr, script, cfg.Timeout)
		export = exp
		if err != nil {
			p := audit.CHRProof{AuditID: rep.AuditID, ReportSHA256: audit.FileSHA256(raw),
				CHR: audit.CHRInfo{ImageSHA256: imageHash(cfg.Image)}, Level: cfg.Level,
				Accelerator: accelerator(cfg.Level),
				Import:      audit.ImportResult{Accepted: false, Reason: err.Error()}}
			if _, werr := audit.WriteCHRProof(outDir, p, []byte(export), []byte(transcript)); werr != nil {
				fmt.Fprintf(os.Stderr, "chr: sceau non écrit: %v\n", werr)
			}
			fmt.Fprintf(os.Stderr, "chr: %v\n", err)
			return 1
		}
	}

	found := audit.ParseRouterOSConfig(export)
	checks := audit.CompareRouterOS(expected, found)
	accepted := importErr == ""
	allPass := accepted
	for _, c := range checks {
		if !c.Pass {
			allPass = false
		}
	}
	p := audit.CHRProof{
		AuditID: rep.AuditID, ReportSHA256: audit.FileSHA256(raw),
		CHR: audit.CHRInfo{ImageSHA256: imageHash(cfg.Image)}, Level: cfg.Level,
		Accelerator: accelerator(cfg.Level),
		Import:      audit.ImportResult{Accepted: accepted, ErrorLine: importErr},
		Checks:      checks,
	}
	path, err := audit.WriteCHRProof(outDir, p, []byte(export), []byte(transcript))
	if err != nil {
		fmt.Fprintf(os.Stderr, "chr: sceau non écrit: %v\n", err)
		return 1
	}
	fmt.Printf("chr: preuve %s · %d/%d vérifications · niveau %s\n", path, passed(checks), len(checks), cfg.Level)
	if !allPass {
		return 1
	}
	return 0
}

func startQEMU(cfg RunConfig) (Transport, error) {
	if _, err := exec.LookPath(cfg.QEMU); err != nil {
		return nil, fmt.Errorf("qemu introuvable (%s) — installez QEMU, voir kit/chr/README.md", cfg.QEMU)
	}
	args := BuildQEMUArgs(cfg)[1:] // le binaire est le premier élément
	cmd := exec.Command(cfg.QEMU, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &qemuTransport{cmd: cmd, in: stdin, out: stdout}, nil
}

func accelerator(level string) string {
	if level == "linux-namespace" {
		return "kvm"
	}
	return "tcg"
}

func imageHash(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return audit.FileSHA256(b)
}

func passed(checks []audit.RouterOSCheck) int {
	n := 0
	for _, c := range checks {
		if c.Pass {
			n++
		}
	}
	return n
}
```

Create `cmd/cgo-chr/transport.go`:

```go
package main

import (
	"io"
	"os/exec"
)

type Transport interface {
	io.ReadWriteCloser
}

type qemuTransport struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.Reader
}

func (q *qemuTransport) Read(p []byte) (int, error)  { return q.out.Read(p) }
func (q *qemuTransport) Write(p []byte) (int, error) { return q.in.Write(p) }
func (q *qemuTransport) Close() error {
	_ = q.in.Close()
	return q.cmd.Process.Kill()
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./cmd/cgo-chr/ -count=1`
Expected: PASS. Then `go vet ./cmd/cgo-chr/` and `gofmt -l cmd/cgo-chr` print nothing. Remove the `var _ = time.Second` line and the `time` import if gofmt/vet flags them.

- [ ] **Step 6: Commit**

```bash
git add cmd/cgo-chr/serial.go cmd/cgo-chr/serial_test.go cmd/cgo-chr/run.go cmd/cgo-chr/transport.go
git commit -m "chr: drive the RouterOS console, seal the proof, and fail loudly"
```

---

### Task 7: Portable wrapper, image pin and human README

**Files:**
- Create: `kit/chr/setup-mikrotik-chr.sh`
- Create: `kit/chr/chr-image.sha256`
- Create: `kit/chr/README.md`

**Interfaces:**
- Consumes: `cmd/cgo-chr` flags (Task 5–6).
- Produces: an executable wrapper accepting the same flags.

- [ ] **Step 1: Write the wrapper**

Create `kit/chr/setup-mikrotik-chr.sh`:

```bash
#!/usr/bin/env bash
# Wrapper portable (Git Bash et Linux) autour de cmd/cgo-chr. Il ne juge rien :
# il localise le dépôt et transmet les arguments. Toute la logique est en Go
# pour rester testable sans VM.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
cd "$root"
exec go run ./cmd/cgo-chr "$@"
```

- [ ] **Step 2: Pin the image and write the human steps**

Create `kit/chr/chr-image.sha256`:

```text
# Image CHR attendue par le wrapper. Le fichier vit sous kit/chr/cache/
# (gitignoré). Remplacer la ligne ci-dessous par le nom et le SHA-256 de
# l'image réellement téléchargée, puis laisser le script refuser toute autre.
# <nom-image>.img  <sha256>
```

Create `kit/chr/README.md` (French, operator-facing):

```markdown
# Validation RouterOS 7 par CHR (QEMU)

Prouve qu'un RouterOS 7.x accepte le script de remédiation du rapport scellé et
que son `export` correspond à la structure voulue. Deux niveaux :
`windows-tcg` (import + export + diff) et `linux-namespace` (idem, plus les
interfaces CHR attachées aux namespaces du banc).

## Prérequis humains (non automatisables)

1. Installer QEMU (`qemu-system-x86_64`). Absent sur DESKTOP-B6K9UCM au
   2026-10-30 ; l'installation est une décision de l'humain.
2. Télécharger l'image CHR sur mikrotik.com et accepter la licence (CHR
   gratuit à débit limité ou licence d'évaluation). Ne jamais commiter l'image.
3. Placer l'image dans `kit/chr/cache/`, renseigner `chr-image.sha256`.

## Usage

```bash
bash kit/chr/setup-mikrotik-chr.sh --dry-run \
  --image kit/chr/cache/chr-7.15.3.img --level windows-tcg

bash kit/chr/setup-mikrotik-chr.sh \
  --report data/audits/<audit_id>/audit_report.json \
  --image kit/chr/cache/chr-7.15.3.img --level windows-tcg
```

Le sceau est écrit dans `data/chr/<audit_id>/` (`proof.json`, `export.txt`,
`transcript.txt`) ; jamais commité. Un refus de RouterOS est conservé mot pour
mot et la preuve dit `accepted: false` — jamais « validée ».

## Ce que cette preuve ne dit pas

Elle ne prouve ni débit, ni latence, ni qu'une production a changé. Le niveau
`linux-namespace` ne s'obtient que si l'attachement aux namespaces réussit.
```

- [ ] **Step 3: Verify shell syntax and executability**

Run: `bash -n kit/chr/setup-mikrotik-chr.sh && chmod +x kit/chr/setup-mikrotik-chr.sh && echo OK`
Expected: `OK`.

- [ ] **Step 4: Add the cache ignore rule**

Append to `.gitignore` (single explicit line, next to the other data ignores):

```text
kit/chr/cache/
```

Run: `git check-ignore -v kit/chr/cache/x.img`
Expected: prints the new rule and exit 0.

- [ ] **Step 5: Commit**

```bash
git add kit/chr/setup-mikrotik-chr.sh kit/chr/chr-image.sha256 kit/chr/README.md .gitignore
git commit -m "kit: CHR wrapper, pinned image and the human steps it cannot automate"
```

---

### Task 8: One validation line in the report card

**Files:**
- Modify: `web/frontend/src/lib/types.ts` (near `AuditReport`, line 140)
- Modify: `web/frontend/src/components/AuditReportCard.tsx` (inside the remediation block, after the ingress buttons around line 264-280)
- Create: `web/frontend/src/components/AuditReportCard.routeros.test.tsx`

**Interfaces:**
- Consumes: `AuditReport.routeros_validation` from the API (Task 4).
- Produces: `data-testid="ros-validation"` line.

- [ ] **Step 1: Write the failing test**

Create `web/frontend/src/components/AuditReportCard.routeros.test.tsx`:

```tsx
import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { AuditReportCard } from './AuditReportCard'
import type { AuditReport } from '../lib/types'

const base = {
  audit_id: 'audit-1', timestamp: '2026-10-30T12:00:00Z', site: 'Département X',
  link_type: '5g', provider: 'Yas',
  measured: {}, phases: {}, unavailable: [], bloat_grade: 'D',
  bloat_verdict: 'queue sous charge', provenance: { source_sha256: 'abc' },
  remediation: { qdisc: 'cake', reason: '', linux_tc: [], routeros: '/queue type add name=cgo-cake kind=cake' },
  cap_sla: {},
} as unknown as AuditReport

describe('AuditReportCard — validation RouterOS', () => {
  it('dit la validation avec la version, le niveau et l’empreinte', () => {
    const html = renderToStaticMarkup(
      <AuditReportCard
        report={{ ...base, routeros_validation: { version: '7.15.3', level: 'windows-tcg', export_sha256: 'abcdef1234567890', checks: [{ name: 'queue type cgo-cake.kind', expected: 'cake', found: 'cake', pass: true }] } } as AuditReport}
        onCopy={() => {}}
      />,
    )
    expect(html).toContain('Configuration RouterOS validée')
    expect(html).toContain('7.15.3')
    expect(html).toContain('abcdef12')
    expect(html).not.toContain('appliquée')
  })

  it('un écart dit non validée', () => {
    const html = renderToStaticMarkup(
      <AuditReportCard
        report={{ ...base, routeros_validation: { version: '7.15.3', level: 'windows-tcg', export_sha256: 'abcdef1234567890', checks: [{ name: 'x', expected: 'cake', found: 'none', pass: false }] } } as AuditReport}
        onCopy={() => {}}
      />,
    )
    expect(html).toContain('non validée')
  })

  it('un script sans preuve dit qu’aucune validation n’a eu lieu', () => {
    const html = renderToStaticMarkup(<AuditReportCard report={base} onCopy={() => {}} />)
    expect(html).toContain('non validée')
    expect(html).toContain('aucune preuve CHR')
  })
})
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd web/frontend && npx vitest run src/components/AuditReportCard.routeros.test.tsx`
Expected: FAIL — no `ros-validation` text in the markup.

- [ ] **Step 3: Add the type and the line**

In `web/frontend/src/lib/types.ts`, above `export interface AuditReport`:

```ts
export interface RouterOSCheck {
  name: string
  expected: string
  found: string
  pass: boolean
}

export interface RouterOSValidation {
  version?: string
  level?: string
  export_sha256?: string
  proof_path?: string
  checks?: RouterOSCheck[]
}
```

and on `AuditReport` add `routeros_validation?: RouterOSValidation`.

In `AuditReportCard.tsx`, inside the remediation block after the ingress buttons, add:

```tsx
{report.remediation?.routeros || report.remediation?.routeros_ingress ? (
  <p className="mono" data-testid="ros-validation" style={{ fontSize: 'var(--fs-tag)', color: 'var(--text-muted)', marginTop: 4 }}>
    {(() => {
      const v = report.routeros_validation
      if (!v) return 'Configuration RouterOS non validée — aucune preuve CHR'
      const ok = (v.checks ?? []).every(c => c.pass)
      const level = v.level === 'linux-namespace' ? 'CHR + banc' : 'CHR'
      const h8 = (v.export_sha256 ?? '').slice(0, 8)
      return ok
        ? `Configuration RouterOS validée (${v.version ?? '7.x'}, ${level}) · export ${h8}`
        : 'Configuration RouterOS non validée — écarts dans le rapport'
    })()}
  </p>
) : null}
```

- [ ] **Step 4: Run the tests and the typecheck**

Run: `cd web/frontend && npx vitest run src/components/AuditReportCard.routeros.test.tsx && npm run typecheck`
Expected: 3 tests pass, typecheck exit 0.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/lib/types.ts web/frontend/src/components/AuditReportCard.tsx web/frontend/src/components/AuditReportCard.routeros.test.tsx
git commit -m "ui: le rapport dit si RouterOS a accepté la configuration, et à quel niveau"
```

---

### Task 9: Close the slice — index, ledger, gates

**Files:**
- Modify: `docs/superpowers/plans/README.md` (row 16: link the plan, status **In progress** until the Windows boot is recorded)
- Modify: `docs/VERIFY.md` (new row: CHR validation, status BLOCKED with the QEMU/image condition)
- Modify: `docs/CHECKPOINTS.md` (append one entry)

**Interfaces:** none.

- [ ] **Step 1: Run the full gates**

```bash
cd web/frontend && npm test && npm run build
cd ../.. && go build ./... && go test ./... -count=1
python tools/verify.py
```

Expected: frontend suite green (46 files + the new one), build exit 0;
Go suite exit 0; `6 passed, 0 failed, 0 not run`.

- [ ] **Step 2: Update the index row 16**

Set the plan cell to `2026-10-30-cgo-chr-router-validation.md`, status
**In progress** (Windows proof pending QEMU), and the next action to
`install QEMU + CHR image, then record the windows-tcg proof`.

- [ ] **Step 3: Add the VERIFY row**

Append:

```markdown
| RouterOS 7 configuration compliance via CHR | BLOCKED | Needs `qemu-system-x86_64` (absent on DESKTOP-B6K9UCM 2026-10-30) and a downloaded CHR image under `kit/chr/cache/`; the Go parser, diff, proof writer and UI line are testable without either. A `windows-tcg` proof upgrades only this row | 2026-10-30 | DESKTOP-B6K9UCM (Windows 10, MINGW64) |
```

- [ ] **Step 4: Append the checkpoint**

Use the standard entry format: branch/HEAD, changed, proven (the commands and
their exact results), unproven (no CHR boot, no Linux attach, no production),
next (install QEMU, run the wrapper, record the proof).

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/plans/README.md docs/VERIFY.md docs/CHECKPOINTS.md
git commit -m "docs: slice 16 planned and recorded blocked on the CHR host prerequisite"
```

---

## Self-review

**Spec coverage**
- §3.1 proof definition and evidence levels → Tasks 3, 6, 7 (level recorded, never upgraded silently).
- §3.2 components → File Structure table, all paths created or modified.
- §3.3 sealed-report input and expected structure → Tasks 4, 6 (`run` reads the report, never regenerates).
- §3.4 flow → Task 6 (`drive` + `run`), Task 7 (wrapper).
- §3.5 artifact schema → Task 3 (`CHRProof`, `cgo.chr-proof/1`, hashes, atomic).
- §3.6 failure ladder → Task 6 exit codes and reasons; Task 5 exits 2 for absent QEMU/image; boot timeout writes a proof with `accepted:false`.
- §3.7 report + UI line → Tasks 4 and 8, with the "never appliquée" test.
- §3.8 acceptance → Task 9 gates; Windows boot and Linux attach are human-gated and recorded as such.
- §3.9 tests → fixture parser, mutated-export fault proof, fake console (injected transport instead of a fake binary — stronger and OS-agnostic); no VM in CI.
- §2.1 guards → Global Constraints; `.gitignore` for the cache; no dependency added.

**Deviations from the spec, deliberate and recorded here**
1. Orchestration lives in `cmd/cgo-chr` (Go) and `kit/chr/setup-mikrotik-chr.sh` is a wrapper. The spec's own rule — "every judgement is Go code with tests" — is what motivates it; a bash console driver cannot be tested without a shell harness.
2. The console seam is an injected `Transport` rather than a fake `qemu` binary on `PATH`. Same guarantee (no VM in CI), portable on Windows.
3. `ImportResult.ErrorLine` carries the first RouterOS refusal verbatim; the proof keeps the whole transcript file beside it.

**Placeholder scan**: no TBD/TODO; every code step carries code; every command has an expected result.

**Type consistency**: `RouterOSObject`/`RouterOSConfig`/`Get` (T1) are used unchanged in T2; `RouterOSCheck` (T2) is used by T3, T4 and the TS mirror in T8 (`checks[].{name,expected,found,pass}`); `CHRProof`/`CHRInfo`/`ImportResult` (T3) are constructed in T6 with the same field names; `RunConfig`/`BuildQEMUArgs` (T5) are consumed by `startQEMU` in T6.
