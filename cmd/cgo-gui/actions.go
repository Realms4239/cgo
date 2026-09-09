//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/Realms4239/cgo/internal/kit"
)

// ---- journal thread-safe (rendu Win32 ; le miroir fichier vit dans core) ----

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

// flushLog — rendu effectif (5 images/s max), APPEND-ONLY : seul le
// delta depuis le dernier flush est injecté (EM_REPLACESEL en fin),
// O(delta) au lieu de réécrire 120 lignes + reflow complet à chaque tick.
// Reset complet (rare) si la mémoire a tronqué (300 lignes max).
func (a *app) flushLog() {
	a.mu.Lock()
	if !a.logDirty {
		a.mu.Unlock()
		return
	}
	a.logDirty = false
	n := len(a.logText)
	if a.shown > n {
		a.shown = 0 // tronqué entre-temps → reset complet
	}
	var delta string
	full := false
	if a.shown == 0 && n > 0 {
		start := 0
		if n > 120 {
			start = n - 120
		}
		delta = strings.Join(a.logText[start:], "\r\n")
		a.shown = n
		full = true
	} else if a.shown < n {
		end := n
		if end-a.shown > 80 {
			end = a.shown + 80 // borne un tick (le reste au suivant)
		}
		delta = "\r\n" + strings.Join(a.logText[a.shown:end], "\r\n")
		a.shown = end
		if end < n {
			a.logDirty = true
		}
	}
	a.mu.Unlock()
	if delta == "" {
		return
	}
	if full {
		setText(a.logEdit, delta)
	} else {
		sendMsg(a.logEdit, emSetsel, ^uintptr(0)>>1, ^uintptr(0)>>1)
		p := utf16(delta)
		sendMsg(a.logEdit, emReplacesel, 0, uintptr(unsafe.Pointer(p)))
	}
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

func (a *app) loadCfg() *kit.Config { return a.cfg.load(a.cfgPath) }

func (a *app) noteHost(ip string) { a.cfg.note(ip) }

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
	rows := scanVMRows()
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
		ssh, dash, locked := probeStatus(c)
		a.mu.Lock()
		a.stSSH, a.stDash, a.stLocked = ssh, dash, locked
		a.mu.Unlock()
		postMsg(a.hwnd, wmAppStatus)
		// Prochaine étape : coûteux (~10-25 s), calculé APRÈS l'affichage
		// du statut pour ne pas le retarder, même single-flight.
		nx := kit.FirstOpen(kit.GatherNext(c))
		a.mu.Lock()
		if nx == nil {
			a.stNext, a.stNextKind, a.stNextArgs = "tout est vert — dashboard prêt", "", nil
		} else if nx.Remedy != "" {
			a.stNext, a.stNextKind, a.stNextArgs = nx.Label+" — "+nx.Remedy, nx.Verb, nx.Args
		} else {
			a.stNext, a.stNextKind, a.stNextArgs = nx.Label+" ("+nx.Detail+")", nx.Verb, nx.Args
		}
		a.stNextDone = true
		a.mu.Unlock()
		postMsg(a.hwnd, wmAppStatus)
	}()
}

// portOpenGUI — dial TCP court, miroir kit.portOpen (non exporté là-bas).
// ---- dispatch boutons ----

// buttonAction — mapping pur bouton → action (testé sans fenêtre) :
// retourne le verbe et ses args CLI. "console:" = console visible,
// "bg:" = fond streamé, "direct:" = appel immédiat (verrouillage).
func (a *app) onButton(id int) {
	kind, args := buttonAction(id)
	if kind == "" {
		return
	}
	a.runKind(kind, args)
}

// runKind — exécute un verbe (bouton ou « Suite » qui rejoue le verbe de
// la prochaine étape calculée). Factorisé de onButton pour que Suite
// emprunte exactement le même chemin que le bouton d'origine.
func (a *app) runKind(kind string, args []string) {
	switch {
	case kind == "direct:lock":
		a.lockSelected()
	case kind == "direct:saveuser":
		a.saveUser()
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
	case kind == "bg:hosttun":
		a.runKit("tunnel-hôte", func(r *kit.Runner) int {
			c := a.loadCfg()
			_, argv, blocked := hostTunCmd(c, filepath.Dir(a.cgoExe), r.Root)
			if blocked != "" {
				fmt.Println(blocked)
				return 3
			}
			cmd := kit.BgCmd("powershell", argv...)
			cmd.Stdout, cmd.Stderr = r.Stdout, r.Stderr
			if err := cmd.Run(); err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					return ee.ExitCode()
				}
				return 1
			}
			return 0
		})
	case kind == "bg:guest":
		a.runKit("invité", func(r *kit.Runner) int { return r.Guest(a.loadCfg(), a.cfgPath, false) })
	case kind == "bg:vnet":
		a.runKit("réseau-hôte", func(r *kit.Runner) int { return r.Vnet(a.loadCfg()) })
	case kind == "console:keysetup":
		a.runKitConsole("poser-clé", "keysetup")
	case kind == "console:dns":
		a.runKitConsole("dns", "dns")
	case kind == "console:tls":
		a.runKitConsole("tls", "tls")
	case kind == "bg:guide":
		a.runKit("guide", func(r *kit.Runner) int {
			for _, ln := range strings.Split(kit.Readme(), "\n") {
				fmt.Fprintln(r.Stdout, ln)
			}
			return 0
		})
	case kind == "direct:suite":
		a.mu.Lock()
		done, nk, na := a.stNextDone, a.stNextKind, a.stNextArgs
		a.mu.Unlock()
		if !done {
			a.appendLog("analyse en cours — patientez 10 s puis Suite")
			return
		}
		if nk == "" {
			a.appendLog("tout est vert — rien à faire (Rescanner pour revérifier)")
			return
		}
		a.appendLog("▶ suite : " + nk)
		a.runKind(nk, na)
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
