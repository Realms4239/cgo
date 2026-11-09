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
