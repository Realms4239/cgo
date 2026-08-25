package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Realms4239/cgo/pkg/results"
)

func hardwareRecommendation(bestQdisc, profile string) string {
	if bestQdisc == "fq_codel" || bestQdisc == "cake" {
		return fmt.Sprintf("Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour %s — %s prouvé en lab", profile, bestQdisc)
	}
	return fmt.Sprintf("Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour %s — pilote isolé d'abord", profile)
}

// HandleTranslate GET /api/hardware/translate?profile=P2 -> {recommendation}
func HandleTranslate(w http.ResponseWriter, r *http.Request) {
	profile := r.URL.Query().Get("profile")
	if profile == "" {
		profile = "P2"
	}
	// try to find best qdisc from Scan if data exists
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
	}
	rec := hardwareRecommendation(bestQdisc, profile)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"recommendation": rec})
}
