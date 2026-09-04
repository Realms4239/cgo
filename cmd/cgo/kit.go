package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/internal/kit"
)

// runKit — moteur de déploiement Go : 21 actions, mêmes codes de sortie
// que l'ancien engine.sh (2 usage/build, 3 scan/pick, 4 hyperviseur,
// 5 timeout SSH, 6 cross/bootstrap, 7 scp, 8 install/logs).
func runKit(args []string) int {
	fs := flag.NewFlagSet("kit", flag.ExitOnError)
	cfgPath := fs.String("config", filepath.Join("kit", "cgo-vm.yaml"), "chemin du yaml machine-local")
	deep := fs.Bool("deep", true, "scan complet des disques (défaut : tout le PC)")
	shallow := fs.Bool("shallow", false, "scan restreint aux conventions (D:/VMs, C:/VMs, racines)")
	public := fs.Bool("public", false, "tunnel cloudflared après deploy")
	yes := fs.Bool("yes", false, "non-interactif (défauts acceptés)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: cgo kit <action> [--config FILE] [--shallow] [--yes]

actions :
  doctor    dépendances locales + config (tout vert avant d'agir)
  scan      trouve les .vmx/.vbox sur TOUT le PC (défaut), sauvegarde l'unique
  keysetup  pose la clé SSH sur la cible via mot de passe (prompts, zéro GUI)
  ensure    SSH up, sinon boot VM + attente (300 s max) + IP auto-découverte
  align     NIC + CPU/mémoire mini du banc (à froid, snapshot auto avant)
  build     porte stricte : go vet + tsc + vite + bundle <600 KB + vitest
  deploy    build + ensure + cross-compile linux + scp + install + health
  svc       pilote le dashboard distant : start|stop|restart|status
  bootstrap paquets VM + veth (idempotent)
  status    SSH + process + health dashboard
  logs      tail du journal serveur VM
  tunnel    cloudflared (CLOUDFLARE_TUNNEL_TOKEN requis)
   snapshot  point de restauration VM (garde-fou avant align/deploy)
   revert    revenir au dernier snapshot (ou --name NOM)
   snapshots liste des instantanés (le nom sert à revert)
   ssh       shell interactif direct dans la VM (Ctrl-D pour sortir)
   ps        processus dashboard en direct (rafraîchi 2 s, q pour sortir)
   backup    rapatrie les runs gelés de la VM vers ./backup (tar.gz horodaté)
   verify    empreintes SHA-256 des archives, recalculées sur la VM
   health    état du dashboard distant (JSON + verdict)
exit codes : 2 usage/build, 3 scan ambigu, 4 hyperviseur absent, 5 timeout SSH,
             6 cross-compile/bootstrap, 7 scp, 8 install/logs`)
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		return 2
	}
	action := fs.Arg(0)
	rest := fs.Args()[1:]
	_ = yes
	// deep par défaut (tout le PC) ; --shallow restreint aux conventions.
	effDeep := *deep && !*shallow

	c, _ := kit.LoadConfig(*cfgPath)
	r := kit.NewRunner()

	switch action {
	case "doctor":
		return r.Doctor(c)
	case "scan":
		return r.Scan(c, *cfgPath, effDeep)
	case "keysetup":
		return r.KeySetup(c, *cfgPath, rest)
	case "ensure":
		return r.Ensure(c, *cfgPath, effDeep)
	case "align":
		return r.Align(c, effDeep)
	case "build":
		return r.Build()
	case "deploy":
		code := r.Deploy(c, *cfgPath, effDeep)
		if code == 0 && *public {
			return r.Tunnel()
		}
		return code
	case "svc":
		sub := "status"
		if len(rest) > 0 {
			sub = rest[0]
		}
		return r.Svc(c, sub)
	case "bootstrap":
		return r.Bootstrap(c)
	case "status":
		return r.Status(c)
	case "logs":
		return r.Logs(c, 40)
	case "tunnel":
		return r.Tunnel()
	case "snapshot":
		name := "cgo-auto"
		if len(rest) > 0 {
			name = rest[0]
		}
		return r.Snapshot(c, name, effDeep)
	case "revert":
		name := ""
		if len(rest) > 0 {
			name = rest[0]
		}
		return r.Revert(c, name, effDeep)
	case "ssh":
		return r.SSHInteractive(c)
	case "snapshots":
		return r.Snapshots(c, effDeep)
	case "verify":
		return r.Verify(c)
	case "health":
		return r.Health(c)
	case "ps":
		return r.Ps(c)
	case "backup":
		dest := "backup"
		if len(rest) > 0 {
			dest = rest[0]
		}
		return r.Backup(c, dest)
	case "schedule":
		at := "02:30"
		profiles := "P2"
		qdiscs := ""
		ccs := ""
		direction := "both"
		reps := 0
		for i := 0; i+1 < len(rest); i += 2 {
			switch rest[i] {
			case "--at":
				at = rest[i+1]
			case "--profiles":
				profiles = rest[i+1]
			case "--qdiscs":
				qdiscs = rest[i+1]
			case "--cc":
				ccs = rest[i+1]
			case "--direction":
				direction = rest[i+1]
			case "--reps":
				fmt.Sscanf(rest[i+1], "%d", &reps)
			}
		}
		return r.Schedule(c, at, profiles, qdiscs, ccs, direction, reps)
	default:
		fmt.Fprintf(os.Stderr, "action inconnue : %s\n", action)
		fs.Usage()
		return 2
	}
}
