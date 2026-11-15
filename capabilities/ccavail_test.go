package capabilities

import (
	"testing"
)

func stubCCs(t *testing.T, list ...string) {
	t.Helper()
	oldCC, oldMarker, oldVersion := readCCList, readMarker, readVersion
	readCCList = func() []string { return list }
	// Noyau « stock » côté provenance : marqueur sans commit épinglé, version
	// sans suffixe g. Sans ça, les tests de défaut dépendraient de l'hôte —
	// sur une machine provisionnée, bbr3 entrerait dans DefaultCCs.
	readMarker = func() (string, error) { return "{}", nil }
	readVersion = func() (string, error) { return "Linux version 0.0.0-test", nil }
	t.Cleanup(func() { readCCList, readMarker, readVersion = oldCC, oldMarker, oldVersion })
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

func TestDefaultCCsWithProvisionedBBRv3(t *testing.T) {
	stubCCs(t, "cubic", "bbr") // liste sans « bbr3 » : la provenance décide
	defer StubProvisioned(pinnedMarker, "Linux version 6.13.7-g90210de4b779 #1 SMP")()
	got := DefaultCCs()
	if len(got) != 3 || got[2] != "bbr3" {
		t.Fatalf("défaut sur hôte provisionné : %v (bbr3 attendu via le pin)", got)
	}
}
