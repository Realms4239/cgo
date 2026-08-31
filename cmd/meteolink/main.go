// meteolink — alias du binaire unique cgo (même noyau, nom compatible).
// Usage :
//   meteolink top [--addr http://localhost:9090] [--interval 250ms]   cartes ASCII, anneaux live
//   meteolink tui|setup|kit|run|serve|version …                      tout cgo, même moteur
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// main — re-exécute le binaire cgo voisin (même archive) avec les mêmes
// arguments ; le nom meteolink reste pour la compat des scripts et npm.
func main() {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "meteolink:", err)
		os.Exit(1)
	}
	dir := filepath.Dir(self)
	cgo := filepath.Join(dir, "cgo")
	if runtime.GOOS == "windows" {
		cgo += ".exe"
	}
	// archive installée : les deux binaires vivent côte à côte ; sinon on
	// cherche cgo dans le PATH ; sinon on prévient honnêtement.
	if _, err := os.Stat(cgo); err == nil {
		passExit(cgo)
	}
	if p, err := exec.LookPath("cgo"); err == nil {
		passExit(p)
	}
	fmt.Fprintln(os.Stderr, "meteolink : binaire cgo introuvable — installez l'archive cgo-<os>-<arch> complète (cgo + meteolink)")
	os.Exit(1)
}

func passExit(path string) {
	cmd := exec.Command(path, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "meteolink:", err)
		os.Exit(1)
	}
	os.Exit(0)
}
