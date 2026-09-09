package kit

import (
	"fmt"
	"strings"
)

// ScheduleUnits — générateur des unités systemd du timer nightly : la
// campagne filtrée se rejoue toute seule sur la VM (NetDevOps : la dérive
// temporelle se lit dans delta sans intervention). Rend (service, timer).
// Un serveur HTTP ne se cronifie pas lui-même : le timer est versionnable
// et auditable côté banc, la campagne tourne dans son répertoire de gels.
func ScheduleUnits(projectDir, at, profiles, qdiscs, ccs, direction string, reps int) (string, string) {
	args := []string{"run", "--profiles", profiles}
	if qdiscs != "" {
		args = append(args, "--qdiscs", qdiscs)
	}
	if ccs != "" {
		args = append(args, "--cc", ccs)
	}
	if direction != "" {
		args = append(args, "--direction", direction)
	}
	if reps <= 0 {
		reps = 3 // défaut matrice : l'API refuse reps=0, jamais de flag vide
	}
	args = append(args, "--reps", fmt.Sprint(reps))
	cmd := "cgo-linux " + strings.Join(args, " ")
	service := fmt.Sprintf(`[Unit]
Description=Meteolink — campagne nocturne (dérive temporelle)
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
WorkingDirectory=%s
ExecStart=/usr/local/bin/%s
`, projectDir, cmd)
	timer := fmt.Sprintf(`[Unit]
Description=Meteolink — déclencheur nocturne de campagne

[Timer]
OnCalendar=*-*-* %s:00
Persistent=true

[Install]
WantedBy=timers.target
`, at)
	return service, timer
}

// Schedule — pose les unités sur la VM et active le timer (idempotent).
// Utilise c.SSH + ssh-cat stdin (mêmes primitives que Deploy/Backup).
func (r *Runner) Schedule(c *Config, at, profiles, qdiscs, ccs, direction string, reps int) int {
	service, timer := ScheduleUnits(c.ProjectDir, at, profiles, qdiscs, ccs, direction, reps)
	for name, content := range map[string]string{"meteolink-campaign.service": service, "meteolink-campaign.timer": timer} {
		// via sshCmd (pas de ssh artisanal) : BatchMode + timeouts + mux
		// unifiés — un ssh nu ici pendait le schedule sans limite.
		base := c.sshCmd()
		cat := bgCmd(base[0], append(base[1:], "cat > /tmp/"+name)...)
		cat.Stdin = strings.NewReader(content)
		if out, err := cat.CombinedOutput(); err != nil {
			r.errf("[schedule] écriture %s: %v (%s)", name, err, string(out))
			return 7
		}
		if out, err := c.SSH("sudo mv /tmp/" + name + " /etc/systemd/system/" + name); err != nil {
			r.errf("[schedule] installation %s: %v (%s)", name, err, out)
			return r.scheduleFallback(c, at, profiles, qdiscs, ccs, direction, reps)
		}
	}
	if out, err := c.SSH("sudo systemctl daemon-reload && sudo systemctl enable --now meteolink-campaign.timer"); err != nil {
		// sans privilège (sudo interactif) : repli documenté — cron
		// utilisateur si présent, sinon la commande root exacte à jouer
		r.errf("[schedule] activation: %v (%s)", err, out)
		return r.scheduleFallback(c, at, profiles, qdiscs, ccs, direction, reps)
	}
	r.out("[schedule] timer nocturne posé — %s, campagne %s%s", at, profiles, extraSuffix(qdiscs, ccs, direction))
	return 0
}

// scheduleFallback — pas de root : cron utilisateur si le binaire existe,
// sinon on imprime la commande root minimale (une seule, à jouer une fois).
func (r *Runner) scheduleFallback(c *Config, at, profiles, qdiscs, ccs, direction string, reps int) int {
	if out, err := c.SSH("command -v crontab"); err == nil && strings.TrimSpace(out) != "" {
		line := ScheduleCronLine(c.ProjectDir, at, profiles, qdiscs, ccs, direction, reps)
		if _, err := c.SSH(fmt.Sprintf("(crontab -l 2>/dev/null | grep -v meteolink-campaign; echo %q) | crontab -", line)); err != nil {
			r.errf("[schedule] crontab: %v", err)
			return 7
		}
		r.out("[schedule] cron utilisateur posé (repli sans root) — %s", line)
		return 0
	}
	r.out("[schedule] ni root ni cron sur le banc — jouer UNE fois en root :")
	r.out("  sudo apt-get install -y cron && sudo systemctl enable --now cron")
	r.out("puis relancer : cgo kit schedule --at %s --profiles %s", at, profiles)
	return 7
}

// ScheduleCronLine — repli sans privilège : cron utilisateur (pas de sudo,
// survit au logout si cron tourne). Rend la ligne crontab complète.
func ScheduleCronLine(projectDir, at, profiles, qdiscs, ccs, direction string, reps int) string {
	var hh, mm string
	if _, err := fmt.Sscanf(at, "%d:%d", &hh, &mm); err != nil {
		hh, mm = "02", "30"
	}
	args := []string{"/usr/local/bin/cgo-linux", "run", "--profiles", profiles}
	if qdiscs != "" {
		args = append(args, "--qdiscs", qdiscs)
	}
	if ccs != "" {
		args = append(args, "--cc", ccs)
	}
	if direction != "" {
		args = append(args, "--direction", direction)
	}
	if reps <= 0 {
		reps = 3
	}
	args = append(args, "--reps", fmt.Sprint(reps))
	return fmt.Sprintf("%s %s * * * cd %s && %s >>%s/campaign-cron.log 2>&1",
		mm, hh, projectDir, strings.Join(args, " "), projectDir)
}

func extraSuffix(qdiscs, ccs, direction string) string {
	var parts []string
	if qdiscs != "" {
		parts = append(parts, "qdiscs="+qdiscs)
	}
	if ccs != "" {
		parts = append(parts, "cc="+ccs)
	}
	if direction != "" {
		parts = append(parts, "direction="+direction)
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, " ") + ")"
}
