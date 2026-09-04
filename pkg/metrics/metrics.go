// Package metrics — percentile, deadline et coût.
// Le navigateur ne recalcule rien; Go est le seul calculateur.
package metrics

import (
	"math"
	"sort"
)

// Percentile of a pre-sorted slice (linear interpolation, nearest-rank-ish).
func Percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	switch n {
	case 0:
		return 0
	case 1:
		return sorted[0]
	}
	rank := (p / 100) * float64(n-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	frac := rank - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

type Summary struct {
	P50, P95, Median float64
	IQRLow, IQRHigh  float64
}

// Summarize trie une copie et rend médiane + IQR + percentiles.
func Summarize(v []float64) Summary {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return Summary{
		Median:  Percentile(s, 50),
		P50:     Percentile(s, 50),
		P95:     Percentile(s, 95),
		IQRLow:  Percentile(s, 25),
		IQRHigh: Percentile(s, 75),
	}
}

// DeadlineOKPct — share of completions strictly under dueMs.
func DeadlineOKPct(completionsMs []float64, dueMs float64) float64 {
	if len(completionsMs) == 0 {
		return 0
	}
	ok := 0
	for _, v := range completionsMs {
		if v < dueMs {
			ok++
		}
	}
	return float64(ok) / float64(len(completionsMs)) * 100
}

// Coût du gaspillage — modèle à paliers opérateurs réels (recherche 2026,
// yas.mg officiel) : le prix à l'octet dépend du forfait où vit
// l'institution. La fibre est ~10× moins chère que le mobile au Go.
//
//	palier           Ar/Go    source
//	Yas Net Day 1 Go  1 000    yas.mg (1000 Ar/24h)
//	Yas Net Month 4,5 5 556    yas.mg (25 000 Ar) — Airtel idem 4,5 Go
//	Yas Net 100 Go    2 000    yas.mg (200 000 Ar)
//	Yas FTTH 100 Go     490    yas.mg (49 000 Ar/mo, hors fibre dédiée)
//
// Default = mobile mensuel 4,5 Go (contexte DSI : liens cellulaires de
// secours) ; l'API expose le choix de palier pour ne pas mentir.
const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

// PriceTier — palier de forfait observé (Ar par Go).
type PriceTier struct {
	Name  string  `json:"name"`
	ARGB  float64 `json:"ar_per_gb"`
	Bytes float64 `json:"bundle_bytes"`
}

// Tiers — tarifs réels des trois opérateurs, mobile + fibre.
var Tiers = []PriceTier{
	{Name: "yas-day-1gb", ARGB: 1000, Bytes: 1 * GB},
	{Name: "yas-month-4.5gb", ARGB: 5556, Bytes: 4.5 * GB},
	{Name: "airtel-month-4.5gb", ARGB: 5556, Bytes: 4.5 * GB},
	{Name: "yas-month-100gb", ARGB: 2000, Bytes: 100 * GB},
	{Name: "yas-ftth-100gb", ARGB: 490, Bytes: 100 * GB},
	{Name: "orange-month-5gb", ARGB: 2000, Bytes: 5 * GB},
}

// DefaultTier — le palier par défaut du calcul.
var DefaultTier = Tiers[1] // yas-month-4.5gb : contexte cellular DSI

// CostARPerH — coût horaire du gaspillage au palier courant.
// wasted × (Ar/Go) / Go = Ar, l'heure vient de la fenêtre d'événement (3 min
// extrapolée ×20 — documentée dans la méthodologie, pas une facturation).
func CostARPerH(wastedBytes uint64) float64 {
	return CostARPerHTier(wastedBytes, DefaultTier)
}

// CostARPerHTier — même calcul, palier explicite (sélectionné via l'API).
func CostARPerHTier(wastedBytes uint64, tier PriceTier) float64 {
	ar := float64(wastedBytes) / GB * tier.ARGB
	return ar * 20 // fenêtre 3 min → heure
}

// JFI — Jain's fairness index (Σx)² / (n·Σx²), 0..1 (1 = perfectly fair).
// Inutilisé tant que l'API n'expose pas les valeurs par répétition (computeJFI côté front).

// JFI — Jain's fairness index (Σx)² / (n·Σx²), 0..1 (1 = perfectly fair).
// Per-flow (contributions individuelles), jamais sur le total seul : un
// total identique peut cacher la famine d'un flux.
func JFI(xs []uint64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var sum, sumSq float64
	for _, x := range xs {
		sum += float64(x)
		sumSq += float64(x) * float64(x)
	}
	if sumSq == 0 {
		return 0
	}
	return sum * sum / (float64(len(xs)) * sumSq)
}

// VoIPR — score R du E-model simplifié (ITU-T G.107) sur les proxys mesurés :
// délai (rtt/2 aller simple), gigue (p95−p50, absorbée par le jitter buffer),
// perte (pct). 0 (insupportable) → 100 (excellent). Idl > 50 = MOS > 3.6
// "utilisable" ; Idl > 80 = très bon. Version simplifiée assumée : sans flux
// UDP de référence, on applique le modèle aux conditions observées — la
// méthode (pas les flux) est celle que prévoyait CONGESTION.md pour E3.
//
//	Id(élai) = Id(0) + 0.024·d + 0.11·(d−177.3)/[d−177.3+ε]  (d en ms, one-way)
//	Ie(perte) = γ(1−ln(1−pct/100))                            (γ≈30 pour G.711)
//	Ij(gigue) ≈ 0.024·jitter  (approx. standard du de-serialiser)
//	R = 93.2 − Id − Ie − Ij   (93.2 = base pour G.711, niveau ≈ 0)
func VoIPR(oneWayDelayMs, jitterMs, lossPct float64) float64 {
	// délai aller simple depuis le RTT mesurés ; le modèle ITU devient sévère
	// au-delà du seuil magique 177,3 ms (le dénominateur d−177,3)
	d := oneWayDelayMs
	if d > 400 {
		d = 400
	}
	id := 0.024 * d
	if d > 177.3 {
		id += 0.11 * (d - 177.3)
	} else {
		id += 0.11 * d / 50 // zone tolérante sous le seuil
	}
	// effet de la gigue : le buffer absorbe mais décale — approximation standard
	ij := 0.024 * jitterMs
	// perte : impairment Equipment du G.711, Ie = γ·ln(1/(1−p)) = −γ·ln(1−p) ;
	// math.Log direct (garde p→1 : perte 100 % = lien mort, pas -Inf gelé)
	ie := 0.0
	if p := 1 - lossPct/100; p > 0 {
		ie = -30.0 * math.Log(p)
	}
	r := 93.2 - id - ie - ij
	if r < 0 {
		return 0
	}
	if r > 100 {
		return 100
	}
	return r
}
