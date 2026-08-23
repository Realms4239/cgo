package profile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

const file = "data/profiles.json"

// Import adds a custom profile and persists it.
func Import(p model.Profile) error {
	if p.ID == "" {
		return nil
	}
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

// Load merges persisted profiles into model.Profiles.
func Load() {
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
