package kit

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Vnet — médecin du réseau HÔTE vers la VM. Né d'une vraie panne : le
// VMnet8 de l'hôte avait perdu son 192.168.174.1 (APIPA 169.254.x) —
// invitée saine mais injoignable, Tools OK, silence total. Sans ce
// diagnostic on accuse la VM, les clés, le pare-feu… alors que c'est
// l'hôte qui n'est plus sur le subnet.
//   - cible loopback (NAT-forward 127.0.0.1) : test du listener local.
//   - cible privée : une interface UP du poste DOIT partager son /24.
//   - sinon : contexte (adaptateurs virtuels) + remède exact ; appliqué
//     direct si privilèges (admin Windows / root), sinon guidé.
func (r *Runner) Vnet(c *Config) int {
	host := strings.TrimSpace(c.SSHHost)
	if host == "" || host == "auto" {
		r.errf("[vnet] cible vide/auto — verrouillez la VM / configurez ssh_host d'abord")
		return 2
	}
	if strings.HasPrefix(host, "127.") || host == "localhost" {
		port := c.SSHPort
		if port == "" {
			port = "22"
		}
		if portOpen(host, port) {
			r.out("[vnet] forward local %s:%s en écoute — côté hôte OK", host, port)
			return 0
		}
		r.errf("[vnet] forward local %s:%s fermé — règle NAT absente ? `kit ensure` la repose", host, port)
		return 3
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() == nil || !ip.IsPrivate() {
		r.out("[vnet] cible %s non-privée — réseau du poste non concerné", host)
		return 0
	}
	sub := vnetSubnet(host)
	iface := hostIfaceOn(sub)
	if iface != "" {
		r.out("[vnet] hôte présent sur %s0/24 (%s) — côté hôte OK", sub, iface)
		return 0
	}
	r.errf("[vnet] KO : aucune interface UP du poste sur %s0/24 (cible %s)", sub, host)
	for _, ln := range virtIfaces() {
		r.out("[vnet]   %s", ln)
	}
	if runtime.GOOS == "windows" {
		return r.vnetFixWindows(sub)
	}
	return r.vnetFixLinux(sub)
}

func vnetSubnet(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return host
	}
	return strings.Join(parts[:3], ".") + "."
}

// hostIfaceOn — nom d'une interface UP non-loopback portant une IP du /24.
func hostIfaceOn(sub string) string {
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

// virtIfaces — adaptateurs virtuels du poste (contexte du diagnostic).
func virtIfaces() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []string
	for _, it := range ifaces {
		nm := strings.ToLower(it.Name)
		if !strings.Contains(nm, "vmnet") && !strings.Contains(nm, "virtualbox") &&
			!strings.Contains(nm, "vboxnet") && !strings.Contains(nm, "host-only") &&
			!strings.Contains(nm, "vmware") {
			continue
		}
		ips := []string{}
		if addrs, err := it.Addrs(); err == nil {
			for _, a := range addrs {
				ips = append(ips, a.String())
			}
		}
		state := "down"
		if it.Flags&net.FlagUp != 0 {
			state = "up"
		}
		out = append(out, fmt.Sprintf("%s : %s [%s]", it.Name, strings.Join(ips, ", "), state))
	}
	return out
}

func elevated() bool {
	if runtime.GOOS == "windows" {
		// `net session` ne réussit qu'élevé (contrat Windows stable).
		return exec.Command("net", "session").Run() == nil
	}
	return os.Geteuid() == 0
}

// vnetFixWindows — restaure 192.168.174.1/24 sur VMnet8 (la panne vue :
// APIPA après perte de config statique). Élevé → appliqué + revérifié ;
// sinon : script admin + one-liner.
func (r *Runner) vnetFixWindows(sub string) int {
	alias := ""
	ifaces, _ := net.Interfaces()
	for _, it := range ifaces {
		if strings.Contains(strings.ToLower(it.Name), "vmnet8") {
			alias = it.Name
		}
	}
	gw := sub + "1"
	if alias == "" {
		r.errf("[vnet] pas d'interface VMnet8 — VMware installé ? VirtualBox ? (vboxnet/host-only pour VBox)")
		return 3
	}
	script := findKitFile(r, "fix-vnet-admin.ps1")
	if elevated() {
		// Le script embarqué EST l'implémentation (pas de doublon inline
		// qui dériverait) ; repli inline si zip incomplet.
		if script != "" {
			r.out("[vnet] élevé : %s -GwIp %s…", script, gw)
			cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
				"-File", script, "-GwIp", gw)
			cmd.Stdout, cmd.Stderr = r.Stdout, r.Stderr
			if err := cmd.Run(); err != nil {
				r.errf("[vnet] script en échec : %v", err)
				return 3
			}
			return 0
		}
		r.out("[vnet] élevé : restauration %s/24 sur « %s »…", gw, alias)
		ps := fmt.Sprintf("New-NetIPAddress -InterfaceAlias '%s' -IPAddress %s -PrefixLength 24; Restart-Service -Name 'VMware NAT Service' -Force; Restart-Service -Name 'VMware DHCP Service' -Force", alias, gw)
		if out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput(); err != nil {
			r.errf("[vnet] restauration refusée : %s", strings.TrimSpace(string(out)))
			return 3
		}
		if hostIfaceOn(sub) != "" {
			r.out("[vnet] restauré — hôte de nouveau sur %s0/24, relancez `kit ensure`", sub)
			return 0
		}
		r.errf("[vnet] appliqué mais %s0/24 toujours absent — redémarrez le poste", sub)
		return 3
	}
	r.errf("[vnet] terminal NON élevé — la réparation demande admin (une fois) :")
	if script != "" {
		r.errf("[vnet]   clic-droit %s → « Exécuter en tant qu'administrateur »", script)
	} else {
		r.errf("[vnet]   powershell ADMIN : New-NetIPAddress -InterfaceAlias '%s' -IPAddress %s -PrefixLength 24", alias, gw)
	}
	return 3
}

// vnetFixLinux — même logique côté sudo (vmnet8/vboxnet0 tombés).
func (r *Runner) vnetFixLinux(sub string) int {
	dev := ""
	ifaces, _ := net.Interfaces()
	for _, it := range ifaces {
		nm := strings.ToLower(it.Name)
		if strings.HasPrefix(nm, "vmnet") || strings.HasPrefix(nm, "vboxnet") {
			dev = it.Name
		}
	}
	gw := sub + "1"
	if dev == "" {
		r.errf("[vnet] aucune interface vmnet*/vboxnet* — hyperviseur démarré ? (VM allumée ?)")
		return 3
	}
	cmd := fmt.Sprintf("sudo ip addr add %s/24 dev %s && sudo ip link set %s up", gw, dev, dev)
	if elevated() {
		r.out("[vnet] root : %s", cmd)
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			r.errf("[vnet] refusé : %v", err)
			return 3
		}
		if hostIfaceOn(sub) != "" {
			r.out("[vnet] restauré — relancez `kit ensure`")
			return 0
		}
		return 3
	}
	r.errf("[vnet] pas root — dans un terminal : %s", cmd)
	return 3
}

// findKitFile — script du kit : dépôt (kit/) puis à côté de l'exe (zips
// opérateur sans source). "" si introuvable (le message guidé prend le relais).
func findKitFile(r *Runner, name string) string {
	if p := filepath.Join(r.Root, "kit", name); fileExists(p) {
		return p
	}
	if ex, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(ex), "kit", name); fileExists(p) {
			return p
		}
		if p := filepath.Join(filepath.Dir(ex), name); fileExists(p) {
			return p
		}
	}
	return ""
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
