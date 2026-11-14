package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Realms4239/cgo/audit"
)

type purgeFixture struct {
	t          *testing.T
	data       string
	sessions   map[string]time.Time
}

func newPurgeFixture(t *testing.T) *purgeFixture {
	t.Helper()
	data := filepath.Join(t.TempDir(), "data")
	f := &purgeFixture{t: t, data: data, sessions: map[string]time.Time{}}
	// Cinq sessions scellées datées par le nom (audit-<unix>) ; une session
	// homonyme sans sceau ; des voisins qui ne sont PAS des sessions scellées.
	now := time.Now().UTC()
	for _, ageDays := range []int{120, 80, 40, 10, 2} {
		when := now.AddDate(0, 0, -ageDays)
		name := fmt.Sprintf("audit-%d", when.Unix())
		f.addSession(name, true)
		f.sessions[name] = when
	}
	// Dossier au nom de session mais sans sceau : intouchable.
	f.addSession("audit-"+fmt.Sprint(now.AddDate(0, 0, -200).Unix()), false)
	// Corpus gelé-style : jamais une cible.
	for _, other := range []string{"camp-p1", "runs", "AQM-run-001"} {
		dir := filepath.Join(data, "audits", other)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, audit.ReportArtifactName), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Corpus hors audits/ : data/runs/... doit survivre.
	if err := os.MkdirAll(filepath.Join(data, "runs", "camp-x"), 0o755); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *purgeFixture) addSession(name string, sealed bool) {
	f.t.Helper()
	dir := filepath.Join(f.data, "audits", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		f.t.Fatal(err)
	}
	if sealed {
		if err := os.WriteFile(filepath.Join(dir, audit.ReportArtifactName), []byte("{}"), 0o644); err != nil {
			f.t.Fatal(err)
		}
	}
}

func (f *purgeFixture) exists(name string) bool {
	f.t.Helper()
	fi, err := os.Stat(filepath.Join(f.data, "audits", name))
	return err == nil && fi.IsDir()
}

func (f *purgeFixture) sealedNames() []string {
	f.t.Helper()
	names := []string{}
	for name := range f.sessions {
		if f.exists(name) {
			names = append(names, name)
		}
	}
	return names
}

func cutoffInDays(t *testing.T, daysAgo int) string {
	t.Helper()
	return time.Now().UTC().AddDate(0, 0, -daysAgo).Format("2006-01-02")
}

func TestPurgeBeforeRemovesOlderThanCutoffAndKeepsLastN(t *testing.T) {
	f := newPurgeFixture(t)
	// Tout est antérieur à maintenant — un --keep de 2 doit sauver les 2 plus
	// récentes scellées, même hors fenêtre d'âge.
	code := purgeCmd([]string{"--keep", "2", "--before", cutoffInDays(t, 0), f.data})
	if code != 0 {
		t.Fatalf("purge = %d, want 0", code)
	}
	left := f.sealedNames()
	if len(left) != 2 {
		t.Fatalf("sessions scellées restantes = %v, want 2", left)
	}
	// data/runs et les dossiers sans sceau survivent.
	if !f.exists("AQM-run-001") || !pathExists(filepath.Join(f.data, "runs", "camp-x")) {
		t.Fatal("corpus gelé touché par la purge")
	}
	// La session sans sceau la plus ancienne doit être écartée, pas supprimée.
	entries, err := os.ReadDir(filepath.Join(f.data, "audits"))
	if err != nil {
		t.Fatal(err)
	}
	unsealed := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(f.data, "audits", e.Name(), audit.ReportArtifactName)); err != nil {
			unsealed++
		}
	}
	if unsealed != 1 {
		t.Fatalf("sessions sans sceau restantes = %d, want 1", unsealed)
	}
}

func TestPurgeKeepsSessionsNewerThanCutoff(t *testing.T) {
	f := newPurgeFixture(t)
	// Coupure il y a 30 jours, --keep 0 : seules les sessions strictement
	// antérieures sont supprimables ; les deux plus récentes survivent.
	code := purgeCmd([]string{"--keep", "0", "--before", cutoffInDays(t, 30), f.data})
	if code != 0 {
		t.Fatalf("purge = %d, want 0", code)
	}
	remaining := f.sealedNames()
	if len(remaining) != 2 {
		t.Fatalf("sessions restantes = %v, want 2 (10 j et 2 j)", remaining)
	}
}

func TestPurgeDryRunDeletesNothing(t *testing.T) {
	f := newPurgeFixture(t)
	before := f.sealedNames()
	code := purgeCmd([]string{"--keep", "2", "--before", cutoffInDays(t, 0), "--dry-run", f.data})
	if code != 0 {
		t.Fatalf("purge --dry-run = %d, want 0", code)
	}
	after := f.sealedNames()
	if len(after) != len(before) {
		t.Fatalf("--dry-run a supprimé %d session(s)", len(before)-len(after))
	}
}

func TestPurgeWithoutBeforeDeletesNothing(t *testing.T) {
	f := newPurgeFixture(t)
	before := f.sealedNames()
	code := purgeCmd([]string{"--keep", "1", f.data})
	if code != 0 {
		t.Fatalf("purge sans --before = %d, want 0", code)
	}
	if len(f.sealedNames()) != len(before) {
		t.Fatal("une suppression a eu lieu sans --before")
	}
}

func TestPurgeBadDateIsUsageError(t *testing.T) {
	f := newPurgeFixture(t)
	if code := purgeCmd([]string{"--before", "14-11-2026", f.data}); code != 2 {
		t.Fatalf("date invalide = %d, want 2", code)
	}
	if code := purgeCmd([]string{"--keep", "-1", f.data}); code != 2 {
		t.Fatalf("--keep négatif = %d, want 2", code)
	}
}

func TestPurgeAcceptsAuditsRootDirectly(t *testing.T) {
	f := newPurgeFixture(t)
	code := purgeCmd([]string{"--keep", "1", "--before", cutoffInDays(t, 0), "--dry-run", filepath.Join(f.data, "audits")})
	if code != 0 {
		t.Fatalf("répertoire audits direct = %d, want 0", code)
	}
}

func TestPurgeNothingToDoIsZero(t *testing.T) {
	f := newPurgeFixture(t)
	before := f.sealedNames()
	// Fenêtre --keep grande : rien d'antérieur ne sort de la fenêtre.
	code := purgeCmd([]string{"--keep", "50", "--before", cutoffInDays(t, 0), f.data})
	if code != 0 {
		t.Fatalf("purge = %d, want 0", code)
	}
	if len(f.sealedNames()) != len(before) {
		t.Fatal("sessions supprimées alors que --keep couvre tout")
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
