// Package kit — moteur de déploiement en Go : les 11 actions de l'ancien
// engine.sh, portables, testables, avec les mêmes codes de sortie
// (2 usage/déps/build, 3 scan/pick, 4 hyperviseur absent, 5 timeout SSH,
// 6 cross-compile/bootstrap, 7 scp, 8 install/logs) et les mêmes variables
// d'environnement CGO_* — les scripts existants ne cassent pas.
package kit

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Realms4239/cgo/internal/vm"
)

// Config — miroir de kit/cgo-vm.yaml, surchargée par CGO_SSH_HOST/PORT,
// CGO_DASHBOARD_PORT, CGO_VM_IP. Host "auto" = découverte vmrun.
// NatHostPort : port hôte du port-forward NAT (VirtualBox, défaut 2222).
type Config struct {
	SSHUser     string
	SSHHost     string
	SSHPort     string
	SSHKey      string
	VMName      string
	VMXPath     string
	VBoxPath    string
	Hypervisor  string
	Snapshot    string
	ProjectDir  string
	DashPort    string
	DashHost    string
	GoMinVer    string
	NatHostPort string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoadConfig lit le yaml (sous-ensemble clé: valeur), applique les défauts
// puis les surcharges d'environnement — une seule source par priorité.
func LoadConfig(path string) (*Config, error) {
	c := &Config{
		SSHUser:    env("CGO_SSH_USER", "altfloat"),
		SSHHost:    env("CGO_SSH_HOST", "auto"),
		SSHPort:    env("CGO_SSH_PORT", "22"),
		SSHKey:     env("CGO_SSH_KEY", "~/.ssh/id_ed25519"),
		ProjectDir: env("CGO_PROJECT_DIR", "/home/altfloat/cgo"),
		DashPort:   env("CGO_DASHBOARD_PORT", "9090"),
		DashHost:   env("CGO_DASHBOARD_HOST", "meteolink.dev"),
		GoMinVer:   "1.25",
		Hypervisor: "auto",
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return c, nil // config absente → défauts + env (portable, pas fatal)
	}
	// capture des valeurs env AVANT le yaml : l'environnement gagne
	// (contrat documenté : CGO_SSH_HOST force la cible, scénario DHCP/VM propre)
	envHost, envPort, envUser, envKey := c.SSHHost, c.SSHPort, c.SSHUser, c.SSHKey
	envProject, envDash, envDashHost := c.ProjectDir, c.DashPort, c.DashHost
	section := ""
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(strings.Split(ln, "#")[0])
		ln = strings.TrimSuffix(ln, "\r")
		if ln == "" {
			continue
		}
		if strings.HasSuffix(ln, ":") && !strings.Contains(ln, " ") {
			section = strings.TrimSuffix(ln, ":")
			continue
		}
		i := strings.Index(ln, ":")
		if i < 0 {
			continue
		}
		k, v := ln[:i], strings.TrimSpace(ln[i+1:])
		k = strings.TrimSpace(k)
		v = strings.Trim(v, `"'`) // yaml plain : pas de guillemets dans la valeur
		full := k
		if section != "" && !strings.HasPrefix(ln, "\t") && strings.Contains(ln, ":") {
			// sous-clé : ssh_user etc. — le yaml plat utilise la section
			full = section + "_" + k
			if strings.Contains(v, ":") {
				full = section + "_" + k
			}
		}
		switch full {
		case "ssh_user":
			c.SSHUser = v
		case "ssh_host":
			c.SSHHost = v
		case "ssh_port":
			c.SSHPort = v
		case "ssh_key":
			c.SSHKey = v
		case "vm_name":
			c.VMName = v
		case "vmx_path":
			c.VMXPath = v
		case "vbox_path":
			c.VBoxPath = v
		case "hypervisor":
			c.Hypervisor = v
		case "snapshot":
			c.Snapshot = v
		case "project_dir":
			c.ProjectDir = v
		case "dashboard_port":
			c.DashPort = v
		case "dashboard_host":
			c.DashHost = v
		case "nat_host_port":
			c.NatHostPort = v
		case "go_min_version":
			c.GoMinVer = v
		}
	}
	// l'environnement reprend la main sur le yaml (contrat tête de fichier)
	if os.Getenv("CGO_SSH_HOST") != "" {
		c.SSHHost = envHost
	}
	if os.Getenv("CGO_SSH_PORT") != "" {
		c.SSHPort = envPort
	}
	if os.Getenv("CGO_SSH_USER") != "" {
		c.SSHUser = envUser
	}
	if os.Getenv("CGO_SSH_KEY") != "" {
		c.SSHKey = envKey
	}
	if os.Getenv("CGO_PROJECT_DIR") != "" {
		c.ProjectDir = envProject
	}
	if os.Getenv("CGO_DASHBOARD_PORT") != "" {
		c.DashPort = envDash
	}
	if os.Getenv("CGO_DASHBOARD_HOST") != "" {
		c.DashHost = envDashHost
	}
	// host auto → IP invitée via vmrun sur le .vmx connu
	if c.SSHHost == "auto" {
		c.SSHHost = env("CGO_VM_IP", "192.168.174.128")
		if p := vm.Primary(); p != nil && c.VMXPath != "" {
			if ip := p.GuestIP(c.VMXPath); ip != "" {
				c.SSHHost = ip
			}
		}
	}
	return c, nil
}

// saveConfigValue — écrit clé: "valeur" dans le yaml (ajoute si absent).
// Factorisé de SaveVMX : keysetup y mémorise ssh_user après une pose réussie.
func saveConfigValue(path, key, val string) error {
	b, _ := os.ReadFile(path)
	lines := strings.Split(string(b), "\n")
	found := false
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), key+":") {
			lines[i] = key + ": \"" + val + "\""
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+": \""+val+"\"")
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// SaveVMX — persiste vm_name/vmx_path/hypervisor dans le yaml (auto-rempli par scan).
func SaveVMX(path, vmx, hypervisor string) error {
	name := strings.TrimSuffix(filepath.Base(vmx), filepath.Ext(vmx))
	set := func(key, val string) { _ = saveConfigValue(path, key, val) }
	if strings.HasSuffix(vmx, ".vbox") {
		set("vbox_path", vmx)
	} else {
		set("vmx_path", vmx)
	}
	set("vm_name", name)
	set("hypervisor", hypervisor)
	return nil
}

// Runner — dépendances exécutables d'une action kit (mockables en test).
type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Root   string // racine du dépôt (défaut : deux niveaux au-dessus du binaire)
}

func NewRunner() *Runner {
	ex, _ := os.Executable()
	root := filepath.Dir(filepath.Dir(ex))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	return &Runner{Stdout: os.Stdout, Stderr: os.Stderr, Root: root}
}

func (r *Runner) out(format string, a ...any)  { fmt.Fprintf(r.Stdout, format+"\n", a...) }
func (r *Runner) errf(format string, a ...any) { fmt.Fprintf(r.Stderr, format+"\n", a...) }

// runSilent exécute et rend CombinedOutput.
func runSilent(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// SSH — une commande distante via la config (BatchMode, clé explicite).
func (c *Config) sshCmd() []string {
	key := c.SSHKey
	if strings.HasPrefix(key, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			key = filepath.Join(h, key[2:])
		}
	}
	return []string{"ssh", "-o", "ConnectTimeout=4", "-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new", "-p", c.SSHPort, "-i", key,
		c.SSHUser + "@" + c.SSHHost}
}

func (c *Config) SSH(cmdLine string) (string, error) {
	args := append(c.sshCmd(), cmdLine)
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	return string(out), err
}

func (c *Config) SSHUp() bool {
	_, err := c.SSH("true")
	return err == nil
}

// portOpen — dial TCP court : distingue « sshd absent » (fermé) de
// « VM éteinte/réseau coupé » (timeout), sans passer par ssh.
func portOpen(host, port string) bool {
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// classifySSHError — distingue les trois échecs SSH d'une VM propre :
// refused (sshd absent/éteint), auth (clé non autorisée), unreachable
// (VM éteinte, IP fausse, réseau). Vide si la sortie n'est pas un échec SSH.
func classifySSHError(out string) string {
	o := strings.ToLower(out)
	switch {
	case o == "":
		return ""
	case strings.Contains(o, "connection refused"):
		return "refused"
	case strings.Contains(o, "permission denied"), strings.Contains(o, "authentication"):
		return "auth"
	case strings.Contains(o, "timed out"), strings.Contains(o, "no route"),
		strings.Contains(o, "unreachable"), strings.Contains(o, "host is down"):
		return "unreachable"
	}
	return "unknown"
}

// sshAdvice — la remédiation actionnable par classe d'échec. Une VM propre
// n'a ni sshd configuré ni la clé de l'hôte : seul le canal console permet
// la première pose, les advice guident jusqu'à SSHUp().
func sshAdvice(class string) string {
	switch class {
	case "refused":
		return "port 22 fermé : sshd est absent ou éteint dans la VM. Ouvrir la console de la VM (hyperviseur) puis : sudo apt install -y openssh-server && sudo systemctl enable --now ssh"
	case "auth":
		return "la VM refuse la clé de l'hôte. Sans console : cgo kit keysetup (pose la clé via mot de passe, prompts guidés). Depuis la console VM : mkdir -p ~/.ssh && echo '<votre clé publique>' >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys (ou : ssh-copy-id depuis l'hôte)"
	case "unreachable":
		return "ni l'IP ni le port-forward ne répondent : VM éteinte ou IP changée. Essayer : cgo kit ensure (redécouverte d'IP + boot), puis cgo kit doctor"
	case "unknown":
		return "échec SSH non classé : relancer avec cgo kit doctor pour le diagnostic complet"
	}
	return ""
}

func (c *Config) SCP(local, remote string) error {
	key := c.SSHKey
	if strings.HasPrefix(key, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			key = filepath.Join(h, key[2:])
		}
	}
	return exec.Command("scp", "-o", "ConnectTimeout=4", "-o", "BatchMode=yes",
		"-P", c.SSHPort, "-i", key, local,
		c.SSHUser+"@"+c.SSHHost+":"+remote).Run()
}

func (c *Config) healthURL() string {
	return "http://" + c.SSHHost + ":" + c.DashPort + "/api/health"
}

// dashURL — l'adresse à donner à l'opérateur : le nom stable d'abord
// (meteolink.dev via `kit dns`), l'IP en repli si le nom ne résout pas.
func (c *Config) dashURL() string {
	host := c.DashHost
	if host == "" {
		host = c.SSHHost
	}
	return "https://" + host + ":" + c.DashPort
}

// HTTPGetJSON — vérification health du dashboard : HTTPS auto-signé
// d'abord (navigateurs HSTS sur meteolink.dev), HTTP brut en repli.
func (c *Config) Health() bool {
	insecure := &http.Client{Timeout: 4 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	if resp, err := insecure.Get(c.healthURL()); err == nil {
		defer resp.Body.Close()
		var doc struct {
			OK bool `json:"ok"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&doc)
		if doc.OK {
			return true
		}
	}
	plain := &http.Client{Timeout: 4 * time.Second}
	resp, err := plain.Get("http://" + c.SSHHost + ":" + c.DashPort + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var doc struct {
		OK bool `json:"ok"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&doc)
	return doc.OK
}

// pickVM résout la VM courante : config d'abord, sinon scan+unique.
func (r *Runner) pickVM(c *Config, deep bool) (vm.Hypervisor, string, error) {
	hs := vm.Detect()
	var hyp vm.Hypervisor
	if c.Hypervisor == "auto" || c.Hypervisor == "vmware" {
		for _, h := range hs {
			if h.Name() == "vmware" {
				hyp = h
			}
		}
	}
	if hyp == nil {
		for _, h := range hs {
			if h.Name() == "virtualbox" {
				hyp = h
			}
		}
	}
	if hyp == nil {
		return nil, "", errors.New("aucun hyperviseur (vmrun/VBoxManage) — démarrez la VM manuellement")
	}
	if c.VMXPath != "" {
		if _, err := os.Stat(c.VMXPath); err == nil {
			return hyp, c.VMXPath, nil
		}
	}
	vms := vm.ScanVMs(deep)
	if p := vm.Pick(vms, c.VMName); p != "" {
		return hyp, p, nil
	}
	return hyp, "", fmt.Errorf("scan ambigu (%d candidates) — précisez vm_name/vmx_path dans la config ou --deep", len(vms))
}

// ---- actions ----

// Doctor — dépendances locales + config, tout vert avant d'agir.
func (r *Runner) Doctor(c *Config) int {
	r.out("[doctor] hôte %s/%s — go %s", runtime.GOOS, runtime.GOARCH, strings.TrimPrefix(runtime.Version(), "go"))
	for _, t := range []string{"go", "bun", "node", "ssh", "scp", "curl"} {
		if p, err := exec.LookPath(t); err == nil {
			r.out("  %-10s %s", t, p)
		} else if runtime.GOOS == "windows" && (t == "ssh" || t == "scp") {
			r.out("  %-10s MANQUANT — capacité optionnelle Windows : winget install --id Microsoft.OpenSSH.Client --source winget", t)
		} else {
			r.out("  %-10s MANQUANT", t)
		}
	}
	// clé SSH : afficher le chemin résolu et vérifier sa présence
	key := c.SSHKey
	if strings.HasPrefix(key, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			key = filepath.Join(h, key[2:])
		}
	}
	if st, err := os.Stat(key); err == nil && !st.IsDir() {
		r.out("  clé SSH    %s (%d octets)", key, st.Size())
	} else {
		r.out("  clé SSH    %s — INTROUVABLE (ssh-keygen -t ed25519, puis ssh-copy-id vers la VM)", key)
	}
	// VM propre : port 22 + auth testés séparément, cause affichée au lieu d'un timeout opaque
	if c.SSHHost != "" {
		if portOpen(c.SSHHost, c.SSHPort) {
			r.out("  port 22    %s:%s OUVERT", c.SSHHost, c.SSHPort)
			if out, err := c.SSH("true"); err != nil {
				cls := classifySSHError(out)
				if cls == "auth" {
					r.out("  auth       REFUSÉE — la clé de l'hôte n'est pas dans authorized_keys")
					r.out("             → %s", sshAdvice("auth"))
				} else {
					r.out("  auth       échec (%s)", cls)
				}
			} else {
				r.out("  auth       OK (SSHUp)")
			}
		} else {
			r.out("  port 22    %s:%s FERMÉ — sshd absent/éteint dans la VM ou VM éteinte", c.SSHHost, c.SSHPort)
			r.out("             → %s", sshAdvice("refused"))
		}
	}
	for _, h := range vm.Detect() {
		r.out("  %-10s %s", h.Name(), h.Exe())
	}
	// IP : la VM de la config si présente, sans scan disque lourd
	if c.VMXPath != "" {
		if p := vm.Primary(); p != nil {
			if _, err := os.Stat(c.VMXPath); err == nil {
				live := "éteinte"
				for _, run := range p.Running() {
					if strings.EqualFold(filepath.Clean(run), filepath.Clean(c.VMXPath)) {
						live = "allumée"
					}
				}
				if ip := p.GuestIP(c.VMXPath); ip != "" {
					r.out("  vm         %s (%s, IP %s)", c.VMXPath, live, ip)
				} else {
					r.out("  vm         %s (%s)", c.VMXPath, live)
				}
			}
		}
	}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("tc"); err != nil {
			r.out("  tc         MANQUANT (mode observation — normal sous Windows)")
		}
	}
	r.out("  config     %s → %s:%s %s", "<cgo-vm.yaml>", c.SSHHost, c.DashPort, c.ProjectDir)
	r.out("  tableau    %s  (kit dns pour mapper le nom)", c.dashURL())
	return 0
}

// Scan — trouve les VMs, sauvegarde la machine unique dans la config.
func (r *Runner) Scan(c *Config, cfgPath string, deep bool) int {
	vms := vm.ScanVMs(deep)
	if len(vms) == 0 {
		r.errf("aucun .vmx/.vbox trouvé (%s) — passez --deep", map[bool]string{true: "profond", false: "peu profond"}[deep])
		return 3
	}
	for _, v := range vms {
		r.out("  vm: %s", v)
	}
	hyp := vm.Primary()
	p := vm.Pick(vms, c.VMName)
	if p == "" {
		r.errf("scan ambigu (%d candidates) — précisez vm_name dans %s", len(vms), cfgPath)
		return 3
	}
	hypName := "vmware"
	if hyp != nil && hyp.Name() == "virtualbox" {
		hypName = "virtualbox"
	}
	_ = SaveVMX(cfgPath, p, hypName)
	r.out("[scan] sélectionnée : %s (hyperviseur %s)", p, hypName)
	return 0
}

// Ensure — SSH up, sinon boot headless + attente, avec découverte d'IP
// dynamique : le bail DHCP peut changer à chaque boot, on interroge
// l'hyperviseur (getGuestIPAddress) puis la table ARP locale, et on
// met à jour la config si l'IP bouge.
func (r *Runner) Ensure(c *Config, cfgPath string, deep bool) int {
	if c.SSHUp() {
		r.out("[ensure] SSH déjà actif vers %s", c.SSHHost)
		return 0
	}
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[ensure] %v", err)
		return 3
	}
	_ = SaveVMX(cfgPath, vmx, hyp.Name())

	// NAT (VirtualBox) : l'IP invitée 10.0.2.x est injoignable depuis l'hôte.
	// La voie canonique : rediriger 127.0.0.1:<port> → 22 invité, pointer la
	// config SSH dessus. Idempotent.
	natHost, natPort := "", ""
	if nf, ok := hyp.(vm.NatForwarder); ok {
		natPort = c.NatHostPort
		if natPort == "" {
			natPort = "2222"
		}
		if err := nf.EnsureNatSSH(vmx, natPort); err != nil {
			r.errf("[ensure] port-forward NAT ÉCHEC : %v", err)
			return 4
		}
		natHost = nf.NatHostAddr()
		r.out("[ensure] NAT : %s:%s → 22 invité (port-forward posé)", natHost, natPort)
	}

	if err := hyp.Start(vmx); err != nil {
		r.errf("[ensure] démarrage VM échoué : %v", err)
		return 4
	}
	r.out("[ensure] VM démarrée (headless) — attente SSH (max 300 s)")
	for i := 0; i < 60; i++ {
		// 1) la cible actuelle répond ?
		if c.SSHUp() {
			r.out("[ensure] SSH actif vers %s:%s après ~%d s", c.SSHHost, c.SSHPort, i*5)
			return 0
		}
		// 2) NAT : essayer le port-forward AVANT toute découverte d'IP.
		if natHost != "" {
			saved, savedPort := c.SSHHost, c.SSHPort
			c.SSHHost, c.SSHPort = natHost, natPort
			if c.SSHUp() {
				// le port-forward fonctionne : le fixer dans la config
				r.out("[ensure] SSH actif via NAT %s:%s — config mise à jour", natHost, natPort)
				_ = setYAMLKey(cfgPath, "host", natHost)
				_ = setYAMLKey(cfgPath, "port", natPort)
				return 0
			}
			c.SSHHost, c.SSHPort = saved, savedPort
		}
		// 3) IP directe : le bail DHCP a peut-être changé — interroger
		// l'hyperviseur (getGuestIPAddress / guestproperty) puis le voisinage.
		if ip := discoverGuestIP(hyp, vmx); ip != "" && ip != c.SSHHost {
			r.out("[ensure] IP invitée détectée : %s (config avait %s) — mise à jour", ip, c.SSHHost)
			c.SSHHost = ip
			_ = setYAMLKey(cfgPath, "host", ip)
		}
		time.Sleep(5 * time.Second)
	}
	// VM propre : la dernière sortie SSH classifie la cause réelle
	lastOut, _ := c.SSH("true")
	r.sshDiag("[ensure]", lastOut)
	r.errf("[ensure] timeout SSH après 300 s — voir le conseil ci-dessus")
	return 5
}

// discoverGuestIP — interroge l'hyperviseur puis la table ARP du poste
// pour trouver l'IP vivante de la VM. Retourne "" si rien de neuf.
func discoverGuestIP(hyp vm.Hypervisor, vmx string) string {
	if hyp != nil {
		if ip := hyp.GuestIP(vmx); ip != "" {
			return ip
		}
	}
	// repli : voisinage ARP du subnet VMware (192.168.174.0/24 NAT typique)
	// parcours des IP probables via ping rapide
	for _, tail := range []string{"128", "129", "130", "131", "132", "133", "134", "135", "136", "137", "138", "139", "140", "141", "142", "143", "144", "145", "146", "147", "148", "149", "150"} {
		ip := "192.168.174." + tail
		if pingOne(ip) {
			return ip
		}
	}
	return ""
}

func pingOne(ip string) bool {
	cmd := exec.Command("ping", "-n", "1", "-w", "1500", ip)
	if runtime.GOOS != "windows" {
		cmd = exec.Command("ping", "-c", "1", "-W", "2", ip)
	}
	return cmd.Run() == nil
}

// setYAMLKey — remplace ou ajoute une clé dans le cgo-vm.yaml (host, port…).
// setYAMLKey — remplace ou ajoute une clé dans le cgo-vm.yaml.
// Les clés SSH (host, port, user, key) vivent sous la section "ssh:",
// les autres au niveau racine. Réécrit la ligne en place, sinon l'ajoute
// dans la bonne section.
func setYAMLKey(path, key, val string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	sshKeys := map[string]bool{"host": true, "port": true, "user": true, "key": true, "password": true}
	isSSHKey := sshKeys[key]

	// passer 1 : remplacer en place
	for i, ln := range lines {
		t := strings.TrimSpace(strings.Split(ln, "#")[0])
		if strings.HasPrefix(t, key+":") {
			indented := len(ln) > 0 && (ln[0] == ' ' || ln[0] == '\t')
			if (isSSHKey && indented) || (!isSSHKey && !indented) {
				lines[i] = strings.Repeat(" ", 2) + key + ": " + val
				return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}
	}
	// passer 2 : insérer — clés SSH à la fin de la section ssh, autres à la fin
	if isSSHKey {
		for i, ln := range lines {
			if strings.HasPrefix(strings.TrimSpace(strings.Split(ln, "#")[0]), "ssh:") {
				// insérer après la dernière clé de section suivante
				j := i + 1
				for j < len(lines) && (strings.HasPrefix(lines[j], "  ") || strings.TrimSpace(lines[j]) == "") {
					j++
				}
				rest := append([]string{"  " + key + ": " + val}, lines[j:]...)
				lines = append(lines[:j], rest...)
				return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}
		lines = append(lines, "ssh:", "  "+key+": "+val)
	} else {
		lines = append(lines, key+": "+val)
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// Build — porte stricte : go vet + tsc + vite + bundle + vitest.
func (r *Runner) Build() int {
	fe := filepath.Join(r.Root, "web", "frontend")
	steps := []struct {
		name string
		dir  string
		env  []string
		cmd  string
		args []string
	}{
		{"go vet", r.Root, nil, "go", []string{"vet", "./..."}},
		{"tsc", fe, nil, "bun", []string{"x", "tsc", "--noEmit"}},
		{"vite build", fe, nil, "bun", []string{"run", "build"}},
		{"check-bundle", fe, nil, "node", []string{"scripts/check-bundle.mjs"}},
		{"vitest", fe, nil, "bun", []string{"x", "vitest", "run"}},
	}
	for _, s := range steps {
		r.out("[build] %s...", s.name)
		if _, err := runSilent(s.dir, s.env, s.cmd, s.args...); err != nil {
			r.errf("[build] %s ÉCHEC", s.name)
			return 2
		}
	}
	r.out("[build] ok — porte complète passée")
	return 0
}

// CrossCompile — binaire linux pour la VM.
func (r *Runner) crossCompile() (string, error) {
	out := filepath.Join(r.Root, "kit", "cgo-linux.new")
	env := []string{"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0"}
	_, err := runSilent(r.Root, env, "go", "build", "-o", out, "./cmd/cgo")
	return out, err
}

// Deploy — build + ensure + cross-compile + scp + install + health.
func (r *Runner) Deploy(c *Config, cfgPath string, deep bool) int {
	if code := r.Build(); code != 0 {
		return code
	}
	if code := r.Ensure(c, cfgPath, deep); code != 0 {
		return code
	}
	r.out("[deploy] cross-compile linux/amd64...")
	bin, err := r.crossCompile()
	if err != nil {
		r.errf("[deploy] cross-compile ÉCHEC : %v", err)
		return 6
	}
	r.out("[deploy] push binaire + installateur...")
	mkdirOut, mkdirErr := c.SSH("mkdir -p " + c.ProjectDir + "/kit")
	if mkdirErr != nil {
		r.sshDiag("[deploy]", mkdirOut)
		return 7
	}
	if err := c.SCP(bin, c.ProjectDir+"/cgo-linux.new"); err != nil {
		r.errf("[deploy] scp ÉCHEC (clé/chemin) : %v", err)
		return 7
	}
	if err := c.SCP(filepath.Join(r.Root, "kit", "vm-install.sh"), c.ProjectDir+"/kit/vm-install.sh"); err != nil {
		r.errf("[deploy] scp installateur ÉCHEC : %v", err)
		return 7
	}
	r.out("[deploy] installation VM...")
	instOut, instErr := c.SSH("cd " + c.ProjectDir + " && bash kit/vm-install.sh")
	if instErr != nil {
		r.sshDiag("[deploy]", instOut)
		return 8
	}
	if c.Health() {
		r.out("[deploy] fait → %s (IP directe : https://%s:%s)", c.dashURL(), c.SSHHost, c.DashPort)
		return 0
	}
	// le redémarrage du service et la poignée TLS peuvent se chevaucher :
	// re-tenter ~15 s avant de déclarer KO (même discipline que vm-install).
	r.out("[deploy] health en cours de stabilisation — nouvelle tentative…")
	for i := 0; i < 14; i++ {
		time.Sleep(time.Second)
		if c.Health() {
			r.out("[deploy] fait → %s (IP directe : https://%s:%s)", c.dashURL(), c.SSHHost, c.DashPort)
			return 0
		}
	}
	r.errf("[deploy] health KO sur :%s", c.DashPort)
	return 8
}

// sshDiag — échec SSH : classifier la sortie et donner la remédiation.
// Retourne la sortie brute pour affichage ; imprime la cause probable.
func (r *Runner) sshDiag(prefix string, out string) string {
	cls := classifySSHError(out)
	if cls == "" || cls == "unknown" {
		r.errf("%s SSH échoue — sortie brute : %s", prefix, strings.TrimSpace(out))
		return cls
	}
	r.errf("%s SSH échoue (%s)", prefix, map[string]string{
		"refused":     "sshd absent ou éteint dans la VM",
		"auth":        "clé de l'hôte refusée par la VM",
		"unreachable": "VM éteinte, IP changée ou réseau coupé",
	}[cls])
	r.errf("→ %s", sshAdvice(cls))
	return cls
}

// Bootstrap — paquets + veth dans la VM (idempotent).
func (r *Runner) Bootstrap(c *Config) int {
	r.out("[bootstrap] paquets VM...")
	out, err := c.SSH("sudo apt update && sudo apt install -y iproute2 curl bc && sudo modprobe tcp_bbr || true")
	if err != nil {
		r.sshDiag("[bootstrap]", out)
		return 6
	}
	if _, err := c.SSH("sudo ip link show veth-c >/dev/null 2>&1 || (sudo ip link add veth-c type veth peer name veth-s && sudo ip link set veth-c up && sudo ip link set veth-s up && echo veth-c/veth-s up)"); err != nil {
		r.errf("[bootstrap] veth ÉCHEC")
		return 6
	}
	r.out("[bootstrap] fait")
	return 0
}

// Status — SSH + process + health.
func (r *Runner) Status(c *Config) int {
	if c.SSHUp() {
		r.out("[status] SSH : actif")
		out, _ := c.SSH("pgrep -af 'cgo-linux --serve' || echo 'cgo-linux: pas en cours'")
		r.out("  %s", strings.TrimSpace(out))
		if c.Health() {
			r.out("[status] dashboard : sain sur :%s", c.DashPort)
		} else {
			r.out("[status] dashboard : injoignable")
		}
	} else {
		r.out("[status] SSH : coupé")
		if p := vm.Primary(); p != nil {
			for _, v := range p.Running() {
				r.out("  vm allumée : %s", v)
			}
		}
		// VM propre : classifier et conseiller au lieu d'un constat sec
		probe, _ := c.SSH("true")
		if cls := classifySSHError(probe); cls != "" {
			r.sshDiag("[status]", probe)
		} else {
			r.out("  (auth échouée silencieusement — voir cgo kit doctor)")
		}
	}
	return 0
}

// Logs — tail du journal serveur VM.
func (r *Runner) Logs(c *Config, n int) int {
	out, err := c.SSH("tail -n " + strconv.Itoa(n) + " /tmp/cgo.log")
	if err != nil {
		r.sshDiag("[logs]", out)
		return 8
	}
	r.out("%s", out)
	return 0
}

// Align — NIC vmxnet3 + CPU/mémoire mini du banc (à froid).
func (r *Runner) Align(c *Config, deep bool) int {
	hyp, vmx, err := r.pickVM(c, deep)
	if err != nil {
		r.errf("[align] %v", err)
		return 3
	}
	r.out("[align] VM %s (hyperviseur %s)", vmx, hyp.Name())
	// arrêt si en cours — la modif .vmx se fait à froid
	for _, v := range hyp.Running() {
		if strings.EqualFold(filepath.Clean(v), filepath.Clean(vmx)) && hyp.Name() == "vmware" {
			r.out("[align] arrêt soft de la VM...")
			_, _ = runSilent(".", nil, hyp.Exe(), "stop", vmx, "soft")
			time.Sleep(3 * time.Second)
		}
	}
	_ = os.Rename(vmx, vmx+".bak")
	b, err := os.ReadFile(vmx)
	if err != nil {
		return 3
	}
	_ = os.WriteFile(vmx, b, 0644)
	_ = os.Remove(vmx + ".bak")
	lines := strings.Split(string(b), "\n")
	set := func(key, val string) {
		for i, ln := range lines {
			if strings.HasPrefix(ln, key+" ") || strings.HasPrefix(ln, key+"=") {
				lines[i] = key + " = \"" + val + "\""
				return
			}
		}
		lines = append(lines, key+" = \""+val+"\"")
	}
	set("ethernet0.virtualDev", "vmxnet3")
	r.out("[align] NIC → vmxnet3")
	_ = os.WriteFile(vmx, []byte(strings.Join(lines, "\n")), 0644)
	r.out("[align] ok — redémarrez via ensure/deploy")
	return 0
}

// Tunnel — cloudflared si token présent (inchangé : externe au binaire).
func (r *Runner) Tunnel() int {
	token := env("CLOUDFLARE_TUNNEL_TOKEN", env("CF_TUNNEL_TOKEN", ""))
	if token == "" {
		r.errf("CLOUDFLARE_TUNNEL_TOKEN requis")
		return 2
	}
	cmd := exec.Command("cloudflared", "tunnel", "run", "--token", token)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return exitCode(cmd.Run())
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}

// Unused import guard (net) — conservé pour dial health futur.
var _ = net.Dial
