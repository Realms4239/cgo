package kit

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// KeySetup — pose la clé publique de l'hôte dans authorized_keys de la cible,
// sans jamais toucher l'interface de l'hyperviseur ni voir le mot de passe :
// l'utilisateur le tape dans l'invite du client ssh lui-même (stdio attaché
// au terminal) — le secret ne transite ni par notre mémoire, ni par nos logs,
// ni par la config. Zéro dépendance ajoutée (ssh système uniquement).
//
// Séquence : clé déjà acceptée ? (idempotent, sortie 0) → prompts
// utilisateur/hôte/port (défauts = config, validés) → résumé du plan +
// confirmation explicite OUI → ssh par mot de passe → vérification par la
// clé (SSHUp) → ssh_user mémorisé dans le yaml.
func (r *Runner) KeySetup(c *Config, cfgPath string, rest []string) int {
	if !r.ensureSSHClient() {
		return 2
	}
	host, user, port := c.SSHHost, c.SSHUser, c.SSHPort
	for i := 0; i+1 < len(rest); i += 2 {
		switch rest[i] {
		case "--host":
			host = rest[i+1]
		case "--user":
			user = rest[i+1]
		case "--port":
			port = rest[i+1]
		}
	}
	if !isTerminal() {
		r.errf("[keysetup] entrée non interactive — relancer dans un terminal")
		return 2
	}
	pub, err := resolvePubkey(c.SSHKey)
	if err != nil {
		r.errf("[keysetup] %v", err)
		return 2
	}
	user = promptLine("utilisateur distant", user)
	host = promptLine("hôte distant", host)
	port = promptLine("port SSH", port)
	if err := validateKeySetupTarget(user, host, port); err != nil {
		r.errf("[keysetup] invalide : %v", err)
		return 2
	}
	// idempotent : la clé passe déjà, rien à poser.
	probe := *c
	probe.SSHUser, probe.SSHHost, probe.SSHPort = user, host, port
	if probe.SSHUp() {
		r.out("[keysetup] clé déjà acceptée par %s@%s — rien à faire", user, host)
		return 0
	}
	if !portOpen(host, port) {
		r.errf("[keysetup] %s:%s fermé — sshd absent ou machine éteinte", host, port)
		return 5
	}
	keyLine, err := os.ReadFile(pub)
	if err != nil {
		r.errf("[keysetup] lecture %s : %v", pub, err)
		return 2
	}
	keyLineStr := strings.TrimSpace(string(keyLine))
	r.out("[keysetup] plan :")
	r.out("  clé locale  %s", pub)
	r.out("  cible       %s@%s:%s (~/.ssh/authorized_keys, sans doublon)", user, host, port)
	r.out("  mot de passe : tapé dans l'invite ssh ci-dessous (jamais stocké)")
	if promptLine("taper OUI pour confirmer", "") != "OUI" {
		r.errf("[keysetup] annulé — rien n'a été touché")
		return 2
	}
	// mot de passe demandé par ssh lui-même sur le terminal.
	remote := fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && touch ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && grep -qxF -- '%s' ~/.ssh/authorized_keys || echo '%s' >> ~/.ssh/authorized_keys", keyLineStr, keyLineStr)
	cmd := exec.Command("ssh", "-o", "ConnectTimeout=8", "-o", "PreferredAuthentications=keyboard-interactive,password",
		"-o", "PubkeyAuthentication=no", "-o", "StrictHostKeyChecking=accept-new",
		"-p", port, user+"@"+host, remote)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, r.Stdout, r.Stderr
	if err := cmd.Run(); err != nil {
		r.errf("[keysetup] pose ÉCHEC : %v (mot de passe ? sshd refuse password auth ?)", err)
		return 5
	}
	if !probe.SSHUp() {
		r.errf("[keysetup] la clé ne passe toujours pas après la pose — authorized_keys à vérifier côté cible")
		return 5
	}
	if user != c.SSHUser {
		_ = saveConfigValue(cfgPath, "ssh_user", user)
	}
	r.out("[keysetup] ok — %s@%s accepte la clé", user, host)
	return 0
}

// validateKeySetupTarget — garde-fous purs (testés) : pas de cible vide,
// utilisateur shell-safe, port numérique 1–65535.
func validateKeySetupTarget(user, host, port string) error {
	if strings.TrimSpace(user) == "" {
		return fmt.Errorf("utilisateur vide")
	}
	if strings.ContainsAny(user, " \t:@\n\"'\\$`") {
		return fmt.Errorf("utilisateur %q : caractères interdits (espace : @ \" ' \\ $ `)", user)
	}
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("hôte vide")
	}
	if strings.ContainsAny(host, " \t\n\"'") {
		return fmt.Errorf("hôte %q : espaces ou quotes interdits", host)
	}
	n, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("port %q : 1–65535 attendu", port)
	}
	return nil
}

// resolvePubkey — chemin clé privée → .pub existante au format OpenSSH
// (type base64 [commentaire]). Refus explicite sinon : pas de clé
// silencieusement absente.
func resolvePubkey(key string) (string, error) {
	if strings.HasPrefix(key, "~/") {
		if h := userHome(); h != "" {
			key = filepath.Join(h, key[2:])
		}
	}
	pub := key + ".pub"
	raw, err := os.ReadFile(pub)
	if err != nil {
		return "", fmt.Errorf("clé publique %s introuvable (ssh-keygen -t ed25519 d'abord)", pub)
	}
	f := strings.Fields(strings.TrimSpace(string(raw)))
	if len(f) < 2 || !(f[0] == "ssh-ed25519" || f[0] == "ssh-rsa" || f[0] == "ecdsa-sha2-nistp256" || f[0] == "ecdsa-sha2-nistp384" || f[0] == "ecdsa-sha2-nistp521" || f[0] == "ssh-dss") {
		return "", fmt.Errorf("clé publique %s : format OpenSSH attendu", pub)
	}
	return pub, nil
}

// promptLine — invite terminal avec défaut entre crochets ; Entrée vide =
// défaut. Pas de secret ici (utilisateur/hôte/port uniquement).
func promptLine(label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s] : ", label, def)
	} else {
		fmt.Printf("%s : ", label)
	}
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// isTerminal — stdin est un terminal (pas un pipe) : les prompts ont un sens.
func isTerminal() bool {
	st, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
