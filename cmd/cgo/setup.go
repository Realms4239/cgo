package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/internal/vm"
	"github.com/Realms4239/cgo/pkg/doctor"
)

// runSetup — wizard de première installation : de git clone au dashboard en
// une seule commande guidée. Idempotent (chaque étape verte est sautée),
// trois modes : interactif (défaut), --yes (défauts, CI), --no-vm (sans
// hyperviseur — observation/audit seulement).
func runSetup(args []string) int {
	fs := flagSetup(args)
	yes := fs.lookupBool("yes")
	noVM := fs.lookupBool("no-vm")

	stdin := bufio.NewReader(os.Stdin)
	ask := func(q, def string) string {
		if yes {
			return def
		}
		fmt.Printf("%s [%s] : ", q, def)
		ln, _ := stdin.ReadString('\n')
		ln = strings.TrimSpace(ln)
		if ln == "" {
			return def
		}
		return ln
	}
	confirm := func(q string) bool {
		return strings.EqualFold(ask(q+" (o/N)", "n"), "o")
	}
	step := func(n int, title string) { fmt.Printf("\n== %d/7 %s ==\n", n, title) }
	root := mustRepoRoot()

	// 1 — bienvenue + détection
	step(1, "Détection")
	fmt.Printf("hôte : %s/%s — go %s\n", runtime.GOOS, runtime.GOARCH, goVer())
	var missing []string
	for _, t := range []string{"go", "bun", "node", "ssh"} {
		if _, err := exec.LookPath(t); err != nil {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		fmt.Printf("dépendances manquantes : %s\n", strings.Join(missing, ", "))
	} else {
		fmt.Println("dépendances : toutes présentes")
	}
	hyps := vm.Detect()
	for _, h := range hyps {
		fmt.Printf("hyperviseur détecté : %s (%s)\n", h.Name(), h.Exe())
	}

	// 2 — installation des manquantes
	step(2, "Dépendances")
	for _, m := range missing {
		switch m {
		case "bun":
			fmt.Println("bun manquant — installez-le : https://bun.sh (powershell -c \"irm bun.sh/install.ps1 | iex\" ou curl -fsSL https://bun.sh/install | bash)")
		case "go":
			fmt.Println("go manquant — https://go.dev/dl/")
		default:
			fmt.Printf("%s manquant — installez-le puis relancez cgo setup\n", m)
		}
	}
	if len(missing) > 0 && !confirm("Continuer quand même ?") {
		return 2
	}

	// 3 — build (frontend + binaire)
	step(3, "Construction")
	distIndex := filepath.Join(root, "web", "frontend", "dist", "index.html")
	distPlaceholder := filepath.Join(root, "web", "frontend", "dist", ".placeholder")
	_, hasIndex := os.Stat(distIndex)
	_, hasPlaceholder := os.Stat(distPlaceholder)
	needsBuild := os.IsNotExist(hasIndex) || !os.IsNotExist(hasPlaceholder)
	if needsBuild {
		fmt.Println("frontend non construit (placeholder)")
		if confirm("Builder le frontend maintenant (bun install + build) ?") {
			fmt.Println("→ bun install && bun run build")
			if out, err := kitRun(filepath.Join(root, "web", "frontend"), "bun", "install"); err != nil {
				fmt.Println(out)
				fmt.Println("bun install ÉCHEC — installez bun puis relancez : https://bun.sh")
				return 2
			}
			if out, err := kitRun(filepath.Join(root, "web", "frontend"), "bun", "run", "build"); err != nil {
				fmt.Println(out)
				fmt.Println("bun build ÉCHEC")
				return 2
			}
			_ = os.Remove(distPlaceholder)
		} else {
			fmt.Println("sans dist réel, le dashboard servira la page placeholder — lancez le build avant de servir")
		}
	} else {
		fmt.Println("frontend dist : présent")
	}
	bin := filepath.Join(root, "bin", "cgo")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if _, err := os.Stat(bin); err != nil {
		if confirm("Compiler le binaire local (go build) ?") {
			if out, err := kitRun(root, "go", "build", "-o", bin, "./cmd/cgo"); err != nil {
				fmt.Println(out)
				return 2
			}
			fmt.Printf("binaire : %s\n", bin)
		}
	} else {
		fmt.Printf("binaire : %s (présent)\n", bin)
	}

	// 4 — contexte
	step(4, "Contexte")
	mode := doctor.Mode("auto")
	fmt.Printf("mode détecté : %s\n", mode)
	context := "observe"
	if mode == "full" {
		context = "full"
	}
	if len(hyps) > 0 && !noVM {
		if confirm("Un hyperviseur est présent — configurer le banc VM ?") {
			context = "vm"
		}
	}
	fmt.Printf("contexte choisi : %s\n", context)

	cfgPath := filepath.Join(root, "kit", "cgo-vm.yaml")
	if context == "vm" {
		// 5 — config VM
		step(5, "Configuration VM")
		if _, err := os.Stat(cfgPath); err != nil {
			ex, _ := filepath.Abs(filepath.Join("kit", "cgo-vm.yaml.example"))
			if b, err := os.ReadFile(ex); err == nil {
				_ = os.WriteFile(cfgPath, b, 0644)
				fmt.Printf("config créée : %s (depuis example)\n", cfgPath)
			}
		} else {
			fmt.Printf("config : %s (présente)\n", cfgPath)
		}
		vms := vm.ScanVMs(false)
		if len(vms) > 0 {
			for _, v := range vms {
				fmt.Printf("  vm trouvée : %s\n", v)
			}
			if p := vm.Pick(vms, ""); p != "" {
				hyp := vm.Primary()
				hn := "vmware"
				if hyp != nil && hyp.Name() == "virtualbox" {
					hn = "virtualbox"
				}
				_ = kit.SaveVMX(cfgPath, p, hn)
				fmt.Printf("  sélectionnée : %s (%s)\n", p, hn)
			}
		}
		user := ask("utilisateur SSH VM", "altfloat")
		host := ask("hôte SSH VM (auto = découverte vmrun)", "auto")
		if host == "auto" && len(vms) > 0 {
			if p := vm.Primary(); p != nil && len(vms) > 0 {
				if ip := p.GuestIP(vms[0]); ip != "" {
					host = ip
					fmt.Printf("  IP invitée découverte : %s\n", ip)
				}
			}
		}
		if host == "auto" {
			// pas d'IP découverte — laisser auto, le prochain ensure la trouvera
			fmt.Println("  IP non découverte — laissée en auto (cgo kit ensure la résoudra)")
		}
		_ = setYAML(cfgPath, "host", host)
		_ = setYAML(cfgPath, "user", user)
		if confirm("Déployer sur la VM maintenant (build + push + health) ?") {
			c, _ := kit.LoadConfig(cfgPath)
			r := kit.NewRunner()
			if code := r.Deploy(c, cfgPath, false); code != 0 {
				fmt.Printf("deploy code %d — relancez : cgo kit deploy\n", code)
			}
		}
	} else {
		step(5, "Configuration VM (sans VM — sauté)")
	}

	// 6 — DNS local
	step(6, "DNS local")
	if confirm("Ajouter meteolink.dev/meteolink.vm au hosts local ?") {
		// portable : .dev → local, .vm → IP découverte ou auto
		addHosts("meteolink.dev", "127.0.0.1")
		vmIP := ""
		for _, v := range vm.ScanVMs(false) {
			if p := vm.Primary(); p != nil {
				if ip := p.GuestIP(v); ip != "" {
					vmIP = ip
					break
				}
			}
		}
		if vmIP == "" {
			vmIP = "auto"
		}
		addHosts("meteolink.vm", vmIP)
		fmt.Println("(sous Windows : relancez en Admin si permission refusée — kit/setup-meteolink-dev.bat)")
	}

	// 7 — docteur + récap
	step(7, "Vérification")
	doctor.Print()
	fmt.Println("\nsetup terminé.")
	next := ask("Lancer [dashboard] (cgo --serve), [tui], ou [quitter] ?", "dashboard")
	switch strings.ToLower(next) {
	case "dashboard", "d":
		fmt.Println("→ cgo --serve  (http://localhost:9090)")
		return runServe()
	case "tui", "t":
		fmt.Println("→ cgo tui")
		return runTUI(nil)
	}
	return 0
}

func addHosts(host, ip string) {
	hosts := hostsPath()
	b, err := os.ReadFile(hosts)
	if err != nil {
		fmt.Printf("  hosts illisible : %s\n", hosts)
		return
	}
	for _, ln := range strings.Split(string(b), "\n") {
		if strings.Contains(ln, host) {
			fmt.Printf("  %s déjà présent\n", host)
			return
		}
	}
	f, err := os.OpenFile(hosts, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("  permission refusée sur %s (Admin requis)\n", hosts)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\t%s\n", ip, host)
	fmt.Printf("  %s → %s ajouté\n", host, ip)
}

func hostsPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

func setYAML(path, key, val string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	prefix := ""
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasSuffix(t, "ssh:") {
			prefix = "  "
			continue
		}
		if strings.HasPrefix(t, key+":") && (len(ln)-len(strings.TrimLeft(ln, " "))) <= 2 {
			lines[i] = prefix + key + ": " + val
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
		}
	}
	return nil
}

func kitRun(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func goVer() string {
	out, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.Fields(string(out))[2], "go"))
}

func mustRepoRoot() string {
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
	}
	if ex, err := os.Executable(); err == nil {
		root := filepath.Dir(filepath.Dir(ex))
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root
		}
	}
	fmt.Fprintln(os.Stderr, "lancez depuis la racine du dépôt (go.mod requis)")
	os.Exit(2)
	return "."
}

// runServe — point d'entrée partagé avec main --serve (utilisé par le menu final).
func runServe() int {
	return 0 // le main gère déjà --serve ; ici on délègue en re-exec
}

// mini-flagset local — évite d'exiger l'ordre des flags.
type setupFlags struct {
	m map[string]bool
}

func flagSetup(args []string) *setupFlags {
	f := &setupFlags{m: map[string]bool{}}
	for _, a := range args {
		switch a {
		case "--yes", "-yes":
			f.m["yes"] = true
		case "--no-vm", "-no-vm":
			f.m["no-vm"] = true
		}
	}
	return f
}
func (f *setupFlags) lookupBool(k string) bool { return f.m[k] }
