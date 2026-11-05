//go:build !linux

package campagne

// Hors Linux, ni /proc ni les qdiscs : la campagne dualns n'y tourne pas.
func readHostCCList() []string      { return nil }
func readHostKernelRelease() string { return "" }
