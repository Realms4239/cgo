package campagne

import (
	"context"
	"sync"
	"testing"
	"time"
)

type errTC struct{}

func (errTC) Run(args ...string) ([]byte, error) {
	return nil, errFakeTC
}

type fakeTCError string

func (e fakeTCError) Error() string { return string(e) }

const errFakeTC = fakeTCError("tc simulé en panne")

// TestFailureIsNotASkip — une erreur tc GÉNUINE n'est ni un skip ni une
// quarantaine : pas d'événement OnSkip, pas d'événement OnQuarantine, aucune
// ligne gelée. Régression du bug ancien où evCancel() précédait le test
// evCtx.Err() : TOUTE erreur ressemblait à un skip (journal "coupée par
// l'opérateur" pour des pannes tc) et le vrai message n'était jamais logué.
func TestFailureIsNotASkip(t *testing.T) {
	var mu sync.Mutex
	var skips, quars int
	OnSkip = func(string, int, string, string, string) {
		mu.Lock()
		skips++
		mu.Unlock()
	}
	OnQuarantine = func(string, int, string, string, string) {
		mu.Lock()
		quars++
		mu.Unlock()
	}
	defer func() { OnSkip = nil; OnQuarantine = nil }()

	deps := fastDeps()
	deps.TC = errTC{}
	m, err := StartMatrixWithID(context.Background(), "run-failcls", []string{"P2"}, 1, deps, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	for i := 0; i < 40 && m.IsRunning(); i++ {
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if skips != 0 {
		t.Fatalf("erreurs tc classées skip: %d", skips)
	}
	if quars != 0 {
		t.Fatalf("erreurs tc journalisées quarantaine: %d", quars)
	}
}
