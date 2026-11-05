package api

import (
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/results"
)

// duelDir — même source que /api/results ; variable pour que le test pointe
// sur un corpus jetable sans écrire dans data/runs.
var duelDir = "data/runs"

var duelProfile = regexp.MustCompile(`^P[0-9]$`)

// duelMinReps — une cellule d'une répétition n'est pas un gagnant de
// référence : l'historique du banc connaît des cellules isolées favorisées par
// l'ordonnanceur. Trois répétitions est le défaut de la matrice.
const duelMinReps = 3

type duelCell struct {
	Qdisc       string   `json:"qdisc"`
	CC          string   `json:"cc"`
	DeadlinePct float64  `json:"deadline_pct"`
	SmallP95Ms  float64  `json:"small_p95_ms"`
	N           int      `json:"n"`
	Runs        []string `json:"runs,omitempty"`
}

type duelResponse struct {
	Profile          string    `json:"profile"`
	DeadlineMs       int       `json:"deadline_ms"`
	MinReps          int       `json:"min_reps"`
	Base             *duelCell `json:"base"`
	Best             *duelCell `json:"best"`
	DeltaDeadlinePts float64   `json:"delta_deadline_pts"`
	DeltaSmallMs     float64   `json:"delta_small_ms"`
	Reason           string    `json:"reason,omitempty"`
}

func qdiscRank(q string) int {
	switch q {
	case "pfifo_fast":
		return 0
	case "fq_codel":
		return 1
	case "cake":
		return 2
	}
	return 9
}

func eligible(g results.Group) bool {
	return g.Direction == "up" && g.DeadlineValidN >= duelMinReps
}

func handleDuel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile := strings.TrimSpace(r.URL.Query().Get("profile"))
		if profile == "" {
			profile = "P2"
		}
		if !duelProfile.MatchString(profile) {
			writeErr(w, r, "profil invalide", http.StatusBadRequest)
			return
		}
		run := r.URL.Query().Get("run")
		if run != "" && !validRunID(run) {
			writeErr(w, r, "id de run invalide", http.StatusBadRequest)
			return
		}
		groups, _ := results.Scan(duelDir, run)
		var cells []results.Group
		for _, g := range groups {
			if g.Profile == profile && eligible(g) {
				cells = append(cells, g)
			}
		}
		resp := duelResponse{Profile: profile, DeadlineMs: metrics.CAPDeadlineMs, MinReps: duelMinReps}
		if len(cells) == 0 {
			resp.Reason = "aucune cellule valide avec au moins " + strconv.Itoa(duelMinReps) + " répétitions — relancez une campagne"
			writeJSON(w, resp)
			return
		}
		sort.Slice(cells, func(i, j int) bool {
			a, b := cells[i], cells[j]
			if a.DeadlineValidMedian != b.DeadlineValidMedian {
				return a.DeadlineValidMedian > b.DeadlineValidMedian
			}
			if a.Smallp95ValidMedian != b.Smallp95ValidMedian {
				return a.Smallp95ValidMedian < b.Smallp95ValidMedian
			}
			if qdiscRank(a.Qdisc) != qdiscRank(b.Qdisc) {
				return qdiscRank(a.Qdisc) < qdiscRank(b.Qdisc)
			}
			return a.CC < b.CC
		})
		best := cellOf(cells[0])
		var base *duelCell
		for _, g := range cells {
			if g.Qdisc == "pfifo_fast" && (base == nil || g.DeadlineValidN > base.N) {
				base = ptr(cellOf(g))
			}
		}
		resp.Best = ptr(best)
		resp.Base = base
		if base != nil {
			resp.DeltaDeadlinePts = round1(best.DeadlinePct - base.DeadlinePct)
			resp.DeltaSmallMs = round1(best.SmallP95Ms - base.SmallP95Ms)
		} else {
			resp.Reason = "aucune ligne de base pfifo_fast exploitable"
		}
		writeJSON(w, resp)
	}
}

func cellOf(g results.Group) duelCell {
	return duelCell{
		Qdisc: g.Qdisc, CC: g.CC,
		DeadlinePct: g.DeadlineValidMedian,
		SmallP95Ms:  g.Smallp95ValidMedian,
		N:           g.DeadlineValidN,
		Runs:        g.Runs,
	}
}

func ptr[T any](v T) *T { return &v }

// round1 — une décimale, arrondi arithmétique. `pkg/api` n'a pas d'arrondi
// réutilisable (vérifié avant d'écrire cette tâche).
func round1(v float64) float64 { return math.Round(v*10) / 10 }
