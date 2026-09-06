package api

import (
	"encoding/json"
	"fmt"
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
	groups, _ := results.Scan("data/runs", "")
	if len(groups) == 0 {
		// F8 : pas encore de données — vide honnête, pas de succès synthétique
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"available": false, "reason": "aucune donnée — lancez campagne"})
		return
	}
	for _, g := range groups {
		if g.Profile == profile && g.Best {
			bestQdisc = g.Qdisc
			break
		}
	}
	rec := results.HardwareRecommendation(bestQdisc, profile)

	// interprétation riche — calculée depuis les mêmes groupes gelés : la
	// meilleure cellule vs la référence pfifo du même profil, écarts et
	// prescription en langage opérateur. Aucune valeur fabriquée.
	var best, pfifo *results.Group
	for i := range groups {
		g := groups[i]
		if g.Profile != profile {
			continue
		}
		if g.Best && best == nil {
			best = &groups[i]
		}
		if g.Qdisc == "pfifo_fast" && pfifo == nil {
			pfifo = &groups[i]
		}
	}
	smallDiff := ""
	if best != nil && pfifo != nil && pfifo.Smallp95Median > 0 && best.Smallp95Median > 0 {
		pct := int((pfifo.Smallp95Median - best.Smallp95Median) / pfifo.Smallp95Median * 100)
		if pct >= 0 {
			smallDiff = fmt.Sprintf("sous charge, %s protège le trafic critique : small p95 %.1f ms vs %.1f ms pour pfifo (−%d %%)", best.Qdisc, best.Smallp95Median, pfifo.Smallp95Median, pct)
		} else {
			smallDiff = fmt.Sprintf("sous charge, %s ne protège pas mieux que pfifo sur ce profil : small p95 %.1f ms vs %.1f ms", best.Qdisc, best.Smallp95Median, pfifo.Smallp95Median)
		}
	} else if best != nil {
		smallDiff = fmt.Sprintf("small p95 mesuré %.1f ms — pas de référence pfifo sur ce profil pour comparer", best.Smallp95Median)
	}
	// régime perte : toutes les cellules valides du profil ratent l'échéance
	// (deadline_med 0 partout) — la retransmission gouverne la sonde, pas la
	// file. Le verdict le dit au lieu de comparer des disciplines
	// indiscernables, et la prochaine étape pointe la vraie expérience :
	// saturer le lien malgré la perte (charge multi-flux) pour que la file
	// ait quelque chose à arbitrer.
	regime := ""
	nValid, nZero := 0, 0
	for i := range groups {
		g := groups[i]
		if g.Profile != profile || g.Count-g.Quarantined <= 0 {
			continue
		}
		nValid++
		if g.DeadlineMedian == 0 {
			nZero++
		}
	}
	if nValid >= 2 && nZero == nValid {
		regime = fmt.Sprintf("régime perte sur %s : %d cellules valides, échéance 0 %% partout — une retransmission coûte un aller simple, aucune discipline ne se distingue ; ne choisissez pas une file sur ce tableau", profile, nValid)
	}
	throughput := ""
	if best != nil && pfifo != nil && pfifo.GoodputMedian > 0 && best.GoodputMedian > 0 {
		pct := int((pfifo.GoodputMedian - best.GoodputMedian) / pfifo.GoodputMedian * 100)
		if pct > 0 {
			throughput = fmt.Sprintf("coût en débit : %.1f Mbit/s vs %.1f (−%d %%)", best.GoodputMedian, pfifo.GoodputMedian, pct)
		} else {
			throughput = fmt.Sprintf("débit préservé : %.1f Mbit/s vs %.1f pour pfifo", best.GoodputMedian, pfifo.GoodputMedian)
		}
	}
	nextSteps := []string{
		"appliquer la prescription sur un équipement de test, en fenêtre de maintenance",
		"mesurer 24 h en surveillance continue (deadline alertée au journal)",
		"comparer ce run au précédent dans Résultats (delta par cellule)",
	}
	if regime != "" {
		smallDiff = regime
		nextSteps = append([]string{
			"rejouer la campagne en charge multi-flux (BulkN) : saturer le lien malgré la perte, et seulement là les disciplines se sépareront",
		}, nextSteps...)
	}
	cli := ""
	if best != nil && profile != "" {
		cli = fmt.Sprintf("sudo tc qdisc replace dev eth0 root %s bandwidth %dmbit", best.Qdisc, int(max(1, best.GoodputMedian*0.9)))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"available": true, "recommendation": rec,
		"verdict": smallDiff, "throughput": throughput, "cli": cli,
		"next_steps": nextSteps,
		"cell": func() map[string]any {
			if best == nil {
				return nil
			}
			return map[string]any{"qdisc": best.Qdisc, "cc": best.CC, "small_p95": best.Smallp95Median, "rtt_p95": best.RTTp95Median, "goodput": best.GoodputMedian, "deadline_ok": best.DeadlineMedian}
		}(),
	})
}
