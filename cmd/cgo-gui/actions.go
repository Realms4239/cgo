//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/internal/vm"
	"golang.org/x/sys/windows"
)

// ---- journal thread-safe ----

// guiLogFh — miroir fichier du journal (cgo-gui-<date>.log à côté de
// l'exe, sinon %TEMP%) : une panne devient envoyable au support.
var guiLogFh = struct {
	f   *os.File
	key string
}{}

func appendGUIFile(s string) {
	day := time.Now().Format("20060102")
	exe, err := os.Executable()
	dir := ""
	if err == nil {
		dir = filepath.Dir(exe)
	}
	clean := strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' {
			return -1
		}
		return r
	}, s)
	line := time.Now().Format("15:04:05") + " " + clean + "\n"
	try := func(dir string) bool {
		if dir == "" {
			return false
		}
		p := filepath.Join(dir, "cgo-gui-"+day+".log")
		if guiLogFh.f != nil && guiLogFh.key == p {
			_, _ = guiLogFh.f.WriteString(line)
			return true
		}
		if f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			if guiLogFh.f != nil {
				_ = guiLogFh.f.Close()
			}
			guiLogFh.f, guiLogFh.key = f, p
			_, _ = f.WriteString(line)
			return true
		}
		return false
	}
	if !try(dir) {
		try(os.TempDir())
	}
}

func (a *app) appendLog(line string) {
	a.mu.Lock()
	a.logText = append(a.logText, line)
	if len(a.logText) > 300 {
		a.logText = a.logText[len(a.logText)-300:]
	}
	a.logDirty = true
	a.mu.Unlock()
	appendGUIFile(line)
}

// flushLog — rendu effectif vers le contrôle (5 images/s max) : réécrire
// 30 Ko + reflow à CHAQUE ligne saturait le thread UI pendant les actions
// bavardes (deploy) — c'était le « freeze à chaque clic ».
func (a *app) flushLog() {
	a.mu.Lock()
	if !a.logDirty {
		a.mu.Unlock()
		return
	}
	a.logDirty = false
	start := 0
	if len(a.logText) > 120 {
		start = len(a.logText) - 120
	}
	text := strings.Join(a.logText[start:], "\r\n")
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
	// cache 45 s : LoadConfig résout l'IP invitée via vmrun (secondes)
	// quand SSHHost=auto — SANS cache, CHAQUE clic pend le thread UI.
	a.mu.Lock()
	if c.SSHHost == "" || c.SSHHost == "auto" {
		if a.cfgHost != "" && time.Since(a.cfgAt) < 45*time.Second {
			c.SSHHost = a.cfgHost
		}
	} else {
		a.cfgHost, a.cfgAt = c.SSHHost, time.Now()
	}
	a.mu.Unlock()
	return c
}

// noteHost — mémorise une IP résolue en fond (refreshStatus/ensure) pour
// les clics suivants. Jamais d'hyperviseur sur le thread UI.
func (a *app) noteHost(ip string) {
	if ip == "" || ip == "auto" {
		return
	}
	a.mu.Lock()
	a.cfgHost, a.cfgAt = ip, time.Now()
	a.mu.Unlock()
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

// refreshVMs — corps SYNCHRONE (les appelants préfixent déjà `go`) :
// l'ancienne double-goroutine effaçait `busy` AVANT la fin du scan.
func (a *app) refreshVMs() {
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
}

// refreshStatus — UNE sonde à la fois (single-flight) : sans garde, un
// clic + le ticker 30 s lançaient DEUX ssh.exe concurrents et le clic
// paraissait « freeze » le temps des handshakes sériels.
func (a *app) refreshStatus() {
	if !a.statFlight.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer a.statFlight.Store(false)
		c := a.loadCfg()
		if c.SSHHost != "" && c.SSHHost != "auto" {
			a.noteHost(c.SSHHost)
		}
		ssh := "—"
		// pré-test TCP instantané (0 ressource) : distingue « VM éteinte »
		// de « sshd absent » SANS payer un handshake ssh de 4 s pour rien.
		if c.SSHHost == "" || c.SSHHost == "auto" || portOpenGUI(c.SSHHost, c.SSHPort) {
			if c.SSHUp() {
				ssh = "actif"
			} else if c.SSHHost != "" && c.SSHHost != "auto" {
				out, _ := c.SSH("true")
				ssh = "coupé (" + shortDiag(out) + ")"
			}
		} else {
			ssh = "VM éteinte / réseau coupé"
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

// portOpenGUI — dial TCP court, miroir kit.portOpen (non exporté là-bas).
func portOpenGUI(host, port string) bool {
	if host == "" || port == "" {
		return false
	}
	d := net.Dialer{Timeout: 2 * time.Second}
	cn, err := d.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return false
	}
	_ = cn.Close()
	return true
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
		a.runKit("logs", func(r *kit.Runner) int {
			c := a.loadCfg()
			if c.VMXPath == "" && c.SSHHost == "" {
				fmt.Println("rien à lire : verrouillez d'abord une VM (liste ci-dessus)")
				return 3
			}
			return r.Logs(c, 40)
		})
	case kind == "bg:verify":
		a.runKit("verify", func(r *kit.Runner) int { return r.Verify(a.loadCfg()) })
	case kind == "bg:backup":
		a.runKit("backup", func(r *kit.Runner) int { return r.Backup(a.loadCfg(), "backup") })
	case kind == "bg:snapshot":
		a.runKit("snapshot", func(r *kit.Runner) int { return r.Snapshot(a.loadCfg(), "cgo-gui", true) })
	case kind == "bg:netinfo":
		a.runKit("netinfo", func(r *kit.Runner) int { return r.Netinfo(a.loadCfg()) })
	case kind == "bg:vmon":
		a.runKit("vm-start", func(r *kit.Runner) int { return vmPowerGUI(a.loadCfg(), r, true) })
	case kind == "bg:vmoff":
		a.runKit("vm-stop", func(r *kit.Runner) int { return vmPowerGUI(a.loadCfg(), r, false) })
	case kind == "bg:nic-toggle":
		a.runKit("nic-toggle", func(r *kit.Runner) int { return nicToggleGUI(a.loadCfg(), r, a.cfgPath) })
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
