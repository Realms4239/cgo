// CGO six-gate verifier. Standard library only; no network; no shell.
// Port of tools/verify.py — same gates, same names, same flags, same exit
// codes (0/1/2), same fault proofs, same allowances.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	pass        = "pass"
	fail        = "fail"
	skip        = "skip"
	seal        = "docs/frozen-manifests.sha256"
	sizeCeiling = 384 * 1024 // mesuré 2026-10-28 : plus gros fichier suivi 352 240 o
)

// timeoutCode is the sentinel exit code _proc returns when the deadline kill hit.
const timeoutCode = -2

var rootDir = func() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gates: cwd unreadable: %v\n", err)
		os.Exit(2)
	}
	return wd
}()

// line_ceiling_problem / MAX_GATE_LINES from the Python runner were dropped:
// they were Python-file hygiene. This runner is covered by the hygiene gate.

type result struct {
	name   string
	status string
	detail string
}
type gate struct {
	name   string
	proves string
	paths  []string
	run    func() result
	prove  func() (bool, string)
	fast   bool
}

func _proc(args []string, timeout time.Duration, cwd string, env []string) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if cwd == "" {
		cwd = rootDir
	}
	cmd.Dir = cwd
	if env != nil {
		cmd.Env = env
	}
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return normalizeNewlines(outBuf.String()), normalizeNewlines(errBuf.String()), timeoutCode, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return normalizeNewlines(outBuf.String()), normalizeNewlines(errBuf.String()), ee.ExitCode(), nil
	}
	if err != nil {
		return normalizeNewlines(outBuf.String()), normalizeNewlines(errBuf.String()), -1, err
	}
	return normalizeNewlines(outBuf.String()), normalizeNewlines(errBuf.String()), 0, nil
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func _tail(text string, limit int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return strings.Join(lines, "\n")
}

func _glob(root, name string) []string {
	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		return nil
	}
	matches, err := filepath.Glob(filepath.Join(root, "*", name))
	if err != nil {
		return nil
	}
	sort.Strings(matches)
	return matches
}

func verdict(results []result) int {
	for _, r := range results {
		if r.status == fail {
			return 1
		}
	}
	for _, r := range results {
		if r.status == skip {
			return 2
		}
	}
	return 0
}

func forwardSlashRel(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func missingProofs(gates []gate) []string {
	var out []string
	for _, g := range gates {
		if g.prove == nil {
			out = append(out, g.name)
		}
	}
	return out
}

// — frozen : l'audit du corpus est la seule source ------------------------

func sealFindings(root string) []string {
	sealPath, runs := filepath.Join(root, filepath.FromSlash(seal)), filepath.Join(root, "book", "data", "runs")
	data, err := os.ReadFile(sealPath)
	if err != nil {
		return []string{fmt.Sprintf("%s: anchor absent", seal)}
	}
	want := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		digest, rel, found := strings.Cut(line, "  ")
		if !found || !strings.HasPrefix(rel, "book/data/runs/") || !strings.HasSuffix(rel, "/manifest.json") {
			return []string{fmt.Sprintf("%s: malformed line %q", seal, line)}
		}
		want[rel] = digest
	}
	got := map[string]string{}
	for _, p := range _glob(runs, "manifest.json") {
		sum, err := sha256File(p)
		if err != nil {
			continue
		}
		got[forwardSlashRel(root, p)] = sum
	}
	var findings []string
	var changed, missing, extra []string
	for r, w := range want {
		if g, ok := got[r]; ok && g != w {
			changed = append(changed, r)
		} else if !ok {
			missing = append(missing, r)
		}
	}
	for r := range got {
		if _, ok := want[r]; !ok {
			extra = append(extra, r)
		}
	}
	sort.Strings(changed)
	sort.Strings(missing)
	sort.Strings(extra)
	for _, r := range changed {
		findings = append(findings, r+": manifest changed")
	}
	for _, r := range missing {
		findings = append(findings, r+": missing run")
	}
	for _, r := range extra {
		findings = append(findings, r+": run not in the anchor")
	}
	return findings
}

func pythonExecutable() string {
	if p, err := exec.LookPath("python"); err == nil {
		return p
	}
	if p, err := exec.LookPath("python3"); err == nil {
		return p
	}
	return ""
}

// runAudit invokes book/tools/audit-archive.py; returns the combined output
// blob and the exit code (timeoutCode on deadline, -1 on spawn failure).
func runAudit(audit, cwd string, timeout time.Duration) (string, int) {
	py := pythonExecutable()
	if py == "" {
		return "python not found on PATH", -1
	}
	stdout, stderr, code, err := _proc([]string{py, audit}, timeout, cwd, nil)
	if err != nil {
		return err.Error(), -1
	}
	return stdout + stderr, code
}

func gateFrozen() result {
	audit := filepath.Join(rootDir, "book", "tools", "audit-archive.py")
	corpus := filepath.Join(rootDir, "book", "data", "runs")
	if st, err := os.Stat(audit); err != nil || st.IsDir() {
		return result{"frozen", skip, "book/ absent: frozen corpus not checked"}
	}
	if st, err := os.Stat(corpus); err != nil || !st.IsDir() {
		return result{"frozen", skip, "book/ absent: frozen corpus not checked"}
	}
	if pythonExecutable() == "" {
		// Go-only path: the Python runner was itself Python, so this skip is new.
		return result{"frozen", skip, "python not found on PATH: frozen audit not run"}
	}
	blob, code := runAudit(audit, rootDir, 120*time.Second)
	if code == timeoutCode {
		return result{"frozen", fail, "audit timed out after 120 s"}
	}
	if code != 0 {
		return result{"frozen", fail, fmt.Sprintf("audit exit %d:\n%s", code, _tail(blob, 20))}
	}
	var oks []string
	for _, ln := range strings.Split(blob, "\n") {
		if strings.HasPrefix(ln, "OK") {
			oks = append(oks, ln)
		}
	}
	if len(oks) < 2 {
		return result{"frozen", fail, "audit exited 0 without two OK lines:\n" + _tail(blob, 20)}
	}
	findings := sealFindings(rootDir)
	if len(findings) != 0 {
		shown := findings
		if len(shown) > 5 {
			shown = shown[:5]
		}
		return result{"frozen", fail, fmt.Sprintf("%d seal finding(s): %s", len(findings), strings.Join(shown, "; "))}
	}
	count := len(_glob(corpus, "manifest.json"))
	return result{"frozen", pass, fmt.Sprintf("%s | %s: %d manifests", strings.Join(oks, " | "), seal, count)}
}

// book/ is an NTFS junction on this tree: WalkDir/Lstat see an irregular
// non-directory and EvalSymlinks does not resolve junctions, while Python's
// copytree follows them. Recursing with Stat/ReadDir keeps the same view.
func copyTree(src, dst string, ignore map[string]bool) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		info, err := os.Stat(s) // follows junctions and symlinks
		if err != nil {
			return err
		}
		if info.IsDir() {
			if ignore[e.Name()] {
				continue
			}
			if err := copyTree(s, d, ignore); err != nil {
				return err
			}
			continue
		}
		data, err := os.ReadFile(s)
		if err != nil {
			return err
		}
		if err := os.WriteFile(d, data, info.Mode().Perm()); err != nil {
			return err
		}
	}
	return nil
}

func proveFrozen() (bool, string) {
	if st, err := os.Stat(filepath.Join(rootDir, "book")); err != nil || !st.IsDir() {
		return false, "book/ absent: the frozen proof cannot run"
	}
	if pythonExecutable() == "" {
		return false, "python not found on PATH: the frozen proof cannot run"
	}
	tmp, err := os.MkdirTemp("", "cgo-frozen-*")
	if err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	defer os.RemoveAll(tmp)
	dst := filepath.Join(tmp, "book")
	if err := copyTree(filepath.Join(rootDir, "book"), dst, map[string]bool{"node_modules": true, ".git": true, "dist": true}); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	audit := filepath.Join(dst, "tools", "audit-archive.py")
	if _, code := runAudit(audit, tmp, 120*time.Second); code != 0 {
		return false, "control failed: the copied audit does not pass"
	}
	manifests := _glob(filepath.Join(dst, "data", "runs"), "manifest.json")
	sealPath := filepath.Join(tmp, filepath.FromSlash(seal))
	if err := os.MkdirAll(filepath.Dir(sealPath), 0o755); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	var sb strings.Builder
	for _, p := range manifests {
		sum, err := sha256File(p)
		if err != nil {
			return false, fmt.Sprintf("frozen proof error: %v", err)
		}
		sb.WriteString(sum + "  " + forwardSlashRel(tmp, p) + "\n")
	}
	if err := os.WriteFile(sealPath, []byte(sb.String()), 0o644); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	if findings := sealFindings(tmp); len(findings) != 0 {
		return false, "control failed: the seal rejects the copied corpus"
	}
	first := manifests[0]
	original, err := os.ReadFile(first)
	if err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	if err := appendBytes(first, []byte("\n")); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	if findings := sealFindings(tmp); len(findings) == 0 {
		return false, "SURVIVED: the seal accepts a rewritten manifest"
	}
	if _, code := runAudit(audit, tmp, 120*time.Second); code != 0 {
		return false, "fault proof invalid: the audit rejects a whitespace-only manifest change"
	}
	if err := os.WriteFile(first, original, 0o644); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	if findings := sealFindings(tmp); len(findings) != 0 {
		return false, "control failed: the restored corpus no longer matches the seal"
	}
	target := ""
	for _, man := range manifests {
		data, err := os.ReadFile(man)
		if err != nil {
			continue
		}
		var doc struct {
			Files []struct {
				File string `json:"File"`
			} `json:"files"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}
		for _, e := range doc.Files {
			if e.File == "" {
				continue
			}
			candidate := filepath.Join(filepath.Dir(man), filepath.FromSlash(e.File))
			if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
				target = candidate
				break
			}
		}
		if target != "" {
			break
		}
	}
	if target == "" {
		return false, "no manifest file found to mutate"
	}
	if err := appendBytes(target, []byte("\n")); err != nil {
		return false, fmt.Sprintf("frozen proof error: %v", err)
	}
	if _, code := runAudit(audit, tmp, 120*time.Second); code == 0 {
		return false, fmt.Sprintf("SURVIVED: audit exit 0 after mutating %s", forwardSlashRel(tmp, target))
	}
	return true, fmt.Sprintf("control ok; seal caught a rewritten manifest; audit caught %s", forwardSlashRel(tmp, target))
}

func appendBytes(path string, extra []byte) error {
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.Write(extra)
	return err
}

// — provenance : en-têtes CSV et enums de valuation -----------------------

var bases = map[string]bool{"measured": true, "estimated": true}
var sources = map[string]bool{"ebpf": true, "tcpinfo": true, "": true}

func csvHeader(path string) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fh.Close()
	buf := make([]byte, 0, 4096)
	chunk := make([]byte, 4096)
	for {
		n, readErr := fh.Read(chunk)
		for _, b := range chunk[:n] {
			if b == '\n' {
				return strings.TrimRight(string(buf), "\r\n"), nil
			}
			buf = append(buf, b)
			if len(buf) > 1<<20 {
				return strings.TrimRight(string(buf), "\r\n"), nil
			}
		}
		if readErr != nil {
			return strings.TrimRight(string(buf), "\r\n"), nil
		}
	}
}

type problem struct{ path, why string }

func headerProblems(csvPaths []string, reference map[string]bool) []problem {
	var out []problem
	for _, p := range csvPaths {
		header, err := csvHeader(p)
		if err != nil || !reference[header] {
			out = append(out, problem{p, "header not in the frozen eras"})
		}
	}
	return out
}

func pyRepr(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	if !strings.ContainsAny(s, "'\"\\\n\r") {
		return "'" + s + "'"
	}
	return fmt.Sprintf("%q", s)
}

func valuationProblem(v interface{}) string {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return "not a JSON object"
	}
	basis := obj["basis"]
	source := obj["source"]
	if source == nil {
		source = ""
	}
	basisStr, basisOK := basis.(string)
	if !basisOK || !bases[basisStr] {
		return fmt.Sprintf("basis=%s (allowed: measured, estimated)", pyRepr(basis))
	}
	sourceStr, sourceOK := source.(string)
	if !sourceOK || !sources[sourceStr] {
		return fmt.Sprintf("source=%s (allowed: ebpf, tcpinfo, empty)", pyRepr(source))
	}
	if basisStr == "estimated" && sourceStr != "" {
		return fmt.Sprintf("estimated with source=%s", pyRepr(sourceStr))
	}
	return ""
}

func gateProvenance() result {
	frozen := filepath.Join(rootDir, "book", "data", "runs")
	if st, err := os.Stat(frozen); err != nil || !st.IsDir() {
		return result{"provenance", skip, "book/data/runs absent: no reference header"}
	}
	reference := map[string]bool{}
	for _, p := range _glob(frozen, "aqm_eval.csv") {
		if header, err := csvHeader(p); err == nil {
			reference[header] = true
		}
	}
	if len(reference) == 0 {
		return result{"provenance", skip, "book/data/runs empty: no reference header"}
	}
	dataCSVs := _glob(filepath.Join(rootDir, "data", "runs"), "aqm_eval.csv")
	sidecars := append(_glob(filepath.Join(rootDir, "data", "runs"), "valuation.json"), _glob(frozen, "valuation.json")...)
	bad := headerProblems(dataCSVs, reference)
	for _, path := range sidecars {
		why := ""
		data, err := os.ReadFile(path)
		if err != nil {
			why = fmt.Sprintf("unreadable: %v", err)
		} else {
			var doc interface{}
			if err := json.Unmarshal(data, &doc); err != nil {
				why = fmt.Sprintf("unreadable: %v", err)
			} else {
				why = valuationProblem(doc)
			}
		}
		if why != "" {
			bad = append(bad, problem{path, why})
		}
	}
	if len(bad) != 0 {
		shown := bad
		if len(shown) > 20 {
			shown = shown[:20]
		}
		lines := make([]string, 0, len(shown))
		for _, p := range shown {
			lines = append(lines, fmt.Sprintf("%s: %s", p.path, p.why))
		}
		return result{"provenance", fail, fmt.Sprintf("%d finding(s):\n%s", len(bad), strings.Join(lines, "\n"))}
	}
	side := fmt.Sprintf("%d valuation sidecars checked", len(sidecars))
	if len(sidecars) == 0 {
		side = "0 valuation sidecars checked; writer covered by tests"
	}
	return result{"provenance", pass, fmt.Sprintf("%d headers checked against %d frozen eras; %s", len(dataCSVs), len(reference), side)}
}

func proveProvenance() (bool, string) {
	tmp, err := os.MkdirTemp("", "cgo-prov-*")
	if err != nil {
		return false, fmt.Sprintf("provenance proof error: %v", err)
	}
	defer os.RemoveAll(tmp)
	good := "run_id,event_id,profile"
	mustMkdir(filepath.Join(tmp, "ref"))
	mustWrite(filepath.Join(tmp, "ref", "aqm_eval.csv"), good+"\n")
	mustMkdir(filepath.Join(tmp, "x"))
	mustWrite(filepath.Join(tmp, "x", "aqm_eval.csv"), good+",extra\n")
	extraProblems := headerProblems([]string{filepath.Join(tmp, "x", "aqm_eval.csv")}, map[string]bool{good: true})
	checks := []struct {
		ok    bool
		label string
	}{
		{len(extraProblems) != 0, "extra column"},
		{valuationProblem(map[string]interface{}{"basis": "guess"}) != "", "basis=guess"},
		{valuationProblem(map[string]interface{}{"basis": "estimated", "source": "ebpf"}) != "", "estimated+ebpf"},
		{valuationProblem(map[string]interface{}{"basis": "measured", "source": "ebpf"}) == "", "control measured+ebpf"},
		{valuationProblem(map[string]interface{}{"basis": "estimated", "source": ""}) == "", "control estimated empty"},
		{len(headerProblems([]string{filepath.Join(tmp, "ref", "aqm_eval.csv")}, map[string]bool{good: true})) == 0, "control exact header"},
	}
	for _, c := range checks {
		if !c.ok {
			return false, fmt.Sprintf("SURVIVED or control failed: %s", c.label)
		}
	}
	return true, "control ok; extra column and two bad sidecars caught"
}

func mustMkdir(path string) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		panic(err)
	}
}

func mustWrite(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}

// — tests : la suite Go, racine et modules imbriqués ----------------------

func gateTests() result {
	if st, err := os.Stat(filepath.Join(rootDir, "web", "frontend", "dist", "assets")); err != nil || !st.IsDir() {
		return result{"tests", skip, "web/frontend/dist/assets absent: build the frontend first"}
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		return result{"tests", skip, "go not found on PATH"}
	}
	dirs := []string{rootDir, filepath.Join(rootDir, "results"), filepath.Join(rootDir, "api")}
	for _, cwd := range dirs {
		tag := ""
		if cwd != rootDir {
			tag = filepath.Base(cwd) + " "
		}
		for _, spec := range []struct {
			label   string
			args    []string
			timeout time.Duration
		}{
			{"go vet", []string{goBin, "vet", "./..."}, 300 * time.Second},
			{"go test", []string{goBin, "test", "./...", "-count=1"}, 600 * time.Second},
		} {
			stdout, stderr, code, err := _proc(spec.args, spec.timeout, cwd, nil)
			if err != nil {
				return result{"tests", fail, fmt.Sprintf("%s%s could not run: %v", tag, spec.label, err)}
			}
			if code == timeoutCode {
				return result{"tests", fail, fmt.Sprintf("%s%s timed out after %d s", tag, spec.label, int(spec.timeout.Seconds()))}
			}
			if code != 0 {
				return result{"tests", fail, fmt.Sprintf("%s%s exit %d:\n%s", tag, spec.label, code, _tail(stdout+stderr, 20))}
			}
		}
	}
	return result{"tests", pass, "go vet + go test in all modules"}
}

func _module(path, body, name string) string {
	if err := os.MkdirAll(path, 0o755); err != nil {
		panic(err)
	}
	mustWrite(filepath.Join(path, "go.mod"), "module fixture\n\ngo 1.25\n")
	mustWrite(filepath.Join(path, name), body)
	return path
}

func proveTests() (bool, string) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return false, "go not found on PATH"
	}
	tmp, err := os.MkdirTemp("", "cgo-tests-*")
	if err != nil {
		return false, fmt.Sprintf("tests proof error: %v", err)
	}
	defer os.RemoveAll(tmp)
	okDir := _module(filepath.Join(tmp, "ok"), "package fixture\n\nimport \"testing\"\n\nfunc TestOK(t *testing.T) {}\n", "ok_test.go")
	badDir := _module(filepath.Join(tmp, "bad"), "package fixture\n\nimport \"testing\"\n\nfunc TestBad(t *testing.T) { t.Fatal(\"boom\") }\n", "ok_test.go")
	vetDir := _module(filepath.Join(tmp, "vet"), "package fixture\n\nimport \"fmt\"\n\nfunc F() { fmt.Printf(\"%d\", \"x\") }\n", "vet.go")
	_, _, okCode, _ := _proc([]string{goBin, "test", "./..."}, 120*time.Second, okDir, nil)
	_, _, vetCode, _ := _proc([]string{goBin, "vet", "./..."}, 120*time.Second, okDir, nil)
	if okCode != 0 || vetCode != 0 {
		return false, "control failed: passing module not green"
	}
	_, _, badCode, _ := _proc([]string{goBin, "test", "./..."}, 120*time.Second, badDir, nil)
	if badCode == 0 {
		return false, "SURVIVED: failing test module passed"
	}
	_, _, vetBadCode, _ := _proc([]string{goBin, "vet", "./..."}, 120*time.Second, vetDir, nil)
	if vetBadCode == 0 {
		return false, "SURVIVED: vet error not caught"
	}
	return true, "control ok; failing test and vet error caught"
}

// — exit : le contrat 0/1/2 du CLI et du runner ---------------------------

var safeInvocations = []struct {
	args   []string
	reason string
}{
	{nil, "usage, no argument"},
	{[]string{"--nope"}, "unknown command: usage then exit 2"},
	{[]string{"version"}, "prints the version only"},
	{[]string{"shape"}, "exits 2 before any tc call"},
	{[]string{"service"}, "prints install usage, exits 2"},
}

var refusalRx = regexp.MustCompile(`(?i)no such file|not found|unknown|refus|erreur|error`)

func classifyExit(code int, out, errStr string) string {
	if code != 0 && code != 1 && code != 2 {
		return fmt.Sprintf("unlisted exit code %d", code)
	}
	blob := out + "\n" + errStr
	if strings.Contains(blob, "panic:") || strings.Contains(blob, "Traceback (most recent call last)") {
		return "panic or traceback in output"
	}
	if code == 0 && refusalRx.MatchString(blob) {
		return "exit 0 while printing a refusal"
	}
	return ""
}

func verdictContract() string {
	checks := []struct {
		results []result
		want    int
	}{
		{[]result{{name: "x", status: pass}}, 0},
		{[]result{{name: "x", status: pass}, {name: "y", status: skip}}, 2},
		{[]result{{name: "x", status: fail}, {name: "y", status: skip}}, 1},
	}
	for _, c := range checks {
		if got := verdict(c.results); got != c.want {
			return fmt.Sprintf("verdict contract broken: wanted %d, got %d", c.want, got)
		}
	}
	return ""
}

var envDropRx = regexp.MustCompile(`(?i)key|token|secret|password|credential`)

func gateExit() result {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return result{"exit", skip, "go not found on PATH"}
	}
	tmp, err := os.MkdirTemp("", "cgo-exit-*")
	if err != nil {
		return result{"exit", fail, fmt.Sprintf("could not create temp dir: %v", err)}
	}
	defer os.RemoveAll(tmp)
	binary := filepath.Join(tmp, "cgo")
	if os.PathListSeparator == ';' {
		binary += ".exe"
	}
	stdout, stderr, code, err := _proc([]string{goBin, "build", "-o", binary, "./cmd/cgo"}, 300*time.Second, rootDir, nil)
	if err != nil {
		return result{"exit", fail, fmt.Sprintf("go build failed:\n%v", err)}
	}
	if code != 0 {
		return result{"exit", fail, fmt.Sprintf("go build failed:\n%s", _tail(stdout+stderr, 20))}
	}
	env := []string{}
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !envDropRx.MatchString(key) {
			env = append(env, kv)
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var findings []string
	for _, probe := range safeInvocations {
		probe := probe
		wg.Add(1)
		go func() {
			defer wg.Done()
			stdout, stderr, code, err := _proc(append([]string{binary}, probe.args...), 30*time.Second, rootDir, env)
			argsStr := "(" + strings.Join(probe.args, ", ") + ")"
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				findings = append(findings, fmt.Sprintf("%s: could not run: %v", argsStr, err))
				return
			}
			if code == timeoutCode {
				findings = append(findings, fmt.Sprintf("%s: timed out after 30 s", argsStr))
				return
			}
			if why := classifyExit(code, stdout, stderr); why != "" {
				findings = append(findings, fmt.Sprintf("%s: %s (%s)", argsStr, why, probe.reason))
			}
		}()
	}
	wg.Wait()
	sort.Strings(findings)
	if broken := verdictContract(); broken != "" {
		findings = append(findings, broken)
	}
	if len(findings) != 0 {
		shown := findings
		if len(shown) > 20 {
			shown = shown[:20]
		}
		return result{"exit", fail, strings.Join(shown, "\n")}
	}
	return result{"exit", pass, fmt.Sprintf("%d safe invocations; verdict contract ok", len(safeInvocations))}
}

func proveExit() (bool, string) {
	fixtures := []struct {
		code     int
		out      string
		errStr   string
		mustFail bool
	}{
		{3, "", "", true},
		{0, "", "Traceback (most recent call last):", true},
		{0, "no such file", "", true},
		{2, "", "", false},
		{1, "usage", "", false},
	}
	for _, f := range fixtures {
		why := classifyExit(f.code, f.out, f.errStr)
		if f.mustFail && why == "" {
			return false, fmt.Sprintf("SURVIVED: exit %d / %q accepted", f.code, f.errStr+f.out)
		}
		if !f.mustFail && why != "" {
			return false, fmt.Sprintf("control failed: exit %d rejected (%s)", f.code, why)
		}
	}
	if broken := verdictContract(); broken != "" {
		return false, broken
	}
	return true, "control ok; exit 3, panic, refusal and verdict contract caught"
}

// — hygiene : secrets, artefacts, adresses, taille ------------------------

var keyRx = regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`)
var tokenRx = regexp.MustCompile(`ghp_[A-Za-z0-9]{30,}|github[_-]?pat|sk-[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16}|xox[baprs]-`)
var netRx = regexp.MustCompile(`\b100\.(6[4-9]|[7-9][0-9]|1[01][0-9]|12[0-7])\.\d{1,3}\.\d{1,3}\b`)

// Deux adresses nommées volontairement dans AGENTS.md et les registres
// append-only : on ne réécrit pas l'historique. Toute autre échoue.
var allowedAddresses = map[string]bool{"100.114.101.13": true, "100.72.251.32": true}
var allowedArtifacts = map[string]bool{"results/runs-frozen.tar.gz": true} // //go:embed : le build en dépend

// 18 archives suivies avant la règle « backup/ jamais commité » : dette nommée,
// tout nouveau fichier sous backup/ échoue.
var legacyArchives = map[string]bool{}

func init() {
	for _, f := range strings.Fields(`
backup/runs-20260902-001002.tar.gz backup/runs-20260903-232145.tar.gz backup/runs-20260904-004749.tar.gz
backup/runs-20260904-024318.tar.gz backup/runs-20260904-034927.tar.gz backup/runs-20260904-053037.tar.gz
backup/runs-20260904-081048.tar.gz backup/runs-20260904-085058.tar.gz backup/runs-20260904-093525.tar.gz
backup/runs-20260904-101207.tar.gz backup/runs-20260904-104558.tar.gz backup/runs-20260904-110337.tar.gz
backup/runs-20260904-111003.tar.gz backup/runs-20260904-113340.tar.gz backup/runs-20260904-115508.tar.gz
backup/runs-20260904-121855.tar.gz backup/runs-20260904-124132.tar.gz backup/runs-20260904-125750.tar.gz
`) {
		legacyArchives[f] = true
	}
}

var userDocRx = regexp.MustCompile(`^(session-.*|handoff-.*|critical-review-.*)\.md$`)
var unpRx = regexp.MustCompile(`^unp[^/]*/`)
var forbiddenPrefixes = []string{"backup/", "data/", "book/", "testbed-bootstrap/", "node_modules/", ".worktrees/"}

func pathProblem(path string) string {
	if allowedArtifacts[path] || legacyArchives[path] {
		return ""
	}
	if unpRx.MatchString(path) {
		return "forbidden tracked path"
	}
	for _, p := range forbiddenPrefixes {
		if strings.HasPrefix(path, p) {
			return "forbidden tracked path"
		}
	}
	if path == "epbfBBR.md" || userDocRx.MatchString(path) {
		return "forbidden user document"
	}
	for _, suffix := range []string{".docx", ".pdf", ".tar.gz", ".zip", ".exe"} {
		if strings.HasSuffix(path, suffix) {
			return "forbidden artifact"
		}
	}
	return ""
}

func scanText(name string, data []byte) []problem {
	text := strings.ToValidUTF8(string(data), "")
	var findings []problem
	if keyRx.MatchString(text) {
		findings = append(findings, problem{name, "private key block"})
	}
	if tokenRx.MatchString(text) {
		findings = append(findings, problem{name, "token shape"})
	}
	for _, m := range netRx.FindAllString(text, -1) {
		if !allowedAddresses[m] {
			findings = append(findings, problem{name, "tailnet address"})
			break
		}
	}
	return findings
}

func hygieneFindings(paths []string, root string) []problem {
	var findings []problem
	for _, path := range paths {
		if why := pathProblem(path); why != "" {
			findings = append(findings, problem{path, why})
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(path))
		st, err := os.Stat(full)
		if err != nil {
			continue
		}
		size := st.Size()
		if size > int64(sizeCeiling) && !strings.HasPrefix(path, "docs/img/") {
			findings = append(findings, problem{path, fmt.Sprintf("oversized (%d bytes > %d)", size, sizeCeiling)})
			continue
		}
		var data []byte
		if size <= 2_000_000 {
			if d, err := os.ReadFile(full); err == nil {
				data = d
			} else {
				continue
			}
		}
		for _, f := range scanText(path, data) {
			findings = append(findings, f)
		}
	}
	return findings
}

func gateHygiene() result {
	stdout, _, code, err := _proc([]string{"git", "ls-files", "-z"}, 60*time.Second, rootDir, nil)
	if err != nil || code != 0 {
		msg := "exit status"
		if err != nil {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("git ls-files exit %d", code)
		}
		return result{"hygiene", skip, fmt.Sprintf("git unavailable: %s", msg)}
	}
	var paths []string
	for _, p := range strings.Split(stdout, "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	findings := hygieneFindings(paths, rootDir)
	if len(findings) != 0 {
		shown := findings
		if len(shown) > 20 {
			shown = shown[:20]
		}
		lines := make([]string, 0, len(shown))
		for _, f := range shown {
			lines = append(lines, fmt.Sprintf("%s: %s", f.path, f.why))
		}
		return result{"hygiene", fail, fmt.Sprintf("%d finding(s):\n%s", len(findings), strings.Join(lines, "\n"))}
	}
	return result{"hygiene", pass, fmt.Sprintf("%d tracked files clean; legacy allowlist: %d backup archives and results/runs-frozen.tar.gz; %d named tailnet addresses", len(paths), len(legacyArchives), len(allowedAddresses))}
}

func proveHygiene() (bool, string) {
	key := "-----BEGIN " + "RSA PRIVATE KEY-----"
	token := "ghp_" + strings.Repeat("A", 36)
	address := strings.Join([]string{"100", "99", "1", "2"}, ".")
	checks := []struct {
		ok    bool
		label string
	}{
		{len(scanText("k", []byte(key))) != 0, "private key block"},
		{len(scanText("t", []byte(token))) != 0, "token shape"},
		{len(scanText("n", []byte(address))) != 0, "tailnet address"},
		{pathProblem("note.docx") != "", "tracked .docx"},
		{pathProblem("pkg/api/x.go") == "", "control ordinary Go file"},
		{len(scanText("c", []byte("KEY=<your-key>\npassword = \"\"\n10.30.0.2\n"))) == 0, "control placeholders"},
	}
	for _, c := range checks {
		if !c.ok {
			return false, fmt.Sprintf("SURVIVED or control failed: %s", c.label)
		}
	}
	return true, "control ok; key, token, tailnet address and .docx caught"
}

// — links : AGENTS.md, registres et index pointent des fichiers réels -----

var linkFiles = []string{"AGENTS.md", "docs/VERIFY.md", "docs/CHECKPOINTS.md", "docs/ERRORS.md", "docs/superpowers/plans/README.md"}

// Références historiques nommées, jamais un motif général — data/ non suivi sauf data/runs/;
// api/frontend supprimé par c4d8325 (Task 0.3) ; cmd/cgo-gui supprimé par Task 0.6 (2026-11-14),
// y compris des références pointant des fichiers précis de cet arbre supprimé ;
// tools/verify.py et tools/verify_test.py supprimés par le runner Go compilé (Task F1, 2026-11-14).
var historicalAbsent = []string{
	"data/audits", "data/link_audit.csv", "data/profiles.json", "api/frontend",
	"cmd/cgo-gui",
	"cmd/cgo-gui/fygui_linux.go", "cmd/cgo-gui/README-SEPARATE.md", "cmd/cgo-gui/verbs_parity_test.go",
	"tools/verify.py", "tools/verify_test.py",
}

var backtickRx = regexp.MustCompile("`([^`\n]+)`")
var mdBacktickRx = regexp.MustCompile("`([^`\n]+\\.md)`")
var tagRx = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
var planNameRx = regexp.MustCompile(`^2026-\d{2}-\d{2}-.*\.md$`)

func linkCandidates(text, root string) []string {
	var out []string
	for _, token := range backtickRx.FindAllStringSubmatch(text, -1) {
		tok := strings.TrimSpace(token[1])
		if tok == "" || strings.Contains(tok, " ") {
			continue
		}
		if strings.ContainsAny(tok, "*<>{}:") {
			continue
		}
		if strings.HasSuffix(tok, "...") {
			continue
		}
		tag := false
		for _, seg := range strings.Split(tok, "/") {
			if tagRx.MatchString(seg) {
				tag = true
				break
			}
		}
		if tag {
			continue
		}
		if strings.HasPrefix(tok, "http") || strings.HasPrefix(tok, "git ") || strings.HasPrefix(tok, "go ") ||
			strings.HasPrefix(tok, "make ") || strings.HasPrefix(tok, "python ") ||
			strings.HasPrefix(tok, "-") || strings.HasPrefix(tok, ".") {
			continue
		}
		if !strings.Contains(tok, "/") {
			continue
		}
		first := strings.Split(tok, "/")[0]
		r, _ := utf8.DecodeRuneInString(first)
		if unicode.IsLetter(r) {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(first))); err == nil {
				out = append(out, tok)
			}
		}
	}
	return out
}

func missingLinks(root string, docs []string) []problem {
	var out []problem
	for _, name := range docs {
		path := filepath.Join(root, filepath.FromSlash(name))
		if _, err := os.Stat(path); err != nil {
			out = append(out, problem{name, "scanned file is absent"})
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, tok := range linkCandidates(string(data), root) {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(strings.TrimRight(tok, "/")))); err != nil {
				out = append(out, problem{name, tok})
			}
		}
	}
	return out
}

func missingPlanLinks(text, root string) []string {
	dirs := []string{filepath.Join(root, "docs", "superpowers", "specs"), filepath.Join(root, "docs", "superpowers", "plans")}
	var out []string
	for _, m := range mdBacktickRx.FindAllStringSubmatch(text, -1) {
		tok := m[1]
		if strings.Contains(tok, "/") || !planNameRx.MatchString(tok) {
			continue
		}
		found := false
		for _, d := range dirs {
			if _, err := os.Stat(filepath.Join(d, tok)); err == nil {
				found = true
				break
			}
		}
		if !found {
			out = append(out, tok)
		}
	}
	return out
}

func gateLinks() result {
	var absent []string
	for _, name := range linkFiles {
		if _, err := os.Stat(filepath.Join(rootDir, filepath.FromSlash(name))); err != nil {
			absent = append(absent, name)
		}
	}
	if len(absent) != 0 {
		return result{"links", skip, fmt.Sprintf("%s absent: links not checked", absent[0])}
	}
	index := filepath.Join(rootDir, "docs", "superpowers", "plans", "README.md")
	bad := missingLinks(rootDir, linkFiles)
	indexData, err := os.ReadFile(index)
	if err == nil {
		for _, tok := range missingPlanLinks(string(indexData), rootDir) {
			bad = append(bad, problem{filepath.Base(index), tok})
		}
	}
	allowed := []string{}
	for _, item := range bad {
		for _, h := range historicalAbsent {
			if item.why == h {
				allowed = append(allowed, item.why)
				break
			}
		}
	}
	filtered := []problem{}
	for _, item := range bad {
		historical := false
		for _, h := range historicalAbsent {
			if item.why == h {
				historical = true
				break
			}
		}
		if !historical {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) != 0 {
		shown := filtered
		if len(shown) > 20 {
			shown = shown[:20]
		}
		lines := make([]string, 0, len(shown))
		for _, b := range shown {
			lines = append(lines, fmt.Sprintf("%s: %s", b.path, b.why))
		}
		return result{"links", fail, fmt.Sprintf("%d missing path(s):\n%s", len(filtered), strings.Join(lines, "\n"))}
	}
	allowedSet := map[string]bool{}
	for _, a := range allowed {
		allowedSet[a] = true
	}
	return result{"links", pass, fmt.Sprintf("%d documents scanned; plan and spec links resolve; %d named historical absent-path allowance(s)", len(linkFiles), len(allowedSet))}
}

func proveLinks() (bool, string) {
	tmp, err := os.MkdirTemp("", "cgo-links-*")
	if err != nil {
		return false, fmt.Sprintf("links proof error: %v", err)
	}
	defer os.RemoveAll(tmp)
	mustMkdir(filepath.Join(tmp, "docs"))
	mustWrite(filepath.Join(tmp, "AGENTS.md"), "see `docs/gone.md`\n")
	if len(missingLinks(tmp, []string{"AGENTS.md"})) == 0 {
		return false, "SURVIVED: a missing path was not caught"
	}
	mustWrite(filepath.Join(tmp, "docs", "gone.md"), "x\n")
	if len(missingLinks(tmp, []string{"AGENTS.md"})) != 0 {
		return false, "control failed: an existing path was rejected"
	}
	mustMkdir(filepath.Join(tmp, "docs", "superpowers", "plans"))
	mustMkdir(filepath.Join(tmp, "docs", "superpowers", "specs"))
	if len(missingPlanLinks("`2026-01-01-gone-design.md`", tmp)) == 0 {
		return false, "SURVIVED: a missing plan link was not caught"
	}
	mustWrite(filepath.Join(tmp, "docs", "superpowers", "specs", "2026-01-01-live-design.md"), "x\n")
	if len(missingPlanLinks("`2026-01-01-live-design.md`", tmp)) != 0 {
		return false, "control failed: an existing spec link was rejected"
	}
	if len(missingPlanLinks("`make verify` `https://example.org/x`", tmp)) != 0 {
		return false, "control failed: a command or URL was treated as a path"
	}
	return true, "control ok; missing path and missing plan link caught"
}

// — registre --------------------------------------------------------------

var gates = []gate{
	{"frozen", "the frozen corpus audit still attests", []string{"book/", "data/"}, gateFrozen, proveFrozen, true},
	{"provenance", "CSV headers and valuation enums cannot lie", []string{"pkg/", "cmd/"}, gateProvenance, proveProvenance, true},
	{"tests", "go vet and the Go suite pass", nil, gateTests, proveTests, false},
	{"exit", "the CLI keeps the 0/1/2 contract", []string{"pkg/", "cmd/"}, gateExit, proveExit, true},
	{"hygiene", "no tracked secret, artifact or oversized file", []string{"tools/"}, gateHygiene, proveHygiene, true},
	{"links", "AGENTS.md and the ledgers name existing paths", []string{"docs/", "AGENTS.md"}, gateLinks, proveLinks, true},
}

var allNames = func() []string {
	names := make([]string, 0, len(gates))
	for _, g := range gates {
		names = append(names, g.name)
	}
	return names
}()

func selectStaged(changed []string) []string {
	for _, p := range changed {
		if p == "cmd/gates/main.go" {
			return append([]string{}, allNames...)
		}
	}
	names := map[string]bool{"hygiene": true}
	for _, g := range gates {
		if len(g.paths) == 0 {
			continue
		}
		for _, p := range changed {
			for _, prefix := range g.paths {
				if strings.HasPrefix(p, prefix) {
					names[g.name] = true
				}
			}
		}
	}
	var out []string
	for _, n := range allNames {
		if names[n] {
			out = append(out, n)
		}
	}
	return out
}

func stagedPaths() []string {
	output, err := exec.Command("git", "diff", "--cached", "--name-only", "-z").Output()
	if err != nil {
		return []string{}
	}
	var paths []string
	for _, p := range strings.Split(string(output), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var strict, fast, staged, list, selftest bool
	var only stringList
	fs := flag.NewFlagSet("gates", flag.ContinueOnError)
	fs.BoolVar(&strict, "strict", false, "SKIP counts as failure (exit 1)")
	fs.BoolVar(&fast, "fast", false, "all gates except tests")
	fs.BoolVar(&staged, "staged", false, "select gates from staged paths")
	fs.Var(&only, "only", "named gate (repeatable)")
	fs.BoolVar(&list, "list", false, "list gates and their proofs")
	fs.BoolVar(&selftest, "selftest", false, "run every gate fault proof")
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if missing := missingProofs(gates); len(missing) != 0 {
		fmt.Fprintf(os.Stderr, "gate without a fault proof: %s\n", strings.Join(missing, ", "))
		os.Exit(2)
	}
	if list {
		for _, g := range gates {
			tag := "full"
			if g.fast {
				tag = "fast"
			}
			fmt.Printf("%-11s %-4s proof  %s\n", g.name, tag, g.proves)
		}
		os.Exit(0)
	}
	var selected []gate
	switch {
	case len(only) != 0:
		var unknown []string
		for _, n := range only {
			known := false
			for _, name := range allNames {
				if n == name {
					known = true
					break
				}
			}
			if !known {
				unknown = append(unknown, n)
			}
		}
		if len(unknown) != 0 {
			fmt.Fprintf(os.Stderr, "unknown gate: %s (known: %s)\n", strings.Join(unknown, ", "), strings.Join(allNames, ", "))
			os.Exit(2)
		}
		for _, g := range gates {
			for _, n := range only {
				if g.name == n {
					selected = append(selected, g)
				}
			}
		}
	case staged:
		chosen := map[string]bool{}
		for _, n := range selectStaged(stagedPaths()) {
			chosen[n] = true
		}
		for _, g := range gates {
			if chosen[g.name] {
				selected = append(selected, g)
			}
		}
	case fast:
		for _, g := range gates {
			if g.fast {
				selected = append(selected, g)
			}
		}
	default:
		selected = append([]gate{}, gates...)
	}
	if selftest {
		survived := 0
		for _, g := range selected {
			caught, detail := g.prove()
			word := "caught  "
			if !caught {
				word = "SURVIVED"
			}
			fmt.Printf("%s %-11s %s\n", word, g.name, detail)
			if !caught {
				survived++
			}
		}
		if survived != 0 {
			os.Exit(1)
		}
		os.Exit(0)
	}
	results := make([]result, 0, len(selected))
	for _, g := range selected {
		results = append(results, g.run())
	}
	for _, r := range results {
		fmt.Printf("%-4s %-11s %s\n", strings.ToUpper(r.status), r.name, r.detail)
	}
	failed, skipped := 0, 0
	for _, r := range results {
		if r.status == fail {
			failed++
		}
		if r.status == skip {
			skipped++
		}
	}
	fmt.Printf("\n%d passed, %d failed, %d not run\n", len(results)-failed-skipped, failed, skipped)
	if strict && skipped != 0 {
		fmt.Printf("strict: %d gate(s) not run count as failure\n", skipped)
		os.Exit(1)
	}
	os.Exit(verdict(results))
}
