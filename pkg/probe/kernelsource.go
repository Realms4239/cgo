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
