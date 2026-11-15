//go:build !linux || !ebpf_object

// cgo-observe — cette plateforme n'a rien à mesurer (Windows, ou noyau
// sans objet CO-RE). Jamais un binaire muet : le motif se refuse bruyamment.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "cgo-observe: obsevation noyau Linux + objet CO-RE requis (GOOS=linux, -tags ebpf_object, make bpf)")
	os.Exit(2)
}
