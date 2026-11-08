package bpf

// UnavailableError — limitation de l'hôte (BTF, capacités, OS, vérificateur).
type UnavailableError struct{ Reason string }

func (e *UnavailableError) Error() string { return "ebpf indisponible : " + e.Reason }

// BuildError — défaut de l'objet embarqué (absent, vide, hash, ELF).
type BuildError struct{ Reason string }

func (e *BuildError) Error() string { return "ebpf : objet invalide : " + e.Reason }
