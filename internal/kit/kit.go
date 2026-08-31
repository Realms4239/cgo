// Package kit — moteur de déploiement en Go : les 11 actions de l'ancien
// engine.sh, portables, testables, avec les mêmes codes de sortie
// (2 usage/déps/build, 3 scan/pick, 4 hyperviseur absent, 5 timeout SSH,
// 6 cross-compile/bootstrap, 7 scp, 8 install/logs) et les mêmes variables
// d'environnement CGO_* — les scripts existants ne cassent pas.
package kit

import (
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
	GoMinVer    string
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
		SSHHost:    env("CGO_SSH_HOST", "192.168.174.128"),
		SSHPort:    env("CGO_SSH_PORT", "22"),
		SSHKey:     env("CGO_SSH_KEY", "~/.ssh/id_ed25519"),
		ProjectDir: env("CGO_PROJECT_DIR", "/home/altfloat/cgo"),
		DashPort:   env("CGO_DASHBOARD_PORT", "9090"),
		GoMinVer:   "1.25",
		Hypervisor: "auto",
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return c, nil // config absente → défauts + env (portable, pas fatal)
	}
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
		case "go_min_version":
			c.GoMinVer = v
		}
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

// SaveVMX — persiste vm_name/vmx_path/hypervisor dans le yaml (auto-rempli par scan).
func SaveVMX(path, vmx, hypervisor string) error {
	name := strings.TrimSuffix(filepath.Base(vmx), filepath.Ext(vmx))
	set := func(key, val string) {
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
		_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	}
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

func (r *Runner) out(format string, a ...any) { fmt.Fprintf(r.Stdout, format+"\n", a...) }
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

// HTTPGetJSON — vérification health du dashboard.
func (c *Config) Health() bool {
	cl := &http.Client{Timeout: 4 * time.Second}
	resp, err := cl.Get(c.healthURL())
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
		} else {
			r.out("  %-10s MANQUANT", t)
		}
	}
	for _, h := range vm.Detect() {
		r.out("  %-10s %s", h.Name(), h.Exe())
	}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("tc"); err != nil {
			r.out("  tc         MANQUANT (mode observation — normal sous Windows)")
		}
	}
	r.out("  config     %s → %s:%s %s", "<cgo-vm.yaml>", c.SSHHost, c.DashPort, c.ProjectDir)
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

// Ensure — SSH up, sinon boot + attente.
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
	if err := hyp.Start(vmx); err != nil {
		r.errf("[ensure] démarrage VM échoué : %v", err)
		return 4
	}
	r.out("[ensure] VM démarrée — attente SSH (max 300 s)")
	for i := 0; i < 60; i++ {
		if c.SSHUp() {
			r.out("[ensure] SSH actif après ~%d s", i*5)
			return 0
		}
		time.Sleep(5 * time.Second)
	}
	r.errf("[ensure] timeout SSH")
	return 5
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
	if _, err := c.SSH("mkdir -p " + c.ProjectDir + "/kit"); err != nil {
		return 7
	}
	if err := c.SCP(bin, c.ProjectDir+"/cgo-linux.new"); err != nil {
		return 7
	}
	if err := c.SCP(filepath.Join(r.Root, "kit", "vm-install.sh"), c.ProjectDir+"/kit/vm-install.sh"); err != nil {
		return 7
	}
	r.out("[deploy] installation VM...")
	if _, err := c.SSH("cd " + c.ProjectDir + " && bash kit/vm-install.sh"); err != nil {
		return 8
	}
	if c.Health() {
		r.out("[deploy] fait → http://%s:%s", c.SSHHost, c.DashPort)
		return 0
	}
	r.errf("[deploy] health KO sur :%s", c.DashPort)
	return 8
}

// Bootstrap — paquets + veth dans la VM (idempotent).
func (r *Runner) Bootstrap(c *Config) int {
	r.out("[bootstrap] paquets VM...")
	if _, err := c.SSH("sudo apt update && sudo apt install -y iproute2 curl bc && sudo modprobe tcp_bbr || true"); err != nil {
		r.errf("[bootstrap] apt ÉCHEC")
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
	}
	return 0
}

// Logs — tail du journal serveur VM.
func (r *Runner) Logs(c *Config, n int) int {
	out, err := c.SSH("tail -n " + strconv.Itoa(n) + " /tmp/cgo.log")
	if err != nil {
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
