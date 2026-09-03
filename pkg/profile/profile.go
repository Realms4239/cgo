package profile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

const file = "data/profiles.json"

// Import enregistre un profil dans model.Profiles ET le persiste (source unique).
func Import(p model.Profile) error {
	if p.ID == "" {
		return nil
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

// Load fusionne les profils persistés dans model.Profiles.
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
		model.Profiles[k] = v
	}
}
