package kit

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Realms4239/cgo/internal/vm"
)

// NextStep — la PROCHAINE action du pipeline, calculée (pas devinée) :
// scan → vnet → boot → cible → port → clé → binaire → dashboard → dns →
// confiance → banc de mesure. Chaque palier est vérifié vite et en lecture
// seule ; le premier non-vert donne l'étape (CLI : commande exacte, GUI :
// bouton pour « Suite »). Même ordre, mêmes mots que le LISEZ-MOI.
type NextStep struct {
	ID     string // vm, vnet, boot, cible, port, cle, binaire, svc, dns, tls, banc
	Label  string // texte banner (« Prochaine : … »)
	State  string // "ok", "ko", "attente" (amont non vert)
	Detail string // preuve courte
	Remedy string // commande CLI exacte
	Verb   string // verbe buttonAction GUI ("" = guidance seule)
	Args   []string
}

// GatherNext — photographie du pipeline. Budgets courts partout (total
// < 30 s pire cas) : c'est appelé à chaque fin d'action + ticker GUI.
func GatherNext(c *Config) []NextStep {
	steps := []NextStep{}
	blocked := false // un KO/amont gèle la pertinence de la suite
	push := func(s NextStep) {
		if blocked && s.State == "ok" {
			s.State = "attente"
			s.Detail = "en attente (" + steps[len(steps)-1].ID + " d'abord)"
		}
		if s.State != "ok" {
			blocked = true
		}
		steps = append(steps, s)
	}

	// 1. VM verrouillée + pilote présent
	// 1. VM verrouillée + pilote présent (VMPath : .vmx OU .vbox —
	// exiger vmx_path pour une VBox était le bug du palier 1).
	var hyp vm.Hypervisor
	locked, hypName, vmx := "", "", ""
	vpath := c.VMPath()
	switch {
	case vpath == "":
		push(NextStep{ID: "vm", Label: "Verrouiller une VM", State: "ko",
			Detail: "aucune VM verrouillée",
			Remedy: "cgo kit scan"})
	default:
		st, err := os.Stat(vpath)
		if err != nil || st.IsDir() {
			push(NextStep{ID: "vm", Label: "VM introuvable", State: "ko",
				Detail: vpath + " n'existe plus",
				Remedy: "cgo kit scan"})
			break
		}
		h, err := selectDriver(vm.Detect(), c.Hypervisor, vpath)
		if err != nil {
			want := vm.HypForPath(vpath)
			push(NextStep{ID: "vm", Label: "Pilote manquant", State: "ko",
				Detail: fmt.Sprintf("%s : pilote %s absent — installez %s", filepath.Base(vpath), want,
					map[string]string{"virtualbox": "VirtualBox", "vmware": "VMware Workstation"}[want]),
				Remedy: "installez " + map[string]string{"virtualbox": "VirtualBox", "vmware": "VMware Workstation"}[want]})
			break
		}
		locked, hypName, hyp, vmx = filepath.Base(vpath), h.Name(), h, vpath
		push(NextStep{ID: "vm", Label: "VM verrouillée", State: "ok",
			Detail: fmt.Sprintf("%s (%s)", locked, hypName)})
	}

	// 2. réseau hôte (cible concrète seulement)
	host := strings.TrimSpace(c.SSHHost)
	sub := ""
	if parts := strings.Split(host, "."); len(parts) == 4 {
		if ip := net.ParseIP(host); ip != nil && ip.To4() != nil && ip.IsPrivate() &&
			!strings.HasPrefix(host, "127.") && host != "localhost" {
			sub = strings.Join(parts[:3], ".") + "."
		} else if strings.HasPrefix(host, "127.") || host == "localhost" {
			sub = "local"
		}
	}
	switch {
	case host == "" || host == "auto":
		push(NextStep{ID: "vnet", Label: "Réseau hôte", State: "attente", Detail: "cible auto"})
	case sub == "local":
		push(NextStep{ID: "vnet", Label: "Réseau hôte", State: "ok", Detail: "forward local"})
	case sub == "":
		push(NextStep{ID: "vnet", Label: "Réseau hôte", State: "ok", Detail: "cible publique"})
	case hostIfaceOn(sub) != "":
		push(NextStep{ID: "vnet", Label: "Réseau hôte", State: "ok", Detail: "hôte sur " + sub + "0/24"})
	default:
		push(NextStep{ID: "vnet", Label: "Réseau hôte tombé", State: "ko",
			Detail: "aucune interface sur " + sub + "0/24",
			Remedy: "cgo kit vnet", Verb: "bg:vnet"})
	}

	// 3. boot
	if hyp == nil {
		push(NextStep{ID: "boot", Label: "VM allumée", State: "attente", Detail: "pas de VM"})
	} else if vmIsRunning(hyp, vmx) {
		push(NextStep{ID: "boot", Label: "VM allumée", State: "ok", Detail: "en cours d'exécution"})
	} else {
		push(NextStep{ID: "boot", Label: "Démarrer la VM", State: "ko",
			Detail: "éteinte", Remedy: "cgo kit ensure", Verb: "bg:vmon"})
	}

	// 4. cible résolue
	if host == "" || host == "auto" {
		push(NextStep{ID: "cible", Label: "Cible SSH", State: "attente",
			Detail: "auto — résolue par la prochaine étape", Remedy: "cgo kit ensure", Verb: "bg:ensure"})
	} else {
		push(NextStep{ID: "cible", Label: "Cible SSH", State: "ok", Detail: host + ":" + sshPort(c)})
	}

	// 5. port TCP
	if !stepOK(steps, "cible") {
		push(NextStep{ID: "port", Label: "Port SSH", State: "attente", Detail: "cible d'abord"})
	} else if portOpen(host, sshPort(c)) {
		push(NextStep{ID: "port", Label: "Port SSH", State: "ok", Detail: host + ":" + sshPort(c) + " ouvert"})
	} else {
		push(NextStep{ID: "port", Label: "Port SSH fermé", State: "ko",
			Detail: host + ":" + sshPort(c) + " muet",
			Remedy: "cgo kit ensure", Verb: "bg:ensure"})
	}

	// 6. clé acceptée (= auth SSH)
	if !stepOK(steps, "port") {
		push(NextStep{ID: "cle", Label: "Clé SSH", State: "attente", Detail: "port d'abord"})
	} else if _, err := c.SSH("true"); err == nil {
		push(NextStep{ID: "cle", Label: "Clé SSH", State: "ok", Detail: "acceptée par " + c.SSHUser + "@" + host})
	} else {
		push(NextStep{ID: "cle", Label: "Poser la clé SSH", State: "ko",
			Detail: "refusée par l'invité (mot de passe une fois)",
			Remedy: "cgo kit keysetup", Verb: "console:keysetup", Args: []string{"keysetup"}})
	}

	// 7. binaire déployé
	proj := c.ProjectDir
	if proj == "" && c.SSHUser != "" {
		proj = "/home/" + c.SSHUser + "/cgo"
	}
	if !stepOK(steps, "cle") {
		push(NextStep{ID: "binaire", Label: "Binaire déployé", State: "attente", Detail: "clé d'abord"})
	} else if out, err := c.SSH("test -x " + shq(proj+"/cgo-linux") + " && echo OK"); err == nil && strings.Contains(out, "OK") {
		push(NextStep{ID: "binaire", Label: "Binaire déployé", State: "ok", Detail: proj + "/cgo-linux"})
	} else {
		push(NextStep{ID: "binaire", Label: "Déployer le binaire", State: "ko",
			Detail: "cgo-linux absent de l'invité",
			Remedy: "cgo kit deploy", Verb: "bg:deploy"})
	}

	// 8. dashboard sain
	if !stepOK(steps, "binaire") {
		push(NextStep{ID: "svc", Label: "Dashboard", State: "attente", Detail: "binaire d'abord"})
	} else if ver, ok := dashProbe(dashProbeURL(c)); ok {
		push(NextStep{ID: "svc", Label: "Dashboard", State: "ok", Detail: "sain (version " + ver + ")"})
	} else if httpOnlyUp(c) {
		push(NextStep{ID: "svc", Label: "Dashboard pré-TLS", State: "ko",
			Detail: "HTTP seul (binaire ≤1.2.2) — svc start ne soignera jamais",
			Remedy: "cgo kit deploy", Verb: "bg:deploy"})
	} else {
		push(NextStep{ID: "svc", Label: "Démarrer le dashboard", State: "ko",
			Detail: "aucune réponse saine",
			Remedy: "cgo kit svc start", Verb: "bg:svc", Args: []string{"start"}})
	}

	// 9. dns (nom → cible)
	dashName := c.DashHost
	if dashName == "" {
		dashName = "meteolink.dev"
	}
	if !stepOK(steps, "svc") {
		push(NextStep{ID: "dns", Label: "Nom meteolink.dev", State: "attente", Detail: "dashboard d'abord"})
	} else if host == "" || host == "auto" {
		push(NextStep{ID: "dns", Label: "Nom meteolink.dev", State: "attente", Detail: "cible d'abord"})
	} else if dnsMapsTo(dashName, host) {
		push(NextStep{ID: "dns", Label: "Nom meteolink.dev", State: "ok", Detail: dashName + " → " + host})
	} else {
		push(NextStep{ID: "dns", Label: "Mapper meteolink.dev (admin)", State: "ko",
			Detail: dashName + " ne pointe pas vers " + host,
			Remedy: "cgo kit dns (terminal admin)", Verb: "console:dns", Args: []string{"dns"}})
	}

	// 10. confiance OS (HTTPS système, pas -k)
	if !stepOK(steps, "dns") {
		push(NextStep{ID: "tls", Label: "Confiance HTTPS", State: "attente", Detail: "nom d'abord"})
	} else if tlsTrusted(dashURLFor(c)) {
		push(NextStep{ID: "tls", Label: "Confiance HTTPS", State: "ok", Detail: "magasin OS OK"})
	} else {
		push(NextStep{ID: "tls", Label: "Confiance HTTPS (admin)", State: "ko",
			Detail: "certificat non reconnu par le système",
			Remedy: "cgo kit tls (terminal admin)", Verb: "console:tls", Args: []string{"tls"}})
	}

	// 11. banc de mesure (campagnes) — DERNIER : ne bloque rien après lui,
	// mais « tout est vert » l'exige. Sans lui (VM fraîche), Démarrer gèle
	// un run à zéro ligne. Sonde guest directe (small + sink + sudo) —
	// sudo testé sur la SORTIE : sudo 1.9.15p5 sort 0 même en refusant -n.
	if !stepOK(steps, "cle") {
		push(NextStep{ID: "banc", Label: "Banc de mesure", State: "attente", Detail: "clé d'abord"})
	} else if out, err := c.SSH("curl -fsS -m3 http://10.200.0.1:8081/small -o /dev/null && timeout 1 bash -c '</dev/tcp/10.200.0.1/5201' && ! sudo -n ip netns list 2>&1 | grep -qi password && echo PLANE_OK"); err == nil && strings.Contains(out, "PLANE_OK") {
		push(NextStep{ID: "banc", Label: "Banc de mesure", State: "ok", Detail: "small + bulk + sudo — campagnes possibles"})
	} else {
		push(NextStep{ID: "banc", Label: "Banc de mesure", State: "ko",
			Detail: "absent — Démarrer gèlerait zéro ligne",
			Remedy: "cgo kit testbed up", Verb: "bg:testbed"})
	}
	return steps
}

// FirstOpen — premier palier non-vert (ko ou attente) : LA prochaine
// étape. Pur et testé ; la collecte (lente) reste dans GatherNext.
func FirstOpen(steps []NextStep) *NextStep {
	for i := range steps {
		if steps[i].State != "ok" {
			return &steps[i]
		}
	}
	return nil
}

func stepOK(steps []NextStep, id string) bool {
	for _, s := range steps {
		if s.ID == id {
			return s.State == "ok"
		}
	}
	return false
}

// Next — `kit next` : checklist complète + prochaine étape, lisible en
// 5 secondes. Code 0 si tout vert, 3 sinon.
func (r *Runner) Next(c *Config) int {
	steps := GatherNext(c)
	for _, s := range steps {
		mark := "[✓]"
		if s.State == "ko" {
			mark = "[✗]"
		} else if s.State == "attente" {
			mark = "[…]"
		}
		line := mark + " " + s.Label
		if s.Detail != "" {
			line += " — " + s.Detail
		}
		if s.State == "ko" && s.Remedy != "" {
			line += "  →  " + s.Remedy
		}
		r.out("%s", line)
	}
	if nx := FirstOpen(steps); nx != nil {
		if nx.Remedy != "" {
			r.out("→ prochaine étape : %s", nx.Remedy)
		} else {
			r.out("→ prochaine étape : %s (%s)", nx.Label, nx.Detail)
		}
		return 3
	}
	r.out("→ tout est vert — dashboard : %s", dashURLFor(c))
	return 0
}

// --- sondes pures-Go (mêmes budgets que partout : courts) ---

func vmIsRunning(hyp vm.Hypervisor, vmx string) bool {
	for _, r := range hyp.Running() {
		if vm.SameVM(r, vmx) {
			return true
		}
	}
	return false
}

func sshPort(c *Config) string {
	if c.SSHPort != "" {
		return c.SSHPort
	}
	return "22"
}

func dashURLFor(c *Config) string {
	host := c.DashHost
	if host == "" {
		host = "meteolink.dev"
	}
	port := c.DashPort
	if port == "" {
		port = "9090"
	}
	return "https://" + host + ":" + port
}

// dashProbeURL — sonde le dashboard sur la CIBLE VIVANTE (IP SSH), pas le
// nom (qui peut être périmé : DHCP change l'IP, hosts pas à jour —
// sonder le nom confond « svc éteint » et « dns périmé », deux paliers
// distincts). Repli nom si pas de cible concrète.
func dashProbeURL(c *Config) string {
	if h := strings.TrimSpace(c.SSHHost); h != "" && h != "auto" {
		port := c.DashPort
		if port == "" {
			port = "9090"
		}
		return "https://" + h + ":" + port
	}
	return dashURLFor(c)
}

func dashProbe(url string) (string, bool) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	cl := &http.Client{Timeout: 5 * time.Second, Transport: tr}
	resp, err := cl.Get(url + "/api/health")
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

func dnsMapsTo(name, ip string) bool {
	addrs, err := lookupHostFast(name)
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if a == ip {
			return true
		}
	}
	return false
}

func tlsTrusted(url string) bool {
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Get(url + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var doc struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil || !doc.OK {
		return false
	}
	return true
}
