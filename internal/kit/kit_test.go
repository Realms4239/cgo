package kit

import (
	"os"
	"strings"
	"testing"
)

// classifySSHError — une VM propre échoue SSH de trois façons distinctes,
// chacune appelle une remédiation différente. Les messages d'OpenSSH sont
// stables d'une version à l'autre ; on matche les fragments canoniques.
func TestClassifySSHError(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want string
	}{
		{"refused", "ssh: connect to host 198.51.100.23 port 22: Connection refused", "refused"},
		{"refused nat", "ssh: connect to host 127.0.0.1 port 2222: Connection refused", "refused"},
		{"no route", "ssh: connect to host 10.0.2.15 port 22: No route to host", "unreachable"},
		{"timeout", "ssh: connect to host 198.51.100.23 port 22: Connection timed out", "unreachable"},
		{"auth", "ubuntu@x: Permission denied (publickey,password).", "auth"},
		{"auth key", "ubuntu@127.0.0.1: Permission denied (publickey).", "auth"},
		{"ok vide", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifySSHError(c.out); got != c.want {
				t.Errorf("classifySSHError(%q) = %q, want %q", c.out, got, c.want)
			}
		})
	}
}

// TestLoadConfigFlatAfterSection — régression : les clés plates APRÈS la
// section ssh: (vm_name, project_dir…) étaient préfixées ssh_ et jetées en
// silence (relues vides). L'indentation décide, pas la position.
func TestLoadConfigFlatAfterSection(t *testing.T) {
	dir := t.TempDir()
	yaml := dir + "/cgo-vm.yaml"
	os.WriteFile(yaml, []byte("ssh:\n  user: testuser\n  host: 198.51.100.23\n  port: 22\n  key: ~/.ssh/id_ed25519\nvm_name: \"ubu\"\nvmx_path: \"D:/VMs/ubu/ubu.vmx\"\nproject_dir: /home/testuser/cgo\ndashboard_port: 9091\n"), 0644)
	c, err := LoadConfig(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if c.SSHUser != "testuser" || c.SSHHost != "198.51.100.23" {
		t.Fatalf("section ssh perdue : user=%q host=%q", c.SSHUser, c.SSHHost)
	}
	if c.VMName != "ubu" || c.VMXPath != "D:/VMs/ubu/ubu.vmx" {
		t.Fatalf("clés plates perdues : name=%q vmx=%q", c.VMName, c.VMXPath)
	}
	if c.ProjectDir != "/home/testuser/cgo" || c.DashPort != "9091" {
		t.Fatalf("clés plates perdues : proj=%q port=%q", c.ProjectDir, c.DashPort)
	}
}

func TestSSHAdvice(t *testing.T) {
	// chaque classe d'échec doit porter un conseil actionnable, pas juste un code
	seen := map[string]string{}
	for _, cls := range []string{"refused", "auth", "unreachable"} {
		adv := sshAdvice(cls)
		if len(adv) < 20 {
			t.Errorf("sshAdvice(%q) trop court : %q", cls, adv)
		}
		seen[cls] = adv
	}
	if seen["refused"] == seen["auth"] {
		t.Error("refused et auth doivent donner des conseils distincts")
	}
	if seen["auth"] == seen["unreachable"] {
		t.Error("auth et unreachable doivent donner des conseils distincts")
	}
}

// TCP joignable sur port 22 : le advice "refused" doit mentionner sshd install,
// pas d'inonder l'opérateur d'auth alors que le port est fermé.
func TestSSHAdviceRefusedMentionsSSHService(t *testing.T) {
	if !strings.Contains(sshAdvice("refused"), "sshd") {
		t.Errorf("sshAdvice(refused) doit orienter vers sshd : %q", sshAdvice("refused"))
	}
}

// LoadConfig : l'environnement DOIT gagner sur le yaml (contrat documenté
// en tête de kit.go) — scénario VM propre/DHCP : CGO_SSH_HOST force la cible
// même quand cgo-vm.yaml retient l'ancienne IP.
func TestLoadConfigEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yaml := dir + "/cgo-vm.yaml"
	if err := os.WriteFile(yaml, []byte("ssh:\n  host: 198.51.100.23\n  port: \"22\"\n  user: testuser\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CGO_SSH_HOST", "10.11.12.13")
	t.Setenv("CGO_SSH_PORT", "2229")
	c, err := LoadConfig(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if c.SSHHost != "10.11.12.13" {
		t.Errorf("SSHHost = %q, want l'env 10.11.12.13 (env doit gagner sur le yaml)", c.SSHHost)
	}
	if c.SSHPort != "2229" {
		t.Errorf("SSHPort = %q, want l'env 2229", c.SSHPort)
	}
	if c.SSHUser != "testuser" {
		t.Errorf("SSHUser = %q, want testuser (du yaml, non surchargé)", c.SSHUser)
	}
}

// shq — un chemin distant avec espace, quote, $ ou ; doit survivre au
// shell distant intact (vérifié en rejouant via sh -c sur le poste).
func TestShq(t *testing.T) {
	for in, want := range map[string]string{
		"/home/u/cgo":        "'/home/u/cgo'",
		"/home/a b/cgo":      "'/home/a b/cgo'",
		"/home/o'b/x":        "'/home/o'\\''b/x'",
		"/home/a$b/c;d`e/x":  "'/home/a$b/c;d`e/x'",
		"":                   "''",
		"/tmp/cgo-backup-1.tgz": "'/tmp/cgo-backup-1.tgz'",
	} {
		if got := shq(in); got != want {
			t.Errorf("shq(%q) = %q, want %q", in, got, want)
		}
	}
}
