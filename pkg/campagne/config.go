package campagne

import (
	"encoding/json"

	"github.com/Realms4239/cgo/pkg/model"
)

// matrixConfig — configuration effective de la matrice, gelée dans le
// manifeste avec son SHA-256. Les profils sont résolus (paramètres réels,
// pas seulement l'identifiant) : une modification d'import change le hash.
type matrixConfig struct {
	Profiles    []model.Profile `json:"profiles"`
	Qdiscs      []string        `json:"qdiscs"`
	CCs         []string        `json:"ccs"`
	Reps        int             `json:"reps"`
	Direction   string          `json:"direction"`
	Target      string          `json:"target,omitempty"`
	BulkAddr    string          `json:"bulk_addr,omitempty"`
	SmallURL    string          `json:"small_url,omitempty"`
	Flows       int             `json:"flows"`
	BaselineSec int             `json:"baseline_sec"`
	ChargeSec   int             `json:"charge_sec"`
	DeadlineMs  float64         `json:"deadline_ms,omitempty"`
}

func buildMatrixConfig(profiles, qdiscs, ccs []string, reps int, deps Deps) ([]byte, error) {
	resolved := make([]model.Profile, 0, len(profiles))
	model.ProfilesMu.RLock()
	for _, id := range profiles {
		p, ok := model.Profiles[id]
		if !ok {
			p = model.Profile{ID: id}
		}
		if p.ID == "" {
			p.ID = id
		}
		resolved = append(resolved, p)
	}
	model.ProfilesMu.RUnlock()

	dir := deps.Direction
	if dir == "" {
		dir = "up"
	}
	flows := deps.Flows
	if flows <= 0 {
		flows = 1
	}
	baseline := deps.BaselineSec
	if baseline == 0 {
		baseline = model.BaselineSec
	}
	charge := deps.ChargeSec
	if charge == 0 {
		charge = model.ChargeSec
	}
	return json.Marshal(matrixConfig{
		Profiles: resolved, Qdiscs: qdiscs, CCs: ccs, Reps: reps,
		Direction: dir, Target: deps.Target, BulkAddr: deps.BulkAddr,
		SmallURL: deps.SmallURL, Flows: flows, BaselineSec: baseline,
		ChargeSec: charge, DeadlineMs: deps.DeadlineMs,
	})
}
