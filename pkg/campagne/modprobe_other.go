//go:build !linux

package campagne

// ensureQdiscModules n'existe que sous Linux (banc netem).
func ensureQdiscModules() {}
