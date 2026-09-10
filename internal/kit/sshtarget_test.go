package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveSSHTarget(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cgo-vm.yaml")
	os.WriteFile(p, []byte("ssh:\n  user: testuser\n  host: auto\n  port: 22\n  key: ~/.ssh/id_ed25519\n"), 0644)
	if err := SaveSSHTarget(p, "marie", "192.168.1.50", "", ""); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.SSHUser != "marie" || c.SSHHost != "192.168.1.50" {
		t.Fatalf("relus user=%q host=%q", c.SSHUser, c.SSHHost)
	}
	if c.SSHPort != "22" {
		t.Fatalf("port perdu : %q", c.SSHPort)
	}
}

// TestSaveSSHTargetMissing — « Sauver » avant tout verrou (yaml pas encore
// né) crée le fichier au lieu de « fichier introuvable » (vu en prod :

// l'ordre naturel taper-user → Sauver échouait).
func TestSaveSSHTargetMissing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "cgo-vm.yaml")
	if err := SaveSSHTarget(p, "fanasina", "", "", ""); err != nil {
		t.Fatalf("création : %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "user: fanasina") {
		t.Fatalf("contenu inattendu : %q", b)
	}
}
