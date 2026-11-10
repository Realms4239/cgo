package campagne

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Le module qui écrit est le dernier rempart : un run-id hostile ne doit
// JAMAIS créer un dossier hors du corpus, quel que soit l'appelant.
// Ces tests échouent sur l'ancien code (aucune validation dans startMatrix).

func TestHostileRunIDRefusedAndNothingWritten(t *testing.T) {
	cases := []string{"..", "../escape", "a/b", `a\b`, `..\..`, "a.b", ".", ""}
	for _, id := range cases {
		dir := t.TempDir()
		_, err := StartMatrixWithID(context.Background(), id, []string{"P2"}, 1, Deps{}, dir)
		if err == nil {
			t.Fatalf("run-id %q : accepté par le banc", id)
		}
		// rien n'a été créé sous le dataDir
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Fatalf("run-id %q : le banc a créé %v", id, entries)
		}
	}
}

func TestFilteredWithIDHostileRefusedAndNothingWritten(t *testing.T) {
	dir := t.TempDir()
	_, err := StartMatrixFilteredWithID(context.Background(), "../freeze",
		[]string{"P2"}, []string{"cake"}, []string{"bbr"}, 1, Deps{}, dir)
	if err == nil {
		t.Fatal("`../freeze` accepté par la reprise")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("le banc a créé %v avec un id hostile", entries)
	}
}

// Un id bien formé reste pleinement fonctionnel après le verrouillage.
func TestValidRunIDStillWrites(t *testing.T) {
	dir := t.TempDir()
	m, err := StartMatrixWithID(context.Background(), "run-jail-test", []string{"P2"}, 1, fastDeps(), dir)
	if err != nil {
		t.Fatalf("id bien formé refusé : %v", err)
	}
	defer stopWait(t, m)
	if _, err := os.Stat(filepath.Join(dir, "run-jail-test")); err != nil {
		t.Fatalf("dossier de run manquant : %v", err)
	}
}

// La complétion du run ne doit pas lire un dossier hors corpus non plus
// (reprise sur un id « ../… » = lecture hors données).
func TestResumeHostileIDRefused(t *testing.T) {
	dir := t.TempDir()
	_, err := StartMatrixFilteredWithID(context.Background(), `run\up`,
		[]string{"P2"}, []string{"cake"}, []string{"bbr"}, 1, Deps{}, dir)
	if err == nil {
		t.Fatal("`run\\up` accepté par la reprise")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("reprise hostile a créé %v", entries)
	}
}
