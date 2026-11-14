package campagne

import "os/exec"

// ensureQdiscModules précharge les disciplines du banc. Sans cela, la
// première cellule CAKE après un boot échoue en "Invalid qdisc name"
// (autoload manqué dans le netns) — vu en prod le 2026-09-05. Best-effort :
// cgo-netctl (NOPASSWD) couvre EXACTEMENT ces trois modules, et l'échec
// reste silencieux (ApplyShaper échoue proprement ensuite si vraiment
// absent). Un modprobe par module : kmod prend UN seul module par appel —
// l'ancien `modprobe sch_netem sch_fq_codel sch_cake` passait les suivants
// comme paramètres de sch_netem (no-op silencieux).
func ensureQdiscModules() {
	for _, m := range []string{"sch_netem", "sch_fq_codel", "sch_cake"} {
		_ = exec.Command("sudo", "-n", "/usr/local/bin/cgo-netctl", "modprobe", m).Run()
	}
}
