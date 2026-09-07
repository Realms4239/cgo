package kit

import (
	"strings"
)

// Netinfo — le réseau de l'invité, en lecture seule, sans privilège :
// adresses live, route par défaut, DNS. Aucune écriture, aucun risque —
// c'est la vérité terrain avant toute décision d'IP (fixe ? DHCP ?).
func (r *Runner) Netinfo(c *Config) int {
	if c.SSHHost == "" {
		r.errf("[netinfo] hôte vide — verrouillez la VM / configurez la cible d'abord")
		return 2
	}
	out, err := c.SSH("ip -br addr show; echo ---; ip route show default; echo ---; grep -h nameserver /etc/resolv.conf 2>/dev/null || true")
	if err != nil {
		r.sshDiag("[netinfo]", out)
		return 5
	}
	r.out("[netinfo] %s :", c.SSHHost)
	for _, ln := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			r.out("  %s", t)
		}
	}
	return 0
}
