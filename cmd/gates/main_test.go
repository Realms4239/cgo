package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestVerdictContract(t *testing.T) {
	if got := verdict([]result{{name: "x", status: pass}}); got != 0 {
		t.Errorf("verdict(pass) = %d, want 0", got)
	}
	if got := verdict([]result{{name: "x", status: pass}, {name: "y", status: skip}}); got != 2 {
		t.Errorf("verdict(pass, skip) = %d, want 2", got)
	}
	if got := verdict([]result{{name: "x", status: fail}, {name: "y", status: skip}}); got != 1 {
		t.Errorf("verdict(fail, skip) = %d, want 1", got)
	}
	if broken := verdictContract(); broken != "" {
		t.Errorf("verdictContract() = %q, want ok", broken)
	}
}

func TestStagedSelection(t *testing.T) {
	cases := []struct {
		changed []string
		want    []string
	}{
		{[]string{"book/data/runs/x"}, []string{"frozen", "hygiene"}},
		{[]string{"pkg/api/x.go"}, []string{"provenance", "exit", "hygiene"}},
		{[]string{"docs/x.md"}, []string{"hygiene", "links"}},
		{[]string{"cmd/gates/main.go"}, allNames},
		{[]string{"README.md"}, []string{"hygiene"}},
		{[]string{"cmd/gates/main_test.go"}, []string{"provenance", "exit", "hygiene"}},
	}
	for _, c := range cases {
		got := selectStaged(c.changed)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("selectStaged(%v) = %v, want %v", c.changed, got, c.want)
		}
	}
}

func TestClassifyExit(t *testing.T) {
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
			t.Errorf("classifyExit(%d, %q, %q) = ok, want a finding", f.code, f.out, f.errStr)
		}
		if !f.mustFail && why != "" {
			t.Errorf("classifyExit(%d, %q, %q) = %q, want ok", f.code, f.out, f.errStr, why)
		}
	}
}

func TestLinksAllowance(t *testing.T) {
	tmp := t.TempDir()
	for _, dir := range []string{"docs", "data", "api", "cmd", "tools", "npm", filepath.Join("docs", "superpowers", "plans"), filepath.Join("docs", "superpowers", "specs")} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Historical references found in the real LINK_FILES: the 7 pre-existing
	// ones (VERIFY.md / CHECKPOINTS.md / plans README) plus the deletion refs
	// for cmd/cgo-gui (Task 0.6) and tools/verify.py / tools/verify_test.py
	// (Task F1). All are named in historicalAbsent — none of them points at
	// something that still exists.
	historical := []string{
		"data/audits",
		"data/audits",
		"data/link_audit.csv",
		"data/link_audit.csv",
		"data/profiles.json",
		"data/profiles.json",
		"api/frontend",
		"cmd/cgo-gui",
		"cmd/cgo-gui/fygui_linux.go",
		"cmd/cgo-gui/README-SEPARATE.md",
		"cmd/cgo-gui/verbs_parity_test.go",
		"tools/verify.py",
		"tools/verify_test.py",
	}
	var docText string
	for _, h := range historical {
		docText += "see `" + h + "`\n"
	}
	if err := os.WriteFile(filepath.Join(tmp, "docs", "CHECKPOINTS.md"), []byte(docText), 0o644); err != nil {
		t.Fatal(err)
	}
	// An 11th, bogus reference that must NOT be allowed.
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("gate: `npm/pkg.nope`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"docs/VERIFY.md", "docs/ERRORS.md", "docs/superpowers/plans/README.md"} {
		if err := os.WriteFile(filepath.Join(tmp, filepath.FromSlash(name)), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bad := missingLinks(tmp, linkFiles)
	allowed := 0
	bogus := 0
	for _, item := range bad {
		historicalHit := false
		for _, h := range historicalAbsent {
			if item.why == h {
				historicalHit = true
				break
			}
		}
		if historicalHit {
			allowed++
			continue
		}
		if item.why == "npm/pkg.nope" {
			bogus++
		}
	}
	if allowed != 13 {
		t.Errorf("historical references allowed = %d, want 13 (findings: %v)", allowed, bad)
	}
	if bogus != 1 {
		t.Errorf("bogus reference allowed = %d, want 1 finding kept (findings: %v)", bogus, bad)
	}
}
