package main

import (
	"strings"
	"testing"
)

// Exhaustivité : chaque bouton câblé a une action, chaque action existe.
func TestGUIButtonMap(t *testing.T) {
	seen := map[string]int{}
	for _, group := range [][]buttonDef{groupActions, accessActions, deployActions, controlActions} {
		for _, b := range group {
			kind, _ := buttonAction(b.id)
			if kind == "" {
				t.Errorf("bouton %d %q sans action", b.id, b.label)
			}
			seen[kind]++
		}
	}
	// actions critiques présentes
	for _, want := range []string{"bg:scan", "bg:deploy", "bg:diag", "bg:mkkey", "console:keysetup", "bg:ensure", "bg:svc", "bg:nic-toggle"} {
		found := false
		for _, group := range [][]buttonDef{groupActions, accessActions, deployActions, controlActions} {
			for _, b := range group {
				if k, _ := buttonAction(b.id); k == want {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("action %q non câblée à un bouton", want)
		}
	}
	// libellés non vides, ids uniques
	ids := map[int]string{}
	for _, group := range [][]buttonDef{groupActions, accessActions, deployActions, controlActions} {
		for _, b := range group {
			if strings.TrimSpace(b.label) == "" {
				t.Errorf("bouton %d sans libellé", b.id)
			}
			if prev, ok := ids[b.id]; ok {
				t.Errorf("id %d dupliqué (%q vs %q)", b.id, prev, b.label)
			}
			ids[b.id] = b.label
		}
	}
}

// Actions console (mot de passe / admin) : jamais en fond silencieux.
func TestGUIConsoleActions(t *testing.T) {
	for _, id := range []int{213, 236, 237} {
		kind, _ := buttonAction(id)
		if !strings.HasPrefix(kind, "console:") {
			t.Errorf("bouton %d : %q devrait être console:", id, kind)
		}
	}
}

// Nouveaux boutons intégration scripts : mappés, jamais muets.
func TestGUIScriptButtons(t *testing.T) {
	for id, want := range map[int]string{245: "bg:hosttun", 246: "bg:guest", 247: "bg:vnet"} {
		kind, _ := buttonAction(id)
		if kind != want {
			t.Errorf("bouton %d : %q, voulu %q", id, kind, want)
		}
	}
}
