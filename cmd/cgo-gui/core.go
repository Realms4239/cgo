// Coeur applicatif PARTAGÉ (Windows + Linux) : aucune API Win32, aucun
// widget — que du pur Go (kit, vm, stdlib). Les deux vues (Win32 natif,
// Fyne) pilotent les mêmes actions : un bouton = un verbe testé sans écran.
package main

import (
	"context"
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
)
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
	{245, "Tunnel hôte", 0, 0, 95}, {246, "Invité", 0, 0, 70}, {247, "Réseau hôte", 0, 0, 95},
	{249, "Guide", 0, 0, 70}, {250, "▶ Suite", 0, 0, 85},
}

// note : le bouton Sauver (248, « direct:saveuser ») n'est dans AUCUN
// groupe : placé à côté du champ utilisateur (Win32 + Fyne), pas dans
// la colonne Contrôle.

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
	case 245:
		return "bg:hosttun", nil
	case 246:
		return "bg:guest", nil
	case 247:
		return "bg:vnet", nil
	case 248:
		return "direct:saveuser", nil
	case 249:
		return "bg:guide", nil
	case 250:
		return "direct:suite", nil
	}
	return "", nil
}

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
	cmd := kit.BgCmd("ssh-keygen", "-t", "ed25519", "-N", "", "-f", key, "-q")
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

func vmPowerGUI(c *kit.Config, r *kit.Runner, start bool) int {
	w := r.Stdout
	if c.VMXPath == "" {
		fmt.Fprintln(w, "aucune VM verrouillée — verrouillez-en une dans la liste")
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
		fmt.Fprintln(w, "hyperviseur absent — installez VMware Workstation ou VirtualBox")
		return 4
	}
	// Enregistre si orpheline (trouvée par scan disque, absente du manager) :
	// SANS ça, startvm échoue sur un nom inconnu. VMware travaille par
	// chemin ; seul VirtualBox exige l'enregistrement.
	if vb, ok := hyp.(interface {
		EnsureRegistered(string) (string, error)
	}); ok {
		name, err := vb.EnsureRegistered(c.VMXPath)
		if err != nil {
			fmt.Fprintln(w, "enregistrement impossible : "+err.Error())
			return 4
		}
		fmt.Fprintln(w, "VM enregistrée : "+name)
	}
	var err error
	if start {
		if vmRunningGUI(hyp, c.VMXPath) {
			fmt.Fprintln(w, "VM déjà allumée — attente de l'IP invitée (max 90 s)…")
		} else {
			if err = hyp.Start(c.VMXPath); err != nil { // headless/nogui — le banc n'a pas besoin d'écran
				fmt.Fprintln(w, "échec : "+err.Error())
				return 4
			}
			fmt.Fprintln(w, "démarrage headless demandé")
		}
	} else {
		err = hyp.Stop(c.VMXPath)
	}
	if err != nil {
		fmt.Fprintln(w, "échec : "+err.Error())
		return 4
	}
	if start {
		// Attend l'IP invitée (boot) : sans ça l'utilisateur clique
		// « Démarrer » puis « Déployer » 3 s plus tard et échoue.
		deadline := time.Now().Add(90 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(5 * time.Second)
			if ip := hyp.GuestIP(c.VMXPath); ip != "" {
				fmt.Fprintln(w, "VM en ligne, IP : "+ip)
				if host := c.SSHHost; host == "" || host == "auto" || host != ip {
					fmt.Fprintln(w, "→ si l'IP a changé, « Diagnostiquer » la détectera et mettra la config à jour")
				}
				return 0
			}
			if !vmRunningGUI(hyp, c.VMXPath) {
				fmt.Fprintln(w, "VM éteinte à nouveau — vérifiez le disque/BIOS dans la console")
				return 4
			}
		}
		fmt.Fprintln(w, "IP non vue en 90 s (tools pas encore prêts ?) — l'état SSH se rafraîchira seul")
	} else {
		fmt.Fprintln(w, "arrêt demandé")
	}
	return 0
}

func nicToggleGUI(c *kit.Config, r *kit.Runner, cfgPath string) int {
	// mode actuel via le même résolveur que partout
	hyp, vmx, err := pickVMGUI(c)
	if err != nil {
		fmt.Fprintln(r.Stdout, err.Error())
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
	// réseau HÔTE d'abord : si le poste n'est même pas sur le subnet de
	// la cible, ssh/TCP ne diront qu'« injoignable » — ici la vraie cause
	// (VMnet tombé, APIPA) avec son remède. Jamais bloquant : un averti.
	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil && ip.IsPrivate() &&
		!strings.HasPrefix(host, "127.") && host != "localhost" {
		parts := strings.Split(host, ".")
		if len(parts) == 4 {
			if sub := strings.Join(parts[:3], ".") + "."; hostIfaceOnGUI(sub) == "" {
				fmt.Fprintln(r.Stdout, ko+" réseau HÔTE : aucune interface sur "+sub+"0/24 — `kit vnet` (VMnet tombé ?)")
			}
		}
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
	cmd := kit.BgCmdCtx(ctx, "ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=4",
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

func vmRunningGUI(hyp vm.Hypervisor, vmx string) bool {
	for _, r := range hyp.Running() {
		if strings.EqualFold(filepath.Clean(r), filepath.Clean(vmx)) {
			return true
		}
	}
	return false
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	_ = cmd.Start()
}

// scanVMRows — inventaire VMs pur (partagé Win32/Fyne) : scan disque
// + hyperviseurs + état + mode réseau. Coûteux (subprocess) : fond.
func scanVMRows() []vmRow {

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
	return rows
}

// probeStatus — sonde SSH/dashboard pure (partagée Win32/Fyne) :
// pré-test TCP instantané puis SSH, jamais d'hyperviseur ici (loadCfg
// appelant). Coûteux (secondes) : fond uniquement.
func probeStatus(c *kit.Config) (ssh, dash, locked string) {
	ssh = "—"
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
	dash = "—"
	if ver, ok := dashHealth(c); ok {
		dash = "ok " + ver
	} else {
		dash = "injoignable"
	}
	locked = ""
	if c.VMName != "" {
		locked = c.VMName
	} else if c.VMXPath != "" {
		locked = filepath.Base(c.VMXPath)
	}
	return ssh, dash, locked
}

// vmRow — une VM vue par le scan (partagé Win32/Fyne).
type vmRow struct {
	path, name, hyp, mode string
	live                  bool
}

// cfgCache — LoadConfig résout l'IP invitée via vmrun (secondes) quand
// SSHHost=auto : SANS cache, CHAQUE clic pend le thread UI. Les deux vues
// partagent ce cache 45 s ; le fond le rafraîchit (note).
type cfgCache struct {
	mu   sync.Mutex
	host string
	at   time.Time
}

func (g *cfgCache) load(path string) *kit.Config {
	c, _ := kit.LoadConfig(path)
	g.mu.Lock()
	defer g.mu.Unlock()
	if c.SSHHost == "" || c.SSHHost == "auto" {
		if g.host != "" && time.Since(g.at) < 45*time.Second {
			c.SSHHost = g.host
		}
	} else {
		g.host, g.at = c.SSHHost, time.Now()
	}
	return c
}

func (g *cfgCache) note(ip string) {
	if ip == "" || ip == "auto" {
		return
	}
	g.mu.Lock()
	g.host, g.at = ip, time.Now()
	g.mu.Unlock()
}

// cfgSSHUser — lit le seul ssh_user du yaml (fichier seul, millisecondes,
// JAMAIS d'hyperviseur/réseau) : pré-remplit le champ utilisateur des GUI
// sans pendre le thread UI (LoadConfig résoudrait l'IP invitée via vmrun).
func cfgSSHUser(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	inSSH := false
	for _, raw := range strings.Split(string(b), "\n") {
		ln := strings.TrimSpace(strings.Split(raw, "#")[0])
		if ln == "" {
			continue
		}
		if !strings.HasPrefix(raw, " ") && !strings.HasPrefix(raw, "\t") {
			inSSH = ln == "ssh:"
			continue
		}
		if inSSH {
			if i := strings.Index(ln, ":"); i > 0 && strings.TrimSpace(ln[:i]) == "user" {
				return strings.Trim(strings.TrimSpace(ln[i+1:]), `"'`)
			}
		}
	}
	return ""
}

// hostTunCmd — invoque le script embarqué host-tunnel.ps1 (7 étapes :
// forwards, hosts, clé, confiance, vérification) avec la VM verrouillée
// et l'utilisateur configuré : zéro Read-Host, sortie streamable.
// Rend (ps1, argv, "") ou ("", nil, message) si bloqué en amont.
func hostTunCmd(c *kit.Config, exeDir, root string) (string, []string, string) {
	if c.SSHUser == "" {
		return "", nil, "utilisateur vide — renseignez ssh_user (TUI/config) puis relancez"
	}
	name := c.VMName
	if name == "" {
		name = c.VMXPath
	}
	if name == "" {
		return "", nil, "aucune VM verrouillée — choisissez d'abord dans la liste"
	}
	ps1 := ""
	for _, p := range []string{
		filepath.Join(exeDir, "host-tunnel.ps1"),
		filepath.Join(root, "kit", "host-tunnel.ps1"),
	} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			ps1 = p
			break
		}
	}
	if ps1 == "" {
		return "", nil, "host-tunnel.ps1 introuvable (zip incomplet ?)"
	}
	return ps1, []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", ps1,
		"-VmName", name, "-User", c.SSHUser}, ""
}

// hostIfaceOnGUI — une interface UP du poste porte-t-elle ce /24 ?
// (miroir kit.hostIfaceOn, non exporté là-bas).
func hostIfaceOnGUI(sub string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, it := range ifaces {
		if it.Flags&net.FlagUp == 0 || it.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := it.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.To4() == nil {
				continue
			}
			if strings.HasPrefix(ip.String(), sub) {
				return it.Name
			}
		}
	}
	return ""
}
