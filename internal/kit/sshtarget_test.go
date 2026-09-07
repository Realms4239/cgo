package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveSSHTarget(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cgo-vm.yaml")
	os.WriteFile(p, []byte("ssh:\n  user: altfloat\n  host: auto\n  port: 22\n  key: ~/.ssh/id_ed25519\n"), 0644)
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
