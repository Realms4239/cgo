package capabilities

import (
	"testing"
)

func stubCCs(t *testing.T, list ...string) {
	t.Helper()
	old := readCCList
	readCCList = func() []string { return list }
	t.Cleanup(func() { readCCList = old })
}

func TestDefaultCCsWithoutBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr")
	got := DefaultCCs()
	if len(got) != 2 || got[0] != "cubic" || got[1] != "bbr" {
		t.Fatalf("défaut sans bbr3 : %v", got)
	}
}

func TestDefaultCCsWithBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr", "bbr3")
	got := DefaultCCs()
	if len(got) != 3 || got[2] != "bbr3" {
		t.Fatalf("défaut avec bbr3 : %v", got)
	}
}
