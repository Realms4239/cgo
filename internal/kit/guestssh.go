package kit

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// GuestSSH — installe openssh-server DANS l'invité via les Tools de
// l'hyperviseur (runProgramInGuest / guestcontrol), SANS ssh préalable :
// c'est la sortie de secours quand le port 22 est fermé et que la console
// appartient à quelqu'un d'autre que vous. Mot de passe invité demandé une
// fois au terminal, jamais persisté. Additions/Tools requis côté invité.
func (r *Runner) GuestSSH(c *Config, cfgPath string, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[guest-ssh] %v", err)
		return 3
	}
	user := promptLine("utilisateur invité", c.SSHUser)
	if strings.TrimSpace(user) == "" {
		r.errf("[guest-ssh] utilisateur vide — abandon")
		return 2
	}
	pass, err := readPassword("mot de passe invité (transmis à la VM, jamais stocké) : ")
	if err != nil {
		r.errf("[guest-ssh] %v", err)
		return 2
	}
	if pass == "" {
		r.errf("[guest-ssh] mot de passe vide — abandon")
		return 2
	}
	// sudo sans mot de passe ? (banc NOPASSWD) — évite d'exposer le secret
	// dans une ligne de commande invitée quand ce n'est pas nécessaire.
	needPipe := true
	if out, err := hyp.RunGuest(user, pass, vmx, "sudo", "-n", "true"); err == nil {
		_ = out
		needPipe = false
		r.out("[guest-ssh] sudo sans mot de passe — installation directe")
	} else {
		r.out("[guest-ssh] sudo demandera le mot de passe — transmis via stdin invité")
	}
	apt := "sudo apt-get update && sudo apt-get install -y openssh-server && sudo systemctl enable --now ssh"
	var out string
	if needPipe {
		// '-' entre quotes simples, échappées à la POSIX
		script := fmt.Sprintf("echo '%s' | sudo -S apt-get update && echo '%s' | sudo -S apt-get install -y openssh-server && echo '%s' | sudo -S systemctl enable --now ssh",
			strings.ReplaceAll(pass, "'", `'\''`), strings.ReplaceAll(pass, "'", `'\''`), strings.ReplaceAll(pass, "'", `'\''`))
		out, err = hyp.RunGuest(user, pass, vmx, "/bin/bash", "-c", script)
	} else {
		out, err = hyp.RunGuest(user, pass, vmx, "/bin/bash", "-c", apt)
	}
	for _, ln := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			r.out("[guest-ssh] invité : %s", firstLine(t))
		}
	}
	if err != nil {
		detail := strings.TrimSpace(out)
		if detail == "" {
			detail = "sans détail — identifiants refusés ? Tools arrêtés ?"
		}
		r.errf("[guest-ssh] échec invité : %s", firstLine(detail))
		r.errf("[guest-ssh] console de secours DANS la VM : sudo apt install -y openssh-server && sudo systemctl enable --now ssh")
		return 8
	}
	// vérification : le port 22 répond-il depuis ici ?
	host := c.SSHHost
	if host == "" || host == "auto" {
		host = hyp.GuestIP(vmx)
	}
	if host != "" {
		d := net.Dialer{Timeout: 5 * time.Second}
		if cn, err := d.Dial("tcp", net.JoinHostPort(host, "22")); err == nil {
			_ = cn.Close()
			r.out("[guest-ssh] port 22 OUVERT sur %s — enchaînez : kit keysetup", host)
			_ = saveConfigValue(cfgPath, "ssh_user", user)
			return 0
		}
		r.out("[guest-ssh] installé, mais %s:22 injoignable d'ici (NAT ?) — relancez : kit ensure", host)
		return 0
	}
	r.out("[guest-ssh] installé — IP non résolue : relancez : kit ensure")
	return 0
}
