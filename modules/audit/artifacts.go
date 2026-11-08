package audit

import (
	"path/filepath"
)

// DefaultDataDir — la zone d'artefacts par défaut, relative au répertoire de
// travail. Le CLI accepte un -data-dir et l'audit l'écrit toujours sous data/ ;
// ce qui est scellé dans un rapport est un chemin RELATIF, pas une promesse
// absolue — c'est ce que la console peut afficher sans mentir.
const DefaultDataDir = "data"

// ReportArtifactName — le nom canonique du rapport scellé.
const ReportArtifactName = "audit_report.json"

// ReportMarkdownName et TelemetryArtifactName — les deux autres artefacts écrits
// dans le même dossier.
const (
	ReportMarkdownName    = "audit_report.md"
	TelemetryArtifactName = "telemetry.csv"
)

// ArtifactDir — LA convention de rangement : un dossier par audit, un
// identifiant assaini par ArtifactID. Six appelsants construisaient ce chemin
// chacun de leur côté (lecture API, stat de présence, écriture du rapport,
// telemetry.csv, calibrate, TUI) ; la désanction d'un seul d'entre eux rendrait
// l'évasion de chemin E11 possible à cet endroit-là.
func ArtifactDir(baseDir, id string) (string, error) {
	clean, err := ArtifactID(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "audits", clean), nil
}

// ArtifactFile — un fichier de l'audit, dans son dossier.
func ArtifactFile(baseDir, id, name string) (string, error) {
	dir, err := ArtifactDir(baseDir, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// ArtifactPointer — le chemin scellé dans le rapport, en séparateurs obliques
// pour que l'artefact lu sur Windows se relise pareil dans la console. Un
// identifiant refusé rend une ABSENCE, jamais un chemin approximatif : mieux
// vaut un tiroir qui dit « non enregistré » qu'un lien vers un dossier voisin.
func ArtifactPointer(id string) string {
	p, err := ArtifactFile(DefaultDataDir, id, ReportArtifactName)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(p)
}
