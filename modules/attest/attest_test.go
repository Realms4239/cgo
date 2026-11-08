package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestVerifyFile — scellé correct passe, scellé faux = erreur.
func TestVerifyFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "data.csv")
	data := []byte("a,b\n1,2\n")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	good := Entry{Path: p, SHA256: hex.EncodeToString(sum[:])}
	if err := VerifyFile(good); err != nil {
		t.Fatalf("scellé correct rejeté : %v", err)
	}
	bad := Entry{Path: p, SHA256: "0000000000000000000000000000000000000000000000000000000000000000"}
	if err := VerifyFile(bad); err == nil {
		t.Fatal("scellé faux accepté")
	}
}

// TestParseEmptyManifestErrors — manifeste vide ou que commentaires = erreur.
func TestParseEmptyManifestErrors(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	comments := filepath.Join(t.TempDir(), "comments.txt")
	if err := os.WriteFile(comments, []byte("# rien ici\n# que des commentaires\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{empty, comments} {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		_, perr := Parse(f)
		_ = f.Close()
		if perr == nil {
			t.Fatalf("manifeste sans entrée accepté : %s", p)
		}
	}
}
