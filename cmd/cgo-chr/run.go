package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Realms4239/cgo/pkg/audit"
)

type reportFile struct {
	AuditID     string             `json:"audit_id"`
	Remediation *audit.Remediation `json:"remediation"`
}

func run(cfg RunConfig, reportPath, exportFile, outDir string) int {
	if ok, reason := levelAllowed(cfg.Level, runtime.GOOS); !ok {
		fmt.Fprintln(os.Stderr, "chr: "+reason)
		return 2
	}
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

// levelAllowed — le niveau linux-namespace exige Linux ; l'affirmer plus tôt
// évite un échec QEMU obscur et une preuve étiquetée à tort.
func levelAllowed(level, goos string) (bool, string) {
	if level == "linux-namespace" && goos != "linux" {
		return false, "niveau linux-namespace sur " + goos + " — l'attachement aux namespaces exige Linux (utilisez windows-tcg ici)"
	}
	return true, ""
}
