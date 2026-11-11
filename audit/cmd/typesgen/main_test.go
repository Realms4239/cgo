package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTypesGenIsCurrent — types.gen.ts doit être la sortie du générateur, octets
// normalisés. Un miroir écrit à la main ou un Go qui change sans régénérer la
// TSUITE casse ici, pas dans la console de l'opérateur.
// Si le frontend n'est pas présent (module audit consommé hors dépôt principal),
// le test se tait : il n'a pas le fichier à comparer.
func TestTypesGenIsCurrent(t *testing.T) {
	target := filepath.Join("..", "..", "web", "frontend", "src", "lib", "types.gen.ts")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Skipf("frontend absent (%v) : types.gen.ts n'est pas vérifiable ici", err)
	}
	gen, err := render()
	if err != nil {
		t.Fatalf("génération : %v", err)
	}
	norm := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
	got := norm(gen)
	want := norm(b)
	if !bytes.Equal(got, want) {
		// première divergence, pour un diagnostic immédiat
		gLines, wLines := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(gLines) && i < len(wLines); i++ {
			if gLines[i] != wLines[i] {
				t.Fatalf("types.gen.ts n'est pas à jour (ligne %d) :\n  généré : %s\n  fichier: %s\nRelancer : go generate ./... dans le module audit", i+1, gLines[i], wLines[i])
			}
		}
		t.Fatalf("types.gen.ts n'est pas à jour (longueur %d ≠ %d) — relancer : go generate ./... dans le module audit", len(gLines), len(wLines))
	}
}
