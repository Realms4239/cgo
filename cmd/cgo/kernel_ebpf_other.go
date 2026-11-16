//go:build !linux || !ebpf_object

package main

import "github.com/Realms4239/cgo/audit"

// attachKernelProbe est un no-op hors du couple (linux, ebpf_object) :
// sans l'objet embarqué, il n'y a rien à attacher, et l'audit le dit en
// tcpinfo — pas en erreur.
func attachKernelProbe(d *audit.Deps) {}
