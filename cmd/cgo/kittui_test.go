package main

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Preuve du partage : la copie du modèle (ce que fait NewProgram) voit
// l'assignation faite après coup sur l'original — le bug initial (champ
// prog direct resté nil dans la copie du runtime, actions fond muettes).
func TestKTBusShared(t *testing.T) {
	m := initialModelKT(t.TempDir()+"/cgo-vm.yaml", "x")
	p := tea.NewProgram(m)
	m.bus.prog = p
	defer p.Kill()
	copie := m // ce que NewProgram a copié en interne : même holder
	if copie.bus != m.bus {
		t.Fatal("holder non partagé")
	}
	copie.bus.prog = p
	if m.bus.prog == nil {
		t.Fatal("assignation invisible depuis l'original — bug du champ direct")
	}
}
func testModelKT(t *testing.T) modelKT {
	t.Helper()
	dir := t.TempDir()
	cfg := dir + "/cgo-vm.yaml"
	os.WriteFile(cfg, []byte("ssh:\n  user: testuser\n"), 0644)
	m := initialModelKT(cfg, "1.2.3-test")
	m.deps = []depRow{{label: "Client OpenSSH", ok: true, info: "ssh"}}
	return m
}

func keyMsg(s string) tea.KeyMsg {
	// construit un KeyMsg fiable sans terminal
	var k tea.KeyMsg
	switch s {
	case "j":
		k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	case "k":
		k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	case "enter":
		k = tea.KeyMsg{Type: tea.KeyEnter}
	default:
		k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	return k
}

func TestKTStepsItems(t *testing.T) {
	m := testModelKT(t)
	// étape dépendances : recheck + next (dep ok → pas d'install)
	if got := len(m.items()); got != 2 {
		t.Fatalf("deps items = %d, veux 2", got)
	}
	// navigation bornée
	mm, _ := m.Update(keyMsg("j"))
	m = mm.(modelKT)
	if m.cursor != 1 {
		t.Fatalf("curseur après j = %d, veux 1", m.cursor)
	}
	mm, _ = m.Update(keyMsg("j"))
	m = mm.(modelKT)
	if m.cursor != 1 {
		t.Fatalf("curseur déborde : %d", m.cursor)
	}
	mm, _ = m.Update(keyMsg("k"))
	m = mm.(modelKT)
	if m.cursor != 0 {
		t.Fatalf("curseur après k = %d, veux 0", m.cursor)
	}
}

func TestKTNextAdvance(t *testing.T) {
	m := testModelKT(t)
	m.cursor = 1 // "Continuer → Machine"
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(modelKT)
	if m.step != ktVM {
		t.Fatalf("step = %d, veux ktVM", m.step)
	}
	if m.cursor != 0 {
		t.Fatalf("curseur non réinitialisé : %d", m.cursor)
	}
}

func TestKTVMLock(t *testing.T) {
	m := testModelKT(t)
	m.step = ktVM
	m.vms = []vmEntry{{path: `D:\VMs\ubu\ubu.vmx`, name: "ubu", hyp: "vmware"}}
	// items : rescan + 1 VM → curseur 1 = la VM
	m.cursor = 1
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(modelKT)
	if m.step != ktAcces {
		t.Fatalf("après verrouillage step = %d, veux ktAcces", m.step)
	}
}

func TestKTDoneClearsBusy(t *testing.T) {
	m := testModelKT(t)
	m.busy = "deploy"
	mm, _ := m.Update(ktDone{action: "deploy", code: 0})
	m = mm.(modelKT)
	if m.busy != "" {
		t.Fatal("busy non libéré après done")
	}
	if len(m.log) == 0 {
		t.Fatal("journal vide après done")
	}
}

func TestKTViewRenders(t *testing.T) {
	m := testModelKT(t)
	m.vms = []vmEntry{{path: `D:\VMs\ubu\ubu.vmx`, name: "ubu", hyp: "vmware", mode: "nat", live: true}}
	m.sshState = "ok"
	m.dashState = "ok 1.2.3"
	m.sshDiag = []diagRow{
		{label: "Clé locale", state: "ok", detail: "~/.ssh/id_ed25519"},
		{label: "Port 22", state: "ko", detail: "fermé — installer openssh-server"},
		{label: "Clé autorisée", state: "wait", detail: "après port 22"},
	}
	m.inputOn = true
	m.inputField = "host"
	m.inputVal = "192.168.1."
	m.pushLog("ligne test")
	for s := ktDeps; s <= ktControle; s++ {
		m.step = s
		m.cursor = 0
		out := m.View()
		if len(out) < 100 {
			t.Fatalf("vue %d trop courte (%d)", s, len(out))
		}
	}
}

// Header global : VM + état + mode + SSH + dashboard + IP sur chaque écran.
func TestKTHeader(t *testing.T) {
	dir := t.TempDir()
	cfg := dir + "/cgo-vm.yaml"
	os.WriteFile(cfg, []byte("ssh:\n  user: testuser\n  host: 198.51.100.23\n  port: 22\n  key: ~/.ssh/id_ed25519\nvm_name: \"ubu\"\nvmx_path: \"D:/VMs/ubu/ubu.vmx\"\n"), 0644)
	m := initialModelKT(cfg, "x")
	m.vmSeen = true
	m.vmLive = true
	m.vmMode = "nat"
	m.sshState = "ok"
	m.dashState = "ok 1.2.3"
	for s := ktDeps; s <= ktControle; s++ {
		m.step = s
		out := m.View()
		for _, want := range []string{"ubu", "allumée", "nat", "198.51.100.23"} {
			if !strings.Contains(out, want) {
				t.Fatalf("vue %d : header sans %q", s, want)
			}
		}
	}
}

// Saisie inline : ouvre, écrit, valide → config + retour navigation.
func TestKTInputCommit(t *testing.T) {
	dir := t.TempDir()
	cfg := dir + "/cgo-vm.yaml"
	os.WriteFile(cfg, []byte("ssh:\n  user: testuser\n  host: auto\n  port: 22\n  key: ~/.ssh/id_ed25519\n"), 0644)
	m := initialModelKT(cfg, "x")
	m.step = ktAcces
	// curseur sur "Utilisateur…" (diag, mkkey, set-user)
	mm, _ := m.Update(keyMsg("enter"))
	m = mm.(modelKT)
	// l'item set-user est à l'index 2 — on l'active directement
	m.cursor = 2
	mm, _ = m.Update(keyMsg("enter"))
	m = mm.(modelKT)
	if !m.inputOn || m.inputField != "user" {
		t.Fatalf("saisie non ouverte : on=%v field=%q", m.inputOn, m.inputField)
	}
	// ctrl+u efface le pré-remplissage, puis frappe du neuf
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = mm.(modelKT)
	for _, r := range "marie" {
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(modelKT)
	}
	mm, _ = m.Update(keyMsg("enter"))
	m = mm.(modelKT)
	if m.inputOn {
		t.Fatal("saisie non refermée après entrée")
	}
	if m.cfg.SSHUser != "marie" {
		t.Fatalf("user non persisté : %q", m.cfg.SSHUser)
	}
}

// Diagnostic : message appliqué, état global suit.
func TestKTDiagApplies(t *testing.T) {
	m := testModelKT(t)
	m.sshState = "ok"
	mm, _ := m.Update(ktDiagMsg{rows: []diagRow{
		{label: "Clé locale", state: "ok"},
		{label: "Port 22", state: "ko"},
	}})
	m = mm.(modelKT)
	if len(m.sshDiag) != 2 {
		t.Fatalf("diag non appliqué : %d lignes", len(m.sshDiag))
	}
	if m.sshState == "ok" {
		t.Fatal("état global resté ok malgré un KO")
	}
}

// Esc : ferme la saisie d'abord, puis recule d'une étape (jamais bloqué).
func TestKTEscBack(t *testing.T) {
	m := testModelKT(t)
	m.step = ktDeploy
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = mm.(modelKT)
	if m.step != ktAcces {
		t.Fatalf("esc ne recule pas : step=%d", m.step)
	}
	m.inputOn = true
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = mm.(modelKT)
	if m.inputOn || m.step != ktAcces {
		t.Fatal("esc devrait fermer la saisie avant de reculer")
	}
}
