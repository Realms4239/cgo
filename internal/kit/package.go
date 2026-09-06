package kit

import (
	"archive/zip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Package — poste opérateur Windows en un zip : l'exe précompilé + exemple
// de config + mode d'emploi. Le binaire embarque déjà le frontend (go:embed)
// et ne demande rien d'autre à l'exécution ; seules les opérations kit
// parlent SSH (doctor le vérifie). Recompiler/redéployer depuis ce poste
// exige en plus Go 1.25+, bun, node et le dépôt — dit dans le README,
// pas découvert à l'échec.
func (r *Runner) Package(c *Config) int {
	exe := filepath.Join(r.Root, "cgo.exe")
	if _, err := exec.LookPath("go"); err == nil {
		r.out("[package] compilation fraîche de cgo.exe…")
		cmd := exec.Command("go", "build", "-o", exe, "./cmd/cgo")
		cmd.Dir = r.Root
		if out, err := cmd.CombinedOutput(); err != nil {
			r.errf("[package] go build ÉCHEC : %v\n%s", err, firstLine(string(out)))
			return 2
		}
	} else if _, err := os.Stat(exe); err != nil {
		r.errf("[package] ni Go (compilation impossible) ni cgo.exe présent — construisez d'abord (go build -o cgo.exe ./cmd/cgo)")
		return 2
	} else {
		r.out("[package] Go absent : on embarque le cgo.exe existant (vérifiez qu'il est à jour)")
	}
	outDir := filepath.Join(r.Root, "dist")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		r.errf("[package] mkdir dist : %v", err)
		return 2
	}
	zpath := filepath.Join(outDir, "cgo-windows-amd64.zip")
	zf, err := os.Create(zpath)
	if err != nil {
		r.errf("[package] création zip : %v", err)
		return 2
	}
	zw := zip.NewWriter(zf)
	add := func(name, src string) error {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	addStr := func(name, s string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(s))
		return err
	}
	fail := func(err error) int {
		_ = zw.Close()
		_ = zf.Close()
		r.errf("[package] %v", err)
		return 2
	}
	if err := add("cgo.exe", exe); err != nil {
		return fail(fmt.Errorf("exe : %w", err))
	}
	if err := add(filepath.Join("kit", "cgo-vm.yaml.example"), filepath.Join(r.Root, "kit", "cgo-vm.yaml.example")); err != nil {
		return fail(fmt.Errorf("config exemple : %w", err))
	}
	if err := addStr("LISEZ-MOI.txt", pcReadme()); err != nil {
		return fail(err)
	}
	if err := zw.Close(); err != nil {
		_ = zf.Close()
		r.errf("[package] finalisation : %v", err)
		return 2
	}
	_ = zf.Close()
	st, _ := os.Stat(zpath)
	r.out("[package] %s (%d octets) — copiez sur le poste, dézippez, lisez LISEZ-MOI.txt", zpath, st.Size())
	return 0
}

func pcReadme() string {
	return `METEOLINK — poste opérateur Windows
====================================
Contenu : cgo.exe (tout embarqué, frontend inclus), exemple de config.

1. Dézippez où vous voulez (ex. C:\cgo), ouvrez un terminal ici.
2. cgo.exe kit doctor — vérifie go/bun/node/ssh/scp/clé/VM.
   SSH manquant ? winget install --id Microsoft.OpenSSH.Client --source winget
3. Config : copiez kit\cgo-vm.yaml.example vers kit\cgo-vm.yaml et ajustez
   ip/clé — OU laissez faire : cgo.exe kit scan trouve la VM tout seul.
4. cgo.exe kit ensure — SSH actif vers la VM.
5. cgo.exe kit dns — mappe meteolink.dev (terminal ADMIN).
6. Ouvrez https://meteolink.dev:9090 — avertissement certificat :
   Avancé → Continuer (une fois), ou cgo.exe kit tls en ADMIN pour la
   confiance totale.

Rôles :
- Ce poste Windows = PILOTAGE (kit, audits terrain cgo audit, TUI, exports).
  Pas de façonnage tc : les campagnes shaping tournent sur la VM Ubuntu.
- La VM Ubuntu = MESURE (banc netem + dashboard). Le binaire linux déployé
  (GOOS=linux amd64, statique, testé Ubuntu 24.04, noyau 6.8 — tout Ubuntu
  20.04+ convient : netem/cake/netns requis) survit aux deploys : les runs
  gelés restent dans ~/cgo/data/runs.

Recompiler/redéployer DEPUIS ce poste exige en plus : Go 1.25+, bun, node
et le dépôt source (kit deploy recompile). Sans eux : exploitation seulement.
`
}
