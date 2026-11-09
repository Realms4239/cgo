package probe

import (
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
