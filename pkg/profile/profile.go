package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

const file = "data/profiles.json"

// Reserved — le référentiel P1–P4 n'est ni importable ni supprimable :
// un import "P3" masquerait le VSAT 5/600/30/1 dans la carte live ET dans
// les campagnes dashboard (Load fusionne le fichier par-dessus), sans
// qu'aucune erreur ne le dise. Vu en prod : un P3 35/60 fantôme.
func Reserved(id string) bool {
	return id == "P1" || id == "P2" || id == "P3" || id == "P4"
}

// Import enregistre un profil dans model.Profiles ET le persiste (source unique).
func Import(p model.Profile) error {
	if p.ID == "" {
		return nil
	}
	if Reserved(p.ID) {
		return fmt.Errorf("profil %s réservé (référentiel P1–P4) — choisissez un autre identifiant, ex. P-site-fibre", p.ID)
	}
	// Tout sous verrou : deux imports simultanés ne doivent ni perdre une
	// écriture (lecture-modification-écriture du fichier) ni courir avec un
	// lecteur de la carte.
	model.ProfilesMu.Lock()
	defer model.ProfilesMu.Unlock()
	model.Profiles[p.ID] = p
	m := map[string]model.Profile{}
	if data, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	m[p.ID] = p
	data, _ := json.MarshalIndent(m, "", "  ")
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// Load fusionne les profils persistés dans model.Profiles — sauf les
// identifiants réservés, qu'un vieux fichier ne doit plus masquer.
func Load() {
	model.ProfilesMu.Lock()
	defer model.ProfilesMu.Unlock()
	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	m := map[string]model.Profile{}
	if err := json.Unmarshal(data, &m); err != nil {
		return
	}
	for k, v := range m {
		if Reserved(k) {
			continue
		}
		model.Profiles[k] = v
	}
}

// Delete retire un profil importé (carte + fichier). Le référentiel est
// intouchable ; un identifiant inconnu est une erreur honnête, pas un OK.
func Delete(id string) error {
	if Reserved(id) {
		return fmt.Errorf("profil %s réservé (référentiel P1–P4) — suppression refusée", id)
	}
	model.ProfilesMu.Lock()
	defer model.ProfilesMu.Unlock()
	m := map[string]model.Profile{}
	if data, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	if _, ok := m[id]; !ok {
		if _, ok := model.Profiles[id]; !ok {
			return fmt.Errorf("profil inconnu : %s", id)
		}
	}
	delete(m, id)
	delete(model.Profiles, id)
	data, _ := json.MarshalIndent(m, "", "  ")
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}
