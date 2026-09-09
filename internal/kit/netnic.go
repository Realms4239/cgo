package kit

import (
	"strings"
)

// Nic — lit ou bascule le mode réseau NIC1 de la VM verrouillée (nat =
// bench isolé + port-forward auto, ponté = IP directe sur le LAN).
// VM ÉTEINTE exigée pour écrire (les deux pilotes rejettent à chaud).
// Sans argument : affiche le mode actuel (lecture seule, sans allumer).
func (r *Runner) Nic(c *Config, cfgPath string, rest []string, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[nic] %v", err)
		return 3
	}
	mode := ""
	if len(rest) > 0 {
		mode = strings.ToLower(strings.TrimSpace(rest[0]))
	}
	if mode == "" {
		r.out("[nic] %s : %s", vmx, hyp.NetMode(vmx))
		return 0
	}
	switch mode {
	case "nat", "bridged", "hostonly", "pont", "ponte":
	default:
		r.errf("[nic] mode %q inconnu — nat|bridged|hostonly", rest[0])
		return 2
	}
	// alias français → valeurs pilote
	if mode == "pont" || mode == "ponte" {
		mode = "bridged"
	}
	if err := hyp.SetNetMode(vmx, mode); err != nil {
		r.errf("[nic] %v", err)
		return 4
	}
	if mode == "nat" {
		r.out("[nic] %s → nat — accès via port-forward auto (127.0.0.1:2222, :9090)", vmx)
	} else {
		r.out("[nic] %s → %s — accès DIRECT à l'IP invitée (redécouverte au prochain ensure)", vmx, mode)
	}
	return 0
}
