//go:build !ebpf_object

package bpf

// ObjectBytes — sans l'étiquette ebpf_object, l'objet n'est pas dans le
// binaire. L'erreur est typée pour que la chaîne de repli puisse la dire.
func ObjectBytes() ([]byte, error) {
	return nil, &BuildError{Reason: "objet non embarqué (make bpf puis -tags ebpf_object)"}
}
