// Commande cgo-attest — vérifie un manifeste de scellés SHA-256.
// Sortie 0 = tout scellé vérifié ; 1 = premier écart (fichier absent ou
// différent). Ne modifie rien : la provenance est additive seulement.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Realms4239/cgo/attest"
)

func main() {
	fs := flag.NewFlagSet("cgo-attest", flag.ExitOnError)
	// Manifeste : une ligne par fichier, format "<chemin> <sha256hex>".
	manifest := fs.String("manifest", "", "chemin du manifeste de scellés (obligatoire)")
	_ = fs.Parse(os.Args[1:])
	if *manifest == "" {
		fmt.Fprintln(os.Stderr, "cgo-attest: --manifest est obligatoire")
		os.Exit(2)
	}
	if err := attest.VerifyManifest(*manifest); err != nil {
		fmt.Fprintln(os.Stderr, "cgo-attest:", err)
		os.Exit(1)
	}
	fmt.Println("cgo-attest: tous les scellés vérifiés")
}
