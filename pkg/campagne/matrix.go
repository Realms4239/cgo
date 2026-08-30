package campagne

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Realms4239/cgo/pkg/model"
)

// OnQuarantine — hook de niveau package posé par l'hôte (cmd/cgo) pour que les
// cellules quarantaine arrivent dans le journal opérateur (Q13). Nil ⇒ silencieux, tests tranquilles.
var OnQuarantine func(runID string, eventID int, profile, qdisc, cc string)

// Matrice d'expérimentation complète: profils × files d'attente × CC ×
// répétitions.
// profils × qdiscs × CC × répétitions = 36 événements (18 en réduit P2).
type Matrix struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	Running bool

	RunID string
	Total int
	Done  int

	// cancel de l'event courant — le skip coupe la cellule, pas la campagne
	skipCancel context.CancelFunc
}

func (m *Matrix) setSkipCancel(c context.CancelFunc) {
	m.mu.Lock()
	m.skipCancel = c
	m.mu.Unlock()
}

// Skip interrompt la cellule en cours sans arrêter la matrice.
func (m *Matrix) Skip() {
	m.mu.Lock()
	c := m.skipCancel
	m.mu.Unlock()
	if c != nil {
		c()
	}
}

// IsRunning est l'accesseur sans course pour Running (H3).
func (m *Matrix) IsRunning() bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Running
}

// Start lance la matrice en arrière-plan ; la progression arrive dans Live.
func StartMatrix(base context.Context, profiles []string, reps int,
	deps Deps, live *Live, dataDir string) (*Matrix, error) {
	return StartMatrixWithID(base, newRunID(), profiles, reps, deps, live, dataDir)
}

func StartMatrixWithID(base context.Context, runID string, profiles []string, reps int,
	deps Deps, live *Live, dataDir string) (*Matrix, error) {
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles")
	}
	if reps <= 0 {
		reps = 3
	}

	m := &Matrix{RunID: runID}
	total := 0
	for range profiles {
		total += len(model.AllQdiscs) * len(model.AllCC) * reps
	}
	m.Total = total
	// la progression voyage avec chaque snapshot — un seul canal de vérité
	deps.TotalEvents = total
	deps.DoneEvents = m.Done

	w, err := OpenRun(dataDir + "/" + m.RunID)
	if err != nil {
		return nil, err
	}
	m.Done = len(w.seen)
	ctx, cancel := context.WithCancel(base)
	m.cancel = cancel
	m.Running = true

	go func() {
		defer func() { w.Freeze("config"); m.mu.Lock(); m.Running = false; m.mu.Unlock() }()
		id := 1
		for _, pid := range profiles {
			prof, ok := model.Profiles[pid]
			if !ok {
				continue
			}
			for _, q := range model.AllQdiscs {
				for _, cc := range model.AllCC {
					for rep := 1; rep <= reps; rep++ {
						select {
						case <-ctx.Done():
							return
						default:
						}
					key := fmt.Sprintf("%s/%d", m.RunID, id)
					if w.seen[key] {
						m.Done = id
						id++
						continue
					}
					ev := model.Event{
						RunID: m.RunID, EventID: id, Profile: pid,
						Qdisc: q, CC: cc, Repetition: rep,
					}
					// contexte annulable par event — le skip coupe la cellule
					// courante sans arrêter la matrice ; la phase de gel reste propre.
					evCtx, evCancel := context.WithCancel(ctx)
					m.setSkipCancel(evCancel)
					done, err := RunEvent(evCtx, ev, prof, deps)
					evCancel()
					if err == nil {
						_ = w.Append(done)
						m.Done = id
						// les cellules quarantaine arrivent dans le journal opérateur (Q13)
						if done.GateStatus == model.GateInvalid && OnQuarantine != nil {
							OnQuarantine(m.RunID, id, pid, string(q), string(cc))
						}
					} else if evCtx.Err() != nil && ctx.Err() == nil {
						// skippé (pas arrêt global) — aucune ligne gelée, reprise possible
						if OnQuarantine != nil {
							OnQuarantine(m.RunID, id, pid, string(q), string(cc))
						}
					} else {
						log.Printf("[campagne] cell %s failed: %v", key, err)
					}
					m.setSkipCancel(nil)
					id++
					}
				}
			}
		}
	}()
	return m, nil
}

func (m *Matrix) Stop() {
	m.mu.Lock()
	c := m.cancel
	m.mu.Unlock()
	if c != nil {
		c()
	}
}

func newRunID() string {
	return fmt.Sprintf("run-%d", nowUnix())
}
