//go:build windows

// Meteolink Kit — mini application Windows native. Fine coquille autour du
// CLI éprouvé : chaque bouton exécute `cgo.exe kit …` à côté d'elle et
// stream la sortie dans le journal. Les actions interactives (mot de
// passe, admin) s'ouvrent dans une console visible ; le reste tourne en
// fond, résultats via PostMessage (jamais de touché UI hors thread).
package main

import (
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	appTitle = "Meteolink Kit — centre de contrôle"
	winW     = 940
	winH     = 700

	idVMList = 101
	idLog    = 102
	idStatus = 103
	idLocked = 104
	idSSH    = 105
	idDash   = 106
)

type buttonDef struct {
	id    int
	label string
	x, y  int
	w     int
}

var groupActions = []buttonDef{
	{201, "Rescanner", 20, 0, 110},
	{202, "Verrouiller", 140, 0, 110},
	{203, "Ouvrir la console", 260, 0, 150},
	{204, "Carte en NAT", 420, 0, 120},
}

var accessActions = []buttonDef{
	{211, "Diagnostiquer", 20, 0, 120},
	{212, "Créer la clé", 150, 0, 110},
	{213, "Poser la clé", 270, 0, 100},
	{214, "Démarrer / Réessayer", 380, 0, 160},
}

var deployActions = []buttonDef{
	{221, "DÉPLOYER", 20, 0, 160},
	{222, "Ouvrir le dashboard", 190, 0, 160},
}

var controlActions = []buttonDef{
	{231, "État", 0, 0, 60}, {232, "Start", 0, 0, 60}, {233, "Stop", 0, 0, 60}, {234, "Restart", 0, 0, 70},
	{235, "Logs", 0, 0, 60}, {236, "DNS", 0, 0, 55}, {237, "TLS", 0, 0, 55}, {238, "Vérifier", 0, 0, 75},
	{239, "Backup", 0, 0, 70}, {240, "Snapshot", 0, 0, 85},
	{241, "Réseau invité", 0, 0, 110}, {242, "Démarrer VM", 0, 0, 105}, {243, "Arrêter VM", 0, 0, 95},
	{244, "Carte NAT/pont", 0, 0, 115},
}

type app struct {
	hwnd    windows.HWND
	hfont   uintptr
	vmList  windows.HWND
	logEdit windows.HWND
	status  windows.HWND
	locked  windows.HWND
	sshLbl  windows.HWND
	dashLbl windows.HWND
	cgoExe   string
	cfgPath  string
	mu       sync.Mutex
	busy     string
	pending  []string
	logText  []string
	vmRows   []vmRow
	stSSH    string
	stDash   string
	stLocked string
}

type vmRow struct {
	path, name, hyp, mode string
	live                  bool
}

var theApp *app
var wndProcPtr = windows.NewCallback(wndProc)

func main() {
	exe, err := os.Executable()
	if err != nil {
		fatalBox("introuvable : " + err.Error())
		return
	}
	dir := filepath.Dir(exe)
	cgo := filepath.Join(dir, "cgo.exe")
	if _, err := os.Stat(cgo); err != nil {
		fatalBox("cgo.exe introuvable à côté de cgo-gui.exe.\nDézippez l'archive complète au même endroit.")
		return
	}
	theApp = &app{
		cgoExe:  cgo,
		cfgPath: filepath.Join(dir, "kit", "cgo-vm.yaml"),
	}
	if err := theApp.run(); err != nil {
		fatalBox(err.Error())
	}
}

func fatalBox(msg string) {
	t, _ := windows.UTF16PtrFromString(msg)
	c, _ := windows.UTF16PtrFromString("Meteolink Kit")
	pMsgBox.Call(uintptr(0), uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), uintptr(0x10))
}
