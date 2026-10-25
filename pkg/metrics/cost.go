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

// B2BTariff — tarif grossiste/apn facturé au gigaoctet (modèle B).
type B2BTariff struct {
	Name    string  `json:"name"`
	ARPerGB float64 `json:"ar_per_gb"`
	// Quoted marque les tarifs relevés sur devis opérateur (pas de grille
	// publique) : le rapport doit le dire, c'est un scénario contractuel.
	Quoted bool `json:"quoted,omitempty"`
}

var (
	// Dérivé de PLATINE 1 : 2 500 000 Ar / 30 720 GiB.
	TariffWholesale = B2BTariff{Name: "b2b-platine-bulk", ARPerGB: 81.38}
	// APN corporate M2M/IoT — devis opérateur (oct. 2026), pas une grille publique.
	TariffCorporateAPN = B2BTariff{Name: "b2b-m2m-apn", ARPerGB: 150, Quoted: true}
)

// RetailEmergencyARPerGB — le prépayé grand public (Yas Net Month 4,5 Go)
// ne survit que comme pénalité de hors-forfait d'urgence, jamais comme
// modèle par défaut d'une infrastructure opérationnelle.
const RetailEmergencyARPerGB = 5556

// MeteredWasteARPerHour — coût horaire du gaspillage sur un lien facturé au
// volume : le ×20 de la directive est le cas particulier fenêtre = 180 s ;
// l'audit passe sa fenêtre de charge MESURÉE, la campagne son 180 s historique.
func MeteredWasteARPerHour(wastedBytes, tariffARPerGB, windowSeconds float64) float64 {
	if windowSeconds <= 0 {
		return 0
	}
	gb := wastedBytes / (1024 * 1024 * 1024)
	return gb * tariffARPerGB * (3600 / windowSeconds)
}

// Objectif opérationnel DGM pour une transaction CAP (sonde 16 KiB). Ce n'est
// PAS un seuil imposé par l'OMM : le Guide WIS 2.0 (WMO-No 1061) définit le
// « temps réel » comme « de quelques secondes à quelques minutes », sans
// seuil numérique par bulletin — le mémoire cite ce cadre comme contexte,
// jamais comme source du 220 ms.
const (
	CAPDeadlineMs         = 220
	CAPBreachThresholdPct = 95.0
)

// CAPSLA — conformité des transactions CAP sous charge (modèle C).
type CAPSLA struct {
	Measured bool    `json:"measured"`
	OKPct    float64 `json:"ok_pct"`
	Breach   bool    `json:"breach"`
	Label    string  `json:"label,omitempty"`
}

// AssessCAPSLA évalue les complétions 16 KiB CHARGÉES contre l'objectif DGM.
// Sans échantillon : non mesuré (Measured=false), jamais un passage par défaut.
func AssessCAPSLA(loadedSmallCompletionsMs []float64) CAPSLA {
	if len(loadedSmallCompletionsMs) == 0 {
		return CAPSLA{}
	}
	pct := DeadlineOKPct(loadedSmallCompletionsMs, CAPDeadlineMs)
	s := CAPSLA{Measured: true, OKPct: pct, Breach: pct < CAPBreachThresholdPct}
	if s.Breach {
		s.Label = "ALERTE OPÉRATIONNELLE — MANQUEMENT SLA (objectif DGM 220 ms, cadre WIS 2.0)"
	}
	return s
}
