package kit

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Package — poste opérateur en une archive : binaire précompilé + exemple
// de config + mode d'emploi. --os windows (défaut, .zip) | linux (.tar.gz).
// Le binaire embarque déjà le frontend (go:embed) et ne demande rien
// d'autre à l'exécution ; seules les opérations kit parlent SSH (doctor le
// vérifie). Recompiler/redéployer depuis ce poste exige en plus Go 1.25+,
// bun, node et le dépôt — dit dans le README, pas découvert à l'échec.
func (r *Runner) Package(c *Config, rest []string) int {
	targetOS := "windows"
	for i := 0; i < len(rest); i++ {
		if rest[i] == "--os" && i+1 < len(rest) {
			targetOS = strings.ToLower(rest[i+1])
		}
	}
	if targetOS == "linux" {
		return r.packageLinux()
	}
	if targetOS != "windows" {
		r.errf("[package] --os %s inconnu (windows|linux)", targetOS)
		return 2
	}
	exe := filepath.Join(r.Root, "cgo.exe")
	if _, err := exec.LookPath("go"); err == nil {
		r.out("[package] compilation fraîche de cgo.exe (windows/amd64)…")
		cmd := exec.Command("go", "build", "-o", exe, "./cmd/cgo")
		cmd.Dir = r.Root
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
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
	// addExec — comme add mais avec le bit +x conservé (le compagnon linux
	// part vers la VM via scp ; l'installateur rechmod de toute façon, mais
	// un zip honnête porte les perms d'origine).
	addExec := func(name, src string) error {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(0755)
		w, err := zw.CreateHeader(h)
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
	// compagnon linux : le poste Windows pilote une VM Ubuntu — `kit deploy`
	// depuis ce zip pousse CE binaire (précompilé, pas de toolchain requise
	// côté opérateur). Sans Go ici : zip sans compagnon (observation seule).
	if _, err := exec.LookPath("go"); err == nil {
		r.out("[package] compagnon linux pour deploy VM…")
		tmpLin := filepath.Join(outDir, "cgo-linux-pkg.exe-tmp")
		cmd := exec.Command("go", "build", "-o", tmpLin, "./cmd/cgo")
		cmd.Dir = r.Root
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.Remove(tmpLin)
			return fail(fmt.Errorf("compagnon linux : %v\n%s", err, firstLine(string(out))))
		}
		if err := addExec("cgo-linux", tmpLin); err != nil {
			_ = os.Remove(tmpLin)
			return fail(fmt.Errorf("compagnon linux : %w", err))
		}
		_ = os.Remove(tmpLin)
	} else {
		r.out("[package] Go absent : zip SANS compagnon linux (kit deploy VM impossible depuis ce zip — observation + audits seuls)")
	}
	// mini-GUI native (Win32 pur, ~200 Ko) + manifeste (styles modernes,
	// DPI) + logo : double-clic = centre de contrôle, zéro terminal.
	if _, err := exec.LookPath("go"); err == nil {
		r.out("[package] mini-GUI Windows…")
		tmpGUI := filepath.Join(outDir, "cgo-gui-pkg.exe-tmp")
		cmd := exec.Command("go", "build", "-ldflags", "-H windowsgui", "-o", tmpGUI, "./cmd/cgo-gui")
		cmd.Dir = r.Root
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.Remove(tmpGUI)
			r.out("[package] GUI ignorée (build) : " + firstLine(string(out)))
		} else {
			if err := add("cgo-gui.exe", tmpGUI); err != nil {
				_ = os.Remove(tmpGUI)
				return fail(fmt.Errorf("gui : %w", err))
			}
			_ = os.Remove(tmpGUI)
			for _, asset := range [][2]string{
				{"cgo-gui.exe.manifest", filepath.Join(r.Root, "cmd", "cgo-gui", "cgo-gui.exe.manifest")},
				{"logo.ico", filepath.Join(r.Root, "assets", "logo.ico")},
			} {
				if err := add(asset[0], asset[1]); err != nil {
					return fail(fmt.Errorf("asset GUI %s : %w", asset[0], err))
				}
			}
		}
	}
	// noms d'entrées en '/' obligatoires (spec zip/tar) — filepath.Join
	// produit '\' sur Windows : l'archive livrerait un fichier littéral
	// "kit\cgo-vm.yaml.example" au lieu d'un dossier kit/ (vu en prod).
	if err := add("kit/cgo-vm.yaml.example", filepath.Join(r.Root, "kit", "cgo-vm.yaml.example")); err != nil {
		return fail(fmt.Errorf("config exemple : %w", err))
	}
	// l'installateur voyage avec le poste : le deploy précompilé (sans
	// source) en a besoin sur la VM — sans lui, push impossible.
	if err := add("kit/vm-install.sh", filepath.Join(r.Root, "kit", "vm-install.sh")); err != nil {
		return fail(fmt.Errorf("installateur : %w", err))
	}
	// guest-setup.sh : la voie manuelle côté invité (console Ubuntu, sans SSH
	// préalable) — le LISEZ-MOI et le guide y renvoient pour le cas sans kit.
	if err := add("kit/guest-setup.sh", filepath.Join(r.Root, "kit", "guest-setup.sh")); err != nil {
		return fail(fmt.Errorf("guest-setup : %w", err))
	}
	// host-tunnel.ps1 à la RACINE du zip (visible immédiatement) : tunnel
	// Host->VM sans kit — forwards, hosts, clé, confiance, vérification.
	if err := add("host-tunnel.ps1", filepath.Join(r.Root, "kit", "host-tunnel.ps1")); err != nil {
		return fail(fmt.Errorf("host-tunnel : %w", err))
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

1. Dézippez où vous voulez (ex. C:\cgo).
   VOIE SIMPLE : double-cliquez cgo-gui.exe — le centre de contrôle
   graphique fait tout (boutons + journal, aucune commande).
   VOIE EXPRESS (un script, zéro kit) : clic-droit host-tunnel.ps1 →
   « Exécuter avec PowerShell » — tunnel complet (forwards, hosts, clé,
   confiance, vérification) avec -WhatIf pour répéter sans rien toucher.
   Windows SmartScreen peut prévenir au premier lancement (binaire non
   signé) : « Informations complémentaires » → Exécuter quand même.
2. Voie terminal : ouvrez un terminal ici, cgo.exe kit tui — CENTRE DE
   CONTRÔLE interactif (flèches + entrée, rien à taper) : dépendances
   (OpenSSH installé auto) → scan VMware/VirtualBox → choix +
   verrouillage VM → clé SSH → deploy → dashboard. Voie normale.
   Voie manuelle (équivalent exact) :
   cgo.exe kit doctor — vérifie go/bun/node/ssh/scp/clé/VM.
   SSH manquant ? Le TUI l'installe, ou : winget install --id Microsoft.OpenSSH.Client --source winget
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

Recompiler DEPUIS ce poste exige en plus : Go 1.25+, bun, node
et le dépôt source. Mais cgo.exe kit deploy pousse le compagnon
cgo-linux (dans ce zip) vers la VM SANS toolchain — le plein pilotage
(scan, clé, boot, deploy, dashboard) ne demande que ce zip + le mot de
passe de la VM.
`
}

// packageLinux — poste opérateur Ubuntu en .tar.gz : même contenu, binaire
// linux/amd64 statique (CGO_ENABLED=0 — tourne sur tout Ubuntu 20.04+,
// prouvé 24.04 noyau 6.8). Testé : VirtualBox/VMware + openssh-client.
func (r *Runner) packageLinux() int {
	bin := filepath.Join(r.Root, "dist", "cgo-linux-pkg")
	if _, err := exec.LookPath("go"); err != nil {
		r.errf("[package] Go requis pour cross-compiler linux — installez Go 1.25+")
		return 2
	}
	r.out("[package] cross-compile linux/amd64…")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/cgo")
	cmd.Dir = r.Root
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		r.errf("[package] go build linux ÉCHEC : %v\n%s", err, firstLine(string(out)))
		return 2
	}
	outDir := filepath.Join(r.Root, "dist")
	tpath := filepath.Join(outDir, "cgo-linux-amd64.tar.gz")
	tf, err := os.Create(tpath)
	if err != nil {
		r.errf("[package] création tar.gz : %v", err)
		return 2
	}
	gz := gzip.NewWriter(tf)
	tw := tar.NewWriter(gz)
	add := func(name, src string, mode int64) error {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(data))}); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	}
	addStr := func(name, s string) error {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(s))}); err != nil {
			return err
		}
		_, err = tw.Write([]byte(s))
		return err
	}
	fail := func(err error) int {
		_ = tw.Close()
		_ = gz.Close()
		_ = tf.Close()
		r.errf("[package] %v", err)
		return 2
	}
	if err := add("cgo", bin, 0755); err != nil {
		return fail(fmt.Errorf("binaire : %w", err))
	}
	if err := add("kit/cgo-vm.yaml.example", filepath.Join(r.Root, "kit", "cgo-vm.yaml.example"), 0644); err != nil {
		return fail(fmt.Errorf("config exemple : %w", err))
	}
	if err := add("kit/vm-install.sh", filepath.Join(r.Root, "kit", "vm-install.sh"), 0644); err != nil {
		return fail(fmt.Errorf("installateur : %w", err))
	}
	if err := add("kit/guest-setup.sh", filepath.Join(r.Root, "kit", "guest-setup.sh"), 0644); err != nil {
		return fail(fmt.Errorf("guest-setup : %w", err))
	}
	if err := addStr("LISEZ-MOI.txt", linuxReadme()); err != nil {
		return fail(err)
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		_ = tf.Close()
		r.errf("[package] finalisation tar : %v", err)
		return 2
	}
	_ = gz.Close()
	_ = tf.Close()
	_ = os.Remove(bin)
	st, _ := os.Stat(tpath)
	r.out("[package] %s (%d octets) — copiez sur le poste Ubuntu, tar xzf, lisez LISEZ-MOI.txt", tpath, st.Size())
	return 0
}

func linuxReadme() string {
	return `METEOLINK — poste opérateur Ubuntu
====================================
Contenu : cgo (binaire linux/amd64 statique, frontend inclus), exemple de config.

Prérequis : openssh-client (sudo apt install -y openssh-client),
hyperviseur + VM Ubuntu du banc, python3 (sondes locales, souvent présent).

1. tar xzf cgo-linux-amd64.tar.gz -C ~/cgo-op && cd ~/cgo-op && chmod +x cgo
2. ./cgo kit tui — CENTRE DE CONTRÔLE interactif (flèches + entrée, rien à
   taper) : dépendances → scan VM → verrouillage → clé SSH → deploy →
   dashboard. C'est la voie normale.
   Voie manuelle (équivalent exact) :
   ./cgo kit doctor — vérifie go/bun/node/ssh/scp/clé/VM.
   Pas de Go/bun/node ? Normal : l'exploitation n'en a pas besoin.
   Seul openssh-client est requis (installé auto par le kit si absent).
3. Config : cp kit/cgo-vm.yaml.example kit/cgo-vm.yaml et ajustez ip/clé —
   OU laissez faire : ./cgo kit scan trouve la VM tout seul (tout le PC).
4. ./cgo kit keysetup — vous demande l'utilisateur, l'hôte, le port, puis
   LE MOT DE PASSE dans l'invite ssh elle-même (jamais stocké) ; pose la clé.
   Si SSH refuse tout (port 22 fermé) : DANS la console Ubuntu de la VM,
   tapez : sudo apt install -y openssh-server && sudo systemctl enable --now ssh
   — puis relancez keysetup. (Astuce : ./cgo kit tui fait tout cela en guidé.)
5. ./cgo kit ensure — SSH actif vers la VM (boot + IP auto si besoin).
6. sudo ./cgo kit dns — mappe meteolink.dev vers la VM.
7. Ouvrez https://meteolink.dev:9090 — avertissement certificat :
   Avancé → Continuer (une fois), ou sudo ./cgo kit tls pour la
   confiance totale (installé dans le magasin système).
8. ./cgo kit deploy — pousse CE binaire testé sur la VM et sert le
   dashboard (mode précompilé : aucune recompilation, aucune chaîne Go
   requise). Avec le dépôt source + Go 1.25+, bun, node : le deploy
   recompile depuis les sources à la place.

Rôles :
- Ce poste Ubuntu = PILOTAGE (kit, audits terrain ./cgo audit, TUI, exports).
  Pas de façonnage tc ici non plus sans droits root + modules : les campagnes
  shaping tournent sur la VM du banc.
- La VM du banc = MESURE (banc netem + dashboard). Les runs gelés survivent
  aux deploys dans ~/cgo/data/runs.
`
}
