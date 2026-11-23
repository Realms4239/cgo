package probe

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/probe/bpf"
)

func TestKernelSourceUsesFacts(t *testing.T) {
	old := kernelFacts
	defer func() { kernelFacts = old }()

	kernelFacts = func() (string, bool, bool, error) { return "linux", true, true, nil }
	if got := KernelSource(); got.Source != "ebpf" || got.Reason != "" {
		t.Fatalf("ready host: %+v", got)
	}
	kernelFacts = func() (string, bool, bool, error) { return "windows", false, false, nil }
	got := KernelSource()
	if got.Source != "tcpinfo" || !strings.Contains(got.Reason, "non-Linux") {
		t.Fatalf("windows host: %+v", got)
	}
	kernelFacts = func() (string, bool, bool, error) {
		return "linux", true, true, &bpf.BuildError{Reason: "objet non embarqué"}
	}
	if got := KernelSource(); got.Source != "tcpinfo" || !strings.Contains(got.Reason, "objet") {
		t.Fatalf("absent object: %+v", got)
	}
}

func TestKernelSourceOnThisHostIsHonest(t *testing.T) {
	got := KernelSource()
	if got.Source == "ebpf" {
		t.Fatal("this Windows host cannot be ebpf")
	}
	if got.Reason == "" {
		t.Fatal("a fallback must name its reason")
	}
}

// Oracle d'observation — une source de repli ne porte JAMAIS de hash d'objet
// ni de compteur de réservation (pas de placement, pas de zéro fabriqué), et
// une source ebpf porte le hash RÉEL fourni par le chargeur.
func TestKernelSourceObservationTruth(t *testing.T) {
	oldFacts, oldSHA := kernelFacts, objectSHA256
	defer func() { kernelFacts, objectSHA256 = oldFacts, oldSHA }()

	// tcpinfo : hostile — aucun hash ne doit apparaître sur une source de repli.
	kernelFacts = func() (string, bool, bool, error) { return "windows", false, false, nil }
	fallback := KernelSource()
	if fallback.Source != "tcpinfo" {
		t.Fatalf("source tcpinfo attendue, obtenue %q", fallback.Source)
	}
	if fallback.ObjectSHA256 != "" {
		t.Fatalf("hash d'objet fabriqué sur une source tcpinfo: %q", fallback.ObjectSHA256)
	}
	if fallback.ReserveFailures != 0 {
		t.Fatalf("compteur de réservation non nul sur une source tcpinfo: %d", fallback.ReserveFailures)
	}

	// ebpf : le hash calculé à partir de l'objet est recopié, pas inventé.
	blob := []byte("ebpf-object-under-test")
	sum := sha256.Sum256(blob)
	want := hex.EncodeToString(sum[:])
	objectSHA256 = func() (string, error) { return want, nil }
	kernelFacts = func() (string, bool, bool, error) { return "linux", true, true, nil }
	ebpf := KernelSource()
	if ebpf.Source != "ebpf" {
		t.Fatalf("source ebpf attendue, obtenue %q", ebpf.Source)
	}
	if ebpf.ObjectSHA256 != want {
		t.Fatalf("hash d'objet=%q, attendu le hash réel %q", ebpf.ObjectSHA256, want)
	}

	// Hostile — le chargeur échoue : jamais de hash de placement, même en ebpf.
	objectSHA256 = func() (string, error) { return "", errors.New("objet absent") }
	if got := KernelSource(); got.ObjectSHA256 != "" {
		t.Fatalf("hash fabriqué malgré l'échec du chargeur: %q", got.ObjectSHA256)
	}
}
