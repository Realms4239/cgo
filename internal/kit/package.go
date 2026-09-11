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
	// testbed.sh : le banc de mesure côté invité (veth + netns + testbedsrv,
	// sudoers auto-posé) — sans lui les campagnes tournent à vide (zéro
	// ligne gelée, « Démarrer → idle → rien » sur VM fraîche). Poussé sur
	// la VM par deploy, exécuté par `kit testbed up` (+x conservé).
	if err := addExec("kit/testbed.sh", filepath.Join(r.Root, "kit", "testbed.sh")); err != nil {
		return fail(fmt.Errorf("testbed : %w", err))
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
Zip : cgo.exe, cgo-gui.exe (logo + version intégrés), cgo-linux
(compagnon VM), kit/*.sh + exemple yaml, DEMARRER.bat, ce LISEZ-MOI.

PREMIER LANCEMENT : un seul prompt UAC — l'app s'élève et tout hérite
(plus aucun terminal admin). Clic-droit « Exécuter en tant
qu'administrateur » n'est plus nécessaire.

MÉTHODE : UNE PRESSION SUR « SUITE » SUFFIT
-------------------------------------------
Double-cliquez DEMARRER.bat. Cliquez « ▶ Suite » UNE fois : la chaîne
avance SEULE, étape après étape (scan + verrouillage auto, boot +
forwards, deploy, dashboard, DNS, confiance, banc), jusqu'à
« tout est vert » ou jusqu'au premier point qui exige une main humaine :
mot de passe (console noire « Poser la clé »), choix d'une VM (plusieurs
trouvées → double-clic), ou pilote à installer. Dans ces cas le journal
dit exactement quoi faire ; corrigez, re-cliquez Suite, ça repart.
La chaîne ne boucle jamais : échec, répétition ou fin l'arrêtent avec
un message. Tout autre bouton cliqué à la main interrompt la chaîne
(vous avez pris la main, c'est normal).

3 CHOSES À SAVOIR
-----------------
- « Utilisateur Ubuntu » + Sauver (ex. fanasina) : SANS lui, tout avorte.
  (Suite s'arrête au palier clé et le dit ; renseignez-le, re-Suite.)
- « Poser la clé » : mot de passe tapé UNE fois dans la console noire,
  jamais stocké. Si la clé passe déjà, ça ne fait rien.
- « Snapshot (arrêt VM) » : snapshot À FROID (arrêt propre d'abord).
- « Banc de mesure » : vérifie le banc invité (campagnes impossibles
  sans lui — « Démarrer » gèlerait zéro ligne).
- Tout survit aux redémarrages : config (écriture insécable), journal
  quotidien, runs gelés côté VM, reprise de campagne par run-id.

COMMANDES (équivalent manuel exact)
-----------------------------------
  cgo.exe kit scan|vnet|ensure|keysetup|deploy|svc|dns|tls
  cgo.exe kit testbed up|check|down   banc de mesure invité
  cgo.exe kit guest|logs|backup|snapshot|revert|health|shipcheck|tui|next

SI ÇA COINCE (par symptôme)
---------------------------
- « Permission denied (publickey) » → kit keysetup, une fois.
- Forward ouvert mais vide → kit deploy PUIS svc start.
- « aucune interface 192.168.x.0/24 » → fix-vnet-admin.ps1 en admin.
- IP changée (DHCP) → kit ensure redécouvre seul.
- Port 2222 occupé → nat_host_port dans kit/cgo-vm.yaml.
- Port 22 fermé → DANS la VM : sudo apt install -y openssh-server.
- ZÉRO ligne gelée → kit testbed up (banc absent).
- 10.0.2.x injoignable : NORMAL sous NAT — le kit passe par 127.0.0.1:2222.
- SmartScreen → « Informations complémentaires » → Exécuter quand même.
- Session VBox verrouillée (E_FAIL) → PowerShell ADMIN :
  Get-Process VBoxHeadless | Stop-Process -Force, puis relancez.

RÔLES : ce poste = PILOTAGE (pas de tc ici). La VM Ubuntu = MESURE
(banc netem + dashboard, runs dans ~/cgo/data/runs, survivent aux deploys).
Recompiler exige Go 1.25+ ; l'exploitation n'a besoin que de ce zip.
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
	// testbed.sh : le banc de mesure (veth + netns + testbedsrv) — voir zip.
	if err := add("kit/testbed.sh", filepath.Join(r.Root, "kit", "testbed.sh"), 0755); err != nil {
		return fail(fmt.Errorf("testbed : %w", err))
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
Tar : cgo + cgo-gui (linux/amd64), exemple de config, ce LISEZ-MOI.

Prérequis : openssh-client, hyperviseur + VM du banc. Rien d'autre
(pas de Go/node ; GUI = X11/Wayland déjà présents). Sans écran :
./cgo kit tui — même pilotage, en texte.

MÉTHODE : UNE PRESSION SUR « SUITE » SUFFIT
-------------------------------------------
Lancez cgo-gui (ou ./cgo kit next en terminal). Cliquez « ▶ Suite » UNE
fois : la chaîne avance SEULE jusqu'à « tout est vert » ou jusqu'au
premier point manuel (mot de passe, choix VM, pilote). Le journal
horodaté dit tout ; preuves dans cgo-gui-<date>.log.

3 CHOSES À SAVOIR
-----------------
- Utilisateur Ubuntu + Sauver : SANS lui, tout avorte.
- « Poser la clé » : mot de passe UNE fois, jamais stocké.
- « Banc de mesure » : sans lui, les campagnes gèlent zéro ligne.
- Tout survit aux redémarrages : config insécable, runs dans
  ~/cgo/data/runs, reprise de campagne par run-id.

COMMANDES (équivalent manuel exact)
-----------------------------------
  ./cgo kit scan|vnet|ensure|keysetup|deploy|svc|testbed up|check|down
  ./cgo kit guest|logs|backup|snapshot|revert|health|shipcheck|tui|next
  sudo ./cgo kit dns|tls   (hosts + magasin système)
  Port 22 fermé côté VM : sudo apt install -y openssh-server.

SI ÇA COINCE (par symptôme)
---------------------------
- « Permission denied (publickey) » → kit keysetup, une fois.
- ZÉRO ligne gelée → kit testbed up (banc absent).
- Forward ouvert mais vide → kit deploy PUIS svc start.
- « aucune interface 192.168.x.0/24 » → vmnet tombé : voir « kit vnet ».
- IP changée (DHCP) → kit ensure redécouvre seul.
- Port 2222 occupé → nat_host_port dans kit/cgo-vm.yaml.
- Binaire ≤1.2.2 + santé HTTPS KO → kit deploy.

RÔLES : ce poste = PILOTAGE. La VM du banc = MESURE (netem + dashboard).
Rebuild (optionnel) : Go 1.25+ ; GUI Linux : gcc + libgl1-mesa-dev +
xorg-dev + libwayland-dev, puis go build -o dist/cgo-gui-linux ./cmd/cgo-gui.
AVEC sources, le deploy recompile ; SANS, il pousse le binaire testé.
`
}
