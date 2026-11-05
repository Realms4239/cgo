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
