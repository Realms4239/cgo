package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Le centre de contrôle se teste sans terminal : Update est pure,
// les actions lourdes partent en goroutines (non testées ici).
func testModelKT() modelKT {
	m := initialModelKT("kit/cgo-vm.yaml.example", "1.2.3-test")
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
	m := testModelKT()
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
	m := testModelKT()
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
	m := testModelKT()
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
	m := testModelKT()
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
	m := testModelKT()
	m.vms = []vmEntry{{path: `D:\VMs\ubu\ubu.vmx`, name: "ubu", hyp: "vmware", live: true}}
	m.sshState = "ok"
	m.dashState = "ok 1.2.3"
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
