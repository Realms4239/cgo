package capabilities

import (
	"testing"

	"github.com/Realms4239/cgo/model"
)

func stubKernelRelease(t *testing.T, release string) {
	t.Helper()
	old := readKernelRelease
	readKernelRelease = func() string { return release }
	t.Cleanup(func() { readKernelRelease = old })
}

func TestCapabilitiesReflectKernelWithoutBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	stubKernelRelease(t, "6.8.0-45-generic")
	got := CapabilitiesNow()
	if got.BBRv3 {
		t.Fatal("bbr3 annoncé sans noyau")
	}
	if len(got.CCs) != 2 || got.CCs[0] != "cubic" || got.CCs[1] != "bbr" {
		t.Fatalf("CCs = %v", got.CCs)
	}
	if got.Kernel != "6.8.0-45-generic" {
		t.Fatalf("kernel = %q", got.Kernel)
	}
	if len(got.Qdiscs) != len(model.AllQdiscs) {
		t.Fatalf("qdiscs = %v", got.Qdiscs)
	}
	assertKnownCCs(t, got.KnownCCs)
}

func TestCapabilitiesReflectKernelWithBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
	stubKernelRelease(t, "7.1.0-cgo-bbr3")
	got := CapabilitiesNow()
	if !got.BBRv3 {
		t.Fatal("bbr3 absent alors que le noyau l'expose")
	}
	if len(got.CCs) != 3 || got.CCs[2] != "bbr3" {
		t.Fatalf("CCs = %v", got.CCs)
	}
}

func assertKnownCCs(t *testing.T, got []string) {
	t.Helper()
	want := map[string]bool{"cubic": true, "bbr": true, "bbr3": true}
	if len(got) != len(want) {
		t.Fatalf("known_ccs = %v", got)
	}
	for _, c := range got {
		if !want[c] {
			t.Fatalf("known_ccs inattendu: %v", got)
		}
	}
}
