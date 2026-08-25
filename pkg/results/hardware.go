package results

import "fmt"

// HardwareRecommendation maps a proven qdisc to COTS hardware translation.
// Single source for results.Scan and api/translate and frontend mirror web/frontend/src/lib/hardware.ts
func HardwareRecommendation(bestQdisc, profile string) string {
	if bestQdisc == "fq_codel" || bestQdisc == "cake" {
		return fmt.Sprintf("Si MikroTik: Queue Tree PCQ/CAKE RouterOS v7+ pour %s — %s prouvé en lab", profile, bestQdisc)
	}
	return fmt.Sprintf("Si ISP/mini-PC gateway: transparent bridge CAKE en amont du CPE pour %s — pilote isolé d'abord", profile)
}
