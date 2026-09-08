//go:build windows

package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/Realms4239/cgo/internal/kit"
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