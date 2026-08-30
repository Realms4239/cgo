// Package doctor — capability report (grilling Q11): one line per capability,
// la même vérité que le badge du tableau de bord. Rien ici ne
// ne mute l'hôte: lecture et rapport.
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | warn | fail
	Detail string `json:"detail"`
}

// Mode résout le mode effectif: override explicite, sinon l'OS décide —
// Windows observe (pas de tc/netem), Linux tourne à pleine puissance.
func Mode(override string) string {
	switch override {
	case "observe", "full":
		return override
	default: // auto
		if runtime.GOOS == "windows" {
			return "observe"
		}
		return "full"
	}
}

// Report liste chaque capacité requise pour le mode donné.
func Report(mode string) (string, []Check) {
	checks := []Check{
		{Name: "os", Status: "ok", Detail: runtime.GOOS + "/" + runtime.GOARCH + " → mode " + mode},
	}
	if p, err := exec.LookPath("tc"); err == nil {
		checks = append(checks, Check{Name: "tc", Status: "ok", Detail: p})
	} else {
		st := "warn"
		if mode == "full" {
			st = "fail"
		}
		checks = append(checks, Check{Name: "tc", Status: st, Detail: "iproute2 absent — façonnage impossible (kit/install.sh)"})
	}
	if runtime.GOOS == "windows" {
		checks = append(checks, Check{Name: "cap_net_admin", Status: "warn", Detail: "n/a — mode observation, aucun façonnage local"})
	} else if os.Geteuid() == 0 {
		checks = append(checks, Check{Name: "cap_net_admin", Status: "ok", Detail: "root"})
	} else {
		checks = append(checks, Check{Name: "cap_net_admin", Status: "warn", Detail: "sans CAP_NET_ADMIN le tc échouera — sudo setcap cap_net_admin+ep ./cgo"})
	}
	bbr := "indéterminé"
	if raw, err := os.ReadFile("/proc/sys/net/ipv4/tcp_available_congestion_control"); err == nil {
		if strings.Contains(string(raw), "bbr") {
			bbr = "disponible"
		} else {
			bbr = "absent — modprobe tcp_bbr (kit/install.sh)"
		}
	}
	st := "warn"
	if mode == "full" && strings.Contains(bbr, "absent") {
		st = "fail"
	}
	checks = append(checks, Check{Name: "bbr", Status: st, Detail: bbr})
	if _, err := exec.LookPath("ping"); err == nil {
		checks = append(checks, Check{Name: "ping", Status: "ok", Detail: "présent"})
	} else {
		checks = append(checks, Check{Name: "ping", Status: "fail", Detail: "absent — audit impossible"})
	}
	return mode, checks
}

// Print écrit le rapport lisible de `cgo doctor`.
func Print() {
	mode, checks := Report(Mode("auto"))
	fmt.Printf("mode: %s\n", mode)
	for _, c := range checks {
		fmt.Printf("  [%s] %s — %s\n", c.Status, c.Name, c.Detail)
	}
}
