package doctor

import (
	"runtime"
	"testing"
)

// TestModeAutoFollowsOS — sur Windows, « auto » observe : le dashboard peut
// mesurer l'affichage sans prétendre exécuter tc/netem sur le poste. Le forçage
// explicite reste disponible pour le banc (--mode full).
func TestModeAutoFollowsOS(t *testing.T) {
	want := "full"
	if runtime.GOOS == "windows" {
		want = "observe"
	}
	for _, override := range []string{"auto", ""} {
		if got := Mode(override); got != want {
			t.Fatalf("Mode(%q) = %q, voulu %q sur %s", override, got, want, runtime.GOOS)
		}
	}
	if got := Mode("full"); got != "full" {
		t.Fatalf("forçage full perdu : %q", got)
	}
	if got := Mode("observe"); got != "observe" {
		t.Fatalf("forçage observe perdu : %q", got)
	}
}
