package capabilities

import (
	"encoding/json"
	"os"
	"strings"
)

// Épinglage du noyau Google BBRv3, verbatim de kit/install-bbr3.sh (la
// référence immuable). Le dépôt google/bbr n'enregistre pas la v3 sous un
// token « bbr3 » : dans cet arbre, « bbr » EST la v3 (TCP_CONG_BBR=y). Un
// hôte provisionné se reconnaît donc à deux preuves concordantes — le
// marqueur déposé par le script d'installation et le suffixe -g<commit> que
// le noyau signe dans /proc/version.
const (
	bbr3Tag         = "bbrv3-2025-03-18"
	bbr3CommitFull  = "90210de4b779d40496dee0b89081780eeddf2a60"
	bbr3CommitShort = "90210de4b779" // 12 premiers caractères du commit épinglé
)

// markerFile — là où kit/install-bbr3.sh dépose le marqueur après installation.
const markerFile = "/var/lib/cgo/bbr3-kernel.json"

// readMarker / readVersion — variables : les tests les remplacent pour
// rester indépendants de l'hôte (même forme que readCCList).
var (
	readMarker  = readHostMarker
	readVersion = readHostVersion
)

func readHostMarker() (string, error) {
	raw, err := os.ReadFile(markerFile)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func readHostVersion() (string, error) {
	raw, err := os.ReadFile("/proc/version")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// StubProvisioned — remplace marqueur et version par un hôte simulé, rend
// une fonction de restauration. Exporté pour que les tests du moteur de
// campagne (hors module) puissent simuler un hôte provisionné (miroir de
// StubHost).
func StubProvisioned(marker string, version string) func() {
	oldMarker, oldVersion := readMarker, readVersion
	readMarker = func() (string, error) { return marker, nil }
	readVersion = func() (string, error) { return version, nil }
	return func() { readMarker, readVersion = oldMarker, oldVersion }
}

// bbr3Marker — seul le commit du marqueur transporte la preuve d'épingle ;
// le reste du JSON (branche, tag, kernel, built) n'est pas exigé.
type bbr3Marker struct {
	Commit string `json:"commit"`
}

// ProvisionedBBRv3 — l'hôte est-il provisionné avec le noyau Google BBRv3
// épinglé ? Les deux preuves sont exigées, aucune ne suffit seule :
//  1. le marqueur /var/lib/cgo/bbr3-kernel.json existe et son commit
//     commence par le pin court (source épinglée, pas une branche dérivée) ;
//  2. /proc/version contient « g » + les 12 premiers caractères du commit
//     (le suffixe que le build Google signe dans la release).
//
// Marqueur absent, invalide, commit hors pin ou version sans le suffixe ⇒
// false. Aucun repli fichier-seul ni version-seul.
func ProvisionedBBRv3() bool {
	raw, err := readMarker()
	if err != nil {
		return false
	}
	var m bbr3Marker
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return false
	}
	if !strings.HasPrefix(m.Commit, bbr3CommitShort) {
		return false
	}
	v, err := readVersion()
	if err != nil {
		return false
	}
	return strings.Contains(v, "g"+m.Commit[:12])
}

// BBRv3Token — le token que le noyau attend pour la v3, ou "" si la v3
// n'est pas lançable. « bbr3 » nu si le noyau l'enregistre ; « bbr » sur un
// hôte provisionné (le seul token v3 de cet arbre) ; sinon refus.
func BBRv3Token() string {
	if CCPresent("bbr3") {
		return "bbr3"
	}
	if ProvisionedBBRv3() && CCPresent("bbr") {
		return "bbr"
	}
	return ""
}
