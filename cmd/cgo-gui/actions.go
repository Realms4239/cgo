//go:build windows

package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/internal/vm"
)

// ---- journal thread-safe ----

func (a *app) appendLog(line string) {
	a.mu.Lock()
	a.logText = append(a.logText, line)
	if len(a.logText) > 300 {
		a.logText = a.logText[len(a.logText)-300:]
	}
	text := strings.Join(a.logText, "\r\n")
	a.mu.Unlock()
	setText(a.logEdit, text)
	sendMsg(a.logEdit, emSetsel, ^uintptr(0)>>1, ^uintptr(0)>>1)
	sendMsg(a.logEdit, emScrollcaret, 0, 0)
}

func (a *app) drainPending() {
	a.mu.Lock()
	q := a.pending
	a.pending = nil
	a.mu.Unlock()
	for _, ln := range q {
		a.appendLog(ln)
	}
}

func (a *app) queueLog(line string) {
	a.mu.Lock()
	a.pending = append(a.pending, line)
	a.mu.Unlock()
	postMsg(a.hwnd, wmAppLog)
}

func (a *app) setStatus(s string) {
	setText(a.status, s)
}

type doneMsg struct {
	label string
	code  int
}

var (
	doneQueue []doneMsg
	doneMu    sync.Mutex
)

func (a *app) onDone() {
	doneMu.Lock()
	var d *doneMsg
	if len(doneQueue) > 0 {
		d = &doneQueue[0]
		doneQueue = doneQueue[1:]
	}
	doneMu.Unlock()
	if d == nil {
		return
	}
	a.mu.Lock()
	a.busy = ""
	a.mu.Unlock()
	if d.code == 0 {
		a.appendLog("✓ " + d.label + " terminé")
	} else {
		a.appendLog("✗ " + d.label + " → code " + itoa(d.code) + " (voir ci-dessus)")
	}
	a.setStatus("Prêt.")
	go a.refreshStatus()
	if d.label == "deploy" && d.code == 0 {
		a.appendLog("dashboard : https://meteolink.dev:9090")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// ---- moteur d'actions ----

type ktWriter struct {
	a *app
}

func (w *ktWriter) Write(p []byte) (int, error) {
	for _, ln := range strings.Split(string(p), "\n") {
		if t := strings.TrimRight(ln, "\r"); strings.TrimSpace(t) != "" {
			w.a.queueLog(t)
		}
	}
	return len(p), nil
}

func (a *app) kitRunner() *kit.Runner {
	r := kit.NewRunner()
	w := &ktWriter{a: a}
	r.Stdout, r.Stderr = w, w
	return r
}

func (a *app) loadCfg() *kit.Config {
	c, _ := kit.LoadConfig(a.cfgPath)
	return c
}

// runKit — action kit NON interactive en fond (sortie streamée).
func (a *app) runKit(label string, fn func(r *kit.Runner) int) {
	a.mu.Lock()
	if a.busy != "" {
		busy := a.busy
		a.mu.Unlock()
		a.appendLog("patience — « " + busy + " » tourne déjà")
		return
	}
	a.busy = label
	a.mu.Unlock()
	a.setStatus("◌ " + label + " …")
	a.appendLog("▸ " + label + " …")
	go func() {
		code := fn(a.kitRunner())
		doneMu.Lock()
		doneQueue = append(doneQueue, doneMsg{label: label, code: code})
		doneMu.Unlock()
		postMsg(a.hwnd, wmAppDone)
	}()
}

// runKitConsole — action INTERACTIVE (mot de passe, admin) : nouvelle
// console visible où l'utilisateur tape, le GUI attend la fermeture.
func (a *app) runKitConsole(label string, args ...string) {
	a.mu.Lock()
	if a.busy != "" {
		busy := a.busy
		a.mu.Unlock()
		a.appendLog("patience — « " + busy + " » tourne déjà")
		return
	}
	a.busy = label
	a.mu.Unlock()
	a.setStatus("◌ " + label + " (console) …")
	a.appendLog("▸ " + label + " — tapez dans la console noire, fermez-la au retour …")
	go func() {
		full := append([]string{"kit", "--config", a.cfgPath}, args...)
		cmd := exec.Command(a.cgoExe, full...)
		cmd.Dir = filepath.Dir(a.cgoExe)
		cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE}
		err := cmd.Run()
		code := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else {
				code = 1
			}
		}
		doneMu.Lock()
		doneQueue = append(doneQueue, doneMsg{label: label, code: code})
		doneMu.Unlock()
		postMsg(a.hwnd, wmAppDone)
	}()
}

// ---- rafraîchissements (directs, pas de parsing de sortie) ----

func (a *app) refreshVMs() {
	go func() {
		paths := vm.ScanVMs(false)
		hyps := vm.Detect()
		byName := map[string]vm.Hypervisor{}
		running := map[string]bool{}
		for _, h := range hyps {
			byName[h.Name()] = h
			for _, r := range h.Running() {
				running[r] = true
			}
		}
		var rows []vmRow
		for _, p := range paths {			hyp := "?"
			lower := strings.ToLower(p)
			if strings.HasSuffix(lower, ".vmx") {
				hyp = "vmware"
			} else if strings.HasSuffix(lower, ".vbox") {
				hyp = "virtualbox"
			}
			live := running[p]
			if !live {
				base := strings.ToLower(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
				for rp := range running {
					if strings.Contains(strings.ToLower(rp), base) {
						live = true
					}
				}
			}
			mode := "inconnu"
			if h, ok := byName[hyp]; ok {
				mode = h.NetMode(p)
			}
			rows = append(rows, vmRow{p, strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)), hyp, mode, live})
		}
		a.mu.Lock()
		a.vmRows = rows
		a.mu.Unlock()
		postMsg(a.hwnd, wmAppVMs)
		postMsg(a.hwnd, wmAppDone2)
	}()
}

func (a *app) refreshStatus() {
	go func() {
		c := a.loadCfg()
		ssh := "—"
		if c.SSHUp() {
			ssh = "actif"
		} else {
			out, _ := c.SSH("true")
			ssh = "coupé (" + shortDiag(out) + ")"
		}
		dash := "—"
		if ver, ok := dashHealth(c); ok {
			dash = "ok " + ver
		} else {
			dash = "injoignable"
		}
		locked := ""
		if c.VMName != "" {
			locked = c.VMName
		} else if c.VMXPath != "" {
			locked = filepath.Base(c.VMXPath)
		}
		a.mu.Lock()
		a.stSSH, a.stDash, a.stLocked = ssh, dash, locked
		a.mu.Unlock()
		postMsg(a.hwnd, wmAppStatus)
	}()
}

func shortDiag(out string) string {
	o := strings.ToLower(out)
	switch {
	case strings.Contains(o, "refused"), strings.Contains(o, "closed"):
		return "port 22 fermé"
	case strings.Contains(o, "denied"):
		return "clé refusée"
	case strings.Contains(o, "timed out"), strings.Contains(o, "timeout"), strings.Contains(o, "unreachable"), strings.Contains(o, "no route"):
		return "injoignable"
	default:
		return "voir journal"
	}
}

func dashHealth(c *kit.Config) (string, bool) {
	host := c.DashHost
	if host == "" {
		host = "meteolink.dev"
	}
	port := c.DashPort
	if port == "" {
		port = "9090"
	}
	// Transport strict : racines système, vérification ON (pas de -k déguisé).
	client := &http.Client{Timeout: 6 * time.Second, Transport: &http.Transport{}}
	resp, err := client.Get("https://" + host + ":" + port + "/api/health")
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	var doc struct {
		OK      bool   `json:"ok"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil || !doc.OK {
		return "", false
	}
	return doc.Version, true
}

// ---- dispatch boutons ----

// buttonAction — mapping pur bouton → action (testé sans fenêtre) :
// retourne le verbe et ses args CLI. "console:" = console visible,
// "bg:" = fond streamé, "direct:" = appel immédiat (verrouillage).
func buttonAction(id int) (string, []string) {
	switch id {
	case 201:
		return "bg:scan", nil
	case 202:
		return "direct:lock", nil
	case 203:
		return "bg:console", nil
	case 204:
		return "bg:nic", []string{"nat"}
	case 211:
		return "bg:diag", nil
	case 212:
		return "bg:mkkey", nil
	case 213:
		return "console:keysetup", []string{"keysetup"}
	case 214:
		return "bg:ensure", nil
	case 221:
		return "bg:deploy", nil
	case 222:
		return "direct:open", nil
	case 231:
		return "bg:svc", []string{"status"}
	case 232:
		return "bg:svc", []string{"start"}
	case 233:
		return "bg:svc", []string{"stop"}
	case 234:
		return "bg:svc", []string{"restart"}
	case 235:
		return "bg:logs", nil
	case 236:
		return "console:dns", []string{"dns"}
	case 237:
		return "console:tls", []string{"tls"}
	case 238:
		return "bg:verify", nil
	case 239:
		return "bg:backup", nil
	case 240:
		return "bg:snapshot", nil
	case 241:
		return "bg:netinfo", nil
	case 242:
		return "bg:vmon", nil
	case 243:
		return "bg:vmoff", nil
	case 244:
		return "bg:nic-toggle", nil
	}
	return "", nil
}

func (a *app) onButton(id int) {
	kind, args := buttonAction(id)
	if kind == "" {
		return
	}
	switch {
	case kind == "direct:lock":
		a.lockSelected()
	case kind == "direct:open":
		openBrowser(a.dashURL())
		a.appendLog("navigateur → " + a.dashURL())
	case kind == "bg:scan":
		a.appendLog("▸ scan des VMs …")
		a.mu.Lock()
		a.busy = "scan"
		a.mu.Unlock()
		a.setStatus("◌ scan …")
		go func() {
			a.refreshVMs()
			a.mu.Lock()
			a.busy = ""
			a.mu.Unlock()
			postMsg(a.hwnd, wmAppDone2)
		}()
	case kind == "bg:diag":
		a.appendLog("diagnostic : clé, port, auth, IP — voir journal")
		a.runKit("diagnostic", func(r *kit.Runner) int { return diagGUI(a.loadCfg(), r) })
	case kind == "bg:mkkey":
		a.runKit("créer-clé", func(r *kit.Runner) int { return mkKeyGUI(a.loadCfg(), r) })
	case kind == "bg:console":
		a.runKit("console", func(r *kit.Runner) int { return openConsoleGUI(a.loadCfg(), r) })
	case kind == "bg:ensure":
		a.runKit("ensure", func(r *kit.Runner) int { return r.Ensure(a.loadCfg(), a.cfgPath, true) })
	case kind == "bg:deploy":
		a.runKit("deploy", func(r *kit.Runner) int { return r.Deploy(a.loadCfg(), a.cfgPath, true) })
	case kind == "bg:svc":
		sub := "status"
		if len(args) > 0 {
			sub = args[0]
		}
		a.runKit("svc "+sub, func(r *kit.Runner) int { return r.Svc(a.loadCfg(), sub) })
	case kind == "bg:logs":
		a.runKit("logs", func(r *kit.Runner) int { return r.Logs(a.loadCfg(), 40) })
	case kind == "bg:verify":
		a.runKit("verify", func(r *kit.Runner) int { return r.Verify(a.loadCfg()) })
	case kind == "bg:backup":
		a.runKit("backup", func(r *kit.Runner) int { return r.Backup(a.loadCfg(), "backup") })
	case kind == "bg:snapshot":
		a.runKit("snapshot", func(r *kit.Runner) int { return r.Snapshot(a.loadCfg(), "cgo-gui", true) })
	case kind == "bg:netinfo":
		a.runKit("netinfo", func(r *kit.Runner) int { return r.Netinfo(a.loadCfg()) })
	case kind == "bg:vmon":
		a.runKit("vm-start", func(r *kit.Runner) int { return vmPowerGUI(a.loadCfg(), true) })
	case kind == "bg:vmoff":
		a.runKit("vm-stop", func(r *kit.Runner) int { return vmPowerGUI(a.loadCfg(), false) })
	case kind == "bg:nic-toggle":
		a.runKit("nic-toggle", func(r *kit.Runner) int { return nicToggleGUI(a.loadCfg(), a.cfgPath) })
	case kind == "console:keysetup":
		a.runKitConsole("poser-clé", "keysetup")
	case kind == "console:dns":
		a.runKitConsole("dns", "dns")
	case kind == "console:tls":
		a.runKitConsole("tls", "tls")
	}
}

func (a *app) dashURL() string {
	host, port := "meteolink.dev", "9090"
	if c := a.loadCfg(); c != nil {
		if c.DashHost != "" {
			host = c.DashHost
		}
		if c.DashPort != "" {
			port = c.DashPort
		}
	}
	return "https://" + host + ":" + port
}
