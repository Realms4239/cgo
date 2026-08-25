package api

import (
	"encoding/json"
	"net/http"

	"github.com/Realms4239/cgo/pkg/results"
)

func hardwareRecommendation(bestQdisc, profile string) string {
	return results.HardwareRecommendation(bestQdisc, profile)
}

// HandleTranslate GET /api/hardware/translate?profile=P2 -> {recommendation} or {available:false}
func HandleTranslate(w http.ResponseWriter, r *http.Request) {
	profile := r.URL.Query().Get("profile")
	if profile == "" {
		profile = "P2"
	}
	bestQdisc := "fq_codel"
	if profile == "P1" {
		bestQdisc = "cake"
	}
	if groups, _ := results.Scan("data/runs", ""); len(groups) > 0 {
		for _, g := range groups {
			if g.Profile == profile && g.Best {
				bestQdisc = g.Qdisc
				break
			}
		}
	} else {
		// F8: no data yet — honest empty, not synthetic success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"available": false, "reason": "aucune donnée — lancez campagne"})
		return
	}
	rec := results.HardwareRecommendation(bestQdisc, profile)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"recommendation": rec})
}
