package metrics

import (
	"fmt"
	"strconv"
)

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
		s.Label = fmt.Sprintf("ALERTE OPÉRATIONNELLE — MANQUEMENT SLA (objectif DGM %d ms, cadre WIS 2.0)", CAPDeadlineMs)
	}
	return s
}

// RetransRatio — part des octets réémis. Basis dit d'où elle vient :
// "measured" (TCP_INFO sur nos sockets), "estimated" (proxy perte sonde ×
// volume), "" (non mesuré — alors aucun scénario n'est chiffré).
type RetransRatio struct {
	Ratio float64 `json:"ratio"`
	Basis string  `json:"basis"`
}

// EconomicsScenario — une valorisation horaire du gaspillage mesuré.
type EconomicsScenario struct {
	Model     string           `json:"model"` // "A" (CIR) | "B" (au volume) | "emergency"
	Name      string           `json:"name"`
	ARPerHour float64          `json:"ar_per_hour"`
	Basis     string           `json:"basis"`
	Numbers   EconomicsNumbers `json:"numbers"`
}

// EconomicsNumbers — la base chiffrée d'un scénario, dans son propre modèle.
// Le texte `Basis` reste lisible par un humain ; l'interface ne parse plus une
// phrase française pour afficher un nombre.
type EconomicsNumbers struct {
	Model          string  `json:"model"`                      // "A" | "B" | "emergency"
	MonthlyOpExMGA float64 `json:"monthly_opex_mga,omitempty"` // A
	CIRMbps        float64 `json:"cir_mbps,omitempty"`         // A
	ARPerGB        float64 `json:"ar_per_gb,omitempty"`        // B et emergency
	RetransRatio   float64 `json:"retrans_ratio"`
	WastedBytes    float64 `json:"wasted_bytes"`
	TransportBytes float64 `json:"transport_bytes"`
	LoadSeconds    float64 `json:"load_seconds"`
}

// numbers — les entrées du scénario telles qu'elles ont été calculées. Aucun
// nombre n'est inventé : un modèle qui ne consomme pas une entrée la laisse à
// zéro, et `omitempty` la retire du JSON plutôt que de la montrer mesurée.
func numbers(model string, opex, cir, arPerGB, ratio float64, transported, wasted, loadSeconds float64) EconomicsNumbers {
	return EconomicsNumbers{
		Model: model, MonthlyOpExMGA: opex, CIRMbps: cir, ARPerGB: arPerGB,
		RetransRatio: ratio, WastedBytes: wasted, TransportBytes: transported,
		LoadSeconds: loadSeconds,
	}
}

// Economics — valorisation complète d'un audit : tous les scénarios, le
// primaire MARQUÉ selon le type de lien déclaré (une correspondance de
// scénario, pas une affirmation sur le contrat réel du site).
type Economics struct {
	PrimaryLinkType  string              `json:"primary_link_type"`
	Primary          *EconomicsScenario  `json:"primary,omitempty"`
	Scenarios        []EconomicsScenario `json:"scenarios"`
	RetransRatio     RetransRatio        `json:"retrans_ratio"`
	WastedBytes      float64             `json:"wasted_bytes"`
	TransportedBytes float64             `json:"transported_bytes"`
	LoadSeconds      float64             `json:"load_seconds"`
	Disclaimer       string              `json:"disclaimer"`
}

func primaryName(linkType string) string {
	switch linkType {
	case "4g", "5g":
		return TariffCorporateAPN.Name
	case "vsat":
		return TariffWholesale.Name
	default: // fiber, other, vide
		return PlanGOLD.Name
	}
}

// PrimaryName expose la correspondance type de lien → scénario primaire :
// le catalogue API et le rapport d'audit lisent la même source.
func PrimaryName(linkType string) string { return primaryName(linkType) }

// EconomicsDisclaimer — le mot des scénarios contractuels, identique sur
// le rapport d'audit et le catalogue API (source unique).
const EconomicsDisclaimer = "Scénarios contractuels (yas.mg 31/08/2026 ; APN 150 Ar/Go sur devis), pas des mesures : le chiffre retenu dépend du contrat du site, pas de la physique du lien."

// BuildEconomics assemble tous les scénarios à partir du gaspillage mesuré.
// transportedBytes/wastedBytes en octets ; loadSeconds = fenêtre de charge
// mesurée (généralisation du ×20 de la directive).
func BuildEconomics(linkType string, ratio RetransRatio, transportedBytes, wastedBytes, loadSeconds float64) *Economics {
	e := &Economics{
		PrimaryLinkType:  linkType,
		RetransRatio:     ratio,
		WastedBytes:      wastedBytes,
		TransportedBytes: transportedBytes,
		LoadSeconds:      loadSeconds,
		Disclaimer:       EconomicsDisclaimer,
	}
	if ratio.Basis == "" || transportedBytes <= 0 {
		return e // non mesuré : aucun zéro fabriqué
	}
	// Modèle A : même ratio, deux plans dédiés.
	for _, p := range []EnterprisePlan{PlanGOLD, PlanPLATINE} {
		e.Scenarios = append(e.Scenarios, EconomicsScenario{
			Model: "A", Name: p.Name,
			ARPerHour: CapacityWasteARPerHour(p.MonthlyOpExMGA, wastedBytes, transportedBytes),
			Basis:     fmt.Sprintf("(part réémise %.4f) × %s Ar/720 h", ratio.Ratio, humanMGA(p.MonthlyOpExMGA)),
			Numbers:   numbers("A", p.MonthlyOpExMGA, p.CIRMbps, 0, ratio.Ratio, transportedBytes, wastedBytes, loadSeconds),
		})
	}
	// Modèle B : gaspillage facturé au volume, deux tarifs B2B.
	for _, tr := range []B2BTariff{TariffWholesale, TariffCorporateAPN} {
		e.Scenarios = append(e.Scenarios, EconomicsScenario{
			Model: "B", Name: tr.Name,
			ARPerHour: MeteredWasteARPerHour(wastedBytes, tr.ARPerGB, loadSeconds),
			Basis:     fmt.Sprintf("%.2f Mo gaspillés / fenêtre %.0f s × %.2f Ar/Go", wastedBytes/1e6, loadSeconds, tr.ARPerGB),
			Numbers:   numbers("B", 0, 0, tr.ARPerGB, ratio.Ratio, transportedBytes, wastedBytes, loadSeconds),
		})
	}
	// Pénalité de secours détail — visible, jamais primaire.
	e.Scenarios = append(e.Scenarios, EconomicsScenario{
		Model: "emergency", Name: "retail-outofbundle",
		ARPerHour: MeteredWasteARPerHour(wastedBytes, RetailEmergencyARPerGB, loadSeconds),
		Basis:     "hors-forfait d'urgence (Net Month 4,5 Go) — pénalité, pas un modèle d'infrastructure",
		Numbers:   numbers("emergency", 0, 0, RetailEmergencyARPerGB, ratio.Ratio, transportedBytes, wastedBytes, loadSeconds),
	})
	want := primaryName(linkType)
	for i := range e.Scenarios {
		if e.Scenarios[i].Name == want {
			prim := e.Scenarios[i]
			e.Primary = &prim
			break
		}
	}
	return e
}

// humanMGA — 1 000 000 → "1 000 000" (lisibilité des basis strings).
func humanMGA(v float64) string {
	s := strconv.FormatFloat(v, 'f', 0, 64)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	return string(out)
}
