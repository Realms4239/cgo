package campagne

import "os/exec"

// ensureQdiscModules précharge les disciplines du banc. Sans cela, la
// première cellule CAKE après un boot échoue en "Invalid qdisc name"
// (autoload manqué dans le netns) — vu en prod le 2026-09-05. Best-effort :
// sudo NOPASSWD couvre modprobe, et l'échec reste silencieux (ApplyShaper
// échoue proprement ensuite si vraiment absent).
func ensureQdiscModules() {
	_ = exec.Command("sudo", "-n", "modprobe", "sch_netem", "sch_fq_codel", "sch_cake").Run()
}
