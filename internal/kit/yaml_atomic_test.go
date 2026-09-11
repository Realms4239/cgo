package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "cgo-vm.yaml")
	if err := writeFileAtomic(p, []byte("a: 1\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "a: 1\n" {
		t.Fatalf("roundtrip = %q, %v", b, err)
	}
	// écrasement : le fichier reste lisible et complet (pas de tronqué).
	if err := writeFileAtomic(p, []byte("a: 2\nb: 3\n"), 0644); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
	b, _ = os.ReadFile(p)
	if string(b) != "a: 2\nb: 3\n" {
		t.Fatalf("overwrite = %q", b)
	}
	// pas de résidu tmp dans le dossier.
	entries, _ := os.ReadDir(dir + "/sub")
	for _, e := range entries {
		if len(e.Name()) > 8 && e.Name()[:8] == ".cgo-tmp" {
			t.Fatalf("résidu tmp : %s", e.Name())
		}
	}
}

func TestResolveProjectDirExplicit(t *testing.T) {
	c := &Config{ProjectDir: "/srv/cgo", SSHUser: "u"}
	if got := c.resolveProjectDir(); got != "/srv/cgo" {
		t.Fatalf("explicite = %q", got)
	}
}

func TestResolveProjectDirFallback(t *testing.T) {
	// hôte injoignable (port fermé local) : pas de $HOME → repli documenté,
	// jamais "" (qui ferait scp vers nulle part).
	c := &Config{SSHUser: "u", SSHHost: "127.0.0.1", SSHPort: "9"}
	if got := c.resolveProjectDir(); got != "/home/u/cgo" {
		t.Fatalf("repli = %q", got)
	}
	if got := (&Config{}).resolveProjectDir(); got != "" {
		t.Fatalf("vide = %q, want \"\"", got)
	}
}
func TestSetYAMLKeyAtomic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cgo-vm.yaml")
	if err := setYAMLKey(p, "host", "127.0.0.1"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := setYAMLKey(p, "host", "192.168.1.2"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	c, err := LoadConfig(p)
	if err != nil || c.SSHHost != "192.168.1.2" {
		t.Fatalf("relit host=%q err=%v", c.SSHHost, err)
	}
}
