package metrics

// Modèle économique entreprise — spéc. 2026-10-25 §4. Les tarifs sont des
// scénarios contractuels sourcés (docs/data-prices.md), pas des mesures.

// EnterprisePlan — liaison dédiée CIR (modèle A : dépréciation de capacité).
type EnterprisePlan struct {
	Name           string  `json:"name"`
	MonthlyOpExMGA float64 `json:"monthly_opex_mga"`
	CIRMbps        float64 `json:"cir_mbps"`
	FUPBytes       float64 `json:"fup_bytes"`
}

// Plans entreprise Yas Grande Entreprise (yas.mg, relevé 31/08/2026).
var (
	PlanGOLD = EnterprisePlan{Name: "enterprise-gold-cir", MonthlyOpExMGA: 1_000_000, CIRMbps: 200}
	// 30 To = 30 × 2^40 octets : 2 500 000 / 30 720 GiB = 81,38 Ar/Go.
	PlanPLATINE = EnterprisePlan{Name: "enterprise-platine1", MonthlyOpExMGA: 2_500_000, FUPBytes: 30 * 1024 * 1024 * 1024 * 1024}
)

// AmortizedARPerSecond — coût amorti de la liaison à la seconde.
// Le mois contractuel est compté à 720 h (30 j × 24 h), convention du mémoire.
func AmortizedARPerSecond(monthlyOpEx float64) float64 {
	return monthlyOpEx / (720 * 3600)
}

// CapacityWasteARPerHour — dépréciation horaire de capacité contractuelle :
// la part des octets transportés gaspillée en réémissions, rapportée au
// douzième de l'OpEx mensuel. Un lien dédié ne facture pas l'octet ; il
// facture la capacité que le bufferbloat rend inutilisable.
func CapacityWasteARPerHour(monthlyOpEx, wastedBytes, transportedBytes float64) float64 {
	if transportedBytes <= 0 {
		return 0
	}
	return (wastedBytes / transportedBytes) * (monthlyOpEx / 720)
}
