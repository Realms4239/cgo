//go:build !linux

package campagne

import "math"

// readCPUUsage — pas d'échantillonnage portable hors Linux : G7 reste non
// évaluée (nil) plutôt que de passer sur une valeur inventée.
func readCPUUsage() float64 { return math.NaN() }
