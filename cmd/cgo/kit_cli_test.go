package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/internal/kit"
)

// keysetup sans terminal : usage → 2, jamais 0 (stdin remplacé par un
// fichier : isTerminal faux, aucun prompt, aucun réseau).
func TestRunKitKeysetupNonInteractive(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cgo-vm.yaml")
	if err := os.WriteFile(cfg, []byte("ssh:\n  user: u\n  host: 192.0.2.1\n  port: 22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	in, err := os.OpenFile(filepath.Join(dir, "in"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	old := os.Stdin
	os.Stdin = in
	defer func() { os.Stdin = old }()
	if code := runKit([]string{"--config", cfg, "keysetup"}); code == 0 {
		t.Error("keysetup non interactif : want non-zero, got 0")
	}
}

// testbed check sans script local : 6, jamais 0 (aucun réseau : le binaire
// de test ne voit pas kit/testbed.sh depuis son répertoire).
func TestRunKitTestbedFailNonZero(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cgo-vm.yaml")
	if err := os.WriteFile(cfg, []byte("ssh:\n  user: u\n  host: 192.0.2.1\n  port: 22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := runKit([]string{"--config", cfg, "testbed"}); code == 0 {
		t.Error("testbed en échec : want non-zero, got 0")
	}
}

// testbed bogus : usage → 2 (avant tout E/S).
func TestRunKitTestbedBadSub(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cgo-vm.yaml")
	if err := os.WriteFile(cfg, []byte("ssh:\n  user: u\n  host: 192.0.2.1\n  port: 22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := runKit([]string{"--config", cfg, "testbed", "bogus"}); code != 2 {
		t.Errorf("testbed bogus : got %d, want 2", code)
	}
}

// testbed down accepté comme sous-commande (le scp vers l'injoignable
// échoue ensuite → 7, pas « inconnue » → 2) ; check sur la même cible →
// non-zéro aussi. Clé inexistante : échec scp immédiat, sans réseau.
func TestTestbedDownAndCheckSSHFail(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "kit"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kit", "testbed.sh"), []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := kit.NewRunner()
	r.Root = dir
	r.Stdout, r.Stderr = io.Discard, io.Discard
	c := &kit.Config{SSHUser: "u", SSHHost: "192.0.2.1", SSHPort: "22",
		SSHKey: filepath.Join(dir, "absente"), ProjectDir: "/home/u/cgo"}
	if code := r.Testbed(c, []string{"down"}); code == 0 {
		t.Error("testbed down en échec : want non-zero, got 0")
	}
	if code := r.Testbed(c, nil); code == 0 {
		t.Error("testbed check en échec : want non-zero, got 0")
	}
}
