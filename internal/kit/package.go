package kit

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		cmd := bgCmd("go", "build", "-o", exe, "./cmd/cgo")
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
		// Sourceless explicit : le dossier cmd/cgo est la preuve de source.
		// Sans lui NI Go, le zip embarquerait un cgo.exe d'âge inconnu —
		// dire lequel, au lieu d'un repli muet (poste opérateur du zip).
		if _, err := os.Stat(filepath.Join(r.Root, "cmd", "cgo")); err != nil {
			ver := selfVersion()
			r.out("[package] poste SANS source ni Go — j'embarque le cgo.exe présent (version %s ; c'est la version testée du zip, rien à faire)", ver)
		} else {
			r.out("[package] Go absent avec source présente : installez Go 1.25+ pour compiler, sinon j'embarque le cgo.exe existant (vérifiez qu'il est à jour)")
		}
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
		cmd := bgCmd("go", "build", "-o", tmpLin, "./cmd/cgo")
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
		cmd := bgCmd("go", "build", "-ldflags", "-H windowsgui", "-o", tmpGUI, "./cmd/cgo-gui")
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
	// fix-vnet-admin.ps1 idem : répare le VMnet8 tombé en APIPA (hôte
	// injoignable alors que la VM est saine) — clic-droit admin, une fois.
	if err := add("fix-vnet-admin.ps1", filepath.Join(r.Root, "kit", "fix-vnet-admin.ps1")); err != nil {
		return fail(fmt.Errorf("fix-vnet : %w", err))
	}
	if err := addStr("LISEZ-MOI.txt", pcReadme()); err != nil {
		return fail(err)
	}
	// DEMARRER.bat : double-clic qui lance le centre de contrôle depuis
	// n'importe où (le .bat cale le dossier, start détache sans console).
	if err := addStr("DEMARRER.bat", "@echo off\r\ncd /d \"%~dp0\"\r\nstart \"\" \"%~dp0cgo-gui.exe\"\r\n"); err != nil {
		return fail(fmt.Errorf("demarrer : %w", err))
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

// Readme — le LISEZ-MOI du poste courant (Windows : pcReadme, Linux :
// linuxReadme) : UNE source pour le zip, `kit readme` et le bouton Guide
// des GUI. Jamais deux textes qui dérivent.
func Readme() string {
	if runtime.GOOS == "windows" {
		return pcReadme()
	}
	return linuxReadme()
}

// selfVersion — version embarquée à la compilation (cmd/cgo la lie via
// -ldflags ; repli local si absent). Pour le message sourceless.
func selfVersion() string {
	if v := os.Getenv("CGO_VERSION"); v != "" {
		return v
	}
	return "inconnue"
}

func pcReadme() string {
	return `METEOLINK — centre de contrôle (poste Windows)
===============================================
Contenu du zip : cgo.exe, cgo-gui.exe (+ manifeste, logo),
cgo-linux (compagnon à pousser vers la VM), host-tunnel.ps1,
fix-vnet-admin.ps1, kit/cgo-vm.yaml.example, kit/vm-install.sh,
kit/guest-setup.sh, ce LISEZ-MOI.

PRINCIPE
--------
Le pilotage suit toujours le même pipeline (mêmes mots que
« cgo kit next » et que le bandeau « Prochaine » de la GUI) :

  1. VM verrouillée   2. réseau hôte   3. VM allumée   4. cible SSH
  5. port SSH         6. clé acceptée  7. binaire       8. dashboard
  9. nom meteolink.dev                    10. confiance HTTPS

Chaque palier dit QUOI faire ensuite. Ne sautez jamais d'étape :
un dashboard muet vient toujours d'un palier amont (clé, réseau…).

VOIE GUI (recommandée, zéro commande)
-------------------------------------
Double-cliquez DEMARRER.bat (ou cgo-gui.exe directement).

  - Liste des VM (VirtualBox/VMware) : Rescanner, double-clic = verrouiller.
  - Champ « Utilisateur Ubuntu » + Sauver (ex. fanasina) : SANS lui,
    diagnostic + clé + deploy avortent — c'est normal, renseignez-le.
  - Bandeau « Prochaine : … » : l'étape calculée en continu.
  - Bouton « Suite » : exécute l'étape du bandeau (forwards, boot,
    clé, deploy, dashboard, DNS, confiance — chacun son bouton aussi).
  - Bouton « Guide » : réaffiche ce texte dans le journal.
  - Journal : chaque action raconte tout ; garde anti-double-clic
    (« patience — … tourne déjà ») + boutons d'action grisés pendant
    qu'une action tourne (la liste des VM reste sélectionnable).
  - Démarrages longs (boot, deploy) : progression journalisée en continu.
  - Preuves : cgo-gui-<date>.log à côté de l'exe (envoyable au support).

VOIE TERMINAL
-------------
  cgo.exe kit tui      centre de contrôle interactif (flèches + entrée)
  cgo.exe kit next     la checklist + la prochaine étape, en 10 secondes
  cgo.exe kit readme   réaffiche ce texte
  cgo.exe kit doctor   dépendances locales + config

VOIE ZÉRO-KIT (sans cgo.exe)
----------------------------
  host-tunnel.ps1  tunnel complet en 7 étapes (forwards, hosts, clé,
                   confiance, vérification). -WhatIf pour répéter sans
                   rien toucher. PowerShell 5.1+ : le fichier DOIT garder
                   son BOM UTF-8 (sinon erreur d'accolade au parsing).

COMMANDES (voie manuelle, équivalent exact du guidé)
----------------------------------------------------
  cgo.exe kit scan         trouve les .vmx/.vbox (hyperviseur par extension)
  cgo.exe kit vnet         médecin du réseau HÔTE (VMnet tombé ? APIPA ?)
  cgo.exe kit ensure       boot headless + forwards NAT + attente SSH (bavard)
  cgo.exe kit keysetup     pose LA CLÉ (mot de passe tapé UNE fois, jamais stocké)
  cgo.exe kit deploy       pousse cgo-linux (précompilé, sans toolchain)
  cgo.exe kit svc start    (re)lance le dashboard   |  svc stop/restart/status
  cgo.exe kit dns          mappe meteolink.dev — TERMINAL ADMIN
  cgo.exe kit tls          confiance HTTPS — TERMINAL ADMIN
  cgo.exe kit guest        prépare l'invité (guest-setup.sh via SSH)
  cgo.exe kit logs         40 dernières lignes du dashboard
  cgo.exe kit backup       archive les runs   |  snapshot/revert : garde-fous VM
  cgo.exe kit health       santé JSON   |  shipcheck : les 6 portes avant release

SI ÇA COINCE (par symptôme, pas au hasard)
------------------------------------------
  - « Permission denied (publickey) » : la clé de CE poste n'est pas
    dans l'invité (clé régénérée ? autre PC ?) → kit keysetup, une fois.
  - Forward « ouvert » mais handshake vide : le backend est muet
    (svc éteint, binaire jamais déployé) → kit deploy PUIS svc start.
  - « aucune interface sur 192.168.x.0/24 » : VMnet tombé (APIPA).
    Clic-droit fix-vnet-admin.ps1 → Exécuter en tant qu'administrateur ;
    100 % PowerShell : voir « kit vnet ». Sans clic-droit admin : impossible
    (frontière Windows, pas un bug).
  - .ps1 « accolade manquante » : BOM UTF-8 perdu (ré-extrayez le zip).
  - SmartScreen au lancement : binaire non signé → « Informations
    complémentaires » → Exécuter quand même.
  - IP changée après reboot (DHCP) : kit ensure la redécouvre et met la
    config à jour tout seul.
  - Port 22 fermé : DANS la console Ubuntu :
    sudo apt install -y openssh-server && sudo systemctl enable --now ssh
  - Port 2222 occupé : un autre forward/service l'utilise → nat_host_port
    dans kit/cgo-vm.yaml.

RÔLES
-----
  - Ce poste Windows = PILOTAGE (kit, audits terrain cgo audit, TUI, exports).
    Pas de façonnage tc : les campagnes shaping tournent sur la VM Ubuntu.
  - La VM Ubuntu = MESURE (banc netem + dashboard). Le binaire linux déployé
    (GOOS=linux amd64, statique, testé Ubuntu 24.04, noyau 6.8 — tout Ubuntu
    20.04+ convient : netem/cake/netns requis) survit aux deploys : les runs
    gelés restent dans ~/cgo/data/runs.

RECOMPILER DEPUIS CE POSTE exige en plus : Go 1.25+, bun, node et le dépôt
source. Mais cgo.exe kit deploy pousse le compagnon cgo-linux (dans ce zip)
vers la VM SANS toolchain — le plein pilotage (scan, clé, boot, deploy,
dashboard) ne demande que ce zip + le mot de passe de la VM.
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
	cmd := bgCmd("go", "build", "-o", bin, "./cmd/cgo")
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
	// GUI Linux (Fyne) : compilée SUR Ubuntu (CGO/GL natifs — pas de
	// cross depuis Windows). Le build VM dépose dist/cgo-gui-linux ;
	// absent → tarball sans GUI (kit tui couvre), jamais d'échec.
	if gl, err := os.Stat(filepath.Join(r.Root, "dist", "cgo-gui-linux")); err == nil && !gl.IsDir() {
		if err := add("cgo-gui", filepath.Join(r.Root, "dist", "cgo-gui-linux"), 0755); err != nil {
			return fail(fmt.Errorf("gui linux : %w", err))
		}
		r.out("[package] GUI Linux embarquée (double-clic sur bureau Ubuntu)")
	} else {
		r.out("[package] sans cgo-gui (buildez sur Ubuntu : go build -o dist/cgo-gui-linux ./cmd/cgo-gui)")
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
	return `METEOLINK — centre de contrôle (poste Ubuntu)
=============================================
Contenu : cgo + cgo-gui (binaires linux/amd64), exemple de config.

Prérequis : openssh-client (sudo apt install -y openssh-client),
hyperviseur + VM Ubuntu du banc, python3 (sondes locales, souvent présent).
AUCUN autre paquet : ni Go, ni node, ni webkit — la GUI Fyne ne demande
que les libs déjà présentes sur tout bureau Ubuntu (libgl1, X11/Wayland).
Sans écran (serveur, SSH) : ./cgo kit tui — même pilotage, en texte.

PRINCIPE
--------
Même pipeline que partout (mêmes mots que « ./cgo kit next ») :

  1. VM verrouillée   2. réseau hôte   3. VM allumée   4. cible SSH
  5. port SSH         6. clé acceptée  7. binaire       8. dashboard
  9. nom meteolink.dev                    10. confiance HTTPS

VOIE GUI
--------
Double-cliquez cgo-gui (bureau Ubuntu) : mêmes 27 boutons, même journal
et même bandeau « Prochaine » que la version Windows — « Suite » exécute
l'étape, « Guide » réaffiche ce texte, preuves dans cgo-gui-<date>.log.

VOIE TERMINAL
-------------
  ./cgo kit tui      interactif (flèches + entrée, rien à taper)
  ./cgo kit next     la checklist + la prochaine étape, en 10 secondes
  ./cgo kit readme   réaffiche ce texte
  ./cgo kit doctor   dépendances locales + config
  Pas de Go/bun/node ? Normal : l'exploitation n'en a pas besoin.

COMMANDES (voie manuelle, équivalent exact du guidé)
----------------------------------------------------
  ./cgo kit scan         trouve les .vmx/.vbox (hyperviseur par extension)
  ./cgo kit vnet         médecin du réseau HÔTE (vmnet/vboxnet tombé ?)
  ./cgo kit ensure       boot headless + forwards NAT + attente SSH (bavard)
  ./cgo kit keysetup     pose LA CLÉ (mot de passe tapé UNE fois, jamais stocké).
    Si SSH refuse tout (port 22 fermé) : DANS la console Ubuntu de la VM :
    sudo apt install -y openssh-server && sudo systemctl enable --now ssh
  ./cgo kit deploy       pousse CE binaire testé (précompilé, sans toolchain)
  ./cgo kit svc start    (re)lance le dashboard  |  stop/restart/status/logs
  sudo ./cgo kit dns     mappe meteolink.dev vers la VM
  sudo ./cgo kit tls     confiance HTTPS (magasin système)
  ./cgo kit guest        prépare l'invité (guest-setup.sh via SSH)
  ./cgo kit backup       archive les runs  |  snapshot/revert : garde-fous VM
  ./cgo kit health       santé JSON  |  shipcheck : les 6 portes avant release

SI ÇA COINCE (par symptôme)
---------------------------
  - « Permission denied (publickey) » : clé de CE poste absente de
    l'invité → kit keysetup, une fois.
  - Forward « ouvert » mais handshake vide : backend muet (svc éteint,
    binaire jamais déployé) → kit deploy PUIS svc start.
  - « aucune interface sur 192.168.x.0/24 » : vmnet/vboxnet tombé →
    sudo ip addr add 192.168.x.1/24 dev <iface> (voir « kit vnet »).
  - IP changée après reboot (DHCP) : kit ensure la redécouvre seul.
  - Port 2222 occupé : nat_host_port dans kit/cgo-vm.yaml.

RÔLES
-----
  - Ce poste Ubuntu = PILOTAGE (kit, audits terrain ./cgo audit, TUI, exports).
  - La VM du banc = MESURE (banc netem + dashboard). Les runs gelés survivent
    aux deploys dans ~/cgo/data/runs.

AVEC le dépôt source + Go 1.25+, bun, node : le deploy recompile depuis
les sources ; la GUI Linux se rebuilt par : go build -o dist/cgo-gui-linux
./cmd/cgo-gui (gcc + libgl1-mesa-dev + xorg-dev + libwayland-dev requis).
`
}
