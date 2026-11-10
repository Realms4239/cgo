package capabilities

import "github.com/Realms4239/cgo/model"

// readCCList et readKernelRelease sont des variables : les tests les
// remplacent pour rester indépendants de l'hôte (un noyau BBRv3 ne doit pas
// changer le nombre de lignes attendu par les tests).
var (
	readCCList        = readHostCCList
	readKernelRelease = readHostKernelRelease
)

// CCPresent — le noyau expose-t-il exactement ce nom ? Le token, pas le
// préfixe : « bbr » ne doit jamais se lire dans « bbr3 ».
func CCPresent(cc string) bool {
	for _, c := range readCCList() {
		if c == cc {
			return true
		}
	}
	return false
}

// DefaultCCs — matrice pleine : cubic et bbr (v1) toujours, bbr3 seulement
// si le noyau l'enregistre. Une demande explicite de bbr3 reste gatée dans
// startMatrix : l'omission silencieuse n'est permise que pour le défaut.
func DefaultCCs() []string {
	out := make([]string, 0, len(model.AllCC)+1)
	for _, c := range model.AllCC {
		out = append(out, string(c))
	}
	if CCPresent(string(model.BBRv3)) {
		out = append(out, string(model.BBRv3))
	}
	return out
}

// KernelRelease — la release du noyau ("" si illisible).
func KernelRelease() string { return readKernelRelease() }
