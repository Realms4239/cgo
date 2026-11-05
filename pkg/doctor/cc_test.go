package doctor

import "testing"

func TestCCChecksSeparatesBBRVersions(t *testing.T) {
	got := ccChecksFrom("full", []string{"reno", "cubic", "bbr"}, true)
	if got[0].Name != "bbr" || got[0].Status != "ok" {
		t.Fatalf("bbr v1 : %+v", got[0])
	}
	if got[1].Name != "bbr3" || got[1].Status != "warn" {
		t.Fatalf("bbr3 absent doit rester warn : %+v", got[1])
	}

	got = ccChecksFrom("full", []string{"reno", "cubic", "bbr3"}, true)
	if got[0].Status != "fail" {
		t.Fatalf("bbr v1 absent en mode full doit échouer : %+v", got[0])
	}
	if got[1].Status != "ok" {
		t.Fatalf("bbr3 présent : %+v", got[1])
	}

	got = ccChecksFrom("full", nil, false)
	if got[0].Status != "warn" || got[1].Status != "warn" {
		t.Fatalf("fichier illisible : %+v", got)
	}
}
