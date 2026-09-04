package kit

import (
	"fmt"
	"os/exec"
	"strings"
)

// ScheduleUnits — générateur des unités systemd du timer nightly : la
// campagne filtrée se rejoue toute seule sur la VM (NetDevOps : la dérive
// temporelle se lit dans delta sans intervention). Rend (service, timer).
// Un serveur HTTP ne se cronifie pas lui-même : le timer est versionnable
// et auditable côté banc, la campagne tourne dans son répertoire de gels.
func ScheduleUnits(at, profiles, qdiscs, ccs, direction string, reps int) (string, string) {
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
WorkingDirectory=/home/altfloat/cgo
ExecStart=/usr/local/bin/%s
`, cmd)
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
	service, timer := ScheduleUnits(at, profiles, qdiscs, ccs, direction, reps)
	for name, content := range map[string]string{"meteolink-campaign.service": service, "meteolink-campaign.timer": timer} {
		cat := exec.Command("ssh", "-o", "ConnectTimeout=6", "-o", "StrictHostKeyChecking=accept-new",
			"-p", c.SSHPort, "-i", expandKey(c.SSHKey),
			c.SSHUser+"@"+c.SSHHost, "cat > /tmp/"+name)
		cat.Stdin = strings.NewReader(content)
		if out, err := cat.CombinedOutput(); err != nil {
			r.errf("[schedule] écriture %s: %v (%s)", name, err, string(out))
			return 7
		}
		if out, err := c.SSH("sudo mv /tmp/" + name + " /etc/systemd/system/" + name); err != nil {
			r.errf("[schedule] installation %s: %v (%s)", name, err, out)
			return 7
		}
	}
	if out, err := c.SSH("sudo systemctl daemon-reload && sudo systemctl enable --now meteolink-campaign.timer"); err != nil {
		r.errf("[schedule] activation: %v (%s)", err, out)
		return 7
	}
	r.out("[schedule] timer nocturne posé — %s, campagne %s%s", at, profiles, extraSuffix(qdiscs, ccs, direction))
	return 0
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
