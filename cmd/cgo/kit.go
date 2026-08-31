package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/internal/kit"
)

// runKit — moteur de déploiement Go, miroir des 11 actions de l'ancien
// engine.sh avec les mêmes codes de sortie (2 usage/build, 3 scan/pick,
// 4 hyperviseur, 5 timeout SSH, 6 cross/bootstrap, 7 scp, 8 install/logs).
func runKit(args []string) int {
	fs := flag.NewFlagSet("kit", flag.ExitOnError)
	cfgPath := fs.String("config", filepath.Join("kit", "cgo-vm.yaml"), "chemin du yaml machine-local")
	deep := fs.Bool("deep", false, "scan complet des disques (lent)")
	public := fs.Bool("public", false, "tunnel cloudflared après deploy")
	yes := fs.Bool("yes", false, "non-interactif (défauts acceptés)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: cgo kit <action> [--config FILE] [--deep] [--yes]

actions :
  doctor    dépendances locales + config (tout vert avant d'agir)
  scan      trouve les .vmx/.vbox (D:/VMs, C:/VMs, racines), sauvegarde l'unique
  ensure    SSH up, sinon boot VM + attente (300 s max)
  align     NIC vmxnet3 + CPU/mémoire mini du banc (à froid)
  build     porte stricte : go vet + tsc + vite + bundle <600 KB + vitest
  deploy    build + ensure + cross-compile linux + scp + install + health
  bootstrap paquets VM + veth (idempotent)
  status    SSH + process + health dashboard
  logs      tail du journal serveur VM
  tunnel    cloudflared (CLOUDFLARE_TUNNEL_TOKEN requis)
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
	_ = yes

	c, _ := kit.LoadConfig(*cfgPath)
	r := kit.NewRunner()

	switch action {
	case "doctor":
		return r.Doctor(c)
	case "scan":
		return r.Scan(c, *cfgPath, *deep)
	case "ensure":
		return r.Ensure(c, *cfgPath, *deep)
	case "align":
		return r.Align(c, *deep)
	case "build":
		return r.Build()
	case "deploy":
		code := r.Deploy(c, *cfgPath, *deep)
		if code == 0 && *public {
			return r.Tunnel()
		}
		return code
	case "bootstrap":
		return r.Bootstrap(c)
	case "status":
		return r.Status(c)
	case "logs":
		return r.Logs(c, 40)
	case "tunnel":
		return r.Tunnel()
	default:
		fmt.Fprintf(os.Stderr, "action inconnue : %s\n", action)
		fs.Usage()
		return 2
	}
}
