package campagne

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestSkipJournalsViaHook — une cellule skippée doit être traçable après
// coup via OnSkip (posé par l'hôte comme OnQuarantine) : run, event,
// cellule. Sans hook, le skip reste silencieux mais non paniquant.
func TestSkipJournalsViaHook(t *testing.T) {
	var mu sync.Mutex
	var got []string
	OnSkip = func(runID string, eventID int, profile, qdisc, cc string) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, profile+"/"+qdisc+"/"+cc)
	}
	defer func() { OnSkip = nil }()

	deps := fastDeps()
	// les fenêtres instantanées de fastDeps rendraient la cellule en ~10 ms —
	// la sonde Small bloque jusqu'à annulation : la cellule 1 vit
	// réellement, la fenêtre de skip est garantie ouverte
	deps.Small = func(ctx context.Context) (float64, error) {
		<-ctx.Done()
		return 25, nil
	}
	m, err := StartMatrixWithID(context.Background(), "run-skip-j", []string{"P2"}, 1, deps, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop()
	// laisser la 1re cellule démarrer (setSkipCancel posé par la boucle
	// matrice juste avant RunEvent) — le bulk bloqué garde la fenêtre ouverte
	time.Sleep(100 * time.Millisecond)
	m.Skip()
	for i := 0; i < 60 && m.IsRunning(); i++ {
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("skip muet — OnSkip jamais appelé")
	}
	if got[0] == "//" {
		t.Fatalf("journal sans cellule: %v", got)
	}
}
