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

	"github.com/Realms4239/cgo/model"
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
	checks = append(checks, ccChecks(mode)...)
	checks = append(checks, benchOffloadCheck())
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

// ParseOffloadFeatures lit la sortie « ethtool -k » : lignes
// « feature: state » où state ∈ {on, off} suffixé facultativement de
// [fixed]/[requested on]. Rend feature → état actif (true = on). Les
// en-têtes et lignes non « feature: state » sont ignorées : le parseur
// sert la porte bench-offloads, il ne devine rien.
func ParseOffloadFeatures(out string) map[string]bool {
	m := map[string]bool{}
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		i := strings.Index(ln, ":")
		if i <= 0 {
			continue
		}
		name := strings.TrimSpace(ln[:i])
		state := strings.TrimSpace(ln[i+1:])
		if name == "" || state == "" {
			continue
		}
		// suffixes [fixed], [requested on]… : l'état nominal reste le premier mot.
		if j := strings.Index(state, " "); j > 0 {
			state = state[:j]
		}
		switch state {
		case "on":
			m[name] = true
		case "off":
			m[name] = false
		}
	}
	return m
}

// benchOffloadPairs — les six interfaces de transit du banc durci
// (kit/cgo-bench-dualns.sh), nom-namespace → interface.
var benchOffloadPairs = []struct{ Ns, Iface string }{
	{"ns_client", "c-cli"},
	{"ns_wan_emulator", "c-wan"},
	{"ns_wan_emulator", "w-edge"},
	{"ns_edge_router", "e-wan"},
	{"ns_edge_router", "e-srv"},
	{"ns_server", "s-edge"},
}

// benchOffloadCheck — la porte honnête du banc : fail si GRO/GSO/TSO actif
// sur une interface de transit (un banc qui ment fausse chaque mesure),
// n/a (warn) si le banc n'est pas monté, ok les six vérifiées off.
// Lecture seule : jamais de mutation de l'hôte.
func benchOffloadCheck() Check {
	if runtime.GOOS != "linux" {
		return Check{Name: "bench-offloads", Status: "warn", Detail: "n/a — hors Linux (mode observation)"}
	}
	if out, err := exec.Command("ip", "netns", "list").Output(); err != nil || len(out) == 0 {
		return Check{Name: "bench-offloads", Status: "warn", Detail: "n/a — banc absent (kit/run-dualns.sh up)"}
	}
	// Le banc est monté : CHAQUE interface de transit doit relire off.
	for _, p := range benchOffloadPairs {
		out, err := exec.Command("sudo", "-n", "ip", "netns", "exec", p.Ns,
			"ethtool", "-k", p.Iface).Output()
		if err != nil {
			return Check{Name: "bench-offloads", Status: "warn",
				Detail: "n/a — ethtool illisible sur " + p.Ns + "/" + p.Iface}
		}
		st := ParseOffloadFeatures(string(out))
		for _, feat := range []string{"tcp-segmentation-offload", "generic-segmentation-offload", "generic-receive-offload"} {
			if st[feat] {
				return Check{Name: "bench-offloads", Status: "fail",
					Detail: feat + " ACTIF sur " + p.Ns + "/" + p.Iface + " — super-paquets possibles, mesures faussées (kit/cgo-bench-dualns.sh up réapplique)"}
			}
		}
	}
	return Check{Name: "bench-offloads", Status: "ok", Detail: "6 interfaces de transit relues off (GRO/GSO/TSO)"}
}

// ccChecks — deux lignes distinctes : « bbr » (v1, mainline) et « bbr3 »
// (noyau BBRv3). Le token exact, jamais le préfixe : contains("bbr")
// déclarait bbr disponible sur un noyau qui n'expose que bbr3.
func ccChecks(mode string) []Check {
	raw, err := os.ReadFile("/proc/sys/net/ipv4/tcp_available_congestion_control")
	if err != nil {
		return ccChecksFrom(mode, nil, false)
	}
	return ccChecksFrom(mode, model.ParseCCList(string(raw)), true)
}

func ccChecksFrom(mode string, list []string, readable bool) []Check {
	if !readable {
		return []Check{
			{Name: "bbr", Status: "warn", Detail: "indéterminé — /proc/sys/net/ipv4/tcp_available_congestion_control illisible"},
			{Name: "bbr3", Status: "warn", Detail: "indéterminé — même fichier illisible"},
		}
	}
	present := map[string]bool{}
	for _, c := range list {
		present[c] = true
	}
	bbr := Check{Name: "bbr", Status: "ok", Detail: "disponible (BBRv1, mainline)"}
	if !present["bbr"] {
		bbr = Check{Name: "bbr", Status: "warn", Detail: "absent — modprobe tcp_bbr (kit/install.sh)"}
		if mode == "full" {
			bbr.Status = "fail"
		}
	}
	bbr3 := Check{Name: "bbr3", Status: "warn", Detail: "absent — optionnel : kit/install-bbr3.sh (noyau BBRv3, 7.1)"}
	if present["bbr3"] {
		bbr3 = Check{Name: "bbr3", Status: "ok", Detail: "disponible (BBRv3)"}
	}
	return []Check{bbr, bbr3}
}
