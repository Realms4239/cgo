package kit

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Realms4239/cgo/internal/vm"
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
	// Espace invité du NAT VirtualBox : 10.0.2.0/24 vit DANS le NAT, aucun
	// adaptateur hôte ne doit jamais le porter (incident : 10.0.2.1/24 écrit
	// sur VMnet8). Seule voie : le forward 127.0.0.1 — jamais d'écriture ici.
	if isVBoxGuestSpace(host) {
		r.errf("[vnet] cible %s = espace invité NAT VirtualBox — injoignable en direct par construction", host)
		r.errf("[vnet] utilisez le forward local (ssh_host 127.0.0.1 + nat_host_port) — `kit ensure` le repose")
		return 3
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
	hyp := resolveHyp(c)
	if runtime.GOOS == "windows" {
		return r.vnetFixWindows(sub, hyp)
	}
	return r.vnetFixLinux(sub, hyp)
}

// resolveHyp — hyperviseur de la VM verrouillée ("vmware"/"virtualbox"),
// "" si inconnu (pas de VM verrouillée ou pilote absent). En mode inconnu
// on diagnostique seulement : on n'écrit jamais à l'aveugle.
func resolveHyp(c *Config) string {
	if c.VMPath() == "" {
		return ""
	}
	if h, err := selectDriver(vm.Detect(), c.Hypervisor, c.VMPath()); err == nil {
		return h.Name()
	}
	return ""
}

// wantAdapter — l'écriture sur un adaptateur hôte n'est autorisée que si
// l'adaptateur appartient à la famille de l'hyperviseur verrouillé.
// Retourne l'alias choisi ou "" (diagnostic seul).
func wantAdapter(hyp, sub string) string {
	ifaces, _ := net.Interfaces()
	best := ""
	for _, it := range ifaces {
		nm := strings.ToLower(it.Name)
		var family bool
		switch hyp {
		case "vmware":
			family = strings.Contains(nm, "vmnet")
		case "virtualbox":
			family = strings.Contains(nm, "vbox") || strings.Contains(nm, "virtualbox") || strings.Contains(nm, "host-only")
		default:
			continue
		}
		if !family {
			continue
		}
		if hyp == "vmware" && strings.Contains(nm, "vmnet8") {
			return it.Name // NAT VMware : VMnet8 d'abord
		}
		if best == "" {
			best = it.Name
		}
	}
	// VirtualBox hors host-only (ponté/autre) : le LAN décide, l'hôte ne
	// peut rien réparer en écrivant — diagnostic seul.
	if hyp == "virtualbox" && sub != "192.168.56." {
		return ""
	}
	return best
}

// adapterHasOtherSubnet — l'adaptateur porte déjà un /24 privé sain
// différent de la cible : l'écraser casserait un réseau qui marche.
func adapterHasOtherSubnet(alias, sub string) bool {
	ifaces, _ := net.Interfaces()
	for _, it := range ifaces {
		if it.Name != alias {
			continue
		}
		addrs, err := it.Addrs()
		if err != nil {
			return false
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.To4() == nil || !ip.IsPrivate() {
				continue
			}
			s := ip.String()
			if strings.HasPrefix(s, "169.254.") || strings.HasPrefix(s, "fe80") {
				continue
			}
			if !strings.HasPrefix(s, sub) {
				return true
			}
		}
	}
	return false
}

func vnetSubnet(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return host
	}
	return strings.Join(parts[:3], ".") + "."
}

// isVBoxGuestSpace — 10.0.2.0/24 est l'espace invité interne du NAT
// VirtualBox : il ne doit apparaître sur AUCUN adaptateur hôte.
func isVBoxGuestSpace(host string) bool {
	return strings.HasPrefix(strings.TrimSpace(host), "10.0.2.")
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
		return bgCmd("net", "session").Run() == nil
	}
	return os.Geteuid() == 0
}

// vnetFixWindows — restaure sub+1/24 sur l'adaptateur de la famille de
// l'hyperviseur (la panne vue : APIPA après perte de config statique).
// Garde-fous (incident 10.0.2.1/24 sur VMnet8) : jamais d'écriture si
// l'hyperviseur est inconnu, si la cible est l'espace NAT interne VBox,
// ou si l'adaptateur porte déjà un autre /24 sain. Élevé → appliqué +
// revérifié ; sinon : script admin + one-liner.
func (r *Runner) vnetFixWindows(sub, hyp string) int {
	if isVBoxGuestSpace(sub + "0") {
		r.errf("[vnet] refusé : 10.0.2.0/24 est interne au NAT VirtualBox — jamais sur un adaptateur hôte")
		return 3
	}
	alias := wantAdapter(hyp, sub)
	gw := sub + "1"
	if alias == "" {
		if hyp == "" {
			r.errf("[vnet] hyperviseur inconnu (pas de VM verrouillée ?) — réparation manuelle uniquement")
		} else if hyp == "virtualbox" {
			r.errf("[vnet] cible pontée/autre (%s0/24) : côté LAN, l'hôte ne peut rien écrire — vérifiez le réseau", sub)
		} else {
			r.errf("[vnet] pas d'interface vmnet* — VMware installé ?")
		}
		return 3
	}
	if adapterHasOtherSubnet(alias, sub) {
		r.errf("[vnet] refusé : « %s » porte déjà un autre /24 sain — l'écraser casserait un réseau qui marche", alias)
		return 3
	}
	script := findKitFile(r, "fix-vnet-admin.ps1")
	if elevated() {
		// Le script embarqué EST l'implémentation (pas de doublon inline
		// qui dériverait) ; repli inline si zip incomplet.
		if script != "" {
			r.out("[vnet] élevé : %s -GwIp %s…", script, gw)
			cmd := bgCmd("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
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
		if out, err := bgCmd("powershell", "-NoProfile", "-Command", ps).CombinedOutput(); err != nil {
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

// vnetFixLinux — même logique côté sudo (vmnet8/vboxnet0 tombés),
// mêmes garde-fous : pas d'écriture sans hyperviseur connu, jamais sur
// 10.0.2.0/24, jamais par-dessus un autre /24 sain.
func (r *Runner) vnetFixLinux(sub, hyp string) int {
	if isVBoxGuestSpace(sub + "0") {
		r.errf("[vnet] refusé : 10.0.2.0/24 est interne au NAT VirtualBox — jamais sur un adaptateur hôte")
		return 3
	}
	dev := wantAdapter(hyp, sub)
	gw := sub + "1"
	if dev == "" {
		if hyp == "" {
			r.errf("[vnet] hyperviseur inconnu (pas de VM verrouillée ?) — réparation manuelle uniquement")
		} else if hyp == "virtualbox" {
			r.errf("[vnet] cible pontée/autre (%s0/24) : côté LAN, l'hôte ne peut rien écrire — vérifiez le réseau", sub)
		} else {
			r.errf("[vnet] aucune interface vmnet* — hyperviseur démarré ? (VM allumée ?)")
		}
		return 3
	}
	if adapterHasOtherSubnet(dev, sub) {
		r.errf("[vnet] refusé : « %s » porte déjà un autre /24 sain — l'écraser casserait un réseau qui marche", dev)
		return 3
	}
	cmd := fmt.Sprintf("sudo ip addr add %s/24 dev %s && sudo ip link set %s up", gw, dev, dev)
	if elevated() {
		r.out("[vnet] root : %s", cmd)
		if err := bgCmd("sh", "-c", cmd).Run(); err != nil {
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
