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

// OnSkip — hook de niveau package posé par l'hôte : une cellule skippée est
// journalisée (run, event, cellule) au lieu de laisser un trou muet dans le
// gel — la reprise re-joue la cellule, l'audit sait pourquoi elle manque.
var OnSkip func(runID string, eventID int, profile, qdisc, cc string)

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

// Start lance la matrice pleine en arrière-plan (tous qdiscs × toutes CC) ;
// la progression arrive dans Live.
func StartMatrix(base context.Context, profiles []string, reps int,
	deps Deps, dataDir string) (*Matrix, error) {
	return startMatrix(base, newRunID(), profiles, allQdiscStrings(), allCCStrings(), reps, deps, dataDir)
}

// StartMatrixFiltered — sous-matrice ciblée : axes qdiscs/CC explicites.
// Les cellules n=1 deviennent rejouables en 1 événement au lieu de relancer
// les 18 de la matrice pleine. Axes nil/vides ou valeurs inconnues = refus :
// pas de matrice fantôme.
func StartMatrixFiltered(base context.Context, profiles []string, qdiscs, ccs []string, reps int,
	deps Deps, dataDir string) (*Matrix, error) {
	return startMatrix(base, newRunID(), profiles, qdiscs, ccs, reps, deps, dataDir)
}

func StartMatrixWithID(base context.Context, runID string, profiles []string, reps int,
	deps Deps, dataDir string) (*Matrix, error) {
	return startMatrix(base, runID, profiles, allQdiscStrings(), allCCStrings(), reps, deps, dataDir)
}

// StartMatrixFilteredWithID — reprise d'une matrice interrompue : les
// cellules déjà gelées (run/event) sont sautées, seules les manquantes
// s'exécutent. Refuse un runID vide (pas de reprise fantôme).
func StartMatrixFilteredWithID(base context.Context, runID string, profiles []string, qdiscs, ccs []string, reps int,
	deps Deps, dataDir string) (*Matrix, error) {
	if runID == "" {
		return nil, fmt.Errorf("run-id vide : reprise impossible")
	}
	return startMatrix(base, runID, profiles, qdiscs, ccs, reps, deps, dataDir)
}

func allQdiscStrings() []string {
	out := make([]string, len(model.AllQdiscs))
	for i, q := range model.AllQdiscs {
		out[i] = string(q)
	}
	return out
}

func allCCStrings() []string {
	out := make([]string, len(model.AllCC))
	for i, c := range model.AllCC {
		out[i] = string(c)
	}
	return out
}

func startMatrix(base context.Context, runID string, profiles []string, qdiscs, ccs []string, reps int,
	deps Deps, dataDir string) (*Matrix, error) {
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles")
	}
	if reps <= 0 {
		reps = 3
	}
	// axes : filtres explicites (StartMatrixFiltered) — nil ou vide = refus :
	// une matrice sans cellule n'a rien à geler, le silence serait un
	// mensonge de progression. Les entrées pleine-matrice matérialisent
	// "tous" avant d'arriver ici.
	if len(qdiscs) == 0 {
		return nil, fmt.Errorf("no qdiscs selected")
	}
	if len(ccs) == 0 {
		return nil, fmt.Errorf("no cc selected")
	}
	validQ := map[string]bool{}
	for _, q := range model.AllQdiscs {
		validQ[string(q)] = true
	}
	for _, q := range qdiscs {
		if !validQ[q] {
			return nil, fmt.Errorf("qdisc inconnu: %s", q)
		}
	}
	validCC := map[string]bool{}
	for _, c := range model.AllCC {
		validCC[string(c)] = true
	}
	for _, c := range ccs {
		if !validCC[c] {
			return nil, fmt.Errorf("cc inconnu: %s", c)
		}
	}

	m := &Matrix{RunID: runID}
	total := 0
	for range profiles {
		total += len(qdiscs) * len(ccs) * reps
	}
	m.Total = total
	// la progression voyage avec chaque snapshot — un seul canal de vérité
	deps.TotalEvents = total

	w, err := OpenRun(dataDir + "/" + m.RunID)
	if err != nil {
		return nil, err
	}
	m.Done = len(w.seen)
	// DoneEvents APRÈS le comptage de reprise : sinon la progression gelée
	// (reprise d'un run interrompu) reste à 0 dans chaque snapshot
	deps.DoneEvents = m.Done
	ctx, cancel := context.WithCancel(base)
	m.cancel = cancel
	m.Running = true

	go func() {
		defer func() { w.Freeze("config"); m.mu.Lock(); m.Running = false; m.mu.Unlock() }()
		id := 1
		for _, pid := range profiles {
			model.ProfilesMu.RLock()
			prof, ok := model.Profiles[pid]
			model.ProfilesMu.RUnlock()
			if !ok {
				continue
			}
			for _, qs := range qdiscs {
				q := model.Qdisc(qs)
				for _, cs := range ccs {
					cc := model.CC(cs)
					for rep := 1; rep <= reps; rep++ {
						select {
						case <-ctx.Done():
							return
						default:
						}
						// identité de cellule normalisée : Direction "" = "up"
						// (défaut historique des archives) — la clé skip et la clé
						// Append (cellKey) doivent se rejoindre sur le même mot.
						dir := deps.Direction
						if dir == "" {
							dir = "up"
						}
						key := fmt.Sprintf("%s/%s/%s/%s/%d", pid, qs, cs, dir, rep)
						if w.seen[key] {
							m.Done = id
							id++
							continue
						}
						ev := model.Event{
							RunID: m.RunID, EventID: id, Profile: pid,
							Qdisc: q, CC: cc, Repetition: rep,
							Direction: dir,
						}
						// contexte annulable par event — le skip coupe la cellule
						// courante sans arrêter la matrice ; la phase de gel reste propre.
						evCtx, evCancel := context.WithCancel(ctx)
						m.setSkipCancel(evCancel)
						done, err := RunEvent(evCtx, ev, prof, deps)
						// ÉTAT FIGÉ AVANT evCancel : l'annulation ci-dessous
						// empoisonnerait evCtx.Err() et TOUTE erreur ressemblerait
						// à un skip (bug ancien : evCancel() précédait le test,
						// chaque échec tc était classé "skippé" + journalisé
						// quarantaine à tort, et le vrai message logué jamais).
						wasCancelled := evCtx.Err() != nil
						evCancel()
						if err == nil {
							_ = w.Append(done)
							m.Done = id
							// les cellules quarantaine arrivent dans le journal opérateur (Q13)
							if done.GateStatus == model.GateInvalid && OnQuarantine != nil {
								OnQuarantine(m.RunID, id, pid, string(q), string(cc))
							}
						} else if wasCancelled && ctx.Err() == nil {
							// skippé (pas arrêt global) — aucune ligne gelée, reprise
							// possible ; journalisé si l'hôte pose le hook (le trou
							// dans le gel reste explicable après coup). PAS de
							// quarantaine : une cellule coupée n'est pas invalidée
							// par les portes, le message mentirait.
							if OnSkip != nil {
								OnSkip(m.RunID, id, pid, string(q), string(cc))
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
