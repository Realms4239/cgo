package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/audit"
)

const cakeScript = `/queue type add name=cgo-cake kind=cake \
    cake-bandwidth=19000000 cake-rtt=25ms \
    cake-flowmode=triple-isolate cake-nat=yes cake-diffserv=diffserv4 \
    cake-ack-filter=filter cake-wash=yes
/queue tree add name=cgo-wan-out parent=ether1-wan max-limit=19000000 \
    queue=cgo-cake comment="cgo-managed"
`

const ingressScript = `/queue type add name=cgoingress-cake kind=cake \
    cake-bandwidth=19000000 cake-rtt=25ms \
    cake-flowmode=dual-dsthost cake-nat=yes cake-diffserv=diffserv4 \
    cake-wash=yes
/queue tree add name=cgoingress-lan-in parent=ether2-lan max-limit=19000000 \
    queue=cgoingress-cake comment="cgo-managed-ingress"
`

func writeReport(t *testing.T, dir, routeros string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"audit_id":    "audit-test",
		"remediation": map[string]any{"qdisc": "cake", "linux_tc": []string{}, "routeros": routeros},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "report.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeIngressReport — rapport dont la remédiation ne porte que le script
// d'ingress RouterOS ; le script WAN est absent.
func writeIngressReport(t *testing.T, dir, routerosIngress string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"audit_id": "audit-test",
		"remediation": map[string]any{
			"qdisc":            "cake",
			"linux_tc":         []string{},
			"routeros_ingress": routerosIngress,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "report.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestRunReplayWritesProof — le chemin rejeu (sans QEMU) écrit un sceau
// complet : niveau, empreinte, douze vérifications toutes passées.
func TestRunReplayWritesProof(t *testing.T) {
	dir := t.TempDir()
	report := writeReport(t, dir, cakeScript)
	out := filepath.Join(dir, "proof")
	code := run(RunConfig{Level: "windows-tcg"}, report, "../../kit/chr/export-fixture.txt", out)
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	b, err := os.ReadFile(filepath.Join(out, "proof.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p audit.CHRProof
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Schema != "cgo.chr-proof/1" || p.Level != "windows-tcg" {
		t.Fatalf("schema/level: %q/%q", p.Schema, p.Level)
	}
	if !p.Import.Accepted {
		t.Fatal("replay not accepted")
	}
	if len(p.Checks) != 12 {
		t.Fatalf("%d checks, want 12", len(p.Checks))
	}
	for _, c := range p.Checks {
		if !c.Pass {
			t.Fatalf("%s failed in replay: expected %q found %q", c.Name, c.Expected, c.Found)
		}
	}
	if p.ExportSHA256 == "" || p.TranscriptSHA256 == "" {
		t.Fatal("hashes not sealed")
	}
	for _, name := range []string{"export.txt", "transcript.txt"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
	}
}

func TestRunReplayMutatedExportFails(t *testing.T) {
	dir := t.TempDir()
	report := writeReport(t, dir, cakeScript)
	raw, err := os.ReadFile("../../kit/chr/export-fixture.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutated := filepath.Join(dir, "export.txt")
	if err := os.WriteFile(mutated, []byte(strings.Replace(string(raw), "cake-ack-filter=filter", "cake-ack-filter=none", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "proof")
	if code := run(RunConfig{Level: "windows-tcg"}, report, mutated, out); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	// l'échec doit nommer la propriété fautive, pas seulement sortir en 1
	raw, rerr := os.ReadFile(filepath.Join(out, "proof.json"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	var p audit.CHRProof
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
}

// TestRunReplayValidatesIngressScript — le rejeu couvre aussi le script
// d'ingress (queue tree LAN), absent de la fixture WAN d'origine.
func TestRunReplayValidatesIngressScript(t *testing.T) {
	dir := t.TempDir()
	report := writeIngressReport(t, dir, ingressScript)
	out := filepath.Join(dir, "proof")
	if code := run(RunConfig{Level: "windows-tcg"}, report, "../../kit/chr/export-ingress-fixture.txt", out); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	b, err := os.ReadFile(filepath.Join(out, "proof.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p audit.CHRProof
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Checks) == 0 {
		t.Fatal("aucune vérification dans la preuve")
	}
	parentSeen := false
	for _, c := range p.Checks {
		if !c.Pass {
			t.Fatalf("%s failed in replay: expected %q found %q", c.Name, c.Expected, c.Found)
		}
		if strings.Contains(c.Name, "queue tree cgoingress-lan-in.parent") {
			parentSeen = true
		}
	}
	if !parentSeen {
		t.Fatal("vérification queue tree cgoingress-lan-in.parent absente")
	}
}

// TestRunReplayRejectsWrongIngressParent — un parent LAN différent de celui du
// rapport doit faire échouer le rejeu (sortie 1).
func TestRunReplayRejectsWrongIngressParent(t *testing.T) {
	dir := t.TempDir()
	report := writeIngressReport(t, dir, ingressScript)
	raw, err := os.ReadFile("../../kit/chr/export-ingress-fixture.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutated := filepath.Join(dir, "export.txt")
	if err := os.WriteFile(mutated, []byte(strings.Replace(string(raw), "parent=ether2-lan", "parent=ether3-lan", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "proof")
	if code := run(RunConfig{Level: "windows-tcg"}, report, mutated, out); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	raw, rerr := os.ReadFile(filepath.Join(out, "proof.json"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	var p audit.CHRProof
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	var named bool
	for _, c := range p.Checks {
		if !c.Pass && strings.Contains(c.Name, "cgoingress-lan-in.parent") {
			named = true
		}
	}
	if !named {
		t.Fatalf("mutated parent not named among failures: %+v", p.Checks)
	}
}

func TestRunRejectsReportWithoutScript(t *testing.T) {
	dir := t.TempDir()
	report := writeReport(t, dir, "")
	if code := run(RunConfig{Level: "windows-tcg"}, report, "", filepath.Join(dir, "proof")); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestRunRejectsUnknownLevel(t *testing.T) {
	dir := t.TempDir()
	report := writeReport(t, dir, cakeScript)
	if code := run(RunConfig{Level: "kvm"}, report, "../../kit/chr/export-fixture.txt", filepath.Join(dir, "proof")); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}
