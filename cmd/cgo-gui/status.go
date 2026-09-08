//go:build windows

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/internal/vm"
)

// ---- VM list ----

func (a *app) onVMs() {
	a.mu.Lock()
	rows := a.vmRows
	a.mu.Unlock()
	sendMsg(a.vmList, lbResetcontent, 0, 0)
	for _, r := range rows {
		dot := "○"
		if r.live {
			dot = "●"
		}
		mode := r.mode
		if mode == "" || mode == "inconnu" {
			mode = "?"
		}
		lbl := fmt.Sprintf("%s %s [%s · %s]", dot, r.name, r.hyp, mode)
		sendMsg(a.vmList, lbAddstring, 0, ptr(lbl))
	}
	a.appendLog(fmt.Sprintf("%d VM(s) — double-clic = verrouiller", len(rows)))
}

func ptr(s string) uintptr {
	return uintptr(unsafe.Pointer(utf16(s)))
}

func (a *app) lockSelected() {
	idx := int(sendMsg(a.vmList, lbGetcursel, 0, 0))
	a.mu.Lock()
	rows := a.vmRows
	a.mu.Unlock()
	if idx < 0 || idx >= len(rows) {
		a.appendLog("sélectionnez d'abord une VM dans la liste")
		return
	}
	v := rows[idx]
	// hyperviseur réel pour SaveVMX (pas le "?" de détection)
	hyp := v.hyp
	if hyp == "?" {
		if strings.HasSuffix(strings.ToLower(v.path), ".vbox") {
			hyp = "virtualbox"
		} else {
			hyp = "vmware"
		}
	}
	if err := kit.SaveVMX(a.cfgPath, v.path, hyp); err != nil {
		a.appendLog("verrouillage : " + err.Error())
		return
	}
	setText(a.locked, "Verrouillée : "+v.name+" ("+hyp+")")
	a.appendLog("verrouillée : " + v.name)
	go a.refreshStatus()
}

// ---- statuts ----

func (a *app) onStatus() {
	a.mu.Lock()
	ssh, dash, locked := a.stSSH, a.stDash, a.stLocked
	a.mu.Unlock()
	setText(a.sshLbl, "SSH : "+orDash(ssh))
	setText(a.dashLbl, "Dashboard : "+orDash(dash))
	if locked != "" {
		setText(a.locked, "Verrouillée : "+locked)
	}
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func (a *app) onScanDone() {
	a.mu.Lock()
	a.busy = ""
	a.mu.Unlock()
	a.setStatus("Prêt.")
	a.appendLog("✓ scan terminé")
}

// ---- actions directes ----

func mkKeyGUI(c *kit.Config, r *kit.Runner) int {
	key := c.SSHKey
	if strings.HasPrefix(key, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			key = filepath.Join(h, key[2:])
		}
	}
	if _, err := os.Stat(key); err == nil {
		fmt.Fprintln(r.Stdout, "clé déjà présente : "+key+" — rien à faire")
		return 0
	}
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		fmt.Fprintln(r.Stderr, "ssh-keygen introuvable — installez le client OpenSSH")
		return 2
	}
	_ = os.MkdirAll(filepath.Dir(key), 0700)
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", key, "-q")
	cmd.Stdout, cmd.Stderr = r.Stdout, r.Stderr
	if err := cmd.Run(); err != nil {
		return 2
	}
	fmt.Fprintln(r.Stdout, "clé créée : "+key)
	return 0
}

func openConsoleGUI(c *kit.Config, r *kit.Runner) int {	if c.VMXPath == "" {
		fmt.Fprintln(r.Stderr, "aucune VM verrouillée")
		return 3
	}
	for _, h := range vm.Detect() {
		if c.Hypervisor != "" && h.Name() != c.Hypervisor {
			continue
		}
		if err := h.StartGUI(c.VMXPath); err == nil {
			fmt.Fprintln(r.Stdout, "console ouverte")
			return 0
		}
	}
	fmt.Fprintln(r.Stderr, "ouverture impossible — lancez VMware/VirtualBox à la main")
	return 4
}

func vmPowerGUI(c *kit.Config, start bool) int {
	if c.VMXPath == "" {
		fmt.Println("aucune VM verrouillée")
		return 3
	}
	var hyp vm.Hypervisor
	for _, h := range vm.Detect() {
		if c.Hypervisor != "" && h.Name() != c.Hypervisor {
			continue
		}
		hyp = h
	}
	if hyp == nil {
		fmt.Println("hyperviseur absent")
		return 4
	}
	var err error
	if start {
		err = hyp.Start(c.VMXPath)
	} else {
		err = hyp.Stop(c.VMXPath)
	}
	if err != nil {
		fmt.Println("échec : " + err.Error())
		return 4
	}
	fmt.Println("demandé")
	return 0
}

func nicToggleGUI(c *kit.Config, cfgPath string) int {
	r := kit.NewRunner()
	// mode actuel via le même résolveur que partout
	hyp, vmx, err := pickVMGUI(c)
	if err != nil {
		fmt.Println(err.Error())
		return 3
	}
	mode := hyp.NetMode(vmx)
	target := "nat"
	if mode == "nat" {
		target = "bridged"
	}
	return r.Nic(c, cfgPath, []string{target}, true)
}

func pickVMGUI(c *kit.Config) (vm.Hypervisor, string, error) {
	if c.VMXPath != "" {
		for _, h := range vm.Detect() {
			if c.Hypervisor == "" || h.Name() == c.Hypervisor {
				return h, c.VMXPath, nil
			}
		}
	}
	return nil, "", fmt.Errorf("verrouillez d'abord une VM (liste ci-dessus)")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	_ = cmd.Start()
}

// diagGUI — diagnostic d'accès en 4 lignes (clé, IP, port, auth), miroir du
// checklist TUI : chaque KO dit son remède exact. Sortie via le Runner
// (journal), jamais de throw.
func diagGUI(c *kit.Config, r *kit.Runner) int {
	if c.SSHUser == "" {
		fmt.Fprintln(r.Stdout, "[ok] (sauté : utilisateur vide — renseignez-le d'abord)")
		return 2
	}
	key := c.SSHKey
	if strings.HasPrefix(key, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			key = filepath.Join(h, key[2:])
		}
	}
	ok, ko := "[ok]", "[KO]"
	if st, err := os.Stat(key); err == nil && !st.IsDir() {
		fmt.Fprintln(r.Stdout, ok+" clé locale : "+key)
	} else {
		fmt.Fprintln(r.Stdout, ko+" clé locale absente — « Créer la clé » puis « Poser la clé »")
	}
	host := c.SSHHost
	if host == "" || host == "auto" {
		host = ""
		for _, h := range vm.Detect() {
			if c.Hypervisor != "" && h.Name() != c.Hypervisor {
				continue
			}
			if ip := h.GuestIP(c.VMXPath); ip != "" {
				host = ip
			}
		}
		if host == "" {
			fmt.Fprintln(r.Stdout, ko+" IP invitée non résolue — VM éteinte ? « Démarrer / Réessayer »")
			return 0
		}
		fmt.Fprintln(r.Stdout, ok+" IP invitée : "+host+" (hyperviseur)")
	} else {
		fmt.Fprintln(r.Stdout, ok+" IP invitée : "+host+" (configurée)")
	}
	port := c.SSHPort
	if port == "" {
		port = "22"
	}
	d := net.Dialer{Timeout: 3 * time.Second}
	cn, err := d.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Fprintln(r.Stdout, ko+" port "+port+" fermé sur "+host+" — DANS la VM : sudo apt install -y openssh-server && sudo systemctl enable --now ssh")
		return 0
	}
	_ = cn.Close()
	fmt.Fprintln(r.Stdout, ok+" port "+port+" ouvert sur "+host)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=4",
		"-o", "StrictHostKeyChecking=accept-new", "-p", port, "-i", key,
		c.SSHUser+"@"+host, "true")
	if out, err := cmd.CombinedOutput(); err == nil {
		fmt.Fprintln(r.Stdout, ok+" clé acceptée par "+c.SSHUser+"@"+host+" — prêt à déployer")
	} else {
		o := strings.ToLower(string(out))
		if strings.Contains(o, "permission denied") || strings.Contains(o, "denied") {
			fmt.Fprintln(r.Stdout, ko+" clé refusée — « Poser la clé SSH » (mot de passe, une fois)")
		} else {
			fmt.Fprintln(r.Stdout, ko+" auth : "+strings.TrimSpace(string(out)))
		}
	}
	return 0
}
