// Package vm — détection et pilotage des hyperviseurs, pur Go multi-OS.
// Remplace le scan bash historique (appels shell externes silencieux hors
// Windows) : la détection enumère les disques nativement et cherche
// .vmx/.vbox avec la convention D:/VMs, C:/VMs (racine historique tolérée).
package vm

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
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
	// GuestIP interroge l'IP invitée d'une VM allumée ("" si inconnue).
	GuestIP(vmx string) string
}

// ---- VMware ----

type vmware struct{ exe string }

func (v *vmware) Name() string { return "vmware" }
func (v *vmware) Exe() string  { return v.exe }

func (v *vmware) run(args ...string) (string, error) {
	cmd := exec.Command(v.exe, args...)
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

func (v *vmware) GuestIP(vmx string) string {
	out, err := v.run("getGuestIPAddress", vmx, "-wait", "5")
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(strings.Split(out, "\n")[0])
	if !strings.Contains(ip, ".") {
		return ""
	}
	return ip
}

// ---- VirtualBox ----

type virtualbox struct{ exe string }

func (v *virtualbox) Name() string { return "virtualbox" }
func (v *virtualbox) Exe() string  { return v.exe }

func (v *virtualbox) run(args ...string) (string, error) {
	cmd := exec.Command(v.exe, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (v *virtualbox) Running() []string {
	out, err := v.run("list", "runningvms")
	if err != nil {
		return nil
	}
	var vms []string
	for _, ln := range strings.Split(out, "\n") {
		if i := strings.Index(ln, ".vbox"); i > 0 {
			if j := strings.LastIndex(ln[:i], "\""); j >= 0 {
				vms = append(vms, strings.Trim(ln[j:], "\""))
			}
		}
	}
	return vms
}

func (v *virtualbox) Start(vbx string) error {
	name := strings.TrimSuffix(filepath.Base(vbx), ".vbox")
	_, err := v.run("startvm", name, "--type", "headless")
	return err
}

func (v *virtualbox) GuestIP(string) string { return "" } // VBoxManage guestproperty — hors périmètre v1.1

// ---- détection ----

// candidates retourne les chemins de binaire plausibles par OS.
func candidates(tool string) []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(`C:\Program Files (x86)\VMware\VMware Workstation`, tool),
			filepath.Join(`C:\Program Files\VMware\VMware Workstation`, tool),
			filepath.Join(`C:\Program Files\Oracle\VirtualBox`, tool),
			tool, // dans le PATH
		}
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
