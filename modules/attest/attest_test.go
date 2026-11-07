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
