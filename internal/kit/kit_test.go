package kit

import (
	"os"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/internal/vm"
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

// LoadConfigFast — même parse que LoadConfig, SANS résolution réseau :
// même avec host:auto + un chemin VM, aucun subprocess (garantie thread UI).
func TestLoadConfigFastNoResolve(t *testing.T) {
	dir := t.TempDir()
	yaml := dir + "/cgo-vm.yaml"
	yml := "ssh:\n  user: testuser\n  host: auto\n  port: \"22\"\nvm_name: \"ubu\"\nvmx_path: \"D:/VMs/ubu/ubu.vmx\"\ndashboard_host: \"meteolink.dev\"\n"
	if err := os.WriteFile(yaml, []byte(yml), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfigFast(yaml)
	if err != nil {
		t.Fatal(err)
	}
	if c.SSHHost != "auto" {
		t.Errorf("SSHHost = %q, want auto (pas de résolution en Fast)", c.SSHHost)
	}
	if c.SSHUser != "testuser" || c.VMName != "ubu" || c.DashHost != "meteolink.dev" {
		t.Errorf("parse incomplet : %+v", c)
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

// fakeHyp — pilote factice (nom seul utile : les décisions testées ne
// touchent jamais au binaire).
type fakeHyp struct{ name string }

func (f *fakeHyp) Name() string                    { return f.name }
func (f *fakeHyp) Exe() string                     { return "" }
func (f *fakeHyp) Running() []string               { return nil }
func (f *fakeHyp) Start(vmx string) error          { return nil }
func (f *fakeHyp) StartGUI(vmx string) error       { return nil }
func (f *fakeHyp) Stop(vmx string) error           { return nil }
func (f *fakeHyp) GuestIP(vmx string) string       { return "" }
func (f *fakeHyp) NetMode(vmx string) string       { return "" }
func (f *fakeHyp) SetNetMode(vmx, mode string) error { return nil }

// selectDriver — même bug racine, côté ensure/deploy : le .vbox obtient
// le pilote virtualbox même si vmware est détecté en premier.
// FirstOpen — premier palier non-vert : la prochaine étape, jamais un
// palier déjà vert ni un « en attente » plus loin dans la file.
// VMPath — un seul accesseur pour les deux champs (le rapport terrain :
// tout le kit ignorait vbox_path — verrou, boot, Guest, Next).
func TestVMPath(t *testing.T) {
	if got := (&Config{VMXPath: "a.vmx"}).VMPath(); got != "a.vmx" {
		t.Fatalf("vmx seul → %q", got)
	}
	if got := (&Config{VBoxPath: "b.vbox"}).VMPath(); got != "b.vbox" {
		t.Fatalf("vbox seul → %q", got)
	}
	if got := (&Config{VMXPath: "a.vmx", VBoxPath: "b.vbox"}).VMPath(); got != "a.vmx" {
		t.Fatalf("les deux → vmx prioritaire, got %q", got)
	}
	if got := (&Config{}).VMPath(); got != "" {
		t.Fatalf("vide → vide, got %q", got)
	}
}

// SaveVMXPurge — basculer .vmx→.vbox ne laisse pas un vmx_path fantôme
// prioritaire (sinon VMPath ressuscite l'ancienne VM).
func TestSaveVMXPurge(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/cgo-vm.yaml"
	if err := SaveVMX(p, `D:\ubuntu.vmx`, "vmware"); err != nil {
		t.Fatal(err)
	}
	if err := SaveVMX(p, `D:\VMs\Fanasina\Fanasina.vbox`, "virtualbox"); err != nil {
		t.Fatal(err)
	}
	c, _ := LoadConfig(p)
	if c.VMPath() != `D:\VMs\Fanasina\Fanasina.vbox` {
		t.Fatalf("VMPath = %q après bascule", c.VMPath())
	}
	if err := SaveVMX(p, `D:\ubuntu.vmx`, "vmware"); err != nil {
		t.Fatal(err)
	}
	c, _ = LoadConfig(p)
	if c.VMPath() != `D:\ubuntu.vmx` || c.VBoxPath != "" {
		t.Fatalf("retour vmx : path=%q vbox=%q", c.VMPath(), c.VBoxPath)
	}
}

func TestFirstOpen(t *testing.T) {
	allOK := []NextStep{{ID: "vm", State: "ok"}, {ID: "cle", State: "ok"}}
	if nx := FirstOpen(allOK); nx != nil {
		t.Fatalf("tout vert → nil, got %q", nx.ID)
	}
	mixed := []NextStep{
		{ID: "vm", State: "ok"},
		{ID: "cle", State: "ko", Remedy: "cgo kit keysetup", Verb: "console:keysetup"},
		{ID: "binaire", State: "attente"},
	}
	nx := FirstOpen(mixed)
	if nx == nil || nx.ID != "cle" || nx.Verb != "console:keysetup" {
		t.Fatalf("got %+v, want l'étape clé", nx)
	}
	waitOnly := []NextStep{
		{ID: "vm", State: "ok"},
		{ID: "cible", State: "attente", Remedy: "cgo kit ensure"},
		{ID: "port", State: "attente"},
	}
	if nx := FirstOpen(waitOnly); nx == nil || nx.ID != "cible" {
		t.Fatalf("attente bloque aussi, got %+v", nx)
	}
}

// GatherNextPalier1VBox — le palier 1 VOIT une config VBox (vbox_path
// seul) : avant, « aucune VM verrouillée » systématique. Le pilote peut
// manquer ici (pas de VBoxManage) — mais le CHEMIN doit être reconnu
// (détail = le basename, pas « aucune VM »).
func TestGatherNextPalier1VBox(t *testing.T) {
	dir := t.TempDir()
	vbx := dir + "/Fanasina.vbox"
	os.WriteFile(vbx, []byte("faux vbox"), 0644)
	p := dir + "/cgo-vm.yaml"
	os.WriteFile(p, []byte("ssh:\n  user: fanasina\n  host: auto\nvm_name: \"Fanasina\"\nvbox_path: \""+vbx+"\"\nhypervisor: \"virtualbox\"\n"), 0644)
	c, _ := LoadConfig(p)
	if c.VMPath() != vbx {
		t.Fatalf("VMPath ne voit pas vbox_path : %q", c.VMPath())
	}
	steps := GatherNext(c)
	if len(steps) == 0 || steps[0].ID != "vm" {
		t.Fatalf("palier 1 absent : %+v", steps)
	}
	if !strings.Contains(steps[0].Detail, "Fanasina.vbox") && !strings.Contains(steps[0].Detail, "virtualbox") {
		t.Fatalf("palier 1 aveugle au .vbox : %+v", steps[0])
	}
}

func TestSelectDriver(t *testing.T) {
	vb := &fakeHyp{name: "virtualbox"}
	vw := &fakeHyp{name: "vmware"}
	both := []vm.Hypervisor{vw, vb} // vmware EN PREMIER (ordre Detect)
	vbox := `D:\VMs\ubuntu Fanasina\ubuntu Fanasina.vbox`
	vmx := `D:\ubuntu.vmx`
	if h, err := selectDriver(both, "auto", vbox); err != nil || h.Name() != "virtualbox" {
		t.Fatalf("vbox+auto → virtualbox, got %v err %v", h, err)
	}
	if h, err := selectDriver(both, "vmware", vbox); err != nil || h.Name() != "virtualbox" {
		t.Fatalf("vbox+config vmware → virtualbox quand même, got %v err %v", h, err)
	}
	if h, err := selectDriver(both, "auto", vmx); err != nil || h.Name() != "vmware" {
		t.Fatalf("vmx+auto → vmware, got %v err %v", h, err)
	}
	if _, err := selectDriver([]vm.Hypervisor{vw}, "auto", vbox); err == nil {
		t.Fatal("vbox sans VBoxManage devrait échouer explicitement")
	}
	if h, err := selectDriver(both, "virtualbox", `D:\bizarre`); err != nil || h.Name() != "virtualbox" {
		t.Fatalf("chemin ambigu + config vbox → virtualbox, got %v err %v", h, err)
	}
}
// hypNameForScan — régression du bug racine : un .vbox reste virtualbox
// même quand vmrun existe (primaire vmware). Le primaire ne sert que
// pour les chemins sans extension connue.
func TestHypNameForScan(t *testing.T) {
	vb := &fakeHyp{name: "virtualbox"}
	vw := &fakeHyp{name: "vmware"}
	for _, tc := range []struct {
		pick, want string
		prim       vm.Hypervisor
	}{
		{`D:\VMs\ubuntu Fanasina\ubuntu Fanasina.vbox`, "virtualbox", vw},
		{`D:\VMs\ubuntu Fanasina\ubuntu Fanasina.vbox`, "virtualbox", vb},
		{`D:\VMs\ubuntu Fanasina\ubuntu Fanasina.vbox`, "virtualbox", nil},
		{`D:\ubuntu.vmx`, "vmware", vb},
		{`D:\ubuntu.vmx`, "vmware", nil},
		{`D:\bizarre`, "virtualbox", vb},
		{`D:\bizarre`, "vmware", vw},
		{`D:\bizarre`, "vmware", nil},
	} {
		if got := hypNameForScan(tc.pick, tc.prim); got != tc.want {
			t.Errorf("hypNameForScan(%q) = %q, want %q", tc.pick, got, tc.want)
		}
	}
}

// keepExplicitTarget — le forward NAT manuel (loopback) est sacré :
// l'auto-découverte ne l'écrase jamais, même avec une IP invitée valide.
func TestKeepExplicitTarget(t *testing.T) {
	for host, want := range map[string]bool{
		"127.0.0.1": true, "127.0.0.2": true, "localhost": true,
		"::1": true, "[::1]": true, "  127.0.0.1  ": true,
		"192.168.174.128": false, "10.0.2.15": false, "": false,
		"auto": false, "meteolink.dev": false,
	} {
		if got := keepExplicitTarget(host); got != want {
			t.Errorf("keepExplicitTarget(%q) = %v, want %v", host, got, want)
		}
	}
}
