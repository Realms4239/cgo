// Package frontend embeds the built SPA (web/frontend/dist).
// The tracked dist/index.html keeps //go:embed all:dist valid before the
// first frontend build; Makefile rebuilds dist first.
package frontend

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// ValidateEmbeddedAssets refuses a placeholder embed: dist/index.html alone
// serves a blank dashboard. The binary must carry the hashed JS bundle.
func ValidateEmbeddedAssets() error { return validateAssets(FS) }

// validateAssets checks a dist-rooted FS: a real index.html and at least one
// hashed JS bundle.
func validateAssets(fsys fs.FS) error {
	if _, err := fs.Stat(fsys, "index.html"); err != nil {
		return fmt.Errorf("SPA embarqué : dist/index.html manquant : %w", err)
	}
	js, err := fs.Glob(fsys, "assets/*.js")
	if err != nil {
		return fmt.Errorf("SPA embarqué : dist/assets illisible : %w", err)
	}
	if len(js) == 0 {
		return fmt.Errorf("SPA embarqué : aucun dist/assets/*.js — construisez le frontend")
	}
	return nil
}

// FS serves dist contents at root ("/index.html" etc.).
var FS fs.FS = func() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}()
