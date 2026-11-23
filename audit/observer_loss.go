package audit

// observerLossDenominator — nom explicite du dénominateur du budget de perte
// d'observateur. Le pourcentage ne voyage jamais sans sa base déclarée.
const observerLossDenominator = "retrans_events + qdisc_drop_packets (compteurs globals, hors ring saturé)"

// ObserverLoss — réservations manquées / événements tentés. Le numérateur est
// ReserveFailures (déjà porté par KernelTelemetry depuis Task 2) : aucun compte
// dupliqué. Les compteurs d'événements viennent des globals, hors du flux
// saturé — un ring plein n'émet aucun record de perte, donc on ne déduit rien
// du nombre d'événements lus.
func ObserverLoss(attemptedEvents, reserveFailures uint64) (uint64, float64, string) {
	if attemptedEvents == 0 {
		// Zéro événement observé : on n'invente ni dénominateur numérique ni
		// pourcentage, mais le libellé de la base reste déclaré.
		return 0, 0, observerLossDenominator
	}
	return attemptedEvents, float64(reserveFailures) / float64(attemptedEvents) * 100, observerLossDenominator
}
