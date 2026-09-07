package kit

import (
	"os"
	"path/filepath"
	"testing"
)

// validateKeySetupTarget — garde-fous des prompts (aucune E/S) : l'annulation
// d'une cible invalide doit arriver AVANT toute connexion, jamais après.
func TestValidateKeySetupTarget(t *testing.T) {
	cases := []struct {
		name, user, host, port string
		wantErr                bool
	}{
		{"nominal", "testuser", "198.51.100.23", "22", false},
		{"user vide", "", "h", "22", true},
		{"user espace", "a b", "h", "22", true},
		{"user injection", "a;rm", "h", "22", false}, // ; seul n'est pas bloquant (pas d'espace/séparateur ssh)
		{"user arobase", "a@b", "h", "22", true},     // casserait user@host
		{"user deux-points", "a:b", "h", "22", true}, // casserait le port scp/ssh
		{"user dollar", "a$b", "h", "22", true},      // expansion distante
		{"user backtick", "a`b", "h", "22", true},    // substitution distante
		{"host vide", "u", "", "22", true},
		{"host espace", "u", "a b", "22", true},
		{"port vide", "u", "h", "", true},
		{"port zero", "u", "h", "0", true},
		{"port lettres", "u", "h", "ssh", true},
		{"port max", "u", "h", "65535", false},
		{"port trop grand", "u", "h", "65536", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateKeySetupTarget(c.user, c.host, c.port)
			if (err != nil) != c.wantErr {
				t.Errorf("validateKeySetupTarget(%q,%q,%q) err=%v, wantErr=%v",
					c.user, c.host, c.port, err, c.wantErr)
			}
		})
	}
}

// resolvePubkey — refuse une clé absente ou malformée au lieu de poser du vide.
func TestResolvePubkey(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(good, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good+".pub", []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := resolvePubkey(good); err != nil || got != good+".pub" {
		t.Errorf("resolvePubkey bon = %q, %v", got, err)
	}
	if _, err := resolvePubkey(filepath.Join(dir, "absente")); err == nil {
		t.Error("resolvePubkey absente : want error")
	}
	bad := filepath.Join(dir, "malformee")
	if err := os.WriteFile(bad+".pub", []byte("ceci n'est pas une clé\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := resolvePubkey(bad); err == nil {
		t.Error("resolvePubkey malformée : want error")
	}
}
