package campagne

import (
	"context"
	"testing"
	"time"
)

// TestSkipJournalsViaHook — une cellule skippée doit être traçable après
// coup via OnSkip (posé par l'hôte comme OnQuarantine) : run, event,
// cellule. Sans hook, le skip reste silencieux mais non paniquant.
func TestSkipJournalsViaHook(t *testing.T) {
	skipped := make(chan string, 1)
	OnSkip = func(_ string, _ int, profile, qdisc, cc string) {
		skipped <- profile + "/" + qdisc + "/" + cc
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
	defer stopWait(t, m)
	// laisser la 1re cellule démarrer (setSkipCancel posé par la boucle
	// matrice juste avant RunEvent) — le bulk bloqué garde la fenêtre ouverte
	time.Sleep(100 * time.Millisecond)
	m.Skip()
	// la matrice ne finit pas d'elle-même : les cellules 2..9 bloquent aussi
	// dans Small, seul Stop les libère. On attend donc le journal, pas la fin.
	select {
	case cell := <-skipped:
		if cell == "//" {
			t.Fatalf("journal sans cellule: %s", cell)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("skip muet — OnSkip jamais appelé")
	}
}
