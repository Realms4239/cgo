package probe

import (
	"os"
	"runtime"

	"github.com/Realms4239/cgo/probe/bpf"
)

// objectSHA256 est injectable pour vérifier le chemin ebpf sans objet vendéré
// ni clang au runtime : en production c'est le chargeur réel qui fournit le hash.
var objectSHA256 = bpf.ObjectSHA256

// KernelSourceInfo — ce que la télémétrie noyau peut dire sur cet hôte,
// et pourquoi elle ne peut pas dire mieux. Le hash d'objet n'est porté QUE par
// une source ebpf : jamais de placement sur une source de repli.
type KernelSourceInfo struct {
	Source          string `json:"source"`
	Reason          string `json:"reason,omitempty"`
	ObjectSHA256    string `json:"object_sha256,omitempty"`
	ReserveFailures uint64 `json:"reserve_failures,omitempty"`
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
	info := KernelSourceInfo{Source: source, Reason: reason}
	if source == "ebpf" {
		if sha, err := objectSHA256(); err == nil {
			info.ObjectSHA256 = sha
		}
	}
	return info
}
