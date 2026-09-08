package kit

import (
	"strings"
)

// Guest — prépare TOUT côté invité via le script embarqué guest-setup.sh
// (poussé par scp, exécuté par ssh, sortie streamée à la fin) :
// openssh-server, outils invité, sshd, ufw 22+dashboard, binaire, dashboard
// TLS + santé. Idempotent (le script ne réinstalle que le manquant).
// check=true : audit lecture seule (--check), code 1 + liste si manque.
func (r *Runner) Guest(c *Config, check bool) int {
	if !r.ensureSSHClient() {
		return 2
	}
	if strings.TrimSpace(c.SSHHost) == "" {
		r.errf("[guest] hôte vide — verrouillez la VM / configurez ssh_host d'abord")
		return 2
	}
	script := findKitFile(r, "guest-setup.sh")
	if script == "" {
		r.errf("[guest] guest-setup.sh introuvable (kit/ du dépôt ou à côté de l'exe)")
		return 2
	}
	r.out("[guest] envoi du script…")
	if out, key, err := c.SCPOut(script, "/tmp/guest-setup.sh"); err != nil {
		r.errf("[guest] scp ÉCHEC : %v — clé=%s cible=%s@%s", err, key, c.SSHUser, c.SSHHost)
		r.errf("[guest] sortie scp : %s", strings.TrimSpace(out))
		return 7
	}
	args := "bash /tmp/guest-setup.sh"
	if check {
		args += " --check"
	} else {
		r.out("[guest] exécution côté invité (1-3 min, silence normal)…")
	}
	out, err := c.SSH("chmod +x /tmp/guest-setup.sh && " + args)
	for _, ln := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			r.out("[guest] │ %s", t)
		}
	}
	if err != nil {
		if check {
			r.errf("[guest] --check : points manquants ci-dessus — relancez sans --check")
			return 1
		}
		r.sshDiag("[guest]", out)
		return 6
	}
	if check {
		r.out("[guest] PRÊT — rien à faire")
	}
	return 0
}
