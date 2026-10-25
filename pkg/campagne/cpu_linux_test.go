//go:build linux

package campagne

import "testing"

func TestParseProcStatLine(t *testing.T) {
	idle, total, ok := parseProcStatLine("cpu  10 20 30 40 50 60 70 80 90 100")
	if !ok {
		t.Fatal("ligne /proc/stat refusée")
	}
	if idle != 90 { // idle + iowait
		t.Fatalf("idle=%d, attendu 90", idle)
	}
	if total != 360 { // user..steal (guest/nice exclus)
		t.Fatalf("total=%d, attendu 360", total)
	}
	if _, _, ok := parseProcStatLine("intr 1 2 3"); ok {
		t.Fatal("ligne non-cpu acceptée")
	}
	if _, _, ok := parseProcStatLine("cpu 1 2 x"); ok {
		t.Fatal("champ non numérique accepté")
	}
}
