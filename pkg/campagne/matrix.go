package campagne

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Realms4239/cgo/pkg/model"
)

// OnQuarantine — package-level hook set by the host (cmd/cgo) so quarantined
// cells land in the operator journal (Q13). Nil ⇒ silent, tests stay quiet.
var OnQuarantine func(runID string, eventID int, profile, qdisc, cc string)

// Matrix drives the full LIEN experiment matrix (Tableau 3):
// profiles × qdiscs × CC × repetitions = 36 events (18 when reduced to P2).
type Matrix struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	Running bool

	RunID string
	Total int
	Done  int
}

// IsRunning is the race-free accessor for Running (H3).
func (m *Matrix) IsRunning() bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Running
}

// Start launches the matrix in the background; progress lands in Live.
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
						done, err := RunEvent(ctx, ev, prof, deps)
						if err == nil {
							_ = w.Append(done)
							m.Done = id
							// quarantined cells land in the operator journal (Q13)
							if done.GateStatus == model.GateInvalid && OnQuarantine != nil {
								OnQuarantine(m.RunID, id, pid, string(q), string(cc))
							}
						} else {
							log.Printf("[campagne] cell %s failed: %v", key, err)
						}
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
