package kit

import (
	"strings"
)

// Testbed — banc de mesure invité (veth + netns + testbedsrv + sudoers) :
// `kit testbed up` le pose, `kit testbed down` le retire (idempotent, que
// du cgo-srv/veth-c/logs — rien d'autre), `kit testbed check` (défaut)
// l'audit. Sans lui, les campagnes tournent à vide (zéro ligne gelée) —
// la porte PlaneReady côté dashboard refuse alors Démarrer avec ce remède.
// Codes honnêtes partout : tout échec (scp, ssh, script invité) sort non-zéro.
func (r *Runner) Testbed(c *Config, rest []string) int {
	sub := "check"
	if len(rest) > 0 {
		sub = strings.ToLower(rest[0])
	}
	if sub != "up" && sub != "check" && sub != "down" {
		r.errf("[testbed] sous-commande inconnue : %s (up|check|down)", rest[0])
		return 2
	}
	// Le banc vit dans l'invitée : sans verrou ni utilisateur, scp/ssh
	// échoueraient en jargon — le chemin d'abord.
	if !r.requireLockUser("[testbed]", c) {
		return 2
	}
	local := findKitFile(r, "testbed.sh")
	if local == "" {
		r.errf("[testbed] testbed.sh introuvable ici — ré-extrayez l'archive complète (dossier kit/)")
		return 6
	}
	remote := c.ProjectDir + "/kit/testbed.sh"
	if out, _, err := c.SCPOut(local, remote); err != nil {
		r.errf("[testbed] scp ÉCHEC : %v — %s", err, strings.TrimSpace(out))
		return 7
	}
	_, _ = c.SSH("chmod +x " + shq(remote))
	out, err := c.SSH("bash " + shq(remote) + " " + sub)
	if err != nil {
		r.errf("[testbed] %s ÉCHEC sur l'invité", sub)
		if strings.TrimSpace(out) != "" {
			r.errf("[testbed] %s", strings.TrimSpace(out))
		}
		// Le retrait comme la pose peuvent exiger un sudo avec tty
		// (NOPASSWD pas encore en place) : la commande console exacte,
		// pas un code muet.
		if sub == "up" {
			r.errf("[testbed] → sur la console VM (une fois) : sudo bash kit/testbed.sh up")
		} else if sub == "down" {
			r.errf("[testbed] → sur la console VM : sudo bash kit/testbed.sh down")
		}
		return 4
	}
	switch sub {
	case "check":
		r.out("[testbed] banc prêt — les campagnes gèleront des lignes")
	case "up":
		r.out("[testbed] banc posé — campagnes possibles")
	default:
		r.out("[testbed] banc retiré")
	}
	return 0
}
