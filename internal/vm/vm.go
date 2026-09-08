// Package vm — détection et pilotage des hyperviseurs, pur Go multi-OS.
// Remplace le scan bash historique (appels shell externes silencieux hors
// Windows) : la détection enumère les disques nativement et cherche
// .vmx/.vbox avec la convention D:/VMs, C:/VMs (racine historique tolérée).
package vm

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Hypervisor — un pilote d'hyperviseur (VMware vmrun ou VirtualBox VBoxManage).
type Hypervisor interface {
	Name() string
	// Exe retourne le chemin du binaire de contrôle, "" si absent.
	Exe() string
	// Running liste les VMs allumées (chemins .vmx/.vbox).
	Running() []string
	// Start démarre une VM (nogui/headless).
	Start(vmx string) error
	// StartGUI ouvre la console graphique de la VM (pour coller une commande
	// dedans à la main : installation openssh-server, dépannage réseau).
	StartGUI(vmx string) error
	// Stop arrête proprement une VM (ACPI, puis forcé après délai).
	Stop(vmx string) error
	// GuestIP interroge l'IP invitée d'une VM allumée ("" si inconnue).
	GuestIP(vmx string) string
	// NetMode lit le mode réseau configuré SANS allumer la VM : ponté (accès
	// direct), nat (port-forward requis), hôte-only, inconnu. VMware : lu
	// dans le .vmx (ethernet0.connectionType). VirtualBox : showvminfo nicN.
	NetMode(vmx string) string
	// SetNetMode bascule la NIC1 (nat|bridged|hostonly selon pilote) — VM
	// ÉTEINTE exigée, refus explicite sinon. VBox : modifyvm. VMware :
	// réécriture ethernet0.connectionType dans le .vmx.
	SetNetMode(vmx, mode string) error
}

// NatForwarder — hyperviseurs en mode NAT (VirtualBox) : l'IP invitée
// n'est pas joignable depuis l'hôte ; l'accès passe par un port redirigé.
type NatForwarder interface {
	// EnsureNatSSH redirige hostPort → 22 invité + dashPort → 9090 invité
	// (idempotent). Échoue en clair si ports occupés ou VM inverrouillable.
	EnsureNatSSH(vmx, hostPort, dashPort string) error
	// NatHostAddr l'adresse d'accès côté hôte (127.0.0.1).
	NatHostAddr() string
}

// ---- VMware ----

type vmware struct{ exe string }

func (v *vmware) Name() string { return "vmware" }
func (v *vmware) Exe() string  { return v.exe }

func (v *vmware) run(args ...string) (string, error) {
	return v.runCtx(context.Background(), args...)
}

// runCtx — vmrun borné : les REQUÊTES (list, getGuestIPAddress…) ne doivent
// jamais pendre le thread appelant. getGuestIPAddress -wait bloque déjà
// côté vmrun ; start/stop gardent le contexte libre (pilotés en fond).
func (v *vmware) runCtx(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, v.exe, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (v *vmware) Running() []string {
	out, err := v.run("list")
	if err != nil {
		return nil
	}
	var vms []string
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasSuffix(ln, ".vmx") || strings.HasSuffix(ln, ".vbox") {
			vms = append(vms, ln)
		}
	}
	return vms
}

func (v *vmware) Start(vmx string) error {
	_, err := v.run("start", vmx, "nogui")
	return err
}

// StartGUI — ouvre la fenêtre Workstation sur la VM.
func (v *vmware) StartGUI(vmx string) error {
	_, err := v.run("-T", "ws", "start", vmx, "gui")
	return err
}

// Stop — soft d'abord (invité ACPI), hard en dernier recours après délai.
func (v *vmware) Stop(vmx string) error {
	_, err := v.run("stop", vmx, "soft")
	if err != nil {
		time.Sleep(10 * time.Second)
		_, err = v.run("stop", vmx, "hard")
	}
	return err
}

func (v *vmware) GuestIP(vmx string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	out, err := v.runCtx(ctx, "getGuestIPAddress", vmx, "-wait", "5")
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(strings.Split(out, "\n")[0])
	if !strings.Contains(ip, ".") {
		return ""
	}
	return ip
}

// NetMode — ethernetX.connectionType du .vmx, première NIC trouvée :
// bridged → "ponté", nat → "nat", hostonly → "hôte-only", custom → son nom.
// Fichier illisible = "inconnu" (jamais d'erreur : c'est un affichage).
func (v *vmware) NetMode(vmx string) string {
	data, err := os.ReadFile(vmx)
	if err != nil {
		return "inconnu"
	}
	best := ""
	for _, ln := range strings.Split(string(data), "\n") {
		ln = strings.TrimSpace(strings.ToLower(ln))
		if !strings.HasPrefix(ln, "ethernet") || !strings.Contains(ln, "connectiontype") {
			continue
		}
		parts := strings.SplitN(ln, "=", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if best == "" {
			best = val // ethernet0 d'abord (ordre du fichier)
		}
	}
	switch best {
	case "bridged":
		return "ponté"
	case "nat":
		return "nat"
	case "hostonly":
		return "hôte-only"
	case "":
		return "inconnu"
	default:
		return best
	}
}

// applyVmxNetMode — réécriture pure (testée) de ethernet0.connectionType.
// Ajoute la ligne si absente (après le bloc ethernet0, sinon fin de fichier).
func applyVmxNetMode(content, mode string) string {
	var out []string
	done := false
	ethIdx := -1
	for _, ln := range strings.Split(content, "\n") {
		t := strings.TrimSpace(strings.ToLower(ln))
		if strings.HasPrefix(t, "ethernet") {
			ethIdx = len(out)
		}
		if strings.HasPrefix(t, "ethernet0.connectiontype") {
			eq := strings.Index(ln, "=")
			if eq >= 0 {
				ln = ln[:eq+1] + ` "` + mode + `"`
				done = true
			}
		}
		out = append(out, ln)
	}
	if !done {
		line := `ethernet0.connectionType = "` + mode + `"`
		if ethIdx >= 0 && ethIdx+1 <= len(out) {
			out = append(out[:ethIdx+1], append([]string{line}, out[ethIdx+1:]...)...)
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// SetNetMode — nat|bridged|hostonly dans le .vmx (VM éteinte exigée :
// VMware ignore/réécrit la config d'une VM allumée).
func (v *vmware) SetNetMode(vmx, mode string) error {
	for _, r := range v.Running() {
		if SameVM(r, vmx) {
			return fmt.Errorf("VM allumée — éteignez-la d'abord (la config live serait écrasée)")
		}
	}
	data, err := os.ReadFile(vmx)
	if err != nil {
		return err
	}
	// sauvegarde une fois (jamais de .bak en double)
	bak := vmx + ".meteolink-bak"
	if _, err := os.Stat(bak); err != nil {
		_ = os.WriteFile(bak, data, 0644)
	}
	return os.WriteFile(vmx, []byte(applyVmxNetMode(string(data), mode)), 0644)
}

// ---- VirtualBox ----

type virtualbox struct{ exe string }

func (v *virtualbox) Name() string { return "virtualbox" }
func (v *virtualbox) Exe() string  { return v.exe }

func (v *virtualbox) run(args ...string) (string, error) {
	return v.runCtx(context.Background(), args...)
}

// runCtx — VBoxManage borné : guestproperty sur VM éteinte répond vite,
// mais un manager gelé ne doit jamais pendre l'appelant (le GUI clique
// sur le thread UI — 10 s max, jamais l'infini).
func (v *virtualbox) runCtx(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, v.exe, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// parseRunningVMs — `list runningvms` rend `"Nom" {uuid}` (ni chemin ni
// .vbox) : extrait les UUID, jamais de match sur ".vbox" (l'ancien parse
// rendait toujours vide → toute VM allumée passait pour éteinte).
func parseRunningVMs(out string) []string {
	var uuids []string
	for _, ln := range strings.Split(out, "\n") {
		i := strings.Index(ln, "{")
		j := strings.Index(ln, "}")
		if i >= 0 && j > i {
			uuids = append(uuids, strings.TrimSpace(ln[i+1:j]))
		}
	}
	return uuids
}

// runningPaths — résout chaque UUID allumé vers son .vbox (showvminfo).
func (v *virtualbox) runningPaths() []string {
	out, err := v.run("list", "runningvms")
	if err != nil {
		return nil
	}
	var paths []string
	for _, u := range parseRunningVMs(out) {
		info, err := v.run("showvminfo", u, "--machinereadable")
		if err != nil {
			continue
		}
		for _, ln := range strings.Split(info, "\n") {
			if strings.HasPrefix(ln, "CfgFile=") {
				p := strings.Trim(strings.TrimPrefix(ln, "CfgFile="), "\"\r")
				// chemins invité Windows (backslashes) normalisés
				paths = append(paths, filepath.FromSlash(strings.ReplaceAll(p, "\\", "/")))
			}
		}
	}
	return paths
}

// Running — les VMs allumées (chemins .vbox résolus, pas les noms).
func (v *virtualbox) Running() []string {
	return v.runningPaths()
}

func (v *virtualbox) Start(vbx string) error {
	name, err := v.ensureRegistered(vbx)
	if err != nil {
		return err
	}
	_, err = v.run("startvm", name, "--type", "headless")
	return err
}

// resolveName — nom enregistré correspondant au .vbox, SANS effet de bord :
// nom direct vérifié par CfgFile, sinon balayage des enregistrées (renommée
// dans le manager : "ubuntu Fanasina" vs fichier). "" si introuvable.
func (v *virtualbox) resolveName(vbx string) string {
	want := normPath(vbx)
	name := strings.TrimSuffix(filepath.Base(vbx), ".vbox")
	if v.cfgFile(name) == want {
		return name
	}
	out, _ := v.run("list", "vms")
	for _, ln := range strings.Split(out, "\n") {
		nm := quotedName(ln)
		if nm == "" || nm == name {
			continue
		}
		if v.cfgFile(nm) == want {
			return nm
		}
	}
	return ""
}

// cfgFile — le .vbox enregistré sous ce nom ("", si inconnu).
func (v *virtualbox) cfgFile(name string) string {
	out, err := v.run("showvminfo", name, "--machinereadable")
	if err != nil {
		return ""
	}
	for _, ln := range strings.Split(out, "\n") {
		if strings.HasPrefix(ln, "CfgFile=") {
			return normPath(strings.Trim(strings.TrimPrefix(ln, "CfgFile="), "\"\r"))
		}
	}
	return ""
}

func normPath(s string) string {
	return strings.ToLower(filepath.ToSlash(s))
}

func quotedName(ln string) string {
	i := strings.Index(ln, "\"")
	if i < 0 {
		return ""
	}
	rest := ln[i+1:]
	if j := strings.Index(rest, "\""); j > 0 {
		return rest[:j]
	}
	return ""
}

// ensureRegistered — enregistre le .vbox s'il est orphelin (trouvé par scan
// disque mais absent du manager) puis rend son nom. C'est le chaînon qui
// manquait : tout le reste échouait en silence sur VM non enregistrée.
func (v *virtualbox) ensureRegistered(vbx string) (string, error) {
	return v.EnsureRegistered(vbx)
}

// EnsureRegistered — version exportée (GUI, scripts) : idempotente.
func (v *virtualbox) EnsureRegistered(vbx string) (string, error) {
	if n := v.resolveName(vbx); n != "" {
		return n, nil
	}
	if _, err := v.run("registervm", vbx); err != nil {
		return "", fmt.Errorf("enregistrement impossible (%s) : %v — ajoutez-la dans VirtualBox (Machine > Ajouter)", vbx, err)
	}
	if n := v.resolveName(vbx); n != "" {
		return n, nil
	}
	return "", fmt.Errorf("enregistrée mais introuvable : %s", vbx)
}

// StartGUI — ouvre la fenêtre VirtualBox sur la VM.
func (v *virtualbox) StartGUI(vbx string) error {
	name, err := v.ensureRegistered(vbx)
	if err != nil {
		return err
	}
	_, err = v.run("startvm", name, "--type", "gui")
	return err
}

// Stop — ACPI propre d'abord, forcé après délai.
func (v *virtualbox) Stop(vbx string) error {
	name := v.resolveName(vbx)
	if name == "" {
		return fmt.Errorf("VM inconnue de VirtualBox : %s (ajoutez-la : Machine > Ajouter)", vbx)
	}
	_, _ = v.run("controlvm", name, "acpipowerbutton")
	time.Sleep(15 * time.Second)
	_, _ = v.run("controlvm", name, "poweroff")
	return nil
}

// vmID — l'UUID de la VM, requis par guestproperty.
func (v *virtualbox) vmID(name string) string {
	out, _ := v.run("showvminfo", name, "--machinereadable")
	for _, ln := range strings.Split(out, "\n") {
		if strings.HasPrefix(ln, "UUID=") {
			return strings.Trim(strings.TrimPrefix(ln, "UUID="), "\"\r")
		}
	}
	return ""
}

// GuestIP — deux voies complémentaires :
//  1. guestproperty /VirtualBox/GuestInfo/Net/0/V4/IP (si additions invité installées) ;
//  2. port-forwarding NAT : ssh de l'hôte vers 127.0.0.1:2222.
func (v *virtualbox) GuestIP(vbx string) string {
	name := v.resolveName(vbx)
	if name == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// voie 1 : guestproperty directe
	for _, key := range []string{
		"/VirtualBox/GuestInfo/Net/0/V4/IP",
		"/VirtualBox/GuestInfo/Net/1/V4/IP",
	} {
		out, err := v.runCtx(ctx, "guestproperty", "get", name, key)
		if err == nil && strings.Contains(out, "Value:") {
			f := strings.Fields(out)
			for i, w := range f {
				if w == "Value:" && i+1 < len(f) {
					ip := f[i+1]
					if strings.Contains(ip, ".") {
						return ip
					}
				}
			}
		}
	}
	return "" // pas d'additions : l'appelant passe par le port-forward NAT
}

// EnsureNatSSH — sur une VM VirtualBox en NAT, l'IP invitée 10.0.2.x n'est
// pas joignable depuis l'hôte. La voie canonique : rediriger le port hôte
// (2222 par défaut) vers le 22 invité, et le port dashboard vers le 9090
// invité, puis la config SSH pointe 127.0.0.1. Idempotent : retire les
// règles existantes avant. modifyvm exige la VM ÉTEINTE — si elle tourne
// déjà, controlvm applique les mêmes règles À CHAUD (syntaxe nue `natpf1`,
// SANS `--` : le flag long est rejeté à chaud).
func (v *virtualbox) EnsureNatSSH(vbx, hostPort, dashPort string) error {
	name, err := v.ensureRegistered(vbx)
	if err != nil {
		return err
	}
	if hostPort == "" {
		hostPort = "2222"
	}
	if dashPort == "" {
		dashPort = "9090"
	}
	// ports libres AVANT de toucher la VM : un occupant existant (2e VM,
	// service) rendrait l'échec opaque.
	if busyPort(hostPort) {
		return fmt.Errorf("port hôte %s déjà occupé — libérez-le ou configurez nat_host_port", hostPort)
	}
	if busyPort(dashPort) {
		return fmt.Errorf("port hôte %s déjà occupé (dashboard) — libérez-le", dashPort)
	}
	running := false
	for _, r := range v.Running() {
		if SameVM(r, vbx) {
			running = true
		}
	}
	// verbe nu à chaud (controlvm), verbe --modifyvm à froid
	apply := func(hot bool, args ...string) error {
		if hot {
			full := append([]string{"controlvm", name}, args...)
			_, err := v.run(full...)
			return err
		}
		full := append([]string{"modifyvm", name}, args...)
		_, err := v.run(full...)
		return err
	}
	if !running {
		_ = apply(false, "--natpf1", "delete", "cgo-ssh")
		_ = apply(false, "--natpf1", "delete", "cgo-dashboard")
		if err := apply(false, "--natpf1", "cgo-ssh,tcp,,"+hostPort+",,22"); err != nil {
			return fmt.Errorf("règle SSH : %v", err)
		}
		if err := apply(false, "--natpf1", "cgo-dashboard,tcp,,"+dashPort+",,9090"); err != nil {
			return fmt.Errorf("règle dashboard : %v", err)
		}
		return nil
	}
	// VM allumée : modifyvm est rejeté (session verrouillée) → controlvm à
	// chaud, syntaxe NUE `natpf1` (le flag --natpf1 y est invalide).
	_ = apply(true, "natpf1", "delete", "cgo-ssh")
	if err := apply(true, "natpf1", "cgo-ssh,tcp,,"+hostPort+",,22"); err != nil {
		return fmt.Errorf("règle SSH à chaud : %v (VM verrouillée ?)", err)
	}
	_ = apply(true, "natpf1", "delete", "cgo-dashboard")
	if err := apply(true, "natpf1", "cgo-dashboard,tcp,,"+dashPort+",,9090"); err != nil {
		return fmt.Errorf("règle dashboard à chaud : %v", err)
	}
	return nil
}

// SameVM — même machine malgré les variantes de casse/séparateurs
// (chemin .vbox vs nom enregistré, slashes Windows/Unix).
func SameVM(a, b string) bool {
	norm := func(s string) string {
		return strings.ToLower(filepath.ToSlash(s))
	}
	return norm(a) == norm(b)
}

// busyPort — quelque chose écoute déjà ce port TCP local ?
func busyPort(port string) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

// NatHostAddr — l'adresse d'accès SSH quand la VM est en NAT.
func (v *virtualbox) NatHostAddr() string { return "127.0.0.1" }

// NetMode — attachement de la première NIC (showvminfo --machinereadable :
// nic1="nat"|"bridged"|"hostonly"|...) : nat → "nat", bridged → "ponté".
// VM inconnue de VirtualBox = "inconnu" (jamais d'erreur : affichage).
func (v *virtualbox) NetMode(vbx string) string {
	name := v.resolveName(vbx)
	if name == "" {
		return "inconnu"
	}
	out, err := v.run("showvminfo", name, "--machinereadable")
	if err != nil {
		return "inconnu"
	}
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if !strings.HasPrefix(ln, "nic1=") {
			continue
		}
		val := strings.Trim(strings.TrimPrefix(ln, "nic1="), "\"\r")
		switch val {
		case "bridged":
			return "ponté"
		case "nat":
			return "nat"
		case "hostonly":
			return "hôte-only"
		case "none", "":
			return "inconnu"
		default:
			return val
		}
	}
	return "inconnu"
}

// SetNetMode — nat|bridged|hostonly sur NIC1 (modifyvm exige la VM ÉTEINTE).
// oser demander à chaud serait mentir : VirtualBox rejette, on refuse avant.
func (v *virtualbox) SetNetMode(vbx, mode string) error {
	name, err := v.ensureRegistered(vbx)
	if err != nil {
		return err
	}
	for _, r := range v.Running() {
		if SameVM(r, vbx) {
			return fmt.Errorf("VM allumée — éteignez-la d'abord (modifyvm refuse à chaud)")
		}
	}
	if _, err := v.run("modifyvm", name, "--nic1", mode); err != nil {
		return fmt.Errorf("modifyvm --nic1 %s : %v", mode, err)
	}
	return nil
}

// ---- détection ----

// candidates retourne les chemins de binaire plausibles par OS.
// Windows 64 bits d'abord (natif), x86 ensuite : un vmrun/VBoxManage
// 32 bits fonctionne mais le 64 bits est préféré (mémoire, pilotes).
func candidates(tool string) []string {
	switch runtime.GOOS {
	case "windows":
		paths := []string{}
		if runtime.GOARCH == "amd64" {
			paths = append(paths,
				filepath.Join(`C:\Program Files\VMware\VMware Workstation`, tool),
				filepath.Join(`C:\Program Files\Oracle\VirtualBox`, tool),
			)
		}
		return append(paths,
			filepath.Join(`C:\Program Files (x86)\VMware\VMware Workstation`, tool),
			filepath.Join(`C:\Program Files\Oracle\VirtualBox`, tool),
			tool, // dans le PATH
		)
	case "darwin":
		return []string{
			filepath.Join("/Applications/VMware Fusion.app/Contents/Library", tool),
			"/usr/local/bin/" + tool, "/opt/homebrew/bin/" + tool, tool,
		}
	default:
		return []string{"/usr/bin/" + tool, "/usr/local/bin/" + tool, tool}
	}
}

func findExe(tool string) string {
	for _, c := range candidates(tool) {
		if c == tool {
			if p, err := exec.LookPath(tool); err == nil {
				return p
			}
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

// Detect — hyperviseurs disponibles dans l'ordre de préférence (VMware d'abord).
func Detect() []Hypervisor {
	var out []Hypervisor
	e := findExe("vmrun.exe")
	if e == "" {
		e = findExe("vmrun")
	}
	if e != "" {
		out = append(out, &vmware{exe: e})
	}
	vb := findExe("VBoxManage.exe")
	if vb == "" {
		vb = findExe("VBoxManage")
	}
	if vb != "" {
		out = append(out, &virtualbox{exe: vb})
	}
	return out
}

// Primary — le premier hyperviseur détecté, nil si aucun.
func Primary() Hypervisor {
	h := Detect()
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

// roots — disques à scanner selon l'OS (Git-Bash : /c, /d…).
func roots() []string {
	if runtime.GOOS == "windows" {
		// énumération réelle des lecteurs
		var out []string
		for c := 'A'; c <= 'Z'; c++ {
			d := string(c) + `:\`
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				out = append(out, d)
			}
		}
		return out
	}
	if dirs, err := os.ReadDir("/"); err == nil {
		var out []string
		for _, d := range dirs {
			n := d.Name()
			if len(n) == 1 && n[0] >= 'a' && n[0] <= 'z' && d.IsDir() {
				out = append(out, "/"+n)
			}
		}
		return out
	}
	return []string{"/"}
}

// extraRoots — emplacements VMs hors conventions, scannés à profondeur
// bornée (5) dans TOUS les modes : ~/vmware (défaut VMware), /media et /mnt
// (disques externes/2e disque). Séparés de roots() : un scan profond à 100
// sur un disque externe de 2 To prendrait des minutes à chaque ensure.
func extraRoots() []string {
	if runtime.GOOS == "windows" {
		return nil
	}
	var out []string
	if h, err := os.UserHomeDir(); err == nil {
		out = append(out, filepath.Join(h, "vmware"))
	}
	if entries, err := os.ReadDir("/media"); err == nil {
		for _, u := range entries {
			if !u.IsDir() {
				continue
			}
			sub, _ := os.ReadDir(filepath.Join("/media", u.Name()))
			for _, m := range sub {
				if m.IsDir() {
					out = append(out, filepath.Join("/media", u.Name(), m.Name()))
				}
			}
		}
	}
	if entries, err := os.ReadDir("/mnt"); err == nil {
		for _, m := range entries {
			if m.IsDir() {
				out = append(out, filepath.Join("/mnt", m.Name()))
			}
		}
	}
	return out
}

// ScanVMs — recherche .vmx/.vbox : d'abord les conventions D:/VMs, C:/VMs
// (profondeur 2), puis la racine des disques (profondeur 3, --deep pour tout).
// Retourne les chemins canoniques triés.
func ScanVMs(deep bool) []string {
	seen := map[string]bool{}
	var out []string

	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}

	// 1) inventaire hyperviseur (le plus fiable)
	for _, h := range Detect() {
		for _, r := range h.Running() {
			add(r)
		}
	}

	// 2) conventions par disque : <drive>/VMs, <drive>/vms, <home>/VirtualBox VMs
	scanDir := func(dir string, depth int) {
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			return
		}
		_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if depth--; depth < 0 {
					return filepath.SkipDir
				}
				return nil
			}
			n := strings.ToLower(d.Name())
			if strings.HasSuffix(n, ".vmx") || strings.HasSuffix(n, ".vbox") {
				add(p)
			}
			return nil
		})
	}

	for _, r := range roots() {
		scanDir(filepath.Join(r, "VMs"), 2)
		scanDir(filepath.Join(r, "vms"), 2)
	}
	scanDir(filepath.Join(home(), "VirtualBox VMs"), 2)

	// 3) racine des disques en peu profond (ou tout si deep)
	depth := 3
	if deep {
		depth = 100
	}
	for _, r := range roots() {
		scanDir(r, depth)
	}

	// 4) emplacements VMs hors conventions — profondeur bornée 5, deep ou non
	for _, r := range extraRoots() {
		scanDir(r, 5)
	}

	sort.Strings(out)
	return out
}

// Pick — choisit la VM : nom exact si précisé, sinon l'unique trouvée,
// sinon "" (ambiguïté — l'appelant liste les candidates).
func Pick(vms []string, name string) string {
	if name != "" {
		for _, v := range vms {
			if strings.TrimSuffix(filepath.Base(v), filepath.Ext(v)) == name || filepath.Base(v) == name {
				return v
			}
		}
	}
	if len(vms) == 1 {
		return vms[0]
	}
	return ""
}

func home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return h
}
