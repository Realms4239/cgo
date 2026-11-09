//go:build ebpf_object

package bpf

import _ "embed"

//go:embed bpf_probe.o
var object []byte

func ObjectBytes() ([]byte, error) {
	if len(object) == 0 {
		return nil, &BuildError{Reason: "objet vide"}
	}
	return object, nil
}
