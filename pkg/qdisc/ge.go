package qdisc

import "fmt"

// GEModel — chaîne de Markov à 2 états de netem (Gilbert-Elliott), dans la
// sémantique de l'article des implémenteurs (Carry & Briscoe, « Definition
// of a general and intuitive loss model for packet networks and its
// implementation in the Linux kernel Netem ») : la forme documentée est
// « loss gemodel p r 1-h 1-k » avec
//
//	p     = proba de transition bon → mauvais
//	r     = proba de transition mauvais → bon
//	1-h   = proba de perte DANS l'état mauvais
//	1-k   = proba de perte dans l'état bon
//
// Valeurs en décimales (0–1), comme l'exemple travaillé de l'article
// (0.01 0.19 0.81 0.01 → 5 % de perte moyenne en salves). L'ancienne forme
// à CINQ arguments (p r h 1-r k) était invalide sur Linux — corrigée ici.
type GEModel struct {
	GoodToBad float64 `json:"good_to_bad"` // p
	BadToGood float64 `json:"bad_to_good"` // r
	LossBad   float64 `json:"loss_bad"`    // 1-h : perte en état mauvais
	LossGood  float64 `json:"loss_good"`   // 1-k : perte en état bon
}

// GEMeanLoss — perte moyenne théorique du modèle, en POURCENT (0–100) :
// π_mauvais = p/(p+r) ; perte = π·(1-h) + (1−π)·(1-k). Modèle vide ⇒ 0.
// C'est la fonction qui garde les paramètres du banc honnêtes : la cible
// P2 (≈0,544 %) et P3 (≈1,08 %) se VÉRIFIE, elle ne se déclare pas.
func GEMeanLoss(ge GEModel) float64 {
	sum := ge.GoodToBad + ge.BadToGood
	if sum <= 0 {
		return 0
	}
	piBad := ge.GoodToBad / sum
	return (piBad*ge.LossBad + (1-piBad)*ge.LossGood) * 100
}

// ApplyNetemGE — netem racine avec perte Gilbert-Elliott (forme documentée
// à 4 arguments). Handle 1: conservé (un shaper peut s'empiler en 1:1).
func ApplyNetemGE(r TCRunner, iface string, delayMs, jitterMs float64, ge GEModel) error {
	args := []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs),
		"loss", "gemodel",
		fmt.Sprintf("%g", ge.GoodToBad),
		fmt.Sprintf("%g", ge.BadToGood),
		fmt.Sprintf("%g", ge.LossBad),
		fmt.Sprintf("%g", ge.LossGood)}
	if _, err := r.Run(args...); err != nil {
		return fmt.Errorf("netem ge %s: %w", iface, err)
	}
	return nil
}

// ApplyNetemPhysics — le netem du saut de propagation : délai + gigue, et
// perte UNIFORME (ge nil, forme historique à l'octet près — le corpus gelé
// se reproduit par là) ou Gilbert-Elliott (ge fourni, 4 arguments). Une
// seule des deux formes de perte, jamais les deux.
func ApplyNetemPhysics(r TCRunner, iface string, delayMs, jitterMs, lossPct float64, ge *GEModel) error {
	if ge != nil {
		return ApplyNetemGE(r, iface, delayMs, jitterMs, *ge)
	}
	args := []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "netem",
		"delay", fmt.Sprintf("%gms", delayMs), fmt.Sprintf("%gms", jitterMs)}
	if lossPct > 0 {
		args = append(args, "loss", fmt.Sprintf("%g%%", lossPct))
	}
	if _, err := r.Run(args...); err != nil {
		return fmt.Errorf("netem %s: %w", iface, err)
	}
	return nil
}
