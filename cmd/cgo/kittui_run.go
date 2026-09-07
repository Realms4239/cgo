package main

// Moteur du centre de contrôle kit : boucle Update, actions en fond avec
// sortie en direct, suspensions TTY (keysetup, shell), rafraîchissements.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/internal/vm"
	tea "github.com/charmbracelet/bubbletea"
)

// ---- messages ----

type ktLog struct{ line string }
type ktDone struct {
	action string
	code   int
}
type ktVMs struct{ vms []vmEntry }
type ktDepsMsg struct{ rows []depRow }
type ktSSH struct {
	ok     bool
	detail string
}
type ktDash struct{ state string }
type ktResumed struct{}

// ---- writer vers le TUI ----

type ktWriter struct {
	prog *tea.Program
	buf  string
}

func (w *ktWriter) Write(p []byte) (int, error) {
	w.buf += string(p)
	for {
		i := strings.Index(w.buf, "\n")
		if i < 0 {
			break
		}
		line := strings.TrimRight(w.buf[:i], "\r")
		w.buf = w.buf[i+1:]
		if strings.TrimSpace(line) != "" {
			w.prog.Send(ktLog{line: line})
		}
	}
	return len(p), nil
}

// ---- lancement d'actions ----

// runBG — action kit en fond, sortie streamée dans le journal. Le modèle
// reste réactif (spinner, navigation) pendant les minutes de deploy/scan.
func (m *modelKT) runBG(action string, fn func(r *kit.Runner) int) {
	if m.prog == nil {
		return // hors runtime TUI (tests) : pas d'envoi possible
	}
	if m.busy != "" {
		m.pushLog("patience — « " + m.busy + " » tourne déjà")
		return
	}
	m.busy = action
	m.pushLog("▸ " + action + " …")
	go func() {
		r := kit.NewRunner()
		w := &ktWriter{prog: m.prog}
		r.Stdout, r.Stderr = w, w
		code := fn(r)
		if rest := strings.TrimSpace(w.buf); rest != "" {
			m.prog.Send(ktLog{line: rest})
		}
		m.prog.Send(ktDone{action: action, code: code})
	}()
}

// runSuspend — action INTERACTIVE (mot de passe, shell) : le TUI s'efface,
// le sous-processus hérite du terminal, puis le TUI reprend et rafraîchit.
func (m *modelKT) runSuspend(args ...string) tea.Cmd {
	return tea.ExecProcess(exec.Command(os.Args[0], args...), func(err error) tea.Msg {
		return ktResumed{}
	})
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// ---- rafraîchissements ----

func (m *modelKT) refreshDeps() {
	if m.prog == nil {
		return
	}
	var rows []depRow
	sshP, _ := exec.LookPath("ssh")
	scpP, _ := exec.LookPath("scp")
	if sshP != "" && scpP != "" {
		rows = append(rows, depRow{label: "Client OpenSSH", ok: true, info: sshP})
	} else {
		rows = append(rows, depRow{label: "Client OpenSSH", ok: false, info: "manquant — installable d'ici", fixID: "install-ssh"})
	}
	found := []string{}
	for _, h := range vm.Detect() {
		found = append(found, h.Name())
	}
	if len(found) > 0 {
		rows = append(rows, depRow{label: "Hyperviseur", ok: true, info: strings.Join(found, " + ")})
	} else {
		rows = append(rows, depRow{label: "Hyperviseur", ok: false, info: "ni vmrun ni VBoxManage — installez VMware/VirtualBox"})
	}
	if _, err := exec.LookPath("go"); err == nil {
		rows = append(rows, depRow{label: "Toolchain Go", ok: true, info: "recompile depuis sources possible"})
	} else {
		rows = append(rows, depRow{label: "Toolchain Go", ok: true, info: "absente — deploy précompilé OK"})
	}
	if _, err := os.Stat(m.cfgPath); err == nil {
		rows = append(rows, depRow{label: "Config", ok: true, info: m.cfgPath})
	} else {
		rows = append(rows, depRow{label: "Config", ok: true, info: "créée au verrouillage VM"})
	}
	m.prog.Send(ktDepsMsg{rows: rows})
}

func (m *modelKT) refreshVMs(deep bool) {
	if m.prog == nil {
		return
	}
	m.busy = "scan"
	m.pushLog("▸ scan des VMs …")
	go func() {
		paths := vm.ScanVMs(deep)
		running := map[string]bool{}
		for _, h := range vm.Detect() {
			for _, r := range h.Running() {
				running[r] = true
			}
		}
		var out []vmEntry
		for _, p := range paths {
			hyp := "?"
			lower := strings.ToLower(p)
			if strings.HasSuffix(lower, ".vmx") {
				hyp = "vmware"
			} else if strings.HasSuffix(lower, ".vbox") {
				hyp = "virtualbox"
			}
			live := running[p]
			if !live {
				for rp := range running {
					if strings.Contains(strings.ToLower(rp), strings.ToLower(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))) {
						live = true
					}
				}
			}
			out = append(out, vmEntry{path: p, name: strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)), hyp: hyp, live: live})
		}
		m.prog.Send(ktVMs{vms: out})
	}()
}

func (m *modelKT) refreshSSH() {
	if m.prog == nil {
		return
	}
	go func() {
		m.reloadCfg()
		if m.cfg.SSHUp() {
			m.prog.Send(ktSSH{ok: true})
			return
		}
		out, _ := m.cfg.SSH("true")
		m.prog.Send(ktSSH{ok: false, detail: shortSSHDiag(out)})
	}()
}

func shortSSHDiag(out string) string {
	o := strings.ToLower(out)
	switch {
	case strings.Contains(o, "refused"), strings.Contains(o, "closed"):
		return "port 22 fermé — openssh-server à installer DANS la VM"
	case strings.Contains(o, "permission denied"), strings.Contains(o, "denied"):
		return "clé refusée — posez la clé (ci-dessous)"
	case strings.Contains(o, "timed out"), strings.Contains(o, "timeout"), strings.Contains(o, "unreachable"), strings.Contains(o, "no route"):
		return "VM injoignable — allumée ? bonne IP ?"
	case strings.Contains(o, "no such file"), strings.Contains(o, "cannot"):
		return "clé locale introuvable — vérifiez ~/.ssh"
	default:
		if strings.TrimSpace(out) == "" {
			return "échec silencieux — démarrez la VM puis réessayez"
		}
		if len(out) > 90 {
			out = out[:90]
		}
		return strings.TrimSpace(out)
	}
}

func (m *modelKT) refreshDash() {
	if m.prog == nil {
		return
	}
	go func() {
		m.reloadCfg()
		if !m.cfg.Health() {
			m.prog.Send(ktDash{state: "ko"})
			return
		}
		m.prog.Send(ktDash{state: "ok"})
	}()
}

// ---- activation ----

func (m *modelKT) activate(id string) tea.Cmd {
	switch id {
	case "next":
		if m.step < ktControle {
			m.step++
			m.cursor = 0
		}
		if m.step == ktAcces {
			m.refreshSSH()
		}
		if m.step == ktDeploy || m.step == ktControle {
			m.refreshDash()
		}
		return nil
	case "recheck":
		m.refreshDeps()
		return nil
	case "install-ssh":
		m.runBG("install-ssh", func(r *kit.Runner) int {
			if runtime.GOOS == "windows" {
				c := exec.Command("winget", "install", "--id", "Microsoft.OpenSSH.Client", "--source", "winget", "--accept-package-agreements", "--accept-source-agreements")
				c.Stdout, c.Stderr = r.Stdout, r.Stderr
				if err := c.Run(); err != nil {
					return 2
				}
				return 0
			}
			c := exec.Command("sudo", "apt-get", "install", "-y", "openssh-client")
			c.Stdout, c.Stderr = r.Stdout, r.Stderr
			c.Stdin = os.Stdin
			if err := c.Run(); err != nil {
				return 2
			}
			return 0
		})
		return nil
	case "rescan":
		m.refreshVMs(true)
		return nil
	case "ensure":
		m.runBG("ensure", func(r *kit.Runner) int { return r.Ensure(m.cfg, m.cfgPath, true) })
		return nil
	case "retry":
		m.refreshSSH()
		return nil
	case "keysetup":
		return m.runSuspend("kit", "keysetup", "--config", m.cfgPath)
	case "deploy":
		m.runBG("deploy", func(r *kit.Runner) int { return r.Deploy(m.cfg, m.cfgPath, true) })
		return nil
	case "open":
		openBrowser(m.dashURL())
		m.pushLog("navigateur → " + m.dashURL())
		return nil
	case "switch":
		m.step = ktVM
		m.cursor = 0
		m.refreshVMs(false)
		return nil
	case "svc-status", "svc-start", "svc-stop", "svc-restart":
		sub := strings.TrimPrefix(id, "svc-")
		m.runBG("svc "+sub, func(r *kit.Runner) int { return r.Svc(m.cfg, sub) })
		return nil
	case "logs":
		m.runBG("logs", func(r *kit.Runner) int { return r.Logs(m.cfg, 40) })
		return nil
	case "dns":
		m.runBG("dns", func(r *kit.Runner) int { return r.DNS(m.cfg) })
		return nil
	case "tls":
		m.runBG("tls", func(r *kit.Runner) int { return r.TLS(m.cfg) })
		return nil
	case "verify":
		m.runBG("verify", func(r *kit.Runner) int { return r.Verify(m.cfg) })
		return nil
	case "backup":
		m.runBG("backup", func(r *kit.Runner) int { return r.Backup(m.cfg, "backup") })
		return nil
	case "snapshot":
		m.runBG("snapshot", func(r *kit.Runner) int {
			return r.Snapshot(m.cfg, fmt.Sprintf("cgo-tui-%s", time.Now().Format("20060102-150405")), true)
		})
		return nil
	}
	if strings.HasPrefix(id, "vm:") {
		var idx int
		fmt.Sscanf(id, "vm:%d", &idx)
		if idx >= 0 && idx < len(m.vms) {
			v := m.vms[idx]
			if err := kit.SaveVMX(m.cfgPath, v.path, v.hyp); err != nil {
				m.pushLog("verrouillage : " + err.Error())
				return nil
			}
			m.reloadCfg()
			m.pushLog("verrouillée : " + v.name + " (" + v.hyp + ")")
			m.step = ktAcces
			m.cursor = 0
			m.refreshSSH()
		}
		return nil
	}
	return nil
}

// ---- Update ----

func (m modelKT) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.busy != "" && msg.String() != "q" && msg.String() != "ctrl+c" {
			return m, nil // verrou doux pendant une action (q reste possible)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4", "5":
			m.step = ktStep(int(msg.String()[0] - '1'))
			m.cursor = 0
			if m.step == ktVM && len(m.vms) == 0 {
				m.refreshVMs(false)
			}
			if m.step == ktAcces {
				m.refreshSSH()
			}
			if m.step == ktDeploy || m.step == ktControle {
				m.refreshDash()
			}
			return m, nil
		case "j", "down":
			if items := m.items(); m.cursor < len(items)-1 {
				m.cursor++
			}
			return m, nil
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "enter", "o":
			if items := m.items(); m.cursor < len(items) {
				return m, m.activate(items[m.cursor].id)
			}
			return m, nil
		}
	case ktLog:
		m.pushLog(msg.line)
		return m, nil
	case ktDone:
		m.busy = ""
		if msg.code == 0 {
			m.pushLog("✓ " + msg.action + " terminé")
		} else {
			m.pushLog(fmt.Sprintf("✗ %s → code %d (voir ci-dessus)", msg.action, msg.code))
		}
		m.reloadCfg()
		switch msg.action {
		case "install-ssh", "ensure", "svc start", "svc stop", "svc restart", "dns", "tls", "svc status":
			m.refreshSSH()
			m.refreshDash()
		case "deploy":
			m.refreshSSH()
			m.refreshDash()
			m.step = ktControle
			m.cursor = 0
		}
		m.refreshDeps()
		return m, nil
	case ktVMs:
		m.busy = ""
		m.vms = msg.vms
		m.pushLog(fmt.Sprintf("%d VM(s) trouvée(s)", len(msg.vms)))
		return m, nil
	case ktDepsMsg:
		m.deps = msg.rows
		return m, nil
	case ktSSH:
		if msg.ok {
			m.sshState = "ok"
		} else {
			m.sshState = msg.detail
		}
		return m, nil
	case ktDash:
		m.dashState = msg.state
		return m, nil
	case ktResumed:
		m.reloadCfg()
		m.refreshDeps()
		m.refreshSSH()
		m.refreshDash()
		m.pushLog("retour au centre de contrôle")
		return m, nil
	}
	return m, nil
}

// runKitTUI — centre de contrôle kit : scan → verrou → clé → deploy → pilotage.
func runKitTUI(cfgPath, version string) int {
	m := initialModelKT(cfgPath, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.prog = p
	// partage du pointeur programme avec la copie du runtime : les goroutines
	// d'actions envoient leurs lignes via m.prog.Send.
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "kit tui exige un terminal interactif : "+err.Error())
		return 2
	}
	return 0
}
